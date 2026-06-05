package diag

import (
	"encoding/json"
	"io"
	"strings"
)

// Code identifies machine-readable error kinds for AI repair loops.
type Code string

const (
	CodeParse          Code = "parse_error"
	CodeMissingNode    Code = "missing_node"
	CodeMissingPos     Code = "missing_position"
	CodeCollision      Code = "collision"
	CodeEdgeCollision  Code = "edge_collision"
	CodeLayout         Code = "layout_error"
	CodeSuggestCompact Code = "suggest_compact"
	CodeUnknownIcon    Code = "unknown_icon"
	CodeTextOverflow   Code = "text_overflow"
	CodeOccluded       Code = "occluded"
	CodeEdgeHidden     Code = "edge_hidden"
	CodeMisaligned     Code = "misaligned"
	CodeDenseLayout    Code = "dense_layout"
	CodeSparseLayout   Code = "sparse_layout"
	CodeSlideCrowded   Code = "slide_crowded"
	CodeWeakHierarchy  Code = "weak_hierarchy"
	CodeTooManyElements Code = "too_many_elements"
	CodeSuggestAnnotation Code = "suggest_annotation"
	CodeAnnotationBlocks Code = "annotation_blocks"
	CodeArrowDetached    Code = "arrow_detached"
	CodeArrowHidden      Code = "arrow_hidden"
	CodeEdgeLabelChrome  Code = "edge_label_chrome_overlap"
	CodeEdgeLabelOffAxis Code = "edge_label_off_axis"
	CodeEdgeSideMismatch Code = "edge_side_mismatch"
)

// Issue is one actionable problem.
type Issue struct {
	Code    Code     `json:"code"`
	Message string   `json:"message"`
	Fix     string   `json:"fix,omitempty"`
	Example string   `json:"example,omitempty"`
	Path    string   `json:"path,omitempty"`
	Nodes   []string `json:"nodes,omitempty"`
	Edge    []string `json:"edge,omitempty"`
	Line    int      `json:"line,omitempty"`
}

// Report is the full validation result (stdout with --json).
type Report struct {
	OK              bool             `json:"ok"`
	Input           string           `json:"input"`
	Errors          []Issue          `json:"errors"`
	Warnings        []Issue          `json:"warnings"`
	Recommendations []Recommendation `json:"recommendations,omitempty"`
	Stats           Stats            `json:"stats,omitempty"`
	Agent           AgentMeta        `json:"agent,omitempty"`
}

type Stats struct {
	Nodes      int `json:"nodes"`
	Edges      int `json:"edges"`
	Slides     int `json:"slides,omitempty"`
	Collisions int `json:"collisions"`
}

func (r Report) ExitCode() int {
	if r.OK {
		return 0
	}
	return 1
}

func (r Report) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

func (r Report) WriteHuman(w io.Writer) error {
	if r.OK {
		_, err := io.WriteString(w, "ok — ready to render\n")
		if r.Stats.Nodes > 0 || r.Stats.Slides > 0 {
			_, _ = io.WriteString(w, "  nodes: "+itoa(r.Stats.Nodes)+"  edges: "+itoa(r.Stats.Edges))
			if r.Stats.Slides > 0 {
				_, _ = io.WriteString(w, "  slides: "+itoa(r.Stats.Slides))
			}
			_, _ = io.WriteString(w, "\n")
		}
		if len(r.Agent.NextSteps) > 0 {
			_, _ = io.WriteString(w, "  next: "+r.Agent.NextSteps[0]+"\n")
		}
		return err
	}
	_, _ = io.WriteString(w, r.Agent.Summary+"\n\n")
	for _, e := range r.Errors {
		writeIssueHuman(w, "error", e)
	}
	for _, e := range r.Warnings {
		writeIssueHuman(w, "warn", e)
	}
	if len(r.Agent.NextSteps) > 0 {
		_, _ = io.WriteString(w, "\nnext steps:\n")
		for _, s := range r.Agent.NextSteps {
			_, _ = io.WriteString(w, "  • "+s+"\n")
		}
	}
	_, _ = io.WriteString(w, "\nhint: sceno docs guide --json\n")
	return nil
}

func writeIssueHuman(w io.Writer, kind string, e Issue) {
	_, _ = io.WriteString(w, kind+": "+string(e.Code)+": "+e.Message+"\n")
	if e.Fix != "" {
		_, _ = io.WriteString(w, "  fix: "+e.Fix+"\n")
	}
	if e.Example != "" {
		_, _ = io.WriteString(w, "  example:\n")
		for _, line := range strings.Split(e.Example, "\n") {
			_, _ = io.WriteString(w, "    "+line+"\n")
		}
	}
}

