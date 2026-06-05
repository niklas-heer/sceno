package guide

// ShapeEntry documents one shape kind (source of truth for spec + docs).
type ShapeEntry struct {
	Kind    string
	Aliases []string
	Use     string
}

// ShapeCatalog returns documented shapes in display order.
func ShapeCatalog() []ShapeEntry {
	return []ShapeEntry{
		{Kind: "box", Aliases: []string{"card", "process"}, Use: "Default card"},
		{Kind: "ellipse", Aliases: []string{"actor", "circle"}, Use: "People / roles"},
		{Kind: "diamond", Aliases: []string{"decision"}, Use: "Branching"},
		{Kind: "hexagon", Use: "External / API"},
		{Kind: "octagon", Use: "Stop / boundary"},
		{Kind: "cylinder", Aliases: []string{"database", "db"}, Use: "Data store"},
		{Kind: "cloud", Use: "Cloud service"},
		{Kind: "document", Aliases: []string{"doc"}, Use: "Document / subprocess"},
		{Kind: "parallelogram", Aliases: []string{"input", "output", "io"}, Use: "I/O"},
		{Kind: "triangle", Use: "Merge / split"},
		{Kind: "pill", Aliases: []string{"terminal", "start", "end"}, Use: "Start / end"},
		{Kind: "textbox", Use: "Light annotation"},
		{Kind: "note", Aliases: []string{"sticky", "postit"}, Use: "Sticky note (yellow)"},
		{Kind: "infobox", Aliases: []string{"callout"}, Use: "Accent callout + subtitle"},
		{Kind: "info", Use: "Blue infobox (default accent #3b82f6)"},
		{Kind: "warning", Aliases: []string{"warn"}, Use: "Amber infobox (default accent #f59e0b)"},
		{Kind: "tip", Aliases: []string{"hint"}, Use: "Green infobox (default accent #10b981)"},
		{Kind: "lane", Aliases: []string{"container"}, Use: "Dashed swimlane"},
		{Kind: "frame", Aliases: []string{"group"}, Use: "Solid group"},
		{Kind: "code", Aliases: []string{"codeblock"}, Use: "Syntax-highlighted block (lang=, source=)"},
	}
}

// IconPosOptions documents iconPos values.
func IconPosOptions() []string {
	return []string{
		"top-left (default)",
		"top",
		"top-right",
		"center",
		"bottom-left",
		"bottom",
		"bottom-right",
	}
}

// TopicDescriptions maps docs topic names to descriptions.
func TopicDescriptions() map[string]string {
	return map[string]string{
		"guide":      "Agent handbook — workflow, commands, examples, properties, stack_model, visual_rules",
		"spec":       "Full KDL specification (generated from code — diagram, shapes, edges, layout, theme)",
		"goals":      "Product mission, quality bar, ecosystem best practices",
		"practices":  "Authoring workflow, iterate loop, best practices, common mistakes, visual rules",
		"stack":      "Stacked 2D plane validation model — lanes, edges, annotations, nodes, labels",
		"validation": "validate + advise commands, error codes, visual rules, stack model summary",
		"shapes":     "Allowed shape kinds including info, tip, warning callouts",
		"icons":      "Allowed icon names",
		"errors":     "Error and warning codes with fix and example for every validation issue",
	}
}

// DocsCatalogCommands returns CLI commands shown in sceno docs --json.
func DocsCatalogCommands() map[string]string {
	return map[string]string{
		"sceno docs":                   "List topics (add --json for catalog)",
		"sceno docs guide --json":      "Full agent handbook",
		"sceno docs spec":              "KDL specification (generated from code)",
		"sceno docs stack [--json]":    "Stack validation model + visual rules",
		"sceno docs validation --json": "Validation + advise reference",
		"sceno docs practices --json":  "Best practices + common mistakes + visual rules",
		"sceno docs errors --json":     "Error code repair catalog",
		"sceno validate -i f --json":   "Validate spec after every edit",
		"sceno advise -i f --json":     "Stack engine + visual score + recommendations",
		"sceno describe -i f --json":   "Layout feedback without viewing images",
		"sceno suggest -i f --json":    "Prioritized layout recommendations",
	}
}

// ShapeNotes returns extra shape authoring hints for docs.
func ShapeNotes() []string {
	return []string{
		"info, warning, tip are semantic infobox variants with default accent colors",
		"iconPos controls icon placement (top-left default)",
	}
}

// CodeLanguages lists supported code block languages.
func CodeLanguages() []string {
	return []string{"go", "json", "yaml", "bash", "kdl", "text"}
}

// SpecExample returns the canonical spec example with callouts.
func SpecExample() string {
	return `diagram title="My Platform" subtitle="Optional" layout=auto gap=32 padding=24 {

  shape box api "API Gateway" icon=api iconPos=top-left layer=1
  shape cylinder db "PostgreSQL" icon=database layer=2
  shape info note "Context" icon=info subtitle="Read left to right" at=0,2

  edge api -> db fromSide=right toSide=left label="SQL"
}`
}