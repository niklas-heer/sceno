package diag

import "strings"

// AgentMeta is included in JSON validation output for repair loops.
type AgentMeta struct {
	Tool        string   `json:"tool"`
	Version     string   `json:"version,omitempty"`
	RenderReady bool     `json:"render_ready"`
	Summary     string   `json:"summary,omitempty"`
	NextSteps   []string `json:"next_steps,omitempty"`
	Hint        string   `json:"hint,omitempty"`
}

// CodeDoc describes one error code for agents.
type CodeDoc struct {
	Meaning string `json:"meaning"`
	Fix     string `json:"fix"`
	Example string `json:"example,omitempty"`
}

// ErrorCatalog maps machine codes to repair guidance.
var ErrorCatalog = map[Code]CodeDoc{
	CodeParse: {
		Meaning: "KDL syntax or structure is invalid.",
		Fix:     "Use only .kdl files. Check braces, quotes, and shape/edge statements.",
		Example: `diagram title="My Diagram" layout=auto gap=32 padding=24 {
  shape box start "Start" icon=server at=0,0
  shape box end "End" icon=server at=1,0
  edge start -> end fromSide=right toSide=left
}`,
	},
	CodeMissingNode: {
		Meaning: "An edge references a node id that does not exist in the same diagram or slide block.",
		Fix:     "Add `shape box ID \"Label\"` before `edge ID -> other`. Ids are case-sensitive.",
		Example: `shape box api "API" layer=1
shape box db "Database" layer=2
edge api -> db`,
	},
	CodeMissingPos: {
		Meaning: "layout=free requires explicit x and y on every shape.",
		Fix:     "Add x= and y= props, or switch to layout=auto and use layer/row/at.",
		Example: `shape box n "Node" x=100 y=80`,
	},
	CodeCollision: {
		Meaning: "Two nodes overlap after layout (even after auto nudge).",
		Fix:     "Increase diagram gap (e.g. gap=40), separate layer/row, or set at=col,row for each shape.",
		Example: `diagram gap=40 padding=28 layout=auto {
  shape box a "A" at=0,0
  shape box b "B" at=0,1
}`,
	},
	CodeEdgeCollision: {
		Meaning: "A connector crosses a node (routing obstacle).",
		Fix:     "Set fromSide/toSide on the edge, reorder layers, or increase gap.",
		Example: `edge api -> db fromSide=right toSide=left`,
	},
	CodeTextOverflow: {
		Meaning: "Label/subtitle does not fit the node box.",
		Fix:     "Remove fixed w/h, shorten text, use \\n for line breaks, or drop fontSize override.",
		Example: `shape box api "API Gateway" icon=api`,
	},
	CodeUnknownIcon: {
		Meaning: "icon= name is not in the catalog.",
		Fix:     "Run sceno icons or sceno guide --json for the allowed list.",
		Example: `shape box api "API" icon=api`,
	},
	CodeLayout: {
		Meaning: "Layout pipeline failed (often layout=free without coordinates).",
		Fix:     "Use layout=auto with layer/row/at, or provide x and y for every node.",
	},
	CodeSuggestCompact: {
		Meaning: "Layout is valid but sparse — optional improvement.",
		Fix:     "Lower gap, reuse layer numbers, or stack with row=.",
	},
	CodeOccluded: {
		Meaning: "A node drawn on top overlaps another (2D paint order / z-index).",
		Fix:     "Fix collisions, separate at=col,row, or increase gap so groups do not stack.",
		Example: `shape box a "A" at=0,0
shape box b "B" at=0,1`,
	},
	CodeEdgeHidden: {
		Meaning: "A connector runs behind nodes and is hard to see.",
		Fix:     "Set fromSide/toSide, increase gap, or use layout=auto with clearer layer columns.",
		Example: `edge a -> b fromSide=right toSide=left`,
	},
	CodeMisaligned: {
		Meaning: "Nodes or labels are not visually aligned within a column or icon row.",
		Fix:     "Use consistent layer/column, at=col,row, or shorten labels.",
		Example: `shape box api "API" icon=api at=1,0
shape box db "DB" icon=database at=1,1`,
	},
}

// Enrich fills summary, next_steps, render_ready, and per-issue examples.
func (r *Report) Enrich() {
	r.Agent = AgentMeta{
		Tool:        "sceno",
		RenderReady: r.OK,
		Hint:        "Run `sceno guide --json` for full spec, shapes, icons, and workflow.",
	}

	if r.OK {
		r.Agent.Summary = "Spec is valid; safe to render."
		r.Agent.NextSteps = []string{
			"sceno render -i " + quote(r.Input) + " -o output/sceno --all",
		}
		if r.Stats.Slides > 1 {
			r.Agent.NextSteps = append(r.Agent.NextSteps,
				"sceno render -i "+quote(r.Input)+" -o output/deck.slides.html -format slides")
		}
		return
	}

	var parts []string
	if len(r.Errors) > 0 {
		parts = append(parts, itoa(len(r.Errors))+" error(s)")
	}
	if len(r.Warnings) > 0 {
		parts = append(parts, itoa(len(r.Warnings))+" warning(s)")
	}
	r.Agent.Summary = strings.Join(parts, ", ") + " — fix errors before render."

	for i := range r.Errors {
		r.enrichIssue(&r.Errors[i], i+1)
	}
	for i := range r.Warnings {
		r.enrichIssue(&r.Warnings[i], i+1)
	}

	r.Agent.NextSteps = buildNextSteps(*r)
}

func (r *Report) enrichIssue(iss *Issue, n int) {
	if iss.Fix == "" {
		if doc, ok := ErrorCatalog[iss.Code]; ok {
			iss.Fix = doc.Fix
		}
	}
	if iss.Example == "" {
		if doc, ok := ErrorCatalog[iss.Code]; ok && doc.Example != "" {
			iss.Example = doc.Example
		}
	}
}

func buildNextSteps(r Report) []string {
	var steps []string
	for i, e := range r.Errors {
		step := "Fix error " + itoa(i+1) + " (" + string(e.Code) + ")"
		if e.Message != "" {
			step += ": " + e.Message
		}
		if e.Fix != "" {
			step += " — " + e.Fix
		}
		steps = append(steps, step)
	}
	steps = append(steps, "Run: sceno validate -i "+quote(r.Input)+" --json")
	steps = append(steps, "When ok is true: sceno render -i "+quote(r.Input)+" -o output/sceno --all")
	return steps
}

func quote(s string) string {
	if s == "" {
		return "sceno.kdl"
	}
	if strings.ContainsAny(s, " \t") {
		return `"` + s + `"`
	}
	return s
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
