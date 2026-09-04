package scene

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/niklas-heer/sceno/internal/collision"
	"github.com/niklas-heer/sceno/internal/model"
)

func TestSpacingSignedGapsAndCollisionClearance(t *testing.T) {
	base := model.Rect{X: 10, Y: 20, W: 100, H: 80}
	tests := []struct {
		name               string
		other              model.Rect
		x, y, distance     float64
		violates, overlaps bool
	}{
		{"horizontal threshold", model.Rect{X: 120, Y: 20, W: 100, H: 80}, 10, -80, 10, false, false},
		{"horizontal too close", model.Rect{X: 119, Y: 20, W: 100, H: 80}, 9, -80, 9, true, false},
		{"touching", model.Rect{X: 110, Y: 20, W: 100, H: 80}, 0, -80, 0, true, false},
		{"overlap", model.Rect{X: 90, Y: 70, W: 100, H: 80}, -20, -30, 0, true, true},
		{"containment", model.Rect{X: 40, Y: 40, W: 20, H: 20}, -20, -20, 0, true, true},
		{"diagonal distance is not required clearance", model.Rect{X: 118, Y: 108, W: 100, H: 80}, 8, 8, math.Sqrt(128), true, false},
		{"above left", model.Rect{X: -100, Y: -80, W: 100, H: 80}, 10, 20, math.Sqrt(500), false, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := &model.Diagram{Gap: 20, Nodes: []model.Node{
				{ID: "a", Kind: model.ShapeBox, Rect: base},
				{ID: "b", Kind: model.ShapeBox, Rect: test.other},
			}}
			report := AnalyzeSpacing(d, BuildStack(d))
			if len(report.Pairs) != 1 {
				t.Fatalf("pairs: %+v", report)
			}
			p := report.Pairs[0]
			if p.GapX != test.x || p.GapY != test.y || math.Abs(p.Distance-test.distance) > 1e-9 {
				t.Fatalf("incorrect gap/distance: %+v", p)
			}
			if p.ViolatesClearance != test.violates || (p.Overlap != nil) != test.overlaps {
				t.Fatalf("incorrect collision state: %+v", p)
			}
			if p.ViolatesClearance != (len(collision.Find(d.Nodes, report.RequiredClearance)) > 0) {
				t.Fatalf("spacing disagrees with collision engine: %+v", p)
			}
		})
	}
}

func TestSpacingIntentionalOverlapAndContainerRelationships(t *testing.T) {
	d := &model.Diagram{Gap: 28, Nodes: []model.Node{
		{ID: "frame", Kind: model.ShapeFrame, Rect: model.Rect{X: 0, Y: 0, W: 300, H: 200}},
		{ID: "a", Kind: model.ShapeBox, Parent: "frame", Rect: model.Rect{X: 20, Y: 30, W: 100, H: 80}},
		{ID: "b", Kind: model.ShapeBox, Parent: "frame", AllowOverlap: true, Rect: model.Rect{X: 100, Y: 30, W: 100, H: 80}},
		// Explicit parent-child relationships are excluded even for non-container parents.
		{ID: "child", Kind: model.ShapeBox, Parent: "a", Rect: model.Rect{X: 20, Y: 30, W: 40, H: 40}},
	}}
	r := AnalyzeSpacing(d, BuildStack(d))
	if r.TotalPairs != 2 {
		t.Fatalf("container/parent pairs must be excluded: %+v", r.Pairs)
	}
	for _, p := range r.Pairs {
		if p.ViolatesClearance {
			t.Fatalf("allowed overlap treated as collision: %+v", p)
		}
		if p.A == "a" && p.B == "b" && (!p.OverlapAllowed || p.Overlap == nil || p.MeetsClearance) {
			t.Fatalf("intentional overlap geometry hidden: %+v", p)
		}
	}
	if !reflect.DeepEqual(r.Nodes[0].ParentPadding, &Insets{Top: 30, Right: 180, Bottom: 90, Left: 20}) {
		t.Fatalf("parent padding = %+v", r.Nodes[0])
	}
}

func TestSpacingFractionalThresholdMatchesCollisionArithmetic(t *testing.T) {
	a := model.Node{ID: "a", Kind: model.ShapeBox, Rect: model.Rect{X: 2267.058593810488, W: 192.45900716687657, H: 80}}
	margin := 2.526617973017191
	b := model.Node{ID: "b", Kind: model.ShapeBox, Rect: model.Rect{X: a.Rect.Right() + margin, W: 100, H: 80}}
	d := &model.Diagram{Gap: margin * 2, Nodes: []model.Node{a, b}}
	for _, horizontal := range []bool{true, false} {
		if !horizontal {
			for i := range d.Nodes {
				r := &d.Nodes[i].Rect
				r.X, r.Y, r.W, r.H = r.Y, r.X, r.H, r.W
			}
		}
		for _, direction := range []float64{0, math.Inf(-1), math.Inf(1)} {
			// Test exactly on, one representable float below, and one above
			// the clearance boundary without replacing engine math by epsilon.
			probe := *d
			probe.Nodes = append([]model.Node(nil), d.Nodes...)
			if direction != 0 {
				if horizontal {
					probe.Nodes[1].Rect.X = math.Nextafter(probe.Nodes[1].Rect.X, direction)
				} else {
					probe.Nodes[1].Rect.Y = math.Nextafter(probe.Nodes[1].Rect.Y, direction)
				}
			}
			pair := AnalyzeSpacing(&probe, BuildStack(&probe)).Pairs[0]
			collides := len(collision.Find(probe.Nodes, margin)) > 0
			if pair.ViolatesClearance != collides || pair.MeetsClearance == collides {
				t.Fatalf("horizontal=%v direction=%v: spacing %+v disagrees with collision=%v", horizontal, direction, pair, collides)
			}
			if direction == 0 && !pair.MeetsClearance {
				t.Fatalf("exact fractional clearance must pass: %+v", pair)
			}
		}
	}
}

func TestSpacingConsumesExactStackContentAndCanvas(t *testing.T) {
	// Deliberately supply stack geometry that differs from the model: reporting
	// must consume the same measured stack as the engine, not measure again.
	writable := model.Rect{X: 40, Y: 50, W: 60, H: 40}
	stack := Stack{Canvas: model.Rect{X: -10, Y: -20, W: 220, H: 180}, Planes: map[string][]StackItem{
		"nodes": {{ID: "a", Kind: "box", Plane: PlaneNode, Bounds: model.Rect{X: 20, Y: 30, W: 100, H: 80}, Writable: &writable,
			Content: []ContentItem{
				{Kind: "icon", Bounds: model.Rect{X: 38, Y: 55, W: 20, H: 20}},
				{Kind: "title", Bounds: model.Rect{X: 60, Y: 60, W: 32, H: 16}},
			}}},
	}}
	r := AnalyzeSpacing(&model.Diagram{Gap: 32}, stack)
	n := r.Nodes[0]
	if !reflect.DeepEqual(n.ContentBounds, &model.Rect{X: 38, Y: 55, W: 54, H: 21}) ||
		!reflect.DeepEqual(n.ContentPadding, &Insets{Top: 25, Right: 28, Bottom: 34, Left: 18}) ||
		!reflect.DeepEqual(n.WritablePadding, &Insets{Top: 5, Right: 8, Bottom: 14, Left: -2}) {
		t.Fatalf("content padding must use stack measurements and preserve overflow: %+v", n)
	}
	if !reflect.DeepEqual(r.Canvas.NodeMargins, &Insets{Top: 50, Right: 90, Bottom: 50, Left: 30}) {
		t.Fatalf("canvas margins = %+v", r.Canvas)
	}
}

func TestSpacingSelectionIsBoundedAndDeterministic(t *testing.T) {
	d := &model.Diagram{Gap: 20}
	for i := 0; i < 25; i++ {
		d.Nodes = append(d.Nodes, model.Node{ID: fmt.Sprintf("n%02d", i), Kind: model.ShapeBox, Rect: model.Rect{W: 100, H: 80}})
	}
	stack := BuildStack(d)
	want := AnalyzeSpacing(d, stack)
	if !want.Truncated || len(want.Pairs) != SpacingPairLimit || want.TotalPairs != 300 || want.CandidatePairs != 300 {
		t.Fatalf("incorrect bounded selection: total=%d candidates=%d pairs=%d truncated=%v", want.TotalPairs, want.CandidatePairs, len(want.Pairs), want.Truncated)
	}
	for i := 1; i < len(want.Pairs); i++ {
		if !spacingPriority(want.Pairs[i-1], want.Pairs[i]) {
			t.Fatalf("unstable order at %d", i)
		}
	}
	// Reverse stack source order; IDs break geometric ties consistently.
	items := stack.Planes["nodes"]
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
	for i := 0; i < 20; i++ {
		if got := AnalyzeSpacing(d, stack); !reflect.DeepEqual(got, want) {
			t.Fatal("spacing output depends on map/source iteration")
		}
	}
}

func TestSpacingSelectionIncludesNearestClearPeerAndPrioritizesViolations(t *testing.T) {
	d := &model.Diagram{Gap: 20}
	for i := 0; i < 25; i++ {
		d.Nodes = append(d.Nodes, model.Node{ID: fmt.Sprintf("n%02d", i), Kind: model.ShapeBox, AllowOverlap: true, Rect: model.Rect{W: 100, H: 80}})
	}
	d.Nodes = append(d.Nodes,
		model.Node{ID: "z1", Kind: model.ShapeBox, Rect: model.Rect{X: 300, W: 100, H: 80}},
		model.Node{ID: "z2", Kind: model.ShapeBox, Rect: model.Rect{X: 409, W: 100, H: 80}},
	)
	r := AnalyzeSpacing(d, BuildStack(d))
	if r.Pairs[0].A != "z1" || r.Pairs[0].B != "z2" || !r.Pairs[0].ViolatesClearance {
		t.Fatalf("true violation lost behind intentional overlap: %+v", r.Pairs[0])
	}
	d.Nodes = []model.Node{
		{ID: "a", Kind: model.ShapeBox, Rect: model.Rect{W: 100, H: 80}},
		{ID: "b", Kind: model.ShapeBox, Rect: model.Rect{X: 140, W: 100, H: 80}},
		{ID: "c", Kind: model.ShapeBox, Rect: model.Rect{X: 1000, W: 100, H: 80}},
	}
	r = AnalyzeSpacing(d, BuildStack(d))
	if len(r.Pairs) != 2 || r.TotalPairs != 3 || r.CandidatePairs != 2 || r.Truncated {
		t.Fatalf("expected deduplicated nearest pairs: %+v", r)
	}
	if r.Pairs[1].A != "b" || r.Pairs[1].B != "c" {
		t.Fatalf("isolated node lost nearest peer: %+v", r.Pairs)
	}
}

func TestSpacingEvaluationAndJSONContract(t *testing.T) {
	d := &model.Diagram{Gap: 28, Nodes: []model.Node{{ID: "one", Kind: model.ShapeBox, Label: "One", Rect: model.Rect{X: 20, Y: 20, W: 100, H: 80}}}}
	ev := Evaluate(d)
	if !reflect.DeepEqual(ev.Spacing, ev.EngineReport().Spacing) {
		t.Fatal("evaluation and engine spacing differ")
	}
	encoded, err := json.Marshal(ev.EngineReport())
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &raw); err != nil {
		t.Fatal(err)
	}
	var got SpacingReport
	if err := json.Unmarshal(raw["spacing"], &got); err != nil {
		t.Fatal(err)
	}
	if got.Units != "px" || got.RequiredClearance != 14 || len(got.Nodes) != 1 || got.Pairs == nil {
		t.Fatalf("JSON spacing contract: %+v", got)
	}
	if empty := AnalyzeSpacing(nil, Stack{}); empty.Pairs == nil || empty.Nodes == nil || len(empty.Pairs) != 0 {
		t.Fatalf("empty spacing: %+v", empty)
	}
}
