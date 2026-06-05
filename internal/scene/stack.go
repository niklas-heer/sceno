package scene

import (
	"fmt"
	"math"

	"github.com/niklas-heer/sceno/internal/geom"
	"github.com/niklas-heer/sceno/internal/model"
)

// PlaneKind is a stacked 2D layer (back → front). Validation treats the diagram
// as multiple planes projected onto a canvas — like a lightweight 3D stack.
type PlaneKind int

const (
	PlaneBackground PlaneKind = iota
	PlaneLane
	PlaneEdge
	PlaneStructure
	PlaneAnnotation
	PlaneNode
	PlaneLabel
	PlaneChrome
)

func (p PlaneKind) String() string {
	switch p {
	case PlaneBackground:
		return "background"
	case PlaneLane:
		return "lanes"
	case PlaneEdge:
		return "edges"
	case PlaneStructure:
		return "structure"
	case PlaneAnnotation:
		return "annotations"
	case PlaneNode:
		return "nodes"
	case PlaneLabel:
		return "labels"
	case PlaneChrome:
		return "chrome"
	default:
		return "unknown"
	}
}

// StackItem is one drawable on a plane.
type StackItem struct {
	ID     string      `json:"id"`
	Kind   string      `json:"kind"` // lane, edge, node, label, title
	Ref    string      `json:"ref,omitempty"`
	Plane  PlaneKind   `json:"plane"`
	Z      int         `json:"z"`
	Bounds model.Rect  `json:"bounds"`
}

// Stack is the multi-plane scene model for one laid-out diagram.
type Stack struct {
	Canvas model.Rect           `json:"canvas"`
	Planes map[string][]StackItem `json:"planes"`
}

// StackSummary is a compact view for agents.
type StackSummary struct {
	PlaneOrder []string         `json:"plane_order"`
	Counts     map[string]int   `json:"counts"`
	Canvas     string           `json:"canvas,omitempty"`
}

// BuildStack assigns every layout element to a plane (paint order).
func BuildStack(d *model.Diagram) Stack {
	if d == nil {
		return Stack{Planes: map[string][]StackItem{}}
	}
	minX, minY, maxX, maxY := canvasExtents(d)
	stack := Stack{
		Canvas: model.Rect{X: minX, Y: minY, W: maxX - minX, H: maxY - minY},
		Planes: map[string][]StackItem{},
	}
	add := func(plane PlaneKind, item StackItem) {
		item.Plane = plane
		key := plane.String()
		stack.Planes[key] = append(stack.Planes[key], item)
	}

	add(PlaneBackground, StackItem{ID: "canvas", Kind: "background", Bounds: stack.Canvas, Z: int(PlaneBackground)})

	for _, n := range d.Nodes {
		kind := string(n.Kind)
		switch model.NormalizeShape(n.Kind) {
		case model.ShapeLane:
			add(PlaneLane, StackItem{ID: n.ID, Kind: "lane", Ref: n.ID, Bounds: n.Rect, Z: int(PlaneLane)})
		case model.ShapeFrame:
			add(PlaneStructure, StackItem{ID: n.ID, Kind: "frame", Ref: n.ID, Bounds: n.Rect, Z: int(PlaneStructure)})
		case model.ShapeInfobox, model.ShapeNote, model.ShapeTextbox:
			add(PlaneAnnotation, StackItem{ID: n.ID, Kind: kind, Ref: n.ID, Bounds: n.Rect, Z: int(PlaneAnnotation)})
		default:
			if model.IsContainer(n.Kind) {
				add(PlaneStructure, StackItem{ID: n.ID, Kind: kind, Ref: n.ID, Bounds: n.Rect, Z: int(PlaneStructure)})
			} else {
				add(PlaneNode, StackItem{ID: n.ID, Kind: kind, Ref: n.ID, Bounds: n.Rect, Z: int(PlaneNode)})
			}
		}
	}

	pad := d.Gap * 0.35
	if pad < 6 {
		pad = 6
	}
	for _, re := range d.Routed {
		pts := pathToGeom(re.Points)
		b := pathBounds(pts, pad)
		add(PlaneEdge, StackItem{
			ID:     re.Key,
			Kind:   "edge",
			Ref:    re.Key,
			Bounds: b,
			Z:      int(PlaneEdge),
		})
		if re.Edge.Label != "" {
			lb := edgeLabelBounds(pts, re.Edge.Label, d.Gap)
			add(PlaneLabel, StackItem{
				ID:     re.Key + ":label",
				Kind:   "edge_label",
				Ref:    re.Key,
				Bounds: lb,
				Z:      int(PlaneLabel),
			})
		}
	}

	if d.Title != "" || d.Subtitle != "" {
		ch := titleChromeBounds(d, stack.Canvas)
		add(PlaneChrome, StackItem{ID: "title", Kind: "title", Bounds: ch, Z: int(PlaneChrome)})
	}
	return stack
}

func (s Stack) Summary() StackSummary {
	order := []string{
		PlaneBackground.String(), PlaneLane.String(), PlaneEdge.String(),
		PlaneStructure.String(), PlaneAnnotation.String(), PlaneNode.String(),
		PlaneLabel.String(), PlaneChrome.String(),
	}
	counts := map[string]int{}
	for k, items := range s.Planes {
		counts[k] = len(items)
	}
	canvas := ""
	if s.Canvas.W > 0 && s.Canvas.H > 0 {
		canvas = fmt.Sprintf("%.0f×%.0f", s.Canvas.W, s.Canvas.H)
	}
	return StackSummary{PlaneOrder: order, Counts: counts, Canvas: canvas}
}

// Project merges selected planes onto one 2D obstacle set (for routing / blocking checks).
func (s Stack) Project(planes ...PlaneKind) []StackItem {
	var out []StackItem
	want := map[PlaneKind]bool{}
	for _, p := range planes {
		want[p] = true
	}
	for _, items := range s.Planes {
		for _, it := range items {
			if want[it.Plane] {
				out = append(out, it)
			}
		}
	}
	return out
}

func pathBounds(pts []geom.Point, pad float64) model.Rect {
	if len(pts) == 0 {
		return model.Rect{}
	}
	minX, minY := pts[0].X, pts[0].Y
	maxX, maxY := pts[0].X, pts[0].Y
	for _, p := range pts[1:] {
		if p.X < minX {
			minX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	return model.Rect{X: minX - pad, Y: minY - pad, W: maxX - minX + pad*2, H: maxY - minY + pad*2}
}

func edgeLabelBounds(pts []geom.Point, label string, gap float64) model.Rect {
	if len(pts) < 2 || label == "" {
		return model.Rect{}
	}
	w := float64(len(label))*7 + 16
	h := 18.0
	mid := len(pts) / 2
	a, b := pts[mid-1], pts[mid]
	cx := (a.X + b.X) / 2
	cy := (a.Y + b.Y) / 2
	if math.Abs(a.Y-b.Y) < 1 {
		return model.Rect{X: cx - w/2, Y: cy - h - gap*0.3, W: w, H: h}
	}
	return model.Rect{X: cx + gap*0.3, Y: cy - h/2, W: w, H: h}
}

func titleChromeBounds(d *model.Diagram, canvas model.Rect) model.Rect {
	h := 0.0
	if d.Title != "" {
		h += 48
	}
	if d.Subtitle != "" {
		h += 28
	}
	if h > 0 {
		h += 16
	}
	return model.Rect{X: canvas.X, Y: canvas.Y, W: canvas.W, H: h}
}
