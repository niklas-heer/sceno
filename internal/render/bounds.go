package render

import (
	"github.com/niklas-heer/sceno/internal/fonts"
	"github.com/niklas-heer/sceno/internal/measure"
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/theme"
)

func Bounds(d model.Diagram) (minX, minY, maxX, maxY float64) {
	minX, minY = 1e9, 1e9
	maxX, maxY = -1e9, -1e9
	expand := func(x, y float64) {
		if x < minX {
			minX = x
		}
		if y < minY {
			minY = y
		}
		if x > maxX {
			maxX = x
		}
		if y > maxY {
			maxY = y
		}
	}
	for _, n := range d.Nodes {
		expand(n.Rect.X, n.Rect.Y)
		expand(n.Rect.Right(), n.Rect.Bottom())
	}
	for _, re := range d.Routed {
		for _, p := range re.Points {
			if len(p) >= 2 {
				expand(p[0], p[1])
			}
		}
	}
	for _, path := range d.EdgePaths {
		for _, p := range path {
			if len(p) >= 2 {
				expand(p[0], p[1])
			}
		}
	}
	if minX > 1e8 {
		return 0, 0, 800, 600
	}
	pad := d.Padding + 48
	minX -= pad
	minY -= pad
	maxX += pad
	maxY += pad
	if d.Title != "" {
		tw := measure.TextWidth(d.Title, theme.TitleSize, fonts.WeightBold)
		expandText(&minX, &minY, &maxX, &maxY, minX+28, minY+32, tw, theme.TitleSize)
	}
	if d.Subtitle != "" {
		sw := measure.TextWidth(d.Subtitle, theme.SubtitleSize, fonts.WeightRegular)
		expandText(&minX, &minY, &maxX, &maxY, minX+28, minY+56, sw, theme.SubtitleSize)
	}
	return minX, minY, maxX, maxY
}

func expandText(minX, minY, maxX, maxY *float64, x, baseline, textW, size float64) {
	if baseline-size < *minY {
		*minY = baseline - size
	}
	if x+textW > *maxX {
		*maxX = x + textW
	}
	if baseline > *maxY {
		*maxY = baseline
	}
}
