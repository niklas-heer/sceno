package layout

import (
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
	dst := geom.Anchor(b, model.SideLeft)
	if geom.TipGap(gpts[len(gpts)-1], dst) > 1.5 {
		t.Fatalf("path end %.1f,%.1f not on anchor %.1f,%.1f gap=%.1f",
			gpts[len(gpts)-1].X, gpts[len(gpts)-1].Y, dst.X, dst.Y, geom.TipGap(gpts[len(gpts)-1], dst))
	}
}
