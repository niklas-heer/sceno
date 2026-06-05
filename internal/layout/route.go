package layout

import (
	"fmt"
	"math"

	"github.com/niklas-heer/sceno/internal/geom"
	"github.com/niklas-heer/sceno/internal/model"
)

// RouteEdges builds obstacle-aware orthogonal paths snapped to node borders.
func RouteEdges(d *model.Diagram) {
	byID := index(d.Nodes)
	d.Routed = make([]model.RoutedEdge, 0, len(d.Edges))
	d.EdgePaths = make(map[string][][]float64, len(d.Edges))
	pad := d.Gap * 0.75
	if pad < 12 {
		pad = 12
	}

	for i, e := range d.Edges {
		a, okA := byID[e.From]
		b, okB := byID[e.To]
		if !okA || !okB {
			continue
		}
		fs, ts := resolveSides(e, a, b)

		start := geom.Anchor(*a, fs)
		end := geom.Anchor(*b, ts)
		obstacles := obstacleNodes(d.Nodes, e.From, e.To, pad)

		pts := routeWithLane(start, end, obstacles, pad, float64(i)*pad*0.6)
		pts = geom.SimplifyPath(pts)
		if d.Style == model.StyleSketch && len(pts) >= 3 {
			pts = geom.SmoothPath(pts, 8)
			pts[0] = start
			pts[len(pts)-1] = end
		}
		key := fmt.Sprintf("%s-%s-%d", e.From, e.To, i)
		path := pointsToPath(pts)
		re := model.RoutedEdge{Edge: e, Key: key, Points: path}
		d.Routed = append(d.Routed, re)
		d.EdgePaths[key] = path
	}
}

func routeWithLane(start, end geom.Point, obstacles []model.Node, pad, laneOff float64) []geom.Point {
	candidates := [][]geom.Point{
		elbowHV(start, end),
		elbowVH(start, end),
	}
	for i := 0; i < 10; i++ {
		off := pad + laneOff + pad*float64(i)
		candidates = append(candidates,
			corridorRoute(start, end, off, true),
			corridorRoute(start, end, off, false),
		)
	}
	for i := -16; i <= 16; i++ {
		if i == 0 {
			continue
		}
		off := pad*float64(i) + laneOff
		candidates = append(candidates, horizontalBus(start, end, off))
	}
	// Detour above/below when siblings share a column between endpoints
	for _, mult := range []float64{1.5, 2.5, 3.5, 4.5} {
		off := pad * mult
		candidates = append(candidates,
			[]geom.Point{start, {X: start.X, Y: start.Y - off}, {X: end.X, Y: start.Y - off}, end},
			[]geom.Point{start, {X: start.X, Y: start.Y + off}, {X: end.X, Y: start.Y + off}, end},
		)
	}
		best := candidates[0]
	bestScore := 1e18
	for _, c := range candidates {
		if len(c) < 2 {
			continue
		}
		sc := scorePath(c, obstacles, start, end, pad)
		if betterPath(sc, c, bestScore, best) {
			bestScore = sc
			best = c
		}
	}
	return snapPathEnds(best, start, end)
}

// betterPath picks a lower score, or deterministically breaks ties (fewer bends, lexicographic).
func betterPath(sc float64, c []geom.Point, bestScore float64, best []geom.Point) bool {
	const eps = 1e-6
	if sc < bestScore-eps {
		return true
	}
	if sc > bestScore+eps {
		return false
	}
	if len(c) != len(best) {
		return len(c) < len(best)
	}
	for i := range c {
		if c[i].X != best[i].X {
			return c[i].X < best[i].X
		}
		if c[i].Y != best[i].Y {
			return c[i].Y < best[i].Y
		}
	}
	return false
}

// snapPathEnds forces endpoints onto shape border anchors (routing may bend nearby).
func snapPathEnds(pts []geom.Point, start, end geom.Point) []geom.Point {
	if len(pts) == 0 {
		return []geom.Point{start, end}
	}
	out := append([]geom.Point(nil), pts...)
	out[0] = start
	out[len(out)-1] = end
	return out
}

func elbowHV(a, b geom.Point) []geom.Point {
	return []geom.Point{a, {X: b.X, Y: a.Y}, b}
}

func elbowVH(a, b geom.Point) []geom.Point {
	return []geom.Point{a, {X: a.X, Y: b.Y}, b}
}

// corridorRoute runs an orthogonal bus between start/end (not outside the canvas).
func horizontalBus(start, end geom.Point, yOff float64) []geom.Point {
	midY := start.Y + yOff
	return []geom.Point{start, {X: start.X, Y: midY}, {X: end.X, Y: midY}, end}
}

func corridorRoute(start, end geom.Point, pad float64, vertical bool) []geom.Point {
	if vertical {
		midX := (start.X + end.X) / 2
		if start.X <= end.X {
			midX = math.Max(start.X, end.X) + pad
		} else {
			midX = math.Min(start.X, end.X) - pad
		}
		return []geom.Point{start, {X: midX, Y: start.Y}, {X: midX, Y: end.Y}, end}
	}
	midY := (start.Y + end.Y) / 2
	if start.Y <= end.Y {
		midY = math.Max(start.Y, end.Y) + pad
	} else {
		midY = math.Min(start.Y, end.Y) - pad
	}
	return []geom.Point{start, {X: start.X, Y: midY}, {X: end.X, Y: midY}, end}
}

func scorePath(pts []geom.Point, obstacles []model.Node, start, end geom.Point, pad float64) float64 {
	score := pathLength(pts)
	for _, n := range obstacles {
		for i := 1; i < len(pts); i++ {
			if geom.SegmentHitsRect(pts[i-1], pts[i], n.Rect, pad) {
				score += 50000
			}
		}
	}
	score += float64(len(pts)) * 6
	if len(pts) == 2 && math.Abs(pts[0].Y-pts[1].Y) < 1 {
		score -= 800
	}
	if len(pts) == 3 && math.Abs(pts[0].Y-pts[1].Y) < 1 && math.Abs(pts[1].Y-pts[2].Y) < 1 {
		score -= 400
	}
	// Penalize routes that extend far beyond the node bounding span
	span := math.Hypot(end.X-start.X, end.Y-start.Y)
	if pathLength(pts) > span*2.5+pad*4 {
		score += 2000
	}
	return score
}

func pathLength(pts []geom.Point) float64 {
	var sum float64
	for i := 1; i < len(pts); i++ {
		sum += math.Hypot(pts[i].X-pts[i-1].X, pts[i].Y-pts[i-1].Y)
	}
	return sum
}

func obstacleNodes(nodes []model.Node, skipA, skipB string, _ float64) []model.Node {
	var out []model.Node
	for _, n := range nodes {
		if n.ID == skipA || n.ID == skipB || model.IsContainer(n.Kind) {
			continue
		}
		out = append(out, n)
	}
	return out
}

func canvasBounds(d *model.Diagram) model.Rect {
	minX, minY := 1e9, 1e9
	maxX, maxY := -1e9, -1e9
	for _, n := range d.Nodes {
		if n.Rect.X < minX {
			minX = n.Rect.X
		}
		if n.Rect.Y < minY {
			minY = n.Rect.Y
		}
		if n.Rect.Right() > maxX {
			maxX = n.Rect.Right()
		}
		if n.Rect.Bottom() > maxY {
			maxY = n.Rect.Bottom()
		}
	}
	return model.Rect{X: minX, Y: minY, W: maxX - minX, H: maxY - minY}
}

func pointsToPath(pts []geom.Point) [][]float64 {
	out := make([][]float64, len(pts))
	for i, p := range pts {
		out[i] = []float64{p.X, p.Y}
	}
	return out
}

func pathsCross(a, b [][]float64) bool {
	pa := pathToPoints(a)
	pb := pathToPoints(b)
	for i := 1; i < len(pa); i++ {
		for j := 1; j < len(pb); j++ {
			if geom.SegmentsCross(pa[i-1], pa[i], pb[j-1], pb[j]) {
				return true
			}
		}
	}
	return false
}

func pathToPoints(path [][]float64) []geom.Point {
	pts := make([]geom.Point, 0, len(path))
	for _, p := range path {
		if len(p) >= 2 {
			pts = append(pts, geom.Point{X: p[0], Y: p[1]})
		}
	}
	return pts
}

func resolveSides(e model.Edge, a, b *model.Node) (model.Side, model.Side) {
	autoF, autoT := geom.BestSides(*a, *b)
	fs, ts := e.FromSide, e.ToSide
	if fs == "" || fs == model.SideAuto {
		fs = autoF
	}
	if ts == "" || ts == model.SideAuto {
		ts = autoT
	}
	return fs, ts
}

func shiftPathBus(r *model.RoutedEdge, offset float64) {
	pts := pathToPoints(r.Points)
	if len(pts) < 3 {
		return
	}
	for i := 1; i < len(pts)-1; i++ {
		pts[i].X += offset
	}
	r.Points = pointsToPath(pts)
}
