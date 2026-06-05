package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niklas-heer/sceno/internal/export"
	"github.com/niklas-heer/sceno/internal/pipeline"
	"github.com/niklas-heer/sceno/internal/render"
)

func TestRenderAuditAllExamples(t *testing.T) {
	files := collectKDLExamples(t)
	for _, path := range files {
		name := filepath.Base(path)
		t.Run(name, func(t *testing.T) {
			res, err := pipeline.BuildAndEvaluateFile(path, pipeline.DefaultOptions())
			if err != nil {
				t.Fatal(err)
			}
			d := res.Deck.Slides[0]

			svg := render.PolishedSVG(d)
			if !strings.Contains(svg, "<svg") || len(svg) < 200 {
				t.Fatalf("svg too small: %d bytes", len(svg))
			}

			png, err := export.RenderPNG(d, export.StylePolished, 1)
			if err != nil {
				t.Fatal(err)
			}
			if len(png) < 500 {
				t.Fatalf("png too small: %d bytes", len(png))
			}

			if len(d.Nodes) > 0 && !strings.Contains(svg, "<rect") && !strings.Contains(svg, "rounded") {
				t.Fatal("svg missing shape markup")
			}
		})
	}
}

func TestRenderAuditSketchFixture(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "fixtures", "sketch-organic.kdl")
	res, err := pipeline.BuildAndEvaluateFile(path, pipeline.DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	svg := render.SVG(res.Deck.Slides[0])
	if !strings.Contains(svg, "<svg") {
		t.Fatal("sketch svg missing")
	}
}

func collectKDLExamples(t *testing.T) []string {
	t.Helper()
	root := filepath.Join("..", "..", "examples")
	var files []string
	for _, dir := range []string{root, filepath.Join(root, "fixtures")} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".kdl") {
				continue
			}
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	if len(files) < 12 {
		t.Fatalf("expected at least 12 example kdl files, got %d", len(files))
	}
	return files
}
