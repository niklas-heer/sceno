package geom

import (
	"math"

	"github.com/niklas-heer/sceno/internal/model"
)

// ArrowTipLength is an alias for ArrowHeadDepth (legacy name).
const ArrowTipLength = ArrowHeadDepth

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

// CollapseJogs removes tiny orthogonal staircase segments and fold-backs that
// endpoint snapping and lane shifts leave behind. Endpoints are shape anchors
// and never move; interior runs shift by less than tol to absorb the jog.
func CollapseJogs(pts []Point, tol float64) []Point {
	out := SimplifyPath(append([]Point(nil), pts...))
	for iter := 0; iter < 32; iter++ {
		if removeFoldBack(out) {
			out = SimplifyPath(out[:len(out)-1])
			continue
		}
		if !shiftTinyJog(out, tol) {
			break
		}
		out = SimplifyPath(out)
	}
	return out
}

// removeFoldBack drops the middle point of two collinear segments that reverse
// direction; the retraced portion overlaps the direct segment exactly, so the
// drawn line only gets cleaner. Compacts in place, returns true when found.
func removeFoldBack(out []Point) bool {
	const eps = 0.5
	for i := 1; i < len(out)-1; i++ {
		a, b, c := out[i-1], out[i], out[i+1]
		horizontal := math.Abs(a.Y-b.Y) < eps && math.Abs(b.Y-c.Y) < eps && (b.X-a.X)*(c.X-b.X) < 0
		vertical := math.Abs(a.X-b.X) < eps && math.Abs(b.X-c.X) < eps && (b.Y-a.Y)*(c.Y-b.Y) < 0
		if horizontal || vertical {
			copy(out[i:], out[i+1:])
			return true
		}
	}
	return false
}

// shiftTinyJog aligns the runs around one interior segment shorter than tol.
// Anchored endpoints never move, and jogs near the path end shift the previous
// run so the straight approach before the arrowhead keeps its length.
func shiftTinyJog(out []Point, tol float64) bool {
	const eps = 0.5
	for i := 1; i < len(out)-2; i++ {
		a, b := out[i], out[i+1]
		dx, dy := math.Abs(b.X-a.X), math.Abs(b.Y-a.Y)
		vertical := dx < eps && dy > eps && dy < tol
		horizontal := dy < eps && dx > eps && dx < tol
		if !vertical && !horizontal {
			continue
		}
		coord := func(p *Point) *float64 {
			if vertical {
				return &p.Y
			}
			return &p.X
		}
		shiftPrev := func() bool {
			if i-1 <= 0 {
				return false
			}
			*coord(&out[i]) = *coord(&b)
			*coord(&out[i-1]) = *coord(&b)
			return true
		}
		shiftNext := func() bool {
			if i+2 < len(out)-1 {
				*coord(&out[i+1]) = *coord(&a)
				*coord(&out[i+2]) = *coord(&a)
				return true
			}
			if i+2 == len(out)-1 && math.Abs(*coord(&out[i+2])-*coord(&a)) < eps {
				*coord(&out[i+1]) = *coord(&a)
				return true
			}
			return false
		}
		if i+2 >= len(out)-2 { // jog near the target: keep the final approach intact
			if shiftPrev() || shiftNext() {
				return true
			}
		} else if shiftNext() || shiftPrev() {
			return true
		}
	}
	return false
}

func collinear(a, b, c Point) bool {
	const eps = 0.5
	// Same horizontal line
	if math.Abs(a.Y-b.Y) < eps && math.Abs(b.Y-c.Y) < eps {
		return (b.X-a.X)*(c.X-b.X) >= 0
	}
	// Same vertical line
	if math.Abs(a.X-b.X) < eps && math.Abs(b.X-c.X) < eps {
		return (b.Y-a.Y)*(c.Y-b.Y) >= 0
	}
	return false
}

// TrimArrowEnd shortens the path to the stroke end (tip remains at the original anchor).
func TrimArrowEnd(pts []Point) []Point {
	if len(pts) < 2 {
		return pts
	}
	strokeEnd, segmentIndex := pointBeforePathEnd(pts, ArrowHeadDepth)
	out := append([]Point(nil), pts[:segmentIndex]...)
	if len(out) == 0 || out[len(out)-1] != strokeEnd {
		out = append(out, strokeEnd)
	}
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
	Avoid    []EdgeLabelObstacle
}

// EdgeLabelObstacle is chrome or node geometry that a label pill must clear.
type EdgeLabelObstacle struct {
	ID     string
	Kind   string
	Bounds model.Rect
}

// EdgeLabelBox returns the label center and size on the best edge segment.
// Prefer LayoutEdgeLabel — this wrapper keeps legacy call sites working.
func EdgeLabelBox(pts []Point, padX, padY, lineH, fontSize float64, lines []string, maxTextW float64, ctx *EdgeLabelContext) (rx, ry, boxW, boxH float64, horizontal bool) {
	_ = padX
	_ = padY
	return edgeLabelBoxInner(pts, fontSize, lineH, lines, maxTextW, ctx)
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
