package layout

import (
	"math"
	"reflect"
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

func TestRouteEdgesFansOutSharedTargetSide(t *testing.T) {
	d := &model.Diagram{Gap: 32, Nodes: []model.Node{
		{ID: "top", Kind: model.ShapeBox, Rect: model.Rect{X: 0, Y: 0, W: 100, H: 52}},
		{ID: "bottom", Kind: model.ShapeBox, Rect: model.Rect{X: 0, Y: 140, W: 100, H: 52}},
		{ID: "target", Kind: model.ShapeBox, Rect: model.Rect{X: 260, Y: 70, W: 140, H: 60}},
	}, Edges: []model.Edge{
		{From: "bottom", To: "target", FromSide: model.SideRight, ToSide: model.SideLeft},
		{From: "top", To: "target", FromSide: model.SideRight, ToSide: model.SideLeft},
	}}
	RouteEdges(d)
	if len(d.Routed) != 2 {
		t.Fatalf("routes = %d", len(d.Routed))
	}
	tips := map[string]geom.Point{}
	for _, re := range d.Routed {
		pts := geom.SlicesToPath(re.Points)
		tips[re.Edge.From] = pts[len(pts)-1]
	}
	if tips["top"].Y >= tips["bottom"].Y {
		t.Fatalf("fan-out order is not counterpart order: %+v", tips)
	}
	if tips["top"].Y != d.Nodes[2].Rect.Y+geom.SlidingPortInset || tips["bottom"].Y != d.Nodes[2].Rect.Bottom()-geom.SlidingPortInset {
		t.Fatalf("tips do not use full usable span: %+v", tips)
	}
	if gap := math.Abs(tips["bottom"].Y - tips["top"].Y); gap < 14 {
		t.Fatalf("tips remain clustered: %.1fpx", gap)
	}
}

func TestRouteEdgesFanOutIsDeterministic(t *testing.T) {
	makeDiagram := func() *model.Diagram {
		return &model.Diagram{Gap: 32, Nodes: []model.Node{
			{ID: "a", Rect: model.Rect{X: 0, Y: 0, W: 80, H: 48}},
			{ID: "b", Rect: model.Rect{X: 0, Y: 100, W: 80, H: 48}},
			{ID: "t", Rect: model.Rect{X: 240, Y: 40, W: 100, H: 64}},
		}, Edges: []model.Edge{{From: "a", To: "t"}, {From: "b", To: "t"}}}
	}
	a, b := makeDiagram(), makeDiagram()
	RouteEdges(a)
	RouteEdges(b)
	if !reflect.DeepEqual(a.Routed, b.Routed) {
		t.Fatalf("fan-out routing drifted:\n%+v\n%+v", a.Routed, b.Routed)
	}
}

func TestRouteEdgesUsesAdjacentSideWhenSilhouetteCannotSlide(t *testing.T) {
	d := &model.Diagram{Gap: 32, Nodes: []model.Node{
		{ID: "source", Kind: model.ShapeHexagon, Rect: model.Rect{X: 0, Y: 60, W: 140, H: 60}},
		{ID: "near", Kind: model.ShapeBox, Rect: model.Rect{X: 240, Y: 60, W: 100, H: 60}},
		{ID: "far", Kind: model.ShapeBox, Rect: model.Rect{X: 460, Y: 60, W: 100, H: 60}},
	}, Edges: []model.Edge{{From: "source", To: "near"}, {From: "source", To: "far"}}}
	RouteEdges(d)
	if len(d.Routed) != 2 {
		t.Fatalf("routes = %d", len(d.Routed))
	}
	if d.Routed[0].Edge.FromSide == d.Routed[1].Edge.FromSide {
		t.Fatalf("tapered source stacked both ports on %s", d.Routed[0].Edge.FromSide)
	}
}

func TestScorePathPenalizesShortInteriorSegments(t *testing.T) {
	start, end := (geom.Point{X: 0, Y: 0}), (geom.Point{X: 100, Y: 20})
	short := []geom.Point{start, {X: 40, Y: 0}, {X: 40, Y: 10}, {X: 100, Y: 10}, end}
	clear := []geom.Point{start, {X: 40, Y: 0}, {X: 40, Y: 20}, end}
	if scorePath(short, nil, start, end, 24, model.SideRight, model.SideLeft) <= scorePath(clear, nil, start, end, 24, model.SideRight, model.SideLeft) {
		t.Fatalf("short interior jog was not penalized")
	}
}

func TestScorePathPenalizesSameAxisReversal(t *testing.T) {
	start, end := (geom.Point{X: 0, Y: 0}), (geom.Point{X: 20, Y: 60})
	reversed := []geom.Point{start, {X: 50, Y: 0}, {X: 20, Y: 0}, end}
	clear := []geom.Point{start, {X: 20, Y: 0}, end}
	if scorePath(reversed, nil, start, end, 24, model.SideRight, model.SideBottom) <= scorePath(clear, nil, start, end, 24, model.SideRight, model.SideBottom) {
		t.Fatalf("same-axis fold-back was not penalized")
	}
}

func TestScorePathPenalizesEightPixelObstacleBrush(t *testing.T) {
	obstacle := model.Node{ID: "block", Rect: model.Rect{X: 40, Y: 20, W: 40, H: 40}}
	start, end := (geom.Point{X: 0, Y: 13}), (geom.Point{X: 120, Y: 13})
	near := []geom.Point{start, end}
	far := []geom.Point{{X: 0, Y: 5}, {X: 120, Y: 5}}
	nearScore := scorePath(near, []model.Node{obstacle}, start, end, 0, model.SideRight, model.SideLeft)
	farScore := scorePath(far, []model.Node{obstacle}, far[0], far[1], 0, model.SideRight, model.SideLeft)
	if nearScore <= farScore {
		t.Fatalf("7px obstacle brush score %.0f must exceed clear route %.0f", nearScore, farScore)
	}
}
