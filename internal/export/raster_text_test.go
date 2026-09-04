package export

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/render"
	"github.com/niklas-heer/sceno/internal/theme"
)

func decodeRaster(t *testing.T, data []byte, err error) image.Image {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	return img
}

func TestRasterizeSVGScaledNegativeViewBox(t *testing.T) {
	data, err := RasterizeSVG(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="-100 -50 200 100"><rect x="-90" y="-40" width="20" height="20" fill="red"/></svg>`, 2)
	img := decodeRaster(t, data, err)
	r, g, b, a := img.At(30, 30).RGBA()
	if r < 60000 || g > 1000 || b > 1000 || a < 60000 {
		t.Fatalf("negative viewBox origin was not scaled into viewport: RGBA %d %d %d %d", r, g, b, a)
	}
}

func TestPolishedPNGCurvedSilhouettes(t *testing.T) {
	for _, kind := range []model.ShapeKind{model.ShapeCylinder, model.ShapePill} {
		t.Run(string(kind), func(t *testing.T) {
			n := model.Node{ID: "shape", Kind: kind, Fill: "#ff0000", Stroke: "#000000", Rect: model.Rect{X: 428, Y: 116, W: 116, H: 64}}
			d := model.Diagram{Padding: 24, Theme: model.ThemeConfig{Transparent: true}, Nodes: []model.Node{n}}
			data, err := RenderPNG(d, StylePolished, 2)
			img := decodeRaster(t, data, err)
			vp := render.ViewportFrom(d)
			for y := 0; y < img.Bounds().Dy(); y++ {
				for x := 0; x < img.Bounds().Dx(); x++ {
					wx, wy := vp.MinX+float64(x)/2, vp.MinY+float64(y)/2
					if wx >= n.Rect.X-2 && wx <= n.Rect.Right()+2 && wy >= n.Rect.Y-2 && wy <= n.Rect.Bottom()+2 {
						continue
					}
					_, _, _, a := img.At(x, y).RGBA()
					if a > 0 {
						t.Fatalf("stray paint outside shape bounds at %.1f,%.1f", wx, wy)
					}
				}
			}
			px, py := vp.PX(n.Rect.CX(), n.Rect.Bottom()-3, 2)
			r, g, b, _ := img.At(int(px), int(py)).RGBA()
			if r < 60000 || g > 1000 || b > 1000 {
				t.Fatal("curved bottom is missing its fill")
			}
			px, py = vp.PX(n.Rect.X+1, n.Rect.Y+1, 2)
			_, _, _, a := img.At(int(px), int(py)).RGBA()
			if a > 0 {
				t.Fatal("curved silhouette was squared off at the corner")
			}
		})
	}
}

func TestPNGDiagramAndSlideRetainTextCodeAndIcons(t *testing.T) {
	d := model.Diagram{Title: "Architecture overview", Padding: 24, Nodes: []model.Node{
		{ID: "label", Kind: model.ShapeBox, Label: "Visible label", FontSize: 20, Rect: model.Rect{X: -160, Y: 100, W: 240, H: 90}},
		{ID: "code", Kind: model.ShapeCode, Code: "return hello", CodeLang: "go", Rect: model.Rect{X: -160, Y: 240, W: 240, H: 110}},
		{ID: "icon", Kind: model.ShapeBox, Icon: "user", Rect: model.Rect{X: 140, Y: 100, W: 140, H: 90}},
	}}
	vp := render.ViewportFrom(d)
	regions := map[string]model.Rect{
		"title": {X: vp.MinX + theme.CanvasTextInset, Y: vp.MinY + theme.HeaderTitleBaseline - 30, W: 360, H: 32},
		"label": {X: -140, Y: 110, W: 200, H: 70},
		"code":  {X: -150, Y: 241, W: 180, H: 40},
		"icon":  {X: 150, Y: 110, W: 120, H: 70},
	}
	for _, slide := range []bool{false, true} {
		name := "diagram"
		if slide {
			name = "slide"
		}
		t.Run(name, func(t *testing.T) {
			var data []byte
			var err error
			mapPoint := func(x, y float64) (float64, float64) { return vp.PX(x, y, 2) }
			if slide {
				path := filepath.Join(t.TempDir(), "slide.png")
				err = WriteSlidePNG(d, path, Options{Style: StylePolished, Scale: 0.5})
				if err != nil {
					t.Fatal(err)
				}
				data, err = os.ReadFile(path)
				frame := render.SlideFrameFrom(d, render.Aspect16x9)
				mapPoint = func(x, y float64) (float64, float64) {
					return (frame.OffsetX + x*frame.Scale) * 0.5, (frame.OffsetY + y*frame.Scale) * 0.5
				}
			} else {
				data, err = RenderPNG(d, StylePolished, 2)
			}
			img := decodeRaster(t, data, err)
			for name, region := range regions {
				x1, y1 := mapPoint(region.X, region.Y)
				x2, y2 := mapPoint(region.Right(), region.Bottom())
				ink := 0
				for y := int(y1); y < int(y2); y++ {
					for x := int(x1); x < int(x2); x++ {
						r, g, b, a := img.At(x, y).RGBA()
						if a > 30000 && (r < 40000 || g < 40000 || b < 40000) {
							ink++
						}
					}
				}
				if ink < 10 {
					t.Errorf("%s has only %d ink pixels in its computed region", name, ink)
				}
			}
		})
	}
}

func TestRasterizeSVGPreservesTextAndTransforms(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 300 120"><rect width="300" height="120" fill="white"/><g transform="translate(100 20) scale(2)"><text x="5" y="30" font-size="20" font-weight="600" fill="#111111">Visible</text></g></svg>`
	data, err := RasterizeSVG(svg, 1)
	img := decodeRaster(t, data, err)
	ink := 0
	for y := 35; y < 85; y++ {
		for x := 105; x < 290; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r < 30000 && g < 30000 && b < 30000 {
				ink++
			}
		}
	}
	if ink < 100 {
		t.Fatalf("transformed text has only %d ink pixels; rasterization dropped the label", ink)
	}
}

func TestPolishedPNGDoesNotRepaintOccludedIcons(t *testing.T) {
	r := model.Rect{X: 50, Y: 50, W: 180, H: 100}
	d := model.Diagram{Padding: 24, Nodes: []model.Node{
		{ID: "back", Kind: model.ShapeBox, Icon: "user", Rect: r},
		{ID: "front", Kind: model.ShapeBox, Fill: "#ff0000", Rect: r},
	}}
	data, err := RenderPNG(d, StylePolished, 1)
	img := decodeRaster(t, data, err)
	vp := render.ViewportFrom(d)
	for y := r.Y + 10; y < r.Bottom()-10; y++ {
		for x := r.X + 10; x < r.Right()-10; x++ {
			px, py := vp.PX(x, y, 1)
			r, g, b, _ := img.At(int(px), int(py)).RGBA()
			if r < 60000 || g > 1000 || b > 1000 {
				t.Fatalf("covered icon was repainted over foreground node at %.0f,%.0f", x, y)
			}
		}
	}
}

func TestPNGCodeTabsMatchSpacesWithoutMissingGlyphs(t *testing.T) {
	d := model.Diagram{Padding: 24, Nodes: []model.Node{{ID: "code", Kind: model.ShapeCode, Code: "\treturn \"é\"", CodeLang: "go", Rect: model.Rect{X: 40, Y: 80, W: 240, H: 90}}}}
	tabs, err := RenderPNG(d, StylePolished, 1)
	if err != nil {
		t.Fatal(err)
	}
	d.Nodes[0].Code = "    return \"é\""
	spaces, err := RenderPNG(d, StylePolished, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(tabs, spaces) {
		t.Fatal("tab-indented code differs from equivalent spaces (possible missing tab glyph)")
	}
}
