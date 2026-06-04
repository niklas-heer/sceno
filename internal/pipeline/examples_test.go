package pipeline

import (
	"path/filepath"
	"testing"

	"github.com/niklas-heer/sceno/internal/spec"
)

func TestBuildAllExamples(t *testing.T) {
	examples := []string{
		"self-service.kdl",
		"slides-demo.kdl",
		"slides-dark.kdl",
		"shapes-demo.kdl",
	}
	for _, name := range examples {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join("..", "..", "examples", name)
			s, err := spec.LoadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if len(s.Slides) > 0 {
				deck, colls, err := BuildDeck(s, DefaultOptions())
				if err != nil {
					t.Fatal(err)
				}
				if len(deck.Slides) == 0 {
					t.Fatal("empty deck")
				}
				_ = colls
				return
			}
			_, colls, err := BuildFromSpec(s, DefaultOptions())
			if err != nil {
				t.Fatal(err)
			}
			_ = colls
		})
	}
}
