package measure

import (
	"strings"

	"github.com/niklas-heer/sceno/internal/fonts"
	"github.com/niklas-heer/sceno/internal/model"

	"golang.org/x/image/font"
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
	if s == "" {
		return 0
	}
	face, err := fonts.Face(weight, size)
	if err != nil {
		return float64(len(s)) * size * 0.55
	}
	return measureString(face, s)
}

func measureString(face font.Face, s string) float64 {
	d := &font.Drawer{Face: face}
	return float64(d.MeasureString(s).Ceil())
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
	needW, needH := ContentSize(n)
	w, h = needW, needH
	if n.W > 0 && n.W > w {
		w = n.W
	}
	if n.H > 0 && n.H > h {
		h = n.H
	}
	return w, h
}

// Overflow returns how many pixels label content exceeds the node rect (0 = fits).
func Overflow(n model.Node) (overW, overH float64) {
	ns := model.NodeSpec{
		Label:    n.Label,
		Subtitle: n.Subtitle,
		Kind:     n.Kind,
		Icon:     n.Icon,
		FontSize: n.FontSize,
	}
	needW, needH := ContentSize(ns)
	if n.Rect.W < needW {
		overW = needW - n.Rect.W
	}
	if n.Rect.H < needH {
		overH = needH - n.Rect.H
	}
	return overW, overH
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
