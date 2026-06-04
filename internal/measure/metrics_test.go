package measure

import (
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
)

func TestFitSizeGrowsForLabel(t *testing.T) {
	ns := model.NodeSpec{
		ID:       "x",
		Label:    "A very long label that needs space",
		FontSize: 14,
	}
	w, _ := FitSize(ns)
	if w < 150 {
		t.Fatalf("expected wide box, got w=%.0f", w)
	}
}

func TestFitSizeRespectsMinimumW(t *testing.T) {
	ns := model.NodeSpec{
		ID:    "x",
		Label: "Hi",
		W:     200,
	}
	w, _ := FitSize(ns)
	if w < 200 {
		t.Fatalf("expected at least w=200, got %.0f", w)
	}
}

func TestOverflowDetectsSmallBox(t *testing.T) {
	n := model.Node{
		ID:       "x",
		Label:    "Long label text",
		FontSize: 14,
		Rect:     model.Rect{W: 40, H: 40},
	}
	ow, oh := Overflow(n)
	if ow < 1 && oh < 1 {
		t.Fatal("expected overflow")
	}
}

func TestEnsureNodeFits(t *testing.T) {
	n := model.Node{
		ID:       "x",
		Label:    "Long label text",
		FontSize: 14,
		Rect:     model.Rect{W: 40, H: 40},
	}
	EnsureNodeFits(&n)
	ow, oh := Overflow(n)
	if ow > 0 || oh > 0 {
		t.Fatalf("still overflow %.0f %.0f", ow, oh)
	}
}
