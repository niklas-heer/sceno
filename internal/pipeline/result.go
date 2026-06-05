package pipeline

import (
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/scene"
	"github.com/niklas-heer/sceno/internal/spec"
)

// SlideResult is one laid-out slide plus its scene evaluation.
type SlideResult struct {
	Diagram model.Diagram     `json:"diagram"`
	Eval    scene.Evaluation  `json:"evaluation"`
}

// Result is the unified build artifact: geometry + semantics for a full deck.
type Result struct {
	Deck       model.Deck        `json:"deck"`
	Collisions []model.Collision `json:"collisions,omitempty"`
	Slides     []SlideResult     `json:"slides"`
}

// BuildAndEvaluate lays out a spec and runs the scene engine on every slide.
func BuildAndEvaluate(s model.Spec, opt Options) (Result, error) {
	deck, colls, err := BuildDeck(s, opt)
	if err != nil {
		return Result{}, err
	}
	out := Result{
		Deck:       deck,
		Collisions: colls,
	}
	for i := range deck.Slides {
		d := deck.Slides[i]
		out.Slides = append(out.Slides, SlideResult{
			Diagram: d,
			Eval:    scene.Evaluate(&d),
		})
	}
	return out, nil
}

// BuildAndEvaluateFile loads a .kdl file, builds, and evaluates.
func BuildAndEvaluateFile(path string, opt Options) (Result, error) {
	s, err := spec.LoadFile(path)
	if err != nil {
		return Result{}, err
	}
	return BuildAndEvaluate(s, opt)
}

// Evaluations returns scene evaluations for all slides (deck-level advise).
func (r Result) Evaluations() []scene.Evaluation {
	out := make([]scene.Evaluation, len(r.Slides))
	for i, s := range r.Slides {
		out[i] = s.Eval
	}
	return out
}

// MergedEval returns a deck-level evaluation (min score, merged findings).
func (r Result) MergedEval() scene.Evaluation {
	return scene.MergeEvaluations(r.Evaluations())
}
