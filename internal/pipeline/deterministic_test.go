package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niklas-heer/sceno/internal/spec"
	"github.com/niklas-heer/sceno/internal/testutil"
)

func TestBuildDeterministic(t *testing.T) {
	for _, kdl := range testutil.ValidSeeds {
		t.Run(strings.TrimSpace(strings.Split(kdl, "\n")[0]), func(t *testing.T) {
			s, err := spec.LoadKDL([]byte(kdl))
			if err != nil {
				t.Fatal(err)
			}
			assertDeterministic(t, func() (string, error) {
				r, err := BuildAndEvaluate(s, DefaultOptions())
				if err != nil {
					return "", err
				}
				return testutil.DiagramFingerprint(r.Deck.Slides[0]), nil
			})
		})
	}
}

func TestAllFixturesDeterministic(t *testing.T) {
	dir := filepath.Join("..", "..", "examples", "fixtures")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".kdl") {
			continue
		}
		t.Run(e.Name(), func(t *testing.T) {
			path := filepath.Join(dir, e.Name())
			assertDeterministic(t, func() (string, error) {
				r, err := BuildAndEvaluateFile(path, DefaultOptions())
				if err != nil {
					return "", err
				}
				return testutil.DiagramFingerprint(r.Deck.Slides[0]), nil
			})
		})
	}
}

func assertDeterministic(t *testing.T, run func() (string, error)) {
	t.Helper()
	const runs = 8
	first, err := run()
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < runs; i++ {
		h, err := run()
		if err != nil {
			t.Fatal(err)
		}
		if h != first {
			t.Fatalf("non-deterministic layout on run %d (hash %s vs %s)", i+1, first[:12], h[:12])
		}
	}
}
