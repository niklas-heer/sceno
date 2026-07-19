package geom

import (
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
)

func TestAnchorRightEdge(t *testing.T) {
	n := model.Node{
		Kind: model.ShapeBox,
		Rect: model.Rect{X: 100, Y: 50, W: 80, H: 40},
	}
	p := Anchor(n, model.SideRight)
	if p.X != 180 || p.Y != 70 {
		t.Fatalf("got %+v want (180,70)", p)
	}
}

func TestCylinderAnchorsTouchSilhouette(t *testing.T) {
	n := model.Node{
		Kind: model.ShapeCylinder,
		Rect: model.Rect{X: 100, Y: 50, W: 200, H: 80},
	}
	top := Anchor(n, model.SideTop)
	if top.X != 200 || top.Y != 50 {
		t.Fatalf("top anchor %+v; want (200,50) on the top rim silhouette", top)
	}
	bottom := Anchor(n, model.SideBottom)
	if bottom.X != 200 || bottom.Y != 130 {
		t.Fatalf("bottom anchor %+v; want (200,130) on the bottom bulge silhouette", bottom)
	}
}

func TestSlidingAnchorClampsToUsableSideSpan(t *testing.T) {
	n := model.Node{Kind: model.ShapeBox, Rect: model.Rect{X: 100, Y: 200, W: 120, H: 60}}
	top, ok := SlidingAnchor(n, model.SideLeft, -100)
	if !ok || top != (Point{X: 100, Y: 212}) {
		t.Fatalf("top sliding port = %+v ok=%v", top, ok)
	}
	bottom, ok := SlidingAnchor(n, model.SideRight, 999)
	if !ok || bottom != (Point{X: 220, Y: 248}) {
		t.Fatalf("bottom sliding port = %+v ok=%v", bottom, ok)
	}
	cloud := model.Node{Kind: model.ShapeCloud, Rect: n.Rect}
	if got, sliding := SlidingAnchor(cloud, model.SideLeft, 212); sliding || got != Anchor(cloud, model.SideLeft) {
		t.Fatalf("cloud must keep silhouette midpoint anchor, got %+v sliding=%v", got, sliding)
	}
}

func TestBestSidesHorizontal(t *testing.T) {
	from := model.Node{Rect: model.Rect{X: 0, Y: 0, W: 50, H: 50}}
	to := model.Node{Rect: model.Rect{X: 200, Y: 0, W: 50, H: 50}}
	f, tside := BestSides(from, to)
	if f != model.SideRight || tside != model.SideLeft {
		t.Fatalf("got %s %s", f, tside)
	}
}

func TestBestSidesVerticalStack(t *testing.T) {
	from := model.Node{Rect: model.Rect{X: 100, Y: 0, W: 80, H: 50}}
	to := model.Node{Rect: model.Rect{X: 105, Y: 120, W: 80, H: 50}}
	if !StackedVertically(from, to) {
		t.Fatal("expected stacked vertically")
	}
	f, tside := BestSides(from, to)
	if f != model.SideBottom || tside != model.SideTop {
		t.Fatalf("got %s %s want bottom top", f, tside)
	}
}

func TestEdgeAnchorsUseSharedBorderRange(t *testing.T) {
	from := model.Node{Row: 0, Rect: model.Rect{X: 0, Y: 20, W: 80, H: 120}}
	to := model.Node{Row: 0, Rect: model.Rect{X: 200, Y: 20, W: 80, H: 40}}
	start, end := EdgeAnchors(from, to, model.SideRight, model.SideLeft)
	if start.Y != end.Y {
		t.Fatalf("pipeline anchors are not aligned: %v %v", start, end)
	}
	if end.Y < to.Rect.Y || end.Y > to.Rect.Bottom() {
		t.Fatalf("anchor %v is outside shorter target border %+v", end, to.Rect)
	}
}

func TestEdgeAnchorsSlideTowardDiagonalPeer(t *testing.T) {
	from := model.Node{Row: 1, Kind: model.ShapeBox, Rect: model.Rect{X: 0, Y: 160, W: 100, H: 60}}
	to := model.Node{Row: 0, Kind: model.ShapeBox, Rect: model.Rect{X: 240, Y: 40, W: 120, H: 80}}
	start, end := EdgeAnchors(from, to, model.SideRight, model.SideLeft)
	if start.Y != from.Rect.Y+12 {
		t.Fatalf("source port did not slide upward: %+v", start)
	}
	if end.Y != to.Rect.Bottom()-12 {
		t.Fatalf("target port did not slide downward: %+v", end)
	}
}
