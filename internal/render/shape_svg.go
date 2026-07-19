package render

import (
	"fmt"
	"math"
	"strings"

	"github.com/niklas-heer/sceno/internal/geom"
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/theme"
)

// shapeSVG returns the main shape markup for a node (no label/icon).
func shapeSVG(n model.Node, dropShadow bool) string {
	fill := n.Fill
	if fill == "" {
		fill = defaultFill(n.Kind)
	}
	stroke := n.Stroke
	if stroke == "" {
		stroke = paint.Border
	}
	k := model.NormalizeShape(n.Kind)
	r := n.Rect

	switch k {
	case model.ShapeActor:
		if n.Icon != "" {
			return actorIconBackdropSVG(n, dropShadow)
		}
		return actorSVG(n, fill, stroke)
	case model.ShapeEllipse, model.ShapeCircle:
		return fmt.Sprintf(`<ellipse cx="%.1f" cy="%.1f" rx="%.1f" ry="%.1f" fill="%s" stroke="%s" stroke-width="1.5"/>`,
			r.CX(), r.CY(), r.W/2, r.H/2, fill, stroke)
	case model.ShapeDiamond, model.ShapeDecision:
		cx, cy := r.CX(), r.CY()
		return fmt.Sprintf(`<polygon points="%.1f,%.1f %.1f,%.1f %.1f,%.1f %.1f,%.1f" fill="%s" stroke="%s" stroke-width="1.5"/>`,
			cx, r.Y, r.Right(), cy, cx, r.Bottom(), r.X, cy, fill, stroke)
	case model.ShapeHexagon:
		return polygonSVG(hexagonPoints(r), fill, stroke, 1.5)
	case model.ShapeOctagon:
		return polygonSVG(octagonPoints(r), fill, stroke, 1.5)
	case model.ShapeTriangle:
		return polygonSVG(trianglePoints(r, "up"), fill, stroke, 1.5)
	case model.ShapeParallelogram:
		return polygonSVG(parallelogramPoints(r), fill, stroke, 1.5)
	case model.ShapeCylinder, model.ShapeDatabase:
		return cylinderSVG(r, fill, stroke)
	case model.ShapeCloud:
		return cloudSVG(r, fill, stroke)
	case model.ShapeDocument:
		return documentSVG(r, fill, stroke)
	case model.ShapePill, model.ShapeTerminal, model.ShapeStart, model.ShapeEnd:
		return fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s" stroke="%s" stroke-width="1.5" rx="%.1f"/>`,
			r.X, r.Y, r.W, r.H, fill, stroke, r.H/2)
	case model.ShapeTextbox:
		return fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s" stroke="%s" stroke-width="1" rx="6"/>`, r.X, r.Y, r.W, r.H, fill, stroke)
	case model.ShapeNote:
		return fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s" stroke="%s" stroke-width="1" rx="4"/><polygon points="%.1f,%.1f %.1f,%.1f %.1f,%.1f" fill="%s"/>`,
			r.X, r.Y, r.W, r.H, fill, stroke,
			r.Right()-18, r.Bottom(), r.Right(), r.Bottom()-18, r.Right(), r.Bottom(), fill)
	case model.ShapeInfobox, model.ShapeCallout:
		accent := n.Accent
		if accent == "" {
			accent = paint.AccentBrand
		}
		s := fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s" stroke="%s" stroke-width="1" rx="10"/>`, r.X, r.Y, r.W, r.H, fill, stroke)
		s += fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="4" height="%.1f" fill="%s" rx="2"/>`, r.X, r.Y, r.H, accent)
		return s
	case model.ShapeLane, model.ShapeContainer:
		return fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s" stroke="%s" stroke-width="1" stroke-dasharray="6 4" rx="12" opacity="0.95"/>`, r.X, r.Y, r.W, r.H, paint.BgLane, stroke)
	case model.ShapeFrame, model.ShapeGroup:
		return fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s" stroke="%s" stroke-width="1.5" rx="14" opacity="0.92"/>`, r.X, r.Y, r.W, r.H, paint.BgLane, stroke)
	default:
		filter := ""
		if dropShadow {
			filter = ` filter="url(#shadow)"`
		}
		return fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s" stroke="%s" stroke-width="1.5" rx="%d"%s/>`, r.X, r.Y, r.W, r.H, fill, stroke, theme.RadiusCard, filter)
	}
}

func defaultFill(k model.ShapeKind) string {
	switch model.NormalizeShape(k) {
	case model.ShapeNote:
		return "#fef9c3"
	case model.ShapeTextbox:
		return paint.BgTextbox
	default:
		return paint.BgCard
	}
}

func polygonSVG(points string, fill, stroke string, sw float64) string {
	return fmt.Sprintf(`<polygon points="%s" fill="%s" stroke="%s" stroke-width="%.1f"/>`, points, fill, stroke, sw)
}

func hexagonPoints(r model.Rect) string {
	cx, cy := r.CX(), r.CY()
	rx, ry := r.W/2, r.H/2
	var pts []string
	for i := 0; i < 6; i++ {
		a := math.Pi/6 + float64(i)*math.Pi/3
		pts = append(pts, fmt.Sprintf("%.1f,%.1f", cx+rx*math.Cos(a), cy+ry*math.Sin(a)))
	}
	return stringsJoin(pts, " ")
}

func octagonPoints(r model.Rect) string {
	cx, cy := r.CX(), r.CY()
	rx, ry := r.W/2, r.H/2
	var pts []string
	for i := 0; i < 8; i++ {
		a := math.Pi/8 + float64(i)*math.Pi/4
		pts = append(pts, fmt.Sprintf("%.1f,%.1f", cx+rx*math.Cos(a), cy+ry*math.Sin(a)))
	}
	return stringsJoin(pts, " ")
}

func trianglePoints(r model.Rect, dir string) string {
	switch dir {
	case "down":
		return fmt.Sprintf("%.1f,%.1f %.1f,%.1f %.1f,%.1f", r.CX(), r.Bottom(), r.X, r.Y, r.Right(), r.Y)
	default:
		return fmt.Sprintf("%.1f,%.1f %.1f,%.1f %.1f,%.1f", r.CX(), r.Y, r.X, r.Bottom(), r.Right(), r.Bottom())
	}
}

func parallelogramPoints(r model.Rect) string {
	s := r.W * 0.15
	return fmt.Sprintf("%.1f,%.1f %.1f,%.1f %.1f,%.1f %.1f,%.1f",
		r.X+s, r.Y, r.Right(), r.Y, r.Right()-s, r.Bottom(), r.X, r.Bottom())
}

// CylinderRimRY is the rim ellipse radius for a cylinder of the given size.
// Shared by all backends so the silhouette matches across formats.
func CylinderRimRY(w, h float64) float64 {
	return geom.CylinderRimRY(w, h)
}

func cylinderSVG(r model.Rect, fill, stroke string) string {
	ry := CylinderRimRY(r.W, r.H)
	top := r.Y + ry
	bot := r.Bottom() - ry
	rx := r.W / 2
	// Body: sides + bottom bulge; the top seam stays unstroked so no line
	// crosses the interior — the rim ellipse closes the outline.
	body := fmt.Sprintf(`<path d="M %.1f %.1f L %.1f %.1f A %.1f %.1f 0 0 0 %.1f %.1f L %.1f %.1f Z" fill="%s" stroke="none"/>`,
		r.X, top, r.X, bot, rx, ry, r.Right(), bot, r.Right(), top, fill)
	outline := fmt.Sprintf(`<path d="M %.1f %.1f L %.1f %.1f A %.1f %.1f 0 0 0 %.1f %.1f L %.1f %.1f" fill="none" stroke="%s" stroke-width="1.5"/>`,
		r.X, top, r.X, bot, rx, ry, r.Right(), bot, r.Right(), top, stroke)
	rim := fmt.Sprintf(`<ellipse cx="%.1f" cy="%.1f" rx="%.1f" ry="%.1f" fill="%s" stroke="%s" stroke-width="1.5"/>`,
		r.CX(), top, rx, ry, fill, stroke)
	return body + outline + rim
}

func cloudSVG(r model.Rect, fill, stroke string) string {
	start, curves := cloudPath(r)
	var d strings.Builder
	fmt.Fprintf(&d, "M %.1f %.1f", start[0], start[1])
	for _, c := range curves {
		fmt.Fprintf(&d, " C %.1f %.1f %.1f %.1f %.1f %.1f", c[0], c[1], c[2], c[3], c[4], c[5])
	}
	return fmt.Sprintf(`<path d="%s Z" fill="%s" stroke="%s" stroke-width="1.5"/>`, d.String(), fill, stroke)
}

// cloudPath is the shared cloud silhouette for SVG, raster, and PDF. The
// closed sequence passes through every bbox side midpoint, so border anchors
// land on visible paint, while the alternating lobes keep a cloud silhouette
// without overlapping interior strokes.
func cloudPath(r model.Rect) ([2]float64, [][6]float64) {
	start, segments := geom.CloudPath(r)
	curves := make([][6]float64, len(segments))
	for i, segment := range segments {
		curves[i] = [6]float64{segment.Control1.X, segment.Control1.Y, segment.Control2.X, segment.Control2.Y, segment.End.X, segment.End.Y}
	}
	return [2]float64{start.X, start.Y}, curves
}

func documentSVG(r model.Rect, fill, stroke string) string {
	fold := math.Min(r.W*0.22, 22.0)
	body := fmt.Sprintf(`<path d="M %.1f %.1f L %.1f %.1f L %.1f %.1f L %.1f %.1f L %.1f %.1f Z" fill="%s" stroke="%s" stroke-width="1.5"/>`,
		r.X, r.Y, r.Right()-fold, r.Y, r.Right(), r.Y+fold, r.Right(), r.Bottom(), r.X, r.Bottom(), fill, stroke)
	corner := fmt.Sprintf(`<path d="M %.1f %.1f L %.1f %.1f L %.1f %.1f Z" fill="%s" stroke="%s" stroke-width="1.5"/>`,
		r.Right()-fold, r.Y, r.Right(), r.Y+fold, r.Right()-fold, r.Y+fold, paint.BgTextbox, stroke)
	return body + corner
}

// actorIconBackdropSVG is a soft card behind a centered icon (actor nodes with icon=).
func actorIconBackdropSVG(n model.Node, dropShadow bool) string {
	fill := n.Fill
	if fill == "" {
		fill = paint.BgCard
	}
	stroke := n.Stroke
	if stroke == "" {
		stroke = paint.Border
	}
	filter := ""
	if dropShadow {
		filter = ` filter="url(#shadow)"`
	}
	r := n.Rect
	return fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s" stroke="%s" stroke-width="1.5" rx="14"%s/>`,
		r.X, r.Y, r.W, r.H, fill, stroke, filter)
}

// actorSVG draws a UML-style stick figure inside the node bounds.
func actorSVG(n model.Node, fill, stroke string) string {
	g := geom.ActorFigure(n.Rect, n.Label != "" || n.Subtitle != "", 1)
	sw := 1.5
	body := fmt.Sprintf(
		`<circle cx="%.1f" cy="%.1f" r="%.1f" fill="%s" stroke="%s" stroke-width="%.1f"/>`+
			`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="%.1f" stroke-linecap="round"/>`+
			`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="%.1f" stroke-linecap="round"/>`+
			`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="%.1f" stroke-linecap="round"/>`+
			`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="%.1f" stroke-linecap="round"/>`,
		g.CenterX, g.HeadY, g.HeadRadius, fill, stroke, sw,
		g.CenterX, g.ShoulderY, g.CenterX, g.FootY, stroke, sw,
		g.ArmLeft, g.ArmY, g.ArmRight, g.ArmY, stroke, sw,
		g.CenterX, g.FootY, g.LegLeft, g.FootY, stroke, sw,
		g.CenterX, g.FootY, g.LegRight, g.FootY, stroke, sw,
	)
	return body
}

func stringsJoin(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	s := parts[0]
	for i := 1; i < len(parts); i++ {
		s += sep + parts[i]
	}
	return s
}
