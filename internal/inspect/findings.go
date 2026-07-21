package inspect

import (
	"github.com/baselhusam/bareai-cli/internal/doctor"
	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// AnalyzeFindings returns informational findings for an inspect snapshot.
func AnalyzeFindings(snap *snapshot.Snapshot) []snapshot.Finding {
	return doctor.Analyze(snap, doctor.Options{})
}
