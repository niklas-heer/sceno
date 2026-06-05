package guide

import (
	"fmt"
	"sort"
	"strings"

	"github.com/niklas-heer/sceno/internal/diag"
	"github.com/niklas-heer/sceno/internal/scene"
)

// RenderSpecMarkdown generates the KDL spec from code (single source of truth).
func RenderSpecMarkdown() string {
	d := Build()
	var b strings.Builder
	b.WriteString("# Sceno — KDL specification\n\n")
	b.WriteString("Diagram is defined in **[KDL](https://kdl.dev/)** only. Generated from code — run `sceno docs spec`.\n\n")
	b.WriteString("## Quick start\n\n```kdl\n")
	b.WriteString(SpecExample())
	b.WriteString("\n```\n\n```bash\nsceno init -o platform.kdl\nsceno validate -i platform.kdl --json\nsceno advise -i platform.kdl --json\nsceno render -i platform.kdl -o out --all\n```\n\n")

	b.WriteString("## Document structure\n\n")
	b.WriteString("| Statement | Example |\n|-----------|---------|\n")
	b.WriteString("| Diagram block | `diagram title=\"...\" layout=auto { ... }` |\n")
	b.WriteString("| Shape | `shape box id \"Label\" layer=1` |\n")
	b.WriteString("| Edge | `edge from -> to` |\n")
	b.WriteString("| Slide | `slide \"Title\" { ... }` |\n\n")

	b.WriteString("## Shapes\n\nSyntax: **`shape KIND ID \"Label\" props...`** — KIND defaults to `box`.\n\n")
	b.WriteString("| Kind | Aliases | Use |\n|------|---------|-----|\n")
	for _, e := range ShapeCatalog() {
		aliases := strings.Join(e.Aliases, ", ")
		if aliases == "" {
			aliases = "—"
		}
		fmt.Fprintf(&b, "| `%s` | %s | %s |\n", e.Kind, aliases, e.Use)
	}
	b.WriteString("\n## Node properties\n\n| Property | Description |\n|----------|-------------|\n")
	keys := sortedKeys(d.ShapeProps)
	for _, k := range keys {
		fmt.Fprintf(&b, "| `%s` | %s |\n", k, d.ShapeProps[k])
	}
	b.WriteString("\n### Icon placement (`iconPos`)\n\n")
	for _, o := range IconPosOptions() {
		b.WriteString("- `" + strings.Split(o, " ")[0] + "` — " + o + "\n")
	}
	b.WriteString("\nLabels support `\\n` for line breaks in quoted strings.\n\n")

	b.WriteString("## Diagram properties\n\n| Property | Description |\n|----------|-------------|\n")
	for _, k := range sortedKeys(d.DiagramProps) {
		fmt.Fprintf(&b, "| `%s` | %s |\n", k, d.DiagramProps[k])
	}
	b.WriteString("\n## Edges\n\n| Property | Values |\n|----------|--------|\n")
	for _, k := range sortedKeys(d.EdgeProps) {
		fmt.Fprintf(&b, "| `%s` | %s |\n", k, d.EdgeProps[k])
	}
	b.WriteString("\nEdge labels render above horizontal segments and to the right of vertical segments.\n\n")

	b.WriteString("## Layout\n\n")
	b.WriteString("- `layout auto` — grid by `layer` / `row` / `at=col,row` (default)\n")
	b.WriteString("- `layout free` — every shape needs `x` and `y`\n")
	b.WriteString("- Single-row diagrams vertically center nodes for straight horizontal connectors\n\n")

	b.WriteString("## Stack validation\n\n")
	b.WriteString(scene.StackModelDescription() + "\n\n")
	b.WriteString("```bash\nsceno advise -i file.kdl --json\nsceno describe -i file.kdl --json\nsceno validate -i file.kdl --json\n```\n\n")

	b.WriteString("## Slides & theme\n\n")
	b.WriteString("Set `slide=16x9`, `theme=dark`, `background=transparent` on the diagram line.\n\n")
	b.WriteString("Code languages: " + strings.Join(CodeLanguages(), ", ") + "\n\n")

	b.WriteString("## Validation codes\n\n| Code | Severity |\n|------|----------|\n")
	for _, c := range sortedErrorCodes() {
		sev := errorSeverity(c)
		fmt.Fprintf(&b, "| `%s` | %s |\n", c, sev)
	}
	b.WriteString("\nSee `sceno docs errors --json` for fix and example on each code.\n")
	return b.String()
}

// RenderGoalsMarkdown generates goals from code.
func RenderGoalsMarkdown() string {
	g := BuildGoals()
	var b strings.Builder
	b.WriteString("# Sceno — goals\n\n")
	b.WriteString("## Mission\n\n")
	b.WriteString(g.Mission + "\n\n")
	b.WriteString("## Product goals\n\n")
	for i, goal := range g.ProductGoals {
		fmt.Fprintf(&b, "%d. %s\n", i+1, goal)
	}
	b.WriteString("\n## Ecosystem\n\n| Tool | What we take |\n|------|-------------|\n")
	for _, e := range g.Ecosystem {
		fmt.Fprintf(&b, "| **%s** | %s |\n", e.Tool, e.Takes)
	}
	b.WriteString("\n### Layout & readability\n\n")
	for _, r := range g.LayoutRules {
		b.WriteString("- " + r + "\n")
	}
	b.WriteString("\n### Agent workflow\n\n")
	for i, step := range g.AgentWorkflow {
		fmt.Fprintf(&b, "%d. %s\n", i+1, step)
	}
	b.WriteString("\n### Authoring conventions\n\n")
	for _, c := range g.Conventions {
		b.WriteString("- " + c + "\n")
	}
	b.WriteString("\n## Non-goals (for now)\n\n")
	for _, n := range g.NonGoals {
		b.WriteString("- " + n + "\n")
	}
	b.WriteString("\n## Quality bar\n\n| Area | Target |\n|------|--------|\n")
	for _, q := range g.QualityBar {
		fmt.Fprintf(&b, "| %s | %s |\n", q.Area, q.Target)
	}
	b.WriteString("\n## Principles\n\n")
	for _, p := range g.Principles {
		b.WriteString("- **" + strings.SplitN(p, " — ", 2)[0] + "**")
		if parts := strings.SplitN(p, " — ", 2); len(parts) == 2 {
			b.WriteString(" — " + parts[1])
		}
		b.WriteString("\n")
	}
	return b.String()
}

// RenderStackMarkdown generates stack documentation from code.
func RenderStackMarkdown() string {
	d := Build()
	var b strings.Builder
	b.WriteString("# Sceno — stack validation model\n\n")
	b.WriteString(d.StackModel + "\n\n")
	b.WriteString("## Plane order (back → front)\n\n")
	b.WriteString("| Plane | Contents | Purpose |\n|-------|----------|--------|\n")
	for _, p := range scene.PlaneCatalog() {
		fmt.Fprintf(&b, "| `%s` | %s | %s |\n", p.Name, p.Contents, p.Purpose)
	}
	b.WriteString("\n## Visual design rules\n\n")
	for _, r := range scene.VisualRulesCatalog {
		src := r.Source
		if src != "" {
			src = " (" + src + ")"
		}
		fmt.Fprintf(&b, "- **%s** (`%s`)%s — %s\n", r.Name, r.ID, src, r.Description)
	}
	b.WriteString("\n## Commands\n\n")
	b.WriteString("```bash\nsceno advise -i file.kdl --json\nsceno describe -i file.kdl --json\nsceno validate -i file.kdl --json\nexport SCENO_AI_CMD=\"codex exec -\"\nsceno advise -i file.kdl --ai\n```\n")
	return b.String()
}

// RenderErrorsMarkdown generates error catalog from diag.ErrorCatalog.
func RenderErrorsMarkdown() string {
	var b strings.Builder
	b.WriteString("# Sceno — error codes\n\n")
	for _, c := range sortedErrorCodes() {
		doc := diag.ErrorCatalog[diag.Code(c)]
		fmt.Fprintf(&b, "## `%s`\n\n", c)
		b.WriteString(doc.Meaning + "\n\n")
		b.WriteString("**Fix:** " + doc.Fix + "\n\n")
		if doc.Example != "" {
			b.WriteString("```kdl\n" + doc.Example + "\n```\n\n")
		}
	}
	return b.String()
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedErrorCodes() []string {
	codes := make([]string, 0, len(diag.ErrorCatalog))
	for c := range diag.ErrorCatalog {
		codes = append(codes, string(c))
	}
	sort.Strings(codes)
	return codes
}

func errorSeverity(code string) string {
	switch diag.Code(code) {
	case diag.CodeParse, diag.CodeMissingNode, diag.CodeMissingPos,
		diag.CodeCollision, diag.CodeEdgeCollision, diag.CodeLayout,
		diag.CodeTextOverflow, diag.CodeUnknownIcon:
		return "error"
	case diag.CodeDenseLayout, diag.CodeSlideCrowded, diag.CodeTooManyElements,
		diag.CodeAnnotationBlocks, diag.CodeOccluded, diag.CodeEdgeHidden, diag.CodeMisaligned:
		return "warning"
	default:
		return "hint"
	}
}
