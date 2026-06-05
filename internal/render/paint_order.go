package render

import (
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/scene"
)

// paintsBeforeEdges delegates to scene engine paint contract.
func paintsBeforeEdges(k model.ShapeKind) bool {
	return scene.PaintsBeforeEdges(k)
}
