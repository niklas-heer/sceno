package render

import "github.com/niklas-heer/sceno/internal/model"

// SlideAspect presets (declarative slide= in KDL).
const (
	Aspect16x9 = "16x9"
	Aspect4x3  = "4x3"
)

// SlideSize returns pixel dimensions for an aspect preset.
func SlideSize(aspect string) (w, h float64) {
	switch aspect {
	case Aspect4x3:
		return 1600, 1200
	default:
		return 1920, 1080
	}
}

// SlideFrame is a 16:9 (or 4:3) canvas with the diagram scaled and centered.
type SlideFrame struct {
	Width, Height float64
	OffsetX       float64
	OffsetY       float64
	Scale         float64
	Content       Viewport
}

// SlideFrameFrom fits diagram content inside a slide aspect box.
func SlideFrameFrom(d model.Diagram, aspect string) SlideFrame {
	sw, sh := SlideSize(aspect)
	content := ViewportFrom(d)
	pad := 72.0
	if d.Title != "" {
		pad = 96
	}
	availW := sw - pad*2
	availH := sh - pad*2
	sx := availW / content.Width
	sy := availH / content.Height
	scale := sx
	if sy < sx {
		scale = sy
	}
	if scale > 1.2 {
		scale = 1.2
	}
	if scale <= 0 {
		scale = 1
	}
	dw := content.Width * scale
	dh := content.Height * scale
	return SlideFrame{
		Width:   sw,
		Height:  sh,
		OffsetX: (sw - dw) / 2,
		OffsetY: (sh - dh) / 2,
		Scale:   scale,
		Content: content,
	}
}
