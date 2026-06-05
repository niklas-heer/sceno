package model

// IconPosition controls where an icon sits inside a node (PowerPoint-style).
type IconPosition string

const (
	IconTopLeft     IconPosition = "top-left"
	IconTop         IconPosition = "top"
	IconTopRight    IconPosition = "top-right"
	IconCenter      IconPosition = "center"
	IconBottomLeft  IconPosition = "bottom-left"
	IconBottom      IconPosition = "bottom"
	IconBottomRight IconPosition = "bottom-right"
)

// ParseIconPosition normalizes KDL iconPos values.
func ParseIconPosition(s string) IconPosition {
	switch s {
	case "", "top-left", "topleft", "tl", "left-top":
		return IconTopLeft
	case "top", "t", "north":
		return IconTop
	case "top-right", "topright", "tr", "right-top":
		return IconTopRight
	case "center", "centre", "middle", "c":
		return IconCenter
	case "bottom-left", "bottomleft", "bl", "left-bottom":
		return IconBottomLeft
	case "bottom", "b", "south":
		return IconBottom
	case "bottom-right", "bottomright", "br", "right-bottom":
		return IconBottomRight
	default:
		return IconTopLeft
	}
}
