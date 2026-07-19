package render

import (
	"bytes"
	"image/color"
	"strings"
	"testing"

	"github.com/fogleman/gg"
	gofpdf "github.com/go-pdf/fpdf"
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

func TestCloudSVGUsesOneClosedSilhouette(t *testing.T) {
	n := model.Node{Kind: model.ShapeCloud, Fill: "#ff0000", Stroke: "#111111", Rect: model.Rect{X: 10, Y: 20, W: 200, H: 100}}
	got := shapeSVG(n, false)
	if strings.Count(got, "<path") != 1 || strings.Contains(got, "<ellipse") {
		t.Fatalf("cloud must be one path without overlapping ellipses: %s", got)
	}
	for _, midpoint := range []string{"10.0 70.0", "110.0 20.0", "210.0 70.0", "110.0 120.0"} {
		if !strings.Contains(got, midpoint) {
			t.Fatalf("cloud path does not touch bbox midpoint %s: %s", midpoint, got)
		}
	}
}

func TestCloudGGSilhouetteTouchesBBoxMidpoints(t *testing.T) {
	n := model.Node{Kind: model.ShapeCloud, Fill: "#ff0000", Stroke: "#111111", Rect: model.Rect{X: 100, Y: 100, W: 200, H: 100}}
	d := model.Diagram{Padding: 24, Nodes: []model.Node{n}}
	vp := ViewportFrom(d)
	dc := gg.NewContext(500, 300)
	DrawPolishedGG(dc, d, 0, 0, 1, vp)
	for _, p := range [][2]float64{{n.Rect.X + 1, n.Rect.CY()}, {n.Rect.CX(), n.Rect.Y + 1}, {n.Rect.Right() - 1, n.Rect.CY()}, {n.Rect.CX(), n.Rect.Bottom() - 1}} {
		got := pixelAtWorld(dc, vp, p[0], p[1], 1)
		if got.R > 245 && got.G > 245 && got.B > 245 {
			t.Fatalf("bbox midpoint %.0f,%.0f is not painted: %+v", p[0], p[1], got)
		}
	}
	center := pixelAtWorld(dc, vp, n.Rect.CX(), n.Rect.CY(), 1)
	if center.R < 200 || center.G > 80 {
		t.Fatalf("cloud center contains an interior stroke: %+v", center)
	}
}

func TestCloudPDFUsesOneClosedBezierPath(t *testing.T) {
	pdf := gofpdf.NewCustom(&gofpdf.InitType{UnitStr: "pt", Size: gofpdf.SizeType{Wd: 320, Ht: 220}})
	pdf.SetCompression(false)
	pdf.AddPage()
	n := model.Node{Kind: model.ShapeCloud, Fill: "#ff0000", Stroke: "#111111", Rect: model.Rect{X: 50, Y: 50, W: 200, H: 100}}
	drawPolishedNodePDF(pdf, n, 0, 0)
	var out bytes.Buffer
	if err := pdf.Output(&out); err != nil {
		t.Fatal(err)
	}
	stream := out.String()
	if curves := strings.Count(stream, " c\n"); curves != 12 {
		t.Fatalf("cloud PDF path has %d curves, want 12", curves)
	}
	if strings.Count(stream, "h\nB\n") != 1 {
		t.Fatalf("cloud PDF must close and paint exactly one path")
	}
}

func TestActorWithIconUsesCardBackdropInSVG(t *testing.T) {
	n := model.Node{Kind: model.ShapeActor, Icon: "user", Fill: "#ffffff", Stroke: "#111111", Rect: model.Rect{X: 10, Y: 20, W: 120, H: 80}}
	got := shapeSVG(n, false)
	if !strings.Contains(got, "<rect") || strings.Contains(got, "<circle") || strings.Contains(got, "<line") {
		t.Fatalf("actor with icon must use card backdrop, got %s", got)
	}
}

func pixelAtWorld(dc *gg.Context, vp Viewport, x, y, scale float64) color.RGBA {
	px, py := vp.PX(x, y, scale)
	r, g, b, a := dc.Image().At(int(px), int(py)).RGBA()
	return color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
}
