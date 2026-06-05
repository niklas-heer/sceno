package collision

import (
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
)

func TestNoSeparationSameRowDifferentColumns(t *testing.T) {
	nodes := []model.Node{
		{ID: "a", Column: 0, Row: 0, Rect: model.Rect{X: 0, Y: 100, W: 100, H: 80}},
		{ID: "b", Column: 1, Row: 0, Rect: model.Rect{X: 200, Y: 100, W: 100, H: 50}},
	}
	beforeA, beforeB := nodes[0].Rect.Y, nodes[1].Rect.Y
	ResolveWithOptions(nodes, 12, 50, ResolveOptions{PreserveSingleRowAlignment: true})
	if nodes[0].Rect.Y != beforeA || nodes[1].Rect.Y != beforeB {
		t.Fatalf("same-row cross-column nodes moved: a=%v b=%v", nodes[0].Rect.Y, nodes[1].Rect.Y)
	}
}

func TestResolveSeparatesOverlap(t *testing.T) {
	nodes := []model.Node{
		{ID: "a", Column: -1, Rect: model.Rect{X: 0, Y: 0, W: 100, H: 50}},
		{ID: "b", Column: -1, Rect: model.Rect{X: 40, Y: 10, W: 100, H: 50}},
	}
	if c := Find(nodes, 8); len(c) == 0 {
		t.Fatal("expected overlap")
	}
	Resolve(nodes, 8, 50)
	if c := Find(nodes, 8); len(c) > 0 {
		t.Fatalf("still overlapping: %+v", c)
	}
}
