package render

import (
	"strings"
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
)

func TestPathSketchDrawsSharedRouteWithoutSecondSmoothing(t *testing.T) {
	path := [][]float64{{296, 160}, {223, 220}, {196, 220}}
	svg := pathSketch(path, model.Edge{From: "refine", To: "parking", Dashed: true})
	if strings.Contains(svg, " Q ") || strings.Contains(svg, " T ") {
		t.Fatalf("sketch renderer transformed shared route again: %s", svg)
	}
	if !strings.Contains(svg, `d="M 296.0 160.0 L 223.0 220.0 L 205.0 220.0"`) {
		t.Fatalf("stroke does not preserve routed points and 9px head clearance: %s", svg)
	}
}
