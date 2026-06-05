package measure

import (
	"math"
	"strings"

	"github.com/niklas-heer/sceno/internal/fonts"
	"github.com/niklas-heer/sceno/internal/model"
)

// SnapUnit is the internal content grid inside shapes (icon + label bands).
const SnapUnit = 4.0

// Snap rounds v to the nearest SnapUnit (engine grid).
func Snap(v float64) float64 {
	return math.Round(v/SnapUnit) * SnapUnit
}

// ContentLayout is the measured interior grid for icon, title, and subtitle.
type ContentLayout struct {
	IconX, IconY float64
	IconSize     float64
	TitleX       float64
	TitleStartY  float64
	TitleLineH   float64
	TitleLines   int
	SubtitleX    float64
	SubtitleY    float64
	HasSubtitle  bool
	TopAlign     bool
	MinW         float64
	MinH         float64
}

// BuildContentLayout computes snapped interior placement and tight outer bounds.
func BuildContentLayout(n model.Node) ContentLayout {
	fs := n.FontSize
	if fs <= 0 {
		fs = 14
	}
	lines := strings.Split(n.Label, "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = nil
	}
	lineH := Snap(fs * 1.25)
	if lineH < fs {
		lineH = fs
	}

	pos := n.IconPos
	if pos == "" {
		pos = model.IconTopLeft
	}
	k := model.NormalizeShape(n.Kind)

	cl := ContentLayout{
		TitleLineH: lineH,
		TitleLines: len(lines),
		IconSize:   IconSize,
	}

	switch {
	case k == model.ShapeInfobox, k == model.ShapeCallout:
		cl.TopAlign = true
	case pos == model.IconTop, pos == model.IconTopRight:
		cl.TopAlign = true
	case n.Icon != "" && pos == model.IconTopLeft && k == model.ShapeActor:
		cl.TopAlign = true
	}

	padX := Snap(PadX * 0.85)
	padY := Snap(PadY * 0.8)
	if cl.TopAlign {
		padY = Snap(12)
	}

	iconBand := 0.0
	if n.Icon != "" {
		ix, iy := IconRect(n, IconSize)
		cl.IconX = Snap(ix - n.Rect.X)
		cl.IconY = Snap(iy - n.Rect.Y)
		if cl.TopAlign {
			iconBand = Snap(IconPad + IconSize + 6)
		}
	}

	titleBlockH := float64(len(lines)) * lineH
	subBlockH := 0.0
	if n.Subtitle != "" {
		cl.HasSubtitle = true
		subBlockH = Snap(subtitleH + 4)
	}

	innerH := padY + iconBand + titleBlockH
	if cl.HasSubtitle {
		innerH += subBlockH
	}
	if !cl.TopAlign {
		innerH += padY
	} else {
		innerH += Snap(8)
	}
	cl.MinH = Snap(math.Max(innerH, 40))

	maxLineW := 0.0
	for _, line := range lines {
		tw := TextWidth(line, fs, fonts.WeightMedium)
		if tw > maxLineW {
			maxLineW = tw
		}
	}
	contentW := maxLineW
	if n.Icon != "" && !cl.TopAlign {
		contentW += IconColumn
	}
	if n.Subtitle != "" {
		sw := TextWidth(n.Subtitle, fs*0.85, fonts.WeightRegular)
		if sw > contentW {
			contentW = sw
		}
	}
	cl.MinW = Snap(math.Max(contentW+padX, 72))

	cl.TitleX = padX
	if n.Icon != "" && !cl.TopAlign && pos == model.IconTopLeft {
		cl.TitleX = Snap(IconColumn)
	}
	cl.TitleStartY = padY + iconBand + lineH*0.75
	if !cl.TopAlign {
		cl.TitleStartY = (cl.MinH-titleBlockH-iconBand)/2 + lineH*0.75
		if cl.HasSubtitle {
			cl.TitleStartY -= subBlockH / 2
		}
	}
	cl.TitleStartY = Snap(cl.TitleStartY)
	cl.SubtitleX = padX
	cl.SubtitleY = Snap(cl.MinH - 14)
	if cl.TopAlign && cl.HasSubtitle {
		cl.SubtitleY = Snap(cl.TitleStartY + float64(len(lines))*lineH + 4)
	}

	return cl
}

// LayoutFor returns the interior grid for a node (stored layout from pipeline when ready).
func LayoutFor(n model.Node) ContentLayout {
	if n.Interior.Ready {
		return contentFromInterior(n)
	}
	return BuildContentLayout(n)
}

func contentFromInterior(n model.Node) ContentLayout {
	in := n.Interior
	return ContentLayout{
		IconX: in.IconX, IconY: in.IconY, IconSize: in.IconSize,
		TitleX: in.TitleX, TitleStartY: in.TitleStartY, TitleLineH: in.TitleLineH,
		TitleLines: in.TitleLines, SubtitleX: in.SubtitleX, SubtitleY: in.SubtitleY,
		HasSubtitle: in.HasSubtitle, TopAlign: in.TopAlign, MinW: in.MinW, MinH: in.MinH,
	}
}

// ApplyInteriors computes and stores interior layout on every node (pipeline/engine SoT).
func ApplyInteriors(nodes []model.Node) {
	for i := range nodes {
		if model.IsContainer(nodes[i].Kind) || model.NormalizeShape(nodes[i].Kind) == model.ShapeCode {
			continue
		}
		cl := BuildContentLayout(nodes[i])
		nodes[i].Interior = interiorToModel(cl)
	}
}

func interiorToModel(cl ContentLayout) model.InteriorLayout {
	return model.InteriorLayout{
		IconX: cl.IconX, IconY: cl.IconY, IconSize: cl.IconSize,
		TitleX: cl.TitleX, TitleStartY: cl.TitleStartY, TitleLineH: cl.TitleLineH,
		TitleLines: cl.TitleLines, SubtitleX: cl.SubtitleX, SubtitleY: cl.SubtitleY,
		HasSubtitle: cl.HasSubtitle, TopAlign: cl.TopAlign, MinW: cl.MinW, MinH: cl.MinH,
		Ready: true,
	}
}

// LabelLayoutFor returns label region metadata (delegates to content grid).
func LabelLayoutFor(n model.Node) LabelLayout {
	cl := LayoutFor(n)
	contentW := n.Rect.W - PadX
	if n.Icon != "" && !cl.TopAlign {
		contentW -= IconColumn
	}
	return LabelLayout{
		ContentX:    n.Rect.X + cl.TitleX,
		ContentY:    n.Rect.Y,
		ContentW:    contentW,
		TopAlign:    cl.TopAlign,
		IconOffsetY: cl.TitleStartY,
	}
}
