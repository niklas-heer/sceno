package measure

import (
	"math"
	"strings"

	"github.com/niklas-heer/sceno/internal/fonts"
	"github.com/niklas-heer/sceno/internal/model"
)

const (
	PadX        = 28.0
	PadY        = 20.0
	IconColumn  = 36.0
	IconSize    = 20.0
	lineSpacing = 1.32
	subtitleH   = 15.0
)

// TextWidth returns pixel width of s using embedded Inter.
func TextWidth(s string, size float64, weight fonts.Weight) float64 {
	return fonts.TextWidth(s, size, weight)
}

// ContentSize returns the inner size needed for label, subtitle, and icon.
func ContentSize(n model.NodeSpec) (w, h float64) {
	fs := n.FontSize
	if fs <= 0 {
		fs = 14
	}
	lines := strings.Split(n.Label, "\n")
	maxW := 0.0
	for _, line := range lines {
		tw := TextWidth(line, fs, fonts.WeightMedium)
		if tw > maxW {
			maxW = tw
		}
	}
	lineH := fs * lineSpacing
	h = float64(len(lines))*lineH + PadY
	if len(lines) == 0 {
		h = PadY
	}
	w = maxW + PadX
	if n.Icon != "" {
		w += IconColumn
	}
	if n.Subtitle != "" {
		sw := TextWidth(n.Subtitle, fs*0.85, fonts.WeightRegular)
		if sw+PadX > w {
			w = sw + PadX
		}
		h += subtitleH
	}
	k := model.NormalizeShape(n.Kind)
	applyShapePadding(&w, &h, k)
	return w, h
}

func applyShapePadding(w, h *float64, k model.ShapeKind) {
	switch k {
	case model.ShapeActor:
		if *w < 88 {
			*w = 88
		}
		if *h < 80 {
			*h = 80
		}
	case model.ShapeEllipse, model.ShapeCircle, model.ShapeCloud:
		if *w < 120 {
			*w = 120
		}
		if *h < 48 {
			*h = 48
		}
	case model.ShapeDiamond, model.ShapeDecision, model.ShapeHexagon, model.ShapeOctagon:
		*w *= 1.18
		*h *= 1.18
	case model.ShapeCylinder, model.ShapeDatabase:
		*h += 14
	case model.ShapeTriangle:
		*h *= 1.12
	case model.ShapeParallelogram:
		*w *= 1.12
	case model.ShapePill, model.ShapeTerminal, model.ShapeStart, model.ShapeEnd:
		if *h < 40 {
			*h = 40
		}
	case model.ShapeLane, model.ShapeFrame, model.ShapeContainer, model.ShapeGroup:
		*w += 36
		*h += 32
	case model.ShapeInfobox, model.ShapeCallout:
		*w += 8
	}
	if *w < 72 {
		*w = 72
	}
	if *h < 40 {
		*h = 40
	}
	*w = Snap(*w)
	*h = Snap(*h)
}

const (
	codeFontSize = 11.0
	codeLineH    = 15.0
	codePadX     = 24.0
	codePadY     = 20.0
)

// CodeContentSize measures a syntax-highlighted code block.
func CodeContentSize(n model.NodeSpec) (w, h float64) {
	body := n.Code
	if body == "" {
		body = n.Label
	}
	lines := strings.Split(body, "\n")
	if len(lines) == 0 {
		lines = []string{" "}
	}
	maxW := 0.0
	for _, line := range lines {
		tw := TextWidth(line, codeFontSize, fonts.WeightRegular)
		if tw > maxW {
			maxW = tw
		}
	}
	h = float64(len(lines))*codeLineH + codePadY
	w = maxW + codePadX
	if n.Label != "" && n.Label != body {
		h += 18
	}
	if w < 200 {
		w = 200
	}
	if h < 48 {
		h = 48
	}
	return w, h
}

// FitSize returns node dimensions: grows to fit text; explicit w/h are minimums.
func FitSize(n model.NodeSpec) (w, h float64) {
	if model.NormalizeShape(n.Kind) == model.ShapeCode {
		needW, needH := CodeContentSize(n)
		w, h = needW, needH
		if n.W > 0 && n.W > w {
			w = n.W
		}
		if n.H > 0 && n.H > h {
			h = n.H
		}
		return w, h
	}
	// Explicit two-dimensional bounds are a strong compositional preference.
	// Keep them when the label can fit by selecting a smaller readable font;
	// only grow the shape when even the minimum font cannot clear its silhouette.
	if n.W > 0 && n.H > 0 {
		placed := nodeFromSpec(n)
		placed.Rect = model.Rect{W: n.W, H: n.H}
		preferred := n.FontSize
		if preferred <= 0 {
			preferred = 14
		}
		fontSize := fittedFontSize(placed, preferred)
		req := measureContentRequirements(placed, fontSize)
		writable := ShapeWritableRect(placed, req.requiredW/req.requiredH)
		if req.requiredW <= writable.W+.5 && req.requiredH <= writable.H+.5 {
			return Snap(n.W), Snap(n.H)
		}
	}
	cl := BuildContentLayout(nodeFromSpec(n))
	w, h = cl.MinW, cl.MinH
	if n.W > 0 && n.W > w {
		w = n.W
	}
	if n.H > 0 && n.H > h {
		h = n.H
	}
	return Snap(w), Snap(h)
}

func nodeFromSpec(ns model.NodeSpec) model.Node {
	return model.Node{
		Label: ns.Label, Subtitle: ns.Subtitle, Kind: ns.Kind,
		Icon: ns.Icon, IconPos: ns.IconPos, FontSize: ns.FontSize,
		Rect: model.Rect{W: 200, H: 120},
	}
}

// Overflow returns how many pixels label content exceeds the node rect (0 = fits).
func Overflow(n model.Node) (overW, overH float64) {
	cl := LayoutFor(n)
	required := measureContentRequirements(n, cl.FontSize)
	overW = math.Max(0, required.requiredW-cl.WritableW)
	overH = math.Max(0, required.requiredH-cl.WritableH)
	content := textBounds(n, cl)
	safe := model.Rect{X: n.Rect.X + cl.WritableX, Y: n.Rect.Y + cl.WritableY, W: cl.WritableW, H: cl.WritableH}
	if content.W > 0 {
		overW = math.Max(overW, math.Max(safe.X-content.X, content.Right()-safe.Right()))
		overH = math.Max(overH, math.Max(safe.Y-content.Y, content.Bottom()-safe.Bottom()))
	}
	return overW, overH
}

func textBounds(n model.Node, cl ContentLayout) model.Rect {
	fs := cl.FontSize
	if fs <= 0 {
		fs = 14
	}
	var bounds model.Rect
	add := func(r model.Rect) {
		if r.W <= 0 || r.H <= 0 {
			return
		}
		if bounds.W == 0 || bounds.H == 0 {
			bounds = r
			return
		}
		left, top := math.Min(bounds.X, r.X), math.Min(bounds.Y, r.Y)
		right, bottom := math.Max(bounds.Right(), r.Right()), math.Max(bounds.Bottom(), r.Bottom())
		bounds = model.Rect{X: left, Y: top, W: right - left, H: bottom - top}
	}
	for i, line := range strings.Split(n.Label, "\n") {
		if line == "" {
			continue
		}
		w := TextWidth(line, fs, fonts.WeightMedium)
		x := n.Rect.X + cl.WritableX + (cl.WritableW-w)/2
		if cl.InlineIcon {
			x = n.Rect.X + cl.TitleX
		}
		baseline := n.Rect.Y + cl.TitleStartY + float64(i)*cl.TitleLineH
		add(model.Rect{X: x, Y: baseline - fs, W: w, H: cl.TitleLineH})
	}
	if cl.HasSubtitle && n.Subtitle != "" {
		subSize := cl.SubtitleSize
		if subSize <= 0 {
			subSize = fs * .85
		}
		w := TextWidth(n.Subtitle, subSize, fonts.WeightRegular)
		x := n.Rect.X + cl.WritableX + (cl.WritableW-w)/2
		if cl.InlineIcon {
			x = n.Rect.X + cl.TitleX
		}
		baseline := n.Rect.Y + cl.SubtitleY
		add(model.Rect{X: x, Y: baseline - subSize, W: w, H: subSize * 1.25})
	}
	return bounds
}

// EnsureNodeFits expands node rect to fit measured content.
func EnsureNodeFits(n *model.Node) {
	ow, oh := Overflow(*n)
	if ow > 0 {
		n.Rect.W += ow
	}
	if oh > 0 {
		n.Rect.H += oh
	}
}
