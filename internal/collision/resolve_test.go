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

func TestResolveDifferentRowsPreservesColumns(t *testing.T) {
	nodes := []model.Node{
		{ID: "a", Column: 0, Row: 0, Rect: model.Rect{X: 0, Y: 0, W: 100, H: 80}},
		{ID: "b", Column: 1, Row: 1, Rect: model.Rect{X: 60, Y: 40, W: 100, H: 80}},
	}
	wantAX, wantBX := nodes[0].Rect.X, nodes[1].Rect.X
	Resolve(nodes, 12, 50)
	if nodes[0].Rect.X != wantAX || nodes[1].Rect.X != wantBX {
		t.Fatalf("different-row nodes drifted across columns: a=%+v b=%+v", nodes[0].Rect, nodes[1].Rect)
	}
	if c := Find(nodes, 12); len(c) != 0 {
		t.Fatalf("different-row nodes still overlap: %+v", c)
	}
}

func TestContainerIgnoresOutsideNodes(t *testing.T) {
	nodes := []model.Node{
		{ID: "frame", Kind: model.ShapeFrame, Rect: model.Rect{X: 0, Y: 0, W: 300, H: 200}},
		{ID: "outside", Rect: model.Rect{X: 50, Y: 50, W: 80, H: 40}},
	}
	if c := Find(nodes, 8); len(c) != 0 {
		t.Fatalf("container should not collide with outside nodes: %+v", c)
	}
}

func TestFindDescribesCollisionGeometryAndRepairs(t *testing.T) {
	nodes := []model.Node{
		{ID: "a", Rect: model.Rect{X: 0, Y: 0, W: 100, H: 80}},
		{ID: "b", Rect: model.Rect{X: 80, Y: 20, W: 100, H: 80}},
	}
	colls := Find(nodes, 10)
	if len(colls) != 1 {
		t.Fatalf("collisions = %+v", colls)
	}
	c := colls[0]
	if c.Overlap.W != 20 || c.Overlap.H != 60 || c.MoveBX != 30 || c.MoveBY != 70 {
		t.Fatalf("collision details = %+v", c)
	}
}

func TestFindSkipsIntentionalOverlap(t *testing.T) {
	nodes := []model.Node{
		{ID: "a", Rect: model.Rect{W: 100, H: 80}},
		{ID: "b", AllowOverlap: true, Rect: model.Rect{X: 20, Y: 20, W: 100, H: 80}},
	}
	if got := Find(nodes, 0); len(got) != 0 {
		t.Fatalf("intentional overlap reported: %+v", got)
	}
}

func TestResolveSameColumnUsesVerticalSeparation(t *testing.T) {
	nodes := []model.Node{
		{ID: "a", Column: 0, Rect: model.Rect{W: 100, H: 100}},
		{ID: "b", Column: 0, Rect: model.Rect{X: 90, Y: 10, W: 100, H: 100}},
	}
	if moves := Resolve(nodes, 8, 1); moves != 1 {
		t.Fatalf("moves = %d, want one effective move", moves)
	}
	if got := Find(nodes, 8); len(got) != 0 {
		t.Fatalf("same-column collision was not resolved: %+v", got)
	}
	if nodes[0].Rect.X != 0 || nodes[1].Rect.X != 90 {
		t.Fatalf("column X positions changed: %+v", nodes)
	}
}

func TestResolveContainedNodeInOneStep(t *testing.T) {
	for _, fixed := range []int{-1, 0, 1} {
		nodes := []model.Node{
			{ID: "outer", Column: -1, Rect: model.Rect{W: 400, H: 400}},
			{ID: "inner", Column: -1, Rect: model.Rect{X: 170, Y: 160, W: 20, H: 20}},
		}
		var fixedBounds model.Rect
		if fixed >= 0 {
			nodes[fixed].Fixed = true
			fixedBounds = nodes[fixed].Rect
		}
		Resolve(nodes, 8, 1)
		if got := Find(nodes, 8); len(got) != 0 {
			t.Fatalf("fixed=%d: containment not resolved in one step: %+v", fixed, got)
		}
		if fixed >= 0 && nodes[fixed].Rect != fixedBounds {
			t.Fatalf("fixed=%d: fixed node moved", fixed)
		}
	}
}
