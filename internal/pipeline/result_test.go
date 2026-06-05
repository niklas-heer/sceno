package pipeline

import (
	"path/filepath"
	"testing"

	"github.com/niklas-heer/sceno/internal/spec"
)

func TestBuildAndEvaluateHowItWorks(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "how-it-works.kdl")
	s, err := spec.LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	opt := DefaultOptions()
	opt.ResolveCollision = true
	result, err := BuildAndEvaluate(s, opt)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Slides) != 1 {
		t.Fatalf("slides: %d", len(result.Slides))
	}
	if result.Slides[0].Eval.Score <= 0 {
		t.Fatalf("expected positive visual score")
	}
	if len(result.Slides[0].Eval.PaintOrder) == 0 {
		t.Fatal("expected paint order")
	}
	merged := result.MergedEval()
	if merged.Score != result.Slides[0].Eval.Score {
		t.Fatalf("merged score mismatch")
	}
}

func TestBuildAndEvaluateFile(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "self-service.kdl")
	result, err := BuildAndEvaluateFile(path, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Slides) != 1 || len(result.Slides[0].Diagram.Nodes) < 5 {
		t.Fatalf("unexpected deck: slides=%d nodes=%d", len(result.Slides), len(result.Slides[0].Diagram.Nodes))
	}
}
