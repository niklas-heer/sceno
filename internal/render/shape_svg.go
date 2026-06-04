package render

import (
	"fmt"
	"math"

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
		return actorSVG(r, fill, stroke)
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

func cylinderSVG(r model.Rect, fill, stroke string) string {
	ry := math.Min(r.W*0.12, 14.0)
	return fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s" stroke="%s" stroke-width="1.5"/>`+
		`<ellipse cx="%.1f" cy="%.1f" rx="%.1f" ry="%.1f" fill="%s" stroke="%s" stroke-width="1.5"/>`+
		`<ellipse cx="%.1f" cy="%.1f" rx="%.1f" ry="%.1f" fill="%s" stroke="%s" stroke-width="1.5"/>`,
		r.X, r.Y+ry, r.W, r.H-ry*2, fill, stroke,
		r.CX(), r.Y+ry, r.W/2, ry, fill, stroke,
		r.CX(), r.Bottom()-ry, r.W/2, ry, fill, stroke)
}

func cloudSVG(r model.Rect, fill, stroke string) string {
	cx, cy := r.CX(), r.CY()
	rx, ry := r.W/2*0.95, r.H/2*0.85
	return fmt.Sprintf(`<ellipse cx="%.1f" cy="%.1f" rx="%.1f" ry="%.1f" fill="%s" stroke="%s" stroke-width="1.5"/>`+
		`<ellipse cx="%.1f" cy="%.1f" rx="%.1f" ry="%.1f" fill="%s" stroke="%s" stroke-width="1.5"/>`+
		`<ellipse cx="%.1f" cy="%.1f" rx="%.1f" ry="%.1f" fill="%s" stroke="%s" stroke-width="1.5"/>`,
		cx-rx*0.35, cy, rx*0.55, ry, fill, stroke,
		cx+rx*0.3, cy-ry*0.1, rx*0.5, ry*0.9, fill, stroke,
		cx+rx*0.1, cy+ry*0.15, rx*0.65, ry*0.85, fill, stroke)
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
func actorSVG(r model.Rect, fill, stroke string) string {
	cx := r.CX()
	headR := math.Min(r.W*0.16, r.H*0.14)
	if headR < 8 {
		headR = 8
	}
	headCY := r.Y + headR + 6
	shoulderY := headCY + headR + 4
	footY := r.Bottom() - 6
	arm := r.W * 0.32
	leg := r.W * 0.22
	sw := 1.5
	body := fmt.Sprintf(
		`<circle cx="%.1f" cy="%.1f" r="%.1f" fill="%s" stroke="%s" stroke-width="%.1f"/>`+
			`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="%.1f" stroke-linecap="round"/>`+
			`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="%.1f" stroke-linecap="round"/>`+
			`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="%.1f" stroke-linecap="round"/>`+
			`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="%.1f" stroke-linecap="round"/>`,
		cx, headCY, headR, fill, stroke, sw,
		cx, shoulderY, cx, footY, stroke, sw,
		cx-arm, shoulderY+4, cx+arm, shoulderY+4, stroke, sw,
		cx, footY, cx-leg, footY, stroke, sw,
		cx, footY, cx+leg, footY, stroke, sw,
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
