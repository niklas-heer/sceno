package layout

import (
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
)

func TestGridRespectsAtZeroColumn(t *testing.T) {
	d := model.Diagram{
		Gap: 32,
		Nodes: []model.Node{
			{ID: "a", Rect: model.Rect{W: 80, H: 40}, AtSet: true, Layer: 0, Row: 0},
			{ID: "b", Rect: model.Rect{W: 80, H: 40}, AtSet: true, Layer: 0, Row: 1},
		},
		Edges: []model.Edge{{From: "a", To: "b"}},
	}
	Grid(&d, 32)
	if d.Nodes[0].Column != 0 || d.Nodes[1].Column != 0 {
		t.Fatalf("expected both in column 0, got %d and %d", d.Nodes[0].Column, d.Nodes[1].Column)
	}
	if d.Nodes[1].Rect.Y <= d.Nodes[0].Rect.Bottom() {
		t.Fatalf("row 1 should sit below row 0")
	}
}
