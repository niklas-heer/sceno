package layout

import (
	"fmt"
	"math"
	"sort"

	"github.com/niklas-heer/sceno/internal/geom"
	"github.com/niklas-heer/sceno/internal/model"
)

// jogTolerance is the longest staircase residue (px) collapsed after routing.
const jogTolerance = 8.0

const (
	minInteriorSegment   = 16.0
	routeBorderClearance = 8.0
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
		pts = geom.CollapseJogs(pts, jogTolerance)
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
	fanOutSharedPorts(d, byID, pad)
}

type portUse struct {
	route       int
	target      bool
	node        model.Node
	counterpart model.Node
	side        model.Side
	key         string
}

// fanOutSharedPorts distributes sibling endpoints over the straight usable
// span of a node side. Sorting by counterpart position makes the result stable
// and prevents crossed sibling connectors.
func fanOutSharedPorts(d *model.Diagram, byID map[string]*model.Node, pad float64) {
	groups := map[string][]portUse{}
	for i, re := range d.Routed {
		from, to := byID[re.Edge.From], byID[re.Edge.To]
		if from == nil || to == nil {
			continue
		}
		if _, ok := geom.SlidingAnchor(*from, re.Edge.FromSide, sideCenter(*from, re.Edge.FromSide)); ok {
			groupKey := "source\x00" + from.ID + "\x00" + string(re.Edge.FromSide)
			groups[groupKey] = append(groups[groupKey], portUse{route: i, node: *from, counterpart: *to, side: re.Edge.FromSide, key: re.Key})
		}
		if _, ok := geom.SlidingAnchor(*to, re.Edge.ToSide, sideCenter(*to, re.Edge.ToSide)); ok {
			groupKey := "target\x00" + to.ID + "\x00" + string(re.Edge.ToSide)
			groups[groupKey] = append(groups[groupKey], portUse{route: i, target: true, node: *to, counterpart: *from, side: re.Edge.ToSide, key: re.Key})
		}
	}

	starts := make([]geom.Point, len(d.Routed))
	ends := make([]geom.Point, len(d.Routed))
	changed := make([]bool, len(d.Routed))
	for i, re := range d.Routed {
		pts := geom.SlicesToPath(re.Points)
		if len(pts) >= 2 {
			starts[i], ends[i] = pts[0], pts[len(pts)-1]
		}
	}
	for _, uses := range groups {
		if len(uses) < 2 {
			continue
		}
		sort.SliceStable(uses, func(i, j int) bool {
			ai, aj := counterpartOrder(uses[i]), counterpartOrder(uses[j])
			if ai != aj {
				return ai < aj
			}
			bi, bj := counterpartCrossOrder(uses[i]), counterpartCrossOrder(uses[j])
			if bi != bj {
				return bi < bj
			}
			return uses[i].key < uses[j].key
		})
		lo, hi := usableSideSpan(uses[0].node, uses[0].side)
		for i, use := range uses {
			along := lo + (hi-lo)*float64(i)/float64(len(uses)-1)
			p, _ := geom.SlidingAnchor(use.node, use.side, along)
			if use.target {
				ends[use.route] = p
			} else {
				starts[use.route] = p
			}
			changed[use.route] = true
		}
	}

	for i, re := range d.Routed {
		if !changed[i] {
			continue
		}
		a, b := byID[re.Edge.From], byID[re.Edge.To]
		if a == nil || b == nil {
			continue
		}
		obstacles := obstacleNodes(d.Nodes, re.Edge.From, re.Edge.To, pad)
		pts := routeWithLane(starts[i], ends[i], obstacles, pad, 0, re.Edge.FromSide, re.Edge.ToSide)
		pts = geom.CollapseJogs(pts, jogTolerance)
		if d.Style == model.StyleSketch && len(pts) >= 3 {
			pts = geom.SmoothPath(pts, 8)
			pts[0], pts[len(pts)-1] = starts[i], ends[i]
		}
		path := pointsToPath(pts)
		d.Routed[i].Points = path
		d.EdgePaths[re.Key] = path
	}
}

func sideCenter(n model.Node, side model.Side) float64 {
	if geom.IsHorizontalSide(side) {
		return n.Rect.CY()
	}
	return n.Rect.CX()
}

func usableSideSpan(n model.Node, side model.Side) (float64, float64) {
	if geom.IsHorizontalSide(side) {
		return n.Rect.Y + geom.SlidingPortInset, n.Rect.Bottom() - geom.SlidingPortInset
	}
	return n.Rect.X + geom.SlidingPortInset, n.Rect.Right() - geom.SlidingPortInset
}

func counterpartOrder(use portUse) float64 {
	if geom.IsHorizontalSide(use.side) {
		return use.counterpart.Rect.CY()
	}
	return use.counterpart.Rect.CX()
}

func counterpartCrossOrder(use portUse) float64 {
	if geom.IsHorizontalSide(use.side) {
		return use.counterpart.Rect.CX()
	}
	return use.counterpart.Rect.CY()
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
		score += endpointSeparationPenalty(start, pair.from, a, existing, false)
		score += endpointSeparationPenalty(end, pair.to, b, existing, true)
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
	for i := 1; i < len(pts); i++ {
		length := math.Hypot(pts[i].X-pts[i-1].X, pts[i].Y-pts[i-1].Y)
		if i > 1 && i < len(pts)-1 && length < minInteriorSegment {
			score += (minInteriorSegment - length) * 1200
		}
	}
	if reversals := sameAxisReversals(pts); reversals > 0 {
		score += float64(reversals) * 100000
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
			a, b := pts[i-1], pts[i]
			if geom.SegmentHitsRect(a, b, n.Rect, 0) {
				score += 50000
				continue
			}
			if geom.SegmentHitsRect(a, b, n.Rect, pad) {
				score += 2000
			}
			if clearance := segmentRectDistance(a, b, n.Rect); clearance < routeBorderClearance {
				score += (routeBorderClearance - clearance) * 1500
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

func sameAxisReversals(pts []geom.Point) int {
	reversals := 0
	lastX, lastY := 0.0, 0.0
	for i := 1; i < len(pts); i++ {
		dx, dy := pts[i].X-pts[i-1].X, pts[i].Y-pts[i-1].Y
		if math.Abs(dx) >= math.Abs(dy) && math.Abs(dx) >= .5 {
			direction := math.Copysign(1, dx)
			if lastX != 0 && direction != lastX {
				reversals++
			}
			lastX = direction
		} else if math.Abs(dy) >= .5 {
			direction := math.Copysign(1, dy)
			if lastY != 0 && direction != lastY {
				reversals++
			}
			lastY = direction
		}
	}
	return reversals
}

func segmentRectDistance(a, b geom.Point, r model.Rect) float64 {
	if math.Abs(a.Y-b.Y) < .5 {
		lo, hi := math.Min(a.X, b.X), math.Max(a.X, b.X)
		dx := intervalGap(lo, hi, r.X, r.Right())
		dy := intervalGap(a.Y, a.Y, r.Y, r.Bottom())
		return math.Hypot(dx, dy)
	}
	if math.Abs(a.X-b.X) < .5 {
		lo, hi := math.Min(a.Y, b.Y), math.Max(a.Y, b.Y)
		dx := intervalGap(a.X, a.X, r.X, r.Right())
		dy := intervalGap(lo, hi, r.Y, r.Bottom())
		return math.Hypot(dx, dy)
	}
	return math.Min(math.Hypot(a.X-r.CX(), a.Y-r.CY()), math.Hypot(b.X-r.CX(), b.Y-r.CY()))
}

func intervalGap(aLo, aHi, bLo, bHi float64) float64 {
	if aHi < bLo {
		return bLo - aHi
	}
	if bHi < aLo {
		return aLo - bHi
	}
	return 0
}

func endpointSeparationPenalty(point geom.Point, side model.Side, node model.Node, existing []model.RoutedEdge, target bool) float64 {
	penalty := 0.0
	multiplier := 8.0
	if _, sliding := geom.SlidingAnchor(node, side, sideCenter(node, side)); !sliding {
		// Curved/tapered sides only touch the bbox at their midpoint, so they
		// cannot fan out safely. Make an adjacent side decisively cheaper than
		// stacking another endpoint at the same painted point.
		multiplier = 5000
	}
	for _, routed := range existing {
		edgeNode, edgeSide := routed.Edge.From, routed.Edge.FromSide
		index := 0
		if target {
			edgeNode, edgeSide = routed.Edge.To, routed.Edge.ToSide
			index = len(routed.Points) - 1
		}
		if edgeNode != node.ID || edgeSide != side || len(routed.Points) == 0 || index < 0 {
			continue
		}
		other := routed.Points[index]
		if len(other) < 2 {
			continue
		}
		distance := math.Hypot(point.X-other[0], point.Y-other[1])
		if distance < 14 {
			penalty += (14 - distance) * multiplier
		}
	}
	return penalty
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
