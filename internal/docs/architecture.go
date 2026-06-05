package docs

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/niklas-heer/sceno/internal/scene"
	"github.com/niklas-heer/sceno/internal/version"
)

// ArchitectureDoc documents the source-of-truth chain for agents.
type ArchitectureDoc struct {
	Tool           string            `json:"tool"`
	Version        string            `json:"version"`
	Summary        string            `json:"summary"`
	Pipeline       []string          `json:"pipeline"`
	GeometrySoT    string            `json:"geometry_source_of_truth"`
	SemanticsSoT   string            `json:"semantics_source_of_truth"`
	EntryPoint     string            `json:"entry_point"`
	Consumers      map[string]string `json:"consumers"`
	PaintOrder     string            `json:"paint_order"`
	StackModel     string            `json:"stack_model"`
	Principles     []string          `json:"principles"`
	AntiPatterns   []string          `json:"anti_patterns"`
}

func buildArchitectureDoc() ArchitectureDoc {
	return ArchitectureDoc{
		Tool:    "sceno",
		Version: version.Version,
		Summary: "Sceno separates geometry (where things are) from semantics (whether the diagram reads well). One pipeline builds both; every command consumes the same artifact.",
		Pipeline: []string{
			"KDL file",
			"spec.LoadFile + spec.Validate — syntax and referential integrity",
			"pipeline.BuildDeck — geometry: node positions, sizes, routed edges",
			"scene.Evaluate per slide — semantics: stack planes, visual rules, score, paint order",
			"validate / advise / describe / render — consumers of pipeline.Result",
		},
		GeometrySoT:  "model.Diagram from pipeline.BuildDeck (positions, rects, Routed edges)",
		SemanticsSoT: "scene.Evaluation from scene.Evaluate (stack, findings, visual_score, paint_order)",
		EntryPoint:   "validate.LoadAndEvaluate(path, opt) → (pipeline.Result, diag.Report, error)",
		Consumers: map[string]string{
			"sceno validate": "diag.Report from ApplyResult — blocking errors + stack warnings",
			"sceno advise":   "pipeline.Result.MergedEval().EngineReport() — visual score + findings",
			"sceno describe": "pipeline.Result.Slides[n].Eval — scene + engine per slide",
			"sceno render":   "pipeline.Result.Deck — pixels; paint order follows Evaluation.PaintOrder",
		},
		PaintOrder: scene.PaintOrderDescription(),
		StackModel: "background → lanes → edges → structure → annotations → nodes → labels → chrome",
		Principles: []string{
			"Build once per command — no double pipeline.BuildDeck in validate then render",
			"Render projects geometry; it must not re-derive validation rules",
			"Visual rules live in scene.engineRules — single catalog for validate, advise, docs",
			"Backward compat: scene.RunEngine and scene.Analyze wrap scene.Evaluate",
		},
		AntiPatterns: []string{
			"Calling pipeline.BuildDeck separately after validate.Run in the same command",
			"Duplicating overlap or edge-crossing checks outside scene.Evaluate",
			"Inventing paint order in render that contradicts engine stack planes",
		},
	}
}

func writeArchitectureJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(buildArchitectureDoc())
}

func writeArchitectureMarkdown(w io.Writer) error {
	doc := buildArchitectureDoc()
	var b strings.Builder
	b.WriteString("# Sceno — architecture\n\n")
	b.WriteString(doc.Summary + "\n\n")
	b.WriteString("## Pipeline\n\n")
	for i, step := range doc.Pipeline {
		fmt.Fprintf(&b, "%d. %s\n", i+1, step)
	}
	b.WriteString("\n## Source of truth\n\n")
	fmt.Fprintf(&b, "- **Geometry:** %s\n", doc.GeometrySoT)
	fmt.Fprintf(&b, "- **Semantics:** %s\n\n", doc.SemanticsSoT)
	b.WriteString("## Entry point\n\n")
	b.WriteString("`" + doc.EntryPoint + "`\n\n")
	b.WriteString("## Consumers\n\n")
	for cmd, role := range doc.Consumers {
		fmt.Fprintf(&b, "- `%s` — %s\n", cmd, role)
	}
	b.WriteString("\n## Paint order (render contract)\n\n")
	b.WriteString(doc.PaintOrder + "\n\n")
	b.WriteString("## Stack model\n\n")
	b.WriteString(doc.StackModel + "\n\n")
	b.WriteString("## Principles\n\n")
	for _, p := range doc.Principles {
		b.WriteString("- " + p + "\n")
	}
	b.WriteString("\n## Anti-patterns\n\n")
	for _, a := range doc.AntiPatterns {
		b.WriteString("- " + a + "\n")
	}
	_, err := io.WriteString(w, b.String())
	return err
}
