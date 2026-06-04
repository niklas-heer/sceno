package render

import (
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/theme"
)

// paint is the active palette for the current render call (CLI is single-threaded).
var paint theme.Palette

func useDiagramPalette(d model.Diagram) {
	paint = theme.Resolve(d.Theme)
}

func useDeckPalette(deck model.Deck) {
	cfg := deck.Theme
	if cfg.Mode == "" && !cfg.Transparent && len(cfg.Vars) == 0 && len(deck.Slides) > 0 {
		cfg = deck.Slides[0].Theme
	}
	paint = theme.Resolve(cfg)
}
