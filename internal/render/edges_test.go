package render

import (
	"math"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niklas-heer/sceno/internal/geom"
	"github.com/niklas-heer/sceno/internal/pipeline"
)

func TestLabeledEdgeSingleConnectorPath(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "how-it-works.kdl")
	d, _, err := pipeline.Build(path, pipeline.DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	for _, re := range d.Routed {
		if strings.TrimSpace(re.Edge.Label) == "" {
			continue
		}
		ctx := LabelContext(d, re.Edge)
		segs := edgePathSegments(re.Points, re.Edge, ctx)
		if len(segs) != 1 {
			t.Fatalf("edge %s→%s label=%q: want 1 segment, got %d", re.Edge.From, re.Edge.To, re.Edge.Label, len(segs))
		}
		gpts := geom.SimplifyPath(segs[0])
		if len(gpts) < 2 {
			t.Fatal("empty path")
		}
		length := math.Hypot(gpts[len(gpts)-1].X-gpts[0].X, gpts[len(gpts)-1].Y-gpts[0].Y)
		if length < 40 {
			t.Fatalf("edge %s→%s connector too short (%.1fpx) for arrowhead", re.Edge.From, re.Edge.To, length)
		}
		head := ArrowHeadSVG(re.Points, re.Edge)
		if !strings.Contains(head, "<polygon") {
			t.Fatalf("edge %s→%s: want arrow polygon, got: %s", re.Edge.From, re.Edge.To, head)
		}
	}
}

func TestHowItWorksArrowTipsOnBorders(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "how-it-works.kdl")
	d, _, err := pipeline.Build(path, pipeline.DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	for _, re := range d.Routed {
		gpts := geom.SimplifyPath(geom.SlicesToPath(re.Points))
		ag, ok := geom.ArrowGeometryForPath(gpts)
		if !ok {
			t.Fatalf("no arrow geom for %s→%s", re.Edge.From, re.Edge.To)
		}
		target := gpts[len(gpts)-1]
		if geom.TipGap(ag.Tip, target) > geom.MaxArrowTipGap {
			t.Fatalf("tip should equal path end anchor for %s→%s (gap %.1f)", re.Edge.From, re.Edge.To, geom.TipGap(ag.Tip, target))
		}
	}
}
