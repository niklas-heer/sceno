package layout

import (
	"sort"

	"github.com/niklas-heer/sceno/internal/model"
)

// Grid places nodes in layer columns and row slots (stable, non-overlapping base layout).
func Grid(d *model.Diagram, gap float64) {
	ranks := computeRanks(d)
	groups := map[int][]*model.Node{}

	for i := range d.Nodes {
		n := &d.Nodes[i]
		if n.Fixed {
			n.Column = int(n.Rect.X) // preserve free placement column hint
			continue
		}
		col := ranks[n.ID]
		if n.Layer > 0 {
			col = n.Layer
		}
		n.Column = col
		groups[col] = append(groups[col], n)
	}

	maxCol := 0
	for c := range groups {
		if c > maxCol {
			maxCol = c
		}
	}

	x := gap
	for col := 0; col <= maxCol; col++ {
		ns := groups[col]
		if len(ns) == 0 {
			continue
		}
		sort.Slice(ns, func(i, j int) bool {
			if ns[i].Row != ns[j].Row {
				return ns[i].Row < ns[j].Row
			}
			return ns[i].ID < ns[j].ID
		})
		maxRow := 0
		for _, n := range ns {
			if n.Row > maxRow {
				maxRow = n.Row
			}
		}
		rowH := make([]float64, maxRow+1)
		for _, n := range ns {
			if n.Rect.H > rowH[n.Row] {
				rowH[n.Row] = n.Rect.H
			}
		}
		colW := 0.0
		for _, n := range ns {
			if n.Rect.W > colW {
				colW = n.Rect.W
			}
		}
		for _, n := range ns {
			y := gap + titleOffset(d)
			for r := 0; r < n.Row; r++ {
				y += rowH[r] + gap
			}
			n.Rect.X = x + (colW-n.Rect.W)/2
			n.Rect.Y = y
		}
		x += colW + gap*2
	}
}

func titleOffset(d *model.Diagram) float64 {
	off := 0.0
	if d.Title != "" {
		off += 48
	}
	if d.Subtitle != "" {
		off += 28
	}
	if off > 0 {
		off += 16
	}
	return off
}

func computeRanks(d *model.Diagram) map[string]int {
	byID := index(d.Nodes)
	inDeg := map[string]int{}
	out := map[string][]string{}
	for _, e := range d.Edges {
		if _, ok := byID[e.From]; !ok {
			continue
		}
		if _, ok := byID[e.To]; !ok {
			continue
		}
		out[e.From] = append(out[e.From], e.To)
		inDeg[e.To]++
	}

	rank := map[string]int{}
	queue := []string{}
	for id := range byID {
		if inDeg[id] == 0 {
			queue = append(queue, id)
		}
	}
	visited := 0
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		visited++
		for _, to := range out[id] {
			next := rank[id] + 1
			if next > rank[to] {
				rank[to] = next
			}
			inDeg[to]--
			if inDeg[to] == 0 {
				queue = append(queue, to)
			}
		}
	}
	if visited < len(byID) {
		for id, n := range byID {
			if _, ok := rank[id]; !ok {
				rank[id] = n.Layer
			}
		}
	}
	return rank
}
