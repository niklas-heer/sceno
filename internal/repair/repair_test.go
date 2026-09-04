package repair

import (
	"reflect"
	"strings"
	"testing"

	"github.com/niklas-heer/sceno/internal/diag"
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/pipeline"
	"github.com/niklas-heer/sceno/internal/validate"
)

const pairSource = `diagram layout=free gap=20 {
  shape box a "A" x=0 y=150 w=100 h=80
  shape box b "B" x=60 y=150 w=100 h=80
}
`

func collisionRequest(t *testing.T, source string, slide, index int) Request {
	t.Helper()
	_, report, err := validate.EvaluateSource([]byte(source), "preview.kdl", validate.Options{FixCollisions: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, issue := range report.Errors {
		if issue.Code == diag.CodeCollision && issue.SlideIndex == slide {
			p := issue.Repairs[index]
			return Request{SlideIndex: slide, Target: p.Target, Properties: p.Properties, IssueID: IssueID(issue)}
		}
	}
	t.Fatalf("no collision: %+v", report)
	return Request{}
}

func TestPreviewVerifiedCollisionFix(t *testing.T) {
	req := collisionRequest(t, pairSource, 1, 0)
	preview, err := Preview([]byte(pairSource), req)
	if err != nil {
		t.Fatal(err)
	}
	if !preview.Applicable || !preview.AfterReport.OK {
		t.Fatalf("repair rejected: %+v", preview)
	}
	if preview.BeforeSource != pairSource || preview.AfterSource == pairSource || len(preview.Moves) != 1 || preview.Moves[0].Target != "b" {
		t.Fatalf("missing measured changes: %+v", preview)
	}
}

func TestPreviewRejectsNewCollision(t *testing.T) {
	source := strings.Replace(pairSource, "}\n", "  shape box c \"C\" x=190 y=150 w=100 h=80\n}\n", 1)
	req := collisionRequest(t, source, 1, 0)
	preview, err := Preview([]byte(source), req)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Applicable || !strings.Contains(strings.Join(preview.Reasons, " "), "structural") {
		t.Fatalf("new collision should reject: %+v", preview)
	}
}

func TestPreviewRejectsWorsenedExistingCollision(t *testing.T) {
	// a-b overlap by 40px and b-c by 10px. Moving b right to repair a-b
	// increases b-c overlap to 60px without adding a new semantic finding.
	source := strings.Replace(pairSource, "}\n", "  shape box c \"C\" x=150 y=150 w=100 h=80\n}\n", 1)
	req := collisionRequest(t, source, 1, 0)
	preview, err := Preview([]byte(source), req)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Applicable || !strings.Contains(strings.Join(preview.Reasons, " "), "increases an existing structural overlap area") {
		t.Fatalf("worsened existing collision accepted: applicable=%v reasons=%v", preview.Applicable, preview.Reasons)
	}
	area := func(report diag.Report) float64 {
		for _, issue := range report.Errors {
			if issue.Code == diag.CodeCollision && reflect.DeepEqual(issue.Nodes, []string{"b", "c"}) {
				return issue.Geometry.Overlap.W * issue.Geometry.Overlap.H
			}
		}
		t.Fatal("expected b-c collision in both reports")
		return 0
	}
	if area(preview.AfterReport) <= area(preview.BeforeReport) {
		t.Fatal("regression fixture must worsen the already-existing collision")
	}
	for _, issue := range preview.AfterReport.Errors {
		if issue.Code == diag.CodeCollision && reflect.DeepEqual(issue.Nodes, []string{"a", "b"}) {
			t.Fatal("regression fixture must resolve the targeted a-b collision")
		}
	}
}

func TestPreviewAllowsUnchangedUnrelatedCollisionOnSameSlide(t *testing.T) {
	source := strings.Replace(pairSource, "}\n", "  shape box c \"C\" x=500 y=150 w=100 h=80\n  shape box d \"D\" x=560 y=150 w=100 h=80\n}\n", 1)
	preview, err := Preview([]byte(source), collisionRequest(t, source, 1, 0))
	if err != nil {
		t.Fatal(err)
	}
	if !preview.Applicable || preview.AfterReport.OK {
		t.Fatalf("unchanged unrelated collision should remain repairable: applicable=%v reasons=%v", preview.Applicable, preview.Reasons)
	}
}

func TestStructuralEvidenceRejectsSeverityEscalationAndUnknownGeometryChanges(t *testing.T) {
	issue := diag.Issue{SlideIndex: 1, Code: diag.CodeEdgeHidden, Message: "connector is hidden", Edge: []string{"a", "b"}}
	prior := diag.Report{Warnings: []diag.Issue{issue}}
	escalated := diag.Report{Errors: []diag.Issue{issue}}
	if reasons := structuralRegressions(prior, escalated, pipeline.Result{}, pipeline.Result{}); len(reasons) != 1 || !strings.Contains(reasons[0], "escalates") {
		t.Fatalf("warning-to-error escalation must reject: %v", reasons)
	}
	if reasons := structuralRegressions(escalated, prior, pipeline.Result{}, pipeline.Result{}); len(reasons) != 0 {
		t.Fatalf("severity improvement with unchanged evidence should pass: %v", reasons)
	}
	// Route changes can preserve the edge's stack bounding box. Compare exact
	// route points too when the finding itself supplies no geometry.
	before := pipeline.Result{Slides: []pipeline.SlideResult{{Diagram: model.Diagram{Routed: []model.RoutedEdge{{Key: "a-b", Points: [][]float64{{0, 0}, {100, 0}, {100, 100}}}}}}}}
	after := pipeline.Result{Slides: []pipeline.SlideResult{{Diagram: model.Diagram{Routed: []model.RoutedEdge{{Key: "a-b", Points: [][]float64{{0, 0}, {0, 100}, {100, 100}}}}}}}}
	if reasons := structuralRegressions(prior, prior, before, before); len(reasons) != 0 {
		t.Fatalf("unchanged unrelated evidence should pass: %v", reasons)
	}
	want := structuralRegressions(prior, prior, before, after)
	if len(want) != 1 || !strings.Contains(want[0], "non-worsening cannot be verified") {
		t.Fatalf("changed geometry behind unchanged prose must reject: %v", want)
	}
	for i := 0; i < 10; i++ {
		if got := structuralRegressions(prior, prior, before, after); !reflect.DeepEqual(got, want) {
			t.Fatalf("non-deterministic rejection reasons: got=%v want=%v", got, want)
		}
	}
}

func TestStructuralEvidenceMatchesOccurrencesWithoutReusingPriorFindings(t *testing.T) {
	issue := diag.Issue{SlideIndex: 1, Code: diag.CodeCollision, Nodes: []string{"a", "b"}, Geometry: &diag.Geometry{Overlap: model.Rect{W: 10, H: 10}}}
	other := issue
	other.Geometry = &diag.Geometry{Overlap: model.Rect{W: 20, H: 10}}
	prior := diag.Report{Errors: []diag.Issue{issue}, Warnings: []diag.Issue{other}}
	after := diag.Report{Errors: []diag.Issue{issue, other}}
	if reasons := structuralRegressions(prior, after, pipeline.Result{}, pipeline.Result{}); len(reasons) != 1 || !strings.Contains(reasons[0], "escalates") {
		t.Fatalf("same participant counts hide escalation: %v", reasons)
	}
	// Equal counts cannot justify replacing distinct evidence with duplicates.
	after = diag.Report{Errors: []diag.Issue{issue, issue}}
	if reasons := structuralRegressions(prior, after, pipeline.Result{}, pipeline.Result{}); len(reasons) != 1 {
		t.Fatalf("one prior occurrence authorized two findings: %v", reasons)
	}
}

func TestPreviewRejectsStaleUnofferedAndOverlap(t *testing.T) {
	for _, kind := range []string{"stale", "unoffered", "overlap"} {
		t.Run(kind, func(t *testing.T) {
			req := collisionRequest(t, pairSource, 1, 0)
			switch kind {
			case "stale":
				req.IssueID = "old"
			case "unoffered":
				req.Properties = map[string]string{"x": "999"}
			case "overlap":
				req = collisionRequest(t, pairSource, 1, 2)
			}
			preview, err := Preview([]byte(pairSource), req)
			if err != nil {
				t.Fatal(err)
			}
			if preview.Applicable || preview.AfterSource != pairSource {
				t.Fatalf("invalid request changed source: %+v", preview)
			}
		})
	}
}

func TestPreviewScopesRepeatedIDsAndAllowsUnrelatedErrors(t *testing.T) {
	slideBody := strings.TrimSuffix(strings.TrimPrefix(pairSource, "diagram layout=free gap=20 {\n"), "}\n")
	source := "diagram layout=free gap=20 {\n slide \"One\" {\n" + slideBody + " }\n slide \"Two\" {\n" + slideBody + " }\n}\n"
	req := collisionRequest(t, source, 2, 0)
	preview, err := Preview([]byte(source), req)
	if err != nil {
		t.Fatal(err)
	}
	if !preview.Applicable || preview.AfterReport.OK {
		t.Fatalf("unrelated slide collision should remain without blocking local repair: %+v", preview)
	}
	if len(preview.Moves) != 1 || preview.Moves[0].SlideIndex != 2 {
		t.Fatalf("repair leaked slides: %+v", preview.Moves)
	}
}

func TestPreviewRejectsEvidenceFromEarlierSource(t *testing.T) {
	req := collisionRequest(t, pairSource, 1, 0)
	changed := strings.Replace(pairSource, "x=60", "x=61", 1)
	preview, err := Preview([]byte(changed), req)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Applicable || preview.AfterSource != changed || !strings.Contains(strings.Join(preview.Reasons, " "), "stale") {
		t.Fatalf("stale geometry authorized a repair: %+v", preview)
	}
}

func TestPreviewRejectsAmbiguousSourceProperty(t *testing.T) {
	source := strings.Replace(pairSource, "x=60", "x=59 x=60", 1)
	req := collisionRequest(t, source, 1, 0)
	preview, err := Preview([]byte(source), req)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Applicable || preview.AfterSource != source || !strings.Contains(strings.Join(preview.Reasons, " "), "ambiguous") {
		t.Fatalf("ambiguous source was edited: %+v", preview)
	}
}

func TestIssueIDIgnoresDisplayLocationsButKeepsSlideProvenance(t *testing.T) {
	a := diag.Issue{Code: diag.CodeCollision, SlideIndex: 1, Path: "/actual/project/diagram.kdl", Line: 5, Nodes: []string{"a", "b"}}
	b := a
	b.Path, b.Line = "preview.kdl", 10
	if IssueID(a) != IssueID(b) {
		t.Fatal("display path or line changed geometric issue identity")
	}
	b.SlideIndex = 2
	if IssueID(a) == IssueID(b) {
		t.Fatal("issue identity lost slide provenance")
	}
}

func TestPreviewRejectsUnscopedSlideRequest(t *testing.T) {
	req := collisionRequest(t, pairSource, 1, 0)
	req.SlideIndex = 0
	preview, err := Preview([]byte(pairSource), req)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Applicable || preview.AfterSource != pairSource {
		t.Fatalf("unscoped repair should not guess a slide: %+v", preview)
	}
}
