package render

import "github.com/niklas-heer/sceno/internal/model"

// Viewport is the diagram canvas in world coordinates (matches SVG viewBox).
type Viewport struct {
	MinX, MinY, Width, Height float64
}

// ViewportFrom builds the export viewport (same for SVG, PNG, PDF).
func ViewportFrom(d model.Diagram) Viewport {
	minX, minY, maxX, maxY := Bounds(d)
	return Viewport{
		MinX:   minX,
		MinY:   minY,
		Width:  maxX - minX,
		Height: maxY - minY,
	}
}

// PX maps world coordinates to viewport-local pixels at scale.
func (v Viewport) PX(x, y, scale float64) (float64, float64) {
	return (x - v.MinX) * scale, (y - v.MinY) * scale
}

// PixelSize returns raster dimensions for a scale factor.
func (v Viewport) PixelSize(scale float64) (int, int) {
	return int(v.Width * scale), int(v.Height * scale)
}
