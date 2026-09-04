package render

import (
	"math"
	"strings"
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
)

func TestSlideFrameCentersWorldViewport(t *testing.T) {
	d := model.Diagram{Title: "Offset diagram", Padding: 24, Nodes: []model.Node{
		{ID: "a", Kind: model.ShapeBox, Rect: model.Rect{X: -600, Y: 800, W: 240, H: 100}},
	}}
	frame := SlideFrameFrom(d, Aspect16x9)
	vp := frame.Content
	x := frame.OffsetX + (vp.MinX+vp.Width/2)*frame.Scale
	y := frame.OffsetY + (vp.MinY+vp.Height/2)*frame.Scale
	if math.Abs(x-frame.Width/2) > 0.001 || math.Abs(y-frame.Height/2) > 0.001 {
		t.Fatalf("world viewport centered at %.2f,%.2f, want %.2f,%.2f", x, y, frame.Width/2, frame.Height/2)
	}
}

func TestSlidesHTMLPreservesCanonicalCodeScene(t *testing.T) {
	d := model.Diagram{Title: "Code and architecture", Padding: 24, Nodes: []model.Node{
		{ID: "app", Kind: model.ShapeBox, Label: "Application", Rect: model.Rect{X: 40, Y: 100, W: 200, H: 90}},
		{ID: "code", Kind: model.ShapeCode, Code: "return result", CodeLang: "go", Rect: model.Rect{X: 400, Y: 100, W: 200, H: 90}},
	}, Routed: []model.RoutedEdge{{Edge: model.Edge{From: "app", To: "code"}, Points: [][]float64{{240, 145}, {400, 145}}}}}
	html := SlidesHTML(model.Deck{Slides: []model.Diagram{d}})
	if !strings.Contains(html, PolishedSVGSlide(d)) {
		t.Fatal("HTML slides changed the canonical scene, code node, or connector geometry")
	}
	if strings.Contains(html, `class="slide-code"`) || strings.Contains(html, `class="slide-hdr"`) {
		t.Fatal("slide duplicated or moved scene content into separate HTML flow")
	}
}

func TestSlidesHTMLInheritsDeckAspect(t *testing.T) {
	d := model.Diagram{Title: "4:3", Padding: 24, Nodes: []model.Node{{ID: "a", Kind: model.ShapeBox, Rect: model.Rect{X: 40, Y: 100, W: 200, H: 90}}}}
	html := SlidesHTML(model.Deck{SlideAspect: Aspect4x3, Slides: []model.Diagram{d}})
	if !strings.Contains(html, `viewBox="0 0 1600 1200"`) {
		t.Fatal("slide SVG did not inherit the 4:3 deck aspect")
	}
}
