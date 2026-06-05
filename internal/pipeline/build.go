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
		if ns.W > 0 {
			w = ns.W
		}
		if ns.H > 0 {
			h = ns.H
		}
		n := model.Node{
			ID:       ns.ID,
			Label:    ns.Label,
			Subtitle: ns.Subtitle,
			Kind:     ns.Kind,
			Icon:     ns.Icon,
			IconPos:  ns.IconPos,
			CodeLang: ns.CodeLang,
			Code:     ns.Code,
			Fill:     ns.Fill,
			Stroke:   ns.Stroke,
			Accent:   ns.Accent,
			FontSize: ns.FontSize,
			Layer:    ns.Layer,
			Row:      ns.Row,
			Parent:   ns.Parent,
			Column:   -1,
			Rect:     model.Rect{W: w, H: h},
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

	margin := s.Gap / 2
	colls := collision.Find(d.Nodes, margin)
	preserveRow := layout.DiagramSingleRow(d.Nodes)
	if opt.ResolveCollision {
		resolveOpt := collision.ResolveOptions{PreserveSingleRowAlignment: preserveRow}
		for pass := 0; pass < 4; pass++ {
			collision.ResolveWithOptions(d.Nodes, margin, opt.MaxCollisionIters, resolveOpt)
			layout.PackColumns(&d, s.Gap)
			layout.FitParents(&d, s.Padding)
		}
		colls = collision.Find(d.Nodes, margin)
	}

	measure.FitAllNodes(d.Nodes)

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

	return d, colls, nil
}
