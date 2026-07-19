package measure

import (
	"math"

	"github.com/niklas-heer/sceno/internal/model"
)

// ShapeContentInsets returns the silhouette-safe content inset for a placed
// shape. Unlike bbox padding, these margins exclude tapered corners, cloud
// lobes, and cylinder rims where text may technically fit the rectangle but
// still paint outside the visible shape.
func ShapeContentInsets(kind model.ShapeKind, w, h float64) (left, top, right, bottom float64) {
	switch model.NormalizeShape(kind) {
	case model.ShapeCloud:
		return w * .10, h * .18, w * .10, h * .18
	case model.ShapeCylinder, model.ShapeDatabase:
		rim := math.Min(math.Min(w*.10, h*.15), 12)
		return 8, rim + 4, 8, rim + 4
	case model.ShapeHexagon, model.ShapeOctagon:
		return w * .14, h * .08, w * .14, h * .08
	case model.ShapeDiamond, model.ShapeDecision:
		return w * .26, h * .22, w * .26, h * .22
	default:
		return 0, 0, 0, 0
	}
}

// ShapeContentRect is the visible interior available for text and icons.
func ShapeContentRect(n model.Node) model.Rect {
	l, t, r, b := ShapeContentInsets(n.Kind, n.Rect.W, n.Rect.H)
	return model.Rect{X: n.Rect.X + l, Y: n.Rect.Y + t, W: math.Max(0, n.Rect.W-l-r), H: math.Max(0, n.Rect.H-t-b)}
}
