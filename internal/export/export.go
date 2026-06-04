package export

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/render"
)

// Format is an output file type.
type Format string

const (
	FormatSVG    Format = "svg"
	FormatPNG    Format = "png"
	FormatPDF    Format = "pdf"
	FormatHTML   Format = "html"
	FormatSlides Format = "slides" // self-contained HTML presentation (16:9)
)

// Options for rendering.
type Options struct {
	Style RenderStyle
	Scale float64 // PNG raster scale
}

type RenderStyle string

const (
	StyleSketch   RenderStyle = "sketch"
	StylePolished RenderStyle = "polished"
)

// Write emits one format to path.
func Write(d model.Diagram, path string, format Format, opt Options) error {
	if opt.Scale <= 0 {
		opt.Scale = 2
	}
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	switch format {
	case FormatSVG:
		return os.WriteFile(path, []byte(svgContent(d, opt.Style)), 0o644)
	case FormatHTML:
		return os.WriteFile(path, []byte(render.HTML(d)), 0o644)
	case FormatPNG:
		pngData, err := RenderPNG(d, opt.Style, opt.Scale)
		if err != nil {
			return err
		}
		return os.WriteFile(path, pngData, 0o644)
	case FormatPDF:
		return WritePDF(d, path, opt)
	default:
		return fmt.Errorf("unknown format %q", format)
	}
}

// WriteDeck emits slide-oriented output (HTML deck or per-slide PNG base-1, base-2, …).
func WriteDeck(deck model.Deck, path string, format Format, opt Options) error {
	switch format {
	case FormatSlides:
		ext := filepath.Ext(path)
		if ext == "" {
			path += ".html"
		}
		return os.WriteFile(path, []byte(render.SlidesHTML(deck)), 0o644)
	case FormatPNG:
		if len(deck.Slides) == 1 {
			return Write(deck.Slides[0], path, FormatPNG, opt)
		}
		base := strings.TrimSuffix(path, filepath.Ext(path))
		for i, d := range deck.Slides {
			p := fmt.Sprintf("%s-%d.png", base, i+1)
			if err := WriteSlidePNG(d, p, opt); err != nil {
				return err
			}
		}
		return nil
	case FormatSVG:
		if len(deck.Slides) == 1 {
			svg := render.PolishedSVGSlide(deck.Slides[0])
			return os.WriteFile(path, []byte(svg), 0o644)
		}
		base := strings.TrimSuffix(path, filepath.Ext(path))
		for i, d := range deck.Slides {
			p := fmt.Sprintf("%s-%d.svg", base, i+1)
			if err := os.WriteFile(p, []byte(render.PolishedSVGSlide(d)), 0o644); err != nil {
				return err
			}
		}
		return nil
	default:
		if len(deck.Slides) > 0 {
			return Write(deck.Slides[0], path, format, opt)
		}
		return fmt.Errorf("empty deck")
	}
}

// WriteAll writes svg, png, pdf, html, and slides.html next to base path without extension.
func WriteAll(d model.Diagram, basePath string, opt Options) ([]string, error) {
	deck := model.Deck{
		Title:       d.Title,
		Subtitle:    d.Subtitle,
		SlideAspect: d.SlideAspect,
		Slides:      []model.Diagram{d},
	}
	return WriteAllDeck(deck, basePath, opt)
}

// WriteAllDeck writes all export formats for a slide deck.
func WriteAllDeck(deck model.Deck, basePath string, opt Options) ([]string, error) {
	base := strings.TrimSuffix(basePath, filepath.Ext(basePath))
	var written []string
	d := deck.Slides[0]
	for _, f := range []struct {
		ext string
		fmt Format
	}{
		{".svg", FormatSVG},
		{".png", FormatPNG},
		{".pdf", FormatPDF},
		{".html", FormatHTML},
	} {
		p := base + f.ext
		if err := Write(d, p, f.fmt, opt); err != nil {
			return written, fmt.Errorf("%s: %w", f.ext, err)
		}
		written = append(written, p)
	}
	p := base + ".slides.html"
	if err := WriteDeck(deck, p, FormatSlides, opt); err != nil {
		return written, fmt.Errorf("slides: %w", err)
	}
	written = append(written, p)
	return written, nil
}

func svgContent(d model.Diagram, style RenderStyle) string {
	if style == StylePolished {
		return render.PolishedSVG(d)
	}
	return render.SVG(d)
}
