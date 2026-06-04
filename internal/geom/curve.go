package geom

import (
	"fmt"
	"math"
	"strings"

	"github.com/niklas-heer/sceno/internal/model"
)

// SmoothPath samples a Catmull-Rom spline through waypoints (organic / sketch-friendly).
func SmoothPath(pts []Point, samplesPerSeg int) []Point {
	if len(pts) < 2 {
		return pts
	}
	if len(pts) == 2 {
		return pts
	}
	if samplesPerSeg < 4 {
		samplesPerSeg = 4
	}
	var out []Point
	for i := 0; i < len(pts)-1; i++ {
		p0, p1, p2, p3 := pts[i], pts[i+1], pts[i+1], pts[i+1]
		if i > 0 {
			p0 = pts[i-1]
		}
		if i+2 < len(pts) {
			p3 = pts[i+2]
		}
		seg := sampleSegment(p0, p1, p2, p3, samplesPerSeg)
		if len(out) > 0 {
			seg = seg[1:]
		}
		out = append(out, seg...)
	}
	if len(out) > 0 {
		out[0] = pts[0]
		out[len(out)-1] = pts[len(pts)-1]
	}
	return out
}

func sampleSegment(p0, p1, p2, p3 Point, n int) []Point {
	out := make([]Point, 0, n+1)
	for i := 0; i <= n; i++ {
		t := float64(i) / float64(n)
		out = append(out, catmullRom(p0, p1, p2, p3, t))
	}
	return out
}

func catmullRom(p0, p1, p2, p3 Point, t float64) Point {
	t2 := t * t
	t3 := t2 * t
	return Point{
		X: 0.5 * ((2*p1.X) + (-p0.X+p2.X)*t + (2*p0.X-5*p1.X+4*p2.X-p3.X)*t2 + (-p0.X+3*p1.X-3*p2.X+p3.X)*t3),
		Y: 0.5 * ((2*p1.Y) + (-p0.Y+p2.Y)*t + (2*p0.Y-5*p1.Y+4*p2.Y-p3.Y)*t2 + (-p0.Y+3*p1.Y-3*p2.Y+p3.Y)*t3),
	}
}

// PathDSmooth builds an SVG path with quadratic smoothing.
func PathDSmooth(pts []Point) string {
	if len(pts) < 2 {
		return ""
	}
	if len(pts) == 2 {
		return fmt.Sprintf("M %.1f %.1f L %.1f %.1f", pts[0].X, pts[0].Y, pts[1].X, pts[1].Y)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "M %.1f %.1f", pts[0].X, pts[0].Y)
	for i := 1; i < len(pts)-1; i++ {
		mx := (pts[i].X + pts[i+1].X) / 2
		my := (pts[i].Y + pts[i+1].Y) / 2
		fmt.Fprintf(&b, " Q %.1f %.1f %.1f %.1f", pts[i].X, pts[i].Y, mx, my)
	}
	last := pts[len(pts)-1]
	fmt.Fprintf(&b, " T %.1f %.1f", last.X, last.Y)
	return b.String()
}

// PathDForStyle picks rounded orthogonal (polished) or smooth organic (sketch).
func PathDForStyle(pts []Point, sketch bool) string {
	if sketch && len(pts) >= 3 {
		return PathDSmooth(SmoothPath(pts, 6))
	}
	return RoundedPathD(pts, EdgeCornerRadius)
}

// PathVisibleFraction estimates how much of a polyline is outside node obstacles.
func PathVisibleFraction(pts []Point, fromID, toID string, nodes []model.Node, pad float64) float64 {
	if len(pts) < 2 {
		return 1
	}
	var visible, total float64
	for i := 1; i < len(pts); i++ {
		segLen := math.Hypot(pts[i].X-pts[i-1].X, pts[i].Y-pts[i-1].Y)
		if segLen < 1e-6 {
			continue
		}
		total += segLen
		samples := int(segLen/10) + 2
		for s := 0; s <= samples; s++ {
			t := float64(s) / float64(samples)
			p := Point{
				X: pts[i-1].X + (pts[i].X-pts[i-1].X)*t,
				Y: pts[i-1].Y + (pts[i].Y-pts[i-1].Y)*t,
			}
			if !pointHitsObstacle(p, fromID, toID, nodes, pad) {
				visible += segLen / float64(samples+1)
			}
		}
	}
	if total < 1e-6 {
		return 1
	}
	if visible > total {
		return 1
	}
	return visible / total
}

func pointHitsObstacle(p Point, fromID, toID string, nodes []model.Node, pad float64) bool {
	for _, n := range nodes {
		if n.ID == fromID || n.ID == toID || model.IsContainer(n.Kind) {
			continue
		}
		r := PadRect(n.Rect, pad*0.35)
		if pointInRect(p, r) {
			return true
		}
	}
	return false
}

// RectOverlapArea returns intersection area of two rects (0 if none).
func RectOverlapArea(a, b model.Rect) float64 {
	x1 := math.Max(a.X, b.X)
	y1 := math.Max(a.Y, b.Y)
	x2 := math.Min(a.Right(), b.Right())
	y2 := math.Min(a.Bottom(), b.Bottom())
	if x2 <= x1 || y2 <= y1 {
		return 0
	}
	return (x2 - x1) * (y2 - y1)
}
