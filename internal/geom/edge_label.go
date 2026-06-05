package geom

import (
	"strings"

	"github.com/niklas-heer/sceno/internal/fonts"
	"github.com/niklas-heer/sceno/internal/measure"
	"github.com/niklas-heer/sceno/internal/theme"
)

// Edge label layout constants — shared by render and scene validation (must stay in sync).
const (
	EdgeLabelPadX     = 6.0
	EdgeLabelPadY     = 2.0
	EdgeLabelLineMult = 1.2
)

// EdgeLabelLayout is the computed label box used for draw and validate.
type EdgeLabelLayout struct {
	CenterX, CenterY float64
	BoxW, BoxH       float64
	Horizontal       bool
	FontSize         float64
	LineH            float64
	Lines            []string
}

// LayoutEdgeLabel computes label geometry from a routed path (render + engine SoT).
func LayoutEdgeLabel(pts []Point, label string, ctx *EdgeLabelContext) EdgeLabelLayout {
	lines := splitLabelLines(label)
	if len(lines) == 0 || len(pts) < 2 {
		return EdgeLabelLayout{}
	}
	fontSize := float64(theme.SubSize)
	lineH := fontSize * EdgeLabelLineMult
	maxW := 0.0
	for _, line := range lines {
		w := measure.TextWidth(line, fontSize, fonts.WeightMedium)
		if w > maxW {
			maxW = w
		}
	}
	rx, ry, boxW, boxH, horiz := edgeLabelBoxInner(pts, fontSize, lineH, lines, maxW, ctx)
	return EdgeLabelLayout{
		CenterX: rx, CenterY: ry, BoxW: boxW, BoxH: boxH, Horizontal: horiz,
		FontSize: fontSize, LineH: lineH, Lines: lines,
	}
}

func splitLabelLines(label string) []string {
	s := strings.TrimSpace(label)
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func edgeLabelBoxInner(pts []Point, fontSize, lineH float64, lines []string, maxTextW float64, ctx *EdgeLabelContext) (rx, ry, boxW, boxH float64, horizontal bool) {
	if len(pts) < 2 || len(lines) == 0 {
		return 0, 0, 0, 0, true
	}
	x, y, horiz := LabelPlacement(pts)
	if maxTextW < 24 {
		maxTextW = 24
	}
	boxW = maxTextW + EdgeLabelPadX*2
	n := float64(len(lines))
	boxH = fontSize*n + EdgeLabelPadY*2
	if len(lines) > 1 {
		boxH += (lineH - fontSize) * (n - 1)
	}
	if horiz {
		rx = x
		if ctx != nil {
			gapLeft := ctx.From.Right() + 6
			gapRight := ctx.To.X - 6
			if gapRight > gapLeft {
				rx = (gapLeft + gapRight) / 2
			}
		}
		ry = y
		return rx, ry, boxW, boxH, true
	}
	const gap = 12.0
	rx = x + boxW/2 + gap
	ry = y
	return rx, ry, boxW, boxH, false
}

// LabelRect returns axis-aligned bounds for a layout.
func (l EdgeLabelLayout) LabelRect() (x, y, w, h float64) {
	if l.BoxW <= 0 {
		return 0, 0, 0, 0
	}
	r := LabelBoxRect(l.CenterX, l.CenterY, l.BoxW, l.BoxH)
	return r.X, r.Y, r.W, r.H
}

// TextBaselineY returns the SVG/gg baseline for line index i (vertically centered block).
func (l EdgeLabelLayout) TextBaselineY(lineIndex int) float64 {
	if len(l.Lines) == 0 {
		return l.CenterY
	}
	blockH := l.FontSize
	if len(l.Lines) > 1 {
		blockH = l.FontSize*float64(len(l.Lines)) + (l.LineH-l.FontSize)*float64(len(l.Lines)-1)
	}
	top := l.CenterY - blockH/2
	return top + l.FontSize*0.82 + float64(lineIndex)*l.LineH
}
