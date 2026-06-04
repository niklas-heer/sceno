package export

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niklas-heer/sceno/internal/pipeline"
	"github.com/niklas-heer/sceno/internal/render"
	"github.com/niklas-heer/sceno/internal/spec"
)

func TestDarkSlidesHTML(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "slides-dark.kdl")
	s, err := spec.LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	deck, colls, err := pipeline.BuildDeck(s, pipeline.DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	if len(colls) > 0 {
		t.Fatalf("collisions: %+v", colls)
	}
	html := render.SlidesHTML(deck)
	if !strings.Contains(html, "class=\"dark\"") {
		t.Fatal("expected dark body class")
	}
	if !strings.Contains(html, "--code-keyword") || !strings.Contains(html, "code-block") {
		t.Fatal("expected code block theme CSS")
	}
	if !strings.Contains(html, `code-kw">package`) && !strings.Contains(html, "package") {
		t.Fatal("expected highlighted code content")
	}
}

func TestTransparentThemeSVG(t *testing.T) {
	src := `diagram theme=light background=transparent layout=auto gap=28 {
  shape box a "A" at=0,0
  shape box b "B" at=0,1
  edge a -> b
}`
	dir := t.TempDir()
	path := filepath.Join(dir, "t.kdl")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	d, colls, err := pipeline.Build(path, pipeline.DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	if len(colls) > 0 {
		t.Fatal(colls)
	}
	svg := render.PolishedSVG(d)
	if !strings.Contains(svg, `fill="none"`) {
		t.Fatalf("expected transparent canvas in svg")
	}
}

func TestWriteDeckDarkSlides(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "slides-dark.kdl")
	s, err := spec.LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	deck, _, err := pipeline.BuildDeck(s, pipeline.DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	out := filepath.Join(dir, "deck.slides.html")
	if err := WriteDeck(deck, out, FormatSlides, Options{Style: StylePolished}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 500 {
		t.Fatalf("slides html too small: %d", len(data))
	}
}
