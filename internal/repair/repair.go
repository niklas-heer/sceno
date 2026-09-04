// Package repair previews conservative, verified source edits without writing files.
package repair

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/niklas-heer/sceno/internal/diag"
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/pipeline"
	"github.com/niklas-heer/sceno/internal/spec"
	"github.com/niklas-heer/sceno/internal/validate"
)

type Request struct {
	SlideIndex int               `json:"slide_index"`
	Target     string            `json:"target"`
	Properties map[string]string `json:"properties"`
	IssueID    string            `json:"issue_id"`
}

type NodeMove struct {
	SlideIndex int        `json:"slide_index"`
	Target     string     `json:"target"`
	Before     model.Rect `json:"before"`
	After      model.Rect `json:"after"`
	DX         float64    `json:"dx"`
	DY         float64    `json:"dy"`
}

type PreviewResult struct {
	Applicable   bool        `json:"applicable"`
	Reasons      []string    `json:"reasons"`
	BeforeSource string      `json:"before_source"`
	AfterSource  string      `json:"after_source"`
	BeforeScore  float64     `json:"before_score"`
	AfterScore   float64     `json:"after_score"`
	Moves        []NodeMove  `json:"moves"`
	BeforeReport diag.Report `json:"before_report"`
	AfterReport  diag.Report `json:"after_report"`
}

// IssueID fingerprints exact current evidence and offered repairs. Stale evidence
// cannot authorize an edit after the source geometry or candidate values change.
func IssueID(issue diag.Issue) string {
	// File paths and source lines are presentation metadata. Preview evaluates
	// in memory under a display name, and source-preserving edits can shift
	// lines without changing the issue's geometric evidence.
	issue.Path = ""
	issue.Line = 0
	data, _ := json.Marshal(issue)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// Preview only accepts a current engine-offered geometry edit for a specific
// issue. Successful verification may leave unrelated existing issues unresolved.
func Preview(source []byte, request Request) (PreviewResult, error) {
	out := PreviewResult{BeforeSource: string(source), AfterSource: string(source)}
	before, report, err := validate.EvaluateSource(source, "preview.kdl", validate.Options{FixCollisions: true})
	out.BeforeReport = report
	out.AfterReport = report
	if err != nil {
		return out, fmt.Errorf("cannot evaluate source: %w", err)
	}
	out.BeforeScore = float64(before.MergedEval().Score)
	out.AfterScore = out.BeforeScore
	var issue *diag.Issue
	for _, candidate := range allIssues(report) {
		if IssueID(candidate) == request.IssueID {
			copy := candidate
			issue = &copy
			break
		}
	}
	if issue == nil {
		out.Reasons = []string{"The targeted issue is stale or no longer exists."}
		return out, nil
	}
	if request.SlideIndex < 1 || issue.SlideIndex != request.SlideIndex {
		out.Reasons = []string{"The issue does not belong to the requested slide."}
		return out, nil
	}
	offered := false
	for _, candidate := range issue.Repairs {
		if candidate.Action == "set_property" && candidate.Target == request.Target && reflect.DeepEqual(candidate.Properties, request.Properties) {
			offered = true
			break
		}
	}
	if !offered {
		out.Reasons = []string{"The requested properties are not a current offered repair for this issue."}
		return out, nil
	}
	if err := geometryProperties(request.Properties); err != nil {
		out.Reasons = []string{err.Error()}
		return out, nil
	}
	updated, err := spec.SetNodeProperties(source, request.SlideIndex, request.Target, request.Properties)
	if err != nil {
		out.Reasons = []string{err.Error()}
		return out, nil
	}
	out.AfterSource = string(updated)
	after, afterReport, err := validate.EvaluateSource(updated, "preview.kdl", validate.Options{FixCollisions: true})
	out.AfterReport = afterReport
	if err != nil {
		out.Reasons = []string{"The proposed source does not parse or evaluate: " + err.Error()}
		return out, nil
	}
	out.AfterScore = float64(after.MergedEval().Score)
	out.Moves = nodeMoves(before, after)
	targetKey := issueKey(*issue)
	for _, remaining := range allIssues(afterReport) {
		if issueKey(remaining) == targetKey {
			out.Reasons = append(out.Reasons, "The targeted issue remains after the edit.")
			break
		}
	}
	out.Reasons = append(out.Reasons, structuralRegressions(report, afterReport, before, after)...)
	if len(out.Reasons) == 0 {
		out.Applicable = true
		out.Reasons = []string{"The targeted issue is resolved without adding or worsening structural findings."}
	}
	return out, nil
}

// structuralRegressions requires each remaining finding to have unchanged
// evidence and equal or lower severity. Finding counts alone cannot establish
// this: a move can resolve one collision while deepening an existing one.
func structuralRegressions(beforeReport, afterReport diag.Report, before, after pipeline.Result) []string {
	prior := structuralEvidence(beforeReport, before)
	next := structuralEvidence(afterReport, after)
	keys := make([]string, 0, len(next))
	for key := range next {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var reasons []string
	for _, key := range keys {
		old := prior[key]
		current := next[key]
		issue := current[0].issue
		if len(current) > len(old) {
			reasons = append(reasons, fmt.Sprintf("The edit introduces or increases a structural finding on slide %d: %s.", issue.SlideIndex, issue.Code))
			continue
		}
		used := make([]bool, len(old))
		for _, candidate := range current {
			match := -1
			for i, previous := range old {
				if !used[i] && previous.evidence == candidate.evidence && previous.severity >= candidate.severity {
					match = i
					break
				}
			}
			if match >= 0 {
				used[match] = true
				continue
			}
			reason := "changes structural evidence whose non-worsening cannot be verified"
			for i, previous := range old {
				if used[i] {
					continue
				}
				if previous.evidence == candidate.evidence && previous.severity < candidate.severity {
					reason = "escalates a structural warning to an error"
					break
				}
				if previous.issue.Geometry != nil && candidate.issue.Geometry != nil {
					a, b := previous.issue.Geometry.Overlap, candidate.issue.Geometry.Overlap
					if b.W*b.H > a.W*a.H {
						reason = "increases an existing structural overlap area"
					}
				}
			}
			reasons = append(reasons, fmt.Sprintf("The edit %s on slide %d: %s.", reason, candidate.issue.SlideIndex, candidate.issue.Code))
			break
		}
	}
	return reasons
}

func geometryProperties(properties map[string]string) error {
	if len(properties) == 0 {
		return fmt.Errorf("repair has no properties")
	}
	keys := make([]string, 0, len(properties))
	for key := range properties {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := properties[key]
		switch key {
		case "at":
			parts := strings.Split(value, ",")
			if len(parts) != 2 {
				return fmt.Errorf("invalid at coordinate")
			}
			for _, p := range parts {
				n, err := strconv.Atoi(p)
				if err != nil || n < 0 || n > spec.MaxGridCoordinate {
					return fmt.Errorf("invalid at coordinate")
				}
			}
		case "row", "layer", "col", "column":
			n, err := strconv.Atoi(value)
			if err != nil || n < 0 || n > spec.MaxGridCoordinate {
				return fmt.Errorf("invalid grid property %q", key)
			}
		case "dx", "dy", "x", "y", "w", "h":
			n, err := strconv.ParseFloat(value, 64)
			if err != nil || math.IsInf(n, 0) || math.IsNaN(n) || math.Abs(n) > spec.MaxGeometryMagnitude {
				return fmt.Errorf("invalid geometry property %q", key)
			}
			if (key == "w" || key == "h") && n <= 0 {
				return fmt.Errorf("size must be positive")
			}
		default:
			return fmt.Errorf("property %q is not an automatic geometry repair", key)
		}
	}
	return nil
}

func allIssues(report diag.Report) []diag.Issue {
	out := make([]diag.Issue, 0, len(report.Errors)+len(report.Warnings))
	out = append(out, report.Errors...)
	return append(out, report.Warnings...)
}

// issueKey ignores changing prose and geometry, retaining semantic participants.
func issueKey(issue diag.Issue) string {
	nodes := append([]string(nil), issue.Nodes...)
	sort.Strings(nodes)
	edge := append([]string(nil), issue.Edge...)
	b, _ := json.Marshal([]any{issue.SlideIndex, issue.Code, nodes, edge})
	return string(b)
}

type observedIssue struct {
	issue    diag.Issue
	severity int // warnings=1, errors=2
	evidence string
}

func structuralEvidence(report diag.Report, result pipeline.Result) map[string][]observedIssue {
	groups := map[string][]observedIssue{}
	// Match errors first so a later warning cannot consume the only prior
	// error that justifies a remaining error with the same evidence.
	for index, issues := range [][]diag.Issue{report.Errors, report.Warnings} {
		for _, issue := range issues {
			switch issue.Code {
			case diag.CodeDenseLayout, diag.CodeSparseLayout, diag.CodeSuggestCompact, diag.CodeSlideCrowded, diag.CodeWeakHierarchy, diag.CodeTooManyElements, diag.CodeSuggestAnnotation:
				continue
			}
			var evidence any = issue.Geometry
			if issue.Geometry == nil {
				// Some routing and visibility rules only expose prose. Unchanged
				// prose is insufficient when the underlying scene has changed;
				// conservatively retain the full computed stack for that slide.
				var geometry any
				if slide := issue.SlideIndex - 1; slide >= 0 && slide < len(result.Slides) {
					geometry = []any{result.Slides[slide].Eval.SceneStack, result.Slides[slide].Diagram.Routed}
				}
				evidence = []any{issue.Message, geometry}
			}
			data, _ := json.Marshal(evidence)
			key := issueKey(issue)
			groups[key] = append(groups[key], observedIssue{issue: issue, severity: 2 - index, evidence: string(data)})
		}
	}
	return groups
}

func nodeMoves(before, after pipeline.Result) []NodeMove {
	var moves []NodeMove
	for si, slide := range before.Slides {
		if si >= len(after.Slides) {
			continue
		}
		lookup := map[string]model.Node{}
		for _, n := range after.Slides[si].Diagram.Nodes {
			lookup[n.ID] = n
		}
		for _, old := range slide.Diagram.Nodes {
			n, ok := lookup[old.ID]
			if ok && old.Rect != n.Rect {
				moves = append(moves, NodeMove{SlideIndex: si + 1, Target: n.ID, Before: old.Rect, After: n.Rect, DX: n.Rect.X - old.Rect.X, DY: n.Rect.Y - old.Rect.Y})
			}
		}
	}
	return moves
}
