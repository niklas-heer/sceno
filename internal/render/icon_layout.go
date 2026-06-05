package render

import (
	"github.com/niklas-heer/sceno/internal/measure"
	"github.com/niklas-heer/sceno/internal/model"
)

const iconPad = 12.0

// IconRect returns the top-left corner of the icon inside a node.
func IconRect(n model.Node, size float64) (x, y float64) {
	if n.Icon == "" {
		return 0, 0
	}
	pos := n.IconPos
	if pos == "" {
		pos = model.IconTopLeft
	}
	r := n.Rect
	switch pos {
	case model.IconTop:
		return r.X + (r.W-size)/2, r.Y + iconPad
	case model.IconTopRight:
		return r.Right() - iconPad - size, r.Y + iconPad
	case model.IconCenter:
		return r.X + (r.W-size)/2, r.Y + (r.H-size)/2
	case model.IconBottomLeft:
		return r.X + iconPad, r.Bottom() - iconPad - size
	case model.IconBottom:
		return r.X + (r.W-size)/2, r.Bottom() - iconPad - size
	case model.IconBottomRight:
		return r.Right() - iconPad - size, r.Bottom() - iconPad - size
	default: // top-left
		return r.X + iconPad, r.Y + iconPad
	}
}

// LabelLayout describes text placement relative to icon and node bounds.
type LabelLayout struct {
	ContentX, ContentY, ContentW float64
	TopAlign                   bool
	IconOffsetY                float64
}

// LabelLayoutFor returns label region and vertical alignment for a node.
func LabelLayoutFor(n model.Node) LabelLayout {
	pos := n.IconPos
	if pos == "" {
		pos = model.IconTopLeft
	}
	contentW := n.Rect.W - measure.PadX
	iconOff := 0.0
	topAlign := false

	switch pos {
	case model.IconCenter:
		contentW = n.Rect.W - measure.PadX
		topAlign = false
	case model.IconTop, model.IconTopRight:
		contentW = n.Rect.W - measure.PadX
		iconOff = iconPad + iconSize + 4
		topAlign = true
	default:
		if n.Icon != "" {
			contentW -= measure.IconColumn
			iconOff = 14
		}
	}

	k := model.NormalizeShape(n.Kind)
	switch k {
	case model.ShapeInfobox, model.ShapeCallout:
		topAlign = true
	case model.ShapeActor:
		if n.Icon != "" && pos == model.IconTopLeft {
			topAlign = true
			iconOff = iconPad + iconSize + 6
			contentW = n.Rect.W - measure.PadX
		}
	}

	contentX := n.Rect.X + measure.PadX
	if n.Icon != "" && pos == model.IconTopLeft {
		contentX = n.Rect.X + measure.IconColumn
	}

	return LabelLayout{
		ContentX:    contentX,
		ContentY:    n.Rect.Y,
		ContentW:    contentW,
		TopAlign:    topAlign,
		IconOffsetY: iconOff,
	}
}
