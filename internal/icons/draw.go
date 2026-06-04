package icons

import (
	"image"
	"strings"
	"sync"

	"github.com/fogleman/gg"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

const view = 24.0

// Group returns SVG icon as a transformed <g> (renders in browsers and rasterizers).
func Group(name string, x, y, size float64, color string) string {
	p, ok := paths[name]
	if !ok {
		return ""
	}
	p = strings.ReplaceAll(p, "currentColor", color)
	scale := size / view
	return `<g transform="translate(` + fmtFloat(x) + `,` + fmtFloat(y) + `) scale(` + fmtFloat(scale) + `)" fill="none" color="` + color + `">` + p + `</g>`
}

// SVG is an alias for Group (valid inline SVG, no nested <svg>).
func SVG(name string, x, y, size float64, color string) string {
	return Group(name, x, y, size, color)
}

var cache sync.Map // key: name|color|size -> image.Image

// Draw renders a crisp icon onto a gg context.
func Draw(dc *gg.Context, name string, x, y, size float64, color string) {
	px := int(size * 2)
	if px < 16 {
		px = 32
	}
	img := rasterIcon(name, px, color)
	if img == nil {
		return
	}
	dc.Push()
	dc.Translate(x, y)
	s := size / float64(img.Bounds().Dx())
	dc.Scale(s, s)
	dc.DrawImage(img, 0, 0)
	dc.Pop()
}

func rasterIcon(name string, px int, color string) image.Image {
	if px < 8 {
		px = 48
	}
	key := name + "|" + color + "|" + string(rune(px))
	if v, ok := cache.Load(key); ok {
		return v.(image.Image)
	}
	p, ok := paths[name]
	if !ok {
		return nil
	}
	p = strings.ReplaceAll(p, "currentColor", color)
	svg := `<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" color="` + color + `">` + p + `</svg>`
	icon, err := oksvg.ReadIconStream(strings.NewReader(svg), oksvg.WarnErrorMode)
	if err != nil {
		return nil
	}
	icon.SetTarget(0, 0, float64(px), float64(px))
	rgba := image.NewRGBA(image.Rect(0, 0, px, px))
	sc := rasterx.NewScannerGV(px, px, rgba, rgba.Bounds())
	d := rasterx.NewDasher(px, px, sc)
	icon.Draw(d, 1.0)
	cache.Store(key, rgba)
	return rgba
}
