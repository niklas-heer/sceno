package docs

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRunSpecFromCode(t *testing.T) {
	var buf bytes.Buffer
	if err := Run("spec", false, &buf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "iconPos") {
		t.Fatal("spec should be generated from code with iconPos")
	}
}

func TestCatalogJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteCatalogJSON(&buf); err != nil {
		t.Fatal(err)
	}
	var c Catalog
	if err := json.Unmarshal(buf.Bytes(), &c); err != nil {
		t.Fatal(err)
	}
	if c.Tool != "sceno" || c.StartHere == "" || len(c.Topics) < 8 {
		t.Fatalf("catalog: %+v", c)
	}
	if c.Topics["stack"] == "" || c.Topics["validation"] == "" {
		t.Fatal("expected stack and validation topics")
	}
	if c.Commands["sceno advise -i f --json"] == "" {
		t.Fatal("expected advise in catalog commands")
	}
}

func TestRunPracticesJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := Run("practices", true, &buf); err != nil {
		t.Fatal(err)
	}
	var doc PracticesDoc
	if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.BestPractices) < 3 {
		t.Fatal("expected best practices")
	}
	if doc.StackModel == "" || len(doc.VisualRules) < 5 {
		t.Fatal("expected stack model and visual rules in practices")
	}
}

func TestRunStackJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := Run("stack", true, &buf); err != nil {
		t.Fatal(err)
	}
	var doc StackDoc
	if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if doc.StackModel == "" || len(doc.VisualRules) < 5 || doc.Markdown == "" {
		t.Fatalf("stack doc incomplete: %+v", doc)
	}
}

func TestRunValidationJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := Run("validation", true, &buf); err != nil {
		t.Fatal(err)
	}
	var doc ValidationDoc
	if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if doc.ValidateCommand == "" || len(doc.ErrorCodes) < 10 {
		t.Fatalf("validation doc: %+v", doc)
	}
}

func TestRunErrorsJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := Run("errors", true, &buf); err != nil {
		t.Fatal(err)
	}
	var doc ErrorsDoc
	if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if doc.ErrorCodes["missing_node"].Fix == "" {
		t.Fatal("missing error doc")
	}
	if doc.ErrorCodes["dense_layout"].Fix == "" {
		t.Fatal("expected dense_layout in catalog")
	}
	if doc.ErrorCodes["arrow_detached"].Fix == "" {
		t.Fatal("expected arrow_detached in catalog")
	}
}

func TestRunArchitectureJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := Run("architecture", true, &buf); err != nil {
		t.Fatal(err)
	}
	var doc ArchitectureDoc
	if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if doc.EntryPoint == "" || doc.GeometrySoT == "" || doc.SemanticsSoT == "" {
		t.Fatalf("architecture doc incomplete: %+v", doc)
	}
	if len(doc.Pipeline) < 4 || len(doc.Consumers) < 4 {
		t.Fatalf("expected full pipeline doc: %+v", doc)
	}
}

func TestRunGoalsJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := Run("goals", true, &buf); err != nil {
		t.Fatal(err)
	}
	var doc GoalsDoc
	if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Mission == "" || len(doc.ProductGoals) < 5 || len(doc.Principles) < 2 {
		t.Fatalf("goals doc incomplete: %+v", doc)
	}
}

func TestRunIconsJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := Run("icons", true, &buf); err != nil {
		t.Fatal(err)
	}
	var doc IconsDoc
	if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Icons) < 25 || len(doc.Categories) < 5 || len(doc.Tips) < 3 {
		t.Fatalf("icons doc incomplete: icons=%d categories=%d", len(doc.Icons), len(doc.Categories))
	}
	if doc.ByCategory["data"] == nil || doc.Names[0] == "" {
		t.Fatalf("expected by_category and names: %+v", doc)
	}
}

func TestRunShapesJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := Run("shapes", true, &buf); err != nil {
		t.Fatal(err)
	}
	var doc ShapesDoc
	if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	hasInfo := false
	for _, s := range doc.Shapes {
		if s == "info" {
			hasInfo = true
		}
	}
	if !hasInfo || doc.Props["iconPos"] == "" {
		t.Fatalf("shapes doc: %+v", doc)
	}
}
