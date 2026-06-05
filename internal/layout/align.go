package layout

import "github.com/niklas-heer/sceno/internal/model"

// AlignRows applies slide-style smart guides within each container (frame/group
// parent) and at diagram root: siblings sharing a row align to the vertical
// center of that row band; siblings sharing a column align to the horizontal
// center of that column band.
func AlignRows(d *model.Diagram, gap float64) {
	_ = gap
	byParent := map[string][]*model.Node{}
	for i := range d.Nodes {
		n := &d.Nodes[i]
		if !gridPlaced(n) {
			continue
		}
		byParent[n.Parent] = append(byParent[n.Parent], n)
	}
	hasAnnotation := diagramHasAnnotation(d)
	parents := sortedKeys(byParent)
	for _, parent := range parents {
		siblings := byParent[parent]
		alignRowCenters(siblings, parent == "" && hasAnnotation)
		alignColCenters(siblings)
	}
}

func alignRowCenters(siblings []*model.Node, skipRow0 bool) {
	byRow := map[int][]*model.Node{}
	for _, n := range siblings {
		byRow[n.Row] = append(byRow[n.Row], n)
	}
	for _, row := range sortedIntKeys(byRow) {
		ns := byRow[row]
		if skipRow0 && row == 0 {
			continue
		}
		if !axisSnapEligible(ns, func(n *model.Node) int { return n.Column }) {
			continue
		}
		rowTop, rowBottom := ns[0].Rect.Y, ns[0].Rect.Bottom()
		for _, n := range ns[1:] {
			if n.Rect.Y < rowTop {
				rowTop = n.Rect.Y
			}
			if b := n.Rect.Bottom(); b > rowBottom {
				rowBottom = b
			}
		}
		cy := (rowTop + rowBottom) / 2
		for _, n := range ns {
			n.Rect.Y = cy - n.Rect.H/2
		}
	}
}

func alignColCenters(siblings []*model.Node) {
	byCol := map[int][]*model.Node{}
	for _, n := range siblings {
		byCol[n.Column] = append(byCol[n.Column], n)
	}
	for _, col := range sortedIntKeys(byCol) {
		ns := byCol[col]
		if !axisSnapEligible(ns, func(n *model.Node) int { return n.Row }) {
			continue
		}
		colLeft, colRight := ns[0].Rect.X, ns[0].Rect.Right()
		for _, n := range ns[1:] {
			if n.Rect.X < colLeft {
				colLeft = n.Rect.X
			}
			if r := n.Rect.Right(); r > colRight {
				colRight = r
			}
		}
		cx := (colLeft + colRight) / 2
		for _, n := range ns {
			n.Rect.X = cx - n.Rect.W/2
		}
	}
}

// axisSnapEligible is true when 2+ nodes each occupy a unique slot on the cross axis.
func axisSnapEligible(ns []*model.Node, cross func(*model.Node) int) bool {
	if len(ns) < 2 {
		return false
	}
	seen := map[int]int{}
	for _, n := range ns {
		slot := cross(n)
		seen[slot]++
		if seen[slot] > 1 {
			return false
		}
	}
	return true
}

func alignsToGrid(n *model.Node) bool {
	return !n.Fixed && !model.IsContainer(n.Kind) && n.Column >= 0
}

func gridPlaced(n *model.Node) bool {
	if !alignsToGrid(n) {
		return false
	}
	switch model.NormalizeShape(n.Kind) {
	case model.ShapeCallout, model.ShapeNote, model.ShapeTextbox, model.ShapeInfobox:
		// Snap callouts only when explicitly placed on the grid (at= / layer=).
		return n.Layer > 0
	}
	return true
}

func diagramHasAnnotation(d *model.Diagram) bool {
	for _, n := range d.Nodes {
		switch model.NormalizeShape(n.Kind) {
		case model.ShapeCallout, model.ShapeNote, model.ShapeTextbox, model.ShapeInfobox:
			return true
		}
	}
	return false
}
