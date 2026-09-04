package diag

import (
	"strconv"
	"testing"

	"github.com/niklas-heer/sceno/internal/collision"
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

func TestCollisionRepairsPreserveFractionalClearance(t *testing.T) {
	for _, fixed := range []bool{false, true} {
		for _, offset := range []float64{-0.25, 99.75} {
			a := model.Node{ID: "a", Rect: model.Rect{W: 100, H: 100}}
			b := model.Node{ID: "b", Fixed: fixed, DX: 0.5, DY: -0.5, Rect: model.Rect{X: offset, Y: offset, W: 100, H: 100}}
			c := collision.Describe(a, b, 8.25)
			for i, repair := range CollisionRepairs(c, b)[:2] {
				moved := b
				for property, raw := range repair.Properties {
					value, err := strconv.ParseFloat(raw, 64)
					if err != nil {
						t.Fatalf("invalid repair value %q: %v", raw, err)
					}
					switch property {
					case "x":
						moved.Rect.X = value + b.DX
					case "y":
						moved.Rect.Y = value + b.DY
					case "dx":
						moved.Rect.X += value - b.DX
					case "dy":
						moved.Rect.Y += value - b.DY
					}
				}
				if got := collision.Find([]model.Node{a, moved}, 8.25); len(got) != 0 {
					t.Fatalf("fixed=%v offset=%v repair=%d %+v leaves collision: %+v", fixed, offset, i, repair, got)
				}
			}
		}
	}
}
