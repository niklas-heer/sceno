package render

import (
	"fmt"
	"strings"

	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/theme"
)

// SVGOptions controls polished SVG output.
type SVGOptions struct {
	// DropShadow enables feDropShadow (unsupported by some rasterizers).
	DropShadow bool
}

// PolishedSVG renders export-ready SVG.
func PolishedSVG(d model.Diagram) string {
	return PolishedSVGWithOptions(d, SVGOptions{DropShadow: true})
}

// PolishedSVGWithOptions renders with explicit options.
func PolishedSVGWithOptions(d model.Diagram, opt SVGOptions) string {
	useDiagramPalette(d)
	vp := ViewportFrom(d)
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="%.1f %.1f %.1f %.1f" width="%.0f" height="%.0f">`,
		vp.MinX, vp.MinY, vp.Width, vp.Height, vp.Width, vp.Height)
	b.WriteString(polishedSVGDefs(d, opt))
	b.WriteString(polishedSVGCanvas(vp))
	b.WriteString(polishedSVGContent(d, vp, opt))
	b.WriteString(`</svg>`)
	return b.String()
}

// PolishedSVGSlide renders a 16:9 (or 4:3) slide with the diagram scaled to fit.
func PolishedSVGSlide(d model.Diagram) string {
	useDiagramPalette(d)
	aspect := d.SlideAspect
	if aspect == "" {
		aspect = Aspect16x9
	}
	frame := SlideFrameFrom(d, aspect)
	opt := SVGOptions{DropShadow: true}
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %.0f %.0f" width="%.0f" height="%.0f">`,
		frame.Width, frame.Height, frame.Width, frame.Height)
	b.WriteString(polishedSVGDefs(d, opt))
	bg := paint.BgCanvas
	if paint.Transparent {
		bg = "none"
	}
	fmt.Fprintf(&b, `<rect width="%.0f" height="%.0f" fill="%s" rx="%d"/>`, frame.Width, frame.Height, bg, theme.RadiusSlide)
	fmt.Fprintf(&b, `<g transform="translate(%.2f %.2f) scale(%.4f)">`, frame.OffsetX, frame.OffsetY, frame.Scale)
	vp := frame.Content
	b.WriteString(polishedSVGContent(d, vp, opt))
	b.WriteString(`</g></svg>`)
	return b.String()
}

func polishedSVGDefs(d model.Diagram, opt SVGOptions) string {
	var b strings.Builder
	b.WriteString(`<defs>`)
	b.WriteString(SVGFontDefs())
	if opt.DropShadow {
		b.WriteString(`<filter id="shadow" x="-15%" y="-15%" width="130%" height="130%"><feDropShadow dx="0" dy="1" stdDeviation="2" flood-color="` + paint.Shadow + `" flood-opacity="0.06"/></filter>`)
	}
	b.WriteString(SVGArrowMarkers(d))
	b.WriteString(`</defs>`)
	return b.String()
}

func polishedSVGCanvas(vp Viewport) string {
	bg := paint.BgCanvas
	if paint.Transparent {
		bg = "none"
	}
	return fmt.Sprintf(`<rect width="%.1f" height="%.1f" x="%.1f" y="%.1f" fill="%s" rx="%d"/>`,
		vp.Width, vp.Height, vp.MinX, vp.MinY, bg, theme.RadiusSlide)
}

func polishedSVGContent(d model.Diagram, vp Viewport, opt SVGOptions) string {
	var b strings.Builder
	if d.Title != "" {
		b.WriteString(textEl(d.Title, vp.MinX+32, vp.MinY+36, theme.TitleSize, paint.FgPrimary, "700"))
	}
	if d.Subtitle != "" {
		b.WriteString(textEl(d.Subtitle, vp.MinX+32, vp.MinY+62, theme.SubtitleSize, paint.FgMuted, ""))
	}
	for _, n := range d.Nodes {
		if paintsBeforeEdges(n.Kind) {
			b.WriteString(polishedNodeSVG(n, opt.DropShadow))
		}
	}
	for _, re := range d.Routed {
		lctx := LabelContext(d, re.Edge)
		b.WriteString(polishedPath(re.Points, re.Edge, lctx))
	}
	for _, n := range d.Nodes {
		if !paintsBeforeEdges(n.Kind) {
			b.WriteString(polishedNodeSVG(n, opt.DropShadow))
		}
	}
	for _, re := range d.Routed {
		if strings.TrimSpace(re.Edge.Label) == "" {
			continue
		}
		lctx := LabelContext(d, re.Edge)
		b.WriteString(EdgeLabelSVG(re.Points, re.Edge, lctx))
	}
	// Arrowheads after nodes so tips meet borders visibly (not buried under fills).
	for _, re := range d.Routed {
		b.WriteString(ArrowHeadSVG(re.Points, re.Edge))
	}
	return b.String()
}
