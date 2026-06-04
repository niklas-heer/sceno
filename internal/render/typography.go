package render

import (
	"fmt"
	"strings"

	"github.com/niklas-heer/sceno/internal/fonts"
	"github.com/niklas-heer/sceno/internal/theme"
)

func textEl(text string, x, y, size float64, color, weight string) string {
	esc := xmlEsc(text)
	w := ""
	if weight != "" {
		w = ` font-weight="` + weight + `"`
	}
	return fmt.Sprintf(`<text x="%.1f" y="%.1f" font-family="%s,ui-sans-serif,system-ui,sans-serif" font-size="%.0f" fill="%s"%s>%s</text>`,
		x, y, theme.FontFamily, size, color, w, esc)
}

func xmlEsc(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	return strings.ReplaceAll(s, ">", "&gt;")
}

// SVGFontDefs returns embedded Inter @font-face for polished/sketch SVG exports.
func SVGFontDefs() string {
	return fonts.SVGStyle()
}
