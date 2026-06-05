package spec

import _ "embed"

// SpecMarkdown is the full KDL specification.
//
//go:embed SPEC.md
var SpecMarkdown string

// GoalsMarkdown describes product mission and quality bar.
//
//go:embed GOALS.md
var GoalsMarkdown string

// StackMarkdown describes the stacked-plane validation model.
//
//go:embed STACK.md
var StackMarkdown string
