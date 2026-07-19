package scene

import (
	"testing"

	"github.com/niklas-heer/sceno/internal/diag"
	"github.com/niklas-heer/sceno/internal/geom"
	"github.com/niklas-heer/sceno/internal/model"
)

func TestCheckEdgeArrowFlagsBentFinalApproach(t *testing.T) {
	d := model.Diagram{Nodes: []model.Node{
		{ID: "source", Rect: model.Rect{X: -80, Y: -40, W: 80, H: 80}},
		{ID: "target", Rect: model.Rect{X: 86, Y: 26, W: 80, H: 80}},
	}}
	points := []geom.Point{{X: 0, Y: 0}, {X: 30, Y: 0}, {X: 50, Y: 10}, {X: 65, Y: 25}, {X: 74, Y: 40}, {X: 79, Y: 51}, {X: 82, Y: 58}, {X: 84, Y: 62}, {X: 85, Y: 64}, {X: 85.5, Y: 65}, {X: 86, Y: 66}}
	re := model.RoutedEdge{
		Edge:   model.Edge{From: "source", To: "target", FromSide: model.SideRight, ToSide: model.SideLeft},
		Points: geom.PathToSlices(points),
	}
	for _, finding := range checkEdgeArrow(&d, re) {
		if finding.Code != string(diag.CodeArrowDetached) || finding.Geometry == nil || len(finding.Repairs) == 0 {
			continue
		}
		if _, ok := finding.Geometry.Bounds["arrow_approach"]; !ok {
			t.Fatal("arrow finding omitted exact approach geometry")
		}
		return
	}
	t.Fatal("bent final approach was silent")
}
