package scene

import (
	"fmt"
	"sort"
	"strings"

	"github.com/niklas-heer/sceno/internal/diag"
	"github.com/niklas-heer/sceno/internal/model"
)

// EngineReport is the output of the multi-plane validation engine.
type EngineReport struct {
	Stack      StackSummary `json:"stack"`
	RulesRun   []string     `json:"rules_run"`
	Findings   []Finding    `json:"findings"`
	Issues     []diag.Issue `json:"issues,omitempty"`
	Summary    string       `json:"summary"`
	Score      int          `json:"score"` // 0–100 visual quality
	VisualRules []VisualRule `json:"visual_rules,omitempty"`
}

// RunEngine analyzes a laid-out diagram using the stacked-plane model and visual rules.
func RunEngine(d *model.Diagram) EngineReport {
	if d == nil {
		return EngineReport{Score: 0, Summary: "empty diagram"}
	}
	stack := BuildStack(d)
	scene := analyzeCore(d)
	ctx := ruleContext{d: d, stack: stack, scene: scene}

	findings := make([]Finding, 0)
	var rulesRun []string
	for _, r := range engineRules {
		rulesRun = append(rulesRun, r.id)
		findings = append(findings, r.fn(ctx)...)
	}
	findings = dedupeFindings(findings)

	score := scene.Aesthetics.Overall
	for _, f := range findings {
		switch f.Severity {
		case "error":
			score -= 15
		case "warning":
			score -= 8
		case "hint":
			score -= 2
		}
	}
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	var issues []diag.Issue
	for _, f := range findings {
		if f.Severity == "hint" {
			continue
		}
		issues = append(issues, f.ToIssue())
	}

	summary := fmt.Sprintf("stack engine: %d plane(s), %d finding(s), score %d/100",
		len(stack.Planes), len(findings), score)
	if len(findings) > 0 {
		summary += " — " + findings[0].Message
	}

	return EngineReport{
		Stack:       stack.Summary(),
		RulesRun:    rulesRun,
		Findings:    findings,
		Issues:      issues,
		Summary:     summary,
		Score:       score,
		VisualRules: VisualRulesCatalog,
	}
}

func dedupeFindings(in []Finding) []Finding {
	seen := map[string]struct{}{}
	var out []Finding
	for _, f := range in {
		key := f.Code + "|" + f.Message
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool {
		si, sj := severityRank(out[i].Severity), severityRank(out[j].Severity)
		if si != sj {
			return si < sj
		}
		return out[i].Message < out[j].Message
	})
	return out
}

func severityRank(s string) int {
	switch s {
	case "error":
		return 0
	case "warning":
		return 1
	default:
		return 2
	}
}

// FindingsToRecommendations converts engine hints into diag recommendations.
func FindingsToRecommendations(findings []Finding) []diag.Recommendation {
	var out []diag.Recommendation
	for _, f := range findings {
		if f.Severity != "hint" && f.Severity != "warning" {
			continue
		}
		priority := 3
		if f.Severity == "warning" {
			priority = 2
		}
		out = append(out, diag.Recommendation{
			Priority: priority,
			Category: categoryForRule(f.RuleID),
			Code:     f.Code,
			Message:  f.Message,
			Fix:      f.Fix,
			Example:  f.Example,
		})
	}
	return out
}

func categoryForRule(ruleID string) string {
	switch ruleID {
	case "routing_plane", "edge_clarity":
		return "edges"
	case "slide_focus", "hierarchy":
		return "slides"
	case "annotations":
		return "annotations"
	default:
		return "layout"
	}
}

// EngineNarrative is a one-line agent summary.
func EngineNarrative(er EngineReport) string {
	parts := []string{er.Summary}
	if len(er.Findings) > 0 {
		var kinds []string
		for _, f := range er.Findings {
			if f.Severity == "error" || f.Severity == "warning" {
				kinds = append(kinds, f.RuleID)
			}
		}
		if len(kinds) > 0 {
			parts = append(parts, "rules: "+strings.Join(uniqueStrings(kinds), ", "))
		}
	}
	return strings.Join(parts, "; ")
}

func uniqueStrings(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
