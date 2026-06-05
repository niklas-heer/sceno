package geom

import "math"

// ArrowHeadDepth is how far the arrowhead extends from stroke end to the target border.
const ArrowHeadDepth = 9.0

// MaxArrowTipGap is the maximum allowed distance between arrow tip and target anchor.
const MaxArrowTipGap = 2.0

// ArrowGeometry is stroke end + tip for one connector (shared by render and validation).
type ArrowGeometry struct {
	StrokeEnd Point // visible line terminus
	Tip       Point // arrow tip on target border
	Prev      Point // direction reference (last segment start)
}

// ArrowGeometryForPath computes render/validate arrow placement from a routed path.
// The path terminus must be the target anchor on the shape border.
func ArrowGeometryForPath(pts []Point) (ArrowGeometry, bool) {
	gpts := SimplifyPath(pts)
	if len(gpts) < 2 {
		return ArrowGeometry{}, false
	}
	tip := gpts[len(gpts)-1]
	prev := gpts[len(gpts)-2]
	strokeEnd := ShortenEnd(prev, tip, ArrowHeadDepth)
	return ArrowGeometry{StrokeEnd: strokeEnd, Tip: tip, Prev: prev}, true
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
