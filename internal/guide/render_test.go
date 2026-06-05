package guide

import (
	"strings"
	"testing"
)

func TestRenderSpecFromCode(t *testing.T) {
	md := RenderSpecMarkdown()
	for _, want := range []string{"iconPos", "info", "Stack validation", "sceno advise"} {
		if !strings.Contains(md, want) {
			t.Fatalf("spec missing %q", want)
		}
	}
}

func TestRenderGoalsFromCode(t *testing.T) {
	md := RenderGoalsMarkdown()
	if !strings.Contains(md, "Self-documenting") {
		t.Fatal("goals should mention self-documenting")
	}
}

func TestRenderStackFromCode(t *testing.T) {
	md := RenderStackMarkdown()
	if !strings.Contains(md, "annotations") || !strings.Contains(md, "whitespace") {
		t.Fatal("stack doc incomplete")
	}
}

func TestShapeCatalogMatchesAllShapes(t *testing.T) {
	catalog := ShapeCatalog()
	if len(catalog) < 15 {
		t.Fatalf("expected shape catalog entries, got %d", len(catalog))
	}
}

func TestBuildGoalsJSON(t *testing.T) {
	g := BuildGoals()
	if g.Mission == "" || len(g.ProductGoals) < 5 {
		t.Fatalf("goals: %+v", g)
	}
}
