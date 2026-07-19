package spec

import (
	"strings"
	"testing"

	"github.com/niklas-heer/sceno/internal/diag"
	"github.com/niklas-heer/sceno/internal/model"
)

func TestValidateDuplicateID(t *testing.T) {
	s := model.Spec{
		Nodes: []model.NodeSpec{{ID: "a"}, {ID: "a"}},
		Edges: []model.EdgeSpec{},
	}
	issues := Validate(s)
	if len(issues) == 0 {
		t.Fatal("expected duplicate id error")
	}
	if issues[0].Code != diag.CodeParse {
		t.Fatalf("got %s", issues[0].Code)
	}
}

func TestValidateMissingEdgeNode(t *testing.T) {
	s := model.Spec{
		Nodes: []model.NodeSpec{{ID: "a"}},
		Edges: []model.EdgeSpec{{From: "a", To: "missing"}},
	}
	issues := Validate(s)
	if len(issues) == 0 {
		t.Fatal("expected missing node")
	}
	if issues[0].Code != diag.CodeMissingNode {
		t.Fatalf("got %s", issues[0].Code)
	}
}

func TestValidateParentMustBeContainer(t *testing.T) {
	s := model.Spec{Nodes: []model.NodeSpec{
		{ID: "parent", Kind: model.ShapeBox},
		{ID: "child", Kind: model.ShapeBox, Parent: "parent"},
	}}
	issues := Validate(s)
	if len(issues) == 0 || issues[0].Code != diag.CodeParse {
		t.Fatalf("expected invalid parent error, got %+v", issues)
	}
}

func TestValidateRejectsIndirectParentCycle(t *testing.T) {
	s := model.Spec{Nodes: []model.NodeSpec{
		{ID: "a", Kind: model.ShapeFrame, Parent: "b"},
		{ID: "b", Kind: model.ShapeFrame, Parent: "a"},
	}}
	issues := Validate(s)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "parent cycle") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected parent cycle error, got %+v", issues)
	}
}

func TestValidateRejectsPartialFixedPosition(t *testing.T) {
	x := 10.0
	s := model.Spec{Layout: model.LayoutHybrid, Nodes: []model.NodeSpec{{ID: "a", Kind: model.ShapeBox, X: &x}}}
	issues := Validate(s)
	if len(issues) != 1 || issues[0].Code != diag.CodeMissingPos {
		t.Fatalf("issues = %+v", issues)
	}
}

func TestValidateRejectsUnknownLayout(t *testing.T) {
	s := model.Spec{Layout: "magic", Nodes: []model.NodeSpec{{ID: "a", Kind: model.ShapeBox}}}
	issues := Validate(s)
	if len(issues) != 1 || !strings.Contains(issues[0].Message, "unknown layout") {
		t.Fatalf("issues = %+v", issues)
	}
}
