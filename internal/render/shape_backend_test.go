package render

import (
	"image/color"
	"testing"

	"github.com/fogleman/gg"
	"github.com/niklas-heer/sceno/internal/model"
)

func TestGGShapeCatalogUsesSemanticSilhouettes(t *testing.T) {
	for _, kind := range []model.ShapeKind{model.ShapeHexagon, model.ShapeTriangle, model.ShapeParallelogram} {
		t.Run(string(kind), func(t *testing.T) {
			n := model.Node{Kind: kind, Fill: "#ff0000", Rect: model.Rect{X: 100, Y: 100, W: 100, H: 80}}
			d := model.Diagram{Padding: 24, Nodes: []model.Node{n}}
			vp := ViewportFrom(d)
			w, h := vp.PixelSize(2)
			dc := gg.NewContext(w, h)
			DrawPolishedGG(dc, d, 0, 0, 2, vp)

			inside := pixelAtWorld(dc, vp, n.Rect.CX(), n.Rect.CY(), 2)
			outside := pixelAtWorld(dc, vp, n.Rect.X+10, n.Rect.Y+10, 2)
			if inside.R < 200 || inside.G > 60 {
				t.Fatalf("center is not shape fill: %+v", inside)
			}
			if outside.R > 240 && outside.G < 80 {
				t.Fatalf("corner is filled like a generic card: %+v", outside)
			}
		})
	}
}

func pixelAtWorld(dc *gg.Context, vp Viewport, x, y, scale float64) color.RGBA {
	px, py := vp.PX(x, y, scale)
	r, g, b, a := dc.Image().At(int(px), int(py)).RGBA()
	return color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
}
