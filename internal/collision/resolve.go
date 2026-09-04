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
			if related(a, b) || a.AllowOverlap || b.AllowOverlap {
				continue
			}
			if overlaps(a.Rect, b.Rect, margin) {
				out = append(out, Describe(*a, *b, margin))
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
	// Containers are visual grouping — ignore collisions with nodes outside their subtree.
	if model.IsContainer(a.Kind) && b.Parent != a.ID {
		return true
	}
	if model.IsContainer(b.Kind) && a.Parent != b.ID {
		return true
	}
	return false
}

// Resolve pushes overlapping nodes apart; same-column nodes only move vertically.
func Resolve(nodes []model.Node, margin float64, maxIter int) int {
	return ResolveWithOptions(nodes, margin, maxIter, ResolveOptions{})
}

type ResolveOptions struct {
	// PreserveSingleRowAlignment skips vertical nudging between same-row nodes in different columns.
	PreserveSingleRowAlignment bool
}

// ResolveWithOptions resolves overlaps with optional layout hints.
func ResolveWithOptions(nodes []model.Node, margin float64, maxIter int, opt ResolveOptions) int {
	moves := 0
	for iter := 0; iter < maxIter; iter++ {
		moved := false
		for i := 0; i < len(nodes); i++ {
			for j := i + 1; j < len(nodes); j++ {
				a, b := &nodes[i], &nodes[j]
				if a.Fixed && b.Fixed {
					continue
				}
				if related(a, b) || a.AllowOverlap || b.AllowOverlap {
					continue
				}
				sx, sy := separation(a, b, margin, opt.PreserveSingleRowAlignment)
				if sx == 0 && sy == 0 {
					continue
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

// Describe returns exact overlap geometry and minimum single-axis movements for b.
func Describe(a, b model.Node, margin float64) model.Collision {
	ix1 := math.Max(a.Rect.X, b.Rect.X)
	iy1 := math.Max(a.Rect.Y, b.Rect.Y)
	ix2 := math.Min(a.Rect.Right(), b.Rect.Right())
	iy2 := math.Min(a.Rect.Bottom(), b.Rect.Bottom())
	overlap := model.Rect{X: ix1, Y: iy1, W: math.Max(0, ix2-ix1), H: math.Max(0, iy2-iy1)}

	moveX := a.Rect.Right() + margin - b.Rect.X
	if b.Rect.CX() < a.Rect.CX() {
		moveX = a.Rect.X - margin - b.Rect.Right()
	}
	moveY := a.Rect.Bottom() + margin - b.Rect.Y
	if b.Rect.CY() < a.Rect.CY() {
		moveY = a.Rect.Y - margin - b.Rect.Bottom()
	}
	return model.Collision{
		A: a.ID, B: b.ID, ABounds: a.Rect, BBounds: b.Rect, Overlap: overlap,
		MoveBX: moveX, MoveBY: moveY,
	}
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

func separation(a, b *model.Node, gap float64, preserveSingleRow bool) (dx, dy float64) {
	if !overlaps(a.Rect, b.Rect, gap) {
		return 0, 0
	}
	// Use the same minimum escape distance as machine-readable repairs.
	// Intersection width/height underestimates this distance when one node
	// contains the other, leaving collisions unresolved after a move.
	c := Describe(*a, *b, gap)
	sx, sy := -c.MoveBX, -c.MoveBY
	// Choose the permitted axis before choosing the smaller displacement;
	// zeroing an already-chosen horizontal move can leave a pair unmoved.
	if !a.Fixed && !b.Fixed && a.Column >= 0 && a.Column == b.Column {
		return 0, sy
	}
	if a.Row != b.Row {
		return 0, sy
	}
	if preserveSingleRow && a.Column != b.Column {
		return sx, 0
	}
	if math.Abs(sx) < math.Abs(sy) {
		return sx, 0
	}
	return 0, sy
}
