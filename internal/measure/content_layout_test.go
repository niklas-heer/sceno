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
	positions := []model.IconPosition{
		model.IconTopLeft, model.IconTop, model.IconTopRight, model.IconCenter,
		model.IconBottomLeft, model.IconBottom, model.IconBottomRight,
	}
	for _, pos := range positions {
		t.Run(string(pos), func(t *testing.T) {
			nodes := []model.Node{{
				Kind: model.ShapeBox, Label: "Node", Icon: "api", IconPos: pos,
				Rect: model.Rect{W: 124, H: 100},
			}}
			ApplyInteriors(nodes)
			got := nodes[0].Interior
			if got.IconX < got.WritableX-.5 || got.IconY < got.WritableY-.5 ||
				got.IconX+got.IconSize > got.WritableX+got.WritableW+.5 || got.IconY+got.IconSize > got.WritableY+got.WritableH+.5 {
				t.Fatalf("%s icon %.0f,%.0f lies outside writable region %.0f,%.0f %.0fx%.0f", pos,
					got.IconX, got.IconY, got.WritableX, got.WritableY, got.WritableW, got.WritableH)
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

func TestPlacedLayoutChoosesFontThatFitsWritableRegion(t *testing.T) {
	n := model.Node{
		Kind: model.ShapeDiamond, Label: "Long decision label", FontSize: 18,
		Rect: model.Rect{W: 240, H: 120},
	}
	cl := BuildPlacedContentLayout(n)
	if cl.FontSize >= n.FontSize || cl.FontSize < minAutoFontSize {
		t.Fatalf("effective font %.1f should shrink from %.1f but respect minimum %.1f", cl.FontSize, n.FontSize, minAutoFontSize)
	}
	req := measureContentRequirements(n, cl.FontSize)
	if req.requiredW > cl.WritableW+.5 || req.requiredH > cl.WritableH+.5 {
		t.Fatalf("chosen font does not fit: need %.1fx%.1f writable %.1fx%.1f", req.requiredW, req.requiredH, cl.WritableW, cl.WritableH)
	}
}

func TestWritableRegionClearsVisibleBorder(t *testing.T) {
	n := model.Node{Kind: model.ShapeBox, Label: "API", Rect: model.Rect{W: 120, H: 60}}
	cl := BuildPlacedContentLayout(n)
	if cl.WritableX <= .75 || cl.WritableY <= .75 || cl.WritableX+cl.WritableW >= n.Rect.W-.75 || cl.WritableY+cl.WritableH >= n.Rect.H-.75 {
		t.Fatalf("writable region does not clear centered 1.5px border: %+v", cl)
	}
}

func TestActorLabelSitsBelowFigure(t *testing.T) {
	w, h := FitSize(model.NodeSpec{Kind: model.ShapeActor, Label: "Actor"})
	n := model.Node{Kind: model.ShapeActor, Label: "Actor", Rect: model.Rect{W: w, H: h}}
	cl := BuildPlacedContentLayout(n)
	if cl.WritableY <= n.Rect.H*.58 {
		t.Fatalf("actor writable band %.1f overlaps figure region in %.1fpx shape", cl.WritableY, n.Rect.H)
	}
	if ow, oh := Overflow(n); ow > .5 || oh > .5 {
		t.Fatalf("actor label overflows reserved band by %.1fx%.1f", ow, oh)
	}
}
