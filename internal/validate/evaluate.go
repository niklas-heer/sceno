package validate

import (
	"fmt"

	"github.com/niklas-heer/sceno/internal/diag"
	"github.com/niklas-heer/sceno/internal/pipeline"
	"github.com/niklas-heer/sceno/internal/scene"
	"github.com/niklas-heer/sceno/internal/spec"
)

// LoadAndEvaluate is the single entry point: parse → spec check → build → scene evaluate.
func LoadAndEvaluate(path string, opt Options) (pipeline.Result, diag.Report, error) {
	report := diag.Report{Input: path, OK: true}

	s, err := spec.LoadFile(path)
	if err != nil {
		report.OK = false
		report.Errors = append(report.Errors, diag.Issue{
			Code:    diag.CodeParse,
			Message: err.Error(),
			Fix:     "Fix KDL syntax: braces, quotes, shape/edge lines. Run `sceno docs guide --json` for examples.",
			Example: `diagram title="Test" layout=auto gap=32 {
  shape box a "A" at=0,0
}`,
		})
		finishReport(&report)
		return pipeline.Result{}, report, err
	}

	for _, iss := range spec.Validate(s) {
		report.OK = false
		report.Errors = append(report.Errors, iss)
	}
	if !report.OK {
		finishReport(&report)
		return pipeline.Result{}, report, fmt.Errorf("spec invalid")
	}

	popt := pipeline.DefaultOptions()
	popt.ResolveCollision = opt.FixCollisions
	result, err := pipeline.BuildAndEvaluate(s, popt)
	if err != nil {
		report.OK = false
		report.Errors = append(report.Errors, diag.Issue{
			Code:    diag.CodeLayout,
			Message: err.Error(),
			Fix:     "Use layout=auto with layer/row/at, or layout=free with x= and y= on every shape.",
		})
		finishReport(&report)
		return result, report, err
	}

	ApplyResult(&report, result)
	finishReport(&report)
	return result, report, nil
}

// ApplyResult maps pipeline geometry + engine findings into a diag.Report.
func ApplyResult(report *diag.Report, result pipeline.Result) {
	report.Stats.Slides = len(result.Slides)
	report.Stats.Collisions = len(result.Collisions)

	for _, c := range result.Collisions {
		report.OK = false
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

	for si, slide := range result.Slides {
		report.Stats.Nodes += len(slide.Diagram.Nodes)
		report.Stats.Edges += len(slide.Diagram.Edges)
		prefix := ""
		if len(result.Slides) > 1 {
			prefix = fmt.Sprintf("slide %d: ", si+1)
		}
		for _, f := range slide.Eval.Findings {
			iss := f.ToIssue()
			if prefix != "" {
				iss.Message = prefix + iss.Message
			}
			switch f.Severity {
			case "error":
				report.Errors = append(report.Errors, iss)
				report.OK = false
			default:
				report.Warnings = append(report.Warnings, iss)
			}
		}
	}
}

// DeckMergedEngine returns deck-level engine report for advise.
func DeckMergedEngine(result pipeline.Result) scene.EngineReport {
	return result.MergedEval().EngineReport()
}
