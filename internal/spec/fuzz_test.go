package spec_test

import (
	"strings"
	"testing"

	"github.com/niklas-heer/sceno/internal/pipeline"
	"github.com/niklas-heer/sceno/internal/spec"
	"github.com/niklas-heer/sceno/internal/testutil"
)

func FuzzKDLParse(f *testing.F) {
	for _, seed := range testutil.ValidSeeds {
		f.Add([]byte(seed))
	}
	f.Add([]byte("diagram { shape box x \"X\" at=0,0 }"))
	f.Add([]byte("not a diagram"))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64*1024 {
			return
		}
		_, _ = spec.LoadKDL(data)
	})
}

func FuzzKDLBuildValidSeeds(f *testing.F) {
	for _, seed := range testutil.ValidSeeds {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		// Only fuzz from known-valid prefixes to explore gap/title mutations safely.
		if len(data) < 20 || len(data) > 16*1024 {
			return
		}
		if !strings.Contains(string(data), "diagram") {
			return
		}
		s, err := spec.LoadKDL(data)
		if err != nil {
			return
		}
		if len(s.Nodes) == 0 {
			return
		}
		_, err = pipeline.BuildAndEvaluate(s, pipeline.DefaultOptions())
		if err != nil && strings.Contains(err.Error(), "layout free") {
			return
		}
		// Build must not panic; errors are acceptable for malformed specs.
	})
}
