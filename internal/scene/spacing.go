package scene

import (
	"container/heap"
	"math"
	"sort"

	"github.com/niklas-heer/sceno/internal/model"
)

// SpacingPairLimit bounds pair feedback even for dense, invalid diagrams.
const SpacingPairLimit = 200

// Insets are signed distances from an enclosing rectangle's sides to an inner
// rectangle. Negative values mean that the inner rectangle extends outside it.
type Insets struct {
	Top    float64 `json:"top"`
	Right  float64 `json:"right"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
}

// SpacingReport measures existing scene geometry; it never re-runs layout.
// Distances use axis-aligned outer bounds, not silhouette or stroke distances.
type SpacingReport struct {
	Units             string        `json:"units"`
	BoundsBasis       string        `json:"bounds_basis"`
	RequiredClearance float64       `json:"required_clearance"`
	PairSelection     string        `json:"pair_selection"`
	PairLimit         int           `json:"pair_limit"`
	TotalPairs        int           `json:"total_pairs"`
	CandidatePairs    int           `json:"candidate_pairs"`
	Truncated         bool          `json:"truncated"`
	Pairs             []NodeSpacing `json:"pairs"`
	Nodes             []NodePadding `json:"nodes"`
	Canvas            CanvasSpacing `json:"canvas"`
}

// NodeSpacing measures one unordered peer pair. Positive axis gaps mean
// separation, zero means touching, and negative means overlapping projections.
// Clearance is met when at least one axis gap >= RequiredClearance, matching
// collision.Find's rectangular margin test; Euclidean Distance is descriptive.
type NodeSpacing struct {
	A                 string      `json:"a"`
	B                 string      `json:"b"`
	ABounds           model.Rect  `json:"a_bounds"`
	BBounds           model.Rect  `json:"b_bounds"`
	GapX              float64     `json:"gap_x"`
	GapY              float64     `json:"gap_y"`
	Distance          float64     `json:"distance"`
	Overlap           *model.Rect `json:"overlap,omitempty"`
	MeetsClearance    bool        `json:"meets_clearance"`
	OverlapAllowed    bool        `json:"overlap_allowed"`
	ViolatesClearance bool        `json:"violates_clearance"`
}

// NodePadding distinguishes outer-shape padding from spare room inside the
// silhouette-safe writable region. ContentBounds is the union of the stack's
// measured icon/text boxes. Absent content/writable bounds produce absent insets.
type NodePadding struct {
	ID              string      `json:"id"`
	Bounds          model.Rect  `json:"bounds"`
	WritableBounds  *model.Rect `json:"writable_bounds,omitempty"`
	ContentBounds   *model.Rect `json:"content_bounds,omitempty"`
	ContentPadding  *Insets     `json:"content_padding,omitempty"`
	WritablePadding *Insets     `json:"writable_padding,omitempty"`
	Parent          string      `json:"parent,omitempty"`
	ParentPadding   *Insets     `json:"parent_padding,omitempty"`
}

// CanvasSpacing measures the union of node outer bounds against the exported
// canvas. Top whitespace can include the title band; it is not a title gap.
type CanvasSpacing struct {
	Bounds      model.Rect  `json:"bounds"`
	NodeBounds  *model.Rect `json:"node_bounds,omitempty"`
	NodeMargins *Insets     `json:"node_margins,omitempty"`
}

// AnalyzeSpacing returns deterministic, bounded quantitative feedback using
// Stack geometry. Peer eligibility follows collision.Find: containers and
// explicit parent-child pairs are excluded, while allowed overlap is measured
// and marked exempt. Parent insets expose containment separately.
func AnalyzeSpacing(d *model.Diagram, stack Stack) SpacingReport {
	out := SpacingReport{
		Units: "px", BoundsBasis: "axis_aligned_node_bounds",
		PairSelection: "nearest peer per node plus all pairs below required clearance; violations first, then distance and ids",
		PairLimit:     SpacingPairLimit, Pairs: []NodeSpacing{}, Nodes: []NodePadding{},
		Canvas: CanvasSpacing{Bounds: stack.Canvas},
	}
	if d == nil {
		return out
	}
	out.RequiredClearance = d.Gap / 2
	items := stack.Project(PlaneLane, PlaneStructure, PlaneAnnotation, PlaneNode)
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	byID := make(map[string]StackItem, len(items))
	for _, item := range items {
		byID[item.ID] = item
	}
	var peers []StackItem
	for _, item := range items {
		padding := NodePadding{ID: item.ID, Bounds: item.Bounds, WritableBounds: item.Writable, Parent: item.Parent}
		for _, content := range item.Content {
			padding.ContentBounds = unionSpacingBounds(padding.ContentBounds, content.Bounds)
		}
		if padding.ContentBounds != nil {
			padding.ContentPadding = measureInsets(item.Bounds, *padding.ContentBounds)
			if item.Writable != nil {
				padding.WritablePadding = measureInsets(*item.Writable, *padding.ContentBounds)
			}
		}
		if parent, ok := byID[item.Parent]; ok && item.Parent != "" {
			padding.ParentPadding = measureInsets(parent.Bounds, item.Bounds)
		}
		out.Nodes = append(out.Nodes, padding)
		out.Canvas.NodeBounds = unionSpacingBounds(out.Canvas.NodeBounds, item.Bounds)
		if !model.IsContainer(model.ShapeKind(item.Kind)) {
			peers = append(peers, item)
		}
	}
	if out.Canvas.NodeBounds != nil {
		out.Canvas.NodeMargins = measureInsets(stack.Canvas, *out.Canvas.NodeBounds)
	}

	// Keep only the bounded best candidate set and one nearest peer per node,
	// rather than allocating the full O(n²) pair matrix.
	nearest := make(map[string]NodeSpacing, len(peers))
	candidates := &spacingCandidates{}
	addCandidate := func(pair NodeSpacing) {
		out.CandidatePairs++
		if candidates.Len() < SpacingPairLimit {
			heap.Push(candidates, pair)
		} else if spacingPriority(pair, (*candidates)[0]) {
			(*candidates)[0] = pair
			heap.Fix(candidates, 0)
		}
	}
	for i, a := range peers {
		for _, b := range peers[i+1:] {
			if a.Parent == b.ID || b.Parent == a.ID {
				continue
			}
			out.TotalPairs++
			pair := measureNodeSpacing(a, b, out.RequiredClearance)
			for _, id := range []string{a.ID, b.ID} {
				if previous, ok := nearest[id]; !ok || nearerSpacing(pair, previous) {
					nearest[id] = pair
				}
			}
			if !pair.MeetsClearance {
				addCandidate(pair)
			}
		}
	}
	seen := make(map[[2]string]bool, len(nearest))
	for _, pair := range nearest {
		key := [2]string{pair.A, pair.B}
		if pair.MeetsClearance && !seen[key] {
			seen[key] = true
			addCandidate(pair)
		}
	}
	out.Pairs = append(out.Pairs, *candidates...)
	sort.Slice(out.Pairs, func(i, j int) bool {
		return spacingPriority(out.Pairs[i], out.Pairs[j])
	})
	out.Truncated = out.CandidatePairs > len(out.Pairs)
	return out
}

// The heap root is the least useful retained pair, so replacement is O(log n).
type spacingCandidates []NodeSpacing

func (h spacingCandidates) Len() int           { return len(h) }
func (h spacingCandidates) Less(i, j int) bool { return spacingPriority(h[j], h[i]) }
func (h spacingCandidates) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *spacingCandidates) Push(v any)        { *h = append(*h, v.(NodeSpacing)) }
func (h *spacingCandidates) Pop() any {
	v := (*h)[len(*h)-1]
	*h = (*h)[:len(*h)-1]
	return v
}

func spacingPriority(a, b NodeSpacing) bool {
	if a.ViolatesClearance != b.ViolatesClearance {
		return a.ViolatesClearance
	}
	return nearerSpacing(a, b)
}

func measureNodeSpacing(a, b StackItem, required float64) NodeSpacing {
	x := math.Max(a.Bounds.X, b.Bounds.X) - math.Min(a.Bounds.Right(), b.Bounds.Right())
	y := math.Max(a.Bounds.Y, b.Bounds.Y) - math.Min(a.Bounds.Bottom(), b.Bounds.Bottom())
	// Use collision.Find's exact expanded-bound comparisons. Comparing the
	// subtracted gaps above can round differently at fractional thresholds,
	// falsely flagging a pair placed at exactly edge+required by layout.
	tooClose := a.Bounds.Right()+required > b.Bounds.X &&
		b.Bounds.Right()+required > a.Bounds.X &&
		a.Bounds.Bottom()+required > b.Bounds.Y &&
		b.Bounds.Bottom()+required > a.Bounds.Y
	p := NodeSpacing{
		A: a.ID, B: b.ID, ABounds: a.Bounds, BBounds: b.Bounds,
		GapX: x, GapY: y, Distance: math.Hypot(math.Max(0, x), math.Max(0, y)),
		MeetsClearance: !tooClose,
		OverlapAllowed: a.AllowOverlap || b.AllowOverlap,
	}
	p.ViolatesClearance = !p.MeetsClearance && !p.OverlapAllowed
	if x < 0 && y < 0 {
		p.Overlap = &model.Rect{X: math.Max(a.Bounds.X, b.Bounds.X), Y: math.Max(a.Bounds.Y, b.Bounds.Y), W: -x, H: -y}
	}
	return p
}

func nearerSpacing(a, b NodeSpacing) bool {
	if a.Distance != b.Distance {
		return a.Distance < b.Distance
	}
	if a.A != b.A {
		return a.A < b.A
	}
	return a.B < b.B
}

func measureInsets(outer, inner model.Rect) *Insets {
	return &Insets{Top: inner.Y - outer.Y, Right: outer.Right() - inner.Right(), Bottom: outer.Bottom() - inner.Bottom(), Left: inner.X - outer.X}
}

func unionSpacingBounds(current *model.Rect, next model.Rect) *model.Rect {
	if current == nil {
		return &next
	}
	x, y := math.Min(current.X, next.X), math.Min(current.Y, next.Y)
	return &model.Rect{X: x, Y: y, W: math.Max(current.Right(), next.Right()) - x, H: math.Max(current.Bottom(), next.Bottom()) - y}
}
