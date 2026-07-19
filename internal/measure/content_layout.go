package measure

import (
	"math"
	"strings"

	"github.com/niklas-heer/sceno/internal/fonts"
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/theme"
)

// SnapUnit is the internal content grid inside shapes (icon + label bands).
const SnapUnit = 4.0

const InlineIconGap = 12.0

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
	InlineIcon   bool
	MinW         float64
	MinH         float64
}

// ContainerLabelBounds is the chrome-plane title box painted inside a lane or
// container. Label placement and scene inspection consume this same geometry.
func ContainerLabelBounds(n model.Node) model.Rect {
	if !model.IsContainer(n.Kind) || n.Label == "" {
		return model.Rect{}
	}
	return model.Rect{
		X: n.Rect.X + 14, Y: n.Rect.Y + 2,
		W: TextWidth(n.Label, theme.LaneLabelSize, fonts.WeightSemiBold), H: 16,
	}
}

// EffectiveIconPos returns the icon position used by layout (default: compact inline group).
func EffectiveIconPos(n model.Node) model.IconPosition {
	if n.IconPos != "" {
		return n.IconPos
	}
	return model.IconTopLeft
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

	pos := EffectiveIconPos(n)
	k := model.NormalizeShape(n.Kind)

	cl := ContentLayout{
		TitleLineH: lineH,
		TitleLines: len(lines),
		IconSize:   IconSize,
	}

	hasIcon := n.Icon != ""
	cl.InlineIcon = hasIcon && pos == model.IconTopLeft
	topIcon := hasIcon && (pos == model.IconTop || pos == model.IconTopRight)
	bottomIcon := hasIcon && (pos == model.IconBottomLeft || pos == model.IconBottom || pos == model.IconBottomRight)

	switch {
	case cl.InlineIcon:
		cl.TopAlign = false
	case k == model.ShapeInfobox, k == model.ShapeCallout, k == model.ShapeNote:
		cl.TopAlign = true
	case k == model.ShapeActor:
		cl.TopAlign = true
	case topIcon:
		cl.TopAlign = true
	}

	padX := Snap(16)
	padY := Snap(12)
	if cl.InlineIcon {
		padX = Snap(PadX * 0.85)
		padY = Snap(PadY * 0.8)
	}

	topBand := 0.0
	bottomBand := 0.0
	if topIcon {
		topBand = Snap(IconPad + IconSize + 8)
	}
	if bottomIcon {
		bottomBand = Snap(IconPad + IconSize + 8)
	}

	titleBlockH := float64(len(lines)) * lineH
	subBlockH := 0.0
	if n.Subtitle != "" {
		cl.HasSubtitle = true
		subBlockH = Snap(subtitleH + 4)
	}

	innerH := padY + topBand + titleBlockH
	if cl.HasSubtitle {
		innerH += subBlockH
	}
	if cl.InlineIcon {
		innerH = math.Max(innerH, IconSize+padY*2)
	} else {
		innerH += bottomBand + padY
	}
	cl.MinH = Snap(math.Max(innerH, shapeMinH(k)))

	maxLineW := 0.0
	for _, line := range lines {
		tw := TextWidth(line, fs, fonts.WeightMedium)
		if tw > maxLineW {
			maxLineW = tw
		}
	}
	textBlockW := maxLineW
	if n.Subtitle != "" {
		sw := TextWidth(n.Subtitle, fs*0.85, fonts.WeightRegular)
		textBlockW = math.Max(textBlockW, sw)
	}
	contentW := textBlockW
	if cl.InlineIcon {
		contentW += IconSize + InlineIconGap
	}
	cl.MinW = Snap(math.Max(contentW+padX*2, shapeMinW(k)))
	cl.MinW, cl.MinH = expandForSilhouette(k, cl.MinW, cl.MinH)
	if hasIcon {
		cl.MinW = math.Max(cl.MinW, Snap(IconSize+IconPad*2))
	}

	groupH := titleBlockH + subBlockH
	if cl.InlineIcon {
		cl.TitleStartY = (cl.MinH-groupH)/2 + lineH*0.75
	} else if cl.TopAlign {
		cl.TitleX = padX
		cl.TitleStartY = padY + topBand + lineH*0.75
	} else {
		cl.TitleX = padX
		availableH := cl.MinH - bottomBand
		cl.TitleStartY = (availableH-groupH)/2 + lineH*0.75
	}
	cl.TitleStartY = Snap(cl.TitleStartY)
	cl.SubtitleX = padX
	if cl.HasSubtitle {
		cl.SubtitleY = Snap(cl.TitleStartY + float64(len(lines))*lineH + 6)
	}

	if hasIcon {
		layoutW := math.Max(n.Rect.W, cl.MinW)
		layoutH := math.Max(n.Rect.H, cl.MinH)
		if cl.InlineIcon {
			cl.IconX = Snap((layoutW - contentW) / 2)
			cl.IconY = Snap((layoutH - IconSize) / 2)
			cl.TitleX = cl.IconX + IconSize + InlineIconGap
		} else {
			cl.IconX, cl.IconY = iconOffset(pos, layoutW, layoutH)
		}
	}

	return cl
}

func expandForSilhouette(k model.ShapeKind, w, h float64) (float64, float64) {
	switch model.NormalizeShape(k) {
	case model.ShapeCloud:
		w /= .80
		h /= .64
	case model.ShapeCylinder, model.ShapeDatabase:
		w += 16
		rim := math.Min(math.Min(w*.10, h*.15), 12)
		h += 2 * (rim + 4)
	case model.ShapeHexagon, model.ShapeOctagon:
		w /= .72
		h /= .84
	case model.ShapeDiamond, model.ShapeDecision:
		w /= .48
		h /= .56
	}
	return Snap(w), Snap(h)
}

func iconOffset(pos model.IconPosition, w, h float64) (x, y float64) {
	switch pos {
	case model.IconTop:
		x, y = (w-IconSize)/2, IconPad
	case model.IconTopRight:
		x, y = w-IconPad-IconSize, IconPad
	case model.IconCenter:
		x, y = (w-IconSize)/2, (h-IconSize)/2
	case model.IconBottomLeft:
		x, y = IconPad, h-IconPad-IconSize
	case model.IconBottom:
		x, y = (w-IconSize)/2, h-IconPad-IconSize
	case model.IconBottomRight:
		x, y = w-IconPad-IconSize, h-IconPad-IconSize
	default:
		x, y = IconPad, (h-IconSize)/2
	}
	return Snap(x), Snap(y)
}

func shapeMinH(k model.ShapeKind) float64 {
	switch k {
	case model.ShapeActor:
		return 72
	case model.ShapeDiamond, model.ShapeDecision:
		return 56
	default:
		return 44
	}
}

func shapeMinW(k model.ShapeKind) float64 {
	switch k {
	case model.ShapeActor:
		return 72
	case model.ShapeInfobox, model.ShapeCallout:
		return 120
	default:
		return 80
	}
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
		HasSubtitle: in.HasSubtitle, TopAlign: in.TopAlign, InlineIcon: in.InlineIcon,
		MinW: in.MinW, MinH: in.MinH,
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
		HasSubtitle: cl.HasSubtitle, TopAlign: cl.TopAlign, InlineIcon: cl.InlineIcon,
		MinW: cl.MinW, MinH: cl.MinH,
		Ready: true,
	}
}

// TightenToInterior shrinks node rects toward measured content bounds (keeps center).
func TightenToInterior(nodes []model.Node) {
	const slack = 20.0
	for i := range nodes {
		n := &nodes[i]
		if model.IsContainer(n.Kind) || n.Fixed {
			continue
		}
		cl := LayoutFor(*n)
		if cl.MinW <= 0 || cl.MinH <= 0 {
			continue
		}
		targetW := math.Max(cl.MinW, n.MinW)
		targetH := math.Max(cl.MinH, n.MinH)
		// Keep grid column widths — shrinking W without reflowing columns causes overlaps.
		if n.Column < 0 && n.Rect.W > targetW+slack {
			dx := n.Rect.W - targetW
			n.Rect.X += dx / 2
			n.Rect.W = targetW
		}
		if n.Rect.H > targetH+slack {
			n.Rect.H = targetH
		}
	}
}

// LabelLayoutFor returns label region metadata (delegates to content grid).
func LabelLayoutFor(n model.Node) LabelLayout {
	cl := LayoutFor(n)
	contentW := n.Rect.W - PadX
	if cl.InlineIcon {
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
