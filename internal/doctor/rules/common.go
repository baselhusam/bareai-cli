package rules

import (
	"fmt"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

const (
	SeverityCritical = "critical"
	SeverityWarn     = "warn"
	SeverityInfo     = "info"
)

const (
	rankCritical = 10
	rankWarn     = 110
	rankInfo     = 210
)

func formatBytes(n uint64) string {
	const (
		giB = 1024 * 1024 * 1024
		miB = 1024 * 1024
	)
	switch {
	case n >= giB:
		return fmt.Sprintf("%.1f GiB", float64(n)/float64(giB))
	case n >= miB:
		return fmt.Sprintf("%.1f MiB", float64(n)/float64(miB))
	default:
		return fmt.Sprintf("%d B", n)
	}
}

func finding(id, severity, category string, rank int, summary, why, try string) snapshot.Finding {
	return snapshot.Finding{
		ID:       id,
		Severity: severity,
		Category: category,
		Rank:     rank,
		Summary:  summary,
		Why:      why,
		Try:      try,
	}
}
