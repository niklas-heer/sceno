package render

import (
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
)

func TestIconRectPositions(t *testing.T) {
	n := model.Node{
		Icon: "user",
		Rect: model.Rect{X: 0, Y: 0, W: 120, H: 80},
	}
	x, y := IconRect(n, 18)
	if x != 12 || y != 12 {
		t.Fatalf("top-left default: got %v,%v", x, y)
	}
	n.IconPos = model.IconCenter
	x, y = IconRect(n, 18)
	if x != 51 || y != 31 {
		t.Fatalf("center: got %v,%v want 51,31", x, y)
	}
}
