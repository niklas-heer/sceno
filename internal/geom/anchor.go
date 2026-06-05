package geom

import (
	"math"

	"github.com/niklas-heer/sceno/internal/model"
)

// Point on the canvas.
type Point struct {
	X, Y float64
}

// Anchor returns the attachment point on a node's border.
func Anchor(n model.Node, side model.Side) Point {
	if side == "" || side == model.SideAuto {
		side = model.SideRight
	}
	r := n.Rect
	switch model.NormalizeShape(n.Kind) {
	case model.ShapeEllipse, model.ShapeCircle, model.ShapeCloud:
		return ellipseAnchor(r, side)
	case model.ShapeActor:
		return rectAnchor(r, side)
	case model.ShapeDiamond, model.ShapeDecision, model.ShapeHexagon, model.ShapeOctagon:
		return diamondAnchor(r, side)
	case model.ShapeTriangle:
		return triangleAnchor(r, side)
	case model.ShapeCylinder, model.ShapeDatabase:
		return cylinderAnchor(r, side)
	default:
		return rectAnchor(r, side)
	}
}

// StackedVertically is true when nodes share a column band and are meaningfully separated on Y.
func StackedVertically(from, to model.Node) bool {
	band := math.Min(from.Rect.W, to.Rect.W)*0.45 + 12
	if math.Abs(from.Rect.CX()-to.Rect.CX()) > band {
		return false
	}
	minH := math.Min(from.Rect.H, to.Rect.H)
	if minH < 1 {
		minH = 1
	}
	return math.Abs(from.Rect.CY()-to.Rect.CY()) > minH*0.25
}

// BestSides picks exit/entry sides from relative node positions.
// Vertically stacked nodes attach top/bottom; horizontal pipelines use left/right.
func BestSides(from, to model.Node) (model.Side, model.Side) {
	if StackedVertically(from, to) {
		dy := to.Rect.CY() - from.Rect.CY()
		if dy >= 0 {
			return model.SideBottom, model.SideTop
		}
		return model.SideTop, model.SideBottom
	}
	dx := to.Rect.CX() - from.Rect.CX()
	dy := to.Rect.CY() - from.Rect.CY()
	if math.Abs(dy) > math.Abs(dx)*0.85 {
		if dy >= 0 {
			return model.SideBottom, model.SideTop
		}
		return model.SideTop, model.SideBottom
	}
	if dx >= 0 {
		return model.SideRight, model.SideLeft
	}
	return model.SideLeft, model.SideRight
}

// IsHorizontalSide reports left/right attachment sides.
func IsHorizontalSide(s model.Side) bool {
	return s == model.SideLeft || s == model.SideRight
}

func rectAnchor(r model.Rect, side model.Side) Point {
	switch side {
	case model.SideTop:
		return Point{r.CX(), r.Y}
	case model.SideBottom:
		return Point{r.CX(), r.Bottom()}
	case model.SideLeft:
		return Point{r.X, r.CY()}
	default:
		return Point{r.Right(), r.CY()}
	}
}

func ellipseAnchor(r model.Rect, side model.Side) Point {
	cx, cy := r.CX(), r.CY()
	switch side {
	case model.SideTop:
		return Point{cx, r.Y}
	case model.SideBottom:
		return Point{cx, r.Bottom()}
	case model.SideLeft:
		return Point{r.X, cy}
	default:
		return Point{r.Right(), cy}
	}
}

func diamondAnchor(r model.Rect, side model.Side) Point {
	cx, cy := r.CX(), r.CY()
	switch side {
	case model.SideTop:
		return Point{cx, r.Y}
	case model.SideBottom:
		return Point{cx, r.Bottom()}
	case model.SideLeft:
		return Point{r.X, cy}
	default:
		return Point{r.Right(), cy}
	}
}

func triangleAnchor(r model.Rect, side model.Side) Point {
	cx := r.CX()
	switch side {
	case model.SideTop:
		return Point{cx, r.Y}
	case model.SideBottom:
		return Point{cx, r.Bottom()}
	case model.SideLeft:
		return Point{r.X, r.CY()}
	default:
		return Point{r.Right(), r.CY()}
	}
}

func cylinderAnchor(r model.Rect, side model.Side) Point {
	ry := math.Min(r.W*0.12, 14.0)
	switch side {
	case model.SideTop:
		return Point{r.CX(), r.Y + ry}
	case model.SideBottom:
		return Point{r.CX(), r.Bottom() - ry}
	default:
		return rectAnchor(r, side)
	}
}

// PadRect expands a rect for obstacle tests.
func PadRect(r model.Rect, pad float64) model.Rect {
	return model.Rect{X: r.X - pad, Y: r.Y - pad, W: r.W + pad*2, H: r.H + pad*2}
}
