package render

import (
	"fmt"
	"math"
	"strings"

	"github.com/niklas-heer/sceno/internal/geom"
	"github.com/niklas-heer/sceno/internal/model"
)

// SVG renders a hand-drawn-style SVG (Excalidraw-adjacent).
func SVG(d model.Diagram) string {
	minX, minY, maxX, maxY := Bounds(d)
	w := maxX - minX
	h := maxY - minY
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="%.1f %.1f %.1f %.1f" width="%.0f" height="%.0f">`, minX, minY, w, h, w, h)
	b.WriteString(`<defs>`)
	b.WriteString(SVGFontDefs())
	b.WriteString(`<marker id="arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse"><path d="M 0 0 L 10 5 L 0 10 z" fill="#1e1e1e"/></marker></defs>`)
	b.WriteString(`<rect x="` + fmt.Sprintf("%.1f", minX) + `" y="` + fmt.Sprintf("%.1f", minY) + `" width="` + fmt.Sprintf("%.1f", w) + `" height="` + fmt.Sprintf("%.1f", h) + `" fill="#ffffff"/>`)
	if d.Title != "" {
		b.WriteString(textEl(d.Title, (minX+maxX)/2-200, minY+28, 28, "#7048e8", "bold"))
	}
	if d.Subtitle != "" {
		b.WriteString(textEl(d.Subtitle, (minX+maxX)/2-220, minY+58, 16, "#495057", ""))
	}
	// Paint order: lanes → edges → nodes (edges stay visible under shapes).
	for _, n := range d.Nodes {
		if n.Kind == model.ShapeLane {
			b.WriteString(nodeSketch(n, minX, minY))
		}
	}
	for _, re := range d.Routed {
		lctx := LabelContext(d, re.Edge)
		b.WriteString(pathSketch(re.Points, re.Edge))
		b.WriteString(EdgeLabelSketch(re.Points, re.Edge, lctx))
	}
	for _, n := range d.Nodes {
		if n.Kind != model.ShapeLane {
			b.WriteString(nodeSketch(n, minX, minY))
		}
	}
	if len(d.Routed) == 0 {
		for key, path := range d.EdgePaths {
			b.WriteString(pathSketch(path, findEdge(d, key)))
		}
	}
	b.WriteString(`</svg>`)
	return b.String()
}

func nodeSketch(n model.Node, _, _ float64) string {
	fill := n.Fill
	if fill == "" {
		fill = "#e9ecef"
	}
	stroke := n.Stroke
	if stroke == "" {
		stroke = "#1e1e1e"
	}
	var shape string
	switch n.Kind {
	case model.ShapeEllipse:
		shape = wobbleEllipse(n.Rect, stroke, fill)
	default:
		shape = wobbleRect(n.Rect, stroke, fill, n.Kind == model.ShapeLane)
	}
	label := labelSketch(n)
	if n.Kind == model.ShapeLane && n.Label != "" {
		laneTitle := textEl(n.Label, n.Rect.X+12, n.Rect.Y+10, 13, "#495057", "")
		return shape + laneTitle
	}
	return shape + label
}

func labelSketch(n model.Node) string {
	lines := strings.Split(n.Label, "\n")
	if len(lines) == 0 {
		return ""
	}
	lineH := n.FontSize * 1.25
	totalH := float64(len(lines)) * lineH
	startY := n.Rect.CY() - totalH/2 + lineH*0.75
	var b strings.Builder
	for i, line := range lines {
		tw := float64(len(line)) * n.FontSize * 0.55
		x := n.Rect.CX() - tw/2
		y := startY + float64(i)*lineH
		b.WriteString(textEl(line, x, y, n.FontSize, "#1e1e1e", ""))
	}
	return b.String()
}

func wobbleRect(r model.Rect, stroke, fill string, lane bool) string {
	points := wobblePolygon([][2]float64{
		{r.X, r.Y},
		{r.Right(), r.Y},
		{r.Right(), r.Bottom()},
		{r.X, r.Bottom()},
		{r.X, r.Y},
	}, r.X+r.Y)
	d := pathData(points)
	opacity := ""
	if lane {
		opacity = ` opacity="0.55"`
	}
	return fmt.Sprintf(`<path d="%s" fill="%s" stroke="%s" stroke-width="2" fill-opacity="1"%s/>`, d, fill, stroke, opacity)
}

func wobbleEllipse(r model.Rect, stroke, fill string) string {
	cx, cy := r.CX(), r.CY()
	rx, ry := r.W/2, r.H/2
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`<ellipse cx="%.1f" cy="%.1f" rx="%.1f" ry="%.1f" fill="%s" stroke="%s" stroke-width="2"/>`, cx, cy, rx, ry, fill, stroke))
	return b.String()
}

func pathSketch(pts [][]float64, e model.Edge) string {
	if len(pts) < 2 {
		return ""
	}
	gpts := geom.SlicesToPath(pts)
	if len(gpts) >= 3 {
		gpts = geom.SmoothPath(gpts, 6)
	}
	flat := make([][2]float64, len(gpts))
	for i, p := range gpts {
		flat[i] = [2]float64{p.X, p.Y}
	}
	seed := gpts[0].X + gpts[0].Y
	wo := wobblePolyline(flat, seed)
	d := pathData(wo)
	if len(gpts) >= 3 {
		d = geom.PathDSmooth(wobbleToPoints(wo))
	}
	stroke := e.Color
	if stroke == "" {
		stroke = "#1e1e1e"
	}
	dash := ""
	if e.Dashed {
		dash = ` stroke-dasharray="8 6"`
	}
	return fmt.Sprintf(`<path d="%s" fill="none" stroke="%s" stroke-width="2"%s marker-end="url(#arrow)"/>`, d, stroke, dash)
}

// Rough.js-lite wobble via seeded sin noise.
func wobblePolygon(pts [][2]float64, seed float64) [][2]float64 {
	out := make([][2]float64, len(pts))
	for i, p := range pts {
		jitter := 2.5 * math.Sin(seed+float64(i)*1.7)
		out[i] = [2]float64{p[0] + jitter, p[1] + jitter*0.7}
	}
	return out
}

func wobblePolyline(pts [][2]float64, seed float64) [][2]float64 {
	return wobblePolygon(pts, seed)
}

func wobbleToPoints(pts [][2]float64) []geom.Point {
	out := make([]geom.Point, len(pts))
	for i, p := range pts {
		out[i] = geom.Point{X: p[0], Y: p[1]}
	}
	return out
}

func pathData(pts [][2]float64) string {
	if len(pts) == 0 {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "M %.1f %.1f", pts[0][0], pts[0][1])
	for i := 1; i < len(pts); i++ {
		fmt.Fprintf(&b, " L %.1f %.1f", pts[i][0], pts[i][1])
	}
	return b.String()
}

