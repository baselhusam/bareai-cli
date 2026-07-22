package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func overviewText(snap *snapshot.Snapshot, width int) string {
	if snap == nil {
		return "Loading..."
	}

	var b strings.Builder
	fmt.Fprintln(&b, "Overview")

	host := "unavailable"
	if snap.Host != nil && snap.Host.Hostname != "" {
		host = snap.Host.Hostname
	}
	gpuCount := len(snap.GPUs)
	docker := "not available"
	if snap.Docker != nil && snap.Docker.Available {
		running := 0
		for _, c := range snap.Docker.Containers {
			if strings.EqualFold(c.State, "running") {
				running++
			}
		}
		docker = fmt.Sprintf("%d running", running)
	}
	fmt.Fprintf(&b, "  Host: %s   GPUs: %d   Docker: %s   LLMs: %d   DBs: %d\n\n",
		host, gpuCount, docker, len(snap.LLMs), len(snap.Databases))

	if snap.Host != nil {
		h := snap.Host
		fmt.Fprintf(&b, "Host\n")
		fmt.Fprintf(&b, "  Hostname:  %s\n", h.Hostname)
		fmt.Fprintf(&b, "  OS:        %s %s (%s)\n", h.OS, h.PlatformVer, h.Platform)
		fmt.Fprintf(&b, "  CPU:       %s (%d cores)\n", h.CPUModel, h.CPUCores)
		fmt.Fprintf(&b, "  Memory:    %s / %s\n", formatBytes(h.MemUsed), formatBytes(h.MemTotal))
		fmt.Fprintf(&b, "  Load:      %.2f %.2f %.2f\n\n", h.Load1, h.Load5, h.Load15)
	}

	b.WriteString(correlationText(snap.Correlations, width))
	b.WriteString("\n")

	if len(snap.Findings) > 0 {
		fmt.Fprintln(&b, "Findings (top)")
		limit := 3
		if len(snap.Findings) < limit {
			limit = len(snap.Findings)
		}
		for _, f := range snap.Findings[:limit] {
			severity := f.Severity
			if severity == "" {
				severity = "info"
			}
			fmt.Fprintf(&b, "  [%s] %s: %s\n", severity, f.ID, f.Summary)
		}
		if len(snap.Findings) > limit {
			fmt.Fprintf(&b, "  … %d more (bareai doctor)\n", len(snap.Findings)-limit)
		}
		b.WriteString("\n")
	}

	if len(snap.Skipped) > 0 {
		fmt.Fprintln(&b, "Skipped")
		for _, skip := range snap.Skipped {
			fmt.Fprintf(&b, "  - %s: %s\n", skip.Component, skip.Reason)
		}
	}

	return strings.TrimRight(b.String(), "\n")
}

func correlationText(rows []snapshot.Correlation, width int) string {
	var b strings.Builder
	fmt.Fprintln(&b, "Correlation")
	if len(rows) == 0 {
		fmt.Fprintln(&b, "  none")
		return b.String()
	}

	showModels := width >= 80
	if showModels {
		fmt.Fprintf(&b, "  %-24s %-8s %-12s %-5s %-3s %-8s %s\n",
			"ENDPOINT", "RUNTIME", "CONTAINER", "PID", "GPU", "VRAM", "MODELS")
	} else {
		fmt.Fprintf(&b, "  %-18s %-8s %-10s %-5s %-3s %s\n",
			"ENDPOINT", "RUNTIME", "CONTAINER", "PID", "GPU", "VRAM")
	}

	for _, row := range rows {
		pid := "-"
		if row.PID > 0 {
			pid = fmt.Sprintf("%d", row.PID)
		}
		gpu := "-"
		if row.GPUIndex != nil {
			gpu = fmt.Sprintf("%d", *row.GPUIndex)
		}
		vram := "-"
		if row.VRAMBytes > 0 {
			vram = formatBytes(row.VRAMBytes)
		}
		container := row.ContainerName
		if container == "" {
			container = "-"
		}
		models := strings.Join(row.Models, ", ")
		if models == "" {
			models = "-"
		}

		if showModels {
			fmt.Fprintf(&b, "  %-24s %-8s %-12s %-5s %-3s %-8s %s\n",
				truncate(row.Endpoint, 24),
				truncate(row.Runtime, 8),
				truncate(container, 12),
				pid, gpu, vram,
				truncate(models, width-70),
			)
		} else {
			fmt.Fprintf(&b, "  %-18s %-8s %-10s %-5s %-3s %s\n",
				truncate(row.Endpoint, 18),
				truncate(row.Runtime, 8),
				truncate(container, 10),
				pid, gpu, vram,
			)
		}
	}
	return b.String()
}

func probePanelText(snap *snapshot.Snapshot, llmIndex int, result *snapshot.ProbeResult, probing bool, s styles) string {
	var b strings.Builder
	fmt.Fprintln(&b, "Probe")

	if snap == nil || len(snap.LLMs) == 0 {
		fmt.Fprintln(&b, "  No LLM endpoints discovered.")
		fmt.Fprintln(&b, "  Press p to probe once endpoints appear.")
		return strings.TrimRight(b.String(), "\n")
	}

	if llmIndex < 0 || llmIndex >= len(snap.LLMs) {
		llmIndex = 0
	}
	llm := snap.LLMs[llmIndex]
	fmt.Fprintf(&b, "  Target: %s (%s)\n", llm.Endpoint, llm.Runtime)
	fmt.Fprintf(&b, "  %s\n\n", llmSourceLabel(llm))

	if probing {
		fmt.Fprintln(&b, "  Probing...")
		return strings.TrimRight(b.String(), "\n")
	}

	display := result
	if display == nil {
		display = llm.Probe
	}
	if display == nil && llm.Health != nil {
		display = llm.Health
	}

	if display == nil {
		fmt.Fprintln(&b, "  No probe result yet. Press p to run smoke test.")
		return strings.TrimRight(b.String(), "\n")
	}

	status := s.fail.Render("FAIL")
	if display.OK {
		status = s.ok.Render("PASS")
	}
	fmt.Fprintf(&b, "  Result:   %s\n", status)
	fmt.Fprintf(&b, "  Latency:  %dms\n", display.LatencyMS)
	if display.Status > 0 {
		fmt.Fprintf(&b, "  HTTP:     %d\n", display.Status)
	}
	if display.Message != "" {
		fmt.Fprintf(&b, "  Message:  %s\n", display.Message)
	}
	if !display.OK && display.Error != "" {
		fmt.Fprintf(&b, "  Error:    %s\n", display.Error)
	}

	fmt.Fprintf(&b, "\n  Collected: %s", snap.CollectedAt.Format(time.RFC3339))
	return strings.TrimRight(b.String(), "\n")
}
