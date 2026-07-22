package inspect

import "github.com/baselhusam/bareai-cli/internal/snapshot"

// BuildCorrelations produces denormalized correlation rows from snapshot LLM, DB, and GPU data.
func BuildCorrelations(snap *snapshot.Snapshot) []snapshot.Correlation {
	if snap == nil {
		return nil
	}
	llmRows := buildLLMCorrelations(snap.LLMs, snap.GPUs)
	dbRows := buildDBCorrelations(snap.Databases)
	if len(llmRows) == 0 && len(dbRows) == 0 {
		return nil
	}
	out := make([]snapshot.Correlation, 0, len(llmRows)+len(dbRows))
	out = append(out, llmRows...)
	out = append(out, dbRows...)
	return out
}

func buildLLMCorrelations(llms []snapshot.LLM, gpus []snapshot.GPU) []snapshot.Correlation {
	if len(llms) == 0 {
		return nil
	}

	out := make([]snapshot.Correlation, 0, len(llms))
	for _, llm := range llms {
		row := snapshot.Correlation{
			Kind:          snapshot.CorrelationKindLLM,
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
		if row.GPUIndex != nil {
			row.GPUName = gpuNameForIndex(gpus, *row.GPUIndex)
		}
		out = append(out, row)
	}
	return out
}

func buildDBCorrelations(dbs []snapshot.Database) []snapshot.Correlation {
	if len(dbs) == 0 {
		return nil
	}

	out := make([]snapshot.Correlation, 0, len(dbs))
	for _, db := range dbs {
		row := snapshot.Correlation{
			Kind:          snapshot.CorrelationKindDB,
			Endpoint:      db.Address,
			Runtime:       db.Engine,
			ContainerName: db.ContainerName,
			ContainerID:   db.ContainerID,
			PID:           db.PID,
		}
		if db.Health != nil {
			ok := db.Health.OK
			row.HealthOK = &ok
		}
		out = append(out, row)
	}
	return out
}

func gpuNameForIndex(gpus []snapshot.GPU, index int) string {
	for _, gpu := range gpus {
		if gpu.Index == index {
			return gpu.Name
		}
	}
	return ""
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
