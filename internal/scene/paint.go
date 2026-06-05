package scene

import (
	"github.com/niklas-heer/sceno/internal/model"
)

// Z-order constants (back → front). Render, validate, and docs share this contract.
const (
	ZCanvas = iota
	ZBackground
	ZEdge
	ZNode
	ZEdgeLabel
	ZArrow
)

// PaintsBeforeEdges is true for lane/frame/group backgrounds drawn beneath connectors.
func PaintsBeforeEdges(k model.ShapeKind) bool {
	return model.IsContainer(k)
}

// BuildPaintOrder lists draw order for a laid-out diagram (engine source of truth).
func BuildPaintOrder(d *model.Diagram) []PaintItem {
	if d == nil {
		return nil
	}
	var items []PaintItem
	for _, n := range d.Nodes {
		if PaintsBeforeEdges(n.Kind) {
			items = append(items, PaintItem{Z: ZBackground, Kind: containerPaintKind(n.Kind), ID: n.ID})
		}
	}
	for _, re := range d.Routed {
		items = append(items, PaintItem{Z: ZEdge, Kind: "edge", Key: re.Key})
	}
	for _, n := range d.Nodes {
		if !PaintsBeforeEdges(n.Kind) {
			items = append(items, PaintItem{Z: ZNode, Kind: "node", ID: n.ID})
		}
	}
	for _, re := range d.Routed {
		if re.Edge.Label != "" {
			items = append(items, PaintItem{Z: ZEdgeLabel, Kind: "edge_label", Key: re.Key})
		}
	}
	for _, re := range d.Routed {
		items = append(items, PaintItem{Z: ZArrow, Kind: "arrow", Key: re.Key})
	}
	return items
}

func containerPaintKind(k model.ShapeKind) string {
	switch model.NormalizeShape(k) {
	case model.ShapeFrame, model.ShapeGroup:
		return "frame"
	default:
		return "lane"
	}
}
