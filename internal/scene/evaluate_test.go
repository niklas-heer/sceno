package scene

import (
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
)

func TestEvaluateMatchesRunEngine(t *testing.T) {
	d := &model.Diagram{
		Title: "Test",
		Gap:   32,
		Nodes: []model.Node{
			{ID: "a", Kind: model.ShapeBox, Rect: model.Rect{X: 0, Y: 0, W: 80, H: 50}},
			{ID: "b", Kind: model.ShapeBox, Rect: model.Rect{X: 120, Y: 0, W: 80, H: 50}},
		},
		Routed: []model.RoutedEdge{
			{Key: "a-b", Edge: model.Edge{From: "a", To: "b"}, Points: [][]float64{{80, 25}, {120, 25}}},
		},
	}
	ev := Evaluate(d)
	er := RunEngine(d)
	if ev.Score != er.Score {
		t.Fatalf("score mismatch: evaluate=%d runEngine=%d", ev.Score, er.Score)
	}
	if len(ev.Findings) != len(er.Findings) {
		t.Fatalf("findings count: evaluate=%d runEngine=%d", len(ev.Findings), len(er.Findings))
	}
	if len(ev.PaintOrder) == 0 {
		t.Fatal("expected paint order from evaluation")
	}
	if ev.Stack.Counts["nodes"] < 1 {
		t.Fatalf("expected node plane count, stack=%+v", ev.Stack)
	}
}

func TestMergeEvaluationsMinScore(t *testing.T) {
	a := Evaluation{Score: 90, Summary: "slide a"}
	b := Evaluation{Score: 70, Summary: "slide b", Findings: []Finding{{RuleID: "dense", Severity: "warning", Code: "dense_layout", Message: "crowded"}}}
	merged := MergeEvaluations([]Evaluation{a, b})
	if merged.Score != 70 {
		t.Fatalf("expected min score 70, got %d", merged.Score)
	}
	if len(merged.Findings) != 1 {
		t.Fatalf("expected merged findings, got %+v", merged.Findings)
	}
}

func TestEvaluateNilDiagram(t *testing.T) {
	ev := Evaluate(nil)
	if ev.Score != 0 || ev.Summary == "" {
		t.Fatalf("nil diagram: %+v", ev)
	}
}
