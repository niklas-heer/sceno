package measure

import (
	"github.com/niklas-heer/sceno/internal/geom"
	"github.com/niklas-heer/sceno/internal/model"
)

// ShapeContentInsets returns the silhouette-safe content inset for a placed
// shape. Unlike bbox padding, these margins exclude tapered corners, cloud
// lobes, and cylinder rims where text may technically fit the rectangle but
// still paint outside the visible shape.
func ShapeContentInsets(kind model.ShapeKind, w, h float64) (left, top, right, bottom float64) {
	n := model.Node{Kind: kind, Rect: model.Rect{W: w, H: h}}
	writable := geom.WritableRect(n, w/h)
	return writable.X, writable.Y, w - writable.Right(), h - writable.Bottom()
}

// ShapeContentRect is the visible interior available for text and icons.
func ShapeContentRect(n model.Node) model.Rect {
	return ShapeWritableRect(n, n.Rect.W/n.Rect.H)
}

// ShapeWritableRect derives a content rectangle from the shared stroked
// silhouette at the requested content aspect ratio.
func ShapeWritableRect(n model.Node, aspect float64) model.Rect { return geom.WritableRect(n, aspect) }
