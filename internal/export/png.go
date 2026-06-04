package export

import (
	"bytes"
	"image/png"
	"math"

	"fmt"
	"strings"

	"github.com/niklas-heer/sceno/internal/fonts"
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/render"

	"github.com/fogleman/gg"
)

// RenderPNG draws the diagram. Polished mode rasterizes SVG for pixel parity.
func RenderPNG(d model.Diagram, style RenderStyle, scale float64) ([]byte, error) {
	if scale <= 0 {
		scale = 2
	}
	if style == StylePolished {
		svg := render.PolishedSVGWithOptions(d, render.SVGOptions{DropShadow: false})
		if pngData, err := RasterizeSVG(svg, scale); err == nil && len(pngData) > 500 {
			if withIcons, err := overlayIcons(pngData, d, scale); err == nil {
				return withIcons, nil
			}
			return pngData, nil
		}
	}
	return renderPNGVector(d, style, scale)
}

func renderPNGVector(d model.Diagram, style RenderStyle, scale float64) ([]byte, error) {
	vp := render.ViewportFrom(d)
	w, h := vp.PixelSize(scale)
	dc := gg.NewContext(w, h)
	dc.SetRGB(1, 1, 1)
	dc.Clear()

	if style == StylePolished {
		drawPolishedVector(dc, d, vp, scale)
	} else {
		drawSketchVector(dc, d, vp, scale)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func drawPolishedVector(dc *gg.Context, d model.Diagram, vp render.Viewport, scale float64) {
	ox, oy := 0.0, 0.0
	render.DrawPolishedGG(dc, d, ox, oy, scale, vp)
}

func drawSketchVector(dc *gg.Context, d model.Diagram, vp render.Viewport, scale float64) {
	loadSketchFont(dc, 22*scale)
	if d.Title != "" {
		tx, ty := vp.PX(vp.MinX+28, vp.MinY+28, scale)
		dc.SetRGB(0.44, 0.28, 0.91)
		dc.DrawString(d.Title, tx, ty)
	}
	for _, n := range d.Nodes {
		if n.Kind == model.ShapeLane {
			drawSketchNode(dc, n, vp, scale)
		}
	}
	for _, re := range d.Routed {
		drawSketchEdge(dc, re.Points, re.Edge, vp, scale)
	}
	for _, n := range d.Nodes {
		if n.Kind != model.ShapeLane {
			drawSketchNode(dc, n, vp, scale)
		}
	}
}

func loadSketchFont(dc *gg.Context, size float64) {
	render.LoadFont(dc, fonts.WeightRegular, size)
}

func drawSketchNode(dc *gg.Context, n model.Node, vp render.Viewport, scale float64) {
	x, y := vp.PX(n.Rect.X, n.Rect.Y, scale)
	w := n.Rect.W * scale
	h := n.Rect.H * scale
	fr, fg, fb := hexRGB(n.Fill, 0.91, 0.93, 0.94)
	sr, sg, sb := hexRGB(n.Stroke, 0.12, 0.12, 0.12)
	dc.SetRGB(fr, fg, fb)
	dc.SetLineWidth(2 * scale)
	if n.Kind == model.ShapeEllipse {
		dc.DrawEllipse(x+w/2, y+h/2, w/2, h/2)
	} else {
		dc.DrawRoundedRectangle(x, y, w, h, 8*scale)
	}
	dc.FillPreserve()
	dc.SetRGB(sr, sg, sb)
	dc.Stroke()
	loadSketchFont(dc, n.FontSize*0.85*scale)
	dc.SetRGB(0.06, 0.09, 0.16)
	tw, _ := dc.MeasureString(n.Label)
	dc.DrawString(n.Label, x+w/2-tw/2, y+h/2)
}

func drawSketchEdge(dc *gg.Context, pts [][]float64, e model.Edge, vp render.Viewport, scale float64) {
	if len(pts) < 2 {
		return
	}
	r, g, b := hexRGB(e.Color, 0.12, 0.12, 0.12)
	dc.SetRGB(r, g, b)
	dc.SetLineWidth(2 * scale)
	for i := 1; i < len(pts); i++ {
		if len(pts[i]) < 2 || len(pts[i-1]) < 2 {
			continue
		}
		x1, y1 := vp.PX(pts[i-1][0], pts[i-1][1], scale)
		x2, y2 := vp.PX(pts[i][0], pts[i][1], scale)
		wobbleLine(dc, x1, y1, x2, y2)
	}
	dc.Stroke()
	if len(pts) >= 2 {
		drawSketchArrow(dc, pts[len(pts)-2], pts[len(pts)-1], vp, scale)
	}
}

func drawSketchArrow(dc *gg.Context, prev, last []float64, vp render.Viewport, scale float64) {
	if len(prev) < 2 || len(last) < 2 {
		return
	}
	x1, y1 := vp.PX(prev[0], prev[1], scale)
	x2, y2 := vp.PX(last[0], last[1], scale)
	angle := math.Atan2(y2-y1, x2-x1)
	sz := 8 * scale
	dc.MoveTo(x2, y2)
	dc.LineTo(x2-math.Cos(angle-0.4)*sz, y2-math.Sin(angle-0.4)*sz)
	dc.LineTo(x2-math.Cos(angle+0.4)*sz, y2-math.Sin(angle+0.4)*sz)
	dc.ClosePath()
	dc.Fill()
}

func wobbleLine(dc *gg.Context, x1, y1, x2, y2 float64) {
	const steps = 6
	for i := 0; i <= steps; i++ {
		t := float64(i) / steps
		x := x1 + (x2-x1)*t + math.Sin(t*7)*2
		y := y1 + (y2-y1)*t + math.Cos(t*5)*2
		if i == 0 {
			dc.MoveTo(x, y)
		} else {
			dc.LineTo(x, y)
		}
	}
}

func hexRGB(hex string, dr, dg, db float64) (float64, float64, float64) {
	ri, gi, bi := hexRGBInt(hex, int(dr*255), int(dg*255), int(db*255))
	return float64(ri) / 255, float64(gi) / 255, float64(bi) / 255
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
