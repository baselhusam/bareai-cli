package render

import (
	"fmt"
	"io"
	"time"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// WriteProbe renders smoke probe results.
func WriteProbe(w io.Writer, snap *snapshot.Snapshot, noColor bool) error {
	if snap == nil {
		return fmt.Errorf("snapshot is nil")
	}

	label := func(s string) string {
		if noColor {
			return s
		}
		return s
	}

	if _, err := fmt.Fprintf(w, "%s\n", label("bareai probe")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Collected: %s\n\n", snap.CollectedAt.Format(time.RFC3339)); err != nil {
		return err
	}

	if len(snap.LLMs) == 0 {
		if _, err := fmt.Fprintln(w, "No endpoints probed."); err != nil {
			return err
		}
		return writeSkipped(w, snap.Skipped)
	}

	if _, err := fmt.Fprintf(w, "  %-28s %-10s %-8s %s\n", "ENDPOINT", "RUNTIME", "RESULT", "LATENCY"); err != nil {
		return err
	}
	for _, llm := range snap.LLMs {
		result := llm.Probe
		if result == nil && llm.Health != nil {
			result = llm.Health
		}
		status := "fail"
		latency := "-"
		detail := ""
		if result != nil {
			if result.OK {
				status = "pass"
			}
			latency = fmt.Sprintf("%dms", result.LatencyMS)
			if result.Message != "" {
				detail = result.Message
			}
			if !result.OK && result.Error != "" {
				detail = result.Error
			}
		}
		if _, err := fmt.Fprintf(w, "  %-28s %-10s %-8s %s",
			truncate(llm.Endpoint, 28),
			truncate(llm.Runtime, 10),
			status,
			latency,
		); err != nil {
			return err
		}
		if detail != "" {
			if _, err := fmt.Fprintf(w, "  %s", detail); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	return writeSkipped(w, snap.Skipped)
}
