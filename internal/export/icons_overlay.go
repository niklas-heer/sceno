package export

import (
	"bytes"
	"image/png"

	"github.com/niklas-heer/sceno/internal/icons"
	"github.com/niklas-heer/sceno/internal/measure"
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/render"
	"github.com/niklas-heer/sceno/internal/theme"

	"github.com/fogleman/gg"
)

// overlayIcons draws icons on a rasterized PNG (oksvg often skips nested symbols).
func overlayIcons(pngData []byte, d model.Diagram, scale float64) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		return pngData, err
	}
	vp := render.ViewportFrom(d)
	dc := gg.NewContextForImage(img)
	sz := measure.IconSize * scale
	for _, n := range d.Nodes {
		if n.Icon == "" || model.IsContainer(n.Kind) {
			continue
		}
		ix, iy := render.IconRect(n, measure.IconSize)
		x, y := vp.PX(ix, iy, scale)
		icons.Draw(dc, n.Icon, x, y, sz, theme.FgMuted)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return pngData, err
	}
	return buf.Bytes(), nil
}
