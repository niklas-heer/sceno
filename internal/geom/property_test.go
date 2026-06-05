package geom

import (
	"math"
	"math/rand"
	"testing"
	"testing/quick"

	"github.com/niklas-heer/sceno/internal/model"
)

func TestPropertyAnchorOnBorder(t *testing.T) {
	sides := []model.Side{model.SideTop, model.SideRight, model.SideBottom, model.SideLeft}
	f := func(w, h uint16, sideIdx uint8) bool {
		wf, hf := float64(w%360+40), float64(h%280+40)
		n := model.Node{Kind: model.ShapeBox, Rect: model.Rect{W: wf, H: hf}}
		side := sides[sideIdx%4]
		p := Anchor(n, side)
		return anchorOnBorder(n.Rect, p, side, 0.01)
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 200}); err != nil {
		t.Fatal(err)
	}
}

func TestPropertyBestSidesVertical(t *testing.T) {
	f := func(dy uint16) bool {
		offset := float64(dy%200 + 60)
		from := model.Node{Rect: model.Rect{X: 100, Y: 0, W: 80, H: 50}}
		to := model.Node{Rect: model.Rect{X: 105, Y: offset, W: 80, H: 50}}
		if !StackedVertically(from, to) {
			return true
		}
		fs, ts := BestSides(from, to)
		return fs == model.SideBottom && ts == model.SideTop
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Fatal(err)
	}
}

func TestPropertyPathSimplifyReducesOrPreserves(t *testing.T) {
	f := func(n uint8) bool {
		rng := rand.New(rand.NewSource(int64(n)))
		pts := make([]Point, 4+int(n%6))
		for i := range pts {
			pts[i] = Point{X: rng.Float64() * 400, Y: rng.Float64() * 300}
		}
		out := SimplifyPath(pts)
		return len(out) >= 2 && len(out) <= len(pts)
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 150}); err != nil {
		t.Fatal(err)
	}
}

func anchorOnBorder(r model.Rect, p Point, s model.Side, eps float64) bool {
	switch s {
	case model.SideTop:
		return math.Abs(p.Y-r.Y) <= eps && p.X >= r.X-eps && p.X <= r.Right()+eps
	case model.SideBottom:
		return math.Abs(p.Y-r.Bottom()) <= eps && p.X >= r.X-eps && p.X <= r.Right()+eps
	case model.SideLeft:
		return math.Abs(p.X-r.X) <= eps && p.Y >= r.Y-eps && p.Y <= r.Bottom()+eps
	case model.SideRight:
		return math.Abs(p.X-r.Right()) <= eps && p.Y >= r.Y-eps && p.Y <= r.Bottom()+eps
	default:
		return false
	}
}

func TestPropertyArrowTipNearPathEnd(t *testing.T) {
	pts := []Point{{0, 0}, {100, 0}, {100, 80}, {200, 80}}
	ag, ok := ArrowGeometryForPath(pts)
	if !ok {
		t.Fatal("expected arrow geometry")
	}
	if TipGap(ag.Tip, pts[len(pts)-1]) > 0.01 {
		t.Fatalf("tip should match path end")
	}
}
