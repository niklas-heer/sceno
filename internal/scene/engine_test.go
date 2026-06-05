package scene

import (
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
)

func TestBuildStackPlanes(t *testing.T) {
	d := &model.Diagram{
		Title: "Test",
		Gap:   24,
		Nodes: []model.Node{
			{ID: "lane", Kind: model.ShapeLane, Rect: model.Rect{X: 0, Y: 0, W: 400, H: 200}},
			{ID: "a", Kind: model.ShapeBox, Rect: model.Rect{X: 40, Y: 60, W: 80, H: 50}},
			{ID: "tip", Kind: model.ShapeInfobox, Rect: model.Rect{X: 200, Y: 20, W: 120, H: 40}},
		},
		Routed: []model.RoutedEdge{
			{Key: "a-tip-0", Edge: model.Edge{From: "a", To: "tip", Label: "note"}, Points: [][]float64{{120, 85}, {200, 85}}},
		},
	}
	stack := BuildStack(d)
	if len(stack.Planes["lanes"]) != 1 {
		t.Fatalf("expected lane plane item")
	}
	if len(stack.Planes["nodes"]) != 1 {
		t.Fatalf("expected node plane item")
	}
	if len(stack.Planes["annotations"]) != 1 {
		t.Fatalf("expected annotation plane item")
	}
	if len(stack.Planes["edges"]) != 1 {
		t.Fatalf("expected edge plane item")
	}
	if len(stack.Planes["labels"]) != 1 {
		t.Fatalf("expected label plane item")
	}
	if len(stack.Planes["chrome"]) != 1 {
		t.Fatalf("expected chrome plane item")
	}
}

func TestRunEngineHowItWorks(t *testing.T) {
	d := &model.Diagram{
		Title:    "How Sceno Works",
		Subtitle: "pipeline",
		Gap:      24,
		Nodes: []model.Node{
			{ID: "a", Row: 0, Column: 0, Rect: model.Rect{X: 24, Y: 100, W: 88, H: 57}},
			{ID: "b", Row: 0, Column: 1, Rect: model.Rect{X: 160, Y: 100, W: 100, H: 57}},
			{ID: "c", Row: 0, Column: 2, Rect: model.Rect{X: 310, Y: 100, W: 110, H: 57}},
		},
		Routed: []model.RoutedEdge{
			{Key: "a-b", Edge: model.Edge{From: "a", To: "b", Label: "write"}, Points: [][]float64{{112, 128}, {160, 128}}},
			{Key: "b-c", Edge: model.Edge{From: "b", To: "c"}, Points: [][]float64{{260, 128}, {310, 128}}},
		},
	}
	er := RunEngine(d)
	if er.Score < 50 {
		t.Fatalf("expected reasonable score, got %d findings=%+v", er.Score, er.Findings)
	}
	if len(er.RulesRun) < 5 {
		t.Fatalf("expected multiple rules, got %v", er.RulesRun)
	}
}

func TestRunEngineTooManyNodes(t *testing.T) {
	var nodes []model.Node
	for i := 0; i < 18; i++ {
		nodes = append(nodes, model.Node{
			ID:   string(rune('a' + i)),
			Row:  i,
			Rect: model.Rect{X: 10, Y: float64(i * 40), W: 80, H: 30},
		})
	}
	d := &model.Diagram{Gap: 24, Nodes: nodes}
	er := RunEngine(d)
	found := false
	for _, f := range er.Findings {
		if f.RuleID == "element_budget" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected element_budget finding, got %+v", er.Findings)
	}
}
