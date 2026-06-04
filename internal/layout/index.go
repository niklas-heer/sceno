package layout

import "github.com/niklas-heer/sceno/internal/model"

func index(nodes []model.Node) map[string]*model.Node {
	m := make(map[string]*model.Node, len(nodes))
	for i := range nodes {
		m[nodes[i].ID] = &nodes[i]
	}
	return m
}
