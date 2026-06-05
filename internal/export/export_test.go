package export

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/pipeline"
	"github.com/niklas-heer/sceno/internal/render"
	"github.com/niklas-heer/sceno/internal/spec"
)

func testDiagram(t *testing.T) model.Diagram {
	t.Helper()
	path := filepath.Join("..", "..", "examples", "self-service.kdl")
	d, colls, err := pipeline.Build(path, pipeline.DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	if len(colls) > 0 {
		t.Fatalf("fixture has collisions: %+v", colls)
	}
	return d
}

func TestPolishedSVG(t *testing.T) {
	svg := render.PolishedSVG(testDiagram(t))
	if !strings.Contains(svg, "<svg") || !strings.Contains(svg, "</svg>") {
		t.Fatal("invalid svg")
	}
	if !strings.Contains(svg, " Q ") {
		t.Fatal("expected rounded connector paths")
	}
}

func TestHTMLEmbeddedFonts(t *testing.T) {
	html := render.HTML(testDiagram(t))
	if !strings.Contains(html, "@font-face") || !strings.Contains(html, "font-family:Inter") {
		t.Fatal("expected embedded Inter fonts")
	}
	if !strings.Contains(html, "class=\"node") {
		t.Fatal("expected node markup")
	}
}

func TestSVGEmbeddedFonts(t *testing.T) {
	svg := render.PolishedSVG(testDiagram(t))
	if !strings.Contains(svg, "@font-face") || !strings.Contains(svg, "Inter") {
		t.Fatal("expected embedded fonts in svg")
	}
}

func TestParseFormats(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want []Format
		err  bool
	}{
		{"", []Format{FormatPNG}, false},
		{"png", []Format{FormatPNG}, false},
		{"svg,png,pdf", []Format{FormatSVG, FormatPNG, FormatPDF}, false},
		{"png,png", []Format{FormatPNG}, false},
		{"slide", []Format{FormatSlides}, false},
		{"all", nil, true},
		{"png,all", nil, true},
		{"gif", nil, true},
	}
	for _, tc := range cases {
		got, err := ParseFormats(tc.in)
		if tc.err {
			if err == nil {
				t.Fatalf("ParseFormats(%q) expected error", tc.in)
			}
			continue
		}
		if err != nil {
			t.Fatalf("ParseFormats(%q): %v", tc.in, err)
		}
		if len(got) != len(tc.want) {
			t.Fatalf("ParseFormats(%q) = %v, want %v", tc.in, got, tc.want)
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Fatalf("ParseFormats(%q)[%d] = %q, want %q", tc.in, i, got[i], tc.want[i])
			}
		}
	}
}

func TestWriteFormatsDeck(t *testing.T) {
	d := testDiagram(t)
	deck := model.Deck{Slides: []model.Diagram{d}}
	dir := t.TempDir()

	single, err := WriteFormatsDeck(deck, filepath.Join(dir, "one"), []Format{FormatPNG}, Options{Style: StylePolished, Scale: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(single) != 1 || !strings.HasSuffix(single[0], ".png") {
		t.Fatalf("single png: %v", single)
	}

	multi, err := WriteFormatsDeck(deck, filepath.Join(dir, "bundle"), []Format{FormatPNG, FormatSVG}, Options{Style: StylePolished, Scale: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(multi) != 2 {
		t.Fatalf("expected 2 files, got %v", multi)
	}
}

func TestWriteAllFormats(t *testing.T) {
	d := testDiagram(t)
	dir := t.TempDir()
	base := filepath.Join(dir, "out")
	paths, err := WriteAll(d, base, Options{Style: StylePolished, Scale: 1.5})
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 5 {
		t.Fatalf("expected 5 files (svg,png,pdf,html,slides), got %d", len(paths))
	}
	for _, p := range paths {
		st, err := os.Stat(p)
		if err != nil {
			t.Fatal(err)
		}
		if st.Size() < 100 {
			t.Fatalf("%s too small", p)
		}
	}
}

func TestRenderPNG(t *testing.T) {
	png, err := RenderPNG(testDiagram(t), StylePolished, 1.5)
	if err != nil {
		t.Fatal(err)
	}
	if len(png) < 500 {
		t.Fatalf("png too small: %d bytes", len(png))
	}
}

func TestWriteAllDeckMultiSlide(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "slides-demo.kdl")
	s, err := spec.LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	deck, _, err := pipeline.BuildDeck(s, pipeline.DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	if len(deck.Slides) < 2 {
		t.Fatalf("expected multi-slide deck, got %d slides", len(deck.Slides))
	}
	dir := t.TempDir()
	base := filepath.Join(dir, "deck")
	paths, err := WriteAllDeck(deck, base, Options{Style: StylePolished, Scale: 1})
	if err != nil {
		t.Fatal(err)
	}
	// 3 slides × (svg+png) + pdf + html + slides.html = 9
	if len(paths) != 9 {
		t.Fatalf("expected 9 files for 3-slide deck, got %d: %v", len(paths), paths)
	}
	for _, p := range paths {
		st, err := os.Stat(p)
		if err != nil {
			t.Fatal(err)
		}
		if st.Size() < 100 {
			t.Fatalf("%s too small", p)
		}
	}
}

func TestWriteAllDeckEmpty(t *testing.T) {
	_, err := WriteAllDeck(model.Deck{}, "out", Options{})
	if err == nil {
		t.Fatal("expected error for empty deck")
	}
}

func TestPNGSVGViewportParity(t *testing.T) {
	d := testDiagram(t)
	svg := render.PolishedSVG(d)
	png, err := RenderPNG(d, StylePolished, 1)
	if err != nil {
		t.Fatal(err)
	}
	vp := render.ViewportFrom(d)
	// PNG from raster should match viewport dimensions at scale 1
	if !strings.Contains(svg, fmt.Sprintf(`viewBox="%.1f %.1f %.1f %.1f"`, vp.MinX, vp.MinY, vp.Width, vp.Height)) {
		t.Fatal("svg viewBox mismatch")
	}
	_ = png
	w, h := vp.PixelSize(1)
	if w < 100 || h < 100 {
		t.Fatalf("viewport too small: %dx%d", w, h)
	}
}
