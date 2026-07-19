package render

import (
	"strings"
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/pipeline"
)

func TestPaintOrderContainersBeforeEdges(t *testing.T) {
	s := model.Spec{
		Layout: model.LayoutAuto,
		Gap:    32,
		Nodes: []model.NodeSpec{
			{ID: "frame", Kind: model.ShapeFrame, Label: "Group"},
			{ID: "a", Kind: model.ShapeBox, Label: "A", Parent: "frame", Layer: 1, Row: 0},
			{ID: "b", Kind: model.ShapeBox, Label: "B", Parent: "frame", Layer: 2, Row: 0},
		},
		Edges: []model.EdgeSpec{{From: "a", To: "b"}},
	}
	d, _, err := pipeline.BuildFromSpec(s, pipeline.DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	svg := PolishedSVG(d)
	frameFill := strings.Index(svg, paint.BgLane)
	pathStroke := strings.Index(svg, `stroke="`+paint.EdgeDefault+`"`)
	if frameFill < 0 || pathStroke < 0 {
		t.Fatalf("missing frame or edge in SVG:\n%s", svg)
	}
	shadow := strings.Index(svg, `filter="url(#shadow)"`)
	if frameFill > pathStroke {
		t.Fatal("frame background must appear before edge stroke in SVG paint order")
	}
	if shadow >= 0 && pathStroke > shadow {
		t.Fatal("edge stroke must appear before foreground node shadow in SVG paint order")
	}
}

func TestPaintOrderContainerInternalEdges(t *testing.T) {
	s := model.Spec{
		Layout: model.LayoutAuto, Gap: 32,
		Nodes: []model.NodeSpec{
			{ID: "frame", Kind: model.ShapeFrame, Label: "Group"},
			{ID: "producer", Kind: model.ShapeBox, Label: "Producer", Parent: "frame", Layer: 1, Fill: "#dbeafe"},
			{ID: "consumer", Kind: model.ShapeBox, Label: "Consumer", Parent: "frame", Layer: 2, Fill: "#dcfce7"},
		},
		Edges: []model.EdgeSpec{{From: "producer", To: "consumer"}},
	}
	d, _, err := pipeline.BuildFromSpec(s, pipeline.DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	svg := PolishedSVG(d)
	// First pipeline connector should be visible in SVG (stroke path before consumer child fills).
	firstEdge := strings.Index(svg, `stroke="`+paint.EdgeDefault+`"`)
	consumerChild := strings.Index(svg, `#dbeafe`)
	if firstEdge < 0 || consumerChild < 0 {
		t.Fatal("missing edge or consumer node fill")
	}
	if firstEdge > consumerChild {
		t.Fatal("internal frame edges must render before child node fills")
	}
}
