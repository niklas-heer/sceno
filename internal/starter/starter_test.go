package starter_test

import (
	"reflect"
	"testing"

	"github.com/niklas-heer/sceno/internal/pipeline"
	"github.com/niklas-heer/sceno/internal/spec"
	"github.com/niklas-heer/sceno/internal/starter"
)

func TestEveryTemplateMeetsVisualQualityBar(t *testing.T) {
	blocked := map[string]bool{
		"collision": true, "edge_collision": true, "edge_detour": true,
		"edge_hidden": true, "arrow_detached": true, "arrow_hidden": true,
		"arrowhead_cluster": true, "arrow_cluster": true, "edge_label_overlap": true,
		"edge_label_chrome_overlap": true, "edge_label_off_axis": true,
		"edge_side_mismatch": true, "occluded": true, "text_overflow": true,
	}
	for _, template := range starter.List() {
		t.Run(template.Name, func(t *testing.T) {
			source, err := starter.Source(template.Name)
			if err != nil {
				t.Fatal(err)
			}
			s, err := spec.LoadKDL([]byte(source))
			if err != nil {
				t.Fatal(err)
			}
			if issues := spec.Validate(s); len(issues) != 0 {
				t.Fatalf("invalid template: %+v", issues)
			}
			for _, fix := range []bool{false, true} {
				opt := pipeline.DefaultOptions()
				opt.ResolveCollision = fix
				result, err := pipeline.BuildAndEvaluate(s, opt)
				if err != nil {
					t.Fatal(err)
				}
				if len(result.Collisions) != 0 || len(result.Slides) == 0 {
					t.Fatalf("fix=%v: missing slides or collisions: %+v", fix, result.Collisions)
				}
				for i, slide := range result.Slides {
					if slide.Eval.Score < 80 {
						t.Errorf("fix=%v slide=%d score=%d, want >=80", fix, i+1, slide.Eval.Score)
					}
					for _, finding := range slide.Eval.Findings {
						if blocked[finding.Code] || finding.Severity == "error" {
							t.Errorf("fix=%v slide=%d: %s: %s", fix, i+1, finding.Code, finding.Message)
						}
					}
					if len(slide.Eval.Spacing.Nodes) != len(slide.Diagram.Nodes) {
						t.Errorf("slide %d omits measurable node geometry", i+1)
					}
				}
			}
		})
	}
}

func TestCatalogAndDefaultSelection(t *testing.T) {
	list := starter.List()
	if len(list) != 4 {
		t.Fatalf("expected four templates, got %+v", list)
	}
	seen := map[string]bool{}
	for _, template := range list {
		if seen[template.Name] || template.Name == "" || template.Title == "" || template.Description == "" {
			t.Fatalf("incomplete or duplicate template: %+v", template)
		}
		seen[template.Name] = true
	}
	list[0].Name = "caller mutation"
	if reflect.DeepEqual(list, starter.List()) {
		t.Fatal("List exposes mutable catalog storage")
	}
	want, err := starter.Source("service-architecture")
	if err != nil {
		t.Fatal(err)
	}
	for _, alias := range []string{"", "default", " default "} {
		got, err := starter.Source(alias)
		if err != nil || got != want {
			t.Fatalf("alias %q: source mismatch, err=%v", alias, err)
		}
	}
	for _, name := range []string{"missing", "../templates/service-architecture.kdl", "service-architecture.kdl"} {
		if source, err := starter.Source(name); err == nil || source != "" {
			t.Errorf("unknown template %q unexpectedly accepted", name)
		}
	}
}
