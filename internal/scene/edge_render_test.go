package scene

import (
	"math"
	"testing"

	"github.com/niklas-heer/sceno/internal/diag"
	"github.com/niklas-heer/sceno/internal/geom"
	"github.com/niklas-heer/sceno/internal/model"
)

func TestDetachedArrowDetectedOnShortStroke(t *testing.T) {
	d := &model.Diagram{
		Gap: 32,
		Nodes: []model.Node{
			{ID: "a", Kind: model.ShapeBox, Rect: model.Rect{X: 0, Y: 100, W: 80, H: 50}},
			{ID: "b", Kind: model.ShapeBox, Rect: model.Rect{X: 88, Y: 100, W: 80, H: 50}},
		},
		Routed: []model.RoutedEdge{{
			Key:    "a-b",
			Edge:   model.Edge{From: "a", To: "b"},
			Points: [][]float64{{80, 125}, {88, 125}},
		}},
	}
	findings := edgeRenderFindings(d)
	found := false
	for _, f := range findings {
		if f.Code == string(diag.CodeArrowDetached) {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected detached arrow on 8px connector, got %+v", findings)
	}
}

func TestLayoutEdgeLabelMatchesEngine(t *testing.T) {
	pts := []geom.Point{{X: 180, Y: 156}, {X: 230, Y: 156}}
	ctx := &geom.EdgeLabelContext{
		From: model.Rect{X: 40, Y: 116, W: 118, H: 80},
		To:   model.Rect{X: 230, Y: 116, W: 92, H: 80},
	}
	layout := geom.LayoutEdgeLabel(pts, "write", ctx)
	if math.Abs(layout.CenterY-156) > 2 {
		t.Fatalf("label on connector y, got %.1f", layout.CenterY)
	}
}

func TestArrowheadClusterReportsExactTipsAndRepair(t *testing.T) {
	d := &model.Diagram{Nodes: []model.Node{
		{ID: "a", Rect: model.Rect{X: 0, Y: 0, W: 80, H: 40}},
		{ID: "b", Rect: model.Rect{X: 0, Y: 80, W: 80, H: 40}},
		{ID: "target", Rect: model.Rect{X: 200, Y: 40, W: 100, H: 60}},
	}, Routed: []model.RoutedEdge{
		{Edge: model.Edge{From: "a", To: "target", ToSide: model.SideLeft}, Points: [][]float64{{80, 20}, {200, 65}}},
		{Edge: model.Edge{From: "b", To: "target", ToSide: model.SideLeft}, Points: [][]float64{{80, 100}, {200, 72}}},
	}}
	findings := checkArrowheadClusters(d)
	if len(findings) != 1 || findings[0].Code != string(diag.CodeArrowCluster) {
		t.Fatalf("expected one arrowhead cluster, got %+v", findings)
	}
	if findings[0].Geometry == nil || len(findings[0].Geometry.Bounds) != 2 || len(findings[0].Repairs) == 0 {
		t.Fatalf("cluster lacks exact geometry or repair: %+v", findings[0])
	}
}

func TestUnavoidableContainerChromeLabelFindingHasRepairGeometry(t *testing.T) {
	d := &model.Diagram{Gap: 28, Nodes: []model.Node{
		{ID: "lane", Kind: model.ShapeLane, Label: "Operations", Rect: model.Rect{X: 80, Y: 100, W: 240, H: 100}},
		{ID: "a", Kind: model.ShapeBox, Rect: model.Rect{X: 90, Y: 40, W: 100, H: 40}},
		{ID: "b", Kind: model.ShapeBox, Rect: model.Rect{X: 90, Y: 135, W: 100, H: 40}},
	}, Routed: []model.RoutedEdge{{
		Edge:   model.Edge{From: "a", To: "b", Label: "merge", FromSide: model.SideBottom, ToSide: model.SideTop},
		Points: [][]float64{{140, 80}, {140, 135}},
	}}}
	findings := checkEdgeLabel(d, d.Routed[0])
	for _, finding := range findings {
		if finding.Code == string(diag.CodeEdgeLabelChrome) {
			if finding.Geometry == nil || len(finding.Geometry.Bounds) != 2 || len(finding.Repairs) == 0 {
				t.Fatalf("chrome finding lacks geometry/repair: %+v", finding)
			}
			return
		}
	}
	t.Fatalf("expected unavoidable chrome finding, got %+v", findings)
}
