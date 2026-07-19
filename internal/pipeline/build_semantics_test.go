package pipeline

import (
	"testing"

	"github.com/niklas-heer/sceno/internal/measure"
	"github.com/niklas-heer/sceno/internal/model"
)

func TestBuildTreatsExplicitDimensionsAsMinimums(t *testing.T) {
	ns := model.NodeSpec{
		ID: "node", Kind: model.ShapeBox, Label: "A label that needs substantially more width",
		W: 20, H: 180, AtSet: true,
	}
	wantW, _ := measure.FitSize(ns)
	d, _, err := BuildFromSpec(model.Spec{
		Layout: model.LayoutAuto, Gap: 32, Padding: 24, Nodes: []model.NodeSpec{ns},
	}, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	got := d.Nodes[0]
	if got.Rect.W < wantW || got.Rect.H < ns.H {
		t.Fatalf("built size %.0fx%.0f violates minima %.0fx%.0f", got.Rect.W, got.Rect.H, wantW, ns.H)
	}
}

func TestHybridFixedPlacementAndAutoNudge(t *testing.T) {
	x, y := 420.0, 180.0
	s := model.Spec{
		Layout: model.LayoutHybrid, Gap: 32, Padding: 24,
		Nodes: []model.NodeSpec{
			{ID: "flow", Kind: model.ShapeBox, Label: "Flow", AtSet: true},
			{ID: "context", Kind: model.ShapeNote, Label: "Context", X: &x, Y: &y},
			{ID: "nudged", Kind: model.ShapeBox, Label: "Nudged", Layer: 1, AtSet: true, DX: 12, DY: 20},
		},
	}
	d, _, err := BuildFromSpec(s, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]model.Node{}
	for _, n := range d.Nodes {
		byID[n.ID] = n
	}
	base := s
	base.Nodes = append([]model.NodeSpec(nil), s.Nodes...)
	base.Nodes[2].DX, base.Nodes[2].DY = 0, 0
	baseDiagram, _, err := BuildFromSpec(base, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	var baseNudged model.Node
	for _, n := range baseDiagram.Nodes {
		if n.ID == "nudged" {
			baseNudged = n
		}
	}
	if got := byID["context"]; !got.Fixed || got.Rect.X != x || got.Rect.Y != y {
		t.Fatalf("fixed callout = %+v", got)
	}
	if got := byID["nudged"]; got.DX != 12 || got.DY != 20 {
		t.Fatalf("nudge metadata lost: %+v", got)
	} else if got.Rect.X != baseNudged.Rect.X+12 || got.Rect.Y != baseNudged.Rect.Y+20 {
		t.Fatalf("nudge not applied after grid layout: %+v", got.Rect)
	}
}
