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
