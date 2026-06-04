package layout

import (
	"fmt"

	"github.com/niklas-heer/sceno/internal/geom"
	"github.com/niklas-heer/sceno/internal/model"
)

// FindEdgeCollisions detects edges crossing nodes or other edges.
func FindEdgeCollisions(d *model.Diagram) []model.EdgeCollision {
	if d == nil {
		return nil
	}
	var out []model.EdgeCollision
	pad := d.Gap * 0.5

	for _, re := range d.Routed {
		pts := pathToPoints(re.Points)
		for _, n := range d.Nodes {
			if n.ID == re.Edge.From || n.ID == re.Edge.To || model.IsContainer(n.Kind) {
				continue
			}
			for i := 1; i < len(pts)-1; i++ {
				if geom.SegmentHitsRect(pts[i-1], pts[i], n.Rect, pad) {
					out = append(out, model.EdgeCollision{
						EdgeKey: re.Key,
						Kind:    "node_crossing",
						With:    n.ID,
					})
					break
				}
			}
		}
	}

	for i := 0; i < len(d.Routed); i++ {
		for j := i + 1; j < len(d.Routed); j++ {
			if pathsCross(d.Routed[i].Points, d.Routed[j].Points) {
				out = append(out, model.EdgeCollision{
					EdgeKey: d.Routed[i].Key,
					Kind:    "edge_crossing",
					With:    d.Routed[j].Key,
				})
			}
		}
	}
	return out
}

// RerouteCollidingEdges re-routes edges that hit obstacles using outer buses.
func RerouteCollidingEdges(d *model.Diagram) int {
	fixed := 0
	colls := FindEdgeCollisions(d)
	if len(colls) == 0 {
		return 0
	}
	byID := index(d.Nodes)

	// Process node crossings first (errors), then edge crossings.
	for _, c := range colls {
		if c.Kind != "node_crossing" {
			continue
		}
		if rerouteEdge(d, byID, c) {
			fixed++
		}
	}
	for _, c := range colls {
		if c.Kind == "node_crossing" {
			continue
		}
		if rerouteEdge(d, byID, c) {
			fixed++
		}
	}
	return fixed
}

func rerouteEdge(d *model.Diagram, byID map[string]*model.Node, c model.EdgeCollision) bool {
	for i := range d.Routed {
		if d.Routed[i].Key != c.EdgeKey {
			continue
		}
		re := &d.Routed[i]
		a, b := byID[re.Edge.From], byID[re.Edge.To]
		if a == nil || b == nil {
			return false
		}
		fs, ts := resolveSides(re.Edge, a, b)
		start := geom.Anchor(*a, fs)
		end := geom.Anchor(*b, ts)
		obs := obstacleNodes(d.Nodes, re.Edge.From, re.Edge.To, d.Gap)

		for lane := 0; lane < 48; lane++ {
			pts := routeWithLane(start, end, obs, d.Gap, float64(lane)*d.Gap*0.5)
			pts = geom.SimplifyPath(pts)
			if !pathHitsNodes(pts, d.Nodes, re.Edge.From, re.Edge.To, d.Gap*0.5) {
				re.Points = pointsToPath(pts)
				d.EdgePaths[re.Key] = re.Points
				return true
			}
		}
	}
	return false
}

func pathHitsNodes(pts []geom.Point, nodes []model.Node, skipA, skipB string, pad float64) bool {
	for _, n := range nodes {
		if n.ID == skipA || n.ID == skipB || model.IsContainer(n.Kind) {
			continue
		}
		for i := 1; i < len(pts)-1; i++ {
			if geom.SegmentHitsRect(pts[i-1], pts[i], n.Rect, pad) {
				return true
			}
		}
	}
	return false
}

// FormatEdgeCollision for diagnostics.
func FormatEdgeCollision(c model.EdgeCollision) string {
	return fmt.Sprintf("%s crosses %s (%s)", c.EdgeKey, c.With, c.Kind)
}
