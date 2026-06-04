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

// BestSides picks exit/entry sides from relative node positions.
func BestSides(from, to model.Node) (model.Side, model.Side) {
	dx := to.Rect.CX() - from.Rect.CX()
	dy := to.Rect.CY() - from.Rect.CY()
	if math.Abs(dx) >= math.Abs(dy) {
		if dx >= 0 {
			return model.SideRight, model.SideLeft
		}
		return model.SideLeft, model.SideRight
	}
	if dy >= 0 {
		return model.SideBottom, model.SideTop
	}
	return model.SideTop, model.SideBottom
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
