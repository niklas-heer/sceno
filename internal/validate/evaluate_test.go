package validate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAndEvaluateOK(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "how-it-works.kdl")
	result, report, err := LoadAndEvaluate(path, Options{FixCollisions: true})
	if err != nil {
		t.Fatal(err)
	}
	if !report.OK {
		t.Fatalf("expected ok: %+v", report.Errors)
	}
	if len(result.Slides) != 1 {
		t.Fatalf("slides: %d", len(result.Slides))
	}
	engine := DeckMergedEngine(result)
	if engine.Score <= 0 {
		t.Fatalf("expected engine score")
	}
}

func TestLoadAndEvaluateInvalidSpec(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.kdl")
	// missing shapes before edge
	if err := os.WriteFile(path, []byte(`diagram {
  edge a -> b
}`), 0o644); err != nil {
		t.Fatal(err)
	}
	result, report, err := LoadAndEvaluate(path, Options{})
	if report.OK {
		t.Fatal("expected failure")
	}
	if len(result.Slides) != 0 {
		t.Fatal("expected no slides on invalid spec")
	}
	if err == nil {
		t.Fatal("expected error")
	}
}
