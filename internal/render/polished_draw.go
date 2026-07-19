package render

import (
	"bytes"
	"fmt"
	"math"
	"strings"

	"github.com/niklas-heer/sceno/internal/fonts"
	"github.com/niklas-heer/sceno/internal/geom"
	"github.com/niklas-heer/sceno/internal/highlight"
	"github.com/niklas-heer/sceno/internal/icons"
	"github.com/niklas-heer/sceno/internal/measure"
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/theme"

	"github.com/fogleman/gg"
	gofpdf "github.com/go-pdf/fpdf"
)

const polishedIconSize = measure.IconSize

// DrawPolishedGG renders the polished scene (fallback when SVG raster fails).
func DrawPolishedGG(dc *gg.Context, d model.Diagram, ox, oy, scale float64, vp Viewport) {
	useDiagramPalette(d)
	w := vp.Width * scale
	h := vp.Height * scale
	if paint.BgCanvas == "none" {
		dc.SetRGBA(0, 0, 0, 0)
	} else {
		setGGColor(dc, paint.BgCanvas)
	}
	dc.DrawRoundedRectangle(ox, oy, w, h, 12*scale)
	dc.Fill()

	if d.Title != "" {
		tx, ty := vp.PX(vp.MinX+theme.CanvasTextInset, vp.MinY+theme.HeaderTitleBaseline, scale)
		setGGFont(dc, fonts.WeightBold, theme.TitleSize*scale)
		setGGColor(dc, paint.FgPrimary)
		dc.DrawString(d.Title, tx+ox, ty+oy)
	}
	if d.Subtitle != "" {
		tx, ty := vp.PX(vp.MinX+theme.CanvasTextInset, vp.MinY+theme.HeaderSubtitleBaseline, scale)
		setGGFont(dc, fonts.WeightRegular, theme.SubtitleSize*scale)
		setGGColor(dc, paint.FgMuted)
		dc.DrawString(d.Subtitle, tx+ox, ty+oy)
	}

	for _, n := range nodesBeforeEdges(&d) {
		drawPolishedNodeGG(dc, n, vp, ox, oy, scale)
	}
	for _, re := range d.Routed {
		lctx := LabelContext(d, re.Edge)
		drawPolishedEdgeGG(dc, re.Points, re.Edge, lctx, vp, ox, oy, scale)
	}
	for _, n := range nodesAfterEdges(&d) {
		drawPolishedNodeGG(dc, n, vp, ox, oy, scale)
	}
	for _, re := range d.Routed {
		if strings.TrimSpace(re.Edge.Label) == "" {
			continue
		}
		lctx := LabelContext(d, re.Edge)
		drawEdgeLabelGG(dc, re.Points, re.Edge, lctx, vp, ox, oy, scale)
	}
	for _, re := range d.Routed {
		drawArrowHeadGG(dc, re.Points, re.Edge, vp, ox, oy, scale)
	}
}

// DrawPolishedPDF renders the polished scene to gofpdf (PDF export).
func DrawPolishedPDF(pdf *gofpdf.Fpdf, d model.Diagram, minX, minY float64) {
	useDiagramPalette(d)
	registerPDFFonts(pdf)
	w, h := pdf.GetPageSize()
	setPDFFillColor(pdf, paint.BgCanvas)
	pdf.Rect(0, 0, w, h, "F")

	if d.Title != "" {
		setPDFFont(pdf, "B", theme.TitleSize)
		setPDFTextColor(pdf, paint.FgPrimary)
		pdf.Text(theme.CanvasTextInset, theme.HeaderTitleBaseline, d.Title)
	}
	if d.Subtitle != "" {
		setPDFFont(pdf, "", theme.SubtitleSize)
		setPDFTextColor(pdf, paint.FgMuted)
		pdf.Text(theme.CanvasTextInset, theme.HeaderSubtitleBaseline, d.Subtitle)
	}

	for _, n := range nodesBeforeEdges(&d) {
		drawPolishedNodePDF(pdf, n, minX, minY)
	}
	for _, re := range d.Routed {
		lctx := LabelContext(d, re.Edge)
		drawPolishedEdgePDF(pdf, re.Points, re.Edge, lctx, minX, minY)
	}
	for _, n := range nodesAfterEdges(&d) {
		drawPolishedNodePDF(pdf, n, minX, minY)
	}
	for _, re := range d.Routed {
		if strings.TrimSpace(re.Edge.Label) == "" {
			continue
		}
		lctx := LabelContext(d, re.Edge)
		drawEdgeLabelPDF(pdf, re.Points, re.Edge, lctx, minX, minY)
	}
	for _, re := range d.Routed {
		drawArrowHeadPDF(pdf, re.Points, re.Edge, minX, minY)
	}
}

func drawPolishedNodeGG(dc *gg.Context, n model.Node, vp Viewport, ox, oy, scale float64) {
	x, y := vp.PX(n.Rect.X, n.Rect.Y, scale)
	x += ox
	y += oy
	w := n.Rect.W * scale
	h := n.Rect.H * scale
	fill := n.Fill
	if fill == "" {
		fill = paint.BgCard
	}
	stroke := n.Stroke
	if stroke == "" {
		stroke = paint.Border
	}
	fr, fg, fb := hexRGB(fill, 1, 1, 1)
	sr, sg, sb := hexRGB(stroke, 0.88, 0.9, 0.93)
	dc.SetRGB(fr, fg, fb)
	dc.SetLineWidth(1.5 * scale)

	switch model.NormalizeShape(n.Kind) {
	case model.ShapeActor:
		if n.Icon != "" {
			dc.DrawRoundedRectangle(x, y, w, h, 14*scale)
			dc.FillPreserve()
			dc.SetRGB(sr, sg, sb)
			dc.Stroke()
		} else {
			drawActorGG(dc, x, y, w, h, scale, sr, sg, sb)
		}
	case model.ShapeEllipse, model.ShapeCircle:
		dc.DrawEllipse(x+w/2, y+h/2, w/2, h/2)
		dc.FillPreserve()
		dc.SetRGB(sr, sg, sb)
		dc.Stroke()
	case model.ShapeDiamond:
		cx, cy := x+w/2, y+h/2
		drawPolygonGG(dc, [][2]float64{{cx, y}, {x + w, cy}, {cx, y + h}, {x, cy}}, sr, sg, sb)
	case model.ShapeHexagon:
		drawRegularPolygonGG(dc, x+w/2, y+h/2, w/2, h/2, 6, math.Pi/6, sr, sg, sb)
	case model.ShapeOctagon:
		drawRegularPolygonGG(dc, x+w/2, y+h/2, w/2, h/2, 8, math.Pi/8, sr, sg, sb)
	case model.ShapeTriangle:
		drawPolygonGG(dc, [][2]float64{{x + w/2, y}, {x + w, y + h}, {x, y + h}}, sr, sg, sb)
	case model.ShapeParallelogram:
		skew := w * 0.15
		drawPolygonGG(dc, [][2]float64{{x + skew, y}, {x + w, y}, {x + w - skew, y + h}, {x, y + h}}, sr, sg, sb)
	case model.ShapeCylinder:
		ry := CylinderRimRY(w/scale, h/scale) * scale
		cx := x + w/2
		// Body with bottom bulge; the top seam stays unstroked (rim closes it).
		dc.MoveTo(x, y+ry)
		dc.LineTo(x, y+h-ry)
		dc.DrawEllipticalArc(cx, y+h-ry, w/2, ry, math.Pi, 0)
		dc.LineTo(x+w, y+ry)
		dc.ClosePath()
		dc.Fill()
		dc.MoveTo(x, y+ry)
		dc.LineTo(x, y+h-ry)
		dc.DrawEllipticalArc(cx, y+h-ry, w/2, ry, math.Pi, 0)
		dc.LineTo(x+w, y+ry)
		dc.SetRGB(sr, sg, sb)
		dc.Stroke()
		dc.SetRGB(fr, fg, fb)
		dc.DrawEllipse(cx, y+ry, w/2, ry)
		dc.FillPreserve()
		dc.SetRGB(sr, sg, sb)
		dc.Stroke()
	case model.ShapeCloud:
		start, curves := cloudPath(model.Rect{X: x, Y: y, W: w, H: h})
		dc.MoveTo(start[0], start[1])
		for _, c := range curves {
			dc.CubicTo(c[0], c[1], c[2], c[3], c[4], c[5])
		}
		dc.ClosePath()
		dc.SetRGB(fr, fg, fb)
		dc.FillPreserve()
		dc.SetRGB(sr, sg, sb)
		dc.Stroke()
	case model.ShapeDocument:
		fold := math.Min(w*0.22, 22*scale)
		drawPolygonGG(dc, [][2]float64{{x, y}, {x + w - fold, y}, {x + w, y + fold}, {x + w, y + h}, {x, y + h}}, sr, sg, sb)
		dc.MoveTo(x+w-fold, y)
		dc.LineTo(x+w-fold, y+fold)
		dc.LineTo(x+w, y+fold)
		dc.SetRGB(sr, sg, sb)
		dc.Stroke()
	case model.ShapePill:
		dc.DrawRoundedRectangle(x, y, w, h, h/2)
		dc.FillPreserve()
		dc.SetRGB(sr, sg, sb)
		dc.Stroke()
	case model.ShapeTextbox:
		dc.DrawRoundedRectangle(x, y, w, h, 6*scale)
		dc.FillPreserve()
		dc.SetRGB(sr, sg, sb)
		dc.Stroke()
	case model.ShapeInfobox:
		dc.DrawRoundedRectangle(x, y, w, h, 10*scale)
		dc.FillPreserve()
		dc.SetRGB(sr, sg, sb)
		dc.Stroke()
		accent := n.Accent
		if accent == "" {
			accent = theme.Accent
		}
		ar, ag, ab := hexRGB(accent, 0.49, 0.23, 0.93)
		dc.SetRGB(ar, ag, ab)
		dc.DrawRectangle(x, y, 4*scale, h)
		dc.Fill()
	case model.ShapeNote:
		dc.DrawRoundedRectangle(x, y, w, h, 4*scale)
		dc.FillPreserve()
		dc.SetRGB(sr, sg, sb)
		dc.Stroke()
		fold := math.Min(18*scale, math.Min(w, h)*0.28)
		dc.MoveTo(x+w-fold, y+h)
		dc.LineTo(x+w, y+h-fold)
		dc.LineTo(x+w, y+h)
		dc.ClosePath()
		dc.SetRGB(fr, fg, fb)
		dc.FillPreserve()
		dc.SetRGB(sr, sg, sb)
		dc.Stroke()
	case model.ShapeLane:
		lr, lg, lb := hexRGB(paint.BgLane, 0.97, 0.98, 0.99)
		dc.SetRGB(lr, lg, lb)
		dc.SetDash(6*scale, 4*scale)
		dc.DrawRoundedRectangle(x, y, w, h, 12*scale)
		dc.FillPreserve()
		dc.SetRGB(sr, sg, sb)
		dc.Stroke()
		dc.SetDash()
	case model.ShapeFrame, model.ShapeGroup:
		lr, lg, lb := hexRGB(paint.BgLane, 0.97, 0.98, 0.99)
		dc.SetRGB(lr, lg, lb)
		dc.DrawRoundedRectangle(x, y, w, h, 14*scale)
		dc.FillPreserve()
		dc.SetRGB(sr, sg, sb)
		dc.Stroke()
	default:
		dc.DrawRoundedRectangle(x, y, w, h, 12*scale)
		dc.FillPreserve()
		dc.SetRGB(sr, sg, sb)
		dc.Stroke()
	}

	if model.IsContainer(n.Kind) {
		if n.Label != "" {
			setGGFont(dc, fonts.WeightSemiBold, theme.LaneLabelSize*scale)
			setGGColor(dc, paint.FgMuted)
			dc.DrawString(n.Label, x+14*scale, y+14*scale)
		}
		return
	}
	if n.Icon != "" {
		ix, iy := IconRect(n, polishedIconSize)
		px, py := vp.PX(ix, iy, scale)
		icons.Draw(dc, n.Icon, px+ox, py+oy, polishedIconSize*scale, paint.FgSecondary)
	}
	drawPolishedLabelGG(dc, n, x, y, w, h, scale)
}

func drawPolygonGG(dc *gg.Context, points [][2]float64, sr, sg, sb float64) {
	for i, point := range points {
		if i == 0 {
			dc.MoveTo(point[0], point[1])
		} else {
			dc.LineTo(point[0], point[1])
		}
	}
	dc.ClosePath()
	dc.FillPreserve()
	dc.SetRGB(sr, sg, sb)
	dc.Stroke()
}

func drawRegularPolygonGG(dc *gg.Context, cx, cy, rx, ry float64, sides int, offset, sr, sg, sb float64) {
	points := make([][2]float64, sides)
	for i := range points {
		angle := offset + float64(i)*2*math.Pi/float64(sides)
		points[i] = [2]float64{cx + rx*math.Cos(angle), cy + ry*math.Sin(angle)}
	}
	drawPolygonGG(dc, points, sr, sg, sb)
}

func drawPolishedLabelGG(dc *gg.Context, n model.Node, x, y, w, h, scale float64) {
	if n.Label == "" && n.Subtitle == "" {
		return
	}
	fs := n.FontSize
	if fs <= 0 {
		fs = theme.NodeSize
	}
	cl := measure.LayoutFor(n)
	setGGFont(dc, fonts.WeightMedium, fs*scale)
	setGGColor(dc, paint.FgPrimary)
	lines := strings.Split(n.Label, "\n")
	lh := cl.TitleLineH * scale
	for i, line := range lines {
		tw, _ := dc.MeasureString(line)
		tx := x + cl.TitleX*scale
		if cl.TopAlign {
			tx = x + (w-tw)/2
		} else if cl.InlineIcon {
			tx = x + cl.TitleX*scale
		} else {
			tx = x + (w-tw)/2
		}
		dc.DrawString(line, tx, y+cl.TitleStartY*scale+float64(i)*lh)
	}
	if cl.HasSubtitle {
		setGGFont(dc, fonts.WeightRegular, theme.SubSize*scale)
		setGGColor(dc, paint.FgMuted)
		sw, _ := dc.MeasureString(n.Subtitle)
		sx := x + (w-sw)/2
		if cl.InlineIcon {
			sx = x + cl.TitleX*scale
		}
		dc.DrawString(n.Subtitle, sx, y+cl.SubtitleY*scale)
	}
}

func drawPolishedEdgeGG(dc *gg.Context, pts [][]float64, e model.Edge, lctx *geom.EdgeLabelContext, vp Viewport, ox, oy, scale float64) {
	segments := edgePathSegments(pts, e, lctx)
	if len(segments) == 0 {
		return
	}
	stroke := e.Color
	if stroke == "" {
		stroke = paint.EdgeDefault
	}
	r, g, b := hexRGB(stroke, 0.39, 0.45, 0.55)
	dc.SetRGB(r, g, b)
	dc.SetLineWidth(theme.EdgeWidth * scale)
	dc.SetLineCap(gg.LineCapRound)
	dc.SetLineJoin(gg.LineJoinRound)
	if e.Dashed {
		dc.SetDash(5*scale, 4*scale)
	}
	gpts := geom.SimplifyPath(segments[0])
	gpts = geom.TrimArrowEnd(gpts)
	for i := 1; i < len(gpts); i++ {
		x1, y1 := vp.PX(gpts[i-1].X, gpts[i-1].Y, scale)
		x2, y2 := vp.PX(gpts[i].X, gpts[i].Y, scale)
		dc.DrawLine(x1+ox, y1+oy, x2+ox, y2+oy)
	}
	dc.Stroke()
	dc.SetDash()
}

func drawArrowHeadGG(dc *gg.Context, pts [][]float64, e model.Edge, vp Viewport, ox, oy, scale float64) {
	gpts := geom.SimplifyPath(geom.SlicesToPath(pts))
	ag, ok := geom.ArrowGeometryForPath(gpts)
	if !ok {
		return
	}
	stroke := e.Color
	if stroke == "" {
		stroke = paint.FgMuted
	}
	t1, t2, t3 := geom.ArrowHeadPoints(ag.Prev, ag.Tip, theme.ArrowMarkerSize)
	for i, p := range []geom.Point{t1, t2, t3} {
		x, y := vp.PX(p.X, p.Y, scale)
		if i == 0 {
			dc.MoveTo(x+ox, y+oy)
		} else {
			dc.LineTo(x+ox, y+oy)
		}
	}
	dc.ClosePath()
	setGGColor(dc, stroke)
	dc.Fill()
}

func drawPolishedNodePDF(pdf *gofpdf.Fpdf, n model.Node, minX, minY float64) {
	x := n.Rect.X - minX
	y := n.Rect.Y - minY
	w := n.Rect.W
	h := n.Rect.H
	if model.NormalizeShape(n.Kind) == model.ShapeCode {
		drawCodeBlockPDF(pdf, n, x, y, w, h)
		return
	}
	fill := n.Fill
	if fill == "" {
		fill = paint.BgCard
	}
	stroke := n.Stroke
	if stroke == "" {
		stroke = paint.Border
	}
	fr, fg, fb := hexRGBInt(fill, 255, 255, 255)
	sr, sg, sb := hexRGBInt(stroke, 226, 232, 240)
	pdf.SetFillColor(fr, fg, fb)
	pdf.SetDrawColor(sr, sg, sb)
	pdf.SetLineWidth(0.75)

	switch model.NormalizeShape(n.Kind) {
	case model.ShapeActor:
		if n.Icon != "" {
			pdf.RoundedRect(x, y, w, h, 7, "1234", "FD")
		} else {
			drawActorPDF(pdf, x, y, w, h, sr, sg, sb)
		}
	case model.ShapeEllipse, model.ShapeCircle:
		pdf.Ellipse(x+w/2, y+h/2, w/2, h/2, 0, "FD")
	case model.ShapeDiamond:
		cx, cy := x+w/2, y+h/2
		pts := []gofpdf.PointType{
			{X: cx, Y: y},
			{X: x + w, Y: cy},
			{X: cx, Y: y + h},
			{X: x, Y: cy},
		}
		pdf.Polygon(pts, "FD")
	case model.ShapeHexagon:
		pdf.Polygon(regularPolygonPDF(x+w/2, y+h/2, w/2, h/2, 6, math.Pi/6), "FD")
	case model.ShapeOctagon:
		pdf.Polygon(regularPolygonPDF(x+w/2, y+h/2, w/2, h/2, 8, math.Pi/8), "FD")
	case model.ShapeTriangle:
		pdf.Polygon([]gofpdf.PointType{{X: x + w/2, Y: y}, {X: x + w, Y: y + h}, {X: x, Y: y + h}}, "FD")
	case model.ShapeParallelogram:
		skew := w * 0.15
		pdf.Polygon([]gofpdf.PointType{{X: x + skew, Y: y}, {X: x + w, Y: y}, {X: x + w - skew, Y: y + h}, {X: x, Y: y + h}}, "FD")
	case model.ShapeCylinder:
		ry := CylinderRimRY(w, h)
		// Body fill without stroking seams; sides and bottom bulge stroked
		// separately so no line crosses the interior (parity with SVG).
		pdf.Rect(x, y+ry, w, h-2*ry, "F")
		pdf.Arc(x+w/2, y+h-ry, w/2, ry, 0, 180, 360, "FD")
		pdf.Line(x, y+ry, x, y+h-ry)
		pdf.Line(x+w, y+ry, x+w, y+h-ry)
		pdf.Ellipse(x+w/2, y+ry, w/2, ry, 0, "FD")
	case model.ShapeCloud:
		start, curves := cloudPath(model.Rect{X: x, Y: y, W: w, H: h})
		pdf.MoveTo(start[0], start[1])
		for _, c := range curves {
			pdf.CurveBezierCubicTo(c[0], c[1], c[2], c[3], c[4], c[5])
		}
		pdf.ClosePath()
		pdf.DrawPath("FD")
	case model.ShapeDocument:
		fold := math.Min(w*0.22, 22)
		pdf.Polygon([]gofpdf.PointType{{X: x, Y: y}, {X: x + w - fold, Y: y}, {X: x + w, Y: y + fold}, {X: x + w, Y: y + h}, {X: x, Y: y + h}}, "FD")
		pdf.Line(x+w-fold, y, x+w-fold, y+fold)
		pdf.Line(x+w-fold, y+fold, x+w, y+fold)
	case model.ShapePill:
		pdf.RoundedRect(x, y, w, h, h/2, "1234", "FD")
	case model.ShapeTextbox:
		pdf.RoundedRect(x, y, w, h, 3, "1234", "FD")
	case model.ShapeInfobox:
		pdf.RoundedRect(x, y, w, h, 5, "1234", "FD")
		accent := n.Accent
		if accent == "" {
			accent = theme.Accent
		}
		ar, ag, ab := hexRGBInt(accent, 124, 58, 237)
		pdf.SetFillColor(ar, ag, ab)
		pdf.Rect(x, y, 1.5, h, "F")
		pdf.SetFillColor(fr, fg, fb)
	case model.ShapeNote:
		pdf.RoundedRect(x, y, w, h, 2, "1234", "FD")
		fold := math.Min(18.0, math.Min(w, h)*0.28)
		pdf.Polygon([]gofpdf.PointType{{X: x + w - fold, Y: y + h}, {X: x + w, Y: y + h - fold}, {X: x + w, Y: y + h}}, "FD")
	case model.ShapeLane:
		lr, lg, lb := hexRGBInt(paint.BgLane, 248, 250, 252)
		pdf.SetFillColor(lr, lg, lb)
		pdf.SetDashPattern([]float64{4, 3}, 0)
		pdf.RoundedRect(x, y, w, h, 6, "1234", "FD")
		pdf.SetDashPattern(nil, 0)
	default:
		pdf.RoundedRect(x, y, w, h, 6, "1234", "FD")
	}

	if model.IsContainer(n.Kind) {
		if n.Label != "" {
			setPDFFont(pdf, "SB", theme.LaneLabelSize)
			setPDFTextColor(pdf, paint.FgMuted)
			pdf.Text(x+14, y+14, n.Label)
		}
		return
	}
	if n.Icon != "" {
		ix, iy := IconRect(n, polishedIconSize)
		data, err := icons.PNG(n.Icon, 64, paint.FgSecondary)
		if err == nil {
			key := "sceno-icon-" + n.Icon + "-" + strings.TrimPrefix(paint.FgSecondary, "#")
			opt := gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}
			pdf.RegisterImageOptionsReader(key, opt, bytes.NewReader(data))
			pdf.ImageOptions(key, ix-minX, iy-minY, polishedIconSize, polishedIconSize, false, opt, 0, "")
		}
	}
	drawPolishedLabelPDF(pdf, n, x, y, w, h)
}

func regularPolygonPDF(cx, cy, rx, ry float64, sides int, offset float64) []gofpdf.PointType {
	points := make([]gofpdf.PointType, sides)
	for i := range points {
		angle := offset + float64(i)*2*math.Pi/float64(sides)
		points[i] = gofpdf.PointType{X: cx + rx*math.Cos(angle), Y: cy + ry*math.Sin(angle)}
	}
	return points
}

func drawActorPDF(pdf *gofpdf.Fpdf, x, y, w, h float64, sr, sg, sb int) {
	cx := x + w/2
	headR := math.Max(4, math.Min(w*0.16, h*0.14))
	headCY := y + headR + 3
	shoulderY := headCY + headR + 2
	footY := y + h - 3
	pdf.SetDrawColor(sr, sg, sb)
	pdf.Ellipse(cx, headCY, headR, headR, 0, "D")
	pdf.Line(cx, shoulderY, cx, footY)
	pdf.Line(cx-w*0.32, shoulderY+2, cx+w*0.32, shoulderY+2)
	pdf.Line(cx, footY, cx-w*0.22, footY)
	pdf.Line(cx, footY, cx+w*0.22, footY)
}

func drawPolishedLabelPDF(pdf *gofpdf.Fpdf, n model.Node, x, y, w, h float64) {
	if n.Label == "" && n.Subtitle == "" {
		return
	}
	fs := n.FontSize
	if fs <= 0 {
		fs = theme.NodeSize
	}
	setPDFFont(pdf, "M", fs)
	setPDFTextColor(pdf, paint.FgPrimary)
	lines := strings.Split(n.Label, "\n")
	cl := measure.LayoutFor(n)
	lineH := cl.TitleLineH
	for i, line := range lines {
		tw := pdf.GetStringWidth(line)
		tx := x + (w-tw)/2
		if cl.InlineIcon {
			tx = x + cl.TitleX
		}
		pdf.Text(tx, y+cl.TitleStartY+float64(i)*lineH, line)
	}
	if cl.HasSubtitle {
		setPDFFont(pdf, "", theme.SubSize)
		setPDFTextColor(pdf, paint.FgMuted)
		tw := pdf.GetStringWidth(n.Subtitle)
		tx := x + (w-tw)/2
		if cl.InlineIcon {
			tx = x + cl.TitleX
		}
		pdf.Text(tx, y+cl.SubtitleY, n.Subtitle)
	}
}

func drawCodeBlockPDF(pdf *gofpdf.Fpdf, n model.Node, x, y, w, h float64) {
	fill := n.Fill
	if fill == "" {
		fill = paint.BgCode
	}
	stroke := n.Stroke
	if stroke == "" {
		stroke = paint.Border
	}
	setPDFFillColor(pdf, fill)
	setPDFDrawColor(pdf, stroke)
	pdf.RoundedRect(x, y, w, h, 5, "1234", "FD")

	body := n.Code
	if body == "" {
		body = n.Label
	}
	lang := n.CodeLang
	if lang == "" {
		lang = "text"
	}
	lineY := y + codePadY + codeFontSize
	if n.Label != "" && n.Label != body {
		setPDFFont(pdf, "SB", codeFontSize)
		setPDFTextColor(pdf, paint.FgMuted)
		pdf.Text(x+codePadX, y+14, n.Label)
		lineY += 18
	}
	setPDFFont(pdf, "", codeFontSize)
	for _, line := range strings.Split(body, "\n") {
		line = strings.ReplaceAll(line, "\t", "    ")
		lineX := x + codePadX
		for _, span := range highlight.Tokenize(lang, line) {
			setPDFTextColor(pdf, codeColor(span.Kind))
			pdf.Text(lineX, lineY, span.Text)
			lineX += pdf.GetStringWidth(span.Text)
		}
		lineY += codeLineH
		if lineY > y+h-codePadY {
			break
		}
	}
}

func drawPolishedEdgePDF(pdf *gofpdf.Fpdf, pts [][]float64, e model.Edge, lctx *geom.EdgeLabelContext, minX, minY float64) {
	segments := edgePathSegments(pts, e, lctx)
	if len(segments) == 0 {
		return
	}
	stroke := e.Color
	if stroke == "" {
		stroke = paint.EdgeDefault
	}
	r, g, b := hexRGBInt(stroke, 100, 116, 139)
	pdf.SetDrawColor(r, g, b)
	if e.Dashed {
		pdf.SetDashPattern([]float64{4, 3}, 0)
	} else {
		pdf.SetDashPattern(nil, 0)
	}
	gpts := geom.SimplifyPath(segments[0])
	gpts = geom.TrimArrowEnd(gpts)
	for i := 1; i < len(gpts); i++ {
		x1 := gpts[i-1].X - minX
		y1 := gpts[i-1].Y - minY
		x2 := gpts[i].X - minX
		y2 := gpts[i].Y - minY
		pdf.Line(x1, y1, x2, y2)
	}
	pdf.SetDashPattern(nil, 0)
}

func drawArrowHeadPDF(pdf *gofpdf.Fpdf, pts [][]float64, e model.Edge, minX, minY float64) {
	gpts := geom.SimplifyPath(geom.SlicesToPath(pts))
	ag, ok := geom.ArrowGeometryForPath(gpts)
	if !ok {
		return
	}
	stroke := e.Color
	if stroke == "" {
		stroke = paint.FgMuted
	}
	r, g, b := hexRGBInt(stroke, 113, 113, 122)
	pdf.SetFillColor(r, g, b)
	t1, t2, t3 := geom.ArrowHeadPoints(ag.Prev, ag.Tip, theme.ArrowMarkerSize)
	pdf.Polygon([]gofpdf.PointType{
		{X: t1.X - minX, Y: t1.Y - minY},
		{X: t2.X - minX, Y: t2.Y - minY},
		{X: t3.X - minX, Y: t3.Y - minY},
	}, "F")
}

func drawEdgeLabelGG(dc *gg.Context, pts [][]float64, e model.Edge, lctx *geom.EdgeLabelContext, vp Viewport, ox, oy, scale float64) {
	gpts := geom.SimplifyPath(geom.SlicesToPath(pts))
	if len(gpts) < 2 {
		return
	}
	layout := geom.LayoutEdgeLabel(gpts, e.Label, lctx)
	if layout.BoxW <= 0 {
		return
	}
	px, py := vp.PX(layout.CenterX, layout.CenterY, scale)
	px += ox
	py += oy
	boxW := layout.BoxW * scale
	boxH := layout.BoxH * scale
	fontSize := layout.FontSize * scale
	lineH := layout.LineH * scale
	dc.SetRGB(1, 1, 1)
	dc.DrawRoundedRectangle(px-boxW/2, py-boxH/2, boxW, boxH, 4*scale)
	dc.FillPreserve()
	setGGColor(dc, paint.Border)
	dc.SetLineWidth(1 * scale)
	dc.Stroke()
	setGGFont(dc, fonts.WeightMedium, fontSize)
	setGGColor(dc, paint.FgMuted)
	for i, line := range layout.Lines {
		lineW, _ := dc.MeasureString(line)
		_, ty := vp.PX(layout.CenterX, layout.TextBaselineY(i), scale)
		dc.DrawString(line, px-lineW/2, ty+oy)
		_ = lineH
	}
}

func drawEdgeLabelPDF(pdf *gofpdf.Fpdf, pts [][]float64, e model.Edge, lctx *geom.EdgeLabelContext, minX, minY float64) {
	gpts := geom.SimplifyPath(geom.SlicesToPath(pts))
	if len(gpts) < 2 {
		return
	}
	layout := geom.LayoutEdgeLabel(gpts, e.Label, lctx)
	if layout.BoxW <= 0 {
		return
	}
	setPDFFont(pdf, "M", theme.SubSize)
	setPDFTextColor(pdf, paint.FgMuted)
	for i, line := range layout.Lines {
		tw := pdf.GetStringWidth(line)
		pdf.Text(layout.CenterX-minX-tw/2, layout.TextBaselineY(i)-minY, line)
	}
}

func setPDFFillColor(pdf *gofpdf.Fpdf, color string) {
	if color == "none" || color == "" {
		color = "#ffffff"
	}
	r, g, b := hexRGBInt(color, 255, 255, 255)
	pdf.SetFillColor(r, g, b)
}

func setPDFDrawColor(pdf *gofpdf.Fpdf, color string) {
	r, g, b := hexRGBInt(color, 226, 232, 240)
	pdf.SetDrawColor(r, g, b)
}

func setPDFTextColor(pdf *gofpdf.Fpdf, color string) {
	r, g, b := hexRGBInt(color, 15, 23, 42)
	pdf.SetTextColor(r, g, b)
}

func registerPDFFonts(pdf *gofpdf.Fpdf) {
	family := fonts.Family()
	for _, pair := range []struct {
		style string
		data  []byte
	}{
		{"", fonts.RegularBytes()},
		{"M", fonts.MediumBytes()},
		{"SB", fonts.SemiBoldBytes()},
		{"B", fonts.BoldBytes()},
	} {
		pdf.AddUTF8FontFromBytes(family, pair.style, pair.data)
	}
}

func setPDFFont(pdf *gofpdf.Fpdf, style string, size float64) {
	pdf.SetFont(fonts.Family(), style, size)
}

// LoadFont sets an embedded Inter face on a gg context.
func LoadFont(dc *gg.Context, w fonts.Weight, size float64) {
	setGGFont(dc, w, size)
}

func setGGFont(dc *gg.Context, w fonts.Weight, size float64) {
	face, err := fonts.Face(w, size)
	if err == nil {
		dc.SetFontFace(face)
	}
}

func setGGColor(dc *gg.Context, hex string) {
	r, g, b := hexRGB(hex, 0, 0, 0)
	dc.SetRGB(r, g, b)
}

func hexRGB(hex string, dr, dg, db float64) (float64, float64, float64) {
	ri, gi, bi := hexRGBInt(hex, int(dr*255), int(dg*255), int(db*255))
	return float64(ri) / 255, float64(gi) / 255, float64(bi) / 255
}

func drawActorGG(dc *gg.Context, x, y, w, h, scale, sr, sg, sb float64) {
	cx := x + w/2
	headR := math.Min(w*0.16, h*0.14)
	if headR < 8*scale {
		headR = 8 * scale
	}
	headCY := y + headR + 6*scale
	shoulderY := headCY + headR + 4*scale
	footY := y + h - 6*scale
	arm := w * 0.32
	leg := w * 0.22
	dc.SetRGB(sr, sg, sb)
	dc.DrawCircle(cx, headCY, headR)
	dc.Stroke()
	dc.DrawLine(cx, shoulderY, cx, footY)
	dc.Stroke()
	dc.DrawLine(cx-arm, shoulderY+4*scale, cx+arm, shoulderY+4*scale)
	dc.Stroke()
	dc.DrawLine(cx, footY, cx-leg, footY)
	dc.Stroke()
	dc.DrawLine(cx, footY, cx+leg, footY)
	dc.Stroke()
}

func hexRGBInt(hex string, defR, defG, defB int) (int, int, int) {
	hex = strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if len(hex) != 6 {
		return defR, defG, defB
	}
	var r, g, b int
	if _, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b); err != nil {
		return defR, defG, defB
	}
	return r, g, b
}
