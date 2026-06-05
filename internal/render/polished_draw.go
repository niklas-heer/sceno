package render

import (
	"fmt"
	"math"
	"strings"

	"github.com/niklas-heer/sceno/internal/fonts"
	"github.com/niklas-heer/sceno/internal/geom"
	"github.com/niklas-heer/sceno/internal/icons"
	"github.com/niklas-heer/sceno/internal/measure"
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/theme"

	"github.com/fogleman/gg"
	"github.com/jung-kurt/gofpdf"
)

const polishedIconSize = 18

// DrawPolishedGG renders the polished scene (fallback when SVG raster fails).
func DrawPolishedGG(dc *gg.Context, d model.Diagram, ox, oy, scale float64, vp Viewport) {
	useDiagramPalette(d)
	w := vp.Width * scale
	h := vp.Height * scale
	dc.SetRGB(0.98, 0.98, 0.98)
	dc.DrawRoundedRectangle(ox, oy, w, h, 12*scale)
	dc.Fill()

	if d.Title != "" {
		tx, ty := vp.PX(vp.MinX+28, vp.MinY+32, scale)
		setGGFont(dc, fonts.WeightBold, theme.TitleSize*scale)
		setGGColor(dc, paint.FgPrimary)
		dc.DrawString(d.Title, tx+ox, ty+oy)
	}
	if d.Subtitle != "" {
		tx, ty := vp.PX(vp.MinX+28, vp.MinY+56, scale)
		setGGFont(dc, fonts.WeightRegular, theme.SubtitleSize*scale)
		setGGColor(dc, paint.FgMuted)
		dc.DrawString(d.Subtitle, tx+ox, ty+oy)
	}

	for _, n := range d.Nodes {
		if n.Kind == model.ShapeLane {
			drawPolishedNodeGG(dc, n, vp, ox, oy, scale)
		}
	}
	for _, re := range d.Routed {
		drawPolishedEdgeGG(dc, re.Points, re.Edge, vp, ox, oy, scale)
		drawEdgeLabelGG(dc, re.Points, re.Edge, vp, ox, oy, scale)
	}
	for _, n := range d.Nodes {
		if n.Kind != model.ShapeLane {
			drawPolishedNodeGG(dc, n, vp, ox, oy, scale)
		}
	}
}

// DrawPolishedPDF renders the polished scene to gofpdf (PDF export).
func DrawPolishedPDF(pdf *gofpdf.Fpdf, d model.Diagram, minX, minY float64) {
	registerPDFFonts(pdf)
	w, h := pdf.GetPageSize()
	pdf.SetFillColor(250, 250, 250)
	pdf.Rect(0, 0, w, h, "F")

	if d.Title != "" {
		setPDFFont(pdf, "B", theme.TitleSize)
		pdf.SetTextColor(15, 23, 42)
		pdf.Text(minX+28, minY+32, d.Title)
	}
	if d.Subtitle != "" {
		setPDFFont(pdf, "", theme.SubtitleSize)
		pdf.SetTextColor(100, 116, 139)
		pdf.Text(minX+28, minY+56, d.Subtitle)
	}

	for _, n := range d.Nodes {
		if n.Kind == model.ShapeLane {
			drawPolishedNodePDF(pdf, n, minX, minY)
		}
	}
	for _, re := range d.Routed {
		drawPolishedEdgePDF(pdf, re.Points, re.Edge, minX, minY)
		drawEdgeLabelPDF(pdf, re.Points, re.Edge, minX, minY)
	}
	for _, n := range d.Nodes {
		if n.Kind != model.ShapeLane {
			drawPolishedNodePDF(pdf, n, minX, minY)
		}
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
		dc.MoveTo(cx, y)
		dc.LineTo(x+w, cy)
		dc.LineTo(cx, y+h)
		dc.LineTo(x, cy)
		dc.ClosePath()
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
	case model.ShapeLane:
		lr, lg, lb := hexRGB(paint.BgLane, 0.97, 0.98, 0.99)
		dc.SetRGB(lr, lg, lb)
		dc.SetDash(6*scale, 4*scale)
		dc.DrawRoundedRectangle(x, y, w, h, 12*scale)
		dc.FillPreserve()
		dc.SetRGB(sr, sg, sb)
		dc.Stroke()
		dc.SetDash()
	default:
		dc.DrawRoundedRectangle(x, y, w, h, 12*scale)
		dc.FillPreserve()
		dc.SetRGB(sr, sg, sb)
		dc.Stroke()
	}

	if n.Icon != "" {
		ix, iy := IconRect(n, polishedIconSize)
		px, py := vp.PX(ix, iy, scale)
		icons.Draw(dc, n.Icon, px+ox, py+oy, polishedIconSize*scale, paint.FgMuted)
	}

	drawPolishedLabelGG(dc, n, x, y, w, h, scale)
	if model.IsContainer(n.Kind) && n.Label != "" {
		setGGFont(dc, fonts.WeightSemiBold, theme.LaneLabelSize*scale)
		setGGColor(dc, paint.FgMuted)
		dc.DrawString(n.Label, x+14*scale, y+14*scale)
	}
}

func drawPolishedLabelGG(dc *gg.Context, n model.Node, x, y, w, h, scale float64) {
	if n.Label == "" && n.Subtitle == "" {
		return
	}
	fs := n.FontSize
	if fs <= 0 {
		fs = theme.NodeSize
	}
	setGGFont(dc, fonts.WeightMedium, fs*scale)
	setGGColor(dc, paint.FgPrimary)
	lines := strings.Split(n.Label, "\n")
	lh := fs * 1.25 * scale
	iconOff := 0.0
	if n.Icon != "" {
		iconOff = 14 * scale
	}
	totalH := float64(len(lines))*lh + iconOff
	if n.Subtitle != "" {
		totalH += 14 * scale
	}
	startY := y + h/2 - totalH/2 + lh*0.75 + iconOff/2
	contentW := w - measure.PadX * scale
	if n.Icon != "" {
		contentW -= measure.IconColumn * scale
	}
	for i, line := range lines {
		tw, _ := dc.MeasureString(line)
		tx := x + (w-tw)/2
		if n.Icon != "" {
			tx = x + measure.IconColumn*scale + (contentW-tw)/2
		}
		dc.DrawString(line, tx, startY+float64(i)*lh)
	}
	if n.Subtitle != "" {
		setGGFont(dc, fonts.WeightRegular, theme.SubSize*scale)
		setGGColor(dc, paint.FgMuted)
		dc.DrawString(n.Subtitle, x+14*scale, y+h-16*scale)
	}
}

func drawPolishedEdgeGG(dc *gg.Context, pts [][]float64, e model.Edge, vp Viewport, ox, oy, scale float64) {
	if len(pts) < 2 {
		return
	}
	gpts := geom.SlicesToPath(pts)
	gpts = geom.SimplifyPath(gpts)
	gpts = geom.TrimArrowEnd(gpts)
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
	for i := 1; i < len(gpts); i++ {
		x1, y1 := vp.PX(gpts[i-1].X, gpts[i-1].Y, scale)
		x2, y2 := vp.PX(gpts[i].X, gpts[i].Y, scale)
		dc.DrawLine(x1+ox, y1+oy, x2+ox, y2+oy)
	}
	dc.Stroke()
	dc.SetDash()
	if len(gpts) >= 2 {
		drawArrowGG(dc, gpts[len(gpts)-2], gpts[len(gpts)-1], vp, ox, oy, scale, stroke)
	}
}

func drawArrowGG(dc *gg.Context, prev, last geom.Point, vp Viewport, ox, oy, scale float64, stroke string) {
	x1, y1 := vp.PX(prev.X, prev.Y, scale)
	x2, y2 := vp.PX(last.X, last.Y, scale)
	x1 += ox
	y1 += oy
	x2 += ox
	y2 += oy
	angle := math.Atan2(y2-y1, x2-x1)
	sz := 7 * scale
	dc.MoveTo(x2, y2)
	dc.LineTo(x2-math.Cos(angle-0.4)*sz, y2-math.Sin(angle-0.4)*sz)
	dc.LineTo(x2-math.Cos(angle+0.4)*sz, y2-math.Sin(angle+0.4)*sz)
	dc.ClosePath()
	setGGColor(dc, stroke)
	dc.Fill()
}

func drawPolishedNodePDF(pdf *gofpdf.Fpdf, n model.Node, minX, minY float64) {
	x := n.Rect.X - minX
	y := n.Rect.Y - minY
	w := n.Rect.W
	h := n.Rect.H
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

	switch n.Kind {
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
	case model.ShapeLane:
		lr, lg, lb := hexRGBInt(paint.BgLane, 248, 250, 252)
		pdf.SetFillColor(lr, lg, lb)
		pdf.SetDashPattern([]float64{4, 3}, 0)
		pdf.RoundedRect(x, y, w, h, 6, "1234", "FD")
		pdf.SetDashPattern(nil, 0)
	default:
		pdf.RoundedRect(x, y, w, h, 6, "1234", "FD")
	}

	drawPolishedLabelPDF(pdf, n, x, y, w, h)
	if model.IsContainer(n.Kind) && n.Label != "" {
		setPDFFont(pdf, "SB", theme.LaneLabelSize)
		pdf.SetTextColor(100, 116, 139)
		pdf.Text(x+14, y+14, n.Label)
	}
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
	pdf.SetTextColor(15, 23, 42)
	lines := strings.Split(n.Label, "\n")
	lineH := fs * 1.25
	startY := y + h/2 - float64(len(lines))*lineH/2 + lineH*0.7
	for i, line := range lines {
		tw := pdf.GetStringWidth(line)
		pdf.Text(x+w/2-tw/2, startY+float64(i)*lineH, line)
	}
	if n.Subtitle != "" {
		setPDFFont(pdf, "", theme.SubSize)
		pdf.SetTextColor(100, 116, 139)
		pdf.Text(x+14, y+h-12, n.Subtitle)
	}
}

func drawPolishedEdgePDF(pdf *gofpdf.Fpdf, pts [][]float64, e model.Edge, minX, minY float64) {
	if len(pts) < 2 {
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
	for i := 1; i < len(pts); i++ {
		if len(pts[i]) < 2 || len(pts[i-1]) < 2 {
			continue
		}
		x1 := pts[i-1][0] - minX
		y1 := pts[i-1][1] - minY
		x2 := pts[i][0] - minX
		y2 := pts[i][1] - minY
		pdf.Line(x1, y1, x2, y2)
	}
	pdf.SetDashPattern(nil, 0)
	// Arrow head
	if len(pts) >= 2 {
		prev, last := pts[len(pts)-2], pts[len(pts)-1]
		if len(prev) >= 2 && len(last) >= 2 {
			x1 := prev[0] - minX
			y1 := prev[1] - minY
			x2 := last[0] - minX
			y2 := last[1] - minY
			angle := math.Atan2(y2-y1, x2-x1)
			sz := 7.0
			pdf.SetFillColor(r, g, b)
			pdf.Polygon([]gofpdf.PointType{
				{X: x2, Y: y2},
				{X: x2 - math.Cos(angle-0.4)*sz, Y: y2 - math.Sin(angle-0.4)*sz},
				{X: x2 - math.Cos(angle+0.4)*sz, Y: y2 - math.Sin(angle+0.4)*sz},
			}, "F")
		}
	}
}

func drawEdgeLabelGG(dc *gg.Context, pts [][]float64, e model.Edge, vp Viewport, ox, oy, scale float64) {
	label := strings.TrimSpace(e.Label)
	if label == "" || len(pts) < 2 {
		return
	}
	gpts := geom.SimplifyPath(geom.SlicesToPath(pts))
	lines := strings.Split(label, "\n")
	fontSize := theme.SubSize * scale
	lineH := fontSize * 1.35
	maxW := 0.0
	for _, line := range lines {
		setGGFont(dc, fonts.WeightMedium, fontSize)
		lineW, _ := dc.MeasureString(line)
		if lineW > maxW {
			maxW = lineW
		}
	}
	rx, ry, boxW, boxH, _ := geom.EdgeLabelBox(gpts, 6*scale, 4*scale, lineH, fontSize, lines, maxW/scale)
	px, py := vp.PX(rx, ry, scale)
	px += ox
	py += oy
	boxW *= scale
	boxH *= scale
	dc.SetRGB(1, 1, 1)
	dc.DrawRoundedRectangle(px-boxW/2, py-boxH/2, boxW, boxH, 4*scale)
	dc.FillPreserve()
	setGGColor(dc, paint.Border)
	dc.SetLineWidth(1 * scale)
	dc.Stroke()
	setGGFont(dc, fonts.WeightMedium, fontSize)
	setGGColor(dc, paint.FgMuted)
	textY := py - boxH/2 + 4*scale + fontSize*0.85
	for i, line := range lines {
		lineW, _ := dc.MeasureString(line)
		dc.DrawString(line, px-lineW/2, textY+float64(i)*lineH)
	}
}

func drawEdgeLabelPDF(pdf *gofpdf.Fpdf, pts [][]float64, e model.Edge, minX, minY float64) {
	label := strings.TrimSpace(e.Label)
	if label == "" || len(pts) < 2 {
		return
	}
	gpts := geom.SimplifyPath(geom.SlicesToPath(pts))
	lines := strings.Split(label, "\n")
	fontSize := float64(theme.SubSize)
	lineH := fontSize * 1.35
	maxW := 0.0
	for _, line := range lines {
		w := measure.TextWidth(line, fontSize, fonts.WeightMedium)
		if w > maxW {
			maxW = w
		}
	}
	rx, ry, boxW, boxH, _ := geom.EdgeLabelBox(gpts, 6, 4, lineH, fontSize, lines, maxW)
	setPDFFont(pdf, "M", theme.SubSize)
	pdf.SetTextColor(113, 113, 122)
	textY := ry - boxH/2 + 4 + fontSize*0.85
	for i, line := range lines {
		tw := pdf.GetStringWidth(line)
		pdf.Text(rx-minX-tw/2, textY-minY+float64(i)*lineH, line)
	}
	_ = boxW
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
