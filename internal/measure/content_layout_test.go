package measure

import (
	"math"
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
)

func TestBuildContentLayoutSnapsToGrid(t *testing.T) {
	n := model.Node{
		Kind:     model.ShapeBox,
		Label:    "Service",
		Subtitle: "v2",
		Icon:     "api",
		IconPos:  model.IconTop,
		Rect:     model.Rect{W: 120, H: 80},
	}
	cl := BuildContentLayout(n)
	if math.Mod(cl.TitleStartY, SnapUnit) > 0.01 {
		t.Fatalf("title Y not snapped: %.2f", cl.TitleStartY)
	}
	if cl.MinW < 72 || cl.MinH < 40 {
		t.Fatalf("tight bounds too small: %.0f×%.0f", cl.MinW, cl.MinH)
	}
}

func TestLabelLayoutIconOffsetIsRelative(t *testing.T) {
	n := model.Node{
		Kind: model.ShapeBox, Label: "Queue", Icon: "queue", IconPos: model.IconTop,
		Rect: model.Rect{X: 40, Y: 200, W: 100, H: 80},
	}
	layout := LabelLayoutFor(n)
	if layout.IconOffsetY < 40 {
		t.Fatalf("icon offset should be relative to node top, got %.1f", layout.IconOffsetY)
	}
	labelTop := layout.ContentY + layout.IconOffsetY
	_, iy := IconRect(n, IconSize)
	if iy+IconSize > labelTop+2 {
		t.Fatalf("icon bottom %.0f should sit above label top %.0f", iy+IconSize, labelTop)
	}
}

func TestFitSizeUsesSnappedBounds(t *testing.T) {
	w, h := FitSize(model.NodeSpec{Kind: model.ShapeBox, Label: "Hi", Icon: "api"})
	if math.Mod(w, SnapUnit) > 0.01 || math.Mod(h, SnapUnit) > 0.01 {
		t.Fatalf("fit size not snapped: %.2f×%.2f", w, h)
	}
}
