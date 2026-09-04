// Package starter supplies complete, validated KDL starting points embedded in
// the binary. Templates use the same parser, layout, and quality gates as user diagrams.
package starter

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed templates/*.kdl
var sources embed.FS

// Template describes a named starting point for a diagram or presentation.
type Template struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

var catalog = []Template{
	{Name: "service-architecture", Title: "Service architecture", Description: "An API service, persistent data, and asynchronous workers with clear component roles."},
	{Name: "deployment-pipeline", Title: "Deployment pipeline", Description: "A delivery path from source through build, verification, release, and production."},
	{Name: "request-flow", Title: "Request flow", Description: "An authenticated request with explicit success and rejection paths."},
	{Name: "presentation", Title: "Architecture presentation", Description: "Three focused slides covering the audience, system design, and delivery plan."},
}

// List returns the named templates in their presentation order. The returned
// slice is independent of the catalog and can be safely sorted by callers.
func List() []Template {
	return append([]Template(nil), catalog...)
}

// Source returns editable KDL. Empty and "default" select service-architecture.
func Source(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || name == "default" {
		name = "service-architecture"
	}
	for _, template := range catalog {
		if template.Name == name {
			data, err := sources.ReadFile("templates/" + name + ".kdl")
			if err != nil {
				return "", fmt.Errorf("read template %q: %w", name, err)
			}
			return string(data), nil
		}
	}
	return "", fmt.Errorf("unknown template %q; choose service-architecture, deployment-pipeline, request-flow, or presentation", name)
}
