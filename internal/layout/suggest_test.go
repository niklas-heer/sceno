package layout

import (
	"strings"
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
)

func TestCompactSuggestionIgnoresFixedAndOrdinaryPipelineColumns(t *testing.T) {
	d := &model.Diagram{Nodes: []model.Node{
		{ID: "a", Column: 0, Rect: model.Rect{W: 80, H: 40}},
		{ID: "b", Column: 1, Layer: 1, Rect: model.Rect{X: 140, W: 80, H: 40}},
		{ID: "note", Fixed: true, Column: -1, Rect: model.Rect{X: 48, Y: 160, W: 120, H: 60}},
	}}
	for _, issue := range CompactSuggestion(d) {
		if strings.Contains(issue.Message, "column") || strings.Contains(issue.Message, "many distinct layer") {
			t.Fatalf("false column advice for hybrid pipeline: %+v", issue)
		}
	}
}

func TestGridMarksFixedNodesOutsideLogicalColumns(t *testing.T) {
	d := &model.Diagram{Nodes: []model.Node{{ID: "note", Fixed: true, Rect: model.Rect{X: 280, Y: 100, W: 80, H: 40}}}}
	Grid(d, 24)
	if d.Nodes[0].Column != -1 {
		t.Fatalf("fixed node column = %d, want -1", d.Nodes[0].Column)
	}
}
