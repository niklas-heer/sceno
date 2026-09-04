package render

import (
	"fmt"
	"strings"

	"github.com/niklas-heer/sceno/internal/fonts"
	"github.com/niklas-heer/sceno/internal/highlight"
	"github.com/niklas-heer/sceno/internal/model"
	"golang.org/x/image/font"
)

const codeFontSize = 11.0
const codeLineH = 15.0
const codePadX = 12.0
const codePadY = 10.0

func codeBlockSVG(n model.Node) string {
	body := n.Code
	if body == "" {
		body = n.Label
	}
	lang := n.CodeLang
	if lang == "" {
		lang = "text"
	}
	fill := n.Fill
	if fill == "" {
		fill = paint.BgCode
	}
	stroke := n.Stroke
	if stroke == "" {
		stroke = paint.Border
	}
	r := n.Rect
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s" stroke="%s" stroke-width="1" rx="8"/>`, r.X, r.Y, r.W, r.H, fill, stroke))
	b.WriteString(`<g xml:space="preserve">`)
	if n.Label != "" && n.Label != body {
		b.WriteString(textEl(n.Label, r.X+codePadX, r.Y+14, 11, paint.FgMuted, "600"))
	}
	x := r.X + codePadX
	y := r.Y + codePadY
	if n.Label != "" && n.Label != body {
		y += 18
	}
	lineY := y
	for _, line := range strings.Split(body, "\n") {
		line = expandCodeTabs(line)
		if line == "" {
			lineY += codeLineH
			continue
		}
		lineSpans := highlight.Tokenize(lang, line)
		lx := x
		for _, sp := range lineSpans {
			col := codeColor(sp.Kind)
			b.WriteString(textEl(sp.Text, lx, lineY, codeFontSize, col, ""))
			lx += codeSpanWidth(sp.Text)
		}
		lineY += codeLineH
	}
	b.WriteString(`</g>`)
	return b.String()
}

// Expand tabs before highlighting so each token retains the code line's
// column position, including Unicode text before an interior tab.
func expandCodeTabs(line string) string {
	var out strings.Builder
	column := 0
	for _, r := range line {
		if r == '\t' {
			n := 4 - column%4
			out.WriteString(strings.Repeat(" ", n))
			column += n
		} else {
			out.WriteRune(r)
			column++
		}
	}
	return out.String()
}

// Use fractional font advances; rounding each highlighted span compounds
// spacing errors across a line, and byte counts mismeasure Unicode entirely.
func codeSpanWidth(text string) float64 {
	face, err := fonts.Face(fonts.WeightRegular, codeFontSize)
	if err != nil {
		return fonts.TextWidth(text, codeFontSize, fonts.WeightRegular)
	}
	defer face.Close()
	return float64(font.MeasureString(face, text)) / 64
}

func codeColor(k highlight.Kind) string {
	switch k {
	case highlight.Keyword:
		return paint.CodeKeyword
	case highlight.String:
		return paint.CodeString
	case highlight.Comment:
		return paint.CodeComment
	case highlight.Number:
		return paint.CodeNumber
	default:
		return paint.CodeFg
	}
}

// CodeBlockHTML returns highlighted HTML for slide decks.
func CodeBlockHTML(n model.Node) string {
	body := n.Code
	if body == "" {
		body = n.Label
	}
	lang := n.CodeLang
	if lang == "" {
		lang = "text"
	}
	var b strings.Builder
	b.WriteString(`<pre class="code-block"><code>`)
	for _, sp := range highlight.Tokenize(lang, body) {
		cls := ""
		switch sp.Kind {
		case highlight.Keyword:
			cls = "code-kw"
		case highlight.String:
			cls = "code-str"
		case highlight.Comment:
			cls = "code-cm"
		case highlight.Number:
			cls = "code-num"
		}
		if cls != "" {
			b.WriteString(`<span class="` + cls + `">` + xmlEsc(sp.Text) + `</span>`)
		} else {
			b.WriteString(xmlEsc(sp.Text))
		}
	}
	b.WriteString(`</code></pre>`)
	return b.String()
}
