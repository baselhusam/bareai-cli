package tui

import (
	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

const defaultHistoryMax = 40

type metricHistory struct {
	max int

	load1  []float64
	memPct []float64

	gpuUtilSamples map[int][]float64
	gpuMemSamples  map[int][]float64
	llmLatency     map[string][]float64
}

func newMetricHistory(max int) metricHistory {
	if max <= 0 {
		max = defaultHistoryMax
	}
	return metricHistory{
		max:            max,
		gpuUtilSamples: make(map[int][]float64),
		gpuMemSamples:  make(map[int][]float64),
		llmLatency:     make(map[string][]float64),
	}
}

func (h *metricHistory) record(snap *snapshot.Snapshot) {
	if snap == nil {
		return
	}

	if snap.Host != nil {
		host := snap.Host
		h.load1 = h.appendSample(h.load1, loadPct(host.Load1, host.CPUCores))
		h.memPct = h.appendSample(h.memPct, pctUsed(host.MemUsed, host.MemTotal))
	}

	seen := make(map[int]struct{}, len(snap.GPUs))
	for _, gpu := range snap.GPUs {
		seen[gpu.Index] = struct{}{}
		if gpu.Utilization != nil {
			h.gpuUtilSamples[gpu.Index] = h.appendSample(h.gpuUtilSamples[gpu.Index], *gpu.Utilization)
		}
		if gpu.MemoryTotal > 0 {
			h.gpuMemSamples[gpu.Index] = h.appendSample(h.gpuMemSamples[gpu.Index], pctUsed(gpu.MemoryUsed, gpu.MemoryTotal))
		}
	}

	for idx := range h.gpuUtilSamples {
		if _, ok := seen[idx]; !ok {
			delete(h.gpuUtilSamples, idx)
			delete(h.gpuMemSamples, idx)
		}
	}

	seenLLM := make(map[string]struct{}, len(snap.LLMs))
	for _, llm := range snap.LLMs {
		key := llm.Endpoint
		if key == "" {
			key = llm.Runtime
		}
		seenLLM[key] = struct{}{}
		if llm.Health != nil && llm.Health.LatencyMS > 0 {
			h.llmLatency[key] = h.appendSample(h.llmLatency[key], float64(llm.Health.LatencyMS))
		}
	}
	for key := range h.llmLatency {
		if _, ok := seenLLM[key]; !ok {
			delete(h.llmLatency, key)
		}
	}
}

func (h metricHistory) appendSample(samples []float64, value float64) []float64 {
	samples = append(samples, value)
	max := h.max
	if max <= 0 {
		max = defaultHistoryMax
	}
	if len(samples) > max {
		samples = samples[len(samples)-max:]
	}
	return samples
}

func (h metricHistory) hostLoad() []float64  { return h.load1 }
func (h metricHistory) hostMem() []float64   { return h.memPct }
func (h metricHistory) gpuUtil(idx int) []float64 {
	if h.gpuUtilSamples == nil {
		return nil
	}
	return h.gpuUtilSamples[idx]
}
func (h metricHistory) gpuMem(idx int) []float64 {
	if h.gpuMemSamples == nil {
		return nil
	}
	return h.gpuMemSamples[idx]
}

func (h metricHistory) llmHealthLatency(endpoint string) []float64 {
	if h.llmLatency == nil {
		return nil
	}
	return h.llmLatency[endpoint]
}
