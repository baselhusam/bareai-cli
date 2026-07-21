package rules

import (
	"fmt"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// Host returns host-related findings.
func Host(snap *snapshot.Snapshot) []snapshot.Finding {
	if snap == nil || snap.Host == nil {
		return nil
	}
	var out []snapshot.Finding
	h := snap.Host

	if h.MemTotal > 0 {
		usedPct := float64(h.MemUsed) / float64(h.MemTotal)
		if usedPct > 0.9 {
			out = append(out, finding(
				"host.mem_high",
				SeverityWarn,
				"host",
				rankWarn+5,
				fmt.Sprintf("Host memory usage above 90%% (%s / %s)", formatBytes(h.MemUsed), formatBytes(h.MemTotal)),
				"Low free RAM can cause OOM kills for LLM workloads and containers.",
				"bareai status --json  ·  free -h",
			))
		}
	}

	for _, d := range h.Disks {
		if d.Total == 0 {
			continue
		}
		usedPct := float64(d.Used) / float64(d.Total)
		if usedPct > 0.9 {
			out = append(out, finding(
				"host.disk_low",
				SeverityWarn,
				"host",
				rankWarn+10,
				fmt.Sprintf("Disk %s above 90%% used (%s / %s)", d.Mount, formatBytes(d.Used), formatBytes(d.Total)),
				"Model weights and Docker layers often fill data mounts quickly.",
				fmt.Sprintf("df -h %s  ·  bareai status --json", d.Mount),
			))
		}
	}

	if h.CPUCores > 0 && h.Load1 > 0 && h.Load1 > 2*float64(h.CPUCores) {
		out = append(out, finding(
			"host.load_high",
			SeverityInfo,
			"host",
			rankInfo+5,
			fmt.Sprintf("Load average (%.2f) exceeds 2× CPU cores (%d)", h.Load1, h.CPUCores),
			"High load may indicate CPU-bound inference or competing jobs.",
			"uptime  ·  bareai status --json",
		))
	}

	return out
}
