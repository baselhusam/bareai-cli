package inspect

import (
	"fmt"
	"strings"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

const (
	severityInfo = "info"
	severityWarn = "warn"
)

// AnalyzeFindings returns informational findings for an inspect snapshot.
func AnalyzeFindings(snap *snapshot.Snapshot) []snapshot.Finding {
	if snap == nil {
		return nil
	}

	var findings []snapshot.Finding
	findings = append(findings, llmFindings(snap.LLMs)...)
	findings = append(findings, dockerFindings(snap.Docker)...)
	findings = append(findings, gpuFindings(snap.GPUs)...)
	return findings
}

func llmFindings(llms []snapshot.LLM) []snapshot.Finding {
	var out []snapshot.Finding
	for _, llm := range llms {
		if llm.Health != nil && !llm.Health.OK {
			out = append(out, snapshot.Finding{
				ID:       "llm.unreachable",
				Severity: severityWarn,
				Summary:  fmt.Sprintf("%s (%s) is unreachable", llm.Name, llm.Endpoint),
			})
		}
	}
	if len(llms) > 2 {
		out = append(out, snapshot.Finding{
			ID:       "llm.multiple_runtimes",
			Severity: severityInfo,
			Summary:  fmt.Sprintf("%d LLM runtimes discovered on this host", len(llms)),
		})
	}
	return out
}

func dockerFindings(docker *snapshot.Docker) []snapshot.Finding {
	if docker == nil || !docker.Available || docker.NVIDIARuntime {
		return nil
	}

	for _, c := range docker.Containers {
		if !strings.EqualFold(c.State, "running") || !c.GPURequested {
			continue
		}
		return []snapshot.Finding{{
			ID:       "docker.no_nvidia_runtime",
			Severity: severityInfo,
			Summary:  "GPU-requested container running but NVIDIA runtime not registered",
		}}
	}
	return nil
}

func gpuFindings(gpus []snapshot.GPU) []snapshot.Finding {
	var out []snapshot.Finding
	for _, gpu := range gpus {
		if gpu.MemoryTotal == 0 {
			continue
		}
		usedPct := float64(gpu.MemoryUsed) / float64(gpu.MemoryTotal)
		if usedPct > 0.9 {
			out = append(out, snapshot.Finding{
				ID:       "gpu.vram_high",
				Severity: severityWarn,
				Summary:  fmt.Sprintf("GPU %d VRAM usage above 90%% (%s / %s)", gpu.Index, formatBytes(gpu.MemoryUsed), formatBytes(gpu.MemoryTotal)),
			})
		}
	}
	return out
}

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
