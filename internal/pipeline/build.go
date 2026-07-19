package pipeline

import (
	"fmt"

	"github.com/niklas-heer/sceno/internal/collision"
	"github.com/niklas-heer/sceno/internal/layout"
	"github.com/niklas-heer/sceno/internal/measure"
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/spec"
)

type Options struct {
	MaxCollisionIters int
	ResolveCollision  bool
}

func DefaultOptions() Options {
	return Options{MaxCollisionIters: 120, ResolveCollision: true}
}

// Build loads a spec file and produces a diagram.
func Build(path string, opt Options) (model.Diagram, []model.Collision, error) {
	s, err := spec.LoadFile(path)
	if err != nil {
		return model.Diagram{}, nil, err
	}
	return BuildFromSpec(s, opt)
}

// BuildFromSpec builds from an already-loaded spec.
func BuildFromSpec(s model.Spec, opt Options) (model.Diagram, []model.Collision, error) {
	nodes := make([]model.Node, 0, len(s.Nodes))
	for _, ns := range s.Nodes {
		w, h := measure.FitSize(ns)
		n := model.Node{
			ID:           ns.ID,
			Label:        ns.Label,
			Subtitle:     ns.Subtitle,
			Kind:         ns.Kind,
			Icon:         ns.Icon,
			IconPos:      ns.IconPos,
			CodeLang:     ns.CodeLang,
			Code:         ns.Code,
			Fill:         ns.Fill,
			Stroke:       ns.Stroke,
			Accent:       ns.Accent,
			FontSize:     ns.FontSize,
			Layer:        ns.Layer,
			Row:          ns.Row,
			AtSet:        ns.AtSet,
			DX:           ns.DX,
			DY:           ns.DY,
			AllowOverlap: ns.AllowOverlap,
			Parent:       ns.Parent,
			MinW:         ns.W,
			MinH:         ns.H,
			Column:       -1,
			Rect:         model.Rect{W: w, H: h},
		}
		if ns.X != nil && ns.Y != nil {
			n.Rect.X = *ns.X
			n.Rect.Y = *ns.Y
			n.Fixed = true
		}
		nodes = append(nodes, n)
	}

	d := model.Diagram{
		Title:       s.Title,
		Subtitle:    s.Subtitle,
		Layout:      s.Layout,
		Style:       s.Style,
		Gap:         s.Gap,
		Padding:     s.Padding,
		SlideAspect: s.SlideAspect,
		Theme:       s.Theme,
		Nodes:       nodes,
	}
	for _, es := range s.Edges {
		d.Edges = append(d.Edges, model.Edge{
			From:     es.From,
			To:       es.To,
			Label:    es.Label,
			FromSide: es.FromSide,
			ToSide:   es.ToSide,
			Dashed:   es.Dashed,
			Color:    es.Color,
		})
	}

	switch s.Layout {
	case model.LayoutFree:
		for i := range d.Nodes {
			if !d.Nodes[i].Fixed {
				return d, nil, fmt.Errorf("layout free: node %q missing x and y", d.Nodes[i].ID)
			}
		}
	default:
		layout.Grid(&d, s.Gap)
		layout.PackColumns(&d, s.Gap)
	}

	layout.FitParents(&d, s.Padding)
	layout.AlignRows(&d, s.Gap)

	margin := s.Gap / 2
	colls := collision.Find(d.Nodes, margin)
	preserveRow := layout.DiagramSingleRow(d.Nodes)
	if opt.ResolveCollision {
		resolveOpt := collision.ResolveOptions{PreserveSingleRowAlignment: preserveRow}
		for pass := 0; pass < 4; pass++ {
			collision.ResolveWithOptions(d.Nodes, margin, opt.MaxCollisionIters, resolveOpt)
			layout.PackColumns(&d, s.Gap)
			layout.FitParents(&d, s.Padding)
			layout.AlignRows(&d, s.Gap)
		}
		colls = collision.Find(d.Nodes, margin)
	}

	measure.FitAllNodes(d.Nodes)
	layout.AlignRows(&d, s.Gap)
	layout.FitParents(&d, s.Padding)

	layout.RouteEdges(&d)
	for i := 0; i < 16; i++ {
		layout.RerouteCollidingEdges(&d)
		cross := layout.FindEdgeCollisions(&d)
		hasNode := false
		for _, c := range cross {
			if c.Kind == "node_crossing" {
				hasNode = true
				break
			}
		}
		if !hasNode {
			break
		}
	}

	colls = collision.Find(d.Nodes, margin)
	measure.ApplyInteriors(d.Nodes)
	measure.TightenToInterior(d.Nodes)
	measure.ApplyInteriors(d.Nodes)
	layout.AlignRows(&d, s.Gap)
	layout.PackColumns(&d, s.Gap)
	applyNudges(d.Nodes)
	layout.FitParents(&d, s.Padding)
	layout.RouteEdges(&d)
	for i := 0; i < 8; i++ {
		layout.RerouteCollidingEdges(&d)
	}
	colls = collision.Find(d.Nodes, margin)
	return d, colls, nil
}

// applyNudges runs after collision-safe auto layout. A nudge is an explicit
// author override, so any collision it introduces is reported rather than
// silently undoing the requested movement. Nudging a container moves its
// complete subtree as one visual group.
func applyNudges(nodes []model.Node) {
	byParent := map[string][]int{}
	for i := range nodes {
		byParent[nodes[i].Parent] = append(byParent[nodes[i].Parent], i)
	}
	var shift func(string, float64, float64)
	shift = func(parent string, dx, dy float64) {
		for _, i := range byParent[parent] {
			nodes[i].Rect.X += dx
			nodes[i].Rect.Y += dy
			shift(nodes[i].ID, dx, dy)
		}
	}
	for i := range nodes {
		if nodes[i].DX == 0 && nodes[i].DY == 0 {
			continue
		}
		nodes[i].Rect.X += nodes[i].DX
		nodes[i].Rect.Y += nodes[i].DY
		shift(nodes[i].ID, nodes[i].DX, nodes[i].DY)
	}
}
