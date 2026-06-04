// Package version holds build metadata (set via ldflags in CI/releases).
package version

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

//go:embed VERSION
var embedded string

var (
	// Version is the semantic release (from VERSION file or -X ldflag).
	Version = strings.TrimSpace(embedded)
	// Commit is the git SHA at build time (optional).
	Commit = "unknown"
	// Date is the build timestamp UTC (optional).
	Date = ""
)

// Info returns structured metadata for JSON output.
func Info() map[string]string {
	m := map[string]string{
		"version": Version,
		"tool":    "sceno",
	}
	if Commit != "" && Commit != "unknown" {
		m["commit"] = Commit
	}
	if Date != "" {
		m["date"] = Date
	}
	return m
}

// String is a human-readable version line.
func String() string {
	if Commit != "" && Commit != "unknown" {
		return fmt.Sprintf("%s (%s)", Version, Commit)
	}
	return Version
}

// WriteHuman prints version info to w.
func WriteHuman(w io.Writer) {
	fmt.Fprintf(w, "sceno %s\n", Version)
	if Commit != "" && Commit != "unknown" {
		fmt.Fprintf(w, "commit: %s\n", Commit)
	}
	if Date != "" {
		fmt.Fprintf(w, "built: %s\n", Date)
	}
}

// WriteJSON prints version info as JSON.
func WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(Info())
}
