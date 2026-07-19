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

func TestGridReservesBreathingRoomAroundEdgeLabels(t *testing.T) {
	d := model.Diagram{Gap: 24, Nodes: []model.Node{
		{ID: "a", Rect: model.Rect{W: 80, H: 44}, AtSet: true, Layer: 0, Row: 0},
		{ID: "b", Rect: model.Rect{W: 80, H: 44}, AtSet: true, Layer: 1, Row: 0},
		{ID: "c", Rect: model.Rect{W: 80, H: 44}, AtSet: true, Layer: 0, Row: 1},
	}, Edges: []model.Edge{
		{From: "a", To: "b", Label: "long operation"},
		{From: "a", To: "c", Label: "events"},
	}}
	Grid(&d, d.Gap)
	byID := map[string]model.Node{}
	for _, n := range d.Nodes {
		byID[n.ID] = n
	}
	if horizontal := byID["b"].Rect.X - byID["a"].Rect.Right(); horizontal < 105 {
		t.Fatalf("horizontal labeled gap = %.0f, want at least 105", horizontal)
	}
	if vertical := byID["c"].Rect.Y - byID["a"].Rect.Bottom(); vertical < 65 {
		t.Fatalf("vertical labeled gap = %.0f, want at least 65", vertical)
	}
}
