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
		obstacles := obstacleNodes(d.Nodes, e.From, e.To, pad)
		fs, ts, pts := chooseRoute(e, *a, *b, obstacles, pad, d.Routed)
		e.FromSide, e.ToSide = fs, ts
		pts = geom.SimplifyPath(pts)
		if d.Style == model.StyleSketch && len(pts) >= 3 {
			start, end := geom.EdgeAnchors(*a, *b, fs, ts)
			pts = geom.SmoothPath(pts, 8)
			pts[0] = start
			pts[len(pts)-1] = end
		}
		key := fmt.Sprintf("%s-%s-%d", e.From, e.To, i)
		path := pointsToPath(pts)
		re := model.RoutedEdge{Edge: e, Key: key, Points: path} // e carries resolved sides
		d.Routed = append(d.Routed, re)
		d.EdgePaths[key] = path
	}
}

type sidePair struct {
	from model.Side
	to   model.Side
}

func chooseRoute(e model.Edge, a, b model.Node, obstacles []model.Node, pad float64, existing []model.RoutedEdge) (model.Side, model.Side, []geom.Point) {
	bestPair := sidePair{}
	var bestPath []geom.Point
	bestScore := 1e18
	idealFrom, idealTo := geom.BestSides(a, b)
	for _, pair := range routeSidePairs(e, idealFrom, idealTo) {
		start, end := geom.EdgeAnchors(a, b, pair.from, pair.to)
		path := routeWithLane(start, end, obstacles, pad, 0, pair.from, pair.to)
		score := scorePath(path, obstacles, start, end, pad, pair.from, pair.to)
		if pair.from != idealFrom {
			score += 240
		}
		if pair.to != idealTo {
			score += 240
		}
		for _, routed := range existing {
			if pathsCross(pointsToPath(path), routed.Points) {
				score += 20000
			}
		}
		if betterPath(score, path, bestScore, bestPath) {
			bestPair, bestPath, bestScore = pair, path, score
		}
	}
	return bestPair.from, bestPair.to, bestPath
}

func routeSidePairs(e model.Edge, idealFrom, idealTo model.Side) []sidePair {
	from := candidateSides(e.FromSide, idealFrom)
	to := candidateSides(e.ToSide, idealTo)
	seen := map[sidePair]bool{}
	var out []sidePair
	for _, fs := range from {
		for _, ts := range to {
			pair := sidePair{from: fs, to: ts}
			if !seen[pair] {
				seen[pair] = true
				out = append(out, pair)
			}
		}
	}
	return out
}

func candidateSides(explicit, ideal model.Side) []model.Side {
	if explicit != "" && explicit != model.SideAuto {
		return []model.Side{explicit}
	}
	out := []model.Side{ideal}
	for _, side := range []model.Side{model.SideRight, model.SideBottom, model.SideLeft, model.SideTop} {
		if side != ideal {
			out = append(out, side)
		}
	}
	return out
}

func routeWithLane(start, end geom.Point, obstacles []model.Node, pad, laneOff float64, fromSide, toSide model.Side) []geom.Point {
	// Keep enough straight shaft before a corner and arrowhead. Direct aligned
	// routes do not need endpoint stubs and remain compact in tight stacks.
	stub := math.Max(geom.ArrowHeadDepth+geom.EdgeLabelClearRun, math.Min(pad, 36))
	innerStart := anchorStub(start, fromSide, stub)
	innerEnd := anchorStub(end, toSide, stub)
	candidates := [][]geom.Point{
		elbowHV(innerStart, innerEnd),
		elbowVH(innerStart, innerEnd),
	}
	for i := 0; i < 10; i++ {
		off := pad + laneOff + pad*float64(i)
		candidates = append(candidates,
			corridorRoute(innerStart, innerEnd, off, true),
			corridorRoute(innerStart, innerEnd, off, false),
		)
	}
	for i := -16; i <= 16; i++ {
		if i == 0 {
			continue
		}
		off := pad*float64(i) + laneOff
		candidates = append(candidates, horizontalBus(innerStart, innerEnd, off))
	}
	// Detour above/below when siblings share a column between endpoints
	for _, mult := range []float64{1.5, 2.5, 3.5, 4.5} {
		off := pad * mult
		candidates = append(candidates,
			[]geom.Point{innerStart, {X: innerStart.X, Y: innerStart.Y - off}, {X: innerEnd.X, Y: innerStart.Y - off}, innerEnd},
			[]geom.Point{innerStart, {X: innerStart.X, Y: innerStart.Y + off}, {X: innerEnd.X, Y: innerStart.Y + off}, innerEnd},
		)
	}
	completed := make([][]geom.Point, 0, len(candidates)+1)
	if laneOff == 0 && directRouteCompatible(start, end, fromSide, toSide) {
		completed = append(completed, []geom.Point{start, end})
	}
	for i := range candidates {
		completed = append(completed, withEndpointStubs(start, end, candidates[i]))
	}
	candidates = completed
	best := candidates[0]
	bestScore := 1e18
	for _, c := range candidates {
		if len(c) < 2 {
			continue
		}
		sc := scorePath(c, obstacles, start, end, pad, fromSide, toSide)
		if betterPath(sc, c, bestScore, best) {
			bestScore = sc
			best = c
		}
	}
	return snapPathEnds(best, start, end)
}

func directRouteCompatible(start, end geom.Point, fromSide, toSide model.Side) bool {
	dx, dy := end.X-start.X, end.Y-start.Y
	if math.Abs(dx) > 1 && math.Abs(dy) > 1 {
		return false
	}
	leaves := func(side model.Side, x, y float64) bool {
		switch side {
		case model.SideLeft:
			return x < 0
		case model.SideRight:
			return x > 0
		case model.SideTop:
			return y < 0
		case model.SideBottom:
			return y > 0
		default:
			return false
		}
	}
	return leaves(fromSide, dx, dy) && leaves(toSide, -dx, -dy)
}

func anchorStub(p geom.Point, side model.Side, distance float64) geom.Point {
	switch side {
	case model.SideLeft:
		p.X -= distance
	case model.SideRight:
		p.X += distance
	case model.SideTop:
		p.Y -= distance
	case model.SideBottom:
		p.Y += distance
	}
	return p
}

func withEndpointStubs(start, end geom.Point, inner []geom.Point) []geom.Point {
	out := make([]geom.Point, 0, len(inner)+2)
	out = append(out, start)
	out = append(out, inner...)
	out = append(out, end)
	return out
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

func scorePath(pts []geom.Point, obstacles []model.Node, start, end geom.Point, pad float64, fromSide, toSide model.Side) float64 {
	score := pathLength(pts)
	if endpointReverses(pts) {
		score += 100000
	}
	dx := math.Abs(end.X - start.X)
	dy := math.Abs(end.Y - start.Y)
	if geom.IsHorizontalSide(fromSide) && geom.IsHorizontalSide(toSide) && dx > dy {
		if len(pts) >= 3 && math.Abs(pts[0].Y-pts[1].Y) > 2 {
			score += 3000
		}
	}
	if (fromSide == model.SideTop || fromSide == model.SideBottom) &&
		(toSide == model.SideTop || toSide == model.SideBottom) && dy > dx {
		if len(pts) >= 3 && math.Abs(pts[0].X-pts[1].X) > 2 {
			score += 3000
		}
	}
	for _, n := range obstacles {
		for i := 1; i < len(pts); i++ {
			if geom.SegmentHitsRect(pts[i-1], pts[i], n.Rect, pad) {
				score += 50000
			}
		}
	}
	score += float64(len(pts)) * 60
	if len(pts) == 2 && math.Abs(pts[0].Y-pts[1].Y) < 1 {
		score -= 800
	}
	if len(pts) == 3 && math.Abs(pts[0].Y-pts[1].Y) < 1 && math.Abs(pts[1].Y-pts[2].Y) < 1 {
		score -= 400
	}
	// Penalize routes that extend far beyond the node bounding span
	span := math.Hypot(end.X-start.X, end.Y-start.Y)
	if pathLength(pts) > span*1.8+pad*2 {
		score += 5000
	}
	return score
}

func endpointReverses(pts []geom.Point) bool {
	if len(pts) < 3 {
		return false
	}
	dot := func(a, b, c geom.Point) float64 {
		return (b.X-a.X)*(c.X-b.X) + (b.Y-a.Y)*(c.Y-b.Y)
	}
	return dot(pts[0], pts[1], pts[2]) < 0 ||
		dot(pts[len(pts)-3], pts[len(pts)-2], pts[len(pts)-1]) < 0
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
			if segmentsConflict(pa[i-1], pa[i], pb[j-1], pb[j]) {
				return true
			}
		}
	}
	return false
}

func segmentsConflict(a1, a2, b1, b2 geom.Point) bool {
	if geom.SegmentsCross(a1, a2, b1, b2) {
		return true
	}
	const eps = 1.0
	if math.Abs(a1.Y-a2.Y) < eps && math.Abs(b1.Y-b2.Y) < eps && math.Abs(a1.Y-b1.Y) < eps {
		return intervalOverlap(a1.X, a2.X, b1.X, b2.X) > eps
	}
	if math.Abs(a1.X-a2.X) < eps && math.Abs(b1.X-b2.X) < eps && math.Abs(a1.X-b1.X) < eps {
		return intervalOverlap(a1.Y, a2.Y, b1.Y, b2.Y) > eps
	}
	return false
}

func intervalOverlap(a1, a2, b1, b2 float64) float64 {
	amin, amax := math.Min(a1, a2), math.Max(a1, a2)
	bmin, bmax := math.Min(b1, b2), math.Max(b1, b2)
	return math.Min(amax, bmax) - math.Max(amin, bmin)
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
