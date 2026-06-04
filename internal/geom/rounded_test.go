package geom

import (
	"strings"
	"testing"
)

func TestRoundedPathD(t *testing.T) {
	pts := []Point{{0, 0}, {0, 50}, {100, 50}}
	d := RoundedPathD(pts, 8)
	if d == "" {
		t.Fatal("empty path")
	}
	for _, part := range []string{"M", "Q", "L"} {
		if !strings.Contains(d, part) {
			t.Fatalf("expected %q in path, got %q", part, d)
		}
	}
}
