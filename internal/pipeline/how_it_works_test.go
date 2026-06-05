package pipeline

import (
	"math"
	"path/filepath"
	"testing"
)

func TestHowItWorksSingleRow(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "how-it-works.kdl")
	d, colls, err := Build(path, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	if len(colls) > 0 {
		t.Fatalf("unexpected collisions: %+v", colls)
	}
	cys := map[string]float64{}
	for _, n := range d.Nodes {
		cys[n.ID] = n.Rect.CY()
	}
	first := cys["author"]
	for id, cy := range cys {
		if math.Abs(cy-first) > 1 {
			t.Fatalf("node %q not row-aligned: cy=%.1f first=%.1f all=%v", id, cy, first, cys)
		}
	}
}
