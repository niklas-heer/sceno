package geom

import "math"

// ArrowHeadDepth is how far the arrowhead extends from stroke end to the target border.
const ArrowHeadDepth = 9.0

// MaxArrowTipGap is the maximum allowed distance between arrow tip and target anchor.
const MaxArrowTipGap = 2.0

// ArrowGeometry is stroke end + tip for one connector (shared by render and validation).
type ArrowGeometry struct {
	StrokeEnd       Point   // visible line terminus
	Tip             Point   // arrow tip on target border
	StartApproach   Point   // direction reference after the source anchor
	Prev            Point   // direction reference before the arrowhead
	VisibleApproach float64 // straight/smoothed shaft available before the head
}

// ArrowGeometryForPath computes render/validate arrow placement from a routed path.
// The path terminus must be the target anchor on the shape border.
func ArrowGeometryForPath(pts []Point) (ArrowGeometry, bool) {
	gpts := SimplifyPath(pts)
	if len(gpts) < 2 {
		return ArrowGeometry{}, false
	}
	tip := gpts[len(gpts)-1]
	strokeEnd, _ := pointBeforePathEnd(gpts, ArrowHeadDepth)
	prev, _ := pointBeforePathEnd(gpts, ArrowHeadDepth+EdgeLabelClearRun)
	startApproach, _ := pointAfterPathStart(gpts, ArrowHeadDepth+EdgeLabelClearRun)
	lastSegment := math.Hypot(tip.X-gpts[len(gpts)-2].X, tip.Y-gpts[len(gpts)-2].Y)
	visible := math.Max(0, lastSegment-ArrowHeadDepth)
	if len(gpts) > 8 && lastSegment < ArrowHeadDepth {
		// Smoothed paths contain dense samples; use path distance, not one sample.
		visible = math.Min(EdgeLabelClearRun, math.Max(0, pathLength(gpts)-ArrowHeadDepth))
	}
	return ArrowGeometry{StrokeEnd: strokeEnd, Tip: tip, StartApproach: startApproach, Prev: prev, VisibleApproach: visible}, true
}

func pointAfterPathStart(pts []Point, distance float64) (Point, int) {
	remaining := distance
	for i := 1; i < len(pts); i++ {
		a, b := pts[i-1], pts[i]
		segment := math.Hypot(b.X-a.X, b.Y-a.Y)
		if segment < 1e-9 {
			continue
		}
		if remaining <= segment {
			ratio := remaining / segment
			return Point{X: a.X + (b.X-a.X)*ratio, Y: a.Y + (b.Y-a.Y)*ratio}, i
		}
		remaining -= segment
	}
	return pts[len(pts)-1], len(pts) - 1
}

func pointBeforePathEnd(pts []Point, distance float64) (Point, int) {
	remaining := distance
	for i := len(pts) - 1; i > 0; i-- {
		a, b := pts[i-1], pts[i]
		segment := math.Hypot(b.X-a.X, b.Y-a.Y)
		if segment < 1e-9 {
			continue
		}
		if remaining <= segment {
			ratio := remaining / segment
			return Point{X: b.X + (a.X-b.X)*ratio, Y: b.Y + (a.Y-b.Y)*ratio}, i
		}
		remaining -= segment
	}
	return pts[0], 1
}

func pathLength(pts []Point) float64 {
	length := 0.0
	for i := 1; i < len(pts); i++ {
		length += math.Hypot(pts[i].X-pts[i-1].X, pts[i].Y-pts[i-1].Y)
	}
	return length
}

// TipGap returns how far the arrow tip sits from the intended target anchor.
func TipGap(tip, targetAnchor Point) float64 {
	return math.Hypot(tip.X-targetAnchor.X, tip.Y-targetAnchor.Y)
}

// ArrowHeadPoints returns three corners of a filled arrowhead with tip at tip.
func ArrowHeadPoints(prev, tip Point, size float64) (a, b, c Point) {
	angle := math.Atan2(tip.Y-prev.Y, tip.X-prev.X)
	half := 0.42
	a = tip
	b = Point{
		X: tip.X - math.Cos(angle-half)*size,
		Y: tip.Y - math.Sin(angle-half)*size,
	}
	c = Point{
		X: tip.X - math.Cos(angle+half)*size,
		Y: tip.Y - math.Sin(angle+half)*size,
	}
	return a, b, c
}
