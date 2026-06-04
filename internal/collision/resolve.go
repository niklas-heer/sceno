package collision

import (
	"math"

	"github.com/niklas-heer/sceno/internal/model"
)

// Find returns all overlapping node pairs (excluding parent/child).
func Find(nodes []model.Node, margin float64) []model.Collision {
	var out []model.Collision
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			a, b := &nodes[i], &nodes[j]
			if related(a, b) {
				continue
			}
			if overlaps(a.Rect, b.Rect, margin) {
				out = append(out, model.Collision{A: a.ID, B: b.ID})
			}
		}
	}
	return out
}

func related(a, b *model.Node) bool {
	if a.Parent == b.ID || b.Parent == a.ID {
		return true
	}
	if model.IsContainer(a.Kind) && b.Parent == a.ID {
		return true
	}
	if model.IsContainer(b.Kind) && a.Parent == b.ID {
		return true
	}
	return false
}

// Resolve pushes overlapping nodes apart; same-column nodes only move vertically.
func Resolve(nodes []model.Node, margin float64, maxIter int) int {
	moves := 0
	for iter := 0; iter < maxIter; iter++ {
		moved := false
		for i := 0; i < len(nodes); i++ {
			for j := i + 1; j < len(nodes); j++ {
				a, b := &nodes[i], &nodes[j]
				if a.Fixed && b.Fixed {
					continue
				}
				if related(a, b) {
					continue
				}
				sx, sy := separation(a, b, margin)
				if sx == 0 && sy == 0 {
					continue
				}
				// Same column: only vertical separation to preserve DAG columns.
				if !a.Fixed && !b.Fixed && a.Column >= 0 && a.Column == b.Column {
					sx = 0
				}
				applyMove(a, b, sx, sy)
				moved = true
				moves++
			}
		}
		if !moved {
			break
		}
	}
	return moves
}

func applyMove(a, b *model.Node, dx, dy float64) {
	if a.Fixed {
		b.Rect.X -= dx
		b.Rect.Y -= dy
	} else if b.Fixed {
		a.Rect.X += dx
		a.Rect.Y += dy
	} else {
		a.Rect.X += dx / 2
		a.Rect.Y += dy / 2
		b.Rect.X -= dx / 2
		b.Rect.Y -= dy / 2
	}
}

func overlaps(a, b model.Rect, gap float64) bool {
	return a.Right()+gap > b.X &&
		b.Right()+gap > a.X &&
		a.Bottom()+gap > b.Y &&
		b.Bottom()+gap > a.Y
}

func separation(a, b *model.Node, gap float64) (dx, dy float64) {
	overlapX := math.Min(a.Rect.Right(), b.Rect.Right()) - math.Max(a.Rect.X, b.Rect.X)
	overlapY := math.Min(a.Rect.Bottom(), b.Rect.Bottom()) - math.Max(a.Rect.Y, b.Rect.Y)
	if overlapX <= 0 && overlapY <= 0 {
		return 0, 0
	}
	var sx, sy float64
	if overlapX > 0 {
		pushX := overlapX + gap
		if a.Rect.CX() < b.Rect.CX() {
			sx = -pushX
		} else {
			sx = pushX
		}
	}
	if overlapY > 0 {
		pushY := overlapY + gap
		if a.Rect.CY() < b.Rect.CY() {
			sy = -pushY
		} else {
			sy = pushY
		}
	}
	if overlapX > 0 && overlapY > 0 {
		return sx / 2, sy / 2
	}
	if overlapX > 0 {
		return sx, 0
	}
	return 0, sy
}
