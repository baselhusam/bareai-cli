package inspect

import "github.com/baselhusam/bareai-cli/internal/snapshot"

// Enrich populates correlation rows and informational findings on a snapshot.
func Enrich(snap *snapshot.Snapshot) {
	if snap == nil {
		return
	}
	snap.Correlations = BuildCorrelations(snap)
	snap.Findings = AnalyzeFindings(snap)
}
