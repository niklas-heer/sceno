package spec_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niklas-heer/sceno/internal/pipeline"
	"github.com/niklas-heer/sceno/internal/spec"
)

// Mutation tests (fault injection): single-char flips on valid KDL must never panic.
func TestMutationNoPanic(t *testing.T) {
	root := filepath.Join("..", "..", "examples", "fixtures")
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	mutators := []byte{'#', ' ', '\n', 'X', '0', '{', '}', '"'}

	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".kdl") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		step := len(data) / 40
		if step < 1 {
			step = 1
		}
		t.Run(e.Name(), func(t *testing.T) {
			for i := 0; i < len(data); i += step {
				for _, m := range mutators {
					if data[i] == m {
						continue
					}
					mutated := append([]byte(nil), data...)
					mutated[i] = m
					func() {
						defer func() {
							if r := recover(); r != nil {
								t.Fatalf("panic at byte %d mut=%q: %v", i, m, r)
							}
						}()
						s, err := spec.LoadKDL(mutated)
						if err != nil {
							return
						}
						_, _ = pipeline.BuildAndEvaluate(s, pipeline.DefaultOptions())
					}()
				}
			}
		})
	}
}

func TestMutationValidSpecStillBuilds(t *testing.T) {
	base := `diagram layout=auto gap=32 {
  shape box a "A" at=0,0
  shape box b "B" at=1,0
  edge a -> b
}`
	s, err := spec.LoadKDL([]byte(base))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pipeline.BuildAndEvaluate(s, pipeline.DefaultOptions()); err != nil {
		t.Fatal(err)
	}
}
