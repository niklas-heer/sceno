package scene

import (
	"fmt"
	"math"

	"github.com/niklas-heer/sceno/internal/diag"
	"github.com/niklas-heer/sceno/internal/geom"
	"github.com/niklas-heer/sceno/internal/layout"
	"github.com/niklas-heer/sceno/internal/model"
)

// VisualRule documents a design principle the engine enforces.
type VisualRule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Source      string `json:"source,omitempty"`
	Description string `json:"description"`
}

// VisualRulesCatalog is the baked-in design knowledge (diagrams + slides).
var VisualRulesCatalog = []VisualRule{
	{ID: "hierarchy", Name: "Visual hierarchy", Source: "NN/g, IxDF", Description: "Titles and focal nodes should dominate; supporting detail recedes via size and spacing."},
	{ID: "whitespace", Name: "Whitespace", Source: "Gestalt proximity", Description: "Use gap and padding so groups breathe; avoid overcrowded or empty canvases."},
	{ID: "alignment", Name: "Alignment", Source: "PowerPoint grids", Description: "Same column/row nodes share center lines; icons and labels balance."},
	{ID: "edge_clarity", Name: "Edge clarity", Source: "d2/Mermaid", Description: "Connectors stay visible; route around nodes with fromSide/toSide."},
	{ID: "element_budget", Name: "Element budget", Source: "C4 / architecture", Description: "Prefer ≤15 primary nodes per view; split slides or add lanes for more."},
	{ID: "slide_focus", Name: "One idea per slide", Source: "10/20/30, Visme", Description: "Each slide should communicate one core idea with a clear focal point."},
	{ID: "annotations", Name: "Callouts & notes", Source: "PowerPoint", Description: "Use infobox/note/tip for context without blocking the main flow."},
	{ID: "collision_2d", Name: "2D collision plane", Source: "Sceno stack", Description: "Node plane overlaps are checked on a reduced 2D projection."},
	{ID: "routing_plane", Name: "Routing plane", Source: "Sceno stack", Description: "Edges are validated on the edge plane against node obstacles."},
}

// Finding is one rule outcome from the stack engine.
type Finding struct {
	RuleID    string    `json:"rule_id"`
	Severity  string    `json:"severity"` // error, warning, hint
	Plane     PlaneKind `json:"plane,omitempty"`
	Projected bool      `json:"projected_2d,omitempty"`
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Fix       string    `json:"fix,omitempty"`
	Example   string    `json:"example,omitempty"`
	Items     []string  `json:"items,omitempty"`
}

func (f Finding) ToIssue() diag.Issue {
	return diag.Issue{
		Code:    diag.Code(f.Code),
		Message: f.Message,
		Fix:     f.Fix,
		Example: f.Example,
		Nodes:   f.Items,
	}
}

type ruleContext struct {
	d      *model.Diagram
	stack  Stack
	scene  Report
}

type ruleFunc func(ruleContext) []Finding

var engineRules = []struct {
	id string
	fn ruleFunc
}{
	{"collision_2d", ruleCollisionPlane},
	{"routing_plane", ruleRoutingPlane},
	{"whitespace", ruleWhitespace},
	{"hierarchy", ruleHierarchy},
	{"element_budget", ruleElementBudget},
	{"slide_focus", ruleSlideFocus},
	{"annotations", ruleAnnotations},
	{"alignment", ruleAlignment},
	{"edge_clarity", ruleEdgeClarity},
}

func ruleCollisionPlane(ctx ruleContext) []Finding {
	margin := ctx.d.Gap / 2
	if margin < 8 {
		margin = 8
	}
	items := ctx.stack.Project(PlaneNode, PlaneAnnotation)
	var findings []Finding
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			a, b := items[i], items[j]
			if rectsOverlap(a.Bounds, b.Bounds, margin) {
				findings = append(findings, Finding{
					RuleID: "collision_2d", Severity: "warning", Plane: PlaneNode, Projected: true,
					Code: string(diag.CodeOccluded),
					Message: fmt.Sprintf("plane overlap: %q covers %q on node/annotation plane", a.Ref, b.Ref),
					Fix:     "Separate with at=col,row, increase gap, or move annotations to margins.",
					Items:   []string{a.Ref, b.Ref},
				})
			}
		}
	}
	return findings
}

func ruleRoutingPlane(ctx ruleContext) []Finding {
	var out []Finding
	for _, ec := range layout.FindEdgeCollisions(ctx.d) {
		sev := "warning"
		code := diag.CodeEdgeCollision
		if ec.Kind == "node_crossing" {
			sev = "error"
		}
		out = append(out, Finding{
			RuleID: "routing_plane", Severity: sev, Plane: PlaneEdge, Projected: true,
			Code:    string(code),
			Message: layout.FormatEdgeCollision(ec),
			Fix:     "Set fromSide/toSide, increase gap, or reorder layers so the connector routes around obstacles.",
			Example: `edge api -> db fromSide=right toSide=left`,
		})
	}
	return out
}

func ruleWhitespace(ctx ruleContext) []Finding {
	density := ctx.scene.Aesthetics.Density
	var out []Finding
	if density > 0.82 {
		out = append(out, Finding{
			RuleID: "whitespace", Severity: "warning", Plane: PlaneBackground,
			Code: string(diag.CodeDenseLayout),
			Message: fmt.Sprintf("layout is crowded (density %.2f) — hard to scan", density),
			Fix:     "Increase gap/padding, split into slides, or group nodes in lanes.",
		})
	}
	if density > 0 && density < 0.06 && len(ctx.d.Nodes) > 3 {
		out = append(out, Finding{
			RuleID: "whitespace", Severity: "hint", Plane: PlaneBackground,
			Code: string(diag.CodeSparseLayout),
			Message: fmt.Sprintf("layout is very sparse (density %.2f)", density),
			Fix:     "Reduce gap, tighten columns, or use a smaller canvas/slide frame.",
		})
	}
	return out
}

func ruleHierarchy(ctx ruleContext) []Finding {
	n := countPrimaryNodes(ctx.d)
	if n >= 6 && ctx.d.Title == "" {
		return []Finding{{
			RuleID: "hierarchy", Severity: "hint", Plane: PlaneChrome,
			Code: string(diag.CodeWeakHierarchy),
			Message: fmt.Sprintf("%d nodes without a diagram title — viewers lack a focal anchor", n),
			Fix:     "Add title=\"…\" and optional subtitle=\"…\" on the diagram block.",
			Example: `diagram title="Platform Overview" subtitle="Production path" gap=32 {`,
		}}
	}
	return nil
}

func ruleElementBudget(ctx ruleContext) []Finding {
	n := countPrimaryNodes(ctx.d)
	if n > 15 {
		return []Finding{{
			RuleID: "element_budget", Severity: "warning", Plane: PlaneNode,
			Code: string(diag.CodeTooManyElements),
			Message: fmt.Sprintf("%d primary nodes — architecture views read best with ≤15 elements", n),
			Fix:     "Split into multiple slides, add lanes to group detail, or extract a focused view.",
			Example: `slide "Overview" { /* ≤8 nodes */ }
slide "Detail" { /* next slice */ }`,
		}}
	}
	return nil
}

func ruleSlideFocus(ctx ruleContext) []Finding {
	if ctx.d.SlideAspect == "" {
		return nil
	}
	n := countPrimaryNodes(ctx.d)
	if n > 9 {
		return []Finding{{
			RuleID: "slide_focus", Severity: "warning", Plane: PlaneChrome,
			Code: string(diag.CodeSlideCrowded),
			Message: fmt.Sprintf("slide has %d nodes — one idea per slide works best with ≤9 shapes", n),
			Fix:     "Move supporting detail to another slide or use infobox for a single callout.",
		}}
	}
	return nil
}

func ruleAnnotations(ctx ruleContext) []Finding {
	var ann, flow int
	for _, n := range ctx.d.Nodes {
		switch model.NormalizeShape(n.Kind) {
		case model.ShapeInfobox, model.ShapeNote, model.ShapeTextbox:
			ann++
		case model.ShapeLane, model.ShapeFrame:
		default:
			flow++
		}
	}
	var out []Finding
	if flow >= 8 && ann == 0 {
		out = append(out, Finding{
			RuleID: "annotations", Severity: "hint", Plane: PlaneAnnotation,
			Code: string(diag.CodeSuggestAnnotation),
			Message: "complex diagram has no infobox/note — consider a callout for context",
			Fix:     "Add shape infobox key \"Note\" icon=info subtitle=\"…\" at=col,row or shape tip …",
			Example: `shape infobox legend "Legend" icon=info accent="#3b82f6" at=0,2`,
		})
	}
	for _, it := range ctx.stack.Planes[PlaneAnnotation.String()] {
		if blocksMainFlow(ctx, it) {
			out = append(out, Finding{
				RuleID: "annotations", Severity: "warning", Plane: PlaneAnnotation, Projected: true,
				Code: string(diag.CodeAnnotationBlocks),
				Message: fmt.Sprintf("annotation %q sits on the main left→right flow", it.Ref),
				Fix:     "Move infobox/note to top or bottom row (at=col,lastRow) or a margin column.",
				Items:   []string{it.Ref},
			})
		}
	}
	return out
}

func ruleAlignment(ctx ruleContext) []Finding {
	var out []Finding
	for _, a := range ctx.scene.Alignment {
		out = append(out, Finding{
			RuleID: "alignment", Severity: "warning", Plane: PlaneNode,
			Code: string(diag.CodeMisaligned),
			Message: a.Message,
			Fix:     "Use consistent at=col,row within columns; single-row pipelines center automatically.",
			Items:   a.Nodes,
		})
	}
	return out
}

func ruleEdgeClarity(ctx ruleContext) []Finding {
	var out []Finding
	for _, ev := range ctx.scene.EdgeVis {
		if ev.Visible >= 0.82 {
			continue
		}
		out = append(out, Finding{
			RuleID: "edge_clarity", Severity: "warning", Plane: PlaneEdge,
			Code: string(diag.CodeEdgeHidden),
			Message: fmt.Sprintf("edge %s→%s is ~%.0f%% visible", ev.From, ev.To, ev.Visible*100),
			Fix:     "Set fromSide/toSide, increase gap, or reroute around obstacles.",
			Items:   []string{ev.From, ev.To},
		})
	}
	return out
}

func countPrimaryNodes(d *model.Diagram) int {
	n := 0
	for _, node := range d.Nodes {
		k := model.NormalizeShape(node.Kind)
		if model.IsContainer(k) || k == model.ShapeInfobox || k == model.ShapeNote || k == model.ShapeTextbox {
			continue
		}
		n++
	}
	return n
}

func blocksMainFlow(ctx ruleContext, ann StackItem) bool {
	if ctx.stack.Canvas.W < 1 {
		return false
	}
	cx := ann.Bounds.CX()
	midL := ctx.stack.Canvas.X + ctx.stack.Canvas.W*0.25
	midR := ctx.stack.Canvas.X + ctx.stack.Canvas.W*0.75
	if cx < midL || cx > midR {
		return false
	}
	for _, re := range ctx.d.Routed {
		pts := pathToGeom(re.Points)
		for i := 1; i < len(pts); i++ {
			if math.Abs(pts[i-1].Y-pts[i].Y) < 2 && math.Abs(pts[i-1].Y-ann.Bounds.CY()) < ctx.d.Gap {
				if geom.SegmentHitsRect(pts[i-1], pts[i], ann.Bounds, ctx.d.Gap*0.2) {
					return true
				}
			}
		}
	}
	return false
}

func rectsOverlap(a, b model.Rect, gap float64) bool {
	return a.Right()+gap > b.X && b.Right()+gap > a.X &&
		a.Bottom()+gap > b.Y && b.Bottom()+gap > a.Y
}
