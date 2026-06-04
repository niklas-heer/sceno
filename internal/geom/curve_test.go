package geom

import (
	"math"
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
)

func TestSmoothPathPreservesEndpoints(t *testing.T) {
	pts := []Point{{0, 0}, {50, 0}, {50, 80}, {120, 80}}
	out := SmoothPath(pts, 6)
	if len(out) < 4 {
		t.Fatalf("expected smooth path, got %d points", len(out))
	}
	if out[0] != pts[0] || out[len(out)-1] != pts[len(pts)-1] {
		t.Fatalf("endpoints changed: %v -> %v", pts[0], out[0])
	}
}

func TestPathVisibleFraction(t *testing.T) {
	pts := []Point{{0, 50}, {200, 50}}
	nodes := []model.Node{
		{ID: "a", Rect: model.Rect{X: 0, Y: 0, W: 40, H: 100}},
		{ID: "b", Rect: model.Rect{X: 180, Y: 0, W: 40, H: 100}},
		{ID: "block", Rect: model.Rect{X: 80, Y: 20, W: 40, H: 60}},
	}
	frac := PathVisibleFraction(pts, "a", "b", nodes, 4)
	if frac >= 1 {
		t.Fatalf("expected partial visibility, got %v", frac)
	}
	if frac <= 0 {
		t.Fatalf("expected some visibility, got %v", frac)
	}
}

func TestPathDSmoothNonEmpty(t *testing.T) {
	d := PathDSmooth([]Point{{0, 0}, {40, 10}, {80, 0}, {120, 30}})
	if d == "" || !contains(d, "Q") {
		t.Fatalf("expected quadratic path, got %q", d)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestRectOverlapArea(t *testing.T) {
	a := model.Rect{X: 0, Y: 0, W: 100, H: 100}
	b := model.Rect{X: 50, Y: 50, W: 100, H: 100}
	area := RectOverlapArea(a, b)
	if math.Abs(area-2500) > 1 {
		t.Fatalf("overlap area = %v, want 2500", area)
	}
}
