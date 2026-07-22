package doctor

import (
	"sort"

	"github.com/baselhusam/bareai-cli/internal/doctor/rules"
	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// Options controls doctor analysis output.
type Options struct {
	MinSeverity string // info | warn | critical
}

// Analyze returns ranked findings for a snapshot.
func Analyze(snap *snapshot.Snapshot, opts Options) []snapshot.Finding {
	if snap == nil {
		return nil
	}

	var findings []snapshot.Finding
	findings = append(findings, rules.Host(snap)...)
	findings = append(findings, rules.GPU(snap)...)
	findings = append(findings, rules.Docker(snap)...)
	findings = append(findings, rules.LLM(snap)...)
	findings = append(findings, rules.DB(snap)...)
	findings = append(findings, rules.EmptyBox(snap)...)
	findings = append(findings, rules.Skipped(snap)...)

	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Rank != findings[j].Rank {
			return findings[i].Rank < findings[j].Rank
		}
		return findings[i].ID < findings[j].ID
	})

	minRank := severityMinRank(opts.MinSeverity)
	if minRank == 0 {
		return findings
	}
	filtered := findings[:0]
	for _, f := range findings {
		if f.Rank < minRank {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

func severityMinRank(severity string) int {
	switch severity {
	case "critical":
		return 100
	case "warn":
		return 200
	default:
		return 0
	}
}
