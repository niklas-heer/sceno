package validate

import (
	"fmt"

	"github.com/niklas-heer/sceno/internal/collision"
	"github.com/niklas-heer/sceno/internal/diag"
	"github.com/niklas-heer/sceno/internal/model"
)

// Options controls build during validation.
type Options struct {
	FixCollisions bool
}

// Run loads, builds, evaluates, and returns a diagnostic report.
func Run(path string, opt Options) (diag.Report, model.Diagram, error) {
	result, report, err := LoadAndEvaluate(path, opt)
	d := model.Diagram{}
	if len(result.Slides) > 0 {
		d = result.Slides[0].Diagram
	}
	return report, d, err
}

// CollisionsOnly returns overlaps without full spec semantic checks.
func CollisionsOnly(d model.Diagram, margin float64) []diag.Issue {
	var out []diag.Issue
	for _, c := range collision.Find(d.Nodes, margin) {
		var b model.Node
		for _, n := range d.Nodes {
			if n.ID == c.B {
				b = n
				break
			}
		}
		out = append(out, diag.Issue{
			Code:     diag.CodeCollision,
			Message:  fmt.Sprintf("nodes %q and %q overlap or violate minimum clearance", c.A, c.B),
			Nodes:    []string{c.A, c.B},
			Fix:      "Apply one candidate repair and re-run validate.",
			Geometry: diag.CollisionGeometry(c),
			Repairs:  diag.CollisionRepairs(c, b),
		})
	}
	return out
}
