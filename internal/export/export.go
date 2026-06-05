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

// AllFormats is the set written by WriteAllDeck / --all.
var AllFormats = []Format{FormatSVG, FormatPNG, FormatPDF, FormatHTML, FormatSlides}

// ValidFormats lists supported -format values (excluding "all").
var ValidFormats = []string{"png", "svg", "pdf", "html", "slides"}

// ParseFormats splits a comma-separated -format value (default: png).
func ParseFormats(s string) ([]Format, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return []Format{FormatPNG}, nil
	}
	if s == "all" {
		return nil, fmt.Errorf("use --all or -format all alone for every format")
	}
	parts := strings.Split(s, ",")
	out := make([]Format, 0, len(parts))
	seen := map[Format]bool{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "slide" {
			p = "slides"
		}
		if p == "all" {
			return nil, fmt.Errorf("-format all cannot be combined with other formats; use --all")
		}
		f := Format(p)
		if !f.Valid() {
			return nil, fmt.Errorf("unknown format %q (allowed: %s, all)", p, strings.Join(ValidFormats, ", "))
		}
		if !seen[f] {
			seen[f] = true
			out = append(out, f)
		}
	}
	if len(out) == 0 {
		return []Format{FormatPNG}, nil
	}
	return out, nil
}

// Valid reports whether f is a supported export format.
func (f Format) Valid() bool {
	switch f {
	case FormatSVG, FormatPNG, FormatPDF, FormatHTML, FormatSlides:
		return true
	default:
		return false
	}
}

// Extension returns the conventional file suffix for a format.
func (f Format) Extension() string {
	switch f {
	case FormatSlides:
		return ".slides.html"
	case FormatHTML:
		return ".html"
	default:
		return "." + string(f)
	}
}

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
// Single-slide decks write base.svg, base.png, base.pdf, base.html, base.slides.html.
// Multi-slide decks write numbered base-N.svg/png per slide plus base.pdf/html (slide 1) and base.slides.html.
func WriteAllDeck(deck model.Deck, basePath string, opt Options) ([]string, error) {
	if len(deck.Slides) == 0 {
		return nil, fmt.Errorf("empty deck")
	}
	base := strings.TrimSuffix(basePath, filepath.Ext(basePath))
	var written []string

	if len(deck.Slides) == 1 {
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
	} else {
		for i, d := range deck.Slides {
			for _, f := range []struct {
				ext string
				fmt Format
			}{
				{".svg", FormatSVG},
				{".png", FormatPNG},
			} {
				p := fmt.Sprintf("%s-%d%s", base, i+1, f.ext)
				if err := Write(d, p, f.fmt, opt); err != nil {
					return written, fmt.Errorf("%s: %w", f.ext, err)
				}
				written = append(written, p)
			}
		}
		for _, f := range []struct {
			ext string
			fmt Format
		}{
			{".pdf", FormatPDF},
			{".html", FormatHTML},
		} {
			p := base + f.ext
			if err := Write(deck.Slides[0], p, f.fmt, opt); err != nil {
				return written, fmt.Errorf("%s: %w", f.ext, err)
			}
			written = append(written, p)
		}
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

// WriteFormatsDeck writes one or more formats. Single-format uses outPath (appends an
// extension when missing). Multiple formats treat outPath as a base name (extension stripped).
func WriteFormatsDeck(deck model.Deck, outPath string, formats []Format, opt Options) ([]string, error) {
	if len(deck.Slides) == 0 {
		return nil, fmt.Errorf("empty deck")
	}
	if len(formats) == 0 {
		formats = []Format{FormatPNG}
	}
	if len(formats) == 1 {
		p := resolveOutputPath(outPath, formats[0], false)
		if err := writeOneDeck(deck, p, formats[0], opt); err != nil {
			return nil, err
		}
		return []string{p}, nil
	}
	base := strings.TrimSuffix(outPath, filepath.Ext(outPath))
	var written []string
	for _, f := range formats {
		p := base + f.Extension()
		if err := writeOneDeck(deck, p, f, opt); err != nil {
			return written, fmt.Errorf("%s: %w", f, err)
		}
		written = append(written, p)
	}
	return written, nil
}

func resolveOutputPath(path string, format Format, multi bool) string {
	if multi {
		base := strings.TrimSuffix(path, filepath.Ext(path))
		return base + format.Extension()
	}
	if filepath.Ext(path) == "" {
		return path + format.Extension()
	}
	return path
}

func writeOneDeck(deck model.Deck, path string, format Format, opt Options) error {
	if format == FormatSlides {
		return WriteDeck(deck, path, format, opt)
	}
	if len(deck.Slides) == 1 {
		return Write(deck.Slides[0], path, format, opt)
	}
	return WriteDeck(deck, path, format, opt)
}
