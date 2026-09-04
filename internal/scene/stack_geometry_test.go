package scene

import (
	"reflect"
	"testing"

	"github.com/niklas-heer/sceno/internal/measure"
	"github.com/niklas-heer/sceno/internal/model"
)

func TestProjectUsesPaintThenSourceOrder(t *testing.T) {
	stack := BuildStack(&model.Diagram{Nodes: []model.Node{
		{ID: "node2", Kind: model.ShapeBox},
		{ID: "note", Kind: model.ShapeNote},
		{ID: "node1", Kind: model.ShapeBox},
	}})
	for i := 0; i < 50; i++ {
		var ids []string
		for _, item := range stack.Project(PlaneNode, PlaneAnnotation, PlaneNode) {
			ids = append(ids, item.ID)
		}
		if !reflect.DeepEqual(ids, []string{"note", "node2", "node1"}) {
			t.Fatalf("unstable projection or source order: %v", ids)
		}
	}
}

func TestStackExposesSilhouetteWritableBoundsAndEffectiveFont(t *testing.T) {
	d := &model.Diagram{Nodes: []model.Node{{
		ID: "decision", Kind: model.ShapeDiamond, Label: "Choose a deployment path", FontSize: 18,
		Rect: model.Rect{X: 40, Y: 30, W: 180, H: 100},
	}}}
	measure.ApplyInteriors(d.Nodes)
	stack := BuildStack(d)
	items := stack.Planes[PlaneNode.String()]
	if len(items) != 1 {
		t.Fatalf("node stack = %+v", items)
	}
	item := items[0]
	if len(item.Outline) < 4 || item.Writable == nil || item.Writable.W <= 0 || item.Writable.H <= 0 {
		t.Fatalf("missing agent-readable shape geometry: %+v", item)
	}
	if item.FontSize <= 0 || item.FontSize > 18 {
		t.Fatalf("invalid effective font size %.1f", item.FontSize)
	}
}
