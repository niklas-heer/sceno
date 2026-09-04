package validate

import (
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/niklas-heer/sceno/internal/diag"
)

func TestEvaluateSourceMatchesFilePipeline(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "slides-dark.kdl")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	fileResult, fileReport, err := LoadAndEvaluate(path, Options{FixCollisions: true})
	if err != nil {
		t.Fatal(err)
	}
	sourceResult, sourceReport, err := EvaluateSource(data, path, Options{FixCollisions: true})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(sourceResult, fileResult) || !reflect.DeepEqual(sourceReport, fileReport) {
		t.Fatal("in-memory evaluation differs from file evaluation")
	}
	if _, report, err := EvaluateSource(data, "/not/a/real/file.kdl", Options{FixCollisions: true}); err != nil || !report.OK {
		t.Fatalf("EvaluateSource attempted to load its display name: %v", err)
	}
}

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

func TestDeckCollisionRepairsUseTheirOwnSlide(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reused-ids.kdl")
	data := `diagram layout=free {
  slide "Clear" {
    shape box a "A" x=-300 y=150 w=120 h=80
    shape box b "B" x=100 y=150 w=120 h=80
  }
  slide "Collision" {
    shape box a "A" x=70 y=150 w=120 h=80
    shape box b "B" x=80 y=150 w=120 h=80 dx=20
  }
}`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	result, report, err := LoadAndEvaluate(path, Options{FixCollisions: false})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Collisions) != 1 || result.Collisions[0].SlideIndex != 2 {
		t.Fatalf("collision provenance = %+v", result.Collisions)
	}
	// The reused node has identical final bounds but different source offsets.
	first, second := result.Slides[0].Diagram.Nodes[1], result.Slides[1].Diagram.Nodes[1]
	if first.Rect != second.Rect || first.DX == second.DX {
		t.Fatalf("fixture must have ambiguous geometry and distinct source offsets: %+v %+v", first, second)
	}
	if report.OK {
		t.Fatal("deck with a collision should fail validation")
	}
	found := false
	for _, issue := range report.Errors {
		if issue.Code != diag.CodeCollision {
			continue
		}
		found = true
		if issue.SlideIndex != 2 {
			t.Fatalf("collision issue has wrong slide: %+v", issue)
		}
		want := second.Rect.X - second.DX + result.Collisions[0].MoveBX
		got, err := strconv.ParseFloat(issue.Repairs[0].Properties["x"], 64)
		if err != nil || got != want {
			t.Fatalf("repair uses wrong slide's coordinates: got %v, want %v (%v)", got, want, err)
		}
	}
	if !found {
		t.Fatal("missing collision issue")
	}
	for _, issue := range report.Warnings {
		if issue.SlideIndex < 1 || issue.SlideIndex > 2 {
			t.Fatalf("finding lacks slide provenance: %+v", issue)
		}
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
