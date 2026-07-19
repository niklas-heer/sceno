package scene

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/niklas-heer/sceno/internal/composition"
	"github.com/niklas-heer/sceno/internal/diag"
	"github.com/niklas-heer/sceno/internal/geom"
	"github.com/niklas-heer/sceno/internal/measure"
	"github.com/niklas-heer/sceno/internal/model"
)

const (
	minVisibleArrowStroke = geom.EdgeLabelClearRun
	maxLabelAxisDrift     = 6.0
	anchorEps             = 1.5
	minArrowTipSeparation = 14.0
)

// edgeRenderFindings validates arrowheads and edge labels using the same layout as render.
func edgeRenderFindings(d *model.Diagram) []Finding {
	var out []Finding
	for _, re := range d.Routed {
		out = append(out, checkEdgeArrow(d, re)...)
		out = append(out, checkEdgeLabel(d, re)...)
	}
	out = append(out, checkArrowheadClusters(d)...)
	return out
}

func checkEdgeAnchorSides(d *model.Diagram, re model.RoutedEdge) []Finding {
	byID := map[string]model.Node{}
	for _, n := range d.Nodes {
		byID[n.ID] = n
	}
	a, okA := byID[re.Edge.From]
	b, okB := byID[re.Edge.To]
	if !okA || !okB || !geom.StackedVertically(a, b) {
		return nil
	}
	idealF, idealT := geom.BestSides(a, b)
	fs, ts := re.Edge.FromSide, re.Edge.ToSide
	explicit := (fs != "" && fs != model.SideAuto) || (ts != "" && ts != model.SideAuto)
	if fs == "" || fs == model.SideAuto {
		fs = idealF
	}
	if ts == "" || ts == model.SideAuto {
		ts = idealT
	}
	if !geom.IsHorizontalSide(fs) && !geom.IsHorizontalSide(ts) {
		return nil
	}
	sev := "hint"
	if explicit {
		sev = "warning"
	}
	key := re.Edge.From + "→" + re.Edge.To
	return []Finding{{
		RuleID: "edge_clarity", Severity: sev, Plane: PlaneEdge,
		Code:    string(diag.CodeEdgeSideMismatch),
		Message: fmt.Sprintf("edge %s is vertical but uses side anchors (%s→%s); prefer top/bottom", key, fs, ts),
		Fix:     "Omit fromSide/toSide for auto top/bottom, or set fromSide=bottom toSide=top when stacked.",
		Example: `edge gate -> blocked fromSide=bottom toSide=top`,
		Items:   []string{re.Edge.From, re.Edge.To},
	}}
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
	pathStart := gpts[0]
	pathEnd := gpts[len(gpts)-1]
	srcAnchor, dstAnchor, ok := edgeAnchors(re.Edge, byID, pathStart, pathEnd)
	if !ok {
		return nil
	}

	var out []Finding
	key := re.Edge.From + "→" + re.Edge.To
	directSpan := math.Hypot(pathEnd.X-pathStart.X, pathEnd.Y-pathStart.Y)
	routeLen := 0.0
	for i := 1; i < len(gpts); i++ {
		routeLen += math.Hypot(gpts[i].X-gpts[i-1].X, gpts[i].Y-gpts[i-1].Y)
	}
	if len(gpts) >= 4 && directSpan > 0 && routeLen > directSpan*1.8+d.Gap*2 {
		out = append(out, Finding{
			RuleID: "edge_clarity", Severity: "warning", Plane: PlaneEdge,
			Code:    string(diag.CodeEdgeDetour),
			Message: fmt.Sprintf("edge %s takes a %.0fpx route across a %.0fpx direct span with %d bends", key, routeLen, directSpan, len(gpts)-2),
			Fix:     "Use directionally sensible fromSide/toSide anchors, move blockers, or reorder at= slots to follow the reading direction.",
			Items:   []string{re.Edge.From, re.Edge.To},
		})
	}
	ag, ok := geom.ArrowGeometryForPath(gpts)
	if !ok {
		return out
	}
	if !exitsSide(pathStart, ag.StartApproach, re.Edge.FromSide) || !entersSide(ag.Prev, pathEnd, re.Edge.ToSide) {
		out = append(out, Finding{
			RuleID: "edge_clarity", Severity: "warning", Plane: PlaneEdge,
			Code:    string(diag.CodeEdgeSideMismatch),
			Message: fmt.Sprintf("edge %s route direction does not match its %s→%s side anchors", key, re.Edge.FromSide, re.Edge.ToSide),
			Fix:     "Use side anchors matching the relative positions, or omit fromSide/toSide so the router chooses them.",
			Items:   []string{re.Edge.From, re.Edge.To},
		})
	}

	if geom.TipGap(pathStart, srcAnchor) > anchorEps {
		out = append(out, Finding{
			RuleID: "edge_clarity", Severity: "error", Plane: PlaneEdge,
			Code:    string(diag.CodeArrowDetached),
			Message: fmt.Sprintf("edge %s starts %.0fpx from %q border (should attach at anchor)", key, geom.TipGap(pathStart, srcAnchor), re.Edge.From),
			Fix:     "Set fromSide/toSide so routing ends on shape borders.",
			Items:   []string{re.Edge.From, re.Edge.To},
		})
	}
	if geom.TipGap(pathEnd, dstAnchor) > anchorEps {
		out = append(out, Finding{
			RuleID: "edge_clarity", Severity: "error", Plane: PlaneEdge,
			Code:    string(diag.CodeArrowDetached),
			Message: fmt.Sprintf("edge %s ends %.0fpx from %q border — arrow cannot meet the shape", key, geom.TipGap(pathEnd, dstAnchor), re.Edge.To),
			Fix:     "Route must terminate on the target border; increase gap or fix fromSide/toSide.",
			Items:   []string{re.Edge.From, re.Edge.To},
		})
	}

	if geom.TipGap(ag.Tip, dstAnchor) > geom.MaxArrowTipGap {
		out = append(out, Finding{
			RuleID: "edge_clarity", Severity: "error", Plane: PlaneEdge,
			Code:    string(diag.CodeArrowDetached),
			Message: fmt.Sprintf("edge %s arrow tip is %.0fpx from %q border", key, geom.TipGap(ag.Tip, dstAnchor), re.Edge.To),
			Fix:     "Arrow tip must land on the target border (render draws heads after nodes).",
			Items:   []string{re.Edge.From, re.Edge.To},
		})
	}

	strokeLen := ag.VisibleApproach
	if strokeLen < minVisibleArrowStroke {
		target := byID[re.Edge.To]
		approach := model.Rect{
			X: math.Min(ag.Prev.X, ag.Tip.X),
			Y: math.Min(ag.Prev.Y, ag.Tip.Y),
			W: math.Max(1, math.Abs(ag.Tip.X-ag.Prev.X)),
			H: math.Max(1, math.Abs(ag.Tip.Y-ag.Prev.Y)),
		}
		repairSide := alternateTargetSide(byID[re.Edge.From], target, re.Edge.ToSide)
		out = append(out, Finding{
			RuleID: "edge_clarity", Severity: "warning", Plane: PlaneEdge,
			Code:    string(diag.CodeArrowDetached),
			Message: fmt.Sprintf("edge %s has only %.0fpx of straight stroke before arrowhead — head may look hooked or floating", key, strokeLen),
			Fix:     "Keep the final 27px approach straight, increase gap, or route into a different target side.",
			Items:   []string{re.Edge.From, re.Edge.To},
			Geometry: &diag.Geometry{Bounds: map[string]model.Rect{
				"target": target.Rect, "arrow_approach": approach,
			}},
			Repairs: []diag.RepairOption{{
				Action: "set_property", Target: re.Key,
				Properties: map[string]string{"toSide": string(repairSide)},
				Reason:     fmt.Sprintf("approach %q from its %s side to create a distinct straight shaft", re.Edge.To, repairSide),
			}},
		})
	}

	if geom.TipGap(pathEnd, dstAnchor) > anchorEps {
		if dst, ok := byID[re.Edge.To]; ok && arrowTipBuried(ag.Tip, dst) {
			out = append(out, Finding{
				RuleID: "edge_clarity", Severity: "warning", Plane: PlaneEdge,
				Code:    string(diag.CodeArrowHidden),
				Message: fmt.Sprintf("edge %s arrow tip on %q sits inside the shape (%.0fpx off border)", key, re.Edge.To, geom.TipGap(ag.Tip, dstAnchor)),
				Fix:     "Routing must end on the target border anchor; check fromSide/toSide and re-run validate.",
				Items:   []string{re.Edge.From, re.Edge.To},
			})
		}
	}

	return out
}

func alternateTargetSide(source, target model.Node, current model.Side) model.Side {
	if geom.IsHorizontalSide(current) {
		if source.Rect.CY() < target.Rect.CY() {
			return model.SideTop
		}
		return model.SideBottom
	}
	if source.Rect.CX() < target.Rect.CX() {
		return model.SideLeft
	}
	return model.SideRight
}

func exitsSide(a, b geom.Point, side model.Side) bool {
	dx, dy := b.X-a.X, b.Y-a.Y
	switch side {
	case model.SideLeft:
		return dx < 0 && math.Abs(dx) >= math.Abs(dy)
	case model.SideRight:
		return dx > 0 && math.Abs(dx) >= math.Abs(dy)
	case model.SideTop:
		return dy < 0 && math.Abs(dy) >= math.Abs(dx)
	case model.SideBottom:
		return dy > 0 && math.Abs(dy) >= math.Abs(dx)
	default:
		return true
	}
}

func entersSide(a, b geom.Point, side model.Side) bool {
	// Entering a side is the inverse of exiting it from the same anchor.
	return exitsSide(b, a, side)
}

func edgeAnchors(e model.Edge, byID map[string]model.Node, pathStart, pathEnd geom.Point) (src, dst geom.Point, ok bool) {
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
	src, dst = geom.EdgeAnchors(a, b, fs, ts)
	srcAlong, dstAlong := pathStart.X, pathEnd.X
	if geom.IsHorizontalSide(fs) {
		srcAlong = pathStart.Y
	}
	if geom.IsHorizontalSide(ts) {
		dstAlong = pathEnd.Y
	}
	if p, sliding := geom.SlidingAnchor(a, fs, srcAlong); sliding {
		src = p
	}
	if p, sliding := geom.SlidingAnchor(b, ts, dstAlong); sliding {
		dst = p
	}
	return src, dst, true
}

type arrowTip struct {
	key   string
	edge  model.Edge
	point geom.Point
}

func checkArrowheadClusters(d *model.Diagram) []Finding {
	groups := map[string][]arrowTip{}
	for _, re := range d.Routed {
		pts := geom.SimplifyPath(geom.SlicesToPath(re.Points))
		if len(pts) < 2 {
			continue
		}
		key := re.Edge.To + "\x00" + string(re.Edge.ToSide)
		groups[key] = append(groups[key], arrowTip{
			key: re.Edge.From + "→" + re.Edge.To, edge: re.Edge, point: pts[len(pts)-1],
		})
	}
	var out []Finding
	for _, tips := range groups {
		sort.SliceStable(tips, func(i, j int) bool {
			if tips[i].point.X != tips[j].point.X {
				return tips[i].point.X < tips[j].point.X
			}
			if tips[i].point.Y != tips[j].point.Y {
				return tips[i].point.Y < tips[j].point.Y
			}
			return tips[i].key < tips[j].key
		})
		for i := 1; i < len(tips); i++ {
			a, b := tips[i-1], tips[i]
			distance := geom.TipGap(a.point, b.point)
			if distance >= minArrowTipSeparation {
				continue
			}
			bounds := func(p geom.Point) model.Rect { return model.Rect{X: p.X - 1, Y: p.Y - 1, W: 2, H: 2} }
			out = append(out, Finding{
				RuleID: "edge_clarity", Severity: "warning", Plane: PlaneEdge, Projected: true,
				Code:    string(diag.CodeArrowCluster),
				Message: fmt.Sprintf("arrowheads on %q %s side are only %.0fpx apart (%s and %s)", b.edge.To, b.edge.ToSide, distance, a.key, b.key),
				Fix:     "Increase the target size or route one edge to an adjacent target side; shared rectangular ports fan out automatically.",
				Items:   []string{a.key, b.key, b.edge.To},
				Geometry: &diag.Geometry{Bounds: map[string]model.Rect{
					a.key + ":tip": bounds(a.point), b.key + ":tip": bounds(b.point),
				}},
				Repairs: []diag.RepairOption{{
					Action: "set_property", Target: b.key,
					Properties: map[string]string{"toSide": adjacentSide(b.edge.ToSide)},
					Reason:     "move one arrowhead to an adjacent target side and re-run routing",
				}},
			})
		}
	}
	return out
}

func adjacentSide(side model.Side) string {
	switch side {
	case model.SideLeft, model.SideRight:
		return string(model.SideTop)
	default:
		return string(model.SideLeft)
	}
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
	for _, obstacle := range layout.BlockedBy {
		padded := model.Rect{
			X: obstacle.Bounds.X - geom.EdgeLabelObstacleClearance,
			Y: obstacle.Bounds.Y - geom.EdgeLabelObstacleClearance,
			W: obstacle.Bounds.W + geom.EdgeLabelObstacleClearance*2,
			H: obstacle.Bounds.H + geom.EdgeLabelObstacleClearance*2,
		}
		overlap, _ := rectIntersection(lb, padded)
		code := diag.CodeEdgeCollision
		message := fmt.Sprintf("edge label %q on %s cannot keep %.0fpx clearance from node %q", label, key, geom.EdgeLabelObstacleClearance, obstacle.ID)
		if obstacle.Kind == "chrome" {
			code = diag.CodeEdgeLabelChrome
			message = fmt.Sprintf("edge label %q on %s cannot keep %.0fpx clearance from chrome %q", label, key, geom.EdgeLabelObstacleClearance, obstacle.ID)
		}
		out = append(out, Finding{
			RuleID: "edge_clarity", Severity: "warning", Plane: PlaneLabel, Projected: true,
			Code: string(code), Message: message,
			Fix:   "Increase gap, shorten the label, or move an endpoint so the pill has a clear path segment.",
			Items: []string{re.Edge.From, re.Edge.To, obstacle.ID},
			Geometry: &diag.Geometry{Bounds: map[string]model.Rect{
				key + ":label": lb, obstacle.ID: obstacle.Bounds,
			}, Overlap: overlap},
			Repairs: []diag.RepairOption{{
				Action: "set_property", Target: "diagram",
				Properties: map[string]string{"gap": fmt.Sprintf("%.0f", d.Gap+12)},
				Reason:     "add connector room for the label and its 6px clearance",
			}},
		})
	}
	_, pathY, horiz := geom.LabelPlacement(gpts)
	if horiz && math.Abs(layout.CenterY-pathY) > maxLabelAxisDrift {
		out = append(out, Finding{
			RuleID: "edge_clarity", Severity: "warning", Plane: PlaneLabel,
			Code:    string(diag.CodeEdgeLabelOffAxis),
			Message: fmt.Sprintf("edge label %q on %s sits %.0fpx off the connector — use on-line placement", label, key, math.Abs(layout.CenterY-pathY)),
			Fix:     "Labels should center on the connector in the node gap (opaque box over the stroke).",
			Items:   []string{re.Edge.From, re.Edge.To},
		})
	}

	if chrome, ok := diagramChromeBand(d); ok && rectsOverlap(lb, chrome, 2) {
		out = append(out, Finding{
			RuleID: "edge_clarity", Severity: "warning", Plane: PlaneChrome,
			Code:    string(diag.CodeEdgeLabelChrome),
			Message: fmt.Sprintf("edge label %q on %s overlaps title/subtitle chrome", label, key),
			Fix:     "Move labels onto connectors between nodes, not above the diagram header band.",
			Items:   []string{re.Edge.From, re.Edge.To},
		})
	}

	if lctx != nil && len(gpts) == 2 && math.Abs(gpts[0].Y-gpts[1].Y) < 1 {
		gap := 0.0
		switch {
		case lctx.From.Right() <= lctx.To.X:
			gap = lctx.To.X - lctx.From.Right()
		case lctx.To.Right() <= lctx.From.X:
			gap = lctx.From.X - lctx.To.Right()
		}
		if layout.BoxW > gap-8 {
			out = append(out, Finding{
				RuleID: "edge_clarity", Severity: "warning", Plane: PlaneLabel,
				Code:    string(diag.CodeEdgeLabelChrome),
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
	ctx := &geom.EdgeLabelContext{From: a.Rect, To: b.Rect}
	for _, n := range d.Nodes {
		if model.IsContainer(n.Kind) {
			if bounds := measure.ContainerLabelBounds(n); bounds.W > 0 {
				ctx.Avoid = append(ctx.Avoid, geom.EdgeLabelObstacle{ID: n.ID + ":title", Kind: "chrome", Bounds: bounds})
			}
			continue
		}
		if n.ID != e.From && n.ID != e.To {
			ctx.Avoid = append(ctx.Avoid, geom.EdgeLabelObstacle{ID: n.ID, Kind: "node", Bounds: n.Rect})
		}
	}
	if bounds, ok := diagramChromeBand(d); ok {
		ctx.Avoid = append(ctx.Avoid, geom.EdgeLabelObstacle{ID: "diagram:title", Kind: "chrome", Bounds: bounds})
	}
	return ctx
}

func diagramChromeBand(d *model.Diagram) (model.Rect, bool) {
	return composition.ChromeBounds(*d)
}
