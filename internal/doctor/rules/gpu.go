package rules

import (
	"fmt"
	"strings"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// GPU returns GPU-related findings.
func GPU(snap *snapshot.Snapshot) []snapshot.Finding {
	if snap == nil {
		return nil
	}
	var out []snapshot.Finding

	gpuRequested := false
	if snap.Docker != nil {
		for _, c := range snap.Docker.Containers {
			if strings.EqualFold(c.State, "running") && c.GPURequested {
				gpuRequested = true
				break
			}
		}
	}
	if gpuRequested && len(snap.GPUs) == 0 {
		out = append(out, finding(
			"gpu.no_driver",
			SeverityWarn,
			"gpu",
			rankWarn+15,
			"GPU-requested container running but no GPUs detected",
			"Containers may fall back to CPU or fail to start models.",
			"nvidia-smi  ·  bareai gpu --json  ·  docker info | grep -i runtime",
		))
	}

	for _, gpu := range snap.GPUs {
		if gpu.Vendor == "apple" && gpu.Notes != "" {
			out = append(out, finding(
				"gpu.apple_limits",
				SeverityInfo,
				"gpu",
				rankInfo+10,
				fmt.Sprintf("Apple GPU %d: %s", gpu.Index, gpu.Name),
				gpu.Notes,
				"bareai gpu --json  ·  correlate LLMs by PID/container on unified memory",
			))
			break
		}
	}

	for _, gpu := range snap.GPUs {
		if gpu.MemoryTotal == 0 {
			continue
		}
		usedPct := float64(gpu.MemoryUsed) / float64(gpu.MemoryTotal)
		if usedPct > 0.9 {
			free := gpu.MemoryTotal - gpu.MemoryUsed
			out = append(out, finding(
				"gpu.vram_high",
				SeverityWarn,
				"gpu",
				rankWarn+20,
				fmt.Sprintf("GPU %d VRAM usage above 90%% (%s / %s)", gpu.Index, formatBytes(gpu.MemoryUsed), formatBytes(gpu.MemoryTotal)),
				fmt.Sprintf("Only %s free; new model loads may OOM.", formatBytes(free)),
				fmt.Sprintf("bareai gpu --json  ·  nvidia-smi -i %d", gpu.Index),
			))
		}

		if usedPct > 0.5 && gpu.Utilization != nil && *gpu.Utilization < 5 {
			for _, llm := range snap.LLMs {
				if llm.Health != nil && llm.Health.OK && llm.GPUIndex != nil && *llm.GPUIndex == gpu.Index {
					out = append(out, finding(
						"gpu.idle_while_llm",
						SeverityInfo,
						"gpu",
						rankInfo+15,
						fmt.Sprintf("GPU %d has high VRAM (%.0f%%) but low utilization (%.0f%%)", gpu.Index, usedPct*100, *gpu.Utilization),
						"VRAM is reserved but the GPU may be waiting on I/O or batching.",
						fmt.Sprintf("bareai llm --json  ·  nvidia-smi -i %d", gpu.Index),
					))
					break
				}
			}
		}
	}

	return out
}
