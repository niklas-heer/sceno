// Package guide provides self-documentation for humans and AI agents.
package guide

import (
	"encoding/json"
	"io"
	"sort"
	"strings"

	"github.com/niklas-heer/sceno/internal/diag"
	"github.com/niklas-heer/sceno/internal/icons"
	"github.com/niklas-heer/sceno/internal/model"
)

// Document is the full machine-readable guide (sceno guide --json).
type Document struct {
	Tool           string                 `json:"tool"`
	Description    string                 `json:"description"`
	Workflow       []string               `json:"workflow"`
	IterateLoop    []string               `json:"iterate_loop"`
	Commands       map[string]string      `json:"commands"`
	ErrorCodes     map[string]diag.CodeDoc `json:"error_codes"`
	Shapes         []string               `json:"shapes"`
	Icons          []string               `json:"icons"`
	DiagramProps   map[string]string      `json:"diagram_properties"`
	ShapeProps     map[string]string      `json:"shape_properties"`
	EdgeProps      map[string]string      `json:"edge_properties"`
	SpecMinimal    string                 `json:"spec_minimal"`
	SpecSlides     string                 `json:"spec_slides"`
	CommonMistakes []string               `json:"common_mistakes"`
	BestPractices  []string               `json:"best_practices"`
	RenderFormats  []string               `json:"render_formats"`
	GoalsSummary   string                 `json:"goals_summary"`
}

// JSON writes the agent guide.
func JSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(Build())
}

// Markdown writes a readable guide for humans.
func Markdown(w io.Writer) error {
	d := Build()
	var b strings.Builder
	b.WriteString("# Sceno — agent guide\n\n")
	b.WriteString(d.Description + "\n\n")
	b.WriteString("## Workflow\n\n")
	for _, step := range d.Workflow {
		b.WriteString("1. " + step + "\n")
	}
	b.WriteString("\n## Iterate until valid\n\n")
	for _, step := range d.IterateLoop {
		b.WriteString("- " + step + "\n")
	}
	b.WriteString("\n## Commands\n\n")
	for k, v := range d.Commands {
		b.WriteString("- `" + k + "` — " + v + "\n")
	}
	b.WriteString("\n## Minimal spec\n\n```kdl\n")
	b.WriteString(d.SpecMinimal)
	b.WriteString("\n```\n\n## Slides spec\n\n```kdl\n")
	b.WriteString(d.SpecSlides)
	b.WriteString("\n```\n\n## Error codes\n\n")
	codes := make([]string, 0, len(d.ErrorCodes))
	for c := range d.ErrorCodes {
		codes = append(codes, c)
	}
	sort.Strings(codes)
	for _, c := range codes {
		doc := d.ErrorCodes[c]
		b.WriteString("### `" + c + "`\n")
		b.WriteString(doc.Meaning + "\n\n")
		b.WriteString("**Fix:** " + doc.Fix + "\n\n")
		if doc.Example != "" {
			b.WriteString("```kdl\n" + doc.Example + "\n```\n\n")
		}
	}
	b.WriteString("## Shapes\n\n")
	b.WriteString(strings.Join(d.Shapes, ", ") + "\n\n")
	b.WriteString("## Icons\n\n")
	b.WriteString(strings.Join(d.Icons, ", ") + "\n\n")
	b.WriteString("## Common mistakes\n\n")
	for _, m := range d.CommonMistakes {
		b.WriteString("- " + m + "\n")
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// Build returns the guide document.
func Build() Document {
	shapeList := model.AllShapes()
	sort.Strings(shapeList)
	iconList := icons.Names()
	sort.Strings(iconList)

	codes := make(map[string]diag.CodeDoc)
	for c, doc := range diag.ErrorCatalog {
		codes[string(c)] = doc
	}

	return Document{
		Tool:        "sceno",
		Description: "Local-first declarative diagrams in KDL. One spec file → SVG, PNG, PDF, HTML, and slide decks. Optimized for AI edit/validate/render loops.",
		Workflow: []string{
			"sceno init -o sceno.kdl",
			"Edit sceno.kdl (KDL only — see spec_minimal)",
			"sceno validate -i sceno.kdl --json",
			"If ok is false: apply each errors[].fix and errors[].example, then validate again",
			"sceno describe -i sceno.kdl --json  (optional: sanity-check layout without viewing PNG)",
			"sceno render -i sceno.kdl -o output/sceno --all",
		},
		IterateLoop: []string{
			"Always run validate --json after editing the KDL file",
			"Read agent.next_steps and agent.summary in the JSON response",
			"Never invent shape kinds or icon names — use lists in this guide",
			"Use layout=auto with layer, row, or at=col,row unless you need exact x/y",
			"Quote labels with spaces: title=\"My Platform\" not title=My Platform",
			"Use \\n inside quotes for line breaks: \"API\\nGateway\"",
			"Edges only connect node ids defined in the same diagram or slide { } block",
			"Use sceno describe --json to see spatial layout without opening images",
		},
		Commands: map[string]string{
			"sceno guide [--json]":     "This document — start here for AI context",
			"sceno init [-o file.kdl]":  "Create a starter spec",
			"sceno validate -i f --json": "Check spec + layout; returns ok, errors, next_steps",
			"sceno describe -i f --json": "2D scene (layers, occlusion, edge visibility) + ascii_map + visual_problems",
			"sceno render -i f -o out --all": "Export svg, png, pdf, html, slides.html",
			"sceno render -format slides": "HTML presentation (16:9)",
			"sceno spec":               "Full KDL specification (markdown)",
			"sceno shapes":             "List shape kinds",
			"sceno icons":              "List icon names",
			"sceno goals":              "Product goals",
		},
		ErrorCodes:  codes,
		Shapes:      append(shapeList, "code (lang=, source=) — syntax-highlighted block for slides"),
		Icons:       iconList,
		DiagramProps: map[string]string{
			"title":    "Diagram title (quoted if spaces)",
			"subtitle": "Subtitle under title",
			"layout":   "auto (default) | free (requires x,y on every shape)",
			"style":    "polished (default) | sketch",
			"gap":      "Spacing between nodes (default 28)",
			"padding":  "Canvas padding (default 20)",
			"slide":    "16x9 or 4x3 — frame exports as presentation slides",
			"theme":    "light or dark — colors for slides/SVG/HTML",
			"background": "transparent — no canvas fill (PNG/SVG overlays)",
			"foreground": "Override text color (#hex)",
			"card":       "Override card/surface color",
			"border":     "Override border color",
			"muted":      "Override muted surface color",
			"accent":     "Override accent color",
			"var.NAME": "custom theme variable (e.g. var.card=#18181b)",
		},
		ShapeProps: map[string]string{
			"icon":     "Catalog icon name",
			"fill":     "Background #hex",
			"stroke":   "Border #hex",
			"accent":   "Callout stripe #hex",
			"subtitle": "Second line of text",
			"layer":    "Column index (auto layout)",
			"row":      "Row within column",
			"at":       "Shorthand layer,row e.g. at=1,2",
			"w, h":     "Minimum width/height (auto-expands for text)",
			"x, y":     "Fixed position (layout free)",
			"parent":   "Parent lane/container id",
			"lang":     "Code language (go, json, yaml, bash, kdl)",
			"source":   "Code body for shape code (use \\n)",
		},
		EdgeProps: map[string]string{
			"fromSide / toSide": "top | right | bottom | left",
			"dashed":            "true for dashed line",
			"color":             "#hex stroke color",
		},
		SpecMinimal: `diagram title="My Platform" layout=auto style=polished gap=32 padding=24 {

  shape box api "API Gateway" icon=api layer=1
  shape cylinder db "Database" icon=database layer=2
  shape actor ops "Operators" at=0,0

  edge ops -> api fromSide=right toSide=left
  edge api -> db
}`,
		SpecSlides: `diagram title="Talk" slide=16x9 layout=auto gap=36 {

  slide "Overview" {
    shape callout note "Summary" icon=info at=0,0
  }

  slide "Architecture" {
    shape box api "API" icon=api layer=1
    shape box db "DB" icon=database layer=2
    edge api -> db
  }
}`,
		GoalsSummary: "Run sceno goals for mission, quality bar, and ecosystem best practices (d2, Mermaid, Excalidraw, slides, theming, scene analysis).",
		BestPractices: []string{
			"Spec is source of truth — never hand-tweak exports; change KDL and re-render",
			"Validate → describe → render (agents: use --json on validate and describe)",
			"Set fromSide/toSide when edges cross nodes; increase gap for dense diagrams",
			"Use theme=dark for slide decks; background=transparent for embed overlays",
			"Group related nodes in columns (layer/at); avoid single-node orphan columns when possible",
			"Polished style for architecture; sketch style for whiteboard/Excalidraw-like organic edges",
			"Slides: one slide block per screen; mix sceno shapes and code blocks as needed",
			"Borrowed from d2/Mermaid: declarative text, themes, layers; from PowerPoint: slide blocks",
		},
		CommonMistakes: []string{
			"Using YAML/JSON — only .kdl is accepted",
			"title=My Platform without quotes — use title=\"My Platform\"",
			"edge to missing node — define shape before edge in the same block",
			"layout=free without x= and y= on every shape",
			"icon=unknown — run sceno icons",
			"Shapes only in slide { } but edges reference ids from another slide",
			"Duplicate node ids in the same diagram or slide",
		},
		RenderFormats: []string{"svg", "png", "pdf", "html", "slides", "all"},
	}
}
