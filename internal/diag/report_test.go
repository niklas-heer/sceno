package diag

import (
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
)

func TestCollisionRepairsUseOriginalFixedCoordinates(t *testing.T) {
	collision := model.Collision{B: "note", MoveBX: 20, MoveBY: 30}
	node := model.Node{
		ID:    "note",
		Fixed: true,
		DX:    10,
		DY:    -5,
		Rect:  model.Rect{X: 110, Y: 95},
	}

	repairs := CollisionRepairs(collision, node)
	if got := repairs[0].Properties["x"]; got != "120" {
		t.Fatalf("x repair = %q, want 120", got)
	}
	if got := repairs[1].Properties["y"]; got != "130" {
		t.Fatalf("y repair = %q, want 130", got)
	}
}
