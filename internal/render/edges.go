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

// SVGArrowMarkers emits arrow markers for each edge stroke color used.
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
	return fmt.Sprintf(`<marker id="%s" viewBox="0 0 12 12" refX="10" refY="6" markerWidth="%.1f" markerHeight="%.1f" orient="auto" markerUnits="userSpaceOnUse"><path d="M 2 3 L 8 6 L 2 9" fill="none" stroke="%s" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"/></marker>`,
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
	for i, seg := range segments {
		b.WriteString(polishedPathSegment(seg, e, i == len(segments)-1))
	}
	return b.String()
}

func edgePathSegments(pts [][]float64, e model.Edge, ctx *geom.EdgeLabelContext) [][]geom.Point {
	gpts := geom.SimplifyPath(geom.SlicesToPath(pts))
	if len(gpts) < 2 {
		return nil
	}
	label := strings.TrimSpace(e.Label)
	if label == "" || ctx == nil {
		return [][]geom.Point{gpts}
	}
	lines := strings.Split(label, "\n")
	fontSize := float64(theme.SubSize)
	lineH := fontSize * 1.35
	maxW := 0.0
	for _, line := range lines {
		w := measure.TextWidth(line, fontSize, fonts.WeightMedium)
		if w > maxW {
			maxW = w
		}
	}
	rx, ry, boxW, boxH, horiz := geom.EdgeLabelBox(gpts, 6, 4, lineH, fontSize, lines, maxW, ctx)
	if !horiz {
		return [][]geom.Point{gpts}
	}
	box := geom.LabelBoxRect(rx, ry, boxW, boxH)
	return geom.SplitPathForLabel(gpts, box)
}

func polishedPathSegment(pts []geom.Point, e model.Edge, withArrow bool) string {
	if len(pts) < 2 {
		return ""
	}
	gpts := geom.SimplifyPath(pts)
	if withArrow {
		gpts = geom.TrimArrowEnd(gpts)
	}
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
	marker := ""
	if withArrow {
		marker = fmt.Sprintf(` marker-end="url(#%s)"`, arrowMarkerID(stroke))
	}
	return fmt.Sprintf(`<path d="%s" fill="none" stroke="%s" stroke-width="%.2f" stroke-opacity="%s" stroke-linejoin="round" stroke-linecap="round"%s%s/>`,
		pathD, stroke, theme.EdgeWidth, paint.EdgeOpacity, dash, marker)
}

// EdgeLabelSVG renders a readable label on the longest edge segment.
func EdgeLabelSVG(pts [][]float64, e model.Edge, ctx *geom.EdgeLabelContext) string {
	label := strings.TrimSpace(e.Label)
	if label == "" || len(pts) < 2 {
		return ""
	}
	gpts := geom.SimplifyPath(geom.SlicesToPath(pts))
	if len(gpts) < 2 {
		return ""
	}
	lines := strings.Split(label, "\n")
	fontSize := float64(theme.SubSize)
	lineH := fontSize * 1.35
	maxW := 0.0
	for _, line := range lines {
		w := measure.TextWidth(line, fontSize, fonts.WeightMedium)
		if w > maxW {
			maxW = w
		}
	}
	padX, padY := 6.0, 4.0
	rx, ry, boxW, boxH, _ := geom.EdgeLabelBox(gpts, padX, padY, lineH, fontSize, lines, maxW, ctx)
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="4" fill="%s" stroke="%s" stroke-width="1"/>`,
		rx-boxW/2, ry-boxH/2, boxW, boxH, paint.BgCard, paint.Border))
	textY := ry - boxH/2 + padY + fontSize*0.85
	for i, line := range lines {
		lineW := measure.TextWidth(line, fontSize, fonts.WeightMedium)
		b.WriteString(textEl(line, rx-lineW/2, textY+float64(i)*lineH, fontSize, paint.FgMuted, "500"))
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
