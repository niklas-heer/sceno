package render

import (
	"fmt"
	"strings"

	"github.com/niklas-heer/sceno/internal/geom"
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
	// Lucide-style open chevron — reads cleaner than a filled triangle.
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

func polishedPath(pts [][]float64, e model.Edge) string {
	if len(pts) < 2 {
		return ""
	}
	gpts := geom.SlicesToPath(pts)
	gpts = geom.SimplifyPath(gpts)
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
	marker := arrowMarkerID(stroke)
	return fmt.Sprintf(`<path d="%s" fill="none" stroke="%s" stroke-width="%.2f" stroke-opacity="%s" stroke-linejoin="round" stroke-linecap="round"%s marker-end="url(#%s)"/>`,
		pathD, stroke, theme.EdgeWidth, paint.EdgeOpacity, dash, marker)
}

// EdgeLabelSVG renders a readable label on the longest edge segment.
func EdgeLabelSVG(pts [][]float64, e model.Edge) string {
	label := strings.TrimSpace(e.Label)
	if label == "" || len(pts) < 2 {
		return ""
	}
	gpts := geom.SimplifyPath(geom.SlicesToPath(pts))
	if len(gpts) < 2 {
		return ""
	}
	x, y, horiz := geom.LabelPlacement(gpts)
	lines := strings.Split(label, "\n")
	fontSize := float64(theme.SubSize)
	lineH := fontSize * 1.35
	maxW := 0.0
	for _, line := range lines {
		if w := float64(len(line)) * fontSize * 0.58; w > maxW {
			maxW = w
		}
	}
	if maxW < 24 {
		maxW = 24
	}
	padX, padY := 6.0, 4.0
	boxW := maxW + padX*2
	boxH := float64(len(lines))*lineH + padY*2 - (lineH - fontSize)
	rx, ry := x, y
	if horiz {
		ry -= boxH/2 + 6
	} else {
		rx += boxH/2 + 6
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="4" fill="%s" stroke="%s" stroke-width="1"/>`,
		rx-boxW/2, ry-boxH/2, boxW, boxH, paint.BgCard, paint.Border))
	textY := ry - boxH/2 + padY + fontSize*0.85
	for i, line := range lines {
		lineW := float64(len(line)) * fontSize * 0.58
		b.WriteString(textEl(line, rx-lineW/2, textY+float64(i)*lineH, fontSize, paint.FgMuted, "500"))
	}
	return b.String()
}

// EdgeLabelSketch renders a hand-drawn style edge label.
func EdgeLabelSketch(pts [][]float64, e model.Edge) string {
	return EdgeLabelSVG(pts, e)
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
