package render

import (
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/scene"
)

// paintsBeforeEdges delegates to scene engine paint contract.
func paintsBeforeEdges(k model.ShapeKind) bool {
	return scene.PaintsBeforeEdges(k)
}

func nodesBeforeEdges(d *model.Diagram) []model.Node {
	out := scene.NodesOnPlane(d, scene.PlaneLane)
	return append(out, scene.NodesOnPlane(d, scene.PlaneStructure)...)
}

func nodesAfterEdges(d *model.Diagram) []model.Node {
	out := scene.NodesOnPlane(d, scene.PlaneAnnotation)
	return append(out, scene.NodesOnPlane(d, scene.PlaneNode)...)
}
