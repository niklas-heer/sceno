package render

import (
	"github.com/niklas-heer/sceno/internal/composition"
	"github.com/niklas-heer/sceno/internal/model"
)

func Bounds(d model.Diagram) (minX, minY, maxX, maxY float64) {
	return composition.Bounds(d)
}
