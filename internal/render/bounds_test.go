package render

import (
	"math"
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/theme"
)

func TestBoundsReserveHeaderAndBalanceNarrowContent(t *testing.T) {
	d := model.Diagram{
		Title: "Event Pipeline", Subtitle: "Vertical stack with labeled edges", Padding: 28,
		Nodes: []model.Node{{ID: "a", Rect: model.Rect{X: 50, Y: 120, W: 96, H: 48}}},
	}
	minX, minY, maxX, maxY := Bounds(d)
	if gap := d.Nodes[0].Rect.Y - (minY + theme.HeaderSubtitleBaseline); gap < 36 {
		t.Fatalf("subtitle-to-content gap = %.0f, want at least 36", gap)
	}
	left := d.Nodes[0].Rect.CX() - minX
	right := maxX - d.Nodes[0].Rect.CX()
	if math.Abs(left-right) > 1 {
		t.Fatalf("narrow content is not centered: left=%.0f right=%.0f", left, right)
	}
	if bottom := maxY - d.Nodes[0].Rect.Bottom(); bottom > 52 {
		t.Fatalf("bottom whitespace = %.0f, want at most 52", bottom)
	}
}
