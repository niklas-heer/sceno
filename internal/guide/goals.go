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
		Mission: "Bridge text and visual work for humans and AI agents: one deterministic KDL file produces both polished exports and an exact machine-readable account of the canvas. Agents must be able to understand what exists, where it is, what conflicts, and how to repair it in a validate → advise → describe → render loop — locally, without cloud lock-in or export surprises.",
		ProductGoals: []string{
			"**KDL-only specs** — One human-friendly format; no YAML/JSON/Lua drift.",
			"**Agent-readable visual state** — Exact canvas, shape content, routes, stack planes, collisions, and repair candidates without requiring image vision.",
			"**Quantified spacing** — Report actual node clearance, content padding, parent insets, and canvas margins in pixels, with explicit bounds-based semantics and intentional-overlap status.",
			"**Export parity** — SVG, PNG, PDF, HTML, and slides must look like the same diagram.",
			"**Shapes fit text** — Visible silhouettes and internal lines define writable space; Inter metrics select the largest readable font and labels never clip.",
			"**Trustworthy layout** — Collision-free auto grid, obstacle-aware routing, readable labels, attached arrowheads, and deterministic results.",
			"**Controlled visual freedom** — Hybrid/free placement, dx/dy nudges, annotations, and intentional overlap without hiding geometry from feedback.",
			"**Stack scene understanding** — Stacked 2D planes; paint order, occlusion, edge visibility via describe/advise/validate.",
			"**Visual design validation** — Measurable whitespace, hierarchy, content grid, connector clearance, element budget, slide focus, and annotation placement.",
			"**Local editing experience** — Live preview, clickable bounds and source locations, draft-safe source edits, verified repair review with Apply/Undo, and exports from the same evaluated geometry.",
			"**AI-ready feedback** — JSON everywhere; exact geometry plus fix/example/repair hints; optional advise --ai with SCENO_AI_CMD.",
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
			"Collision-free defaults — auto layout routes around shapes and existing connectors",
			"Visible connectors — attached arrowheads, direct paths, and clear labeled shafts",
			"Aligned content — inline icon/text groups and row/column alignment checks",
			"Readable density — aesthetic score; dense_layout / sparse_layout hints",
			"Annotations — info/tip/warning/infobox for context off the main flow",
		},
		AgentWorkflow: []string{
			"sceno docs guide --json once per session",
			"Edit KDL → sceno validate --json until ok: true",
			"sceno advise --json; resolve every structural visual finding",
			"sceno describe --json; inspect slides[n].engine.scene_stack outlines, internal lines, writable/content bounds, effective font sizes, routes, repairs, and ascii_map",
			"sceno render --all; visually sample each target format",
			"Repository changes: mask verify; every shipped example scores at least 80 with no blocked visual code",
		},
		Conventions: []string{
			"Quote strings with spaces: title=\"My Platform\"",
			"Use \\n in labels for line breaks",
			"Define shapes before edges in the same block",
			"Prefer layout=auto with layer/row/at and dx/dy; use hybrid + x/y for breakout elements",
			"Use semantic callouts: shape info, tip, warning, infobox, note",
		},
		NonGoals: []string{
			"Real-time collaborative editing",
			"WYSIWYG drag-and-drop canvas (free placement via x/y is supported)",
			"Import from Visio/Lucidchart",
			"Animation timelines inside slides",
			"Embedded general-purpose scripting — deterministic declarative constraints keep feedback explainable",
			"Built-in LLM — use sceno advise --ai with your CLI instead",
		},
		QualityBar: []QualityEntry{
			{Area: "Typography", Target: "Embedded Inter (OFL), measured widths, silhouette-safe writable bounds, automatic font fitting with a 10px floor"},
			{Area: "Icons", Target: "Embedded vector catalog; parity across SVG/PNG/PDF/HTML/slides"},
			{Area: "Arrows", Target: "Obstacle-free orthogonal routes; 18px visible shaft + 9px head; tips on borders"},
			{Area: "Scene", Target: "One shared geometry model for stack planes, describe, advise, validate, and render"},
			{Area: "Feedback", Target: "Deterministic measurements with units and bounds basis; slide-scoped geometry and repairs; explicit truncation rather than silently incomplete spacing reports"},
			{Area: "Corpus", Target: "Every shipped KDL scores ≥80 with no collision, detour, hidden/detached arrow, label overlap, side mismatch, occlusion, or text overflow"},
			{Area: "Slides", Target: "slide blocks; ≤9 shapes per slide (hint)"},
			{Area: "Preview", Target: "Invalid source never renders; proposals require current evidence, preserve source formatting, and must resolve their target without adding or worsening structural findings; external edits invalidate pending writes"},
			{Area: "CLI", Target: "Single binary; JSON everywhere; self-doc from code"},
		},
		Principles: []string{
			"Spec is source of truth — diagram is computed, not hand-tweaked per format",
			"Feedback and pixels share geometry — an agent sees the same canvas the renderer exports",
			"Fail with advice — errors include fix and example KDL",
			"See without pixels — describe/advise/stack when agents cannot view PNG",
			"Self-documenting — sceno docs generated from code, not duplicate markdown",
		},
	}
}
