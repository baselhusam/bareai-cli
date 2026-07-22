package llm

import (
	"strconv"
	"strings"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func correlate(llm *snapshot.LLM, in Input) {
	if llm == nil {
		return
	}
	if llm.Source != sourceDocker {
		matchContainerByPortOrPID(llm, in.Docker)
	}
	matchGPU(llm, in.GPUs, in.Docker)
}

func matchContainerByPortOrPID(llm *snapshot.LLM, docker *snapshot.Docker) {
	if docker == nil || !docker.Available {
		return
	}
	port, ok := endpointPort(llm.Endpoint)
	if !ok {
		return
	}
	for _, c := range docker.Containers {
		if llm.PID > 0 && c.PID == llm.PID {
			llm.ContainerID = c.ID
			llm.ContainerName = c.Name
			return
		}
		for _, p := range c.Ports {
			if p.PublicPort == port || p.PrivatePort == port {
				llm.ContainerID = c.ID
				llm.ContainerName = c.Name
				if llm.PID == 0 {
					llm.PID = c.PID
				}
				return
			}
		}
	}
}

func matchGPU(llm *snapshot.LLM, gpus []snapshot.GPU, docker *snapshot.Docker) {
	if llm.PID <= 0 {
		return
	}
	var idx *int
	for _, gpu := range gpus {
		for _, proc := range gpu.Processes {
			if proc.PID == llm.PID {
				if idx != nil {
					return
				}
				i := gpu.Index
				idx = &i
			}
		}
	}
	if idx != nil {
		llm.GPUIndex = idx
		return
	}

	// Best-effort fallback when process metrics are unavailable (AMD/Apple).
	if llm.ContainerID != "" || llm.ContainerName != "" {
		if gpuIdx := gpuIndexFromContainer(llm, docker, gpus); gpuIdx != nil {
			llm.GPUIndex = gpuIdx
		}
	}
}

func gpuIndexFromContainer(llm *snapshot.LLM, docker *snapshot.Docker, gpus []snapshot.GPU) *int {
	if docker == nil || !docker.Available || len(gpus) == 0 {
		return nil
	}
	var container *snapshot.DockerContainer
	for i := range docker.Containers {
		c := &docker.Containers[i]
		if llm.ContainerID != "" && c.ID == llm.ContainerID {
			container = c
			break
		}
		if llm.ContainerName != "" && (c.Name == llm.ContainerName || strings.Contains(c.Name, llm.ContainerName)) {
			container = c
			break
		}
	}
	if container == nil || !container.GPURequested {
		return nil
	}

	var candidates []int
	for _, req := range container.DeviceRequests {
		for _, id := range req.DeviceIDs {
			if idx, ok := parseGPUDeviceIndex(id); ok {
				candidates = append(candidates, idx)
			}
		}
	}
	if len(candidates) == 1 {
		idx := candidates[0]
		return &idx
	}
	if len(candidates) == 0 && len(gpus) == 1 && container.GPURequested {
		idx := gpus[0].Index
		return &idx
	}
	return nil
}

func parseGPUDeviceIndex(id string) (int, bool) {
	id = strings.TrimSpace(strings.ToLower(id))
	if id == "" {
		return 0, false
	}
	if idx, err := strconv.Atoi(id); err == nil {
		return idx, true
	}
	if strings.HasPrefix(id, "gpu:") {
		if idx, err := strconv.Atoi(strings.TrimPrefix(id, "gpu:")); err == nil {
			return idx, true
		}
	}
	return 0, false
}
