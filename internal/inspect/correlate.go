package inspect

import "github.com/baselhusam/bareai-cli/internal/snapshot"

// BuildCorrelations produces denormalized correlation rows from LLM and GPU data.
func BuildCorrelations(llms []snapshot.LLM, gpus []snapshot.GPU) []snapshot.Correlation {
	if len(llms) == 0 {
		return nil
	}

	out := make([]snapshot.Correlation, 0, len(llms))
	for _, llm := range llms {
		row := snapshot.Correlation{
			Endpoint:      llm.Endpoint,
			Runtime:       llm.Runtime,
			ContainerName: llm.ContainerName,
			ContainerID:   llm.ContainerID,
			PID:           llm.PID,
			GPUIndex:      llm.GPUIndex,
			Models:        modelIDs(llm.Models),
		}
		if llm.Health != nil {
			ok := llm.Health.OK
			row.HealthOK = &ok
		}
		if llm.PID > 0 {
			if vram := vramForPID(gpus, llm.GPUIndex, llm.PID); vram > 0 {
				row.VRAMBytes = vram
			}
		}
		out = append(out, row)
	}
	return out
}

func modelIDs(models []snapshot.LLMModel) []string {
	if len(models) == 0 {
		return nil
	}
	out := make([]string, 0, len(models))
	for _, m := range models {
		if m.ID != "" {
			out = append(out, m.ID)
		}
	}
	return out
}

func vramForPID(gpus []snapshot.GPU, gpuIndex *int, pid int) uint64 {
	if gpuIndex != nil {
		for _, gpu := range gpus {
			if gpu.Index != *gpuIndex {
				continue
			}
			for _, proc := range gpu.Processes {
				if proc.PID == pid {
					return proc.MemoryUsed
				}
			}
			return 0
		}
	}

	for _, gpu := range gpus {
		for _, proc := range gpu.Processes {
			if proc.PID == pid {
				return proc.MemoryUsed
			}
		}
	}
	return 0
}
