package spec

import (
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
