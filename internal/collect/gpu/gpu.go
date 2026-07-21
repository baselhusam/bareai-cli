package gpu

import (
	"context"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// Collect gathers GPU inventory from all supported vendors.
func Collect(ctx context.Context) ([]snapshot.GPU, []snapshot.Skip, error) {
	var all []snapshot.GPU
	var skips []snapshot.Skip

	if gpus, err := collectNVIDIA(ctx); err != nil {
		skips = append(skips, snapshot.Skip{Component: "gpu.nvidia", Reason: err.Error()})
	} else {
		all = append(all, gpus...)
	}

	if gpus, err := collectAMD(ctx); err != nil {
		skips = append(skips, snapshot.Skip{Component: "gpu.amd", Reason: err.Error()})
	} else {
		all = append(all, gpus...)
	}

	if gpus, err := collectApple(ctx); err != nil {
		skips = append(skips, snapshot.Skip{Component: "gpu.apple", Reason: err.Error()})
	} else {
		all = append(all, gpus...)
	}

	return all, skips, nil
}

// SnapshotGPU collects GPUs and appends a skip when none are found.
func SnapshotGPU(ctx context.Context) ([]snapshot.GPU, []snapshot.Skip) {
	gpus, skips, err := Collect(ctx)
	if err != nil {
		skips = append(skips, snapshot.Skip{Component: "gpu", Reason: err.Error()})
	}
	if len(gpus) == 0 && !hasVendorGPUSkip(skips) {
		skips = append(skips, snapshot.Skip{Component: "gpu", Reason: "no accelerators detected"})
	}
	return gpus, skips
}

func hasVendorGPUSkip(skips []snapshot.Skip) bool {
	for _, skip := range skips {
		if skip.Component == "gpu" || skip.Component == "gpu.nvidia" ||
			skip.Component == "gpu.amd" || skip.Component == "gpu.apple" {
			return true
		}
	}
	return false
}
