package action

import (
	"context"
	"fmt"

	bareaimcp "github.com/baselhusam/bareai-cli/internal/mcp"
	"github.com/baselhusam/bareai-cli/internal/config"
)

func runReprobe(ctx context.Context, req Request, target *Target) (Step, *bool) {
	endpoint := target.Endpoint
	runtime := target.Runtime
	if endpoint == "" && req.Endpoint != "" {
		endpoint = req.Endpoint
	}
	if runtime == "" && req.Runtime != "" {
		runtime = req.Runtime
	}
	if endpoint == "" {
		return Step{Verb: VerbReprobe, OK: false, Error: "no endpoint to probe"}, nil
	}

	cfg := config.Global()
	model := cfg.Probe.Model
	prompt := cfg.Probe.Prompt
	snap := bareaimcp.RunProbeSnapshot(ctx, bareaimcp.ProbeOptions{
		Endpoint: endpoint,
		Runtime:  runtime,
		Model:    model,
		Prompt:   prompt,
	})
	if len(snap.LLMs) == 0 {
		reason := "probe returned no results"
		if len(snap.Skipped) > 0 {
			reason = snap.Skipped[0].Reason
		}
		return Step{Verb: VerbReprobe, OK: false, Error: reason, Summary: endpoint}, nil
	}
	llm := snap.LLMs[0]
	ok := llm.Probe != nil && llm.Probe.OK
	step := Step{
		Verb:    VerbReprobe,
		OK:      ok,
		Summary: fmt.Sprintf("probe %s ok=%v", endpoint, ok),
	}
	if llm.Probe != nil && llm.Probe.Error != "" {
		step.Error = llm.Probe.Error
	}
	return step, &ok
}
