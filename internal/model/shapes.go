package model

import "strings"

// Shape kinds — use in KDL as: shape box id "Label" ...
const (
	ShapeBox            ShapeKind = "box"
	ShapeCard           ShapeKind = "card"           // alias → box
	ShapeEllipse        ShapeKind = "ellipse"
	ShapeCircle         ShapeKind = "circle"
	ShapeActor          ShapeKind = "actor"          // alias → ellipse
	ShapeLane           ShapeKind = "lane"
	ShapeContainer      ShapeKind = "container"      // alias → lane
	ShapeTextbox        ShapeKind = "textbox"
	ShapeNote           ShapeKind = "note"           // sticky note
	ShapeInfobox        ShapeKind = "infobox"
	ShapeCallout        ShapeKind = "callout"        // alias → infobox
	ShapeDiamond        ShapeKind = "diamond"
	ShapeDecision       ShapeKind = "decision"       // alias → diamond
	ShapeHexagon        ShapeKind = "hexagon"
	ShapeCylinder       ShapeKind = "cylinder"
	ShapeDatabase       ShapeKind = "database"       // alias → cylinder
	ShapeCloud          ShapeKind = "cloud"
	ShapeDocument       ShapeKind = "document"
	ShapeParallelogram  ShapeKind = "parallelogram"
	ShapeTriangle       ShapeKind = "triangle"
	ShapePill           ShapeKind = "pill"
	ShapeTerminal       ShapeKind = "terminal"       // alias → pill
	ShapeStart          ShapeKind = "start"          // alias → pill
	ShapeEnd            ShapeKind = "end"            // alias → pill
	ShapeFrame          ShapeKind = "frame"
	ShapeGroup          ShapeKind = "group"          // alias → frame
	ShapeOctagon        ShapeKind = "octagon"
	ShapeCode           ShapeKind = "code" // syntax-highlighted code block (slides & diagrams)
)

// NormalizeShape maps KDL-friendly names to canonical kinds.
func NormalizeShape(k ShapeKind) ShapeKind {
	switch strings.ToLower(string(k)) {
	case "card", "rect", "rectangle":
		return ShapeBox
	case "actor", "person":
		return ShapeActor
	case "users":
		return ShapeEllipse
	case "container":
		return ShapeLane
	case "callout":
		return ShapeInfobox
	case "info":
		return ShapeInfobox
	case "warning", "warn":
		return ShapeInfobox
	case "tip", "hint":
		return ShapeInfobox
	case "decision":
		return ShapeDiamond
	case "database", "db", "storage-cylinder":
		return ShapeCylinder
	case "terminal", "start", "end", "capsule", "terminator":
		return ShapePill
	case "group":
		return ShapeFrame
	case "codeblock", "_codeblock":
		return ShapeCode
	case "sticky", "postit":
		return ShapeNote
	case "input", "output", "io":
		return ShapeParallelogram
	case "process", "subprocess":
		return ShapeBox
	case "document", "doc":
		return ShapeDocument
	case "":
		return ShapeBox
	default:
		return ShapeKind(strings.ToLower(string(k)))
	}
}

// IsContainer returns true for shapes that can parent other nodes.
func IsContainer(k ShapeKind) bool {
	switch NormalizeShape(k) {
	case ShapeLane, ShapeFrame:
		return true
	default:
		return false
	}
}

// AllShapes returns documented shape names for help text.
func AllShapes() []string {
	return []string{
		"box", "card", "ellipse", "circle", "actor",
		"textbox", "note", "infobox", "callout", "info", "warning", "tip",
		"diamond", "decision", "hexagon", "octagon",
		"cylinder", "database", "cloud", "document",
		"parallelogram", "triangle", "pill", "terminal", "start", "end",
		"lane", "container", "frame", "group",
		"code",
	}
}
