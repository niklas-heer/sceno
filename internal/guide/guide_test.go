package guide

import (
	"encoding/json"
	"testing"
)

func TestGuideJSON(t *testing.T) {
	d := Build()
	if d.Tool != "sceno" || d.Version == "" || len(d.Workflow) < 3 {
		t.Fatalf("workflow: %+v version=%q", d.Workflow, d.Version)
	}
	if len(d.Shapes) < 10 || len(d.Icons) < 5 {
		t.Fatalf("catalog empty")
	}
	if d.SpecMinimal == "" || d.ErrorCodes["missing_node"].Fix == "" {
		t.Fatal("missing content")
	}
	if d.DescribeOutput["slides[n].engine.scene_stack.planes.*[].writable_bounds"] == "" {
		t.Fatal("missing agent-readable writable geometry")
	}
	data, err := json.Marshal(d)
	if err != nil || !json.Valid(data) {
		t.Fatal("invalid json")
	}
}
