package advise

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/niklas-heer/sceno/internal/diag"
	"github.com/niklas-heer/sceno/internal/pipeline"
	"github.com/niklas-heer/sceno/internal/scene"
	"github.com/niklas-heer/sceno/internal/spec"
	"github.com/niklas-heer/sceno/internal/validate"
	"github.com/niklas-heer/sceno/internal/version"
)

// Options controls advise depth and optional AI review.
type Options struct {
	FixCollisions bool
	UseAI         bool
	AICmd         string // defaults to SCENO_AI_CMD env
}

// Report combines validation, stack engine, scene analysis, and recommendations.
type Report struct {
	Input           string                  `json:"input"`
	Tool            string                  `json:"tool"`
	Version         string                  `json:"version"`
	ValidationOK    bool                    `json:"validation_ok"`
	VisualScore     int                     `json:"visual_score"`
	Stack           scene.StackSummary      `json:"stack"`
	Engine          scene.EngineReport      `json:"engine"`
	VisualRules     []scene.VisualRule      `json:"visual_rules"`
	Recommendations []diag.Recommendation   `json:"recommendations"`
	Agent           AdviseMeta              `json:"agent"`
	AIReview        string                  `json:"ai_review,omitempty"`
}

type AdviseMeta struct {
	Summary   string   `json:"summary"`
	NextSteps []string `json:"next_steps"`
	Hint      string   `json:"hint"`
}

// Run performs deep visual validation and returns prioritized advice.
func Run(path string, opt Options) (Report, error) {
	vreport, _, err := validate.Run(path, validate.Options{FixCollisions: opt.FixCollisions})
	if err != nil && !strings.Contains(err.Error(), "spec invalid") && !strings.Contains(err.Error(), "layout") {
		return Report{}, err
	}

	var engine scene.EngineReport
	if s, err := spec.LoadFile(path); err == nil {
		popt := pipeline.DefaultOptions()
		popt.ResolveCollision = opt.FixCollisions
		if deck, _, err := pipeline.BuildDeck(s, popt); err == nil && len(deck.Slides) > 0 {
			engine = scene.RunEngine(&deck.Slides[0])
		}
	}
	recs := diag.BuildRecommendations(vreport)
	recs = append(recs, scene.FindingsToRecommendations(engine.Findings)...)
	recs = dedupeRecs(recs)

	out := Report{
		Input:           path,
		Tool:            "sceno",
		Version:         version.Version,
		ValidationOK:    vreport.OK,
		VisualScore:     engine.Score,
		Stack:           engine.Stack,
		Engine:          engine,
		VisualRules:     scene.VisualRulesCatalog,
		Recommendations: recs,
		Agent: AdviseMeta{
			Summary: buildSummary(vreport, engine),
			Hint:    "Stack validation uses layered 2D planes (lanes→edges→annotations→nodes→labels). Run sceno guide --json for shape catalog.",
			NextSteps: buildNextSteps(path, vreport, engine),
		},
	}

	if opt.UseAI {
		cmd := opt.AICmd
		if cmd == "" {
			cmd = os.Getenv("SCENO_AI_CMD")
		}
		if cmd == "" {
			out.Agent.NextSteps = append(out.Agent.NextSteps,
				"Set SCENO_AI_CMD or use --ai-cmd to invoke an external AI CLI (e.g. codex exec -).")
		} else {
			review, aiErr := invokeAI(cmd, out)
			if aiErr != nil {
				out.AIReview = "ai error: " + aiErr.Error()
			} else {
				out.AIReview = review
			}
		}
	}
	return out, nil
}

func buildSummary(v diag.Report, e scene.EngineReport) string {
	if !v.OK {
		return v.Agent.Summary
	}
	return fmt.Sprintf("Valid spec; visual score %d/100. %s", e.Score, scene.EngineNarrative(e))
}

func buildNextSteps(path string, v diag.Report, e scene.EngineReport) []string {
	if !v.OK {
		return v.Agent.NextSteps
	}
	steps := []string{
		"sceno describe -i " + quote(path) + " --json",
		"sceno render -i " + quote(path) + " -o output/sceno --all",
	}
	for _, f := range e.Findings {
		if f.Severity == "warning" && f.Fix != "" {
			steps = append([]string{f.Message + " — " + f.Fix}, steps...)
			break
		}
	}
	return steps
}

func dedupeRecs(in []diag.Recommendation) []diag.Recommendation {
	seen := map[string]struct{}{}
	var out []diag.Recommendation
	for _, r := range in {
		key := r.Category + "|" + r.Message
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, r)
	}
	return out
}

func quote(s string) string {
	if strings.ContainsAny(s, " \t") {
		return `"` + s + `"`
	}
	return s
}

func invokeAI(cmd string, report Report) (string, error) {
	prompt, err := buildAIPrompt(report)
	if err != nil {
		return "", err
	}
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty ai command")
	}
	c := exec.Command(parts[0], parts[1:]...)
	c.Stdin = strings.NewReader(prompt)
	out, err := c.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func buildAIPrompt(report Report) (string, error) {
	payload := map[string]any{
		"task":             "Review this Sceno diagram spec layout and suggest KDL edits for visual quality.",
		"validation_ok":    report.ValidationOK,
		"visual_score":     report.VisualScore,
		"engine_findings":  report.Engine.Findings,
		"recommendations":  report.Recommendations,
		"visual_rules":     report.VisualRules,
		"stack":            report.Stack,
	}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	sb.WriteString("You are a diagram and slide design expert. Given the JSON analysis below, suggest concrete KDL edits.\n")
	sb.WriteString("Focus on: hierarchy, whitespace, alignment, edge clarity, slide focus, infobox placement.\n")
	sb.WriteString("Respond with bullet points and short KDL examples.\n\n")
	sb.Write(b)
	return sb.String(), nil
}

// WriteJSON writes the advise report.
func (r Report) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

// WriteHuman writes a readable summary.
func (r Report) WriteHuman(w io.Writer) error {
	_, _ = io.WriteString(w, r.Agent.Summary+"\n\n")
	_, _ = fmt.Fprintf(w, "visual score: %d/100\n", r.VisualScore)
	_, _ = fmt.Fprintf(w, "stack planes: %v\n\n", r.Stack.Counts)
	if len(r.Recommendations) > 0 {
		_, _ = io.WriteString(w, "recommendations:\n")
		for _, rec := range r.Recommendations {
			_, _ = fmt.Fprintf(w, "  [%s] %s\n", rec.Category, rec.Message)
			if rec.Fix != "" {
				_, _ = fmt.Fprintf(w, "    fix: %s\n", rec.Fix)
			}
		}
	}
	if r.AIReview != "" {
		_, _ = io.WriteString(w, "\nai review:\n"+r.AIReview+"\n")
	}
	return nil
}
