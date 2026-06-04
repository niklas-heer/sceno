package collision

import (
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
)

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
