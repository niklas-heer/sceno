package geom

import (
	"math"
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
)

func TestWritableRectStaysInsideTaperedSilhouette(t *testing.T) {
	tests := []model.ShapeKind{model.ShapeDiamond, model.ShapeTriangle, model.ShapeHexagon, model.ShapeParallelogram, model.ShapeCloud}
	for _, kind := range tests {
		t.Run(string(kind), func(t *testing.T) {
			n := model.Node{Kind: kind, Rect: model.Rect{X: 40, Y: 20, W: 220, H: 120}}
			writable := WritableRect(n, 2.4)
			if writable.W <= 0 || writable.H <= 0 {
				t.Fatalf("no writable region for %s", kind)
			}
			for _, corner := range []Point{{writable.X, writable.Y}, {writable.Right(), writable.Y}, {writable.Right(), writable.Bottom()}, {writable.X, writable.Bottom()}} {
				if !pointInPolygon(corner, ShapeOutline(n)) {
					t.Fatalf("writable corner %+v outside %s silhouette: %+v", corner, kind, writable)
				}
			}
		})
	}
}

func TestWritableRectAccountsForStrokeAndCylinderRim(t *testing.T) {
	n := model.Node{Kind: model.ShapeCylinder, Rect: model.Rect{X: 10, Y: 20, W: 200, H: 100}}
	got := WritableRect(n, 2)
	ry := CylinderRimRY(n.Rect.W, n.Rect.H)
	if got.Y < n.Rect.Y+2*ry+ShapeBorderClearance-.01 {
		t.Fatalf("writable region crosses top rim: got %+v rim bottom %.1f", got, n.Rect.Y+2*ry)
	}
	if got.X <= n.Rect.X+ShapeStrokeWidth(n.Kind)/2 {
		t.Fatalf("writable region ignores centered border stroke: %+v", got)
	}
}

func TestCloudOutlineUsesRenderedCubicPath(t *testing.T) {
	r := model.Rect{X: 10, Y: 20, W: 200, H: 100}
	start, curves := CloudPath(r)
	outline := ShapeOutline(model.Node{Kind: model.ShapeCloud, Rect: r})
	if outline[0] != start || len(outline) != len(curves)*8+1 {
		t.Fatalf("outline does not sample cloud render path: start=%+v points=%d", outline[0], len(outline))
	}
	last := outline[len(outline)-1]
	if math.Hypot(last.X-start.X, last.Y-start.Y) > .01 {
		t.Fatalf("cloud outline is not closed: start=%+v last=%+v", start, last)
	}
}

func TestActorReservesSeparateWritableLabelBand(t *testing.T) {
	n := model.Node{Kind: model.ShapeActor, Label: "Actor", Rect: model.Rect{X: 20, Y: 30, W: 100, H: 120}}
	figure := ActorFigure(n.Rect, true, 1)
	writable := WritableRect(n, 2)
	if writable.Y < figure.FootY+ShapeBorderClearance {
		t.Fatalf("actor label band crosses figure: foot %.1f writable %+v", figure.FootY, writable)
	}
	if len(ShapeOutline(n)) != 0 || len(ShapeInternalLines(n)) < 20 {
		t.Fatalf("actor must expose disjoint visible strokes instead of a fake closed outline")
	}
}

func pointInPolygon(p Point, polygon []Point) bool {
	inside := false
	for i, a := range polygon {
		b := polygon[(i+1)%len(polygon)]
		if (a.Y > p.Y) != (b.Y > p.Y) && p.X < (b.X-a.X)*(p.Y-a.Y)/(b.Y-a.Y)+a.X {
			inside = !inside
		}
	}
	return inside
}
