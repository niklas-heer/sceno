// Package docs serves self-documentation for humans and AI agents.
package docs

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/niklas-heer/sceno/internal/diag"
	"github.com/niklas-heer/sceno/internal/guide"
	"github.com/niklas-heer/sceno/internal/icons"
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/spec"
	"github.com/niklas-heer/sceno/internal/version"
)

// Topic is a documentation section.
type Topic string

const (
	TopicGuide     Topic = "guide"
	TopicSpec      Topic = "spec"
	TopicGoals     Topic = "goals"
	TopicPractices Topic = "practices"
	TopicShapes    Topic = "shapes"
	TopicIcons     Topic = "icons"
	TopicErrors    Topic = "errors"
)

// AllTopics lists available doc topics in display order.
var AllTopics = []Topic{
	TopicGuide, TopicSpec, TopicGoals, TopicPractices, TopicShapes, TopicIcons, TopicErrors,
}

// Catalog is the machine-readable index (sceno docs --json with no topic).
type Catalog struct {
	Tool        string            `json:"tool"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	StartHere   string            `json:"start_here"`
	Topics      map[string]string `json:"topics"`
	Commands    map[string]string `json:"commands"`
}

// PracticesDoc is best-practices for agents.
type PracticesDoc struct {
	Tool           string   `json:"tool"`
	Version        string   `json:"version"`
	Workflow       []string `json:"workflow"`
	IterateLoop    []string `json:"iterate_loop"`
	BestPractices  []string `json:"best_practices"`
	CommonMistakes []string `json:"common_mistakes"`
	RenderFormats  []string `json:"render_formats"`
}

// ShapesDoc lists shape kinds.
type ShapesDoc struct {
	Tool    string   `json:"tool"`
	Version string   `json:"version"`
	Shapes  []string `json:"shapes"`
	Usage   string   `json:"usage"`
}

// IconsDoc lists icon names.
type IconsDoc struct {
	Tool    string   `json:"tool"`
	Version string   `json:"version"`
	Icons   []string `json:"icons"`
	Usage   string   `json:"usage"`
}

// ErrorsDoc is the full error catalog for repair loops.
type ErrorsDoc struct {
	Tool       string                 `json:"tool"`
	Version    string                 `json:"version"`
	ErrorCodes map[string]diag.CodeDoc `json:"error_codes"`
}

// WriteCatalogJSON writes the docs index.
func WriteCatalogJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(BuildCatalog())
}

// BuildCatalog returns the docs index.
func BuildCatalog() Catalog {
	topics := map[string]string{
		string(TopicGuide):     "Agent handbook — workflow, commands, examples, properties",
		string(TopicSpec):      "Full KDL specification (diagram, shapes, edges, layout, theme)",
		string(TopicGoals):     "Product mission, quality bar, ecosystem best practices",
		string(TopicPractices): "Authoring workflow, iterate loop, best practices, common mistakes",
		string(TopicShapes):    "Allowed shape kinds",
		string(TopicIcons):     "Allowed icon names",
		string(TopicErrors):    "Error codes with fix and example for every validation issue",
	}
	return Catalog{
		Tool:        "sceno",
		Version:     version.Version,
		Description: "Self-documenting CLI — every topic is available as markdown or JSON for AI agents.",
		StartHere:   "sceno docs guide --json",
		Topics:      topics,
		Commands: map[string]string{
			"sceno docs":                    "List topics (add --json for catalog)",
			"sceno docs guide --json":       "Full agent handbook",
			"sceno docs spec":               "KDL specification",
			"sceno docs practices --json":   "Best practices + common mistakes",
			"sceno docs errors --json":      "Error code repair catalog",
			"sceno validate -i f --json":    "Validate spec after every edit",
			"sceno describe -i f --json":    "Layout feedback without viewing images",
		},
	}
}

// WriteHumanUsage prints available topics.
func WriteHumanUsage(w io.Writer) {
	c := BuildCatalog()
	fmt.Fprintf(w, "sceno docs — self-documentation for humans and AI agents\n\n")
	fmt.Fprintln(w, "Start here: sceno docs guide --json")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Topics:")
	for _, t := range AllTopics {
		fmt.Fprintf(w, "  %-12s %s\n", t, c.Topics[string(t)])
	}
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage: sceno docs TOPIC [--json]")
	fmt.Fprintln(w, "       sceno docs --json          (topic catalog)")
}

// Run writes documentation for a topic.
func Run(topic string, jsonOut bool, w io.Writer) error {
	t := Topic(strings.ToLower(strings.TrimSpace(topic)))
	if topic == "" {
		if jsonOut {
			return WriteCatalogJSON(w)
		}
		WriteHumanUsage(w)
		return nil
	}
	switch t {
	case TopicGuide:
		if jsonOut {
			return guide.JSON(w)
		}
		return guide.Markdown(w)
	case TopicSpec:
		if jsonOut {
			return writeSpecJSON(w)
		}
		_, err := io.WriteString(w, spec.SpecMarkdown)
		return err
	case TopicGoals:
		_, err := io.WriteString(w, spec.GoalsMarkdown)
		return err
	case TopicPractices:
		if jsonOut {
			return writePracticesJSON(w)
		}
		return writePracticesMarkdown(w)
	case TopicShapes:
		if jsonOut {
			return writeShapesJSON(w)
		}
		return writeShapesHuman(w)
	case TopicIcons:
		if jsonOut {
			return writeIconsJSON(w)
		}
		return writeIconsHuman(w)
	case TopicErrors:
		if jsonOut {
			return writeErrorsJSON(w)
		}
		return writeErrorsMarkdown(w)
	default:
		return fmt.Errorf("unknown docs topic %q — run sceno docs for topics", topic)
	}
}

func writeSpecJSON(w io.Writer) error {
	out := map[string]string{
		"tool":    "sceno",
		"version": version.Version,
		"format":  "kdl",
		"spec":    spec.SpecMarkdown,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func writePracticesJSON(w io.Writer) error {
	g := guide.Build()
	doc := PracticesDoc{
		Tool:           "sceno",
		Version:        version.Version,
		Workflow:       g.Workflow,
		IterateLoop:    g.IterateLoop,
		BestPractices:  g.BestPractices,
		CommonMistakes: g.CommonMistakes,
		RenderFormats:  g.RenderFormats,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(doc)
}

func writePracticesMarkdown(w io.Writer) error {
	g := guide.Build()
	var b strings.Builder
	b.WriteString("# Sceno — best practices\n\n")
	b.WriteString("## Workflow\n\n")
	for i, step := range g.Workflow {
		fmt.Fprintf(&b, "%d. %s\n", i+1, step)
	}
	b.WriteString("\n## Iterate until valid\n\n")
	for _, step := range g.IterateLoop {
		b.WriteString("- " + step + "\n")
	}
	b.WriteString("\n## Best practices\n\n")
	for _, p := range g.BestPractices {
		b.WriteString("- " + p + "\n")
	}
	b.WriteString("\n## Common mistakes\n\n")
	for _, m := range g.CommonMistakes {
		b.WriteString("- " + m + "\n")
	}
	b.WriteString("\n## Export formats\n\n")
	b.WriteString(strings.Join(g.RenderFormats, ", ") + "\n")
	_, err := io.WriteString(w, b.String())
	return err
}

func writeShapesJSON(w io.Writer) error {
	shapes := model.AllShapes()
	sort.Strings(shapes)
	doc := ShapesDoc{
		Tool:    "sceno",
		Version: version.Version,
		Shapes:  shapes,
		Usage:   `shape KIND id "Label" props...`,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(doc)
}

func writeShapesHuman(w io.Writer) error {
	fmt.Fprintln(w, `Shapes (use as: shape KIND id "Label" ...):`)
	for _, s := range model.AllShapes() {
		fmt.Fprintln(w, " ", s)
	}
	return nil
}

func writeIconsJSON(w io.Writer) error {
	names := icons.Names()
	sort.Strings(names)
	doc := IconsDoc{
		Tool:    "sceno",
		Version: version.Version,
		Icons:   names,
		Usage:   "icon=name on shape lines",
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(doc)
}

func writeIconsHuman(w io.Writer) error {
	fmt.Fprintln(w, "Icons (use as: icon=name):")
	for _, name := range icons.Names() {
		fmt.Fprintln(w, " ", name)
	}
	return nil
}

func writeErrorsJSON(w io.Writer) error {
	codes := make(map[string]diag.CodeDoc, len(diag.ErrorCatalog))
	for c, doc := range diag.ErrorCatalog {
		codes[string(c)] = doc
	}
	doc := ErrorsDoc{Tool: "sceno", Version: version.Version, ErrorCodes: codes}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(doc)
}

func writeErrorsMarkdown(w io.Writer) error {
	var b strings.Builder
	b.WriteString("# Sceno — error codes\n\n")
	codes := make([]string, 0, len(diag.ErrorCatalog))
	for c := range diag.ErrorCatalog {
		codes = append(codes, string(c))
	}
	sort.Strings(codes)
	for _, c := range codes {
		doc := diag.ErrorCatalog[diag.Code(c)]
		fmt.Fprintf(&b, "## `%s`\n\n", c)
		b.WriteString(doc.Meaning + "\n\n")
		b.WriteString("**Fix:** " + doc.Fix + "\n\n")
		if doc.Example != "" {
			b.WriteString("```kdl\n" + doc.Example + "\n```\n\n")
		}
	}
	_, err := io.WriteString(w, b.String())
	return err
}
