package pipeline

import (
	"path/filepath"
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
)

func TestBuildExampleNoCollisions(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "self-service.kdl")
	d, colls, err := Build(path, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Nodes) < 5 {
		t.Fatalf("expected nodes, got %d", len(d.Nodes))
	}
	if len(colls) > 0 {
		t.Fatalf("collisions remain: %+v", colls)
	}
}

func TestBuildFromSpecFreeRequiresPositions(t *testing.T) {
	_, _, err := BuildFromSpec(model.Spec{
		Layout: model.LayoutFree,
		Nodes:  []model.NodeSpec{{ID: "a", Label: "A"}},
	}, DefaultOptions())
	if err == nil {
		t.Fatal("expected error for missing position")
	}
}
