package mcp

import (
	"context"

	"github.com/baselhusam/bareai-cli/internal/collect"
	"github.com/baselhusam/bareai-cli/internal/doctor"
	"github.com/baselhusam/bareai-cli/internal/inspect"
	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// CollectEnriched runs collectors and populates correlation rows.
func CollectEnriched(ctx context.Context, light bool) *snapshot.Snapshot {
	opts := collect.FullOptions()
	if light {
		opts = collect.LightRefreshOptions()
	}
	snap := collect.SnapshotWithOptions(ctx, opts)
	inspect.Enrich(snap)
	return snap
}

// CollectDoctor runs full collection, enrich, and doctor analysis.
func CollectDoctor(ctx context.Context, minSeverity string) *snapshot.Snapshot {
	snap := CollectEnriched(ctx, false)
	snap.Findings = doctor.Analyze(snap, doctor.Options{MinSeverity: minSeverity})
	return snap
}

// CollectLLMs discovers LLM runtimes with optional model listing.
func CollectLLMs(ctx context.Context, listModels bool) *snapshot.Snapshot {
	return collect.SnapshotWithOptions(ctx, collect.Options{
		ProbeLLM:     true,
		ListModels:   listModels,
		FetchMetrics: true,
		ProbeDB:      false,
		DockerDetail: false,
	})
}

// CollectDatabases discovers and probes local databases.
func CollectDatabases(ctx context.Context) *snapshot.Snapshot {
	return collect.SnapshotWithOptions(ctx, collect.Options{
		ProbeLLM:     false,
		ListModels:   false,
		FetchMetrics: false,
		ProbeDB:      true,
		DockerDetail: false,
	})
}
