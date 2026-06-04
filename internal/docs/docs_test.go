package docs

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestCatalogJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteCatalogJSON(&buf); err != nil {
		t.Fatal(err)
	}
	var c Catalog
	if err := json.Unmarshal(buf.Bytes(), &c); err != nil {
		t.Fatal(err)
	}
	if c.Tool != "sceno" || c.StartHere == "" || len(c.Topics) < 5 {
		t.Fatalf("catalog: %+v", c)
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
}
