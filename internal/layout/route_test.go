package layout

import (
	"math"
	"testing"

	"github.com/niklas-heer/sceno/internal/geom"
	"github.com/niklas-heer/sceno/internal/model"
)

func TestSnapPathEnds(t *testing.T) {
	start := geom.Point{X: 10, Y: 20}
	end := geom.Point{X: 100, Y: 200}
	pts := []geom.Point{
		{X: 50, Y: 50},
		{X: 80, Y: 180},
		{X: 95, Y: 195},
	}
	out := snapPathEnds(pts, start, end)
	if out[0] != start || out[len(out)-1] != end {
		t.Fatalf("snap failed: start=%v end=%v full=%v", out[0], out[len(out)-1], out)
	}
}

func TestRouteEdgesSnapsToAnchors(t *testing.T) {
	a := model.Node{ID: "a", Rect: model.Rect{X: 0, Y: 0, W: 80, H: 40}}
	b := model.Node{ID: "b", Rect: model.Rect{X: 200, Y: 120, W: 80, H: 40}}
	d := &model.Diagram{
		Gap:   32,
		Style: model.StylePolished,
		Nodes: []model.Node{a, b},
		Edges: []model.Edge{{From: "a", To: "b"}},
	}
	RouteEdges(d)
	if len(d.Routed) != 1 {
		t.Fatal("expected one routed edge")
	}
	re := d.Routed[0]
	gpts := geom.SlicesToPath(re.Points)
	if len(gpts) < 2 {
		t.Fatal("empty path")
	}
	_, dst := geom.EdgeAnchors(a, b, re.Edge.FromSide, re.Edge.ToSide)
	if geom.TipGap(gpts[len(gpts)-1], dst) > 1.5 {
		t.Fatalf("path end %.1f,%.1f not on anchor %.1f,%.1f gap=%.1f",
			gpts[len(gpts)-1].X, gpts[len(gpts)-1].Y, dst.X, dst.Y, geom.TipGap(gpts[len(gpts)-1], dst))
	}
}

func TestRouteEdgesApproachesExplicitTargetSideTangentially(t *testing.T) {
	a := model.Node{ID: "a", Rect: model.Rect{X: 400, Y: 0, W: 100, H: 60}}
	b := model.Node{ID: "b", Rect: model.Rect{X: 0, Y: 240, W: 100, H: 60}}
	d := &model.Diagram{
		Gap: 32, Style: model.StylePolished, Nodes: []model.Node{a, b},
		Edges: []model.Edge{{From: "a", To: "b", FromSide: model.SideRight, ToSide: model.SideLeft}},
	}
	RouteEdges(d)
	pts := geom.SimplifyPath(geom.SlicesToPath(d.Routed[0].Points))
	prev, end := pts[len(pts)-2], pts[len(pts)-1]
	if math.Abs(prev.Y-end.Y) > 0.1 || prev.X >= end.X {
		t.Fatalf("left anchor must be approached horizontally from outside: %v", pts)
	}
}

func TestRouteEdgesChooseClearAutoAnchors(t *testing.T) {
	d := &model.Diagram{Gap: 32, Nodes: []model.Node{
		{ID: "source", Kind: model.ShapeBox, Rect: model.Rect{X: 0, Y: 0, W: 80, H: 60}},
		{ID: "blocker", Kind: model.ShapeBox, Rect: model.Rect{X: 120, Y: 0, W: 100, H: 60}},
		{ID: "target", Kind: model.ShapeBox, Rect: model.Rect{X: 240, Y: 120, W: 80, H: 60}},
	}, Edges: []model.Edge{{From: "source", To: "target"}}}
	RouteEdges(d)
	if len(d.Routed) != 1 {
		t.Fatalf("routed edges = %d", len(d.Routed))
	}
	re := d.Routed[0]
	if re.Edge.FromSide == model.SideRight && re.Edge.ToSide == model.SideLeft {
		t.Fatalf("router kept blocked default anchors: %+v", re.Edge)
	}
	for i := 1; i < len(re.Points); i++ {
		a := geom.Point{X: re.Points[i-1][0], Y: re.Points[i-1][1]}
		b := geom.Point{X: re.Points[i][0], Y: re.Points[i][1]}
		if geom.SegmentHitsRect(a, b, d.Nodes[1].Rect, 8) {
			t.Fatalf("route crosses blocker: %v", re.Points)
		}
	}
}

func TestRouteEdgesAvoidOverlappingOutgoingConnectors(t *testing.T) {
	d := &model.Diagram{Gap: 32, Nodes: []model.Node{
		{ID: "source", Kind: model.ShapeBox, Rect: model.Rect{X: 0, Y: 80, W: 80, H: 60}},
		{ID: "top", Kind: model.ShapeBox, Rect: model.Rect{X: 220, Y: 40, W: 80, H: 60}},
		{ID: "bottom", Kind: model.ShapeBox, Rect: model.Rect{X: 220, Y: 160, W: 80, H: 60}},
	}, Edges: []model.Edge{{From: "source", To: "top"}, {From: "source", To: "bottom"}}}
	RouteEdges(d)
	if collisions := FindEdgeCollisions(d); len(collisions) != 0 {
		t.Fatalf("outgoing connectors collide: %+v routes=%+v", collisions, d.Routed)
	}
}
