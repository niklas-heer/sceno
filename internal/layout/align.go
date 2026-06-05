package layout

import (
	"github.com/niklas-heer/sceno/internal/model"
)

// AlignRows vertically centers every node within its row band (same row index shares a center line).
func AlignRows(d *model.Diagram, gap float64) {
	if len(d.Nodes) == 0 {
		return
	}
	rowMaxH := map[int]float64{}
	maxRow := 0
	for i := range d.Nodes {
		n := &d.Nodes[i]
		if n.Fixed {
			continue
		}
		if n.Row > maxRow {
			maxRow = n.Row
		}
		if n.Rect.H > rowMaxH[n.Row] {
			rowMaxH[n.Row] = n.Rect.H
		}
	}
	y := gap + titleOffset(d)
	for r := 0; r <= maxRow; r++ {
		cy := y + rowMaxH[r]/2
		for i := range d.Nodes {
			n := &d.Nodes[i]
			if n.Fixed || n.Row != r {
				continue
			}
			n.Rect.Y = cy - n.Rect.H/2
		}
		y += rowMaxH[r] + gap
	}
}
