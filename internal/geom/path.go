package geom

import "math"

// ArrowTipLength is how far to shorten the last segment for marker geometry.
const ArrowTipLength = 11.0

// ShortenEnd moves point b toward a by dist (for arrowhead clearance).
func ShortenEnd(a, b Point, dist float64) Point {
	dx := b.X - a.X
	dy := b.Y - a.Y
	l := math.Hypot(dx, dy)
	if l <= dist || l < 1e-6 {
		return a
	}
	t := (l - dist) / l
	return Point{X: a.X + dx*t, Y: a.Y + dy*t}
}

// SimplifyPath removes collinear intermediate points.
func SimplifyPath(pts []Point) []Point {
	if len(pts) <= 2 {
		return pts
	}
	out := []Point{pts[0]}
	for i := 1; i < len(pts)-1; i++ {
		if !collinear(out[len(out)-1], pts[i], pts[i+1]) {
			out = append(out, pts[i])
		}
	}
	out = append(out, pts[len(pts)-1])
	return out
}

func collinear(a, b, c Point) bool {
	const eps = 0.5
	// Same horizontal line
	if math.Abs(a.Y-b.Y) < eps && math.Abs(b.Y-c.Y) < eps {
		return true
	}
	// Same vertical line
	if math.Abs(a.X-b.X) < eps && math.Abs(b.X-c.X) < eps {
		return true
	}
	return false
}

// TrimArrowEnd shortens the path terminus so arrow markers sit on the border.
func TrimArrowEnd(pts []Point) []Point {
	if len(pts) < 2 {
		return pts
	}
	out := make([]Point, len(pts))
	copy(out, pts)
	last := len(out) - 1
	out[last] = ShortenEnd(out[last-1], out[last], ArrowTipLength)
	return out
}

// PathToSlices converts points to [][]float64.
func PathToSlices(pts []Point) [][]float64 {
	out := make([][]float64, len(pts))
	for i, p := range pts {
		out[i] = []float64{p.X, p.Y}
	}
	return out
}

// SlicesToPath converts [][]float64 to points.
func SlicesToPath(path [][]float64) []Point {
	pts := make([]Point, 0, len(path))
	for _, p := range path {
		if len(p) >= 2 {
			pts = append(pts, Point{X: p[0], Y: p[1]})
		}
	}
	return pts
}
