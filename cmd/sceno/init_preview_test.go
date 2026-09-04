package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niklas-heer/sceno/internal/starter"
)

func TestInitListsMetadataWithoutWriting(t *testing.T) {
	t.Chdir(t.TempDir())
	var output bytes.Buffer
	if err := runInit([]string{"--list", "--json"}, &output); err != nil {
		t.Fatal(err)
	}
	var templates []starter.Template
	if err := json.Unmarshal(output.Bytes(), &templates); err != nil {
		t.Fatal(err)
	}
	if len(templates) != len(starter.List()) || templates[0].Name != "service-architecture" {
		t.Fatalf("unexpected template metadata: %+v", templates)
	}
	if _, err := os.Stat("sceno.kdl"); !os.IsNotExist(err) {
		t.Fatalf("listing templates wrote a file: %v", err)
	}
}

func TestInitProtectsExistingFilesAndSupportsExplicitReplacement(t *testing.T) {
	t.Chdir(t.TempDir())
	var output bytes.Buffer
	if err := runInit(nil, &output); err != nil {
		t.Fatal(err)
	}
	want, _ := starter.Source("service-architecture")
	got, err := os.ReadFile("sceno.kdl")
	if err != nil || string(got) != want || !strings.Contains(output.String(), "sceno preview 'sceno.kdl'") {
		t.Fatalf("default creation failed: err=%v output=%s", err, &output)
	}
	if err := os.WriteFile("sceno.kdl", []byte("user edits"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runInit([]string{"--template", "request-flow"}, io.Discard); err == nil || !strings.Contains(err.Error(), "--force") {
		t.Fatalf("existing file must be protected: %v", err)
	}
	if got, _ := os.ReadFile("sceno.kdl"); string(got) != "user edits" {
		t.Fatal("refused creation changed existing user edits")
	}
	output.Reset()
	if err := runInit([]string{"--force", "--template", "request-flow", "--json"}, &output); err != nil {
		t.Fatal(err)
	}
	want, _ = starter.Source("request-flow")
	if got, _ := os.ReadFile("sceno.kdl"); string(got) != want {
		t.Fatal("explicit replacement did not write the selected template")
	}
	var result struct {
		Path      string           `json:"path"`
		Template  starter.Template `json:"template"`
		NextSteps []string         `json:"next_steps"`
	}
	if err := json.Unmarshal(output.Bytes(), &result); err != nil || result.Path != "sceno.kdl" || result.Template.Name != "request-flow" || len(result.NextSteps) != 2 {
		t.Fatalf("creation metadata: %+v, err=%v", result, err)
	}
}

func TestInitOutputPathAndInvalidTemplate(t *testing.T) {
	t.Chdir(t.TempDir())
	path := filepath.Join("new folder", "O'Brien")
	var output bytes.Buffer
	if err := runInit([]string{"-o", path, "--template", "default"}, &output); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path + ".kdl"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), `'new folder/O'"'"'Brien.kdl'`) {
		t.Fatalf("preview command does not shell-quote the path: %s", &output)
	}
	if err := runInit([]string{"-o", "missing.kdl", "--template", "no-such-template"}, io.Discard); err == nil {
		t.Fatal("unknown template accepted")
	}
	if _, err := os.Stat("missing.kdl"); !os.IsNotExist(err) {
		t.Fatalf("invalid template left an output file: %v", err)
	}
}

func TestPreviewArguments(t *testing.T) {
	for _, test := range []struct {
		args []string
		path string
		port int
		open bool
	}{
		{[]string{"diagram.kdl"}, "diagram.kdl", 0, true},
		{[]string{"diagram.kdl", "--no-open", "--port", "4123"}, "diagram.kdl", 4123, false},
		{[]string{"--port=8123", "-i", "my diagram.kdl", "--no-open"}, "my diagram.kdl", 8123, false},
		{[]string{"--no-open=false", "--", "-diagram.kdl"}, "-diagram.kdl", 0, true},
		{[]string{"--port", "0", "diagram.kdl", "--no-open"}, "diagram.kdl", 0, false},
	} {
		path, port, open, err := parsePreviewArgs(test.args, io.Discard)
		if err != nil || path != test.path || port != test.port || open != test.open {
			t.Errorf("args=%q: path=%q port=%d open=%v err=%v", test.args, path, port, open, err)
		}
	}
	for _, args := range [][]string{
		nil, {"a.kdl", "b.kdl"}, {"-i", "a.kdl", "b.kdl"},
		{"a.kdl", "--port", "-1"}, {"a.kdl", "--port=65536"},
		{"a.kdl", "--port", "bad"}, {"a.kdl", "--unknown"}, {"--port"},
	} {
		if _, _, _, err := parsePreviewArgs(args, io.Discard); err == nil {
			t.Errorf("invalid preview arguments accepted: %q", args)
		}
	}
}
