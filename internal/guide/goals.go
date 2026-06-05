package guide

// GoalsDocument is product mission and quality bar (source of truth for sceno goals).
type GoalsDocument struct {
	Mission       string
	ProductGoals  []string
	Ecosystem     []EcosystemEntry
	LayoutRules   []string
	AgentWorkflow []string
	Conventions   []string
	NonGoals      []string
	QualityBar    []QualityEntry
	Principles    []string
}

type EcosystemEntry struct {
	Tool  string
	Takes string
}

type QualityEntry struct {
	Area   string
	Target string
}

// BuildGoals returns the goals document from code.
func BuildGoals() GoalsDocument {
	return GoalsDocument{
		Mission: "Build the best **local-first** tool for architecture and system diagrams: one KDL file in, polished diagrams out — without a browser editor, without cloud lock-in, without export surprises. Optimized for humans **and** AI agents that edit specs in a validate → advise → describe → render loop.",
		ProductGoals: []string{
			"**KDL-only specs** — One human-friendly format; no YAML/JSON/Lua drift.",
			"**Export parity** — SVG, PNG, PDF, HTML, and slides must look like the same diagram.",
			"**Boxes fit text** — Node size driven by Inter metrics; labels never clip.",
			"**Trustworthy layout** — Auto grid, obstacle-aware routing, single-row centering, collision resolution.",
			"**Stack scene understanding** — Stacked 2D planes; paint order, occlusion, edge visibility via describe/advise/validate.",
			"**Visual design validation** — Whitespace, hierarchy, C4 element budget, slide focus, annotation placement.",
			"**AI-ready validation** — JSON everywhere; fix + example on errors; optional advise --ai with SCENO_AI_CMD.",
			"**PowerPoint familiarity** — Shapes, lanes, callouts (info/tip/warning), iconPos, dashed policy lines.",
			"**Slide-ready export** — slide blocks, 16x9, HTML deck, per-slide PNG/SVG.",
			"**Theming** — theme=dark, background=transparent, var.* overrides.",
			"**Code on slides** — Syntax-highlighted code blocks in slides HTML and SVG.",
		},
		Ecosystem: []EcosystemEntry{
			{Tool: "d2", Takes: "Declarative source of truth; themes; validate before export"},
			{Tool: "Mermaid", Takes: "Text-first diagrams; familiar edges; dark/light themes"},
			{Tool: "Excalidraw", Takes: "Sketch aesthetic (style=sketch); organic connectors"},
			{Tool: "PlantUML", Takes: "Precise architecture layout; code in decks"},
			{Tool: "Structurizr", Takes: "Consistent notation; clear layers; ≤15 elements per view"},
			{Tool: "PowerPoint / Keynote", Takes: "Slide titles, 16:9, callouts, one idea per slide"},
			{Tool: "Figma / shadcn", Takes: "Design tokens, subtle borders, dark mode"},
		},
		LayoutRules: []string{
			"Logical grouping — columns/layers and proximity clusters",
			"Visible connectors — fromSide/toSide when paths cross nodes",
			"Aligned labels — iconPos + column alignment checks",
			"Readable density — aesthetic score; dense_layout / sparse_layout hints",
			"Annotations — info/tip/warning/infobox for context off the main flow",
		},
		AgentWorkflow: []string{
			"sceno docs guide --json once per session",
			"Edit KDL → sceno validate --json until ok: true",
			"sceno advise --json for visual polish",
			"sceno describe --json to sanity-check layout",
			"sceno render --all",
		},
		Conventions: []string{
			"Quote strings with spaces: title=\"My Platform\"",
			"Use \\n in labels for line breaks",
			"Define shapes before edges in the same block",
			"Prefer layout=auto with layer/row/at; layout=free + x/y for free placement",
			"Use semantic callouts: shape info, tip, warning, infobox, note",
		},
		NonGoals: []string{
			"Real-time collaborative editing",
			"WYSIWYG drag-and-drop canvas (free placement via x/y is supported)",
			"Import from Visio/Lucidchart",
			"Animation timelines inside slides",
			"Built-in LLM — use sceno advise --ai with your CLI instead",
		},
		QualityBar: []QualityEntry{
			{Area: "Typography", Target: "Embedded Inter (OFL), measured widths"},
			{Area: "Icons", Target: "Crisp SVG/PNG; iconPos placement"},
			{Area: "Arrows", Target: "Orthogonal (polished); labels on H/V segments"},
			{Area: "Scene", Target: "Stack planes; describe + advise + validate"},
			{Area: "Slides", Target: "slide blocks; ≤9 shapes per slide (hint)"},
			{Area: "CLI", Target: "Single binary; JSON everywhere; self-doc from code"},
		},
		Principles: []string{
			"Spec is source of truth — diagram is computed, not hand-tweaked per format",
			"Fail with advice — errors include fix and example KDL",
			"See without pixels — describe/advise/stack when agents cannot view PNG",
			"Self-documenting — sceno docs generated from code, not duplicate markdown",
		},
	}
}
