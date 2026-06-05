package render

import (
	"fmt"
	"strings"

	"github.com/niklas-heer/sceno/internal/fonts"
	"github.com/niklas-heer/sceno/internal/geom"
	"github.com/niklas-heer/sceno/internal/measure"
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/theme"
)

// SVGArrowMarkers emits arrow markers for sketch style exports.
func SVGArrowMarkers(d model.Diagram) string {
	seen := map[string]struct{}{paint.EdgeDefault: {}}
	for _, re := range d.Routed {
		c := re.Edge.Color
		if c == "" {
			c = paint.EdgeDefault
		}
		seen[c] = struct{}{}
	}
	var b strings.Builder
	for c := range seen {
		b.WriteString(svgArrowMarker(c))
	}
	return b.String()
}

func svgArrowMarker(color string) string {
	id := arrowMarkerID(color)
	return fmt.Sprintf(`<marker id="%s" viewBox="0 0 10 10" refX="0" refY="5" markerWidth="%.1f" markerHeight="%.1f" orient="auto" markerUnits="userSpaceOnUse"><path d="M 0 1.5 L 9 5 L 0 8.5 Z" fill="%s"/></marker>`,
		id, theme.ArrowMarkerSize, theme.ArrowMarkerSize, color)
}

func arrowMarkerID(color string) string {
	if color == "" || color == paint.EdgeDefault {
		return "arrow"
	}
	s := strings.TrimPrefix(color, "#")
	for _, r := range s {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
			continue
		}
		return "arrow-custom"
	}
	return "arrow-" + strings.ToLower(s)
}

// LabelContext returns endpoint rects for edge label placement.
func LabelContext(d model.Diagram, e model.Edge) *geom.EdgeLabelContext {
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

func polishedPath(pts [][]float64, e model.Edge, ctx *geom.EdgeLabelContext) string {
	segments := edgePathSegments(pts, e, ctx)
	var b strings.Builder
	for _, seg := range segments {
		b.WriteString(polishedStrokeSegment(seg, e))
	}
	return b.String()
}

func edgePathSegments(pts [][]float64, e model.Edge, ctx *geom.EdgeLabelContext) [][]geom.Point {
	gpts := geom.SimplifyPath(geom.SlicesToPath(pts))
	if len(gpts) < 2 {
		return nil
	}
	return [][]geom.Point{gpts}
}

func polishedStrokeSegment(pts []geom.Point, e model.Edge) string {
	if len(pts) < 2 {
		return ""
	}
	gpts := geom.SimplifyPath(pts)
	gpts = geom.TrimArrowEnd(gpts)
	if len(gpts) < 2 {
		return ""
	}
	pathD := geom.RoundedPathD(gpts, theme.EdgeCorner)
	stroke := e.Color
	if stroke == "" {
		stroke = paint.EdgeDefault
	}
	dash := ""
	if e.Dashed {
		dash = ` stroke-dasharray="5 4"`
	}
	return fmt.Sprintf(`<path d="%s" fill="none" stroke="%s" stroke-width="%.2f" stroke-opacity="%s" stroke-linejoin="round" stroke-linecap="round"%s/>`,
		pathD, stroke, theme.EdgeWidth, paint.EdgeOpacity, dash)
}

// ArrowHeadSVG draws a filled arrowhead with tip on the target border (paint after nodes).
func ArrowHeadSVG(pts [][]float64, e model.Edge) string {
	gpts := geom.SimplifyPath(geom.SlicesToPath(pts))
	ag, ok := geom.ArrowGeometryForPath(gpts)
	if !ok {
		return ""
	}
	stroke := e.Color
	if stroke == "" {
		stroke = paint.FgMuted
	}
	t1, t2, t3 := geom.ArrowHeadPoints(ag.Prev, ag.Tip, theme.ArrowMarkerSize)
	return fmt.Sprintf(`<polygon points="%.2f,%.2f %.2f,%.2f %.2f,%.2f" fill="%s"/>`,
		t1.X, t1.Y, t2.X, t2.Y, t3.X, t3.Y, stroke)
}

// EdgeLabelSVG renders a readable label on the longest edge segment.
func EdgeLabelSVG(pts [][]float64, e model.Edge, ctx *geom.EdgeLabelContext) string {
	gpts := geom.SimplifyPath(geom.SlicesToPath(pts))
	if len(gpts) < 2 {
		return ""
	}
	layout := geom.LayoutEdgeLabel(gpts, e.Label, ctx)
	if layout.BoxW <= 0 {
		return ""
	}
	x, y, boxW, boxH := layout.LabelRect()
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="4" fill="%s" stroke="%s" stroke-width="1"/>`,
		x, y, boxW, boxH, paint.BgCard, paint.Border))
	for i, line := range layout.Lines {
		lineW := measure.TextWidth(line, layout.FontSize, fonts.WeightMedium)
		b.WriteString(textEl(line, layout.CenterX-lineW/2, layout.TextBaselineY(i), layout.FontSize, paint.FgMuted, "500"))
	}
	return b.String()
}

// EdgeLabelSketch renders a hand-drawn style edge label.
func EdgeLabelSketch(pts [][]float64, e model.Edge, ctx *geom.EdgeLabelContext) string {
	return EdgeLabelSVG(pts, e, ctx)
}

func findEdge(d model.Diagram, key string) model.Edge {
	for i, e := range d.Edges {
		k := fmt.Sprintf("%s-%s-%d", e.From, e.To, i)
		if k == key {
			return e
		}
	}
	if len(d.Edges) > 0 {
		return d.Edges[0]
	}
	return model.Edge{}
}
