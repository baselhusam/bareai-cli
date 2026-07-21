package rules

import (
	"fmt"
	"strings"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// Skipped returns findings for collectors that could not run.
func Skipped(snap *snapshot.Snapshot) []snapshot.Finding {
	if snap == nil {
		return nil
	}
	var out []snapshot.Finding
	for _, skip := range snap.Skipped {
		component := skip.Component
		try := skipTryHint(component)
		out = append(out, finding(
			"skip.collector_failed",
			SeverityInfo,
			"probe",
			rankInfo+30,
			fmt.Sprintf("Collector %s skipped: %s", component, skip.Reason),
			"Partial snapshot; some subsystems were unavailable.",
			try,
		))
	}
	return out
}

func skipTryHint(component string) string {
	base := strings.ToLower(component)
	switch {
	case strings.HasPrefix(base, "docker"):
		return "docker ps  ·  bareai docker --json"
	case strings.HasPrefix(base, "gpu"):
		return "nvidia-smi  ·  bareai gpu --json"
	case strings.HasPrefix(base, "llm"):
		return "bareai llm --json"
	case base == "host":
		return "bareai status --json"
	default:
		return "bareai inspect --json"
	}
}
