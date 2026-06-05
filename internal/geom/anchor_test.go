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
