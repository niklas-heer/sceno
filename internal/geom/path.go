package geom

import (
	"math"

	"github.com/niklas-heer/sceno/internal/model"
)

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

// LabelPlacement picks the midpoint of the longest segment for edge labels.
func LabelPlacement(pts []Point) (x, y float64, horizontal bool) {
	if len(pts) < 2 {
		return 0, 0, true
	}
	bestLen := -1.0
	var mid Point
	horiz := true
	for i := 1; i < len(pts); i++ {
		a, b := pts[i-1], pts[i]
		dx, dy := b.X-a.X, b.Y-a.Y
		l := math.Hypot(dx, dy)
		if l > bestLen {
			bestLen = l
			mid = Point{X: (a.X + b.X) / 2, Y: (a.Y + b.Y) / 2}
			horiz = math.Abs(dx) >= math.Abs(dy)
		}
	}
	return mid.X, mid.Y, horiz
}

// EdgeLabelContext supplies endpoint nodes so labels clear shapes and sit in the gap.
type EdgeLabelContext struct {
	From, To model.Rect
}

// EdgeLabelBox returns the label center and size on the best edge segment.
// When ctx is set, horizontal labels center in the node gap and sit above both shapes.
func EdgeLabelBox(pts []Point, padX, padY, lineH, fontSize float64, lines []string, maxTextW float64, ctx *EdgeLabelContext) (rx, ry, boxW, boxH float64, horizontal bool) {
	if len(pts) < 2 || len(lines) == 0 {
		return 0, 0, 0, 0, true
	}
	x, y, horiz := LabelPlacement(pts)
	if maxTextW < 24 {
		maxTextW = 24
	}
	boxW = maxTextW + padX*2
	boxH = float64(len(lines))*lineH + padY*2 - (lineH - fontSize)
	const gap = 12.0
	if horiz {
		rx = x
		if ctx != nil {
			gapLeft := ctx.From.Right() + 6
			gapRight := ctx.To.X - 6
			if gapRight > gapLeft {
				rx = (gapLeft + gapRight) / 2
			}
			minTop := ctx.From.Y
			if ctx.To.Y < minTop {
				minTop = ctx.To.Y
			}
			ry = minTop - gap - boxH/2
		} else {
			ry = y - boxH/2 - gap
		}
		return rx, ry, boxW, boxH, true
	}
	rx = x + boxW/2 + gap
	ry = y
	return rx, ry, boxW, boxH, false
}

// LabelBoxRect returns the axis-aligned bounds for a label box.
func LabelBoxRect(rx, ry, boxW, boxH float64) model.Rect {
	return model.Rect{X: rx - boxW/2, Y: ry - boxH/2, W: boxW, H: boxH}
}

// SplitPathForLabel breaks a path around a horizontal label box (gap in the connector).
func SplitPathForLabel(pts []Point, box model.Rect) [][]Point {
	if len(pts) < 2 || box.W <= 0 {
		return [][]Point{pts}
	}
	gpts := SimplifyPath(pts)
	if len(gpts) < 2 {
		return [][]Point{pts}
	}
	var out [][]Point
	for i := 1; i < len(gpts); i++ {
		a, b := gpts[i-1], gpts[i]
		if math.Abs(a.Y-b.Y) > 1 || math.Abs(b.X-a.X) < 1 {
			out = append(out, []Point{a, b})
			continue
		}
		// Horizontal segment — split around label x span
		left, right := a, b
		if left.X > right.X {
			left, right = right, left
		}
		pad := 4.0
		lx := box.X - pad
		rx := box.Right() + pad
		if rx <= left.X || lx >= right.X {
			out = append(out, []Point{a, b})
			continue
		}
		if lx > left.X {
			out = append(out, []Point{a, Point{X: lx, Y: a.Y}})
		}
		if rx < right.X {
			out = append(out, []Point{Point{X: rx, Y: b.Y}, b})
		}
	}
	if len(out) == 0 {
		return [][]Point{pts}
	}
	return out
}
