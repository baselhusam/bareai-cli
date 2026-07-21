package render

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// WriteLLM renders discovered LLM runtimes.
func WriteLLM(w io.Writer, snap *snapshot.Snapshot, noColor bool) error {
	if snap == nil {
		return fmt.Errorf("snapshot is nil")
	}

	label := func(s string) string {
		if noColor {
			return s
		}
		return s
	}

	if _, err := fmt.Fprintf(w, "%s\n", label("bareai llm")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Collected: %s\n\n", snap.CollectedAt.Format(time.RFC3339)); err != nil {
		return err
	}

	if len(snap.LLMs) == 0 {
		if _, err := fmt.Fprintln(w, "No LLM runtimes discovered."); err != nil {
			return err
		}
		return writeSkipped(w, snap.Skipped)
	}

	for i, llm := range snap.LLMs {
		if i > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if err := writeLLMDetail(w, llm); err != nil {
			return err
		}
	}

	return writeSkipped(w, snap.Skipped)
}

func writeLLMDetail(w io.Writer, llm snapshot.LLM) error {
	source := llmSourceLabel(llm)
	if _, err := fmt.Fprintf(w, "%s  %s  (%s)\n", llm.Name, llm.Endpoint, source); err != nil {
		return err
	}
	if llm.Health != nil {
		status := "fail"
		if llm.Health.OK {
			status = "ok"
		}
		line := fmt.Sprintf("  Health: %s  %dms", status, llm.Health.LatencyMS)
		if llm.Health.Message != "" {
			line += "  " + llm.Health.Message
		}
		if !llm.Health.OK && llm.Health.Error != "" {
			line += "  " + llm.Health.Error
		}
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	if len(llm.Models) > 0 {
		names := make([]string, 0, len(llm.Models))
		for _, m := range llm.Models {
			if m.ID != "" {
				names = append(names, m.ID)
			}
		}
		if _, err := fmt.Fprintf(w, "  Models: %s\n", strings.Join(names, ", ")); err != nil {
			return err
		}
	}
	if llm.GPUIndex != nil {
		if _, err := fmt.Fprintf(w, "  GPU: %d\n", *llm.GPUIndex); err != nil {
			return err
		}
	}
	if len(llm.Metrics) > 0 {
		parts := make([]string, 0, len(llm.Metrics))
		for k, v := range llm.Metrics {
			parts = append(parts, fmt.Sprintf("%s=%.0f", k, v))
		}
		if _, err := fmt.Fprintf(w, "  Metrics: %s\n", strings.Join(parts, ", ")); err != nil {
			return err
		}
	}
	return nil
}

func llmSourceLabel(llm snapshot.LLM) string {
	switch llm.Source {
	case "docker":
		if llm.ContainerName != "" {
			return "docker: " + llm.ContainerName
		}
		return "docker"
	case "process":
		if llm.PID > 0 {
			return fmt.Sprintf("process pid %d", llm.PID)
		}
		return "process"
	case "port":
		return "port scan"
	default:
		return llm.Source
	}
}

func writeLLMSummary(w io.Writer, llms []snapshot.LLM) error {
	if len(llms) == 0 {
		if _, err := fmt.Fprintln(w, "LLM:         none discovered"); err != nil {
			return err
		}
		return nil
	}

	healthy := 0
	runtimes := make(map[string]bool)
	for _, llm := range llms {
		if llm.Runtime != "" {
			runtimes[strings.ToLower(llm.Runtime)] = true
		}
		if llm.Health != nil && llm.Health.OK {
			healthy++
		}
	}
	names := make([]string, 0, len(runtimes))
	for name := range runtimes {
		names = append(names, name)
	}
	sortStrings(names)

	line := fmt.Sprintf("LLM:         %d runtimes (%s)", len(llms), strings.Join(names, ", "))
	if healthy > 0 {
		line += fmt.Sprintf(" — %d healthy", healthy)
	}
	if _, err := fmt.Fprintln(w, line); err != nil {
		return err
	}
	return nil
}

func sortStrings(s []string) {
	for i := 0; i < len(s); i++ {
		for j := i + 1; j < len(s); j++ {
			if s[j] < s[i] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}
