package layout

import (
	"sort"

	"github.com/niklas-heer/sceno/internal/model"
)

// PackColumns enforces vertical spacing between nodes sharing a column.
func PackColumns(d *model.Diagram, gap float64) {
	groups := map[int][]*model.Node{}
	for i := range d.Nodes {
		n := &d.Nodes[i]
		if n.Fixed || n.Column < 0 {
			continue
		}
		groups[n.Column] = append(groups[n.Column], n)
	}
	for _, ns := range groups {
		sort.Slice(ns, func(i, j int) bool {
			if ns[i].Rect.Y != ns[j].Rect.Y {
				return ns[i].Rect.Y < ns[j].Rect.Y
			}
			return ns[i].Row < ns[j].Row
		})
		for i := 1; i < len(ns); i++ {
			prev := ns[i-1]
			cur := ns[i]
			minY := prev.Rect.Bottom() + gap
			if cur.Rect.Y < minY {
				cur.Rect.Y = minY
			}
		}
	}
}
