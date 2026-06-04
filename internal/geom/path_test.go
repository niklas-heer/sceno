package geom

import "testing"

func TestTrimArrowEnd(t *testing.T) {
	pts := []Point{{0, 0}, {100, 0}}
	out := TrimArrowEnd(pts)
	if out[len(out)-1].X >= 100 {
		t.Fatalf("expected shortened end, got %v", out[len(out)-1])
	}
}

func TestSimplifyPath(t *testing.T) {
	pts := []Point{{0, 0}, {50, 0}, {100, 0}}
	out := SimplifyPath(pts)
	if len(out) != 2 {
		t.Fatalf("expected 2 points, got %d", len(out))
	}
}
