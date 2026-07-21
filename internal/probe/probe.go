package probe

import (
	"context"
	"net/http"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// Enrich runs health, model listing, and metrics for a discovered LLM.
func Enrich(ctx context.Context, client *http.Client, llm *snapshot.LLM, adapter Adapter, withModels bool) {
	if llm == nil || adapter == nil {
		return
	}
	health := adapter.Health(ctx, client, llm.Endpoint)
	llm.Health = &health
	if withModels {
		if models, err := adapter.ListModels(ctx, client, llm.Endpoint); err == nil {
			llm.Models = models
		}
	}
	if metrics := FetchMetrics(ctx, client, llm.Endpoint, adapter); metrics != nil {
		llm.Metrics = metrics
	}
}

// Smoke runs a one-hit completion probe against an LLM endpoint.
func Smoke(ctx context.Context, client *http.Client, llm snapshot.LLM, adapter Adapter, model, prompt string) snapshot.ProbeResult {
	if adapter == nil {
		return snapshot.ProbeResult{OK: false, Error: "unknown runtime"}
	}
	if model == "" {
		model = defaultModel(adapter.Runtime(), llm.Models)
	}
	if prompt == "" {
		prompt = "Hello"
	}
	return adapter.Smoke(ctx, client, llm.Endpoint, model, prompt)
}
