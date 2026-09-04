package preview

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/niklas-heer/sceno/internal/diag"
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/pipeline"
	"github.com/niklas-heer/sceno/internal/render"
	"github.com/niklas-heer/sceno/internal/repair"
	"github.com/niklas-heer/sceno/internal/scene"
	"github.com/niklas-heer/sceno/internal/spec"
	"github.com/niklas-heer/sceno/internal/validate"
)

// Rendering currently uses a shared palette. Keep every preview render and
// export serialized, including separate sessions in the same process.
var renderMu sync.Mutex

type State struct {
	Revision      string    `json:"revision"`
	Filename      string    `json:"filename"`
	Source        string    `json:"source"`
	Valid         bool      `json:"valid"`
	RenderReady   bool      `json:"render_ready"`
	Score         int       `json:"score"`
	Summary       string    `json:"summary"`
	UndoAvailable bool      `json:"undo_available"`
	UpdatedAt     time.Time `json:"updated_at"`
	Slides        []Slide   `json:"slides"`
	Issues        []Issue   `json:"issues"`
	FileError     string    `json:"file_error,omitempty"`
}

type Slide struct {
	Index   int                 `json:"index"`
	Title   string              `json:"title"`
	SVG     string              `json:"svg"`
	Canvas  model.Rect          `json:"canvas"`
	Nodes   []Node              `json:"nodes"`
	Issues  []Issue             `json:"issues"`
	Spacing scene.SpacingReport `json:"spacing"`
}

type Node struct {
	ID     string     `json:"id"`
	Label  string     `json:"label"`
	Kind   string     `json:"kind"`
	Bounds model.Rect `json:"bounds"`
	Line   int        `json:"line"`
	Source string     `json:"source"`
}

type Issue struct {
	ID         string              `json:"id"`
	SlideIndex int                 `json:"slide_index"`
	Severity   string              `json:"severity"`
	Category   string              `json:"category"`
	Code       diag.Code           `json:"code"`
	Message    string              `json:"message"`
	Fix        string              `json:"fix"`
	Line       int                 `json:"line,omitempty"`
	Nodes      []string            `json:"nodes,omitempty"`
	Geometry   *diag.Geometry      `json:"geometry,omitempty"`
	Repairs    []diag.RepairOption `json:"repairs"`
}

func revision(source []byte) string { return fmt.Sprintf("%x", sha256.Sum256(source)) }

func buildState(source []byte, path string) (out State, result pipeline.Result) {
	out = State{
		Revision: revision(source), Filename: filepath.Base(path), Source: string(source),
		UpdatedAt: time.Now().UTC(), Slides: []Slide{}, Issues: []Issue{},
	}
	// The source is editable while this long-lived session is running. A
	// layout/render bug must become visible feedback, never kill the watcher.
	defer func() {
		if problem := recover(); problem != nil {
			out.Slides = []Slide{}
			out.Score = 0
			out.Issues = []Issue{}
			invalidatePreview(&out, fmt.Sprintf("Layout could not be computed: %v. Update the source to continue.", problem))
			result = pipeline.Result{}
		}
	}()
	var report diag.Report
	var err error
	result, report, err = validate.EvaluateSource(source, path, validate.Options{FixCollisions: true})
	out.Valid, out.RenderReady = report.OK, report.OK
	out.Score, out.Summary = result.MergedEval().Score, report.Agent.Summary
	if err != nil && out.Summary == "" {
		out.Summary = err.Error()
	}
	locations, _ := spec.SourceLocations(source)
	locationByNode := make(map[string]spec.SourceLocation, len(locations))
	for _, location := range locations {
		locationByNode[fmt.Sprintf("%d:%s", location.SlideIndex, location.Target)] = location
	}
	severity := make(map[string]string)
	for i, slide := range result.Slides {
		for _, finding := range slide.Eval.Findings {
			severity[fmt.Sprintf("%d:%s:%s", i+1, finding.Code, finding.Message)] = finding.Severity
		}
	}
	appendIssue := func(iss diag.Issue, level string) {
		category := issueCategory(iss.Code)
		if category == "structural" {
			out.RenderReady = false
		}
		item := Issue{ID: repair.IssueID(iss), SlideIndex: iss.SlideIndex, Severity: level,
			Category: category, Code: iss.Code, Message: iss.Message, Fix: iss.Fix,
			Geometry: iss.Geometry, Line: iss.Line, Nodes: iss.Nodes, Repairs: []diag.RepairOption{}}
		for _, option := range iss.Repairs {
			if option.Action == "set_property" && option.Properties["overlap"] == "" {
				item.Repairs = append(item.Repairs, option)
			}
		}
		out.Issues = append(out.Issues, item)
	}
	for _, iss := range report.Errors {
		appendIssue(iss, "error")
	}
	for _, iss := range report.Warnings {
		level := "warning"
		message := iss.Message
		if len(result.Slides) > 1 {
			prefix := fmt.Sprintf("slide %d: ", iss.SlideIndex)
			if len(message) >= len(prefix) && message[:len(prefix)] == prefix {
				message = message[len(prefix):]
			}
		}
		if known := severity[fmt.Sprintf("%d:%s:%s", iss.SlideIndex, iss.Code, message)]; known != "" {
			level = known
		}
		appendIssue(iss, level)
	}
	if out.Valid {
		if out.RenderReady {
			out.Summary = "Ready to export. Optional suggestions are listed separately."
		} else {
			out.Summary = "The source is valid. Resolve structural findings before sharing."
		}
	}
	for i, slide := range result.Slides {
		view := Slide{Index: i + 1, Title: slide.Diagram.Title, Canvas: slide.Eval.SceneStack.Canvas,
			Nodes: []Node{}, Issues: []Issue{}, Spacing: slide.Eval.Spacing}
		// Invalid sources get diagnostics but never a new render. The browser
		// may retain its last valid view with a clearly marked stale state.
		if report.OK {
			view.SVG = renderPreviewSVG(slide.Diagram)
			view.SVG = safeSVG(view.SVG)
			if view.SVG == "" {
				invalidatePreview(&out, fmt.Sprintf("Slide %d could not be displayed safely. Check its labels and paint properties before continuing.", i+1))
			}
		}
		for _, n := range slide.Diagram.Nodes {
			loc := locationByNode[fmt.Sprintf("%d:%s", i+1, n.ID)]
			view.Nodes = append(view.Nodes, Node{ID: n.ID, Label: n.Label, Kind: string(n.Kind), Bounds: n.Rect, Line: loc.Line, Source: loc.Source})
		}
		for _, issue := range out.Issues {
			if issue.SlideIndex == i+1 || (issue.SlideIndex == 0 && i == 0) {
				view.Issues = append(view.Issues, issue)
			}
		}
		out.Slides = append(out.Slides, view)
	}
	if !out.Valid {
		for i := range out.Slides {
			out.Slides[i].SVG = ""
		}
	}
	if !ensureSerializableState(&out) {
		result = pipeline.Result{}
	}
	return out, result
}

// A defensive last boundary keeps unexpected computed NaN/Inf out of the API.
// Preserve the editable source even if an engine regression escapes validation.
func ensureSerializableState(state *State) bool {
	if _, err := json.Marshal(state); err == nil {
		return true
	}
	state.Slides = []Slide{}
	state.Issues = []Issue{}
	state.Score = 0
	invalidatePreview(state, "The computed layout contains invalid geometry. Adjust the source to continue.")
	return false
}

func renderPreviewSVG(diagram model.Diagram) string {
	renderMu.Lock()
	defer renderMu.Unlock()
	return render.PolishedSVG(diagram)
}

func invalidatePreview(state *State, message string) {
	state.Valid, state.RenderReady = false, false
	state.Summary = message
	state.Issues = append(state.Issues, Issue{
		ID: revision([]byte(message)), Severity: "error", Category: "structural",
		Code: diag.CodeLayout, Message: message, Repairs: []diag.RepairOption{},
	})
}

func issueCategory(code diag.Code) string {
	switch code {
	case diag.CodeDenseLayout, diag.CodeSlideCrowded, diag.CodeTooManyElements, diag.CodeWeakHierarchy:
		return "readability"
	case diag.CodeSparseLayout, diag.CodeSuggestCompact, diag.CodeSuggestAnnotation, diag.CodeMisaligned:
		return "style"
	default:
		return "structural"
	}
}
