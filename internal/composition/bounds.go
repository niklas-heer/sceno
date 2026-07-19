// Package composition defines canvas geometry shared by rendering and analysis.
package composition

import (
	"math"

	"github.com/niklas-heer/sceno/internal/fonts"
	"github.com/niklas-heer/sceno/internal/measure"
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/theme"
)

// Bounds returns the export viewport, including balanced padding and title chrome.
func Bounds(d model.Diagram) (minX, minY, maxX, maxY float64) {
	minX, minY = 1e9, 1e9
	maxX, maxY = -1e9, -1e9
	expand := func(x, y float64) {
		minX = math.Min(minX, x)
		minY = math.Min(minY, y)
		maxX = math.Max(maxX, x)
		maxY = math.Max(maxY, y)
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

	pad := math.Max(d.Padding, 16) + theme.CanvasPaddingExtra
	minX -= pad
	maxX += pad
	maxY += pad
	headerHeight := pad
	if d.Subtitle != "" {
		headerHeight = theme.HeaderFullHeight
	} else if d.Title != "" {
		headerHeight = theme.HeaderTitleOnlyHeight
	}
	minY -= headerHeight

	// Long chrome widens symmetrically so diagram content remains centered.
	requiredTextW := 0.0
	if d.Title != "" {
		requiredTextW = measure.TextWidth(d.Title, theme.TitleSize, fonts.WeightBold)
	}
	if d.Subtitle != "" {
		requiredTextW = math.Max(requiredTextW, measure.TextWidth(d.Subtitle, theme.SubtitleSize, fonts.WeightRegular))
	}
	if requiredW := requiredTextW + theme.CanvasTextInset*2; requiredW > maxX-minX {
		center := (minX + maxX) / 2
		minX = center - requiredW/2
		maxX = center + requiredW/2
	}
	return minX, minY, maxX, maxY
}

// ChromeBounds returns the painted title/subtitle text region in canvas coordinates.
func ChromeBounds(d model.Diagram) (model.Rect, bool) {
	if d.Title == "" && d.Subtitle == "" {
		return model.Rect{}, false
	}
	minX, minY, _, _ := Bounds(d)
	w := 0.0
	top := math.Inf(1)
	bottom := math.Inf(-1)
	if d.Title != "" {
		w = measure.TextWidth(d.Title, theme.TitleSize, fonts.WeightBold)
		top = minY + theme.HeaderTitleBaseline - theme.TitleSize
		bottom = minY + theme.HeaderTitleBaseline + 4
	}
	if d.Subtitle != "" {
		w = math.Max(w, measure.TextWidth(d.Subtitle, theme.SubtitleSize, fonts.WeightRegular))
		top = math.Min(top, minY+theme.HeaderSubtitleBaseline-theme.SubtitleSize)
		bottom = math.Max(bottom, minY+theme.HeaderSubtitleBaseline+4)
	}
	return model.Rect{X: minX + theme.CanvasTextInset, Y: top, W: w, H: bottom - top}, true
}
