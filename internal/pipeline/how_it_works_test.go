package pipeline

import (
	"math"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niklas-heer/sceno/internal/diag"
	"github.com/niklas-heer/sceno/internal/geom"
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/scene"
)

func TestHowItWorksSingleRow(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "how-it-works.kdl")
	d, colls, err := Build(path, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	if len(colls) > 0 {
		t.Fatalf("unexpected collisions: %+v", colls)
	}
	cys := map[string]float64{}
	for _, n := range d.Nodes {
		cys[n.ID] = n.Rect.CY()
	}
	first := cys["author"]
	for id, cy := range cys {
		if math.Abs(cy-first) > 1 {
			t.Fatalf("node %q not row-aligned: cy=%.1f first=%.1f all=%v", id, cy, first, cys)
		}
	}
}

func TestHowItWorksEdgeAnchorsOnBorders(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "how-it-works.kdl")
	d, _, err := Build(path, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]model.Node{}
	for _, n := range d.Nodes {
		byID[n.ID] = n
	}
	const eps = 1.5
	for _, re := range d.Routed {
		if len(re.Points) < 2 {
			t.Fatalf("edge %s→%s has no path", re.Edge.From, re.Edge.To)
		}
		a, okA := byID[re.Edge.From]
		b, okB := byID[re.Edge.To]
		if !okA || !okB {
			continue
		}
		fs, ts := re.Edge.FromSide, re.Edge.ToSide
		if fs == "" || fs == model.SideAuto {
			fs, _ = geom.BestSides(a, b)
		}
		if ts == "" || ts == model.SideAuto {
			_, ts = geom.BestSides(a, b)
		}
		start := geom.SlicesToPath(re.Points)[0]
		end := geom.SlicesToPath(re.Points)[len(re.Points)-1]
		wantStart := geom.Anchor(a, fs)
		wantEnd := geom.Anchor(b, ts)
		if math.Hypot(start.X-wantStart.X, start.Y-wantStart.Y) > eps {
			t.Fatalf("edge %s→%s start not on border: got %v want %v", re.Edge.From, re.Edge.To, start, wantStart)
		}
		if math.Hypot(end.X-wantEnd.X, end.Y-wantEnd.Y) > eps {
			t.Fatalf("edge %s→%s end not on border: got %v want %v", re.Edge.From, re.Edge.To, end, wantEnd)
		}
	}
}

func TestHowItWorksNodeSpacing(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "how-it-works.kdl")
	d, _, err := Build(path, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	minGap := d.Gap * 0.5
	if minGap < 8 {
		minGap = 8
	}
	nodes := d.Nodes
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			a, b := nodes[i], nodes[j]
			if a.Rect.Right()+minGap > b.Rect.X && b.Rect.Right()+minGap > a.Rect.X &&
				a.Rect.Bottom()+minGap > b.Rect.Y && b.Rect.Bottom()+minGap > a.Rect.Y {
				t.Fatalf("nodes %q and %q are too close (gap < %.0f)", a.ID, b.ID, minGap)
			}
		}
	}
}

func TestHowItWorksEdgeRenderValidation(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "how-it-works.kdl")
	d, _, err := Build(path, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	ev := scene.Evaluate(&d)
	for _, f := range ev.Findings {
		switch f.Code {
		case string(diag.CodeArrowDetached), string(diag.CodeArrowHidden), string(diag.CodeEdgeLabelChrome), string(diag.CodeEdgeLabelOffAxis):
			t.Fatalf("README diagram should pass edge render rules: %+v", f)
		}
	}
}

func TestHowItWorksIncludesDescribeStep(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "how-it-works.kdl")
	d, _, err := Build(path, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, n := range d.Nodes {
		if n.ID == "describe" && strings.Contains(strings.ToLower(n.Label), "describe") {
			found = true
		}
	}
	if !found {
		t.Fatal("README diagram should include sceno describe step")
	}
}
