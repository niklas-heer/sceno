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

func TestValidateBoundsGridCoordinates(t *testing.T) {
	for _, placement := range []string{"at=0,-1", "at=2147483647,0", "row=10001", "layer=-1"} {
		source := "diagram {\n shape box a \"A\" " + placement + "\n}\n"
		s, err := LoadKDL([]byte(source))
		if err != nil {
			t.Fatal(err)
		}
		issues := Validate(s)
		if len(issues) == 0 || !strings.Contains(issues[0].Message, "grid coordinates") {
			t.Fatalf("unbounded placement accepted %s: %+v", placement, issues)
		}
	}
	s, err := LoadKDL([]byte("diagram {\n shape box a \"A\" at=10000,10000\n}\n"))
	if err != nil {
		t.Fatal(err)
	}
	if issues := Validate(s); len(issues) != 0 {
		t.Fatalf("documented grid boundary rejected: %+v", issues)
	}
}

func TestValidateRejectsNonfiniteGeometry(t *testing.T) {
	for _, property := range []string{"w=NaN", "h=+Inf", "fontSize=NaN", "fontSize=-Inf", "x=+Inf y=0", "x=0 y=NaN", "dx=NaN", "dy=-Inf"} {
		s, err := LoadKDL([]byte("diagram {\n shape box a \"A\" " + property + "\n}\n"))
		if err != nil {
			t.Fatal(err)
		}
		issues := Validate(s)
		if len(issues) == 0 || !strings.Contains(issues[0].Message, "finite number") {
			t.Fatalf("nonfinite node geometry accepted %s: %+v", property, issues)
		}
	}
	for _, property := range []string{"gap=NaN", "gap=-Inf", "padding=+Inf", "padding=-Inf"} {
		s, err := LoadKDL([]byte("diagram " + property + " {\n shape box a \"A\"\n}\n"))
		if err != nil {
			t.Fatal(err)
		}
		if issues := Validate(s); len(issues) == 0 {
			t.Fatalf("nonfinite diagram geometry accepted %s", property)
		}
	}
}

func TestValidateRejectsHugeFiniteGeometry(t *testing.T) {
	for _, property := range []string{"w=1e308", "h=1000001", "fontSize=-1e308", "x=1e308 y=0", "x=0 y=-1e308", "dx=1e308", "dy=-1e308"} {
		s, err := LoadKDL([]byte("diagram {\n shape box a \"A\" " + property + "\n}\n"))
		if err != nil {
			t.Fatal(err)
		}
		issues := Validate(s)
		if len(issues) == 0 || !strings.Contains(issues[0].Message, "1000000") {
			t.Fatalf("huge finite node geometry accepted %s: %+v", property, issues)
		}
	}
	for _, property := range []string{"gap=1e308", "gap=-1e308", "padding=1000001", "padding=-1e308"} {
		s, err := LoadKDL([]byte("diagram " + property + " {\n shape box a \"A\"\n}\n"))
		if err != nil {
			t.Fatal(err)
		}
		if issues := Validate(s); len(issues) == 0 {
			t.Fatalf("huge finite spacing accepted %s", property)
		}
	}
	s, err := LoadKDL([]byte("diagram gap=1000000 padding=1000000 {\n shape box a \"A\" x=-1000000 y=1000000 w=1000000 h=1000000\n}\n"))
	if err != nil {
		t.Fatal(err)
	}
	if issues := Validate(s); len(issues) != 0 {
		t.Fatalf("bounded geometry rejected: %+v", issues)
	}
}
