package scene

import (
	"fmt"
	"math"
	"strings"

	"github.com/niklas-heer/sceno/internal/diag"
	"github.com/niklas-heer/sceno/internal/fonts"
	"github.com/niklas-heer/sceno/internal/geom"
	"github.com/niklas-heer/sceno/internal/measure"
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/theme"
)

const (
	minVisibleArrowStroke = 18.0
	maxLabelAxisDrift     = 6.0
	anchorEps             = 1.5
)

// edgeRenderFindings validates arrowheads and edge labels using the same layout as render.
func edgeRenderFindings(d *model.Diagram) []Finding {
	var out []Finding
	for _, re := range d.Routed {
		out = append(out, checkEdgeArrow(d, re)...)
		out = append(out, checkEdgeLabel(d, re)...)
	}
	return out
}

func checkEdgeArrow(d *model.Diagram, re model.RoutedEdge) []Finding {
	gpts := geom.SimplifyPath(geom.SlicesToPath(re.Points))
	if len(gpts) < 2 {
		return nil
	}
	byID := map[string]model.Node{}
	for _, n := range d.Nodes {
		byID[n.ID] = n
	}
	srcAnchor, dstAnchor, ok := edgeAnchors(re.Edge, byID)
	if !ok {
		return nil
	}

	var out []Finding
	key := re.Edge.From + "→" + re.Edge.To
	pathStart := gpts[0]
	pathEnd := gpts[len(gpts)-1]

	if geom.TipGap(pathStart, srcAnchor) > anchorEps {
		out = append(out, Finding{
			RuleID: "edge_clarity", Severity: "error", Plane: PlaneEdge,
			Code: string(diag.CodeArrowDetached),
			Message: fmt.Sprintf("edge %s starts %.0fpx from %q border (should attach at anchor)", key, geom.TipGap(pathStart, srcAnchor), re.Edge.From),
			Fix:     "Set fromSide/toSide so routing ends on shape borders.",
			Items:   []string{re.Edge.From, re.Edge.To},
		})
	}
	if geom.TipGap(pathEnd, dstAnchor) > anchorEps {
		out = append(out, Finding{
			RuleID: "edge_clarity", Severity: "error", Plane: PlaneEdge,
			Code: string(diag.CodeArrowDetached),
			Message: fmt.Sprintf("edge %s ends %.0fpx from %q border — arrow cannot meet the shape", key, geom.TipGap(pathEnd, dstAnchor), re.Edge.To),
			Fix:     "Route must terminate on the target border; increase gap or fix fromSide/toSide.",
			Items:   []string{re.Edge.From, re.Edge.To},
		})
	}

	ag, ok := geom.ArrowGeometryForPath(gpts)
	if !ok {
		return out
	}
	if geom.TipGap(ag.Tip, dstAnchor) > geom.MaxArrowTipGap {
		out = append(out, Finding{
			RuleID: "edge_clarity", Severity: "error", Plane: PlaneEdge,
			Code: string(diag.CodeArrowDetached),
			Message: fmt.Sprintf("edge %s arrow tip is %.0fpx from %q border", key, geom.TipGap(ag.Tip, dstAnchor), re.Edge.To),
			Fix:     "Arrow tip must land on the target border (render draws heads after nodes).",
			Items:   []string{re.Edge.From, re.Edge.To},
		})
	}

	strokeLen := math.Hypot(ag.StrokeEnd.X-ag.Prev.X, ag.StrokeEnd.Y-ag.Prev.Y)
	if strokeLen < minVisibleArrowStroke {
		out = append(out, Finding{
			RuleID: "edge_clarity", Severity: "warning", Plane: PlaneEdge,
			Code: string(diag.CodeArrowDetached),
			Message: fmt.Sprintf("edge %s has only %.0fpx of stroke before arrowhead — head may look floating", key, strokeLen),
			Fix:     "Increase gap between nodes or shorten edge labels.",
			Items:   []string{re.Edge.From, re.Edge.To},
		})
	}

	if geom.TipGap(pathEnd, dstAnchor) > anchorEps {
		if dst, ok := byID[re.Edge.To]; ok && arrowTipBuried(ag.Tip, dst) {
			out = append(out, Finding{
				RuleID: "edge_clarity", Severity: "warning", Plane: PlaneEdge,
				Code: string(diag.CodeArrowHidden),
				Message: fmt.Sprintf("edge %s arrow tip on %q sits inside the shape (%.0fpx off border)", key, re.Edge.To, geom.TipGap(ag.Tip, dstAnchor)),
				Fix:     "Routing must end on the target border anchor; check fromSide/toSide and re-run validate.",
				Items:   []string{re.Edge.From, re.Edge.To},
			})
		}
	}

	return out
}

func edgeAnchors(e model.Edge, byID map[string]model.Node) (src, dst geom.Point, ok bool) {
	a, okA := byID[e.From]
	b, okB := byID[e.To]
	if !okA || !okB {
		return geom.Point{}, geom.Point{}, false
	}
	fs, ts := e.FromSide, e.ToSide
	if fs == "" || fs == model.SideAuto {
		fs, _ = geom.BestSides(a, b)
	}
	if ts == "" || ts == model.SideAuto {
		_, ts = geom.BestSides(a, b)
	}
	return geom.Anchor(a, fs), geom.Anchor(b, ts), true
}

// arrowTipBuried is true when the tip sits more than 2px inside the node interior (not on border).
func arrowTipBuried(tip geom.Point, n model.Node) bool {
	r := n.Rect
	pad := 2.0
	inside := tip.X > r.X+pad && tip.X < r.Right()-pad &&
		tip.Y > r.Y+pad && tip.Y < r.Bottom()-pad
	return inside
}

func checkEdgeLabel(d *model.Diagram, re model.RoutedEdge) []Finding {
	label := strings.TrimSpace(re.Edge.Label)
	if label == "" {
		return nil
	}
	gpts := geom.SimplifyPath(geom.SlicesToPath(re.Points))
	if len(gpts) < 2 {
		return nil
	}
	lctx := edgeLabelContext(d, re.Edge)
	layout := geom.LayoutEdgeLabel(gpts, label, lctx)
	if layout.BoxW <= 0 {
		return nil
	}
	lx, ly, lw, lh := layout.LabelRect()
	lb := model.Rect{X: lx, Y: ly, W: lw, H: lh}
	key := re.Edge.From + "→" + re.Edge.To

	var out []Finding
	_, pathY, horiz := geom.LabelPlacement(gpts)
	if horiz && math.Abs(layout.CenterY-pathY) > maxLabelAxisDrift {
		out = append(out, Finding{
			RuleID: "edge_clarity", Severity: "warning", Plane: PlaneLabel,
			Code: string(diag.CodeEdgeLabelOffAxis),
			Message: fmt.Sprintf("edge label %q on %s sits %.0fpx off the connector — use on-line placement", label, key, math.Abs(layout.CenterY-pathY)),
			Fix:     "Labels should center on the connector in the node gap (opaque box over the stroke).",
			Items:   []string{re.Edge.From, re.Edge.To},
		})
	}

	if chrome, ok := diagramChromeBand(d); ok && rectsOverlap(lb, chrome, 2) {
		out = append(out, Finding{
			RuleID: "edge_clarity", Severity: "warning", Plane: PlaneChrome,
			Code: string(diag.CodeEdgeLabelChrome),
			Message: fmt.Sprintf("edge label %q on %s overlaps title/subtitle chrome", label, key),
			Fix:     "Move labels onto connectors between nodes, not above the diagram header band.",
			Items:   []string{re.Edge.From, re.Edge.To},
		})
	}

	if lctx != nil {
		gap := lctx.To.X - lctx.From.Right()
		if layout.BoxW > gap-8 {
			out = append(out, Finding{
				RuleID: "edge_clarity", Severity: "warning", Plane: PlaneLabel,
				Code: string(diag.CodeEdgeLabelChrome),
				Message: fmt.Sprintf("edge label %q on %s is wider (%.0fpx) than the node gap (%.0fpx)", label, key, layout.BoxW, gap),
				Fix:     "Shorten the label, increase diagram gap, or remove the label from tight pipelines.",
				Items:   []string{re.Edge.From, re.Edge.To},
			})
		}
	}
	return out
}

func edgeLabelContext(d *model.Diagram, e model.Edge) *geom.EdgeLabelContext {
	byID := map[string]model.Node{}
	for _, n := range d.Nodes {
		byID[n.ID] = n
	}
	a, okA := byID[e.From]
	b, okB := byID[e.To]
	if !okA || !okB {
		return nil
	}
	return &geom.EdgeLabelContext{From: a.Rect, To: b.Rect}
}

// diagramChromeBand approximates the title/subtitle region (matches render.Bounds).
func diagramChromeBand(d *model.Diagram) (model.Rect, bool) {
	if d.Title == "" && d.Subtitle == "" {
		return model.Rect{}, false
	}
	minX, minY := 1e9, 1e9
	for _, n := range d.Nodes {
		if n.Rect.X < minX {
			minX = n.Rect.X
		}
		if n.Rect.Y < minY {
			minY = n.Rect.Y
		}
	}
	if minX > 1e8 {
		return model.Rect{}, false
	}
	pad := d.Padding + 48
	bandX := minX - pad + 28
	bandTop := minY - pad + float64(theme.TitleSize)*0.25
	bandH := float64(theme.TitleSize) + 8
	if d.Title != "" {
		tw := measure.TextWidth(d.Title, theme.TitleSize, fonts.WeightBold)
		if tw < 120 {
			tw = 120
		}
		if d.Subtitle != "" {
			sw := measure.TextWidth(d.Subtitle, theme.SubtitleSize, fonts.WeightRegular)
			if sw > tw {
				tw = sw
			}
			bandH = 62 - 32 + float64(theme.SubtitleSize) + float64(theme.TitleSize)
		}
		return model.Rect{X: bandX, Y: bandTop, W: tw + 16, H: bandH + 12}, true
	}
	return model.Rect{X: bandX, Y: bandTop, W: 200, H: bandH}, true
}
