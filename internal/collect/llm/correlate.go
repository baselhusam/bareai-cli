package llm

import "github.com/baselhusam/bareai-cli/internal/snapshot"

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
