package render

import (
	"fmt"
	"strings"

	"github.com/niklas-heer/sceno/internal/highlight"
	"github.com/niklas-heer/sceno/internal/model"
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
		if line == "" {
			lineY += codeLineH
			continue
		}
		lineSpans := highlight.Tokenize(lang, line)
		lx := x
		for _, sp := range lineSpans {
			col := codeColor(sp.Kind)
			b.WriteString(textEl(sp.Text, lx, lineY, codeFontSize, col, ""))
			lx += float64(len(sp.Text)) * codeFontSize * 0.58
		}
		lineY += codeLineH
	}
	return b.String()
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

func codeNodesForHTML(d model.Diagram) []model.Node {
	var out []model.Node
	for _, n := range d.Nodes {
		if model.NormalizeShape(n.Kind) == model.ShapeCode {
			out = append(out, n)
		}
	}
	return out
}
