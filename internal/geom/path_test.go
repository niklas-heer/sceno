package geom

import (
	"math"
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
)

func TestSimplifyPathPreservesDirectionReversal(t *testing.T) {
	pts := []Point{{X: 0, Y: 0}, {X: 20, Y: 0}, {X: 10, Y: 0}}
	got := SimplifyPath(pts)
	if len(got) != 3 {
		t.Fatalf("direction reversal removed: %+v", got)
	}
}

func TestEdgeLabelBoxHorizontal(t *testing.T) {
	pts := []Point{{X: 0, Y: 0}, {X: 100, Y: 0}}
	layout := LayoutEdgeLabel(pts, "write", nil)
	if !layout.Horizontal {
		t.Fatal("expected horizontal")
	}
	if math.Abs(layout.CenterY) > 1 {
		t.Fatalf("horizontal label should sit on connector, ry=%v", layout.CenterY)
	}
	if layout.BoxW <= 0 || layout.BoxH <= 0 {
		t.Fatalf("invalid box size %v x %v", layout.BoxW, layout.BoxH)
	}
	if layout.CenterX != 50 {
		t.Fatalf("expected center x=50, got %v", layout.CenterX)
	}
	if layout.BoxH > 16 {
		t.Fatalf("label box should be compact, got height %.1f", layout.BoxH)
	}
}

func TestEdgeLabelBoxClearsNodes(t *testing.T) {
	pts := []Point{{X: 158, Y: 156}, {X: 230, Y: 156}}
	ctx := &EdgeLabelContext{
		From: model.Rect{X: 40, Y: 116, W: 118, H: 80},
		To:   model.Rect{X: 230, Y: 116, W: 92, H: 80},
	}
	rx, ry, boxW, boxH, horiz := EdgeLabelBox(pts, 6, 4, 14, 12, []string{"write"}, 40, ctx)
	if !horiz {
		t.Fatal("expected horizontal")
	}
	if math.Abs(ry-156) > 2 {
		t.Fatalf("label should sit on connector y, ry=%.1f", ry)
	}
	box := LabelBoxRect(rx, ry, boxW, boxH)
	if box.Y < ctx.From.Y || box.Bottom() > ctx.From.Bottom() {
		t.Fatalf("label should stay within connector band, box=%+v nodes=%+v", box, ctx.From)
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
	if rx != 0 {
		t.Fatalf("vertical label should center on connector axis, rx=%v", rx)
	}
	if boxW <= 0 || boxH <= 0 {
		t.Fatalf("invalid box size")
	}
}

func TestLayoutEdgeLabelSlidesAwayFromChrome(t *testing.T) {
	pts := []Point{{X: 100, Y: 100}, {X: 100, Y: 260}}
	chrome := model.Rect{X: 70, Y: 165, W: 80, H: 16}
	layout := LayoutEdgeLabel(pts, "merge", &EdgeLabelContext{Avoid: []EdgeLabelObstacle{{ID: "lane:title", Kind: "chrome", Bounds: chrome}}})
	x, y, w, h := layout.LabelRect()
	label := model.Rect{X: x, Y: y, W: w, H: h}
	if labelRectTooClose(label, chrome, EdgeLabelObstacleClearance) {
		t.Fatalf("label did not slide away from chrome: label=%+v chrome=%+v", label, chrome)
	}
	if len(layout.BlockedBy) != 0 {
		t.Fatalf("clear placement incorrectly marked blocked: %+v", layout.BlockedBy)
	}
}

func TestLayoutEdgeLabelReportsUnavoidableChrome(t *testing.T) {
	pts := []Point{{X: 100, Y: 100}, {X: 100, Y: 150}}
	chrome := model.Rect{X: 60, Y: 90, W: 80, H: 70}
	layout := LayoutEdgeLabel(pts, "merge", &EdgeLabelContext{Avoid: []EdgeLabelObstacle{{ID: "lane:title", Kind: "chrome", Bounds: chrome}}})
	if len(layout.BlockedBy) != 1 || layout.BlockedBy[0].ID != "lane:title" {
		t.Fatalf("expected unavoidable chrome blocker, got %+v", layout.BlockedBy)
	}
}

func TestLayoutEdgeLabelKeepsSixPixelsFromNodeBorder(t *testing.T) {
	from := model.Rect{X: 0, Y: 80, W: 100, H: 40}
	to := model.Rect{X: 210, Y: 80, W: 100, H: 40}
	pts := []Point{{X: 100, Y: 100}, {X: 210, Y: 100}}
	layout := LayoutEdgeLabel(pts, "REST", &EdgeLabelContext{From: from, To: to})
	x, y, w, h := layout.LabelRect()
	label := model.Rect{X: x, Y: y, W: w, H: h}
	if labelRectTooClose(label, from, EdgeLabelObstacleClearance) || labelRectTooClose(label, to, EdgeLabelObstacleClearance) {
		t.Fatalf("label violates endpoint clearance: %+v", label)
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
