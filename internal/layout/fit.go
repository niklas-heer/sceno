package layout

import "github.com/niklas-heer/sceno/internal/model"

// FitParents expands lane/container nodes around their children.
func FitParents(d *model.Diagram, pad float64) {
	byID := index(d.Nodes)
	children := map[string][]*model.Node{}
	for i := range d.Nodes {
		n := &d.Nodes[i]
		if n.Parent != "" {
			children[n.Parent] = append(children[n.Parent], n)
		}
	}
	for pid, kids := range children {
		p, ok := byID[pid]
		if !ok || !model.IsContainer(p.Kind) {
			continue
		}
		minX, minY := kids[0].Rect.X, kids[0].Rect.Y
		maxR, maxB := kids[0].Rect.Right(), kids[0].Rect.Bottom()
		for _, c := range kids[1:] {
			if c.Rect.X < minX {
				minX = c.Rect.X
			}
			if c.Rect.Y < minY {
				minY = c.Rect.Y
			}
			if c.Rect.Right() > maxR {
				maxR = c.Rect.Right()
			}
			if c.Rect.Bottom() > maxB {
				maxB = c.Rect.Bottom()
			}
		}
		p.Rect.X = minX - pad
		p.Rect.Y = minY - pad - 22
		p.Rect.W = maxR - minX + pad*2
		p.Rect.H = maxB - minY + pad*2 + 22
	}
}
