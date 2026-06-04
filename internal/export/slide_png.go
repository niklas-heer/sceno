package export

import (
	"os"
	"path/filepath"

	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/render"
)

// WriteSlidePNG renders a 16:9 slide at native resolution (1920×1080 by default).
func WriteSlidePNG(d model.Diagram, path string, opt Options) error {
	svg := render.PolishedSVGSlide(d)
	pngData, err := RasterizeSVG(svg, 1)
	if err != nil {
		return Write(d, path, FormatPNG, opt)
	}
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, pngData, 0o644)
}
