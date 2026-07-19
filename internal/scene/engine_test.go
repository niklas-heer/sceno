package scene

import (
	"testing"

	"github.com/niklas-heer/sceno/internal/diag"
	"github.com/niklas-heer/sceno/internal/model"
)

func TestBuildStackPlanes(t *testing.T) {
	d := &model.Diagram{
		Title: "Test",
		Gap:   24,
		Nodes: []model.Node{
			{ID: "lane", Kind: model.ShapeLane, Rect: model.Rect{X: 0, Y: 0, W: 400, H: 200}},
			{ID: "a", Label: "API", Icon: "api", Kind: model.ShapeBox, Rect: model.Rect{X: 40, Y: 60, W: 100, H: 80}},
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
	if content := stack.Planes["nodes"][0].Content; len(content) < 2 || content[0].Kind != "icon" || content[0].Bounds.W <= 0 {
		t.Fatalf("expected exact node content geometry, got %+v", content)
	}
	if got := stack.Planes["labels"][0].Bounds; got.W <= 0 || got.H <= 0 {
		t.Fatalf("expected exact label bounds, got %+v", got)
	}
}

func TestEvaluationExposesCompleteSceneStack(t *testing.T) {
	d := &model.Diagram{Gap: 24, Nodes: []model.Node{
		{ID: "frame", Kind: model.ShapeFrame, Rect: model.Rect{W: 300, H: 180}},
		{ID: "node", Kind: model.ShapeBox, Parent: "frame", Rect: model.Rect{X: 40, Y: 60, W: 80, H: 40}},
	}}
	ev := Evaluate(d)
	items := ev.SceneStack.Planes["nodes"]
	if len(items) != 1 || items[0].Parent != "frame" {
		t.Fatalf("scene stack lost parent containment: %+v", items)
	}
	if len(ev.SceneStack.Planes["structure"]) != 1 {
		t.Fatalf("scene stack lost structure plane: %+v", ev.SceneStack.Planes)
	}
}

func TestIntentionalOverlapRemainsVisibleWithoutOcclusionWarning(t *testing.T) {
	d := &model.Diagram{Gap: 24, Nodes: []model.Node{
		{ID: "base", Kind: model.ShapeBox, Rect: model.Rect{W: 120, H: 80}},
		{ID: "badge", Kind: model.ShapeNote, AllowOverlap: true, Rect: model.Rect{X: 80, Y: 40, W: 80, H: 50}},
	}}
	ev := Evaluate(d)
	if len(ev.Scene.Occlusions) != 0 {
		t.Fatalf("intentional overlap reported as occlusion: %+v", ev.Scene.Occlusions)
	}
	items := ev.SceneStack.Planes[PlaneAnnotation.String()]
	if len(items) != 1 || !items[0].AllowOverlap {
		t.Fatalf("scene stack lost overlap intent: %+v", items)
	}
}

func TestEdgeSideMismatchDetectsWrongArrivalDirection(t *testing.T) {
	d := &model.Diagram{Gap: 24, Nodes: []model.Node{
		{ID: "a", Kind: model.ShapeBox, Rect: model.Rect{W: 80, H: 40}},
		{ID: "b", Kind: model.ShapeBox, Rect: model.Rect{X: 200, Y: 120, W: 80, H: 40}},
	}, Routed: []model.RoutedEdge{{
		Key: "a-b-0", Edge: model.Edge{From: "a", To: "b", FromSide: model.SideRight, ToSide: model.SideLeft},
		Points: [][]float64{{80, 20}, {200, 20}, {200, 140}},
	}}}
	ev := Evaluate(d)
	for _, f := range ev.Findings {
		if f.Code == string(diag.CodeEdgeSideMismatch) && f.Severity == "warning" {
			return
		}
	}
	t.Fatalf("missing side mismatch finding: %+v", ev.Findings)
}

func TestEdgeLabelOverlapReportsExactGeometry(t *testing.T) {
	d := &model.Diagram{Gap: 24, Nodes: []model.Node{
		{ID: "a", Kind: model.ShapeBox, Rect: model.Rect{W: 80, H: 40}},
		{ID: "b", Kind: model.ShapeBox, Rect: model.Rect{X: 200, W: 80, H: 40}},
	}, Routed: []model.RoutedEdge{
		{Key: "a-b-0", Edge: model.Edge{From: "a", To: "b", Label: "forward"}, Points: [][]float64{{80, 20}, {200, 20}}},
		{Key: "b-a-1", Edge: model.Edge{From: "b", To: "a", Label: "retry"}, Points: [][]float64{{200, 20}, {80, 20}}},
	}}
	ev := Evaluate(d)
	for _, f := range ev.Findings {
		if f.Code == string(diag.CodeEdgeLabelOverlap) {
			if f.Geometry == nil || f.Geometry.Overlap.W <= 0 || f.Geometry.Overlap.H <= 0 || len(f.Geometry.Bounds) != 2 {
				t.Fatalf("missing exact label collision geometry: %+v", f.Geometry)
			}
			return
		}
	}
	t.Fatalf("missing edge label overlap finding: %+v", ev.Findings)
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
