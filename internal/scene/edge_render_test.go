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
