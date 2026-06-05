package geom

import (
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
)

func TestEdgeLabelBoxHorizontal(t *testing.T) {
	pts := []Point{{X: 0, Y: 0}, {X: 100, Y: 0}}
	rx, ry, boxW, boxH, horiz := EdgeLabelBox(pts, 6, 4, 14, 12, []string{"write"}, 40, nil)
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

func TestEdgeLabelBoxClearsNodes(t *testing.T) {
	pts := []Point{{X: 180, Y: 156}, {X: 230, Y: 156}}
	ctx := &EdgeLabelContext{
		From: model.Rect{X: 40, Y: 116, W: 118, H: 80},
		To:   model.Rect{X: 230, Y: 116, W: 92, H: 80},
	}
	rx, ry, boxW, boxH, horiz := EdgeLabelBox(pts, 6, 4, 14, 12, []string{"write"}, 40, ctx)
	if !horiz {
		t.Fatal("expected horizontal")
	}
	box := LabelBoxRect(rx, ry, boxW, boxH)
	if box.Bottom() > ctx.From.Y-4 {
		t.Fatalf("label should sit above nodes, box bottom=%.1f node top=%.1f", box.Bottom(), ctx.From.Y)
	}
	gapCenter := (ctx.From.Right() + 6 + ctx.To.X - 6) / 2
	if rx < gapCenter-2 || rx > gapCenter+2 {
		t.Fatalf("label x should center in node gap, got %.1f want ~%.1f", rx, gapCenter)
	}
}

func TestEdgeLabelBoxVertical(t *testing.T) {
	pts := []Point{{X: 0, Y: 0}, {X: 0, Y: 80}}
	rx, _, boxW, boxH, horiz := EdgeLabelBox(pts, 6, 4, 14, 12, []string{"ok?"}, 30, nil)
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

func TestSplitPathForLabel(t *testing.T) {
	pts := []Point{{X: 0, Y: 50}, {X: 200, Y: 50}}
	box := model.Rect{X: 85, Y: 20, W: 30, H: 16}
	parts := SplitPathForLabel(pts, box)
	if len(parts) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(parts))
	}
	if len(parts[0]) != 2 || parts[0][1].X >= box.X {
		t.Fatalf("left segment should end before label: %+v", parts[0])
	}
	if len(parts[1]) != 2 || parts[1][0].X <= box.Right() {
		t.Fatalf("right segment should start after label: %+v", parts[1])
	}
}
