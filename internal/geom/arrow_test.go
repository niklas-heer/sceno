package geom

import "testing"

func TestArrowGeometryTipOnBorder(t *testing.T) {
	pts := []Point{{X: 100, Y: 50}, {X: 200, Y: 50}}
	ag, ok := ArrowGeometryForPath(pts)
	if !ok {
		t.Fatal("expected geometry")
	}
	if TipGap(ag.Tip, pts[1]) > MaxArrowTipGap {
		t.Fatalf("tip should be at border, gap=%v", TipGap(ag.Tip, pts[1]))
	}
	if ag.StrokeEnd.X >= ag.Tip.X {
		t.Fatalf("stroke should end before tip: stroke=%v tip=%v", ag.StrokeEnd, ag.Tip)
	}
}

func TestArrowGeometryRejectsBentSmoothedTail(t *testing.T) {
	pts := []Point{{0, 0}, {30, 0}, {50, 10}, {65, 25}, {74, 40}, {79, 51}, {82, 58}, {84, 62}, {85, 64}, {85.5, 65}, {86, 66}}
	ag, ok := ArrowGeometryForPath(pts)
	if !ok {
		t.Fatal("expected arrow geometry")
	}
	if ag.VisibleApproach >= EdgeLabelClearRun {
		t.Fatalf("bent tail falsely reported %.2fpx of straight approach", ag.VisibleApproach)
	}
	if ag.StartApproach.X <= pts[0].X {
		t.Fatalf("start approach should follow the route: %+v", ag.StartApproach)
	}
	trimmed := TrimArrowEnd(pts)
	if len(trimmed) >= len(pts) {
		t.Fatalf("trim should remove dense tail: before=%d after=%d", len(pts), len(trimmed))
	}
	if gap := TipGap(trimmed[len(trimmed)-1], ag.StrokeEnd); gap > 0.01 {
		t.Fatalf("trim end does not match arrow geometry: %.2f", gap)
	}
}

func TestArrowGeometryCountsCollinearSamplesAsStraightApproach(t *testing.T) {
	pts := []Point{{0, 0}, {40, 0}, {60, 20}, {70, 20}, {80, 20}, {90, 20}, {100, 20}}
	ag, ok := ArrowGeometryForPath(pts)
	if !ok {
		t.Fatal("expected arrow geometry")
	}
	if ag.VisibleApproach < EdgeLabelClearRun {
		t.Fatalf("straight sampled approach = %.2f, want at least %.2f", ag.VisibleApproach, EdgeLabelClearRun)
	}
}
