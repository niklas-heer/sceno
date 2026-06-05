package render

import (
	"strings"

	"github.com/niklas-heer/sceno/internal/fonts"
	"github.com/niklas-heer/sceno/internal/icons"
	"github.com/niklas-heer/sceno/internal/measure"
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/theme"
)

const iconSize = measure.IconSize

func polishedNodeSVG(n model.Node, dropShadow bool) string {
	var b strings.Builder
	k := model.NormalizeShape(n.Kind)
	if k == model.ShapeCode {
		b.WriteString(codeBlockSVG(n))
		return b.String()
	}
	b.WriteString(shapeSVG(n, dropShadow))
	if n.Icon != "" {
		ix, iy := IconRect(n, iconSize)
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
		fs = theme.NodeSize
	}
	cl := measure.LayoutFor(n)
	lines := strings.Split(n.Label, "\n")
	lh := cl.TitleLineH
	contentW := n.Rect.W - measure.PadX
	if n.Icon != "" && !cl.TopAlign {
		contentW -= measure.IconColumn
	}
	var b strings.Builder
	for i, line := range lines {
		tw := measure.TextWidth(line, fs, fonts.WeightMedium)
		tx := n.Rect.X + cl.TitleX
		if !cl.TopAlign || n.Icon == "" {
			tx = n.Rect.X + cl.TitleX + (contentW-tw)/2
		}
		if n.Icon != "" && (n.IconPos == "" || n.IconPos == model.IconTopLeft) && !cl.TopAlign {
			tx = n.Rect.X + measure.IconColumn + (contentW-tw)/2
		}
		if cl.TopAlign && n.Icon != "" {
			tx = n.Rect.X + (n.Rect.W-tw)/2
		}
		y := n.Rect.Y + cl.TitleStartY + float64(i)*lh
		b.WriteString(textEl(line, tx, y, fs, paint.FgPrimary, "500"))
	}
	if cl.HasSubtitle {
		b.WriteString(textEl(n.Subtitle, n.Rect.X+cl.SubtitleX, n.Rect.Y+cl.SubtitleY, theme.SubSize, paint.FgMuted, ""))
	}
	return b.String()
}
