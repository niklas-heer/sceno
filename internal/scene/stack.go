package scene

import (
	"fmt"
	"strings"

	"github.com/niklas-heer/sceno/internal/composition"
	"github.com/niklas-heer/sceno/internal/fonts"
	"github.com/niklas-heer/sceno/internal/geom"
	"github.com/niklas-heer/sceno/internal/measure"
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/theme"
)

// PlaneKind is a stacked 2D layer (back → front). Validation treats the diagram
// as multiple planes projected onto a canvas — like a lightweight 3D stack.
type PlaneKind int

const (
	PlaneBackground PlaneKind = iota
	PlaneLane
	PlaneStructure
	PlaneEdge
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
	ID            string        `json:"id"`
	Kind          string        `json:"kind"` // lane, edge, node, label, title
	Ref           string        `json:"ref,omitempty"`
	Parent        string        `json:"parent,omitempty"`
	Plane         PlaneKind     `json:"plane"`
	Z             int           `json:"z"`
	Order         int           `json:"order"`
	AllowOverlap  bool          `json:"allow_overlap,omitempty"`
	Bounds        model.Rect    `json:"bounds"`
	Outline       [][]float64   `json:"outline,omitempty"`
	InternalLines [][]float64   `json:"internal_lines,omitempty"`
	Writable      *model.Rect   `json:"writable_bounds,omitempty"`
	FontSize      float64       `json:"effective_font_size,omitempty"`
	Content       []ContentItem `json:"content,omitempty"`
}

// ContentItem exposes internal visual geometry so agents can reason about
// icons and text without reconstructing renderer measurements.
type ContentItem struct {
	Kind   string     `json:"kind"`
	Value  string     `json:"value,omitempty"`
	Bounds model.Rect `json:"bounds"`
}

// Stack is the multi-plane scene model for one laid-out diagram.
type Stack struct {
	Canvas model.Rect             `json:"canvas"`
	Planes map[string][]StackItem `json:"planes"`
}

// StackSummary is a compact view for agents.
type StackSummary struct {
	PlaneOrder []string       `json:"plane_order"`
	Counts     map[string]int `json:"counts"`
	Canvas     string         `json:"canvas,omitempty"`
}

// BuildStack assigns every layout element to a plane (paint order).
func BuildStack(d *model.Diagram) Stack {
	if d == nil {
		return Stack{Planes: map[string][]StackItem{}}
	}
	minX, minY, maxX, maxY := composition.Bounds(*d)
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

	for order, n := range d.Nodes {
		kind := string(n.Kind)
		plane := PlaneForNode(n)
		outline := geom.PathToSlices(geom.ShapeOutline(n))
		lines := geom.ShapeInternalLines(n)
		internalLines := make([][]float64, len(lines))
		for i, line := range lines {
			internalLines[i] = []float64{line[0].X, line[0].Y, line[1].X, line[1].Y}
		}
		var writable *model.Rect
		fontSize := 0.0
		if !model.IsContainer(n.Kind) && model.NormalizeShape(n.Kind) != model.ShapeCode {
			cl := measure.LayoutFor(n)
			bounds := model.Rect{X: n.Rect.X + cl.WritableX, Y: n.Rect.Y + cl.WritableY, W: cl.WritableW, H: cl.WritableH}
			writable, fontSize = &bounds, cl.FontSize
		}
		add(plane, StackItem{
			ID: n.ID, Kind: kind, Ref: n.ID, Parent: n.Parent, Bounds: n.Rect,
			Z: int(plane), Order: order, AllowOverlap: n.AllowOverlap,
			Outline: outline, InternalLines: internalLines, Writable: writable, FontSize: fontSize, Content: nodeContent(n),
		})
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
			lb := edgeLabelBounds(d, pts, re.Edge)
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
		if ch, ok := composition.ChromeBounds(*d); ok {
			add(PlaneChrome, StackItem{ID: "title", Kind: "title", Bounds: ch, Z: int(PlaneChrome)})
		}
	}
	return stack
}

func nodeContent(n model.Node) []ContentItem {
	if model.IsContainer(n.Kind) {
		if n.Label == "" {
			return nil
		}
		return []ContentItem{{
			Kind: "container_label", Value: n.Label,
			Bounds: measure.ContainerLabelBounds(n),
		}}
	}
	if model.NormalizeShape(n.Kind) == model.ShapeCode {
		return []ContentItem{{
			Kind: "code", Value: n.Code,
			Bounds: model.Rect{X: n.Rect.X + 12, Y: n.Rect.Y + 12, W: n.Rect.W - 24, H: n.Rect.H - 24},
		}}
	}

	var out []ContentItem
	if n.Icon != "" {
		x, y := measure.IconRect(n, measure.IconSize)
		out = append(out, ContentItem{Kind: "icon", Value: n.Icon, Bounds: model.Rect{X: x, Y: y, W: measure.IconSize, H: measure.IconSize}})
	}
	cl := measure.LayoutFor(n)
	fs := cl.FontSize
	if fs <= 0 {
		fs = theme.NodeSize
	}
	for i, line := range strings.Split(n.Label, "\n") {
		if line == "" {
			continue
		}
		w := measure.TextWidth(line, fs, fonts.WeightMedium)
		x := n.Rect.X + cl.WritableX + (cl.WritableW-w)/2
		if cl.InlineIcon {
			x = n.Rect.X + cl.TitleX
		}
		baseline := n.Rect.Y + cl.TitleStartY + float64(i)*cl.TitleLineH
		out = append(out, ContentItem{Kind: "title", Value: line, Bounds: model.Rect{X: x, Y: baseline - fs, W: w, H: cl.TitleLineH}})
	}
	if cl.HasSubtitle {
		subSize := cl.SubtitleSize
		if subSize <= 0 {
			subSize = theme.SubSize
		}
		w := measure.TextWidth(n.Subtitle, subSize, fonts.WeightRegular)
		x := n.Rect.X + cl.WritableX + (cl.WritableW-w)/2
		if cl.InlineIcon {
			x = n.Rect.X + cl.TitleX
		}
		out = append(out, ContentItem{
			Kind: "subtitle", Value: n.Subtitle,
			Bounds: model.Rect{X: x, Y: n.Rect.Y + cl.SubtitleY - subSize, W: w, H: subSize * 1.25},
		})
	}
	return out
}

func (s Stack) Summary() StackSummary {
	order := []string{
		PlaneBackground.String(), PlaneLane.String(), PlaneStructure.String(),
		PlaneEdge.String(), PlaneAnnotation.String(), PlaneNode.String(),
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
	// Map iteration must not change collision pair order or repair targets.
	// Project in semantic paint order, retaining source order within planes.
	for plane := PlaneBackground; plane <= PlaneChrome; plane++ {
		if want[plane] {
			out = append(out, s.Planes[plane.String()]...)
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

func edgeLabelBounds(d *model.Diagram, pts []geom.Point, edge model.Edge) model.Rect {
	if len(pts) < 2 || edge.Label == "" {
		return model.Rect{}
	}
	ctx := edgeLabelContext(d, edge)
	layout := geom.LayoutEdgeLabel(pts, edge.Label, ctx)
	x, y, w, h := layout.LabelRect()
	return model.Rect{X: x, Y: y, W: w, H: h}
}
