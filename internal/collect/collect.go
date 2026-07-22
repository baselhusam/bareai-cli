package collect

import (
	"context"
	"sync"

	dockercollect "github.com/baselhusam/bareai-cli/internal/collect/docker"
	dbcollect "github.com/baselhusam/bareai-cli/internal/collect/db"
	hostcollect "github.com/baselhusam/bareai-cli/internal/collect/host"
	gpucollect "github.com/baselhusam/bareai-cli/internal/collect/gpu"
	llmcollect "github.com/baselhusam/bareai-cli/internal/collect/llm"
	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// Options controls snapshot collection behavior.
type Options struct {
	ProbeLLM     bool
	ListModels   bool
	FetchMetrics bool
	ProbeDB      bool
}

// FullOptions returns options for one-shot CLI commands.
func FullOptions() Options {
	return Options{
		ProbeLLM:     true,
		ListModels:   true,
		FetchMetrics: true,
		ProbeDB:      true,
	}
}

// LightRefreshOptions returns options for periodic TUI refresh.
func LightRefreshOptions() Options {
	return Options{
		ProbeLLM:     true,
		ListModels:   false,
		FetchMetrics: true,
		ProbeDB:      true,
	}
}

// Snapshot builds a partial infrastructure snapshot with host, GPU, Docker, and LLM data.
func Snapshot(ctx context.Context) *snapshot.Snapshot {
	return SnapshotWithOptions(ctx, Options{})
}

// SnapshotWithOptions builds a snapshot with configurable LLM probing.
func SnapshotWithOptions(ctx context.Context, opts Options) *snapshot.Snapshot {
	snap := snapshot.New()

	var (
		host   snapshot.Host
		hostOK bool
		hostMu sync.Mutex
		wg     sync.WaitGroup
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		h, err := hostcollect.Collect(ctx)
		hostMu.Lock()
		defer hostMu.Unlock()
		if err != nil {
			snap.Skipped = append(snap.Skipped, snapshot.Skip{
				Component: "host",
				Reason:    err.Error(),
			})
			return
		}
		host = h
		hostOK = true
	}()

	gpus, gpuSkips := gpucollect.SnapshotGPU(ctx)
	snap.GPUs = gpus
	snap.Skipped = append(snap.Skipped, gpuSkips...)

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

	wg.Wait()
	if hostOK {
		snap.Host = &host
	}

	llms, llmSkips, err := llmcollect.Collect(ctx, llmcollect.Input{
		Docker:       snap.Docker,
		GPUs:         snap.GPUs,
		Probe:        opts.ProbeLLM,
		ListModels:   opts.ListModels,
		FetchMetrics: opts.FetchMetrics,
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

	dbs, dbSkips, err := dbcollect.Collect(ctx, dbcollect.Input{
		Docker: snap.Docker,
		Probe:  opts.ProbeDB,
	})
	if err != nil {
		snap.Skipped = append(snap.Skipped, snapshot.Skip{
			Component: "db",
			Reason:    err.Error(),
		})
	} else {
		snap.Databases = dbs
		snap.Skipped = append(snap.Skipped, dbSkips...)
	}

	return snap
}
