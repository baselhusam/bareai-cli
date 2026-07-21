package collect

import (
	"context"

	dockercollect "github.com/baselhusam/bareai-cli/internal/collect/docker"
	hostcollect "github.com/baselhusam/bareai-cli/internal/collect/host"
	gpucollect "github.com/baselhusam/bareai-cli/internal/collect/gpu"
	llmcollect "github.com/baselhusam/bareai-cli/internal/collect/llm"
	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// Options controls snapshot collection behavior.
type Options struct {
	ProbeLLM bool
}

// Snapshot builds a partial infrastructure snapshot with host, GPU, Docker, and LLM data.
func Snapshot(ctx context.Context) *snapshot.Snapshot {
	return SnapshotWithOptions(ctx, Options{})
}

// SnapshotWithOptions builds a snapshot with configurable LLM probing.
func SnapshotWithOptions(ctx context.Context, opts Options) *snapshot.Snapshot {
	snap := snapshot.New()

	host, err := hostcollect.Collect(ctx)
	if err != nil {
		snap.Skipped = append(snap.Skipped, snapshot.Skip{
			Component: "host",
			Reason:    err.Error(),
		})
	} else {
		snap.Host = &host
	}

	gpus, skips := gpucollect.SnapshotGPU(ctx)
	snap.GPUs = gpus
	snap.Skipped = append(snap.Skipped, skips...)

	docker, dockerSkips, err := dockercollect.Collect(ctx)
	if err != nil {
		snap.Skipped = append(snap.Skipped, snapshot.Skip{
			Component: "docker",
			Reason:    err.Error(),
		})
	} else {
		snap.Docker = &docker
		snap.Skipped = append(snap.Skipped, dockerSkips...)
	}

	llms, llmSkips, err := llmcollect.Collect(ctx, llmcollect.Input{
		Docker: snap.Docker,
		GPUs:   snap.GPUs,
		Probe:  opts.ProbeLLM,
	})
	if err != nil {
		snap.Skipped = append(snap.Skipped, snapshot.Skip{
			Component: "llm",
			Reason:    err.Error(),
		})
	} else {
		snap.LLMs = llms
		snap.Skipped = append(snap.Skipped, llmSkips...)
	}

	return snap
}
