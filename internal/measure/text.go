package measure

import "github.com/niklas-heer/sceno/internal/model"

// Size estimates node dimensions from label (font metrics); boxes fit text.
func Size(n model.NodeSpec) (w, h float64) {
	return FitSize(n)
}
