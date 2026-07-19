package measure

import (
	"fmt"
	"math"
	"strconv"

	"github.com/niklas-heer/sceno/internal/diag"
	"github.com/niklas-heer/sceno/internal/model"
)

// FindTextOverflow reports nodes whose label does not fit their rect.
func FindTextOverflow(nodes []model.Node) []diag.Issue {
	var out []diag.Issue
	for _, n := range nodes {
		if model.IsContainer(n.Kind) {
			continue
		}
		ow, oh := Overflow(n)
		if ow < 1 && oh < 1 {
			continue
		}
		cl := LayoutFor(n)
		writable := model.Rect{X: n.Rect.X + cl.WritableX, Y: n.Rect.Y + cl.WritableY, W: cl.WritableW, H: cl.WritableH}
		content := textBounds(n, cl)
		format := func(v float64) string { return strconv.FormatFloat(math.Ceil(v), 'f', 0, 64) }
		out = append(out, diag.Issue{
			Code:    diag.CodeTextOverflow,
			Message: fmt.Sprintf("node %q: content still overflows its silhouette-safe writable region by %.0f×%.0f px at %.1fpx text", n.ID, ow, oh, cl.FontSize),
			Fix:     "Remove fixed w/h, add a line break, shorten the label, or apply the proposed silhouette-aware size.",
			Nodes:   []string{n.ID},
			Geometry: &diag.Geometry{Bounds: map[string]model.Rect{
				n.ID + ":shape": n.Rect, n.ID + ":writable": writable, n.ID + ":content": content,
			}},
			Repairs: []diag.RepairOption{{
				Action: "set_property", Target: n.ID,
				Properties: map[string]string{"w": format(math.Max(n.Rect.W, cl.MinW)), "h": format(math.Max(n.Rect.H, cl.MinH))},
				Reason:     "grow the shape until preferred-size text clears the stroked silhouette and internal seams",
			}},
		})
	}
	return out
}

// FitAllNodes expands rects so content fits (call after layout if sizes were fixed).
func FitAllNodes(nodes []model.Node) {
	for i := range nodes {
		EnsureNodeFits(&nodes[i])
	}
}
