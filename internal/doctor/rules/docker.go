package rules

import (
	"strings"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// Docker returns Docker-related findings.
func Docker(snap *snapshot.Snapshot) []snapshot.Finding {
	if snap == nil {
		return nil
	}
	var out []snapshot.Finding

	docker := snap.Docker
	if docker == nil || !docker.Available {
		for _, llm := range snap.LLMs {
			if llm.Source == "docker" {
				out = append(out, finding(
					"docker.unavailable",
					SeverityInfo,
					"docker",
					rankInfo+10,
					"Docker-sourced LLM discovered but Docker engine unavailable",
					"Discovery may be stale or Docker Desktop/daemon is stopped.",
					"docker ps  ·  bareai docker --json",
				))
				break
			}
		}
		return out
	}

	if !docker.NVIDIARuntime {
		for _, c := range docker.Containers {
			if !strings.EqualFold(c.State, "running") || !c.GPURequested {
				continue
			}
			out = append(out, finding(
				"docker.no_nvidia_runtime",
				SeverityInfo,
				"docker",
				rankInfo+20,
				"GPU-requested container running but NVIDIA runtime not registered",
				"GPU passthrough requires nvidia-container-toolkit and the nvidia runtime.",
				"docker info | grep -i runtime  ·  bareai docker --json",
			))
			break
		}
	}

	return out
}
