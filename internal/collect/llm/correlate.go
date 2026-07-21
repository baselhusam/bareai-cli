package llm

import (
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
	matchGPU(llm, in.GPUs)
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

func matchGPU(llm *snapshot.LLM, gpus []snapshot.GPU) {
	if llm.PID <= 0 {
		return
	}
	var idx *int
	for _, gpu := range gpus {
		for _, proc := range gpu.Processes {
			if proc.PID == llm.PID {
				if idx != nil {
					// ambiguous
					return
				}
				i := gpu.Index
				idx = &i
			}
		}
	}
	if idx != nil {
		llm.GPUIndex = idx
	}
}

func sourceLabel(llm snapshot.LLM) string {
	switch llm.Source {
	case sourceDocker:
		if llm.ContainerName != "" {
			return "docker: " + llm.ContainerName
		}
		return "docker"
	case sourceProcess:
		if llm.PID > 0 {
			return "process pid " + itoa(llm.PID)
		}
		return "process"
	case sourcePort:
		return "port scan"
	default:
		return llm.Source
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}

func runtimeList(llms []snapshot.LLM) string {
	seen := make(map[string]bool)
	var names []string
	for _, llm := range llms {
		r := strings.ToLower(llm.Runtime)
		if r == "" || seen[r] {
			continue
		}
		seen[r] = true
		names = append(names, r)
	}
	return strings.Join(names, ", ")
}
