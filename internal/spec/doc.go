package spec

import "github.com/niklas-heer/sceno/internal/guide"

// SpecMarkdown returns the KDL specification (generated from code).
func SpecMarkdown() string { return guide.RenderSpecMarkdown() }

// GoalsMarkdown returns product goals (generated from code).
func GoalsMarkdown() string { return guide.RenderGoalsMarkdown() }

// StackMarkdown returns the stack validation model (generated from code).
func StackMarkdown() string { return guide.RenderStackMarkdown() }
