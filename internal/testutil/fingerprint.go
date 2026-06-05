// Package testutil provides deterministic helpers for property and regression tests.
package testutil

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"sort"

	"github.com/niklas-heer/sceno/internal/geom"
	"github.com/niklas-heer/sceno/internal/model"
)

// DiagramFingerprint returns a stable hash of laid-out geometry (positions, sizes, routes).
func DiagramFingerprint(d model.Diagram) string {
	h := sha256.New()
	fmt.Fprintf(h, "title=%q gap=%.2f\n", d.Title, d.Gap)
	ids := make([]string, len(d.Nodes))
	for i, n := range d.Nodes {
		ids[i] = n.ID
	}
	sort.Strings(ids)
	byID := map[string]model.Node{}
	for _, n := range d.Nodes {
		byID[n.ID] = n
	}
	for _, id := range ids {
		n := byID[id]
		fmt.Fprintf(h, "node %s %.2f,%.2f %.2fx%.2f col=%d row=%d int=%v\n",
			id, n.Rect.X, n.Rect.Y, n.Rect.W, n.Rect.H, n.Column, n.Row, n.Interior.Ready)
		if n.Interior.Ready {
			in := n.Interior
			fmt.Fprintf(h, "  interior icon=%.0f,%.0f title=%.0f\n", in.IconX, in.IconY, in.TitleStartY)
		}
	}
	edgeKeys := make([]string, 0, len(d.Routed))
	for _, re := range d.Routed {
		edgeKeys = append(edgeKeys, re.Key)
	}
	sort.Strings(edgeKeys)
	routeByKey := map[string]model.RoutedEdge{}
	for _, re := range d.Routed {
		routeByKey[re.Key] = re
	}
	for _, key := range edgeKeys {
		re := routeByKey[key]
		fmt.Fprintf(h, "edge %s %s->%s", key, re.Edge.From, re.Edge.To)
		for _, p := range re.Points {
			if len(p) >= 2 {
				fmt.Fprintf(h, " %.1f,%.1f", p[0], p[1])
			}
		}
		fmt.Fprintln(h)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// InteriorInBounds reports whether stored interior coords sit inside the node rect.
func InteriorInBounds(n model.Node) bool {
	if !n.Interior.Ready || n.Icon == "" && n.Label == "" {
		return true
	}
	in := n.Interior
	r := n.Rect
	if in.IconSize > 0 {
		if in.IconX < 0 || in.IconY < 0 ||
			in.IconX+in.IconSize > r.W+0.5 || in.IconY+in.IconSize > r.H+0.5 {
			return false
		}
	}
	if in.TitleStartY < 0 || in.TitleStartY > r.H {
		return false
	}
	if in.HasSubtitle && (in.SubtitleY < 0 || in.SubtitleY > r.H) {
		return false
	}
	return true
}

// AnchorOnBorder reports whether point p lies on side s of rect r (within eps).
func AnchorOnBorder(r model.Rect, p geom.Point, s model.Side, eps float64) bool {
	switch s {
	case model.SideTop:
		return math.Abs(p.Y-r.Y) <= eps && p.X >= r.X-eps && p.X <= r.Right()+eps
	case model.SideBottom:
		return math.Abs(p.Y-r.Bottom()) <= eps && p.X >= r.X-eps && p.X <= r.Right()+eps
	case model.SideLeft:
		return math.Abs(p.X-r.X) <= eps && p.Y >= r.Y-eps && p.Y <= r.Bottom()+eps
	case model.SideRight:
		return math.Abs(p.X-r.Right()) <= eps && p.Y >= r.Y-eps && p.Y <= r.Bottom()+eps
	default:
		return false
	}
}
