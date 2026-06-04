package diag

import (
	"sort"
	"strings"
)

// Recommendation is an actionable layout or authoring hint (non-blocking).
type Recommendation struct {
	Priority int    `json:"priority"` // 1 = highest
	Category string `json:"category"` // layout, edges, labels, slides, style
	Code     string `json:"code,omitempty"`
	Message  string `json:"message"`
	Fix      string `json:"fix"`
	Example  string `json:"example,omitempty"`
}

// BuildRecommendations turns warnings and validation context into prioritized hints.
func BuildRecommendations(r Report) []Recommendation {
	var out []Recommendation
	seen := map[string]struct{}{}

	add := func(priority int, category, code, message, fix, example string) {
		key := category + "|" + code + "|" + message
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, Recommendation{
			Priority: priority,
			Category: category,
			Code:     code,
			Message:  message,
			Fix:      fix,
			Example:  example,
		})
	}

	for _, w := range r.Warnings {
		cat := categoryForCode(w.Code)
		priority := priorityForCode(w.Code)
		add(priority, cat, string(w.Code), w.Message, w.Fix, w.Example)
	}

	if r.OK && len(r.Warnings) == 0 {
		add(3, "workflow", "render_ready",
			"Spec validates cleanly — layout looks good from geometry checks.",
			"Run sceno describe -i "+quote(r.Input)+" --json, then sceno render --all.",
			"sceno render -i "+quote(r.Input)+" -o output/sceno --all")
	}

	if !r.OK && len(r.Errors) > 0 {
		add(1, "validation", string(r.Errors[0].Code),
			"Fix blocking errors before render.",
			r.Errors[0].Fix,
			r.Errors[0].Example)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Priority != out[j].Priority {
			return out[i].Priority < out[j].Priority
		}
		return out[i].Message < out[j].Message
	})
	return out
}

func categoryForCode(code Code) string {
	switch code {
	case CodeEdgeCollision, CodeEdgeHidden:
		return "edges"
	case CodeMisaligned, CodeSuggestCompact, CodeCollision, CodeOccluded:
		return "layout"
	case CodeTextOverflow:
		return "labels"
	default:
		return "spec"
	}
}

func priorityForCode(code Code) int {
	switch code {
	case CodeEdgeCollision:
		return 2
	case CodeEdgeHidden, CodeMisaligned, CodeOccluded:
		return 2
	case CodeSuggestCompact:
		return 3
	default:
		return 2
	}
}

// EnrichRecommendations attaches recommendations and extends agent next_steps.
func (r *Report) EnrichRecommendations(recs []Recommendation) {
	if len(recs) == 0 {
		return
	}
	for _, rec := range recs {
		if rec.Priority > 2 {
			continue
		}
		step := rec.Message
		if rec.Fix != "" {
			step += " — " + rec.Fix
		}
		if !containsStep(r.Agent.NextSteps, step) {
			r.Agent.NextSteps = append([]string{step}, r.Agent.NextSteps...)
		}
	}
}

func containsStep(steps []string, step string) bool {
	prefix := strings.Split(step, " — ")[0]
	for _, s := range steps {
		if s == step || strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}
