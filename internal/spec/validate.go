package spec

import (
	"fmt"
	"math"
	"strings"

	"github.com/niklas-heer/sceno/internal/diag"
	"github.com/niklas-heer/sceno/internal/icons"
	"github.com/niklas-heer/sceno/internal/model"
)

// MaxGridCoordinate bounds indexed grid allocation for editable source. It
// still permits far more rows/columns than a readable diagram needs.
const MaxGridCoordinate = 10000

// MaxGeometryMagnitude bounds source coordinates, dimensions, spacing, nudges,
// and font sizes in pixels. Keeping intermediate geometry within this generous
// range prevents finite source values from overflowing layout computations.
const MaxGeometryMagnitude = 1_000_000

type geometryProperty struct {
	name   string
	number float64
}

// Validate checks referential integrity and layout constraints before build.
func Validate(s model.Spec) []diag.Issue {
	for _, value := range []geometryProperty{{"gap", s.Gap}, {"padding", s.Padding}} {
		if !boundedGeometry(value.number) {
			return []diag.Issue{{Code: diag.CodeParse, Message: fmt.Sprintf("diagram %s must be a finite number between -%d and %d", value.name, MaxGeometryMagnitude, MaxGeometryMagnitude), Fix: "Use bounded numeric gap and padding values, such as gap=32 padding=24."}}
		}
	}
	if s.Layout != "" && s.Layout != model.LayoutAuto && s.Layout != model.LayoutHybrid && s.Layout != model.LayoutFree {
		return []diag.Issue{{
			Code:    diag.CodeParse,
			Message: fmt.Sprintf("unknown layout %q", s.Layout),
			Fix:     "Use layout=auto, layout=hybrid, or layout=free.",
		}}
	}
	if len(s.Slides) == 0 && len(s.Nodes) == 0 {
		if len(s.Edges) > 0 {
			return validateBody(nil, s.Edges, "", s.Layout)
		}
		return []diag.Issue{{
			Code:    diag.CodeParse,
			Message: "diagram has no shapes — add shape statements or slide { } blocks",
			Fix:     "Run `sceno init -o sceno.kdl` or add: shape box start \"Start\" at=0,0",
			Example: `shape box start "Start" icon=server at=0,0`,
		}}
	}
	if len(s.Slides) > 0 {
		var issues []diag.Issue
		for i, sl := range s.Slides {
			path := fmt.Sprintf("slide[%d]", i)
			if sl.Title == "" {
				issues = append(issues, diag.Issue{
					Code:    diag.CodeParse,
					Message: fmt.Sprintf("%s: slide missing title", path),
					Fix:     "Use: slide \"Slide title\" { ... }",
					Path:    path,
				})
			}
			issues = append(issues, validateBody(sl.Nodes, sl.Edges, path, s.Layout)...)
		}
		return issues
	}
	return validateBody(s.Nodes, s.Edges, "", s.Layout)
}

func validateBody(nodes []model.NodeSpec, edges []model.EdgeSpec, pathPrefix string, layout model.LayoutMode) []diag.Issue {
	var issues []diag.Issue
	ids := map[string]int{}
	byID := map[string]model.NodeSpec{}
	path := func(p string) string {
		if pathPrefix == "" {
			return p
		}
		return pathPrefix + "." + p
	}

	for i, n := range nodes {
		if n.ID == "" {
			issues = append(issues, diag.Issue{
				Code:    diag.CodeParse,
				Message: fmt.Sprintf("shape missing id"),
				Path:    path(fmt.Sprintf("nodes[%d]", i)),
				Fix:     "Use: shape box myid \"Label\" — every shape needs a unique id.",
			})
			continue
		}
		if prev, ok := ids[n.ID]; ok {
			issues = append(issues, diag.Issue{
				Code:    diag.CodeParse,
				Message: fmt.Sprintf("duplicate node id %q", n.ID),
				Fix:     "Rename one of the shapes to a unique id.",
				Nodes:   []string{n.ID},
				Path:    path(fmt.Sprintf("nodes[%d] (also nodes[%d])", i, prev)),
			})
		}
		ids[n.ID] = i
		byID[n.ID] = n
		if n.Row < 0 || n.Row > MaxGridCoordinate || n.Layer < 0 || n.Layer > MaxGridCoordinate {
			issues = append(issues, diag.Issue{
				Code:    diag.CodeParse,
				Message: fmt.Sprintf("node %q grid coordinates must be between 0 and %d", n.ID, MaxGridCoordinate),
				Nodes:   []string{n.ID},
				Fix:     "Use bounded nonnegative row/layer or at coordinates; start with at=0,0.",
			})
		}
		values := []geometryProperty{
			{"w", n.W}, {"h", n.H}, {"fontSize", n.FontSize}, {"dx", n.DX}, {"dy", n.DY},
		}
		if n.X != nil {
			values = append(values, geometryProperty{"x", *n.X})
		}
		if n.Y != nil {
			values = append(values, geometryProperty{"y", *n.Y})
		}
		for _, value := range values {
			if !boundedGeometry(value.number) {
				issues = append(issues, diag.Issue{Code: diag.CodeParse, Message: fmt.Sprintf("node %q %s must be a finite number between -%d and %d", n.ID, value.name, MaxGeometryMagnitude, MaxGeometryMagnitude), Nodes: []string{n.ID}, Fix: "Replace nonfinite or excessively large values with bounded numeric geometry."})
			}
		}
		if !isKnownShape(n.Kind) {
			issues = append(issues, diag.Issue{
				Code:    diag.CodeParse,
				Message: fmt.Sprintf("unknown shape %q on node %q", n.Kind, n.ID),
				Fix:     "Use: sceno docs shapes — allowed: " + strings.Join(model.AllShapes(), ", "),
				Nodes:   []string{n.ID},
			})
		}
		if n.Icon != "" && !icons.Has(n.Icon) {
			issues = append(issues, diag.Issue{
				Code:    diag.CodeUnknownIcon,
				Message: fmt.Sprintf("unknown icon %q on node %q", n.Icon, n.ID),
				Fix:     "Use: sceno docs icons — allowed: " + strings.Join(iconNames(), ", "),
				Nodes:   []string{n.ID},
				Example: `shape box api "API" icon=api`,
			})
		}
		if (n.X == nil) != (n.Y == nil) {
			issues = append(issues, diag.Issue{
				Code:    diag.CodeMissingPos,
				Message: fmt.Sprintf("node %q must set x and y together", n.ID),
				Nodes:   []string{n.ID},
				Fix:     "Add the missing coordinate, or remove both and use at/layer/row with optional dx/dy.",
				Example: `shape note context "Context" x=120 y=80`,
			})
		}
		if model.NormalizeShape(n.Kind) == model.ShapeCode && n.Code == "" && !strings.Contains(n.Label, "\n") {
			issues = append(issues, diag.Issue{
				Code:    diag.CodeParse,
				Message: fmt.Sprintf("code block %q needs source= or multiline body", n.ID),
				Nodes:   []string{n.ID},
				Fix:     "Use: code myid lang=go source=\"package main\\n\" or a quoted multiline label.",
				Example: `code demo lang=go source="package main\nfunc main() {}" at=0,0`,
			})
		}
		if n.Parent != "" && n.Parent == n.ID {
			issues = append(issues, diag.Issue{
				Code:    diag.CodeParse,
				Message: fmt.Sprintf("node %q cannot be its own parent", n.ID),
				Nodes:   []string{n.ID},
				Fix:     "Set parent= to another lane/container id.",
			})
		}
	}

	for i, e := range edges {
		if e.From == "" || e.To == "" {
			issues = append(issues, diag.Issue{
				Code:    diag.CodeParse,
				Message: "edge needs from and to node ids",
				Path:    path(fmt.Sprintf("edges[%d]", i)),
				Fix:     "Use: edge fromId -> toId",
				Example: `edge start -> end fromSide=right toSide=left`,
			})
			continue
		}
		if _, ok := ids[e.From]; !ok {
			issues = append(issues, diag.Issue{
				Code:    diag.CodeMissingNode,
				Message: fmt.Sprintf("edge references unknown node %q", e.From),
				Edge:    []string{e.From, e.To},
				Fix:     fmt.Sprintf("Add: shape box %s \"Label\" before the edge.", e.From),
				Example: fmt.Sprintf("shape box %s \"%s\"\nedge %s -> %s", e.From, e.From, e.From, e.To),
			})
		}
		if _, ok := ids[e.To]; !ok {
			issues = append(issues, diag.Issue{
				Code:    diag.CodeMissingNode,
				Message: fmt.Sprintf("edge references unknown node %q", e.To),
				Edge:    []string{e.From, e.To},
				Fix:     fmt.Sprintf("Add: shape box %s \"Label\" before the edge.", e.To),
				Example: fmt.Sprintf("shape box %s \"%s\"\nedge %s -> %s", e.To, e.To, e.From, e.To),
			})
		}
	}

	for _, n := range nodes {
		if n.Parent != "" {
			parent, ok := byID[n.Parent]
			if !ok {
				issues = append(issues, diag.Issue{
					Code:    diag.CodeMissingNode,
					Message: fmt.Sprintf("node %q parent %q not found", n.ID, n.Parent),
					Nodes:   []string{n.ID, n.Parent},
					Fix:     "Define the parent lane/container shape first in the same slide or diagram.",
				})
			} else if !model.IsContainer(parent.Kind) {
				issues = append(issues, diag.Issue{
					Code:    diag.CodeParse,
					Message: fmt.Sprintf("node %q parent %q is not a lane, frame, or container", n.ID, n.Parent),
					Nodes:   []string{n.ID, n.Parent},
					Fix:     "Use a lane/frame/group as parent, or remove parent= from the child.",
				})
			}
		}
	}

	reportedCycle := map[string]bool{}
	for _, n := range nodes {
		seen := map[string]bool{}
		id := n.ID
		for id != "" && !reportedCycle[id] {
			if seen[id] {
				issues = append(issues, diag.Issue{
					Code:    diag.CodeParse,
					Message: fmt.Sprintf("parent cycle includes node %q", id),
					Nodes:   []string{id},
					Fix:     "Remove one parent= link so container nesting forms a tree.",
				})
				for member := range seen {
					reportedCycle[member] = true
				}
				break
			}
			seen[id] = true
			parent, ok := byID[id]
			if !ok {
				break
			}
			id = parent.Parent
		}
	}

	if layout == model.LayoutFree {
		for _, n := range nodes {
			if n.X == nil || n.Y == nil {
				issues = append(issues, diag.Issue{
					Code:    diag.CodeMissingPos,
					Message: fmt.Sprintf("node %q missing x/y (layout: free)", n.ID),
					Nodes:   []string{n.ID},
					Fix:     "Add x= and y= to the shape, or use layout=auto with layer/row/at.",
					Example: `shape box n "Node" x=120 y=80`,
				})
			}
		}
	}

	return issues
}

func finiteGeometry(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

func boundedGeometry(value float64) bool {
	return finiteGeometry(value) && math.Abs(value) <= MaxGeometryMagnitude
}

func iconNames() []string {
	return icons.Names()
}
