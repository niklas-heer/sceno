package diag

import "testing"

func TestEnrichOK(t *testing.T) {
	r := Report{OK: true, Input: "sceno.kdl", Stats: Stats{Nodes: 3, Edges: 2}}
	r.Enrich()
	if !r.Agent.RenderReady || len(r.Agent.NextSteps) == 0 {
		t.Fatalf("agent: %+v", r.Agent)
	}
}

func TestEnrichErrors(t *testing.T) {
	r := Report{
		OK: false,
		Input: "bad.kdl",
		Errors: []Issue{{
			Code:    CodeMissingNode,
			Message: "edge references unknown node \"x\"",
		}},
	}
	r.Enrich()
	if r.Agent.Summary == "" || len(r.Agent.NextSteps) < 2 {
		t.Fatalf("agent: %+v", r.Agent)
	}
	if r.Errors[0].Fix == "" {
		t.Fatal("expected fix from catalog")
	}
}
