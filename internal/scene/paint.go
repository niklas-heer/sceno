package scene

import (
	"github.com/niklas-heer/sceno/internal/model"
)

// Z-order constants (back → front). Render, validate, and docs share this contract.
const (
	ZCanvas = iota
	ZLane
	ZStructure
	ZEdge
	ZAnnotation
	ZNode
	ZEdgeLabel
	ZArrow
)

// PaintsBeforeEdges is true for lane/frame/group backgrounds drawn beneath connectors.
func PaintsBeforeEdges(k model.ShapeKind) bool {
	return model.IsContainer(k)
}

// NodesOnPlane returns nodes in deterministic paint order within one semantic plane.
// Source order is the z-order for intentional overlaps on the same plane.
func NodesOnPlane(d *model.Diagram, plane PlaneKind) []model.Node {
	if d == nil {
		return nil
	}
	var out []model.Node
	for _, n := range d.Nodes {
		if PlaneForNode(n) == plane {
			out = append(out, n)
		}
	}
	return out
}

// BuildPaintOrder lists draw order for a laid-out diagram (engine source of truth).
func BuildPaintOrder(d *model.Diagram) []PaintItem {
	if d == nil {
		return nil
	}
	var items []PaintItem
	for _, n := range NodesOnPlane(d, PlaneLane) {
		items = append(items, PaintItem{Z: ZLane, Kind: "lane", ID: n.ID})
	}
	for _, n := range NodesOnPlane(d, PlaneStructure) {
		items = append(items, PaintItem{Z: ZStructure, Kind: containerPaintKind(n.Kind), ID: n.ID})
	}
	for _, re := range d.Routed {
		items = append(items, PaintItem{Z: ZEdge, Kind: "edge", Key: re.Key})
	}
	for _, n := range NodesOnPlane(d, PlaneAnnotation) {
		items = append(items, PaintItem{Z: ZAnnotation, Kind: "annotation", ID: n.ID})
	}
	for _, n := range NodesOnPlane(d, PlaneNode) {
		items = append(items, PaintItem{Z: ZNode, Kind: "node", ID: n.ID})
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
