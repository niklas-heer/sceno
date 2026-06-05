package scene_test

import (
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/pipeline"
	"github.com/niklas-heer/sceno/internal/scene"
	"github.com/niklas-heer/sceno/internal/testutil"
)

func TestPropertyEvaluateDeterministic(t *testing.T) {
	path := "../../examples/fixtures/ledger-database.kdl"
	r1, err := pipeline.BuildAndEvaluateFile(path, pipeline.DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	r2, err := pipeline.BuildAndEvaluateFile(path, pipeline.DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	e1, e2 := r1.Slides[0].Eval, r2.Slides[0].Eval
	if e1.Score != e2.Score {
		t.Fatalf("score drift %d vs %d", e1.Score, e2.Score)
	}
	if len(e1.Findings) != len(e2.Findings) {
		t.Fatalf("findings count drift %d vs %d", len(e1.Findings), len(e2.Findings))
	}
	if len(e1.PaintOrder) != len(e2.PaintOrder) {
		t.Fatal("paint order length drift")
	}
}

func TestPropertyPaintOrderStable(t *testing.T) {
	d := &model.Diagram{
		Gap: 32,
		Nodes: []model.Node{
			{ID: "lane", Kind: model.ShapeLane, Rect: model.Rect{W: 400, H: 200}},
			{ID: "a", Kind: model.ShapeBox, Rect: model.Rect{X: 40, Y: 40, W: 80, H: 40}},
		},
		Edges: []model.Edge{{From: "a", To: "a"}},
		Routed: []model.RoutedEdge{{
			Key: "a-a-0", Edge: model.Edge{From: "a", To: "a"},
			Points: [][]float64{{40, 60}, {120, 60}},
		}},
	}
	p1 := scene.BuildPaintOrder(d)
	p2 := scene.BuildPaintOrder(d)
	if len(p1) != len(p2) {
		t.Fatal("paint order unstable length")
	}
	for i := range p1 {
		if p1[i].Kind != p2[i].Kind || p1[i].Z != p2[i].Z {
			t.Fatalf("paint order drift at %d", i)
		}
	}
}

func TestPropertyInteriorReadyAfterBuild(t *testing.T) {
	path := "../../examples/fixtures/kubernetes-mesh.kdl"
	res, err := pipeline.BuildAndEvaluateFile(path, pipeline.DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range res.Deck.Slides[0].Nodes {
		if model.IsContainer(n.Kind) {
			continue
		}
		if !n.Interior.Ready {
			t.Fatalf("node %q missing interior layout", n.ID)
		}
		if !testutil.InteriorInBounds(n) {
			t.Fatalf("node %q interior out of bounds", n.ID)
		}
	}
}
