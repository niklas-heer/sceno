package spec

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
)

func TestLoadKDL(t *testing.T) {
	data := `diagram title="Test" layout=auto gap=40 {
  shape box a "A" at=0,0
  shape box b "B" at=1,0
  edge a -> b fromSide=right toSide=left
}`
	s, err := LoadKDL([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if s.Title != "Test" || s.Gap != 40 || len(s.Nodes) != 2 || len(s.Edges) != 1 {
		t.Fatalf("unexpected spec: %+v", s)
	}
	if s.Nodes[0].Kind != model.ShapeBox {
		t.Fatalf("expected box, got %s", s.Nodes[0].Kind)
	}
}

func TestLoadKDLHybridControls(t *testing.T) {
	data := `diagram layout=hybrid {
  shape note callout "Context" x=40 y=60 dx=8 dy=-4 overlap=allow
}`
	s, err := LoadKDL([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	n := s.Nodes[0]
	if n.X == nil || n.Y == nil || n.DX != 8 || n.DY != -4 || !n.AllowOverlap {
		t.Fatalf("hybrid controls not parsed: %+v", n)
	}
}

func TestKDLArrowEdge(t *testing.T) {
	data := `diagram {
  shape box a "A" at=0,0
  shape box b "B" at=1,0
  edge a -> b
}`
	s, err := LoadKDL([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Edges) != 1 || s.Edges[0].From != "a" || s.Edges[0].To != "b" {
		t.Fatalf("edge: %+v", s.Edges)
	}
}

func TestKDLEdgeLabel(t *testing.T) {
	data := `diagram {
  shape box a "A" at=0,0
  shape box b "B" at=1,0
  edge a -> b "hello"
  edge b -> a label="world"
}`
	s, err := LoadKDL([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Edges) != 2 {
		t.Fatalf("edges: %d", len(s.Edges))
	}
	if s.Edges[0].Label != "hello" || s.Edges[1].Label != "world" {
		t.Fatalf("labels: %+v", s.Edges)
	}
}

func TestNormalizeShapeAliases(t *testing.T) {
	data := `diagram {
  shape actor dev "Dev" at=0,0
  shape callout tip "Tip" at=1,0
  shape decision chk "?" at=2,0
}`
	s, err := LoadKDL([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if s.Nodes[0].Kind != model.ShapeActor {
		t.Fatalf("actor: %s", s.Nodes[0].Kind)
	}
	if s.Nodes[1].Kind != model.ShapeInfobox {
		t.Fatalf("callout: %s", s.Nodes[1].Kind)
	}
	if s.Nodes[2].Kind != model.ShapeDiamond {
		t.Fatalf("decision: %s", s.Nodes[2].Kind)
	}
}

func TestKDLEscapeNewlineInLabel(t *testing.T) {
	data := `diagram {
  shape box api "API / Git\nTrigger" at=0,0
}`
	s, err := LoadKDL([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if s.Nodes[0].Label != "API / Git\nTrigger" {
		t.Fatalf("label: %q", s.Nodes[0].Label)
	}
}

func TestKDLQuotedPropWithSpaces(t *testing.T) {
	data := `diagram title="Self-Service Platform" subtitle="Pulumi · Policy Pack" layout=auto gap=32 {
  shape callout tip "Tip" subtitle="Platform team owns this" at=0,0
}`
	s, err := LoadKDL([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if s.Title != "Self-Service Platform" {
		t.Fatalf("title: %q", s.Title)
	}
	if s.Subtitle != "Pulumi · Policy Pack" {
		t.Fatalf("subtitle: %q", s.Subtitle)
	}
	if s.Nodes[0].Subtitle != "Platform team owns this" {
		t.Fatalf("node subtitle: %q", s.Nodes[0].Subtitle)
	}
}

func TestKDLSlideBlocks(t *testing.T) {
	data := `diagram title="Deck" slide=16x9 layout=auto gap=32 {
  slide "One" {
    shape box a "A" at=0,0
    shape box b "B" at=1,0
    edge a -> b
  }
  slide "Two" {
    shape box c "C" at=0,0
  }
}`
	s, err := LoadKDL([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if s.SlideAspect != "16x9" || len(s.Slides) != 2 {
		t.Fatalf("slides: aspect=%q n=%d", s.SlideAspect, len(s.Slides))
	}
	if s.Slides[0].Title != "One" || len(s.Slides[0].Nodes) != 2 {
		t.Fatalf("slide1: %+v", s.Slides[0])
	}
}

func TestLoadKDLFixture(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "self-service.kdl")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	s, err := LoadKDL(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Nodes) < 10 {
		t.Fatalf("expected nodes, got %d", len(s.Nodes))
	}
	if s.Title != "Self-Service Infrastructure Platform" {
		t.Fatalf("title: %q", s.Title)
	}
	if s.Subtitle != "Pulumi Components · Policy Pack · Dedicated Runner" {
		t.Fatalf("subtitle: %q", s.Subtitle)
	}
}

func TestLoadSlidesDarkPreservesCompleteCode(t *testing.T) {
	s, err := LoadFile(filepath.Join("..", "..", "examples", "slides-dark.kdl"))
	if err != nil {
		t.Fatal(err)
	}
	want := "package main\n\nimport \"github.com/pulumi/pulumi/sdk/v3/go/pulumi\"\n\nfunc main() {\n\tpulumi.Run(func(ctx *pulumi.Context) error {\n\t\treturn nil\n\t})\n}"
	if len(s.Slides) != 2 || len(s.Slides[1].Nodes) != 2 {
		t.Fatalf("unexpected slides: %+v", s.Slides)
	}
	n := s.Slides[1].Nodes[0]
	if n.Code != want {
		t.Fatalf("code source changed during parsing:\ngot:  %q\nwant: %q", n.Code, want)
	}
	if n.W != 520 || n.H != 200 || !n.AtSet {
		t.Fatalf("properties after code source were lost: %+v", n)
	}
}

func TestKDLQuotedValuesDecodeExactlyOnce(t *testing.T) {
	s, err := LoadKDL([]byte(`diagram title="Quoted \"title\"  with spaces" {
  shape box a "Literal \\n and \"quotes\"" subtitle="https://example.com/a?x=1 // text" at=0,0 // real comment
  code sample source="fmt.Println(\"literal \\n\")\n// code comment\n\treturn nil" at=1,0
}`))
	if err != nil {
		t.Fatal(err)
	}
	if s.Title != "Quoted \"title\"  with spaces" || len(s.Nodes) != 2 {
		t.Fatalf("quoted header or nodes corrupted: %+v", s)
	}
	if s.Nodes[0].Label != `Literal \n and "quotes"` || s.Nodes[0].Subtitle != "https://example.com/a?x=1 // text" {
		t.Fatalf("quoted label/property corrupted: %+v", s.Nodes[0])
	}
	if want := "fmt.Println(\"literal \\n\")\n// code comment\n\treturn nil"; s.Nodes[1].Code != want {
		t.Fatalf("code escaped twice or truncated: got %q, want %q", s.Nodes[1].Code, want)
	}
}

func TestKDLRejectsUnterminatedQuotedValues(t *testing.T) {
	for _, line := range []string{`shape box a "unterminated`, `shape box a "A" subtitle="unterminated`, `code sample source="trailing escape\`} {
		if _, err := LoadKDL([]byte("diagram {\n" + line + "\n}")); err == nil {
			t.Fatalf("unterminated string silently accepted: %s", line)
		}
	}
}
