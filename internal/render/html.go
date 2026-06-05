package render

import (
	"fmt"
	"strings"

	"github.com/niklas-heer/sceno/internal/fonts"
	"github.com/niklas-heer/sceno/internal/icons"
	"github.com/niklas-heer/sceno/internal/model"
)

// HTML renders a self-contained polished page (embedded Inter, shadcn-style tokens).
func HTML(d model.Diagram) string {
	useDiagramPalette(d)
	minX, minY, maxX, maxY := Bounds(d)
	ox, oy := minX-48, minY-48
	cw := maxX - minX + 96
	ch := maxY - minY + 96
	var b strings.Builder
	b.WriteString("<!DOCTYPE html><html lang=\"en\"><head><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width,initial-scale=1\">")
	fmt.Fprintf(&b, "<title>%s</title>", xmlEsc(d.Title))
	b.WriteString(fonts.HTMLStyle())
	b.WriteString("<style>")
	b.WriteString(paint.CSSVars())
	b.WriteString(`
*{box-sizing:border-box}body{margin:0;font-family:Inter,ui-sans-serif,system-ui,sans-serif;background:var(--background);color:var(--foreground);-webkit-font-smoothing:antialiased}
.viewport{width:100%;min-height:100vh;overflow:auto;padding:32px 20px 48px}
.canvas{position:relative;margin:0 auto;background:var(--card);border:1px solid var(--border);border-radius:calc(var(--radius) * 2);box-shadow:0 1px 2px var(--ring),0 20px 50px rgb(9 9 11 / 5%)}
.edges,.edge-labels{position:absolute;inset:0;pointer-events:none}
.edges{z-index:1}
.edge-labels{z-index:3}
.node{position:absolute;z-index:2;border:1px solid var(--border);background:var(--card);border-radius:calc(var(--radius) + 4px);padding:14px 16px 14px 44px;box-shadow:0 1px 2px var(--ring);font-size:13px;line-height:1.5;white-space:pre-wrap;font-weight:500}
.node .title{font-weight:600;letter-spacing:-0.01em}.node .sub{font-size:11px;color:var(--muted-foreground);margin-top:4px;display:block;font-weight:400}
.node .ico{position:absolute;left:14px;top:50%;transform:translateY(-50%);width:22px;height:22px;color:var(--muted-foreground)}
.node.ellipse{border-radius:9999px;text-align:center;padding:14px 20px}.node.actor{border-radius:14px;text-align:center;padding:52px 16px 14px}
.node.textbox{background:var(--muted);border-radius:var(--radius);padding:12px 14px}
.node.infobox{border-left:4px solid var(--accent);padding-left:16px;border-radius:calc(var(--radius) + 2px)}
.node.lane,.node.frame{z-index:0;background:var(--muted);border-style:dashed;border-radius:16px;padding-top:38px;box-shadow:none}
.node.frame{border-style:solid;opacity:.92}
.node.lane .lane-lbl{position:absolute;left:14px;top:12px;font-size:10px;font-weight:600;letter-spacing:.08em;text-transform:uppercase;color:var(--muted-foreground)}
.node.diamond{clip-path:polygon(50% 0%,100% 50%,50% 100%,0% 50%)}
.header{position:absolute;left:32px;top:28px;z-index:2;pointer-events:none}
.header h1{font-size:1.75rem;font-weight:700;margin:0;letter-spacing:-0.03em;line-height:1.2}
.header p{font-size:0.875rem;color:var(--muted-foreground);margin:8px 0 0;font-weight:400}
</style></head><body><div class="viewport"><div class="canvas" style="`)
	fmt.Fprintf(&b, "width:%.0fpx;height:%.0fpx", cw, ch)
	b.WriteString(`">`)
	if d.Title != "" {
		b.WriteString(`<div class="header"><h1>` + xmlEsc(d.Title) + `</h1>`)
		if d.Subtitle != "" {
			b.WriteString(`<p>` + xmlEsc(d.Subtitle) + `</p>`)
		}
		b.WriteString(`</div>`)
	}
	for _, n := range d.Nodes {
		if paintsBeforeEdges(n.Kind) {
			b.WriteString(htmlNode(n, ox, oy))
		}
	}
	fmt.Fprintf(&b, `<svg class="edges" xmlns="http://www.w3.org/2000/svg" width="%.0f" height="%.0f" viewBox="%.1f %.1f %.1f %.1f">`, cw, ch, ox, oy, cw, ch)
	b.WriteString(`<defs>`)
	b.WriteString(SVGFontDefs())
	b.WriteString(SVGArrowMarkers(d))
	b.WriteString(`</defs>`)
	for _, re := range d.Routed {
		lctx := LabelContext(d, re.Edge)
		b.WriteString(polishedPath(re.Points, re.Edge, lctx))
	}
	b.WriteString(`</svg>`)
	for _, n := range d.Nodes {
		if !paintsBeforeEdges(n.Kind) {
			b.WriteString(htmlNode(n, ox, oy))
		}
	}
	b.WriteString(`<svg class="edge-labels" xmlns="http://www.w3.org/2000/svg" width="` + fmt.Sprintf("%.0f", cw) + `" height="` + fmt.Sprintf("%.0f", ch) + `" viewBox="` + fmt.Sprintf("%.1f %.1f %.1f %.1f", ox, oy, cw, ch) + `">`)
	for _, re := range d.Routed {
		if strings.TrimSpace(re.Edge.Label) == "" {
			continue
		}
		lctx := LabelContext(d, re.Edge)
		b.WriteString(EdgeLabelSVG(re.Points, re.Edge, lctx))
	}
	for _, re := range d.Routed {
		b.WriteString(ArrowHeadSVG(re.Points, re.Edge))
	}
	b.WriteString(`</svg></div></div></body></html>`)
	return b.String()
}

func htmlNode(n model.Node, ox, oy float64) string {
	cls := "node"
	k := model.NormalizeShape(n.Kind)
	switch k {
	case model.ShapeEllipse, model.ShapeCircle:
		cls += " ellipse"
	case model.ShapeActor:
		cls += " actor"
	case model.ShapeTextbox, model.ShapeNote:
		cls += " textbox"
	case model.ShapeInfobox, model.ShapeCallout:
		cls += " infobox"
	case model.ShapeDiamond, model.ShapeDecision:
		cls += " diamond"
	case model.ShapeLane, model.ShapeContainer:
		cls += " lane"
	case model.ShapeFrame, model.ShapeGroup:
		cls += " frame"
	case model.ShapePill, model.ShapeTerminal:
		cls += " pill"
	}
	fill := n.Fill
	if fill == "" {
		fill = paint.BgCard
	}
	accent := n.Accent
	if accent == "" {
		accent = paint.AccentBrand
	}
	stroke := n.Stroke
	if stroke == "" {
		stroke = paint.Border
	}
	var inner strings.Builder
	if n.Kind == model.ShapeLane && n.Label != "" {
		inner.WriteString(`<span class="lane-lbl">` + xmlEsc(n.Label) + `</span>`)
	} else {
		if n.Icon != "" {
			inner.WriteString(`<span class="ico">` + iconHTML(n.Icon) + `</span>`)
		}
		inner.WriteString(`<span class="title">` + xmlEsc(n.Label) + `</span>`)
		if n.Subtitle != "" {
			inner.WriteString(`<span class="sub">` + xmlEsc(n.Subtitle) + `</span>`)
		}
	}
	style := fmt.Sprintf("left:%.0fpx;top:%.0fpx;width:%.0fpx;min-height:%.0fpx;background:%s;border-color:%s;--accent:%s",
		n.Rect.X-ox, n.Rect.Y-oy, n.Rect.W, n.Rect.H, fill, stroke, accent)
	return fmt.Sprintf(`<div class="%s" style="%s">%s</div>`, cls, style, inner.String())
}

func iconHTML(name string) string {
	svg := icons.SVG(name, 0, 0, 22, "currentColor")
	return strings.Replace(svg, `x="0" y="0"`, `x="0" y="0" style="display:block"`, 1)
}
