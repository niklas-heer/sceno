package geom

import (
	"math"
	"strings"

	"github.com/niklas-heer/sceno/internal/fonts"
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/theme"
)

// Edge label layout constants — shared by render and scene validation (must stay in sync).
const (
	EdgeLabelPadX     = 6.0
	EdgeLabelPadY     = 2.0
	EdgeLabelLineMult = 1.2
	// EdgeLabelClearRun is the minimum visible connector on each side of a label.
	EdgeLabelClearRun = 18.0
	// EdgeLabelObstacleClearance is the minimum pill-to-node/chrome gap.
	EdgeLabelObstacleClearance = 6.0
)

// EdgeLabelLayout is the computed label box used for draw and validate.
type EdgeLabelLayout struct {
	CenterX, CenterY float64
	BoxW, BoxH       float64
	Horizontal       bool
	FontSize         float64
	LineH            float64
	Lines            []string
	BlockedBy        []EdgeLabelObstacle
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
		w := fonts.TextWidth(line, fontSize, fonts.WeightMedium)
		if w > maxW {
			maxW = w
		}
	}
	rx, ry, boxW, boxH, horiz, blocked := layoutEdgeLabelBox(pts, fontSize, lineH, lines, maxW, ctx)
	return EdgeLabelLayout{
		CenterX: rx, CenterY: ry, BoxW: boxW, BoxH: boxH, Horizontal: horiz,
		FontSize: fontSize, LineH: lineH, Lines: lines, BlockedBy: blocked,
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
	rx, ry, boxW, boxH, horizontal, _ = layoutEdgeLabelBox(pts, fontSize, lineH, lines, maxTextW, ctx)
	return
}

func layoutEdgeLabelBox(pts []Point, fontSize, lineH float64, lines []string, maxTextW float64, ctx *EdgeLabelContext) (rx, ry, boxW, boxH float64, horizontal bool, blocked []EdgeLabelObstacle) {
	if len(pts) < 2 || len(lines) == 0 {
		return 0, 0, 0, 0, true, nil
	}
	if maxTextW < 24 {
		maxTextW = 24
	}
	boxW = maxTextW + EdgeLabelPadX*2
	n := float64(len(lines))
	boxH = fontSize*n + EdgeLabelPadY*2
	if len(lines) > 1 {
		boxH += (lineH - fontSize) * (n - 1)
	}
	obstacles := labelObstacles(ctx)
	rx, ry, horizontal, blocked = bestLabelPosition(pts, boxW, boxH, obstacles)
	return
}

func labelObstacles(ctx *EdgeLabelContext) []EdgeLabelObstacle {
	if ctx == nil {
		return nil
	}
	out := make([]EdgeLabelObstacle, 0, len(ctx.Avoid)+2)
	if ctx.From.W > 0 && ctx.From.H > 0 {
		out = append(out, EdgeLabelObstacle{ID: "from", Kind: "node", Bounds: ctx.From})
	}
	if ctx.To.W > 0 && ctx.To.H > 0 {
		out = append(out, EdgeLabelObstacle{ID: "to", Kind: "node", Bounds: ctx.To})
	}
	out = append(out, ctx.Avoid...)
	return out
}

type labelCandidate struct {
	x, y       float64
	horizontal bool
	score      float64
	blocked    []EdgeLabelObstacle
}

func bestLabelPosition(pts []Point, boxW, boxH float64, obstacles []EdgeLabelObstacle) (float64, float64, bool, []EdgeLabelObstacle) {
	best := labelCandidate{score: math.Inf(1)}
	for i := 1; i < len(pts); i++ {
		a, b := pts[i-1], pts[i]
		dx, dy := b.X-a.X, b.Y-a.Y
		horizontal := math.Abs(dx) >= math.Abs(dy)
		length := math.Hypot(dx, dy)
		if length < 1 {
			continue
		}
		mid := (a.X + b.X) / 2
		lo, hi := math.Min(a.X, b.X), math.Max(a.X, b.X)
		half := boxW / 2
		cross := (a.Y + b.Y) / 2
		if !horizontal {
			mid = (a.Y + b.Y) / 2
			lo, hi = math.Min(a.Y, b.Y), math.Max(a.Y, b.Y)
			half = boxH / 2
			cross = (a.X + b.X) / 2
		}
		usableLo, usableHi := lo+half+EdgeLabelClearRun, hi-half-EdgeLabelClearRun
		if usableLo > usableHi {
			usableLo, usableHi = lo+half, hi-half
		}
		if usableLo > usableHi {
			usableLo, usableHi = mid, mid
		}
		positions := []float64{clamp(mid, usableLo, usableHi), usableLo, usableHi}
		relaxedLo, relaxedHi := lo+half+EdgeLabelObstacleClearance, hi-half-EdgeLabelObstacleClearance
		if relaxedLo <= relaxedHi {
			positions = append(positions, clamp(mid, relaxedLo, relaxedHi), relaxedLo, relaxedHi)
		}
		for _, obstacle := range obstacles {
			r := obstacle.Bounds
			if horizontal {
				positions = append(positions,
					clamp(r.X-EdgeLabelObstacleClearance-half, usableLo, usableHi),
					clamp(r.Right()+EdgeLabelObstacleClearance+half, usableLo, usableHi))
				if relaxedLo <= relaxedHi {
					positions = append(positions,
						clamp(r.X-EdgeLabelObstacleClearance-half, relaxedLo, relaxedHi),
						clamp(r.Right()+EdgeLabelObstacleClearance+half, relaxedLo, relaxedHi))
				}
			} else {
				positions = append(positions,
					clamp(r.Y-EdgeLabelObstacleClearance-half, usableLo, usableHi),
					clamp(r.Bottom()+EdgeLabelObstacleClearance+half, usableLo, usableHi))
				if relaxedLo <= relaxedHi {
					positions = append(positions,
						clamp(r.Y-EdgeLabelObstacleClearance-half, relaxedLo, relaxedHi),
						clamp(r.Bottom()+EdgeLabelObstacleClearance+half, relaxedLo, relaxedHi))
				}
			}
		}
		for _, pos := range positions {
			candidate := labelCandidate{x: pos, y: cross, horizontal: horizontal}
			if !horizontal {
				candidate.x, candidate.y = cross, pos
			}
			box := LabelBoxRect(candidate.x, candidate.y, boxW, boxH)
			overlapArea := 0.0
			for _, obstacle := range obstacles {
				if labelRectTooClose(box, obstacle.Bounds, EdgeLabelObstacleClearance) {
					candidate.blocked = append(candidate.blocked, obstacle)
					overlapArea += expandedOverlapArea(box, obstacle.Bounds, EdgeLabelObstacleClearance)
				}
			}
			candidate.score = math.Abs(pos-mid) - length*.05
			if !horizontal {
				candidate.score += 1
			}
			candidate.score += float64(len(candidate.blocked))*1e6 + overlapArea*1e3
			if candidate.score < best.score-1e-6 {
				best = candidate
			}
		}
	}
	if math.IsInf(best.score, 1) {
		x, y, horizontal := LabelPlacement(pts)
		return x, y, horizontal, nil
	}
	return best.x, best.y, best.horizontal, best.blocked
}

func clamp(v, lo, hi float64) float64 {
	return math.Max(lo, math.Min(v, hi))
}

func labelRectTooClose(label, obstacle model.Rect, clearance float64) bool {
	expanded := model.Rect{X: obstacle.X - clearance, Y: obstacle.Y - clearance, W: obstacle.W + clearance*2, H: obstacle.H + clearance*2}
	return label.X < expanded.Right() && label.Right() > expanded.X && label.Y < expanded.Bottom() && label.Bottom() > expanded.Y
}

func expandedOverlapArea(label, obstacle model.Rect, clearance float64) float64 {
	expanded := model.Rect{X: obstacle.X - clearance, Y: obstacle.Y - clearance, W: obstacle.W + clearance*2, H: obstacle.H + clearance*2}
	w := math.Max(0, math.Min(label.Right(), expanded.Right())-math.Max(label.X, expanded.X))
	h := math.Max(0, math.Min(label.Bottom(), expanded.Bottom())-math.Max(label.Y, expanded.Y))
	return w * h
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
