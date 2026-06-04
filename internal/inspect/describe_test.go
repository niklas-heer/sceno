package inspect

import (
	"encoding/json"
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
