package render

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// WriteStatus renders a human-readable status report.
func WriteStatus(w io.Writer, snap *snapshot.Snapshot, noColor bool) error {
	if snap == nil {
		return fmt.Errorf("snapshot is nil")
	}

	label := func(s string) string {
		if noColor {
			return s
		}
		return s
	}

	if _, err := fmt.Fprintf(w, "%s\n", label("bareai status")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Collected: %s\n\n", snap.CollectedAt.Format(time.RFC3339)); err != nil {
		return err
	}

	if snap.Host != nil {
		if err := writeHostSection(w, snap.Host); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintln(w, "Host: unavailable"); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if err := writeGPUSummary(w, snap.GPUs); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "Docker:      not collected yet (Phase 3)"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "LLM runtimes: not collected yet (Phase 4)"); err != nil {
		return err
	}

	if len(snap.Skipped) > 0 {
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "Skipped:"); err != nil {
			return err
		}
		for _, skip := range snap.Skipped {
			if _, err := fmt.Fprintf(w, "  - %s: %s\n", skip.Component, skip.Reason); err != nil {
				return err
			}
		}
	}

	return nil
}

func writeHostSection(w io.Writer, h *snapshot.Host) error {
	if _, err := fmt.Fprintln(w, "Host"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  Hostname:  %s\n", h.Hostname); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  OS:        %s %s (%s)\n", h.OS, h.PlatformVer, h.Platform); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  Arch:      %s\n", h.Arch); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  Uptime:    %s\n", formatDuration(h.Uptime)); err != nil {
		return err
	}

	cpuLabel := h.CPUModel
	if cpuLabel == "" {
		cpuLabel = "unknown"
	}
	if _, err := fmt.Fprintf(w, "  CPU:       %s (%d cores, %d logical)\n", cpuLabel, h.CPUCores, h.CPULogical); err != nil {
		return err
	}

	if hasLoad(h) {
		if _, err := fmt.Fprintf(w, "  Load:      %.2f / %.2f / %.2f (1/5/15 min)\n", h.Load1, h.Load5, h.Load15); err != nil {
			return err
		}
	}

	memPct := float64(0)
	if h.MemTotal > 0 {
		memPct = float64(h.MemUsed) / float64(h.MemTotal) * 100
	}
	if _, err := fmt.Fprintf(w, "  Memory:    %s / %s (%.0f%% used, %s available)\n",
		formatBytes(h.MemUsed), formatBytes(h.MemTotal), memPct, formatBytes(h.MemAvailable)); err != nil {
		return err
	}

	if len(h.Disks) == 0 {
		if _, err := fmt.Fprintln(w, "  Disks:     none reported"); err != nil {
			return err
		}
		return nil
	}

	if _, err := fmt.Fprintln(w, "  Disks:"); err != nil {
		return err
	}
	for _, d := range h.Disks {
		pct := float64(0)
		if d.Total > 0 {
			pct = float64(d.Used) / float64(d.Total) * 100
		}
		mount := d.Mount
		if d.FSType != "" {
			mount = fmt.Sprintf("%s (%s)", d.Mount, d.FSType)
		}
		if _, err := fmt.Fprintf(w, "    %-24s %s / %s (%.0f%% used)\n",
			mount, formatBytes(d.Used), formatBytes(d.Total), pct); err != nil {
			return err
		}
	}

	return nil
}

func hasLoad(h *snapshot.Host) bool {
	return h.Load1 != 0 || h.Load5 != 0 || h.Load15 != 0
}

func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}

	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	parts := make([]string, 0, 3)
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}

	return strings.Join(parts, " ")
}
