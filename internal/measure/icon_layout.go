package measure

import "github.com/niklas-heer/sceno/internal/model"

const IconPad = 12.0

// IconRect returns the top-left corner of the icon inside a node.
func IconRect(n model.Node, size float64) (x, y float64) {
	if n.Icon == "" {
		return 0, 0
	}
	if n.Interior.Ready && n.Interior.IconSize > 0 {
		return n.Rect.X + n.Interior.IconX, n.Rect.Y + n.Interior.IconY
	}
	pos := n.IconPos
	if pos == "" {
		pos = model.IconTopLeft
	}
	r := n.Rect
	switch pos {
	case model.IconTop:
		return r.X + (r.W-size)/2, r.Y + IconPad
	case model.IconTopRight:
		return r.Right() - IconPad - size, r.Y + IconPad
	case model.IconCenter:
		return r.X + (r.W-size)/2, r.Y + (r.H-size)/2
	case model.IconBottomLeft:
		return r.X + IconPad, r.Bottom() - IconPad - size
	case model.IconBottom:
		return r.X + (r.W-size)/2, r.Bottom() - IconPad - size
	case model.IconBottomRight:
		return r.Right() - IconPad - size, r.Bottom() - IconPad - size
	default:
		return r.X + IconPad, r.Y + IconPad
	}
}

// LabelLayout describes text placement relative to icon and node bounds.
type LabelLayout struct {
	ContentX, ContentY, ContentW float64
	TopAlign                     bool
	IconOffsetY                  float64
}
