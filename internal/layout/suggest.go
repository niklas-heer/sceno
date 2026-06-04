package layout

import (
	"fmt"

	"github.com/niklas-heer/sceno/internal/diag"
	"github.com/niklas-heer/sceno/internal/model"
)

// CompactSuggestion analyzes layout density and returns hints.
func CompactSuggestion(d *model.Diagram) []diag.Issue {
	if d == nil {
		return nil
	}
	var hints []diag.Issue
	if len(d.Nodes) == 0 {
		return hints
	}

	minX, minY, maxX, maxY := bounds(d)
	canvasArea := (maxX - minX) * (maxY - minY)
	var nodeArea float64
	for _, n := range d.Nodes {
		if n.Kind != model.ShapeLane {
			nodeArea += n.Rect.W * n.Rect.H
		}
	}
	if canvasArea > 0 {
		ratio := nodeArea / canvasArea
		if ratio < 0.12 {
			hints = append(hints, diag.Issue{
				Code:    diag.CodeSuggestCompact,
				Message: fmt.Sprintf("diagram uses %.0f%% of canvas — layout is sparse", ratio*100),
				Fix:     "Reduce gap (e.g. gap: 20), merge layers, or set row values to pack nodes tighter.",
			})
		}
	}

	// Suggest merging columns with few nodes
	colCount := map[int]int{}
	for _, n := range d.Nodes {
		if n.Column >= 0 {
			colCount[n.Column]++
		}
	}
	for col, cnt := range colCount {
		if cnt == 1 {
			hints = append(hints, diag.Issue{
				Code:    diag.CodeSuggestCompact,
				Message: fmt.Sprintf("column %d has only one node", col),
				Fix:     "Assign the same layer to related nodes or use row to stack within a column.",
			})
		}
	}

	maxLayer := 0
	for _, n := range d.Nodes {
		if n.Layer > maxLayer {
			maxLayer = n.Layer
		}
	}
	if maxLayer > len(d.Nodes)/2 {
		hints = append(hints, diag.Issue{
			Code:    diag.CodeSuggestCompact,
			Message: "many distinct layer values widen the diagram horizontally",
			Fix:     "Reuse layer numbers for parallel branches (same column, different row).",
		})
	}

	return hints
}

func bounds(d *model.Diagram) (minX, minY, maxX, maxY float64) {
	minX, minY = 1e9, 1e9
	maxX, maxY = -1e9, -1e9
	for _, n := range d.Nodes {
		if n.Rect.X < minX {
			minX = n.Rect.X
		}
		if n.Rect.Y < minY {
			minY = n.Rect.Y
		}
		if n.Rect.Right() > maxX {
			maxX = n.Rect.Right()
		}
		if n.Rect.Bottom() > maxY {
			maxY = n.Rect.Bottom()
		}
	}
	return
}
