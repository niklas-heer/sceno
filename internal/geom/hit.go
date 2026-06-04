package geom

import "github.com/niklas-heer/sceno/internal/model"

// SegmentHitsRect returns true if segment ab intersects padded rect.
func SegmentHitsRect(a, b Point, r model.Rect, pad float64) bool {
	r = PadRect(r, pad)
	// Quick reject
	if (a.X < r.X && b.X < r.X) || (a.X > r.Right() && b.X > r.Right()) ||
		(a.Y < r.Y && b.Y < r.Y) || (a.Y > r.Bottom() && b.Y > r.Bottom()) {
		return false
	}
	// Endpoint inside
	if pointInRect(a, r) || pointInRect(b, r) {
		return true
	}
	// Liang-Barsky / Cohen-Sutherland simplified: test against 4 edges
	corners := []Point{{r.X, r.Y}, {r.Right(), r.Y}, {r.Right(), r.Bottom()}, {r.X, r.Bottom()}}
	for i := 0; i < 4; i++ {
		c := corners[(i+1)%4]
		if segmentsIntersect(a, b, corners[i], c) {
			return true
		}
	}
	return false
}

func pointInRect(p Point, r model.Rect) bool {
	return p.X >= r.X && p.X <= r.Right() && p.Y >= r.Y && p.Y <= r.Bottom()
}

func segmentsIntersect(a, b, c, d Point) bool {
	d1 := cross(c, d, a)
	d2 := cross(c, d, b)
	d3 := cross(a, b, c)
	d4 := cross(a, b, d)
	if ((d1 > 0 && d2 < 0) || (d1 < 0 && d2 > 0)) && ((d3 > 0 && d4 < 0) || (d3 < 0 && d4 > 0)) {
		return true
	}
	return false
}

func cross(a, b, c Point) float64 {
	return (c.X-a.X)*(b.Y-a.Y) - (c.Y-a.Y)*(b.X-a.X)
}

// PathHitsNode tests all segments against a node rect.
func PathHitsNode(pts []Point, n model.Node, pad float64, skipEndpoints bool) bool {
	for i := 1; i < len(pts); i++ {
		a, b := pts[i-1], pts[i]
		if skipEndpoints && (i == 1 || i == len(pts)-1) {
			// still check middle segments fully
		}
		if SegmentHitsRect(a, b, n.Rect, pad) {
			// Allow touching the attached node at endpoints
			if skipEndpoints && i == 1 && (near(a, n) || near(b, n)) {
				continue
			}
			if skipEndpoints && i == len(pts)-1 && (near(a, n) || near(b, n)) {
				continue
			}
			return true
		}
	}
	return false
}

func near(p Point, n model.Node) bool {
	return p.X >= n.Rect.X-2 && p.X <= n.Rect.Right()+2 && p.Y >= n.Rect.Y-2 && p.Y <= n.Rect.Bottom()+2
}

// SegmentsCross tests if two segments intersect.
func SegmentsCross(a1, a2, b1, b2 Point) bool {
	return segmentsIntersect(a1, a2, b1, b2)
}
