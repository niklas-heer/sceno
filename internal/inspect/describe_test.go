package inspect

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/niklas-heer/sceno/internal/diag"
)

func TestDescribeSelfService(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "self-service.kdl")
	r, err := Run(path, Options{FixCollisions: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Slides) != 1 {
		t.Fatalf("slides: %d", len(r.Slides))
	}
	if r.Tool != "sceno" || r.Version == "" {
		t.Fatalf("missing tool metadata: tool=%q version=%q", r.Tool, r.Version)
	}
	s := r.Slides[0]
	if s.Narrative == "" || s.ASCIIMap == "" || len(s.Nodes) < 5 {
		t.Fatalf("incomplete: narrative=%q nodes=%d", s.Narrative, len(s.Nodes))
	}
	if len(s.Edges) == 0 {
		t.Fatal("expected edges")
	}
	data, err := json.Marshal(r)
	if err != nil || !json.Valid(data) {
		t.Fatal("invalid json")
	}
}

func TestDescribeDoesNotLeakCollisionsAcrossReusedSlideIDs(t *testing.T) {
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
	r, err := Run(path, Options{FixCollisions: false})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Slides) != 2 || r.Slides[0].Stats.Overlaps != 0 || r.Slides[1].Stats.Overlaps != 1 {
		t.Fatalf("collision counts leaked between slides: %+v", r.Slides)
	}
	for _, p := range r.Slides[0].VisualProblems {
		if p.Code == string(diag.CodeCollision) {
			t.Fatalf("clear slide inherited another slide's collision: %+v", p)
		}
	}
}

func TestGroupIssuesPrefersStructuredSlideIndex(t *testing.T) {
	report := diag.Report{Errors: []diag.Issue{{SlideIndex: 2, Code: diag.CodeCollision, Message: "slide 1: stale display text"}}}
	grouped := groupIssues(report, 2)
	if len(grouped[0]) != 0 || len(grouped[1]) != 1 {
		t.Fatalf("structured slide provenance ignored: %+v", grouped)
	}
}

func TestDedupeVisualProblems(t *testing.T) {
	in := []VisualProblem{
		{Severity: "warning", Code: "misaligned", Message: "a"},
		{Severity: "warning", Code: "misaligned", Message: "a"},
		{Severity: "warning", Code: "edge_collision", Message: "b"},
	}
	out := dedupeVisualProblems(in)
	if len(out) != 2 {
		t.Fatalf("expected 2 unique problems, got %d", len(out))
	}
}

func TestDescribeSlidesDarkScene(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "slides-dark.kdl")
	r, err := Run(path, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Slides) != 2 {
		t.Fatalf("slides: %d", len(r.Slides))
	}
	if r.Slides[0].Scene.PaintOrder == nil {
		t.Fatal("missing scene.paint_order")
	}
	for _, s := range r.Slides {
		for _, p := range s.VisualProblems {
			if p.Code == string(diag.CodeSuggestCompact) {
				t.Fatalf("suggest_compact should not appear in visual_problems: %+v", p)
			}
		}
	}
}

func TestDescribeSlidesDemo(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "slides-demo.kdl")
	r, err := Run(path, Options{FixCollisions: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Slides) != 3 {
		t.Fatalf("slides: %d", len(r.Slides))
	}
	for _, s := range r.Slides {
		if s.Narrative == "" {
			t.Fatalf("slide %d missing narrative", s.Index)
		}
	}
}
