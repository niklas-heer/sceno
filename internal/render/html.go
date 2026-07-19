package render

import (
	"strings"

	"github.com/niklas-heer/sceno/internal/model"
)

// HTML renders a self-contained page around the canonical polished SVG.
// Keeping a single diagram renderer guarantees icon, text, route, and paint-order parity.
func HTML(d model.Diagram) string {
	useDiagramPalette(d)
	bg := paint.BgCanvas
	if paint.Transparent {
		bg = "transparent"
	}
	var b strings.Builder
	b.WriteString(`<!DOCTYPE html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">`)
	b.WriteString(`<title>` + xmlEsc(d.Title) + `</title><style>`)
	b.WriteString(`*{box-sizing:border-box}body{margin:0;min-height:100vh;display:grid;place-items:center;padding:32px 20px;background:`)
	b.WriteString(bg)
	b.WriteString(`}.diagram{width:min(100%,1600px)}.diagram svg{display:block;width:100%;height:auto;filter:drop-shadow(0 20px 50px rgb(9 9 11 / 5%))}</style></head><body><main class="diagram" aria-label="`)
	b.WriteString(xmlEsc(d.Title))
	b.WriteString(`">`)
	b.WriteString(PolishedSVG(d))
	b.WriteString(`</main></body></html>`)
	return b.String()
}
