package geom

import (
	"fmt"
	"math"
	"strings"
)

// EdgeCornerRadius is the default bend radius for orthogonal connectors.
const EdgeCornerRadius = 10.0

// RoundedPathD returns an SVG path with quadratic fillets at each bend.
func RoundedPathD(pts []Point, radius float64) string {
	n := len(pts)
	if n < 2 {
		return ""
	}
	if n == 2 {
		return fmt.Sprintf("M %.2f %.2f L %.2f %.2f", pts[0].X, pts[0].Y, pts[1].X, pts[1].Y)
	}
	r := radius
	if r <= 0 {
		r = EdgeCornerRadius
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("M %.2f %.2f", pts[0].X, pts[0].Y))

	for i := 1; i < n-1; i++ {
		p0, p1, p2 := pts[i-1], pts[i], pts[i+1]
		v1x, v1y := p1.X-p0.X, p1.Y-p0.Y
		v2x, v2y := p2.X-p1.X, p2.Y-p1.Y
		len1 := math.Hypot(v1x, v1y)
		len2 := math.Hypot(v2x, v2y)
		if len1 < 1e-6 || len2 < 1e-6 {
			continue
		}
		rc := math.Min(r, math.Min(len1/2, len2/2))
		xA := p1.X - v1x/len1*rc
		yA := p1.Y - v1y/len1*rc
		xB := p1.X + v2x/len2*rc
		yB := p1.Y + v2y/len2*rc
		if i == 1 {
			fmt.Fprintf(&b, " L %.2f %.2f", xA, yA)
		}
		fmt.Fprintf(&b, " Q %.2f %.2f %.2f %.2f", p1.X, p1.Y, xB, yB)
	}
	last := pts[n-1]
	fmt.Fprintf(&b, " L %.2f %.2f", last.X, last.Y)
	return b.String()
}

// RoundedPathSamples approximates a rounded path as line segments (for raster backends).
func RoundedPathSamples(pts []Point, radius float64, segmentsPerCorner int) []Point {
	d := RoundedPathD(pts, radius)
	if d == "" {
		return pts
	}
	_ = d
	// Fallback: return simplified polyline; GG/PDF use round joins on this path.
	return SimplifyPath(pts)
}
