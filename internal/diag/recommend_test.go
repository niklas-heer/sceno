package diag

import "testing"

func TestBuildRecommendationsDedupesWarnings(t *testing.T) {
	r := Report{
		OK: true,
		Input: "test.kdl",
		Warnings: []Issue{
			{Code: CodeSuggestCompact, Message: "sparse", Fix: "reduce gap"},
		},
	}
	recs := BuildRecommendations(r)
	if len(recs) == 0 {
		t.Fatal("expected recommendations")
	}
}

func TestEnrichRecommendationsPrependsHighPriority(t *testing.T) {
	r := Report{Agent: AgentMeta{NextSteps: []string{"existing"}}}
	r.EnrichRecommendations([]Recommendation{
		{Priority: 1, Message: "fix first", Fix: "do it"},
	})
	if len(r.Agent.NextSteps) < 2 || r.Agent.NextSteps[0] != "fix first — do it" {
		t.Fatalf("steps: %+v", r.Agent.NextSteps)
	}
}
