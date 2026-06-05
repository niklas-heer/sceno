package geom

import (
	"testing"
)

func TestEdgeLabelBoxHorizontal(t *testing.T) {
	pts := []Point{{X: 0, Y: 0}, {X: 100, Y: 0}}
	rx, ry, boxW, boxH, horiz := EdgeLabelBox(pts, 6, 4, 14, 12, []string{"write"}, 40)
	if !horiz {
		t.Fatal("expected horizontal")
	}
	if ry >= 0 {
		t.Fatalf("horizontal label should sit above segment, ry=%v", ry)
	}
	if boxW <= 0 || boxH <= 0 {
		t.Fatalf("invalid box size %v x %v", boxW, boxH)
	}
	if rx != 50 {
		t.Fatalf("expected center x=50, got %v", rx)
	}
}

func TestEdgeLabelBoxVertical(t *testing.T) {
	pts := []Point{{X: 0, Y: 0}, {X: 0, Y: 80}}
	rx, _, boxW, boxH, horiz := EdgeLabelBox(pts, 6, 4, 14, 12, []string{"ok?"}, 30)
	if horiz {
		t.Fatal("expected vertical segment")
	}
	if rx <= 0 {
		t.Fatalf("vertical label should sit to the right, rx=%v", rx)
	}
	if boxW <= 0 || boxH <= 0 {
		t.Fatalf("invalid box size")
	}
}
