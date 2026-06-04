// Package inspect turns laid-out diagrams into text for agents that cannot view images.
package inspect

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sort"
	"strings"

	"github.com/niklas-heer/sceno/internal/diag"
	"github.com/niklas-heer/sceno/internal/layout"
	"github.com/niklas-heer/sceno/internal/measure"
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/pipeline"
	"github.com/niklas-heer/sceno/internal/render"
	"github.com/niklas-heer/sceno/internal/scene"
	"github.com/niklas-heer/sceno/internal/spec"
	"github.com/niklas-heer/sceno/internal/validate"
	"github.com/niklas-heer/sceno/internal/version"
)

// Options for describe.
type Options struct {
	FixCollisions bool
	ASCIIWidth    int // default 56
}

// Report is textual visual feedback (--json for agents).
type Report struct {
	Input        string       `json:"input"`
	Tool         string       `json:"tool"`
	Version      string       `json:"version"`
	ValidationOK bool         `json:"validation_ok"`
	RenderReady  bool         `json:"render_ready"`
	Purpose      string       `json:"purpose"`
	Agent        DescribeMeta `json:"agent"`
	Slides       []SlideView  `json:"slides"`
}

// DescribeMeta explains how to use this output.
type DescribeMeta struct {
	Summary  string   `json:"summary"`
	ReadFirst []string `json:"read_first"`
	Hint     string   `json:"hint"`
}

// SlideView describes one laid-out slide or diagram.
type SlideView struct {
	Index           int              `json:"index"`
	Title           string           `json:"title,omitempty"`
	Subtitle        string           `json:"subtitle,omitempty"`
	Narrative       string           `json:"narrative"`
	Canvas          CanvasInfo       `json:"canvas"`
	Header          string           `json:"header,omitempty"`
	Columns         []ColumnSummary  `json:"columns,omitempty"`
	Nodes           []NodeView       `json:"nodes"`
	Edges           []EdgeView       `json:"edges"`
	Relationships   []string         `json:"relationships"`
	Scene           scene.Report     `json:"scene"`
	VisualProblems  []VisualProblem  `json:"visual_problems"`
	ASCIIMap        string           `json:"ascii_map"`
	Stats           ViewStats        `json:"stats"`
}

type CanvasInfo struct {
	MinX   float64 `json:"min_x"`
	MinY   float64 `json:"min_y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type ViewStats struct {
	Nodes      int `json:"nodes"`
	Edges      int `json:"edges"`
	Columns    int `json:"columns"`
	Overlaps   int `json:"overlaps"`
	Crossings  int `json:"edge_node_crossings"`
}

type ColumnSummary struct {
	Column int      `json:"column"`
	Nodes  []string `json:"node_ids"`
	Rows   []int    `json:"rows,omitempty"`
}

type NodeView struct {
	ID          string  `json:"id"`
	Label       string  `json:"label"`
	Kind        string  `json:"kind"`
	Icon        string  `json:"icon,omitempty"`
	Subtitle    string  `json:"subtitle,omitempty"`
	Column      int     `json:"column,omitempty"`
	Row         int     `json:"row,omitempty"`
	Region      string  `json:"region"`
	Position    string  `json:"position"`
	Bounds      RectView `json:"bounds"`
	Size        string  `json:"size"`
	Fill        string  `json:"fill,omitempty"`
}

type RectView struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	W      float64 `json:"w"`
	H      float64 `json:"h"`
	Center string  `json:"center"`
}

type EdgeView struct {
	From        string  `json:"from"`
	To          string  `json:"to"`
	Label       string  `json:"label,omitempty"`
	Route       string  `json:"route"`
	Attachment  string  `json:"attachment"`
	Style       string  `json:"style,omitempty"`
	Segments    int     `json:"segments"`
	Organic     bool    `json:"organic_route,omitempty"`
	VisiblePct  float64 `json:"visible_pct,omitempty"`
	CrossesNode string  `json:"crosses_node,omitempty"`
}

type VisualProblem struct {
	Severity    string   `json:"severity"` // error | warning
	Code        string   `json:"code"`
	Message     string   `json:"message"`
	Fix         string   `json:"fix,omitempty"`
	Where       string   `json:"where"`
	Involves    []string `json:"involves,omitempty"`
}

// Run builds layout and returns a describe report.
func Run(path string, opt Options) (Report, error) {
	if opt.ASCIIWidth <= 0 {
		opt.ASCIIWidth = 56
	}
	vreport, _, _ := validate.Run(path, validate.Options{FixCollisions: opt.FixCollisions})

	s, err := spec.LoadFile(path)
	if err != nil {
		return Report{}, err
	}
	popt := pipeline.DefaultOptions()
	popt.ResolveCollision = opt.FixCollisions
	deck, colls, err := pipeline.BuildDeck(s, popt)
	if err != nil {
		return Report{}, err
	}

	report := Report{
		Input:        path,
		Tool:         "sceno",
		Version:      version.Version,
		ValidationOK: vreport.OK,
		RenderReady:  vreport.OK,
		Purpose:      "Textual description of how the diagram looks after layout — for agents that cannot view SVG/PNG.",
		Agent: DescribeMeta{
			ReadFirst: []string{
				"narrative — plain-language overview",
				"scene — 2D paint order, groups, occlusion, edge visibility",
				"ascii_map — coarse spatial map (* = node, lines = edges)",
				"visual_problems — what looks wrong and where",
				"nodes / edges — exact positions and routes",
			},
			Hint: "Pair with sceno validate --json for repair steps; use describe after render-worthy validate passes to sanity-check layout.",
		},
	}

	// Map validation issues by slide (single diagram = slide 0)
	issuesBySlide := groupIssues(vreport, len(deck.Slides))

	margin := s.Gap / 2
	if margin < 8 {
		margin = 8
	}

	for i, d := range deck.Slides {
		slideColls := filterCollisions(colls, &d)
		sv := describeSlide(d, i, issuesBySlide[i], slideColls, margin, opt.ASCIIWidth)
		report.Slides = append(report.Slides, sv)
	}

	report.Agent.Summary = buildSummary(report)
	return report, nil
}

func describeSlide(d model.Diagram, index int, issues []diag.Issue, colls []model.Collision, margin float64, asciiW int) SlideView {
	minX, minY, maxX, maxY := render.Bounds(d)
	cw, ch := maxX-minX, maxY-minY

	sv := SlideView{
		Index:     index + 1,
		Title:     d.Title,
		Subtitle:  d.Subtitle,
		Canvas:    CanvasInfo{MinX: minX, MinY: minY, Width: cw, Height: ch},
		Stats:     ViewStats{Nodes: len(d.Nodes), Edges: len(d.Edges)},
	}

	if d.Title != "" {
		sv.Header = fmt.Sprintf("Title %q at top of canvas", d.Title)
		if d.Subtitle != "" {
			sv.Header += fmt.Sprintf("; subtitle %q below it", d.Subtitle)
		}
	}

	sv.Columns = summarizeColumns(d.Nodes)
	for _, n := range sortNodes(d.Nodes) {
		sv.Nodes = append(sv.Nodes, describeNode(n, minX, minY, maxX, maxY))
	}

	sv.Scene = scene.Analyze(&d)

	edgeHitsNode := map[string]string{}
	var edgeHitsEdge []model.EdgeCollision
	for _, ec := range layout.FindEdgeCollisions(&d) {
		switch ec.Kind {
		case "node_crossing":
			edgeHitsNode[ec.EdgeKey] = ec.With
		case "edge_crossing":
			edgeHitsEdge = append(edgeHitsEdge, ec)
		}
	}

	visByKey := map[string]scene.EdgeVis{}
	for _, ev := range sv.Scene.EdgeVis {
		visByKey[ev.Key] = ev
	}
	for i, re := range d.Routed {
		ev := describeEdge(re, i, edgeHitsNode[re.Key], visByKey[re.Key], d.Style)
		sv.Edges = append(sv.Edges, ev)
		if ev.CrossesNode != "" {
			sv.Stats.Crossings++
		}
	}

	sv.Relationships = spatialRelationships(d.Nodes)
	sv.VisualProblems = visualProblems(d, issues, colls, margin, edgeHitsNode, edgeHitsEdge)
	sv.VisualProblems = append(sv.VisualProblems, sceneVisualProblems(sv.Scene, minX, minY, maxX, maxY)...)
	sv.VisualProblems = dedupeVisualProblems(sv.VisualProblems)
	sv.Stats.Overlaps = countSeverity(sv.VisualProblems, "collision")
	sv.Stats.Columns = len(sv.Columns)
	sv.ASCIIMap = asciiMap(d, minX, minY, maxX, maxY, asciiW)
	sv.Narrative = buildNarrative(sv)

	return sv
}

func describeNode(n model.Node, minX, minY, maxX, maxY float64) NodeView {
	k := model.NormalizeShape(n.Kind)
	label := n.Label
	if label == "" {
		label = n.ID
	}
	label = strings.ReplaceAll(label, "\n", " / ")
	nv := NodeView{
		ID:       n.ID,
		Label:    label,
		Kind:     string(k),
		Icon:     n.Icon,
		Subtitle: n.Subtitle,
		Column:   n.Column,
		Row:      n.Row,
		Region:   region(n.Rect.CX(), n.Rect.CY(), minX, minY, maxX, maxY),
		Bounds: RectView{
			X: n.Rect.X, Y: n.Rect.Y, W: n.Rect.W, H: n.Rect.H,
			Center: fmt.Sprintf("(%.0f, %.0f)", n.Rect.CX(), n.Rect.CY()),
		},
		Size: fmt.Sprintf("%.0f×%.0f px", n.Rect.W, n.Rect.H),
		Fill: n.Fill,
	}
	nv.Position = fmt.Sprintf("column %d row %d; %s; center %s",
		n.Column, n.Row, nv.Region, nv.Bounds.Center)
	if n.Layer > 0 && n.Column != n.Layer {
		nv.Position = fmt.Sprintf("layer %d %s; center %s", n.Layer, nv.Region, nv.Bounds.Center)
	}
	return nv
}

func describeEdge(re model.RoutedEdge, idx int, crosses string, vis scene.EdgeVis, diagramStyle model.RenderStyle) EdgeView {
	e := re.Edge
	attach := ""
	if e.FromSide != "" || e.ToSide != "" {
		attach = fmt.Sprintf("%s → %s", sideOrDefault(e.FromSide, "auto"), sideOrDefault(e.ToSide, "auto"))
	} else {
		attach = "auto sides (from relative positions)"
	}
	style := ""
	if e.Dashed {
		style = "dashed"
	}
	if e.Color != "" {
		if style != "" {
			style += ", "
		}
		style += "color " + e.Color
	}
	ev := EdgeView{
		From:        e.From,
		To:          e.To,
		Label:       e.Label,
		Route:       describeRoute(re.Points),
		Attachment:  attach,
		Style:       style,
		Segments:    segmentCount(re.Points),
		CrossesNode: crosses,
	}
	if vis.Key != "" {
		ev.Organic = vis.Organic
		ev.VisiblePct = vis.Visible
	} else if diagramStyle == model.StyleSketch {
		ev.Organic = segmentCount(re.Points) >= 3
	}
	return ev
}

func describeRoute(pts [][]float64) string {
	if len(pts) < 2 {
		return "no path"
	}
	gpts := make([]struct{ x, y float64 }, 0, len(pts))
	for _, p := range pts {
		if len(p) >= 2 {
			gpts = append(gpts, struct{ x, y float64 }{p[0], p[1]})
		}
	}
	var parts []string
	p0 := gpts[0]
	pl := gpts[len(gpts)-1]
	parts = append(parts, fmt.Sprintf("from (%.0f, %.0f)", p0.x, p0.y))
	for i := 1; i < len(gpts); i++ {
		parts = append(parts, segmentWords(gpts[i-1].x, gpts[i-1].y, gpts[i].x, gpts[i].y))
	}
	parts = append(parts, fmt.Sprintf("to (%.0f, %.0f)", pl.x, pl.y))
	return strings.Join(parts, ", ")
}

func segmentWords(x1, y1, x2, y2 float64) string {
	dx, dy := x2-x1, y2-y1
	if math.Abs(dx) < 2 && math.Abs(dy) < 2 {
		return "short hop"
	}
	var segs []string
	if math.Abs(dy) >= 2 {
		dir := "down"
		if dy < 0 {
			dir = "up"
		}
		segs = append(segs, fmt.Sprintf("%s %.0fpx", dir, math.Abs(dy)))
	}
	if math.Abs(dx) >= 2 {
		dir := "right"
		if dx < 0 {
			dir = "left"
		}
		segs = append(segs, fmt.Sprintf("%s %.0fpx", dir, math.Abs(dx)))
	}
	if len(segs) == 0 {
		return "micro-adjust"
	}
	return strings.Join(segs, " then ")
}

func segmentCount(pts [][]float64) int {
	if len(pts) < 2 {
		return 0
	}
	return len(pts) - 1
}

func region(cx, cy, minX, minY, maxX, maxY float64) string {
	if maxX-minX < 1 || maxY-minY < 1 {
		return "center"
	}
	col := 0
	if cx > minX+(maxX-minX)/3 {
		col = 1
	}
	if cx > minX+2*(maxX-minX)/3 {
		col = 2
	}
	row := 0
	if cy > minY+(maxY-minY)/3 {
		row = 1
	}
	if cy > minY+2*(maxY-minY)/3 {
		row = 2
	}
	names := [][]string{
		{"top-left", "top-center", "top-right"},
		{"middle-left", "center", "middle-right"},
		{"bottom-left", "bottom-center", "bottom-right"},
	}
	return names[row][col]
}

func spatialRelationships(nodes []model.Node) []string {
	if len(nodes) < 2 {
		return nil
	}
	sorted := sortNodes(nodes)
	var rels []string
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			a, b := sorted[i], sorted[j]
			if r := relativePhrase(a, b); r != "" {
				rels = append(rels, fmt.Sprintf("%q (%s) is %s %q (%s)", a.ID, shortLabel(a), r, b.ID, shortLabel(b)))
			}
		}
		if len(rels) >= 12 {
			break
		}
	}
	return rels
}

func relativePhrase(a, b model.Node) string {
	dx := b.Rect.CX() - a.Rect.CX()
	dy := b.Rect.CY() - a.Rect.CY()
	const thresh = 24.0
	if math.Abs(dx) < thresh && math.Abs(dy) < thresh {
		return "overlapping or very close to"
	}
	if math.Abs(dx) >= math.Abs(dy) {
		if dx > thresh {
			return "to the right of"
		}
		if dx < -thresh {
			return "to the left of"
		}
	}
	if dy > thresh {
		return "below"
	}
	if dy < -thresh {
		return "above"
	}
	return ""
}

func visualProblems(d model.Diagram, issues []diag.Issue, colls []model.Collision, margin float64, edgeHitsNode map[string]string, edgeHitsEdge []model.EdgeCollision) []VisualProblem {
	var out []VisualProblem
	byID := map[string]model.Node{}
	for _, n := range d.Nodes {
		byID[n.ID] = n
	}

	minX, minY, maxX, maxY := render.Bounds(d)
	for _, c := range colls {
		a, b := byID[c.A], byID[c.B]
		where := fmt.Sprintf("nodes %q and %q", c.A, c.B)
		if a.ID != "" && b.ID != "" {
			where = fmt.Sprintf("%s (%s) overlaps %s (%s) near %s",
				a.ID, region(a.Rect.CX(), a.Rect.CY(), minX, minY, maxX, maxY),
				b.ID, region(b.Rect.CX(), b.Rect.CY(), minX, minY, maxX, maxY),
				overlapRegion(a.Rect, b.Rect))
		}
		out = append(out, VisualProblem{
			Severity: "error",
			Code:     string(diag.CodeCollision),
			Message:  fmt.Sprintf("nodes %q and %q overlap on canvas", c.A, c.B),
			Fix:      "Increase gap or separate at=col,row positions.",
			Where:    where,
			Involves: []string{c.A, c.B},
		})
	}

	for key, nodeID := range edgeHitsNode {
		where := fmt.Sprintf("connector crosses node %q", nodeID)
		if n, ok := byID[nodeID]; ok {
			where = fmt.Sprintf("connector crosses %q (%s) in %s", nodeID, shortLabel(n),
				region(n.Rect.CX(), n.Rect.CY(), minX, minY, maxX, maxY))
		}
		out = append(out, VisualProblem{
			Severity: "error",
			Code:     string(diag.CodeEdgeCollision),
			Message:  fmt.Sprintf("edge %s passes through node %q", key, nodeID),
			Fix:      "Set fromSide/toSide or move nodes to different layers.",
			Where:    where,
			Involves: []string{key, nodeID},
		})
	}

	for _, ec := range edgeHitsEdge {
		out = append(out, VisualProblem{
			Severity: "warning",
			Code:     string(diag.CodeEdgeCollision),
			Message:  fmt.Sprintf("edge routes %s and %s cross each other", ec.EdgeKey, ec.With),
			Fix:      "Usually acceptable; adjust layers or fromSide/toSide if confusing.",
			Where:    "connector paths intersect between nodes (see ascii_map)",
			Involves: []string{ec.EdgeKey, ec.With},
		})
	}

	for _, iss := range issues {
		if iss.Code == diag.CodeCollision || iss.Code == diag.CodeTextOverflow || iss.Code == diag.CodeEdgeCollision ||
			iss.Code == diag.CodeSuggestCompact {
			continue
		}
		sev := "warning"
		if iss.Code == diag.CodeEdgeCollision || iss.Code == diag.CodeMissingNode {
			sev = "error"
		}
		where := issueWhere(iss, byID, minX, minY, maxX, maxY)
		out = append(out, VisualProblem{
			Severity: sev,
			Code:     string(iss.Code),
			Message:  iss.Message,
			Fix:      iss.Fix,
			Where:    where,
			Involves: append(iss.Nodes, iss.Edge...),
		})
	}

	for _, iss := range measure.FindTextOverflow(d.Nodes) {
		n := byID[iss.Nodes[0]]
		where := iss.Nodes[0]
		if n.ID != "" {
			where = fmt.Sprintf("node %q (%s) at %s — label may clip", n.ID, shortLabel(n), region(n.Rect.CX(), n.Rect.CY(), minX, minY, maxX, maxY))
		}
		out = append(out, VisualProblem{
			Severity: "error",
			Code:     string(diag.CodeTextOverflow),
			Message:  iss.Message,
			Fix:      iss.Fix,
			Where:    where,
			Involves: iss.Nodes,
		})
	}

	return out
}

func dedupeVisualProblems(problems []VisualProblem) []VisualProblem {
	if len(problems) < 2 {
		return problems
	}
	seen := make(map[string]struct{}, len(problems))
	out := make([]VisualProblem, 0, len(problems))
	for _, p := range problems {
		key := p.Severity + "|" + p.Code + "|" + p.Message
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, p)
	}
	return out
}

func issueWhere(iss diag.Issue, byID map[string]model.Node, minX, minY, maxX, maxY float64) string {
	if len(iss.Nodes) == 1 {
		if n, ok := byID[iss.Nodes[0]]; ok {
			return fmt.Sprintf("node %q (%s), %s, center (%.0f, %.0f)",
				n.ID, shortLabel(n), region(n.Rect.CX(), n.Rect.CY(), minX, minY, maxX, maxY), n.Rect.CX(), n.Rect.CY())
		}
		return "node " + iss.Nodes[0]
	}
	if len(iss.Nodes) == 2 {
		return fmt.Sprintf("between %q and %q", iss.Nodes[0], iss.Nodes[1])
	}
	if len(iss.Edge) == 2 {
		return fmt.Sprintf("edge %s → %s", iss.Edge[0], iss.Edge[1])
	}
	if iss.Path != "" {
		return iss.Path
	}
	return "sceno"
}

func overlapRegion(a, b model.Rect) string {
	ix1 := math.Max(a.X, b.X)
	iy1 := math.Max(a.Y, b.Y)
	ix2 := math.Min(a.Right(), b.Right())
	iy2 := math.Min(a.Bottom(), b.Bottom())
	if ix2 > ix1 && iy2 > iy1 {
		return fmt.Sprintf("overlap box center (%.0f, %.0f)", (ix1+ix2)/2, (iy1+iy2)/2)
	}
	return "shared area"
}

func summarizeColumns(nodes []model.Node) []ColumnSummary {
	groups := map[int][]model.Node{}
	for _, n := range nodes {
		if model.IsContainer(n.Kind) {
			continue
		}
		c := n.Column
		if c < 0 {
			c = n.Layer
		}
		groups[c] = append(groups[c], n)
	}
	cols := make([]int, 0, len(groups))
	for c := range groups {
		cols = append(cols, c)
	}
	sort.Ints(cols)
	var out []ColumnSummary
	for _, c := range cols {
		ns := groups[c]
		sort.Slice(ns, func(i, j int) bool { return ns[i].Row < ns[j].Row })
		cs := ColumnSummary{Column: c}
		for _, n := range ns {
			cs.Nodes = append(cs.Nodes, n.ID)
			cs.Rows = append(cs.Rows, n.Row)
		}
		out = append(out, cs)
	}
	return out
}

func buildNarrative(sv SlideView) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Canvas %.0f×%.0f px with %d nodes and %d edges. ",
		sv.Canvas.Width, sv.Canvas.Height, sv.Stats.Nodes, sv.Stats.Edges)
	if sv.Header != "" {
		b.WriteString(sv.Header + ". ")
	}
	if len(sv.Columns) > 0 {
		b.WriteString("Columns (left to right): ")
		for i, c := range sv.Columns {
			if i > 0 {
				b.WriteString("; ")
			}
			b.WriteString(fmt.Sprintf("col %d has %s", c.Column, strings.Join(c.Nodes, ", ")))
		}
		b.WriteString(". ")
	}
	if sv.Scene.Aesthetics.Overall > 0 {
		b.WriteString(scene.NarrativeSummary(sv.Scene) + ". ")
	}
	errs, warns := countProblemsBySeverity(sv.VisualProblems)
	switch {
	case errs > 0:
		fmt.Fprintf(&b, "%d visual error(s), %d warning(s) — see visual_problems and ascii_map.", errs, warns)
	case warns > 0:
		fmt.Fprintf(&b, "%d visual warning(s) — see visual_problems and ascii_map.", warns)
	default:
		b.WriteString("No visual problems detected — layout looks clean from geometry.")
	}
	return strings.TrimSpace(b.String())
}

func sceneVisualProblems(sr scene.Report, minX, minY, maxX, maxY float64) []VisualProblem {
	var out []VisualProblem
	for _, iss := range sr.Issues {
		if iss.Code == diag.CodeSuggestCompact {
			continue
		}
		sev := "warning"
		if iss.Code == diag.CodeOccluded {
			sev = "error"
		}
		out = append(out, VisualProblem{
			Severity: sev,
			Code:     string(iss.Code),
			Message:  iss.Message,
			Fix:      iss.Fix,
			Where:    issueWhere(iss, nil, minX, minY, maxX, maxY),
			Involves: iss.Nodes,
		})
	}
	return out
}

func buildSummary(r Report) string {
	if len(r.Slides) == 0 {
		return "No slides to describe."
	}
	if len(r.Slides) == 1 {
		s := r.Slides[0]
		return fmt.Sprintf("1 diagram: %s", s.Narrative)
	}
	return fmt.Sprintf("%d slides described; read slides[].narrative and ascii_map per slide.", len(r.Slides))
}

func asciiMap(d model.Diagram, minX, minY, maxX, maxY float64, width int) string {
	w := width
	if w < 24 {
		w = 24
	}
	h := int(math.Max(8, math.Round((maxY-minY)/(maxX-minX+1)*float64(w/2))))
	if h > 24 {
		h = 24
	}
	grid := make([][]rune, h)
	for y := 0; y < h; y++ {
		grid[y] = make([]rune, w)
		for x := 0; x < w; x++ {
			grid[y][x] = ' '
		}
	}

	worldToCell := func(x, y float64) (int, int) {
		cx := int((x - minX) / (maxX - minX + 1) * float64(w-1))
		cy := int((y - minY) / (maxY - minY + 1) * float64(h-1))
		if cx < 0 {
			cx = 0
		}
		if cy < 0 {
			cy = 0
		}
		if cx >= w {
			cx = w - 1
		}
		if cy >= h {
			cy = h - 1
		}
		return cx, cy
	}

	// Edges as · or -
	for _, re := range d.Routed {
		for i := 1; i < len(re.Points); i++ {
			if len(re.Points[i]) < 2 || len(re.Points[i-1]) < 2 {
				continue
			}
			x1, y1 := re.Points[i-1][0], re.Points[i-1][1]
			x2, y2 := re.Points[i][0], re.Points[i][1]
			steps := int(math.Max(math.Abs(x2-x1), math.Abs(y2-y1)) / 20)
			if steps < 2 {
				steps = 2
			}
			for s := 0; s <= steps; s++ {
				t := float64(s) / float64(steps)
				x := x1 + (x2-x1)*t
				y := y1 + (y2-y1)*t
				cx, cy := worldToCell(x, y)
				if grid[cy][cx] == ' ' {
					if math.Abs(x2-x1) > math.Abs(y2-y1) {
						grid[cy][cx] = '·'
					} else {
						grid[cy][cx] = '│'
					}
				}
			}
		}
	}

	// Nodes as * + initial
	for _, n := range d.Nodes {
		if model.IsContainer(n.Kind) {
			continue
		}
		cx, cy := worldToCell(n.Rect.CX(), n.Rect.CY())
		if r := []rune(nodeInitial(n.ID)); len(r) > 0 {
			grid[cy][cx] = r[0]
			if len(r) > 1 && cx+1 < w {
				if grid[cy][cx+1] == ' ' || grid[cy][cx+1] == '·' || grid[cy][cx+1] == '│' {
					grid[cy][cx+1] = r[1]
				}
			}
		} else {
			grid[cy][cx] = '*'
		}
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("ASCII map (%d×%d cells, top-left = canvas top-left):\n", w, h))
	for y := 0; y < h; y++ {
		b.WriteString("  ")
		b.WriteString(string(grid[y]))
		b.WriteByte('\n')
	}
	b.WriteString("Legend: letter = node id initial, ·/│ = edge path, space = empty\n")
	return b.String()
}

func nodeInitial(id string) string {
	if id == "" {
		return "?"
	}
	return strings.ToUpper(id[:1])
}

func shortLabel(n model.Node) string {
	s := strings.ReplaceAll(n.Label, "\n", " ")
	if s == "" {
		return n.ID
	}
	if len(s) > 40 {
		return s[:37] + "..."
	}
	return s
}

func sortNodes(nodes []model.Node) []model.Node {
	out := append([]model.Node(nil), nodes...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Rect.X != out[j].Rect.X {
			return out[i].Rect.X < out[j].Rect.X
		}
		return out[i].Rect.Y < out[j].Rect.Y
	})
	return out
}

func sideOrDefault(s model.Side, def string) string {
	if s == "" {
		return def
	}
	return string(s)
}

func groupIssues(v diag.Report, slideCount int) [][]diag.Issue {
	if slideCount <= 1 {
		all := append(append([]diag.Issue{}, v.Errors...), v.Warnings...)
		return [][]diag.Issue{all}
	}
	out := make([][]diag.Issue, slideCount)
	for _, iss := range append(v.Errors, v.Warnings...) {
		idx := 0
		if strings.HasPrefix(iss.Message, "slide ") {
			var n int
			if _, err := fmt.Sscanf(iss.Message, "slide %d:", &n); err == nil && n >= 1 && n <= slideCount {
				idx = n - 1
			}
		}
		out[idx] = append(out[idx], iss)
	}
	return out
}

func filterCollisions(colls []model.Collision, d *model.Diagram) []model.Collision {
	if len(colls) == 0 {
		return nil
	}
	ids := map[string]bool{}
	for _, n := range d.Nodes {
		ids[n.ID] = true
	}
	var out []model.Collision
	for _, c := range colls {
		if ids[c.A] && ids[c.B] {
			out = append(out, c)
		}
	}
	return out
}

func countProblemsBySeverity(problems []VisualProblem) (errors, warnings int) {
	for _, p := range problems {
		if p.Severity == "error" {
			errors++
		} else {
			warnings++
		}
	}
	return errors, warnings
}

func countSeverity(problems []VisualProblem, code string) int {
	n := 0
	for _, p := range problems {
		if p.Code == code {
			n++
		}
	}
	return n
}

// WriteJSON emits the report.
func (r Report) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

// WriteHuman emits a readable description.
func (r Report) WriteHuman(w io.Writer) error {
	for _, s := range r.Slides {
		if len(r.Slides) > 1 {
			_, _ = fmt.Fprintf(w, "=== slide %d: %s ===\n\n", s.Index, s.Title)
		}
		_, _ = fmt.Fprintf(w, "%s\n\n", s.Narrative)
		if len(s.VisualProblems) > 0 {
			_, _ = io.WriteString(w, "visual problems:\n")
			for _, p := range s.VisualProblems {
				_, _ = fmt.Fprintf(w, "  [%s] %s\n    where: %s\n", p.Severity, p.Message, p.Where)
				if p.Fix != "" {
					_, _ = fmt.Fprintf(w, "    fix: %s\n", p.Fix)
				}
			}
			_, _ = io.WriteString(w, "\n")
		}
		_, _ = io.WriteString(w, s.ASCIIMap)
		_, _ = io.WriteString(w, "\nnodes:\n")
		for _, n := range s.Nodes {
			_, _ = fmt.Fprintf(w, "  %s [%s] %s — %s\n", n.ID, n.Kind, n.Label, n.Position)
		}
		_, _ = io.WriteString(w, "\nedges:\n")
		for _, e := range s.Edges {
			line := fmt.Sprintf("  %s → %s: %s", e.From, e.To, e.Route)
			if e.CrossesNode != "" {
				line += fmt.Sprintf(" (crosses %s)", e.CrossesNode)
			}
			_, _ = fmt.Fprintf(w, line+"\n")
		}
		if len(s.Relationships) > 0 {
			_, _ = io.WriteString(w, "\nrelationships:\n")
			for _, rel := range s.Relationships {
				_, _ = fmt.Fprintf(w, "  %s\n", rel)
			}
		}
		_, _ = io.WriteString(w, "\n")
	}
	_, _ = io.WriteString(w, r.Agent.Hint + "\n")
	return nil
}
