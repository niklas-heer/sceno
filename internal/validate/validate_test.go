package validate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateJSONOK(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "self-service.kdl")
	report, _, err := Run(path, Options{FixCollisions: true})
	if err != nil {
		t.Fatal(err)
	}
	if !report.OK {
		t.Fatalf("expected ok, errors: %+v warnings: %+v", report.Errors, report.Warnings)
	}
	if len(report.Warnings) == 0 {
		t.Log("no compact suggestions (optional)")
	}
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(data) {
		t.Fatal("invalid json")
	}
}

func TestValidateBrokenSpec(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.kdl")
	if err := os.WriteFile(path, []byte(`diagram {
  edge a -> b
}`), 0o644); err != nil {
		t.Fatal(err)
	}
	report, _, _ := Run(path, Options{})
	if report.OK {
		t.Fatal("expected failure")
	}
	if len(report.Errors) == 0 || report.Agent.Summary == "" {
		t.Fatalf("expected enriched errors: %+v", report)
	}
}

func TestValidateSlidesDemo(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "slides-demo.kdl")
	report, _, err := Run(path, Options{FixCollisions: true})
	if err != nil {
		t.Fatal(err)
	}
	if !report.OK {
		t.Fatalf("slides-demo should validate: %+v", report.Errors)
	}
	if report.Stats.Slides != 3 {
		t.Fatalf("slides: %d", report.Stats.Slides)
	}
}

func TestValidateJSONHasAgent(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "self-service.kdl")
	report, _, _ := Run(path, Options{FixCollisions: true})
	report.Enrich()
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(data) {
		t.Fatal("invalid json")
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["agent"]; !ok {
		t.Fatal("missing agent field")
	}
}
