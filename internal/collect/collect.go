package collect

import (
	"context"

	dockercollect "github.com/baselhusam/bareai-cli/internal/collect/docker"
	hostcollect "github.com/baselhusam/bareai-cli/internal/collect/host"
	gpucollect "github.com/baselhusam/bareai-cli/internal/collect/gpu"
	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// Snapshot builds a partial infrastructure snapshot with host, GPU, and Docker data.
func Snapshot(ctx context.Context) *snapshot.Snapshot {
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

	return snap
}
