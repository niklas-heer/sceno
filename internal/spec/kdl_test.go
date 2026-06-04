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
