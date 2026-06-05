package layout

import (
	"math"
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
)

func TestAlignRowsSharedCenter(t *testing.T) {
	d := model.Diagram{
		Gap: 24,
		Nodes: []model.Node{
			{ID: "a", Row: 0, Column: 0, Rect: model.Rect{W: 100, H: 80, X: 10, Y: 5}},
			{ID: "b", Row: 0, Column: 1, Rect: model.Rect{W: 120, H: 50, X: 200, Y: 99}},
		},
	}
	AlignRows(&d, d.Gap)
	cyA := d.Nodes[0].Rect.CY()
	cyB := d.Nodes[1].Rect.CY()
	if math.Abs(cyA-cyB) > 0.01 {
		t.Fatalf("row centers differ: %.2f vs %.2f", cyA, cyB)
	}
}

func TestGridSingleRowAligned(t *testing.T) {
	d := model.Diagram{
		Gap: 24,
		Nodes: []model.Node{
			{ID: "a", Layer: 0, Row: 0, Rect: model.Rect{W: 88, H: 80}},
			{ID: "b", Layer: 1, Row: 0, Rect: model.Rect{W: 100, H: 57}},
			{ID: "c", Layer: 2, Row: 0, Rect: model.Rect{W: 110, H: 57}},
		},
		Edges: []model.Edge{
			{From: "a", To: "b"},
			{From: "b", To: "c"},
		},
	}
	Grid(&d, d.Gap)
	cys := make([]float64, len(d.Nodes))
	for i, n := range d.Nodes {
		cys[i] = n.Rect.CY()
	}
	for i := 1; i < len(cys); i++ {
		if math.Abs(cys[i]-cys[0]) > 0.01 {
			t.Fatalf("grid row not aligned: %v", cys)
		}
	}
}

func TestGridMultiRowTopAligned(t *testing.T) {
	d := model.Diagram{
		Gap: 24,
		Nodes: []model.Node{
			{ID: "a", Layer: 0, Row: 0, Rect: model.Rect{W: 88, H: 80}},
			{ID: "b", Layer: 0, Row: 1, Rect: model.Rect{W: 88, H: 50}},
		},
	}
	Grid(&d, d.Gap)
	if d.Nodes[0].Rect.Y >= d.Nodes[1].Rect.Y {
		t.Fatalf("expected row 1 below row 0")
	}
}
