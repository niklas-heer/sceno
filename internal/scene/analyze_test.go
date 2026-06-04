package scene

import (
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
)

func TestAnalyzePaintOrder(t *testing.T) {
	d := &model.Diagram{
		Style: model.StyleSketch,
		Nodes: []model.Node{
			{ID: "lane", Kind: model.ShapeLane, Rect: model.Rect{X: 0, Y: 0, W: 400, H: 300}},
			{ID: "a", Rect: model.Rect{X: 20, Y: 40, W: 80, H: 50}},
			{ID: "b", Rect: model.Rect{X: 200, Y: 40, W: 80, H: 50}},
		},
		Routed: []model.RoutedEdge{
			{Key: "a-b-0", Edge: model.Edge{From: "a", To: "b"}, Points: [][]float64{{20, 65}, {100, 65}, {200, 65}}},
		},
		Gap: 32,
	}
	r := Analyze(d)
	if len(r.PaintOrder) < 3 {
		t.Fatalf("paint order: %+v", r.PaintOrder)
	}
	if r.PaintOrder[0].Kind != "lane" {
		t.Fatalf("lane should paint first: %+v", r.PaintOrder[0])
	}
	foundEdge, foundNode := false, false
	for _, p := range r.PaintOrder {
		if p.Kind == "edge" {
			foundEdge = true
		}
		if p.Kind == "node" {
			foundNode = true
		}
	}
	if !foundEdge || !foundNode {
		t.Fatalf("expected edge and node layers: %+v", r.PaintOrder)
	}
	if r.Style != "sketch" {
		t.Fatalf("style = %q", r.Style)
	}
}

func TestNarrativeSummary(t *testing.T) {
	r := Report{Style: "polished", Aesthetics: AestheticScore{Overall: 88}}
	s := NarrativeSummary(r)
	if s == "" {
		t.Fatal("empty summary")
	}
}
