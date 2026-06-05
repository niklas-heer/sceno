package scene

import (
	"fmt"
	"strings"

	"github.com/niklas-heer/sceno/internal/diag"
	"github.com/niklas-heer/sceno/internal/model"
)

// Evaluation is the single source of truth for scene semantics after layout.
// All validation, advise, describe, and render paint-order contracts derive from this.
type Evaluation struct {
	Scene       Report       `json:"scene"`
	Stack       StackSummary `json:"stack"`
	PaintOrder  []PaintItem  `json:"paint_order"`
	Findings    []Finding    `json:"findings"`
	Score       int          `json:"visual_score"`
	RulesRun    []string     `json:"rules_run"`
	Summary     string       `json:"summary"`
	VisualRules []VisualRule `json:"visual_rules"`
}

// Evaluate runs the full stack engine on a laid-out diagram.
// Geometry must already be computed by pipeline.Build — this does not lay out nodes.
func Evaluate(d *model.Diagram) Evaluation {
	if d == nil {
		return Evaluation{Score: 0, Summary: "empty diagram", VisualRules: VisualRulesCatalog}
	}

	stack := BuildStack(d)
	sceneReport := analyzeCore(d)
	ctx := ruleContext{d: d, stack: stack, scene: sceneReport}

	var findings []Finding
	var rulesRun []string
	for _, r := range engineRules {
		rulesRun = append(rulesRun, r.id)
		findings = append(findings, r.fn(ctx)...)
	}
	findings = dedupeFindings(findings)

	score := sceneReport.Aesthetics.Overall
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

	summary := fmt.Sprintf("stack engine: %d plane(s), %d finding(s), score %d/100",
		len(stack.Planes), len(findings), score)
	if len(findings) > 0 {
		summary += " — " + findings[0].Message
	}

	// Scene issues mirror engine findings (non-hint) for describe/advise consumers.
	sceneReport.Issues = findingsToIssues(findings)
	sceneReport.Stack = stack.Summary()

	return Evaluation{
		Scene:       sceneReport,
		Stack:       stack.Summary(),
		PaintOrder:  sceneReport.PaintOrder,
		Findings:    findings,
		Score:       score,
		RulesRun:    rulesRun,
		Summary:     summary,
		VisualRules: VisualRulesCatalog,
	}
}

// EngineReport converts Evaluation to the legacy engine report shape (JSON compat).
func (ev Evaluation) EngineReport() EngineReport {
	var issues []diag.Issue
	for _, f := range ev.Findings {
		if f.Severity == "hint" {
			continue
		}
		issues = append(issues, f.ToIssue())
	}
	return EngineReport{
		Stack:       ev.Stack,
		RulesRun:    ev.RulesRun,
		Findings:    ev.Findings,
		Issues:      issues,
		Summary:     ev.Summary,
		Score:       ev.Score,
		VisualRules: ev.VisualRules,
	}
}

// MergeEvaluations combines per-slide evaluations (deck-level advise: min score, merged findings).
func MergeEvaluations(evals []Evaluation) Evaluation {
	if len(evals) == 0 {
		return Evaluation{VisualRules: VisualRulesCatalog}
	}
	if len(evals) == 1 {
		return evals[0]
	}
	merged := Evaluation{
		Score:       100,
		VisualRules: VisualRulesCatalog,
	}
	seenRules := map[string]struct{}{}
	for _, ev := range evals {
		if ev.Score < merged.Score {
			merged.Score = ev.Score
			merged.Stack = ev.Stack
			merged.Summary = ev.Summary
		}
		merged.Findings = append(merged.Findings, ev.Findings...)
		for _, r := range ev.RulesRun {
			if _, ok := seenRules[r]; !ok {
				seenRules[r] = struct{}{}
				merged.RulesRun = append(merged.RulesRun, r)
			}
		}
	}
	merged.Findings = dedupeFindings(merged.Findings)
	merged.Summary = fmt.Sprintf("%d slides evaluated; deck score %d/100; %d finding(s)",
		len(evals), merged.Score, len(merged.Findings))
	return merged
}

func findingsToIssues(findings []Finding) []diag.Issue {
	var out []diag.Issue
	for _, f := range findings {
		if f.Severity == "hint" {
			continue
		}
		out = append(out, f.ToIssue())
	}
	return out
}

// PaintOrderDescription documents the render contract for agents.
func PaintOrderDescription() string {
	return strings.Join([]string{
		"Paint order (back → front) is fixed and shared by render and engine:",
		"1. canvas background",
		"2. lanes, frames, groups (container backgrounds)",
		"3. edges (connector strokes)",
		"4. nodes (shapes)",
		"5. edge labels",
		"6. arrowheads",
		"scene.PaintsBeforeEdges / scene.BuildPaintOrder are the source of truth; render delegates to them.",
	}, "\n")
}
