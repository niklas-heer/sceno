package spec

import "github.com/niklas-heer/sceno/internal/model"

func isKnownShape(k model.ShapeKind) bool {
	n := model.NormalizeShape(k)
	for _, s := range model.AllShapes() {
		if model.NormalizeShape(model.ShapeKind(s)) == n {
			return true
		}
	}
	// canonical kinds not in alias list
	switch n {
	case model.ShapeBox, model.ShapeEllipse, model.ShapeCircle, model.ShapeLane,
		model.ShapeTextbox, model.ShapeInfobox, model.ShapeDiamond,
		model.ShapeHexagon, model.ShapeCylinder, model.ShapeCloud, model.ShapeDocument,
		model.ShapeParallelogram, model.ShapeTriangle, model.ShapePill, model.ShapeFrame,
		model.ShapeOctagon, model.ShapeNote:
		return true
	}
	return false
}
