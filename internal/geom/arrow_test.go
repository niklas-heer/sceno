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
