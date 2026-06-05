package pipeline

import (
	"testing"

	"github.com/niklas-heer/sceno/internal/geom"
	"github.com/niklas-heer/sceno/internal/measure"
	"github.com/niklas-heer/sceno/internal/model"
)

func TestVerticalStackIconDoesNotOverlapLabel(t *testing.T) {
	res, err := BuildAndEvaluateFile("../../examples/fixtures/vertical-stack.kdl", DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range res.Deck.Slides[0].Nodes {
		if n.Icon == "" {
			continue
		}
		ix, iy := measure.IconRect(n, measure.IconSize)
		iconBox := model.Rect{X: ix, Y: iy, W: measure.IconSize, H: measure.IconSize}
		layout := measure.LabelLayoutFor(n)
		labelTop := layout.ContentY + layout.IconOffsetY
		labelBox := model.Rect{
			X: layout.ContentX,
			Y: labelTop,
			W: layout.ContentW,
			H: n.Rect.Bottom() - labelTop,
		}
		if rectsOverlap(iconBox, labelBox, 2) {
			cl := measure.BuildContentLayout(n)
			t.Fatalf("%s icon overlaps label: icon=%+v label=%+v topAlign=%v titleStart=%.1f",
				n.ID, iconBox, labelBox, cl.TopAlign, cl.TitleStartY)
		}
	}
}

func TestVerticalStackUsesTopBottomAnchors(t *testing.T) {
	res, err := BuildAndEvaluateFile("../../examples/fixtures/vertical-stack.kdl", DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	d := res.Deck.Slides[0]
	byID := map[string]model.Node{}
	for _, n := range d.Nodes {
		byID[n.ID] = n
	}
	// ingest->queue should use bottom->top when stacked in one column
	for _, e := range d.Edges {
		if e.From == "ingest" && e.To == "queue" {
			a, b := byID[e.From], byID[e.To]
			if !geom.StackedVertically(a, b) {
				t.Fatalf("ingest and queue should be stacked vertically (cols %d %d)", a.Column, b.Column)
			}
			fs, ts := e.FromSide, e.ToSide
			if fs == "" || fs == model.SideAuto {
				fs, _ = geom.BestSides(a, b)
			}
			if ts == "" || ts == model.SideAuto {
				_, ts = geom.BestSides(a, b)
			}
			if fs != model.SideBottom || ts != model.SideTop {
				t.Fatalf("ingest->queue anchors got %s→%s want bottom→top", fs, ts)
			}
		}
	}
}

func TestPipelineStoresInteriorLayout(t *testing.T) {
	res, err := BuildAndEvaluateFile("../../examples/fixtures/vertical-stack.kdl", DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	n := res.Deck.Slides[0].Nodes[0]
	if !n.Interior.Ready {
		t.Fatal("expected interior layout from pipeline")
	}
	if n.Interior.TitleStartY < 40 {
		t.Fatalf("title start too small: %.1f", n.Interior.TitleStartY)
	}
}

func TestTridentGateBlockedVerticalEdge(t *testing.T) {
	res, err := BuildAndEvaluateFile("../../examples/trident-architecture.kdl", DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	d := res.Deck.Slides[0]
	for _, re := range d.Routed {
		if re.Edge.From == "gate" && re.Edge.To == "blocked" {
			if re.Edge.FromSide != model.SideBottom || re.Edge.ToSide != model.SideTop {
				t.Fatalf("gate->blocked sides %s→%s", re.Edge.FromSide, re.Edge.ToSide)
			}
			return
		}
	}
	t.Fatal("gate->blocked edge not found")
}

func rectsOverlap(a, b model.Rect, gap float64) bool {
	return a.Right()+gap > b.X && b.Right()+gap > a.X &&
		a.Bottom()+gap > b.Y && b.Bottom()+gap > a.Y
}
