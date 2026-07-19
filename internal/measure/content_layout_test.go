package measure

import (
	"math"
	"testing"

	"github.com/niklas-heer/sceno/internal/fonts"
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

func TestApplyInteriorsPreservesExplicitIconPositions(t *testing.T) {
	tests := []struct {
		pos   model.IconPosition
		wantX float64
		wantY float64
	}{
		{model.IconTopLeft, 28, 40},
		{model.IconTop, 52, 12},
		{model.IconTopRight, 92, 12},
		{model.IconCenter, 52, 40},
		{model.IconBottomLeft, 12, 68},
		{model.IconBottom, 52, 68},
		{model.IconBottomRight, 92, 68},
	}
	for _, tc := range tests {
		t.Run(string(tc.pos), func(t *testing.T) {
			nodes := []model.Node{{
				Kind: model.ShapeBox, Label: "Node", Icon: "api", IconPos: tc.pos,
				Rect: model.Rect{W: 124, H: 100},
			}}
			ApplyInteriors(nodes)
			if got := nodes[0].Interior; got.IconX != tc.wantX || got.IconY != tc.wantY {
				t.Fatalf("icon offset = %.0f,%.0f want %.0f,%.0f", got.IconX, got.IconY, tc.wantX, tc.wantY)
			}
		})
	}
}

func TestDefaultIconPositionIsInline(t *testing.T) {
	if got := EffectiveIconPos(model.Node{}); got != model.IconTopLeft {
		t.Fatalf("default icon position = %q want top-left", got)
	}
}

func TestInlineIconAndTextAreCenteredAsAGroup(t *testing.T) {
	n := model.Node{Kind: model.ShapeBox, Label: "Build", Icon: "workflow", Rect: model.Rect{W: 220, H: 64}}
	cl := BuildContentLayout(n)
	textW := TextWidth(n.Label, 14, fonts.WeightMedium)
	groupLeft := cl.IconX
	groupRight := cl.TitleX + textW
	if math.Abs((groupLeft+groupRight)/2-n.Rect.W/2) > SnapUnit {
		t.Fatalf("inline group not centered: left=%.0f right=%.0f node=%.0f", groupLeft, groupRight, n.Rect.W)
	}
	if gap := cl.TitleX - (cl.IconX + cl.IconSize); gap != InlineIconGap {
		t.Fatalf("inline gap = %.0f want %.0f", gap, InlineIconGap)
	}
}
