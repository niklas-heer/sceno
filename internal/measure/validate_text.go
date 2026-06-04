package measure

import (
	"fmt"

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
		out = append(out, diag.Issue{
			Code:    diag.CodeTextOverflow,
			Message: fmt.Sprintf("node %q: text overflows by %.0f×%.0f px", n.ID, ow, oh),
			Fix:     "Remove fixed w/h, shorten label, or increase fontSize; re-render auto-expands boxes by default.",
			Nodes:   []string{n.ID},
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
