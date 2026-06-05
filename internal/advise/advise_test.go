package advise

import (
	"path/filepath"
	"testing"
)

func TestAdviseHowItWorks(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "how-it-works.kdl")
	report, err := Run(path, Options{FixCollisions: true})
	if err != nil {
		t.Fatal(err)
	}
	if !report.ValidationOK {
		t.Fatalf("expected valid spec")
	}
	if report.VisualScore <= 0 {
		t.Fatalf("expected positive visual score")
	}
	if len(report.Engine.RulesRun) < 5 {
		t.Fatalf("expected engine rules")
	}
}
