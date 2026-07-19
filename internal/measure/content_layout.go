package measure

import (
	"math"
	"strings"

	"github.com/niklas-heer/sceno/internal/fonts"
	"github.com/niklas-heer/sceno/internal/geom"
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
	FontSize     float64
	SubtitleSize float64
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
	WritableX    float64
	WritableY    float64
	WritableW    float64
	WritableH    float64
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

// minAutoFontSize prevents "fitting" content by making it illegibly small.
const minAutoFontSize = 10.0

type contentRequirements struct {
	fontSize, subtitleSize, lineH float64
	lines                         []string
	padX, padY                    float64
	topBand, bottomBand           float64
	titleBlockH, subBlockH        float64
	textBlockW, groupW            float64
	requiredW, requiredH          float64
	inline, topIcon, bottomIcon   bool
}

// BuildContentLayout computes preferred-size geometry for auto sizing.
func BuildContentLayout(n model.Node) ContentLayout { return buildContentLayout(n, false) }

// BuildPlacedContentLayout chooses the largest font that fits an already
// placed shape, then positions content inside its silhouette-derived region.
func BuildPlacedContentLayout(n model.Node) ContentLayout { return buildContentLayout(n, true) }

func buildContentLayout(n model.Node, fitPlaced bool) ContentLayout {
	preferred := n.FontSize
	if preferred <= 0 {
		preferred = 14
	}
	fontSize := preferred
	if fitPlaced && n.Rect.W > 0 && n.Rect.H > 0 {
		fontSize = fittedFontSize(n, preferred)
	}
	req := measureContentRequirements(n, fontSize)
	preferredReq := measureContentRequirements(n, preferred)
	k := model.NormalizeShape(n.Kind)
	minW, minH := fitShapeForContent(n, preferredReq.requiredW, preferredReq.requiredH)

	layoutW, layoutH := math.Max(n.Rect.W, minW), math.Max(n.Rect.H, minH)
	if fitPlaced && n.Rect.W > 0 && n.Rect.H > 0 {
		layoutW, layoutH = n.Rect.W, n.Rect.H
	}
	probe := n
	probe.Rect = model.Rect{W: layoutW, H: layoutH}
	writable := ShapeWritableRect(probe, req.requiredW/math.Max(req.requiredH, 1))
	if writable.W <= 0 || writable.H <= 0 {
		writable = model.Rect{W: layoutW, H: layoutH}
	}

	cl := ContentLayout{
		IconSize: IconSize, FontSize: fontSize, SubtitleSize: req.subtitleSize,
		TitleLineH: req.lineH, TitleLines: len(req.lines), HasSubtitle: n.Subtitle != "",
		InlineIcon: req.inline, MinW: minW, MinH: minH,
		WritableX: writable.X, WritableY: writable.Y, WritableW: writable.W, WritableH: writable.H,
	}
	cl.TopAlign = !req.inline && (k == model.ShapeInfobox || k == model.ShapeNote || k == model.ShapeActor || req.topIcon)

	textGroupH := req.titleBlockH + req.subBlockH
	if req.inline {
		groupH := math.Max(IconSize, textGroupH)
		groupLeft := writable.X + (writable.W-req.groupW)/2
		groupTop := writable.Y + (writable.H-groupH)/2
		cl.IconX = math.Max(writable.X, math.Min(Snap(groupLeft), writable.Right()-req.groupW))
		cl.IconY = Snap(groupTop + (groupH-IconSize)/2)
		cl.TitleX = cl.IconX + IconSize + InlineIconGap
		cl.TitleStartY = groupTop + (groupH-textGroupH)/2 + req.lineH*.75
	} else {
		cl.TitleX = writable.X + req.padX
		if cl.TopAlign {
			cl.TitleStartY = writable.Y + req.padY + req.topBand + req.lineH*.75
		} else {
			availableH := writable.H - req.topBand - req.bottomBand
			cl.TitleStartY = writable.Y + req.topBand + (availableH-textGroupH)/2 + req.lineH*.75
		}
	}
	cl.TitleStartY = Snap(cl.TitleStartY)
	cl.SubtitleX = cl.TitleX
	if cl.HasSubtitle {
		cl.SubtitleY = Snap(cl.TitleStartY + float64(len(req.lines))*req.lineH + 6)
	}
	clampTextVertically(&cl, req, writable)
	if n.Icon != "" && !req.inline {
		cl.IconX, cl.IconY = iconOffsetInRect(EffectiveIconPos(n), writable)
	}
	return cl
}

func clampTextVertically(cl *ContentLayout, req contentRequirements, writable model.Rect) {
	if len(req.lines) == 0 && !cl.HasSubtitle {
		return
	}
	top, bottom := math.Inf(1), math.Inf(-1)
	if len(req.lines) > 0 {
		top = cl.TitleStartY - req.fontSize
		lastBaseline := cl.TitleStartY + float64(len(req.lines)-1)*req.lineH
		bottom = lastBaseline - req.fontSize + req.lineH
	}
	if cl.HasSubtitle {
		subTop := cl.SubtitleY - req.subtitleSize
		subBottom := subTop + req.subtitleSize*1.25
		top, bottom = math.Min(top, subTop), math.Max(bottom, subBottom)
	}
	shift := 0.0
	if top < writable.Y {
		shift = writable.Y - top
	}
	if bottom+shift > writable.Bottom() {
		shift -= bottom + shift - writable.Bottom()
	}
	cl.TitleStartY += shift
	if cl.HasSubtitle {
		cl.SubtitleY += shift
	}
}

func measureContentRequirements(n model.Node, fs float64) contentRequirements {
	lines := strings.Split(n.Label, "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = nil
	}
	lineH := Snap(fs * 1.25)
	if lineH < fs {
		lineH = fs
	}
	pos := EffectiveIconPos(n)
	hasIcon := n.Icon != ""
	req := contentRequirements{
		fontSize: fs, subtitleSize: fs * .85, lineH: lineH, lines: lines,
		padX:       Snap(math.Max(4, 16-geom.ShapeBorderClearance-geom.ShapeStrokeWidth(n.Kind)/2)),
		padY:       Snap(math.Max(4, 12-geom.ShapeBorderClearance-geom.ShapeStrokeWidth(n.Kind)/2)),
		inline:     hasIcon && pos == model.IconTopLeft,
		topIcon:    hasIcon && (pos == model.IconTop || pos == model.IconTopRight),
		bottomIcon: hasIcon && (pos == model.IconBottomLeft || pos == model.IconBottom || pos == model.IconBottomRight),
	}
	if req.inline {
		clearance := geom.ShapeBorderClearance + geom.ShapeStrokeWidth(n.Kind)/2
		req.padX = Snap(math.Max(4, PadX*.85-clearance))
		req.padY = Snap(math.Max(4, PadY*.8-clearance))
	}
	if req.topIcon {
		req.topBand = Snap(IconPad + IconSize + 8)
	}
	if req.bottomIcon {
		req.bottomBand = Snap(IconPad + IconSize + 8)
	}
	req.titleBlockH = float64(len(lines)) * lineH
	if n.Subtitle != "" {
		req.subBlockH = Snap(req.subtitleSize + 4)
	}
	for _, line := range lines {
		req.textBlockW = math.Max(req.textBlockW, TextWidth(line, fs, fonts.WeightMedium))
	}
	if n.Subtitle != "" {
		req.textBlockW = math.Max(req.textBlockW, TextWidth(n.Subtitle, req.subtitleSize, fonts.WeightRegular))
	}
	req.groupW = req.textBlockW
	if req.inline {
		req.groupW += IconSize + InlineIconGap
	}
	req.requiredW = req.groupW + req.padX*2
	req.requiredH = req.padY*2 + req.topBand + req.titleBlockH + req.subBlockH + req.bottomBand
	if req.inline {
		req.requiredH = math.Max(req.requiredH, IconSize+req.padY*2)
	}
	req.requiredW = math.Max(req.requiredW, 1)
	req.requiredH = math.Max(req.requiredH, 1)
	return req
}

func fittedFontSize(n model.Node, preferred float64) float64 {
	for fs := preferred; fs >= minAutoFontSize; fs -= .5 {
		req := measureContentRequirements(n, fs)
		writable := ShapeWritableRect(n, req.requiredW/req.requiredH)
		if req.requiredW <= writable.W+.5 && req.requiredH <= writable.H+.5 {
			return fs
		}
	}
	return math.Min(preferred, minAutoFontSize)
}

func fitShapeForContent(n model.Node, needW, needH float64) (float64, float64) {
	k := model.NormalizeShape(n.Kind)
	w, h := math.Max(needW, shapeMinW(k)), math.Max(needH, shapeMinH(k))
	aspect := needW / math.Max(needH, 1)
	for range 24 {
		probe := n
		probe.Rect = model.Rect{W: w, H: h}
		writable := ShapeWritableRect(probe, aspect)
		if writable.W > 0 && writable.H > 0 {
			scale := math.Max(needW/writable.W, needH/writable.H)
			if scale <= 1.001 {
				return SnapUp(w), SnapUp(h)
			}
			w, h = w*scale+2, h*scale+2
			continue
		}
		w, h = w+SnapUnit*2, h+SnapUnit*2
	}
	return SnapUp(w), SnapUp(h)
}

func SnapUp(v float64) float64 { return math.Ceil(v/SnapUnit) * SnapUnit }

func iconOffsetInRect(pos model.IconPosition, r model.Rect) (x, y float64) {
	switch pos {
	case model.IconTop:
		x, y = r.X+(r.W-IconSize)/2, r.Y+IconPad
	case model.IconTopRight:
		x, y = r.Right()-IconPad-IconSize, r.Y+IconPad
	case model.IconCenter:
		x, y = r.X+(r.W-IconSize)/2, r.Y+(r.H-IconSize)/2
	case model.IconBottomLeft:
		x, y = r.X+IconPad, r.Bottom()-IconPad-IconSize
	case model.IconBottom:
		x, y = r.X+(r.W-IconSize)/2, r.Bottom()-IconPad-IconSize
	case model.IconBottomRight:
		x, y = r.Right()-IconPad-IconSize, r.Bottom()-IconPad-IconSize
	default:
		x, y = r.X+IconPad, r.Y+(r.H-IconSize)/2
	}
	x, y = Snap(x), Snap(y)
	return math.Max(r.X, math.Min(x, r.Right()-IconSize)), math.Max(r.Y, math.Min(y, r.Bottom()-IconSize))
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
	return BuildPlacedContentLayout(n)
}

func contentFromInterior(n model.Node) ContentLayout {
	in := n.Interior
	return ContentLayout{
		IconX: in.IconX, IconY: in.IconY, IconSize: in.IconSize, FontSize: in.FontSize, SubtitleSize: in.SubtitleSize,
		TitleX: in.TitleX, TitleStartY: in.TitleStartY, TitleLineH: in.TitleLineH,
		TitleLines: in.TitleLines, SubtitleX: in.SubtitleX, SubtitleY: in.SubtitleY,
		HasSubtitle: in.HasSubtitle, TopAlign: in.TopAlign, InlineIcon: in.InlineIcon,
		MinW: in.MinW, MinH: in.MinH,
		WritableX: in.WritableX, WritableY: in.WritableY, WritableW: in.WritableW, WritableH: in.WritableH,
	}
}

// ApplyInteriors computes and stores interior layout on every node (pipeline/engine SoT).
func ApplyInteriors(nodes []model.Node) {
	for i := range nodes {
		if model.IsContainer(nodes[i].Kind) || model.NormalizeShape(nodes[i].Kind) == model.ShapeCode {
			continue
		}
		cl := BuildPlacedContentLayout(nodes[i])
		nodes[i].Interior = interiorToModel(cl)
	}
}

func interiorToModel(cl ContentLayout) model.InteriorLayout {
	return model.InteriorLayout{
		IconX: cl.IconX, IconY: cl.IconY, IconSize: cl.IconSize, FontSize: cl.FontSize, SubtitleSize: cl.SubtitleSize,
		TitleX: cl.TitleX, TitleStartY: cl.TitleStartY, TitleLineH: cl.TitleLineH,
		TitleLines: cl.TitleLines, SubtitleX: cl.SubtitleX, SubtitleY: cl.SubtitleY,
		HasSubtitle: cl.HasSubtitle, TopAlign: cl.TopAlign, InlineIcon: cl.InlineIcon,
		MinW: cl.MinW, MinH: cl.MinH,
		WritableX: cl.WritableX, WritableY: cl.WritableY, WritableW: cl.WritableW, WritableH: cl.WritableH,
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
