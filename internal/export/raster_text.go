package export

import (
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/niklas-heer/sceno/internal/fonts"
	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

var rasterFonts struct {
	sync.Once
	faces [4]*sfnt.Font
	err   error
}

var scalarSVGScale = regexp.MustCompile(`\bscale\(\s*([+-]?(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)(?:[eE][+-]?[0-9]+)?)\s*\)`)

// prepareSVGForRaster replaces Sceno's simple text elements in place. oksvg does
// not render text or embedded font-face rules. Glyph paths preserve the exact
// SVG baseline, weight, color, enclosing transforms, and semantic paint order.
// The public SVG remains searchable text with embedded fonts.
func prepareSVGForRaster(svg string) (string, error) {
	decoder := xml.NewDecoder(strings.NewReader(svg))
	var out strings.Builder
	copied := int64(0)
	for {
		start := decoder.InputOffset()
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		el, ok := token.(xml.StartElement)
		if !ok {
			continue
		}
		if el.Name.Local == "style" {
			var css string
			if err := decoder.DecodeElement(&css, &el); err != nil {
				return "", err
			}
			if strings.Contains(css, "@font-face") {
				// Glyphs are already embedded as paths. oksvg's CSS parser
				// cannot parse data-URL font faces and would reject the scene.
				out.WriteString(svg[copied:start])
				copied = decoder.InputOffset()
			}
			continue
		}
		if el.Name.Local != "text" {
			end := decoder.InputOffset()
			raw := svg[start:end]
			var rx, ry string
			// oksvg interprets scale(s) as scale(s, 0), collapsing inline
			// icons and complete slide scenes. Keep both axes explicit.
			for _, a := range el.Attr {
				if a.Name.Local == "rx" {
					rx = a.Value
				}
				if a.Name.Local == "ry" {
					ry = a.Value
				}
				if a.Name.Local != "transform" {
					continue
				}
				value := scalarSVGScale.ReplaceAllString(a.Value, "scale($1,$1)")
				if value == a.Value {
					continue
				}
				raw = strings.Replace(raw, a.Value, value, 1)
			}
			// SVG inherits an omitted rectangle radius from the other axis;
			// oksvg otherwise leaves it zero and paints square corners.
			if el.Name.Local == "rect" && (rx == "") != (ry == "") {
				key, value := "ry", rx
				if rx == "" {
					key, value = "rx", ry
				}
				pos := strings.LastIndex(raw, ">")
				if pos > 0 && raw[pos-1] == '/' {
					pos--
				}
				raw = raw[:pos] + " " + key + `="` + value + `"` + raw[pos:]
			}
			if raw != svg[start:end] {
				out.WriteString(svg[copied:start])
				out.WriteString(raw)
				copied = end
			}
			continue
		}
		var label string
		if err := decoder.DecodeElement(&label, &el); err != nil {
			return "", err
		}
		attrs := make(map[string]string, len(el.Attr))
		for _, a := range el.Attr {
			attrs[a.Name.Local] = a.Value
		}
		x, err := svgTextNumber(attrs, "x", 0)
		if err != nil {
			return "", err
		}
		y, err := svgTextNumber(attrs, "y", 0)
		if err != nil {
			return "", err
		}
		size, err := svgTextNumber(attrs, "font-size", 16)
		if err != nil {
			return "", err
		}
		path, err := textGlyphPath(label, x, y, size, fonts.WeightFromCSS(attrs["font-weight"]))
		if err != nil {
			return "", err
		}
		out.WriteString(svg[copied:start])
		out.WriteString(`<path d="` + path + `"`)
		for _, a := range el.Attr {
			switch a.Name.Local {
			case "fill", "fill-opacity", "opacity", "transform", "clip-path":
				out.WriteString(" " + a.Name.Local + `="`)
				value := a.Value
				if a.Name.Local == "transform" {
					value = scalarSVGScale.ReplaceAllString(value, "scale($1,$1)")
				}
				if err := xml.EscapeText(&out, []byte(value)); err != nil {
					return "", err
				}
				out.WriteByte('"')
			}
		}
		out.WriteString(`/>`)
		copied = decoder.InputOffset()
	}
	if copied == 0 {
		return svg, nil
	}
	out.WriteString(svg[copied:])
	return out.String(), nil
}

func svgTextNumber(attrs map[string]string, key string, fallback float64) (float64, error) {
	if attrs[key] == "" {
		return fallback, nil
	}
	v, err := strconv.ParseFloat(attrs[key], 64)
	if err != nil {
		return 0, fmt.Errorf("SVG text %s: %w", key, err)
	}
	return v, nil
}

func textGlyphPath(text string, x, y, size float64, weight fonts.Weight) (string, error) {
	rasterFonts.Do(func() {
		for i, data := range [][]byte{fonts.RegularBytes(), fonts.MediumBytes(), fonts.SemiBoldBytes(), fonts.BoldBytes()} {
			rasterFonts.faces[i], rasterFonts.err = sfnt.Parse(data)
			if rasterFonts.err != nil {
				return
			}
		}
	})
	if rasterFonts.err != nil {
		return "", rasterFonts.err
	}
	f := rasterFonts.faces[weight]
	var buf sfnt.Buffer
	ppem := fixed.Int26_6(size * 64)
	var path strings.Builder
	var previous sfnt.GlyphIndex
	for i, r := range text {
		glyph, err := f.GlyphIndex(&buf, r)
		if err != nil {
			return "", err
		}
		if i > 0 {
			kern, err := f.Kern(&buf, previous, glyph, ppem, font.HintingNone)
			if err != nil && err != sfnt.ErrNotFound {
				return "", err
			}
			x += float64(kern) / 64
		}
		segments, err := f.LoadGlyph(&buf, glyph, ppem, nil)
		if err != nil {
			return "", err
		}
		for _, segment := range segments {
			var op string
			var count int
			switch segment.Op {
			case sfnt.SegmentOpMoveTo:
				if path.Len() > 0 {
					path.WriteByte('Z')
				}
				op, count = "M", 1
			case sfnt.SegmentOpLineTo:
				op, count = "L", 1
			case sfnt.SegmentOpQuadTo:
				op, count = "Q", 2
			case sfnt.SegmentOpCubeTo:
				op, count = "C", 3
			}
			path.WriteString(op)
			for j := 0; j < count; j++ {
				p := segment.Args[j]
				fmt.Fprintf(&path, " %.3f %.3f", x+float64(p.X)/64, y+float64(p.Y)/64)
			}
		}
		advance, err := f.GlyphAdvance(&buf, glyph, ppem, font.HintingNone)
		if err != nil {
			return "", err
		}
		x += float64(advance) / 64
		previous = glyph
	}
	if path.Len() > 0 {
		path.WriteByte('Z')
	}
	return path.String(), nil
}
