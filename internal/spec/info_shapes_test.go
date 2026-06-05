package spec

import (
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
)

func TestInfoShapeDefaultAccent(t *testing.T) {
	s, err := LoadFile("testdata/info-shapes.kdl")
	if err != nil {
		t.Fatal(err)
	}
	accent := map[string]string{}
	for _, n := range s.Nodes {
		accent[n.ID] = n.Accent
	}
	if accent["info"] != "#3b82f6" {
		t.Fatalf("info accent: %q", accent["info"])
	}
	if accent["warn"] != "#f59e0b" {
		t.Fatalf("warn accent: %q", accent["warn"])
	}
	if accent["tip"] != "#10b981" {
		t.Fatalf("tip accent: %q", accent["tip"])
	}
	if s.Nodes[0].Kind != model.ShapeInfobox {
		t.Fatalf("kind %q", s.Nodes[0].Kind)
	}
}
