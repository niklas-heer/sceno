package export

import (
	"bytes"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/pipeline"
	"github.com/niklas-heer/sceno/internal/render"
	"github.com/niklas-heer/sceno/internal/spec"
)

func TestAllAndExplicitExportsUseIdenticalSlideFrames(t *testing.T) {
	for _, count := range []int{1, 2} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			deck := model.Deck{SlideAspect: render.Aspect4x3}
			for i := 0; i < count; i++ {
				deck.Slides = append(deck.Slides, model.Diagram{Title: "Slide", Nodes: []model.Node{{ID: "a", Kind: model.ShapeBox, Label: "Visible", Rect: model.Rect{X: -100, Y: 100, W: 160, H: 80}}}})
			}
			dir := t.TempDir()
			opt := Options{Style: StylePolished, Scale: 0.25}
			if _, err := WriteAllDeck(deck, filepath.Join(dir, "all"), opt); err != nil {
				t.Fatal(err)
			}
			for _, format := range []Format{FormatPNG, FormatSVG} {
				explicitDir := filepath.Join(dir, "explicit-"+string(format))
				if _, err := WriteFormatsDeck(deck, filepath.Join(explicitDir, "render"), []Format{format}, opt); err != nil {
					t.Fatal(err)
				}
				for i := 0; i < count; i++ {
					suffix := format.Extension()
					if count > 1 {
						suffix = fmt.Sprintf("-%d%s", i+1, suffix)
					}
					all, err := os.ReadFile(filepath.Join(dir, "all"+suffix))
					if err != nil {
						t.Fatal(err)
					}
					explicit, err := os.ReadFile(filepath.Join(explicitDir, "render"+suffix))
					if err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(all, explicit) {
						t.Fatalf("--all differs from explicit %s for slide %d", format, i+1)
					}
					if format == FormatPNG {
						config, err := png.DecodeConfig(bytes.NewReader(all))
						if err != nil || config.Width != 400 || config.Height != 300 {
							t.Fatalf("expected 4:3 slide at quarter scale, got %+v, %v", config, err)
						}
					}
				}
			}
			if deck.Slides[0].SlideAspect != "" {
				t.Fatal("export mutated the caller's diagram")
			}
		})
	}
}

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
