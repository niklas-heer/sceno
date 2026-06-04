package render

import (
	"strings"

	"github.com/niklas-heer/sceno/internal/fonts"
	"github.com/niklas-heer/sceno/internal/icons"
	"github.com/niklas-heer/sceno/internal/measure"
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/theme"
)

const iconSize = 18

func polishedNodeSVG(n model.Node, dropShadow bool) string {
	var b strings.Builder
	k := model.NormalizeShape(n.Kind)
	if k == model.ShapeCode {
		b.WriteString(codeBlockSVG(n))
		return b.String()
	}
	if k == model.ShapeActor && n.Icon != "" {
		b.WriteString(actorIconBackdropSVG(n, dropShadow))
	} else {
		b.WriteString(shapeSVG(n, dropShadow))
	}
	if n.Icon != "" {
		ix, iy := n.Rect.X+12, n.Rect.Y+12
		if k == model.ShapeActor {
			ix = n.Rect.X + (n.Rect.W-iconSize)/2
			iy = n.Rect.Y + n.Rect.H*0.14
		}
		b.WriteString(icons.Group(n.Icon, ix, iy, iconSize, paint.FgMuted))
	}
	b.WriteString(polishedLabel(n))
	if model.IsContainer(k) && n.Label != "" {
		b.WriteString(textEl(n.Label, n.Rect.X+14, n.Rect.Y+14, theme.LaneLabelSize, paint.FgMuted, "600"))
	}
	return b.String()
}

func polishedLabel(n model.Node) string {
	if n.Label == "" && n.Subtitle == "" {
		return ""
	}
	fs := n.FontSize
	if fs <= 0 {
		fs = theme.NodeSize // typographic scale constant
	}
	lines := strings.Split(n.Label, "\n")
	lineH := fs * 1.25
	iconOff := 0.0
	contentW := n.Rect.W - measure.PadX
	if n.Icon != "" {
		iconOff = 14
		contentW -= measure.IconColumn
	}
	totalH := float64(len(lines))*lineH + iconOff
	if n.Subtitle != "" {
		totalH += 14
	}
	padTop := 14.0
	padLeft := 14.0
	topAlign := false
	k := model.NormalizeShape(n.Kind)
	var startY float64
	switch k {
	case model.ShapeInfobox, model.ShapeCallout:
		topAlign = true
		padLeft = 18
		padTop = 16
		startY = n.Rect.Y + padTop + lineH*0.75 + iconOff
	case model.ShapeActor:
		topAlign = true
		startY = n.Rect.Y + n.Rect.H*0.56 + lineH*0.75 + iconOff
	default:
		startY = n.Rect.CY() - totalH/2 + lineH*0.75 + iconOff/2
	}
	var b strings.Builder
	for i, line := range lines {
		tw := measure.TextWidth(line, fs, fonts.WeightMedium)
		x := labelX(n, tw, contentW)
		if topAlign && n.Icon == "" {
			x = n.Rect.X + padLeft + (contentW-tw)/2
		}
		y := startY + float64(i)*lineH
		b.WriteString(textEl(line, x, y, fs, paint.FgPrimary, "500"))
	}
	if n.Subtitle != "" {
		subY := n.Rect.Bottom() - 16
		if topAlign {
			subY = startY + float64(len(lines))*lineH + 4
		}
		b.WriteString(textEl(n.Subtitle, n.Rect.X+padLeft, subY, theme.SubSize, paint.FgMuted, ""))
	}
	return b.String()
}

func labelX(n model.Node, textW, contentW float64) float64 {
	if n.Icon != "" {
		return n.Rect.X + measure.IconColumn + (contentW-textW)/2
	}
	return n.Rect.X + (n.Rect.W-textW)/2
}
