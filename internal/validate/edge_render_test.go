package validate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateCatchesDetachedArrow(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tight.kdl")
	content := `diagram layout=auto gap=8 {
  shape box a "A" at=0,0
  shape box b "B" at=1,0
  edge a -> b fromSide=right toSide=left
}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	report, _, _ := Run(path, Options{FixCollisions: true})
	found := false
	for _, iss := range append(report.Errors, report.Warnings...) {
		if iss.Code == "arrow_detached" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected arrow_detached, got errors=%+v warnings=%+v", report.Errors, report.Warnings)
	}
}
