package export

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"math"
	"strings"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

// RasterizeSVG renders SVG markup to PNG bytes at the given scale.
func RasterizeSVG(svg string, scale float64) ([]byte, error) {
	if math.IsNaN(scale) || math.IsInf(scale, 0) {
		return nil, fmt.Errorf("PNG scale must be finite")
	}
	if scale <= 0 {
		scale = 2
	}
	svg, err := prepareSVGForRaster(svg)
	if err != nil {
		return nil, fmt.Errorf("prepare SVG: %w", err)
	}
	icon, err := oksvg.ReadIconStream(strings.NewReader(svg), oksvg.WarnErrorMode)
	if err != nil {
		return nil, fmt.Errorf("parse svg: %w", err)
	}
	width, height := icon.ViewBox.W*scale, icon.ViewBox.H*scale
	// Bound memory before converting floats or allocating image/scanner buffers.
	// Vector exports remain available for canvases beyond this raster budget.
	const maxDimension, maxPixels = 32768, 32_000_000
	if math.IsNaN(width) || math.IsNaN(height) || math.IsInf(width, 0) || math.IsInf(height, 0) || width > maxDimension || height > maxDimension || width*height > maxPixels {
		return nil, fmt.Errorf("PNG dimensions exceed the limit of %d px per side or %d pixels; reduce the export scale or use SVG/PDF", maxDimension, maxPixels)
	}
	w := int(width)
	h := int(height)
	if w < 1 || h < 1 {
		return nil, fmt.Errorf("invalid svg dimensions %dx%d", w, h)
	}
	// SetTarget subtracts viewBox origins before scaling, leaving nonzero
	// origins shifted at export scales other than 1. Map world coordinates
	// to pixels explicitly, using the actual rounded raster dimensions.
	sx, sy := float64(w)/icon.ViewBox.W, float64(h)/icon.ViewBox.H
	icon.Transform = rasterx.Identity.Scale(sx, sy).Translate(-icon.ViewBox.X, -icon.ViewBox.Y)
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	scanner := rasterx.NewScannerGV(w, h, rgba, rgba.Bounds())
	dasher := rasterx.NewDasher(w, h, scanner)
	icon.Draw(dasher, 1.0)
	var buf bytes.Buffer
	if err := png.Encode(&buf, rgba); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
