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
