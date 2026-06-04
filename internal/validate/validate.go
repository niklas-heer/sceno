package validate

import (
	"fmt"

	"github.com/niklas-heer/sceno/internal/collision"
	"github.com/niklas-heer/sceno/internal/diag"
	"github.com/niklas-heer/sceno/internal/layout"
	"github.com/niklas-heer/sceno/internal/scene"
	"github.com/niklas-heer/sceno/internal/measure"
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/pipeline"
	"github.com/niklas-heer/sceno/internal/spec"
)

// Options controls build during validation.
type Options struct {
	FixCollisions bool
}

// Run loads, builds, and returns a diagnostic report.
func Run(path string, opt Options) (diag.Report, model.Diagram, error) {
	report := diag.Report{Input: path, OK: true}

	s, err := spec.LoadFile(path)
	if err != nil {
		report.OK = false
		report.Errors = append(report.Errors, diag.Issue{
			Code:    diag.CodeParse,
			Message: err.Error(),
			Fix:     "Fix KDL syntax: braces, quotes, shape/edge lines. Run `sceno guide --json` for examples.",
			Example: `diagram title="Test" layout=auto gap=32 {
  shape box a "A" at=0,0
}`,
		})
		finishReport(&report)
		return report, model.Diagram{}, err
	}

	issues := spec.Validate(s)
	for _, iss := range issues {
		report.OK = false
		report.Errors = append(report.Errors, iss)
	}
	if !report.OK {
		finishReport(&report)
		return report, model.Diagram{}, fmt.Errorf("spec invalid")
	}

	popt := pipeline.DefaultOptions()
	popt.ResolveCollision = opt.FixCollisions
	deck, colls, err := pipeline.BuildDeck(s, popt)
	if err != nil {
		report.OK = false
		report.Errors = append(report.Errors, diag.Issue{
			Code:    diag.CodeLayout,
			Message: err.Error(),
			Fix:     "Use layout=auto with layer/row/at, or layout=free with x= and y= on every shape.",
		})
		finishReport(&report)
		return report, model.Diagram{}, err
	}
	d := model.Diagram{}
	if len(deck.Slides) > 0 {
		d = deck.Slides[0]
	}

	report.Stats.Slides = len(deck.Slides)
	for _, slide := range deck.Slides {
		report.Stats.Nodes += len(slide.Nodes)
		report.Stats.Edges += len(slide.Edges)
	}
	report.Stats.Collisions = len(colls)

	if len(colls) > 0 {
		report.OK = false
		for _, c := range colls {
			report.Errors = append(report.Errors, diag.Issue{
				Code:    diag.CodeCollision,
				Message: fmt.Sprintf("nodes %q and %q overlap", c.A, c.B),
				Fix:     "Increase diagram gap (e.g. gap=40), set different at=col,row on each shape, or separate layer values.",
				Nodes:   []string{c.A, c.B},
				Example: fmt.Sprintf(`diagram gap=40 layout=auto {
  shape box %s "%s" at=0,0
  shape box %s "%s" at=0,1
}`, c.A, c.A, c.B, c.B),
			})
		}
	}

	for si, slide := range deck.Slides {
		edgeColls := layout.FindEdgeCollisions(&slide)
		for _, ec := range edgeColls {
			iss := diag.Issue{
				Code:    diag.CodeEdgeCollision,
				Message: layout.FormatEdgeCollision(ec),
				Fix:     "Adjust layer/row, set fromSide/toSide on the edge, or increase gap.",
			}
			if len(deck.Slides) > 1 {
				iss.Message = fmt.Sprintf("slide %d: %s", si+1, iss.Message)
			}
			if ec.Kind == "node_crossing" {
				report.Errors = append(report.Errors, iss)
				report.OK = false
			} else {
				report.Warnings = append(report.Warnings, iss)
			}
		}
		for _, iss := range measure.FindTextOverflow(slide.Nodes) {
			if len(deck.Slides) > 1 {
				iss.Message = fmt.Sprintf("slide %d: %s", si+1, iss.Message)
			}
			report.Errors = append(report.Errors, iss)
			report.OK = false
		}
		report.Warnings = append(report.Warnings, layout.CompactSuggestion(&slide)...)
		sr := scene.Analyze(&slide)
		for _, iss := range sr.Issues {
			if iss.Code == diag.CodeOccluded || iss.Code == diag.CodeEdgeHidden {
				if len(deck.Slides) > 1 {
					iss.Message = fmt.Sprintf("slide %d: %s", si+1, iss.Message)
				}
				report.Warnings = append(report.Warnings, iss)
			} else if iss.Code == diag.CodeMisaligned || iss.Code == diag.CodeSuggestCompact {
				if len(deck.Slides) > 1 {
					iss.Message = fmt.Sprintf("slide %d: %s", si+1, iss.Message)
				}
				report.Warnings = append(report.Warnings, iss)
			}
		}
	}

	finishReport(&report)
	return report, d, nil
}

// CollisionsOnly returns overlaps without full spec semantic checks.
func CollisionsOnly(d model.Diagram, margin float64) []diag.Issue {
	var out []diag.Issue
	for _, c := range collision.Find(d.Nodes, margin) {
		out = append(out, diag.Issue{
			Code:    diag.CodeCollision,
			Message: fmt.Sprintf("nodes %q and %q overlap", c.A, c.B),
			Nodes:   []string{c.A, c.B},
			Fix:     "Increase gap or separate layers.",
		})
	}
	return out
}
