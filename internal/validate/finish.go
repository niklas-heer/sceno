package validate

import "github.com/niklas-heer/sceno/internal/diag"

func finishReport(report *diag.Report) {
	report.Enrich()
	recs := diag.BuildRecommendations(*report)
	report.Recommendations = recs
	report.EnrichRecommendations(recs)
}
