package geom

import (
	"math"
	"sort"

	"github.com/niklas-heer/sceno/internal/model"
)

const (
	// ShapeBorderClearance is the minimum clear space between content and the
	// visible edge of a shape, including its centered stroke.
	ShapeBorderClearance = 6.0
	shapeScanlines       = 48
)

// CubicCurve is one cubic Bézier segment in a closed shape silhouette.
type CubicCurve struct {
	Control1 Point
	Control2 Point
	End      Point
}

// ActorGeometry is the shared line geometry for the UML actor backends.
type ActorGeometry struct {
	CenterX, HeadY, HeadRadius float64
	ShoulderY, ArmY, FootY     float64
	ArmLeft, ArmRight          float64
	LegLeft, LegRight          float64
}

// ActorFigure computes the stick figure and optionally reserves the lower
// portion of the bbox as a separate writable label band.
func ActorFigure(r model.Rect, reserveLabel bool, unit float64) ActorGeometry {
	figureBottom := r.Bottom() - 6*unit
	if reserveLabel {
		figureBottom = r.Y + r.H*.58
	}
	figureH := math.Max(20*unit, figureBottom-r.Y)
	headR := math.Max(4*unit, math.Min(r.W*.16, figureH*.14))
	cx := r.CX()
	headY := r.Y + headR + 4*unit
	shoulderY := math.Min(figureBottom-8*unit, headY+headR+3*unit)
	arm := r.W * .32
	leg := r.W * .22
	return ActorGeometry{
		CenterX: cx, HeadY: headY, HeadRadius: headR,
		ShoulderY: shoulderY, ArmY: shoulderY + 3*unit, FootY: figureBottom,
		ArmLeft: cx - arm, ArmRight: cx + arm, LegLeft: cx - leg, LegRight: cx + leg,
	}
}

// ShapeStrokeWidth returns the visible outline width used by polished backends.
func ShapeStrokeWidth(kind model.ShapeKind) float64 {
	switch model.NormalizeShape(kind) {
	case model.ShapeTextbox, model.ShapeNote, model.ShapeInfobox, model.ShapeLane:
		return 1
	default:
		return 1.5
	}
}

// CylinderRimRY is the shared rim radius for cylinder geometry and rendering.
func CylinderRimRY(w, h float64) float64 {
	return math.Min(math.Min(w*.10, h*.15), 12)
}

// CloudPath returns the exact closed cubic path used for cloud rendering.
func CloudPath(r model.Rect) (Point, []CubicCurve) {
	point := func(x, y float64) Point { return Point{X: r.X + x*r.W, Y: r.Y + y*r.H} }
	curve := func(c1x, c1y, c2x, c2y, x, y float64) CubicCurve {
		return CubicCurve{Control1: point(c1x, c1y), Control2: point(c2x, c2y), End: point(x, y)}
	}
	return point(0, .5), []CubicCurve{
		curve(0, .38, .04, .30, .14, .28),
		curve(.14, .14, .26, .08, .38, .15),
		curve(.40, .06, .44, 0, .50, 0),
		curve(.57, 0, .62, .06, .64, .14),
		curve(.77, .07, .91, .17, .90, .32),
		curve(.97, .35, 1, .42, 1, .50),
		curve(1, .62, .92, .72, .80, .72),
		curve(.76, .86, .64, .88, .56, .82),
		curve(.55, .92, .53, 1, .50, 1),
		curve(.45, 1, .41, .91, .38, .84),
		curve(.25, .90, .10, .82, .12, .68),
		curve(.05, .64, 0, .58, 0, .50),
	}
}

// ShapeOutline returns a deterministic clockwise polyline of the painted
// silhouette. Curves are sampled densely enough for writable-space analysis;
// exact cubic cloud commands remain available through CloudPath.
func ShapeOutline(n model.Node) []Point {
	r := n.Rect
	k := model.NormalizeShape(n.Kind)
	switch k {
	case model.ShapeEllipse, model.ShapeCircle:
		return ellipseOutline(r.CX(), r.CY(), r.W/2, r.H/2, 64)
	case model.ShapeDiamond:
		return []Point{{r.CX(), r.Y}, {r.Right(), r.CY()}, {r.CX(), r.Bottom()}, {r.X, r.CY()}}
	case model.ShapeHexagon:
		return regularPolygon(r, 6, math.Pi/6)
	case model.ShapeOctagon:
		return regularPolygon(r, 8, math.Pi/8)
	case model.ShapeTriangle:
		return []Point{{r.CX(), r.Y}, {r.Right(), r.Bottom()}, {r.X, r.Bottom()}}
	case model.ShapeParallelogram:
		skew := r.W * .15
		return []Point{{r.X + skew, r.Y}, {r.Right(), r.Y}, {r.Right() - skew, r.Bottom()}, {r.X, r.Bottom()}}
	case model.ShapeCylinder:
		return cylinderOutline(r)
	case model.ShapeCloud:
		return cloudOutline(r)
	case model.ShapeDocument:
		fold := math.Min(r.W*.22, 22)
		return []Point{{r.X, r.Y}, {r.Right() - fold, r.Y}, {r.Right(), r.Y + fold}, {r.Right(), r.Bottom()}, {r.X, r.Bottom()}}
	case model.ShapePill:
		return roundedRectOutline(r, r.H/2)
	case model.ShapeTextbox:
		return roundedRectOutline(r, 6)
	case model.ShapeNote:
		return roundedRectOutline(r, 4)
	case model.ShapeInfobox:
		return roundedRectOutline(r, 10)
	case model.ShapeLane:
		return roundedRectOutline(r, 12)
	case model.ShapeFrame:
		return roundedRectOutline(r, 14)
	case model.ShapeActor:
		if n.Icon != "" {
			return roundedRectOutline(r, 14)
		}
		return nil
	default:
		return roundedRectOutline(r, 12)
	}
}

// WritableRect returns the largest axis-aligned rectangle of the requested
// aspect ratio that stays inside the stroked silhouette and avoids internal
// seams such as cylinder rims, note folds, and infobox accent rules.
func WritableRect(n model.Node, aspect float64) model.Rect {
	r := n.Rect
	if r.W <= 0 || r.H <= 0 {
		return model.Rect{}
	}
	if aspect <= 0 || math.IsNaN(aspect) || math.IsInf(aspect, 0) {
		aspect = r.W / r.H
	}
	if model.NormalizeShape(n.Kind) == model.ShapeActor && n.Icon == "" {
		clearance := ShapeBorderClearance + ShapeStrokeWidth(n.Kind)/2
		figure := ActorFigure(r, n.Label != "" || n.Subtitle != "", 1)
		available := model.Rect{
			X: r.X + clearance, Y: figure.FootY + clearance,
			W: math.Max(0, r.W-2*clearance), H: math.Max(0, r.Bottom()-clearance-(figure.FootY+clearance)),
		}
		return fitRectAspect(available, aspect)
	}
	outline := ShapeOutline(n)
	clearance := ShapeBorderClearance + ShapeStrokeWidth(n.Kind)/2
	minY, maxY := r.Y+clearance, r.Bottom()-clearance
	if minY >= maxY || len(outline) < 3 {
		return model.Rect{}
	}

	type span struct{ left, right float64 }
	spans := make([]span, shapeScanlines+1)
	valid := make([]bool, len(spans))
	for i := range spans {
		y := minY + (maxY-minY)*float64(i)/shapeScanlines
		left, right, ok := insetHorizontalSpan(outline, y, clearance)
		spans[i], valid[i] = span{left, right}, ok && right > left
	}

	blocked := shapeInteriorObstacles(n, clearance)
	best := model.Rect{}
	bestArea := 0.0
	for top := 0; top < shapeScanlines; top++ {
		if !valid[top] {
			continue
		}
		left, right := spans[top].left, spans[top].right
		for bottom := top + 1; bottom <= shapeScanlines; bottom++ {
			if !valid[bottom] {
				break
			}
			left = math.Max(left, spans[bottom].left)
			right = math.Min(right, spans[bottom].right)
			if right <= left {
				break
			}
			topY := minY + (maxY-minY)*float64(top)/shapeScanlines
			bottomY := minY + (maxY-minY)*float64(bottom)/shapeScanlines
			availableW, availableH := right-left, bottomY-topY
			w, h := availableW, availableW/aspect
			if h > availableH {
				h, w = availableH, availableH*aspect
			}
			candidate := model.Rect{X: (left + right - w) / 2, Y: (topY + bottomY - h) / 2, W: w, H: h}
			if intersectsAny(candidate, blocked) {
				continue
			}
			if area := w * h; area > bestArea {
				best, bestArea = candidate, area
			}
		}
	}
	return best
}

// ShapeInternalLines returns visible seams or disjoint strokes that are not
// represented by the closed silhouette outline.
func ShapeInternalLines(n model.Node) [][2]Point {
	r := n.Rect
	switch model.NormalizeShape(n.Kind) {
	case model.ShapeActor:
		if n.Icon != "" {
			return nil
		}
		g := ActorFigure(r, n.Label != "" || n.Subtitle != "", 1)
		lines := [][2]Point{
			{{X: g.CenterX, Y: g.ShoulderY}, {X: g.CenterX, Y: g.FootY}},
			{{X: g.ArmLeft, Y: g.ArmY}, {X: g.ArmRight, Y: g.ArmY}},
			{{X: g.CenterX, Y: g.FootY}, {X: g.LegLeft, Y: g.FootY}},
			{{X: g.CenterX, Y: g.FootY}, {X: g.LegRight, Y: g.FootY}},
		}
		head := ellipseOutline(g.CenterX, g.HeadY, g.HeadRadius, g.HeadRadius, 24)
		for i, point := range head {
			lines = append(lines, [2]Point{point, head[(i+1)%len(head)]})
		}
		return lines
	case model.ShapeCylinder:
		ry := CylinderRimRY(r.W, r.H)
		rim := ellipseOutline(r.CX(), r.Y+ry, r.W/2, ry, 32)
		lines := make([][2]Point, len(rim))
		for i, point := range rim {
			lines[i] = [2]Point{point, rim[(i+1)%len(rim)]}
		}
		return lines
	case model.ShapeDocument:
		fold := math.Min(r.W*.22, 22)
		return [][2]Point{
			{{X: r.Right() - fold, Y: r.Y}, {X: r.Right() - fold, Y: r.Y + fold}},
			{{X: r.Right() - fold, Y: r.Y + fold}, {X: r.Right(), Y: r.Y + fold}},
		}
	case model.ShapeNote:
		fold := math.Min(18.0, math.Min(r.W, r.H)*.28)
		return [][2]Point{{{X: r.Right() - fold, Y: r.Bottom()}, {X: r.Right(), Y: r.Bottom() - fold}}}
	case model.ShapeInfobox:
		return [][2]Point{{{X: r.X + 4, Y: r.Y}, {X: r.X + 4, Y: r.Bottom()}}}
	default:
		return nil
	}
}

func fitRectAspect(r model.Rect, aspect float64) model.Rect {
	if r.W <= 0 || r.H <= 0 {
		return model.Rect{}
	}
	w, h := r.W, r.W/aspect
	if h > r.H {
		h, w = r.H, r.H*aspect
	}
	return model.Rect{X: r.CX() - w/2, Y: r.CY() - h/2, W: w, H: h}
}

func insetHorizontalSpan(outline []Point, y, clearance float64) (float64, float64, bool) {
	type hit struct {
		x     float64
		inset float64
	}
	hits := make([]hit, 0, 4)
	for i, a := range outline {
		b := outline[(i+1)%len(outline)]
		if !((a.Y <= y && b.Y > y) || (b.Y <= y && a.Y > y)) {
			continue
		}
		t := (y - a.Y) / (b.Y - a.Y)
		dx, dy := b.X-a.X, b.Y-a.Y
		inset := clearance * math.Hypot(dx, dy) / math.Abs(dy)
		hits = append(hits, hit{x: a.X + t*dx, inset: inset})
	}
	if len(hits) < 2 {
		return 0, 0, false
	}
	sort.Slice(hits, func(i, j int) bool { return hits[i].x < hits[j].x })
	return hits[0].x + hits[0].inset, hits[len(hits)-1].x - hits[len(hits)-1].inset, true
}

func shapeInteriorObstacles(n model.Node, clearance float64) []model.Rect {
	r := n.Rect
	switch model.NormalizeShape(n.Kind) {
	case model.ShapeCylinder:
		ry := CylinderRimRY(r.W, r.H)
		return []model.Rect{{X: r.X, Y: r.Y, W: r.W, H: 2*ry + clearance}}
	case model.ShapeDocument:
		fold := math.Min(r.W*.22, 22)
		return []model.Rect{{X: r.Right() - fold - clearance, Y: r.Y, W: fold + clearance, H: fold + clearance}}
	case model.ShapeNote:
		fold := math.Min(18.0, math.Min(r.W, r.H)*.28)
		return []model.Rect{{X: r.Right() - fold - clearance, Y: r.Bottom() - fold - clearance, W: fold + clearance, H: fold + clearance}}
	case model.ShapeInfobox:
		return []model.Rect{{X: r.X, Y: r.Y, W: 4 + clearance, H: r.H}}
	default:
		return nil
	}
}

func intersectsAny(r model.Rect, obstacles []model.Rect) bool {
	for _, obstacle := range obstacles {
		if r.X < obstacle.Right() && r.Right() > obstacle.X && r.Y < obstacle.Bottom() && r.Bottom() > obstacle.Y {
			return true
		}
	}
	return false
}

func regularPolygon(r model.Rect, sides int, offset float64) []Point {
	points := make([]Point, sides)
	for i := range points {
		a := offset + float64(i)*2*math.Pi/float64(sides)
		points[i] = Point{X: r.CX() + r.W/2*math.Cos(a), Y: r.CY() + r.H/2*math.Sin(a)}
	}
	return points
}

func ellipseOutline(cx, cy, rx, ry float64, steps int) []Point {
	points := make([]Point, steps)
	for i := range points {
		a := -math.Pi/2 + float64(i)*2*math.Pi/float64(steps)
		points[i] = Point{X: cx + rx*math.Cos(a), Y: cy + ry*math.Sin(a)}
	}
	return points
}

func roundedRectOutline(r model.Rect, radius float64) []Point {
	radius = math.Min(math.Max(0, radius), math.Min(r.W, r.H)/2)
	if radius == 0 {
		return []Point{{r.X, r.Y}, {r.Right(), r.Y}, {r.Right(), r.Bottom()}, {r.X, r.Bottom()}}
	}
	const arcSteps = 6
	centers := []Point{{r.Right() - radius, r.Y + radius}, {r.Right() - radius, r.Bottom() - radius}, {r.X + radius, r.Bottom() - radius}, {r.X + radius, r.Y + radius}}
	starts := []float64{-math.Pi / 2, 0, math.Pi / 2, math.Pi}
	var points []Point
	for corner, center := range centers {
		for i := 0; i <= arcSteps; i++ {
			a := starts[corner] + float64(i)*math.Pi/2/arcSteps
			points = append(points, Point{X: center.X + radius*math.Cos(a), Y: center.Y + radius*math.Sin(a)})
		}
	}
	return points
}

func cylinderOutline(r model.Rect) []Point {
	ry := CylinderRimRY(r.W, r.H)
	const steps = 16
	points := make([]Point, 0, steps*2+2)
	for i := 0; i <= steps; i++ {
		a := math.Pi + float64(i)*math.Pi/steps
		points = append(points, Point{X: r.CX() + r.W/2*math.Cos(a), Y: r.Y + ry + ry*math.Sin(a)})
	}
	for i := 0; i <= steps; i++ {
		a := float64(i) * math.Pi / steps
		points = append(points, Point{X: r.CX() + r.W/2*math.Cos(a), Y: r.Bottom() - ry + ry*math.Sin(a)})
	}
	return points
}

func cloudOutline(r model.Rect) []Point {
	start, curves := CloudPath(r)
	points := []Point{start}
	current := start
	for _, curve := range curves {
		for step := 1; step <= 8; step++ {
			t := float64(step) / 8
			u := 1 - t
			points = append(points, Point{
				X: u*u*u*current.X + 3*u*u*t*curve.Control1.X + 3*u*t*t*curve.Control2.X + t*t*t*curve.End.X,
				Y: u*u*u*current.Y + 3*u*u*t*curve.Control1.Y + 3*u*t*t*curve.Control2.Y + t*t*t*curve.End.Y,
			})
		}
		current = curve.End
	}
	return points
}
