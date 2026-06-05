package render

import (
	"github.com/niklas-heer/sceno/internal/measure"
	"github.com/niklas-heer/sceno/internal/model"
)

// IconRect returns the top-left corner of the icon inside a node.
func IconRect(n model.Node, size float64) (x, y float64) {
	return measure.IconRect(n, size)
}

// LabelLayout describes text placement relative to icon and node bounds.
type LabelLayout = measure.LabelLayout

// LabelLayoutFor returns label region and vertical alignment for a node.
func LabelLayoutFor(n model.Node) LabelLayout {
	return measure.LabelLayoutFor(n)
}
