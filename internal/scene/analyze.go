// Package scene provides game-engine-like 2D understanding: paint order, occlusion,
// grouping, edge visibility, and aesthetic checks for agents and validation.
package scene

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/niklas-heer/sceno/internal/diag"
	"github.com/niklas-heer/sceno/internal/geom"
	"github.com/niklas-heer/sceno/internal/model"
)

// Z-order constants (back → front), matching polished render order.
const (
	ZCanvas = iota
	ZLane
	ZEdge
	ZNode
)

// Report is a full 2D scene analysis for one laid-out diagram.
type Report struct {
	Style        string         `json:"style"`
	PaintOrder   []PaintItem    `json:"paint_order"`
	Groups       []Group        `json:"groups"`
	Occlusions   []Occlusion    `json:"occlusions"`
	EdgeVis      []EdgeVis      `json:"edge_visibility"`
	Alignment    []AlignIssue   `json:"alignment"`
	Aesthetics   AestheticScore `json:"aesthetics"`
	Issues       []diag.Issue   `json:"issues"`
}

type PaintItem struct {
	Z    int    `json:"z"`
	Kind string `json:"kind"` // lane | edge | node
	ID   string `json:"id,omitempty"`
	Key  string `json:"key,omitempty"`
}

type Group struct {
	Kind    string   `json:"kind"` // column | proximity
	Label   string   `json:"label"`
	NodeIDs []string `json:"node_ids"`
	Bounds  string   `json:"bounds,omitempty"`
}

type Occlusion struct {
	Over   string  `json:"covers"`
	Under  string  `json:"covered"`
	AreaPx float64 `json:"overlap_px"`
	Note   string  `json:"note"`
}

type EdgeVis struct {
	Key       string  `json:"key"`
	From      string  `json:"from"`
	To        string  `json:"to"`
	Visible   float64 `json:"visible_fraction"`
	Organic   bool    `json:"organic_route"`
	HiddenPx  float64 `json:"hidden_estimate_px,omitempty"`
}

type AlignIssue struct {
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Nodes   []string `json:"nodes,omitempty"`
}

type AestheticScore struct {
	Overall      int      `json:"overall"` // 0–100
	Density      float64  `json:"density"` // content / canvas
	EdgeClarity  float64  `json:"edge_clarity_avg"`
	GroupSpacing float64  `json:"group_spacing_ok"`
	Notes        []string `json:"notes,omitempty"`
}

// Analyze inspects a laid-out diagram in 2D (layers, visibility, aesthetics).
func Analyze(d *model.Diagram) Report {
	if d == nil {
		return Report{}
	}
	style := string(d.Style)
	if style == "" {
		style = string(model.StylePolished)
	}
	r := Report{
		Style:      style,
		PaintOrder: buildPaintOrder(d),
		Groups:     findGroups(d),
	}
	r.Occlusions = findOcclusions(d)
	r.EdgeVis = edgeVisibility(d)
	r.Alignment = alignmentIssues(d)
	r.Aesthetics = scoreAesthetics(d, r)
	r.Issues = buildIssues(d, r)
	return r
}

func buildPaintOrder(d *model.Diagram) []PaintItem {
	var items []PaintItem
	for _, n := range d.Nodes {
		if n.Kind == model.ShapeLane {
			items = append(items, PaintItem{Z: ZLane, Kind: "lane", ID: n.ID})
		}
	}
	for _, re := range d.Routed {
		items = append(items, PaintItem{Z: ZEdge, Kind: "edge", Key: re.Key})
	}
	for _, n := range d.Nodes {
		if n.Kind != model.ShapeLane {
			items = append(items, PaintItem{Z: ZNode, Kind: "node", ID: n.ID})
		}
	}
	return items
}

func findGroups(d *model.Diagram) []Group {
	byCol := map[int][]model.Node{}
	for _, n := range d.Nodes {
		if model.IsContainer(n.Kind) {
			continue
		}
		col := n.Column
		if col == 0 {
			col = n.Layer
		}
		byCol[col] = append(byCol[col], n)
	}
	cols := make([]int, 0, len(byCol))
	for c := range byCol {
		cols = append(cols, c)
	}
	sort.Ints(cols)
	var groups []Group
	for _, c := range cols {
		ns := byCol[c]
		sort.Slice(ns, func(i, j int) bool { return ns[i].Row < ns[j].Row })
		ids := make([]string, len(ns))
		for i, n := range ns {
			ids[i] = n.ID
		}
		b := unionBounds(ns)
		groups = append(groups, Group{
			Kind:    "column",
			Label:   fmt.Sprintf("column %d", c),
			NodeIDs: ids,
			Bounds:  fmt.Sprintf("(%.0f,%.0f)–(%.0f,%.0f)", b.X, b.Y, b.Right(), b.Bottom()),
		})
	}
	// Proximity clusters (same row band, close X)
	gap := d.Gap
	if gap < 16 {
		gap = 16
	}
	for i := 0; i < len(d.Nodes); i++ {
		a := d.Nodes[i]
		if model.IsContainer(a.Kind) {
			continue
		}
		var cluster []string
		for j := i + 1; j < len(d.Nodes); j++ {
			b := d.Nodes[j]
			if model.IsContainer(b.Kind) {
				continue
			}
			if math.Abs(a.Rect.CY()-b.Rect.CY()) > gap*1.2 {
				continue
			}
			if math.Abs(a.Rect.CX()-b.Rect.CX()) < gap*3 {
				if len(cluster) == 0 {
					cluster = []string{a.ID}
				}
				cluster = append(cluster, b.ID)
			}
		}
		if len(cluster) >= 2 {
			groups = append(groups, Group{
				Kind:    "proximity",
				Label:   "row cluster",
				NodeIDs: cluster,
			})
		}
	}
	return groups
}

func findOcclusions(d *model.Diagram) []Occlusion {
	var out []Occlusion
	nodes := nonLaneNodes(d)
	for i := 0; i < len(nodes); i++ {
		for j := 0; j < len(nodes); j++ {
			if i == j {
				continue
			}
			// Later in paint order (higher index in Nodes slice) covers earlier when overlapping.
			ni, nj := nodeIndex(d, nodes[i].ID), nodeIndex(d, nodes[j].ID)
			if ni <= nj {
				continue
			}
			area := geom.RectOverlapArea(nodes[i].Rect, nodes[j].Rect)
			if area < 4 {
				continue
			}
			out = append(out, Occlusion{
				Over:   nodes[i].ID,
				Under:  nodes[j].ID,
				AreaPx: area,
				Note:   "node drawn on top overlaps another (collision or stacking)",
			})
		}
	}
	return out
}

func nodeIndex(d *model.Diagram, id string) int {
	for i, n := range d.Nodes {
		if n.ID == id {
			return i
		}
	}
	return -1
}

func edgeVisibility(d *model.Diagram) []EdgeVis {
	pad := d.Gap * 0.5
	if pad < 10 {
		pad = 10
	}
	sketch := d.Style == model.StyleSketch
	var out []EdgeVis
	for _, re := range d.Routed {
		pts := pathToGeom(re.Points)
		if sketch && len(pts) >= 3 {
			pts = geom.SmoothPath(pts, 6)
		}
		frac := geom.PathVisibleFraction(pts, re.Edge.From, re.Edge.To, d.Nodes, pad)
		ev := EdgeVis{
			Key:     re.Key,
			From:    re.Edge.From,
			To:      re.Edge.To,
			Visible: math.Round(frac*100) / 100,
			Organic: sketch || len(pts) > 4,
		}
		if frac < 1 {
			ev.HiddenPx = pathLen(pts) * (1 - frac)
		}
		out = append(out, ev)
	}
	return out
}

func alignmentIssues(d *model.Diagram) []AlignIssue {
	var out []AlignIssue
	tol := d.Gap / 4
	if tol < 6 {
		tol = 6
	}

	byCol := map[int][]model.Node{}
	for _, n := range d.Nodes {
		if model.IsContainer(n.Kind) {
			continue
		}
		col := n.Column
		if col == 0 {
			col = n.Layer
		}
		byCol[col] = append(byCol[col], n)
	}
	for col, ns := range byCol {
		if len(ns) < 2 {
			continue
		}
		var xs []float64
		for _, n := range ns {
			xs = append(xs, n.Rect.CX())
		}
		sort.Float64s(xs)
		span := xs[len(xs)-1] - xs[0]
		if span > tol*2 {
			ids := make([]string, len(ns))
			for i, n := range ns {
				ids[i] = n.ID
			}
			out = append(out, AlignIssue{
				Code:    "column_misaligned",
				Message: fmt.Sprintf("column %d nodes are not vertically aligned (%.0fpx spread)", col, span),
				Nodes:   ids,
			})
		}
	}

	for _, n := range d.Nodes {
		if n.Icon == "" || n.Label == "" || model.IsContainer(n.Kind) {
			continue
		}
		// Icon sits left; label should not be centered on full box when icon present.
		centerOff := math.Abs(n.Rect.CX() - (n.Rect.X+n.Rect.W*0.58))
		if centerOff > n.Rect.W*0.22 {
			out = append(out, AlignIssue{
				Code:    "label_icon_balance",
				Message: fmt.Sprintf("node %q label may not balance with icon column", n.ID),
				Nodes:   []string{n.ID},
			})
		}
	}
	return out
}

func scoreAesthetics(d *model.Diagram, r Report) AestheticScore {
	minX, minY, maxX, maxY := canvasExtents(d)
	cw := maxX - minX
	ch := maxY - minY
	if cw < 1 {
		cw = 1
	}
	if ch < 1 {
		ch = 1
	}
	content := 0.0
	for _, n := range d.Nodes {
		if !model.IsContainer(n.Kind) {
			content += n.Rect.W * n.Rect.H
		}
	}
	density := content / (cw * ch)

	var edgeSum float64
	for _, ev := range r.EdgeVis {
		edgeSum += ev.Visible
	}
	edgeAvg := 1.0
	if len(r.EdgeVis) > 0 {
		edgeAvg = edgeSum / float64(len(r.EdgeVis))
	}

	groupOK := 1.0
	if len(r.Groups) > 1 {
		gap := d.Gap
		if gap < 16 {
			gap = 16
		}
		// penalize if column groups are too tight
		for i := 1; i < len(r.Groups); i++ {
			if r.Groups[i].Kind != "column" {
				continue
			}
		}
		_ = groupOK
	}

	score := 100
	if density < 0.08 {
		score -= 15
	}
	if density > 0.85 {
		score -= 10
	}
	if edgeAvg < 0.75 {
		score -= int((0.75 - edgeAvg) * 40)
	}
	if len(r.Occlusions) > 0 {
		score -= 20 * len(r.Occlusions)
	}
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	notes := []string{}
	if d.Style == model.StyleSketch {
		notes = append(notes, "sketch style uses organic curved connectors")
	}
	if density < 0.1 {
		notes = append(notes, "layout is sparse — consider lower gap or tighter layers")
	}
	if edgeAvg < 0.9 && len(r.EdgeVis) > 0 {
		notes = append(notes, "some arrows run behind nodes — increase gap or set fromSide/toSide")
	}

	return AestheticScore{
		Overall:     score,
		Density:     math.Round(density*1000) / 1000,
		EdgeClarity: math.Round(edgeAvg*100) / 100,
		Notes:       notes,
	}
}

func buildIssues(d *model.Diagram, r Report) []diag.Issue {
	var issues []diag.Issue
	for _, o := range r.Occlusions {
		issues = append(issues, diag.Issue{
			Code:    diag.CodeOccluded,
			Message: fmt.Sprintf("node %q visually covers %q (%.0f px² overlap)", o.Over, o.Under, o.AreaPx),
			Fix:     "Separate nodes with gap/layer/row, or fix collision nudge.",
			Nodes:   []string{o.Over, o.Under},
		})
	}
	for _, ev := range r.EdgeVis {
		if ev.Visible >= 0.82 {
			continue
		}
		issues = append(issues, diag.Issue{
			Code:    diag.CodeEdgeHidden,
			Message: fmt.Sprintf("edge %s→%s is ~%.0f%% visible (runs behind nodes)", ev.From, ev.To, ev.Visible*100),
			Fix:     "Set fromSide/toSide, increase gap, or reroute with a wider lane offset.",
			Nodes:   []string{ev.From, ev.To},
		})
	}
	for _, a := range r.Alignment {
		code := diag.CodeMisaligned
		issues = append(issues, diag.Issue{
			Code:    code,
			Message: a.Message,
			Fix:     "Align nodes in the same column (consistent layer/at) or balance icon+label layout.",
			Nodes:   a.Nodes,
		})
	}
	if r.Aesthetics.Overall < 55 && r.Aesthetics.Density < 0.1 {
		issues = append(issues, diag.Issue{
			Code:    diag.CodeSuggestCompact,
			Message: "scene looks sparse and low-contrast for its canvas",
			Fix:     "Reduce gap, stack rows, or add grouping lanes.",
		})
	}
	return issues
}

func nonLaneNodes(d *model.Diagram) []model.Node {
	var out []model.Node
	for _, n := range d.Nodes {
		if n.Kind != model.ShapeLane {
			out = append(out, n)
		}
	}
	return out
}

func unionBounds(ns []model.Node) model.Rect {
	if len(ns) == 0 {
		return model.Rect{}
	}
	minX, minY := ns[0].Rect.X, ns[0].Rect.Y
	maxX, maxY := ns[0].Rect.Right(), ns[0].Rect.Bottom()
	for _, n := range ns[1:] {
		if n.Rect.X < minX {
			minX = n.Rect.X
		}
		if n.Rect.Y < minY {
			minY = n.Rect.Y
		}
		if n.Rect.Right() > maxX {
			maxX = n.Rect.Right()
		}
		if n.Rect.Bottom() > maxY {
			maxY = n.Rect.Bottom()
		}
	}
	return model.Rect{X: minX, Y: minY, W: maxX - minX, H: maxY - minY}
}

func canvasExtents(d *model.Diagram) (minX, minY, maxX, maxY float64) {
	minX, minY = 1e9, 1e9
	maxX, maxY = -1e9, -1e9
	for _, n := range d.Nodes {
		if n.Rect.X < minX {
			minX = n.Rect.X
		}
		if n.Rect.Y < minY {
			minY = n.Rect.Y
		}
		if n.Rect.Right() > maxX {
			maxX = n.Rect.Right()
		}
		if n.Rect.Bottom() > maxY {
			maxY = n.Rect.Bottom()
		}
	}
	if maxX < minX {
		return 0, 0, 400, 300
	}
	return minX, minY, maxX, maxY
}

func pathToGeom(path [][]float64) []geom.Point {
	pts := make([]geom.Point, 0, len(path))
	for _, p := range path {
		if len(p) >= 2 {
			pts = append(pts, geom.Point{X: p[0], Y: p[1]})
		}
	}
	return pts
}

func pathLen(pts []geom.Point) float64 {
	var sum float64
	for i := 1; i < len(pts); i++ {
		sum += math.Hypot(pts[i].X-pts[i-1].X, pts[i].Y-pts[i-1].Y)
	}
	return sum
}

// NarrativeSummary is a short agent-readable scene description.
func NarrativeSummary(r Report) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("2D scene (%s): paint order lanes→edges→nodes", r.Style))
	if len(r.Groups) > 0 {
		parts = append(parts, fmt.Sprintf("%d logical group(s)", len(r.Groups)))
	}
	if len(r.Occlusions) > 0 {
		parts = append(parts, fmt.Sprintf("%d occlusion(s)", len(r.Occlusions)))
	}
	hidden := 0
	for _, ev := range r.EdgeVis {
		if ev.Visible < 0.82 {
			hidden++
		}
	}
	if hidden > 0 {
		parts = append(parts, fmt.Sprintf("%d edge(s) partly hidden", hidden))
	}
	parts = append(parts, fmt.Sprintf("aesthetic score %d/100", r.Aesthetics.Overall))
	return strings.Join(parts, "; ")
}
