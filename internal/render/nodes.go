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
	if model.IsContainer(k) {
		if n.Label != "" {
			b.WriteString(containerLabelSVG(n))
		}
		return b.String()
	}
	if n.Icon != "" {
		ix, iy := IconRect(n, iconSize)
		b.WriteString(icons.Group(n.Icon, ix, iy, iconSize, paint.FgSecondary))
	}
	b.WriteString(polishedLabel(n))
	return b.String()
}

func containerLabelSVG(n model.Node) string {
	return textEl(n.Label, n.Rect.X+14, n.Rect.Y+14, theme.LaneLabelSize, paint.FgMuted, "600")
}

func polishedLabel(n model.Node) string {
	if n.Label == "" && n.Subtitle == "" {
		return ""
	}
	cl := measure.LayoutFor(n)
	fs := cl.FontSize
	if fs <= 0 {
		fs = theme.NodeSize
	}
	lines := strings.Split(n.Label, "\n")
	lh := cl.TitleLineH
	var b strings.Builder
	for i, line := range lines {
		tw := measure.TextWidth(line, fs, fonts.WeightMedium)
		tx := n.Rect.X + cl.WritableX + (cl.WritableW-tw)/2
		if cl.InlineIcon {
			tx = n.Rect.X + cl.TitleX
		}
		y := n.Rect.Y + cl.TitleStartY + float64(i)*lh
		b.WriteString(textEl(line, tx, y, fs, paint.FgPrimary, "500"))
	}
	if cl.HasSubtitle {
		subSize := cl.SubtitleSize
		if subSize <= 0 {
			subSize = theme.SubSize
		}
		sw := measure.TextWidth(n.Subtitle, subSize, fonts.WeightRegular)
		sx := n.Rect.X + cl.WritableX + (cl.WritableW-sw)/2
		if cl.InlineIcon {
			sx = n.Rect.X + cl.TitleX
		}
		b.WriteString(textEl(n.Subtitle, sx, n.Rect.Y+cl.SubtitleY, subSize, paint.FgMuted, ""))
	}
	return b.String()
}
