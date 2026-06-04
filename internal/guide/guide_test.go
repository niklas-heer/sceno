package guide

import (
	"encoding/json"
	"testing"
)

func TestGuideJSON(t *testing.T) {
	d := Build()
	if d.Tool != "sceno" || len(d.Workflow) < 3 {
		t.Fatalf("workflow: %+v", d.Workflow)
	}
	if len(d.Shapes) < 10 || len(d.Icons) < 5 {
		t.Fatalf("catalog empty")
	}
	if d.SpecMinimal == "" || d.ErrorCodes["missing_node"].Fix == "" {
		t.Fatal("missing content")
	}
	data, err := json.Marshal(d)
	if err != nil || !json.Valid(data) {
		t.Fatal("invalid json")
	}
}
