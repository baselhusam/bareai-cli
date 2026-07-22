package mcp

import (
	"context"

	dockercollect "github.com/baselhusam/bareai-cli/internal/collect/docker"
	gpucollect "github.com/baselhusam/bareai-cli/internal/collect/gpu"
	llmcollect "github.com/baselhusam/bareai-cli/internal/collect/llm"
	"github.com/baselhusam/bareai-cli/internal/probe"
	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// ProbeOptions controls MCP/CLI probe runs.
type ProbeOptions struct {
	Endpoint string
	Runtime  string
	Model    string
	Prompt   string
}

// RunProbeSnapshot probes one endpoint or all discovered LLMs.
func RunProbeSnapshot(ctx context.Context, opts ProbeOptions) *snapshot.Snapshot {
	snap := snapshot.New()
	client := probe.NewClient(ctx)

	if opts.Endpoint != "" {
		llm := snapshot.LLM{
			Endpoint: opts.Endpoint,
			Runtime:  opts.Runtime,
			Name:     opts.Runtime,
		}
		adapter := probe.ResolveAdapter(ctx, client, llm)
		if adapter == nil {
			snap.Skipped = append(snap.Skipped, snapshot.Skip{
				Component: "probe",
				Reason:    "unknown runtime for endpoint",
			})
			return snap
		}
		llm.Runtime = adapter.Runtime()
		llm.Name = llm.Runtime
		if opts.Model == "" {
			if models, err := adapter.ListModels(ctx, client, llm.Endpoint); err == nil {
				llm.Models = models
			}
		}
		result := probe.Smoke(ctx, client, llm, adapter, opts.Model, opts.Prompt)
		llm.Probe = &result
		snap.LLMs = []snapshot.LLM{llm}
		return snap
	}

	docker, dockerSkips, err := dockercollect.Collect(ctx, dockercollect.Options{Detail: false})
	snap.Docker = &docker
	snap.Skipped = append(snap.Skipped, dockerSkips...)
	if err != nil {
		snap.Skipped = append(snap.Skipped, snapshot.Skip{
			Component: "docker",
			Reason:    err.Error(),
		})
	}

	gpus, gpuSkips := gpucollect.SnapshotGPU(ctx)
	snap.GPUs = gpus
	snap.Skipped = append(snap.Skipped, gpuSkips...)

	llms, llmSkips, err := llmcollect.Collect(ctx, llmcollect.Input{
		Docker:       snap.Docker,
		GPUs:         snap.GPUs,
		Probe:        true,
		ListModels:   true,
		FetchMetrics: true,
	})
	if err != nil {
		snap.Skipped = append(snap.Skipped, snapshot.Skip{
			Component: "llm",
			Reason:    err.Error(),
		})
		return snap
	}
	snap.Skipped = append(snap.Skipped, llmSkips...)
	snap.LLMs = probe.SmokeAll(ctx, client, llms, opts.Model, opts.Prompt)
	return snap
}
