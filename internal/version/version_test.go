package version

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestVersionNonEmpty(t *testing.T) {
	if strings.TrimSpace(Version) == "" {
		t.Fatal("Version must not be empty")
	}
}

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteJSON(&buf); err != nil {
		t.Fatal(err)
	}
	var m map[string]string
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatal(err)
	}
	if m["version"] != Version || m["tool"] != "sceno" {
		t.Fatalf("unexpected json: %+v", m)
	}
}
