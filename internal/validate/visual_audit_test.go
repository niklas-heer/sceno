package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niklas-heer/sceno/internal/diag"
)

var hardVisualCodes = map[diag.Code]bool{
	diag.CodeArrowDetached: true,
	diag.CodeArrowHidden:   true,
}

func TestVisualAuditExamplesAndFixtures(t *testing.T) {
	root := filepath.Join("..", "..", "examples")
	dirs := []string{root, filepath.Join(root, "fixtures")}
	var files []string
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".kdl") {
				continue
			}
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	if len(files) < 5 {
		t.Fatalf("expected examples + fixtures, got %d files", len(files))
	}

	for _, path := range files {
		name := filepath.Base(path)
		t.Run(name, func(t *testing.T) {
			_, report, err := LoadAndEvaluate(path, Options{FixCollisions: true})
			if err != nil {
				t.Fatalf("load/evaluate: %v errors=%v warnings=%v", err, report.Errors, report.Warnings)
			}
			if !report.OK {
				t.Fatalf("validate not ok: errors=%v warnings=%v", report.Errors, report.Warnings)
			}
			for _, iss := range report.Errors {
				if hardVisualCodes[iss.Code] {
					t.Fatalf("hard visual error %s: %s", iss.Code, iss.Message)
				}
			}
		})
	}
}
