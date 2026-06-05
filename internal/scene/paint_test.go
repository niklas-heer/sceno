package scene

import (
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
)

func TestBuildPaintOrderContainersFirst(t *testing.T) {
	d := &model.Diagram{
		Nodes: []model.Node{
			{ID: "frame", Kind: model.ShapeFrame, Rect: model.Rect{W: 400, H: 200}},
			{ID: "a", Rect: model.Rect{W: 80, H: 40}},
		},
		Routed: []model.RoutedEdge{
			{Key: "a-b", Edge: model.Edge{From: "a", To: "b"}, Points: [][]float64{{0, 0}, {100, 0}}},
		},
	}
	order := BuildPaintOrder(d)
	if len(order) < 3 {
		t.Fatalf("order: %+v", order)
	}
	if order[0].Kind != "frame" || order[0].Z != ZBackground {
		t.Fatalf("frame first: %+v", order[0])
	}
	edgeIdx, nodeIdx := -1, -1
	for i, p := range order {
		if p.Kind == "edge" {
			edgeIdx = i
		}
		if p.Kind == "node" {
			nodeIdx = i
		}
	}
	if edgeIdx < 0 || nodeIdx < 0 || edgeIdx > nodeIdx {
		t.Fatalf("edge before node: %+v", order)
	}
}
