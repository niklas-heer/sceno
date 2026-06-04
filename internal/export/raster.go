package export

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"strings"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

// RasterizeSVG renders SVG markup to PNG bytes at the given scale.
func RasterizeSVG(svg string, scale float64) ([]byte, error) {
	if scale <= 0 {
		scale = 2
	}
	icon, err := oksvg.ReadIconStream(strings.NewReader(svg), oksvg.WarnErrorMode)
	if err != nil {
		return nil, fmt.Errorf("parse svg: %w", err)
	}
	w := int(icon.ViewBox.W * scale)
	h := int(icon.ViewBox.H * scale)
	if w < 1 || h < 1 {
		return nil, fmt.Errorf("invalid svg dimensions %dx%d", w, h)
	}
	icon.SetTarget(0, 0, float64(w), float64(h))
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
