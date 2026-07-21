package probe

import (
	"context"
	"net/http"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// ResolveAdapter returns the best adapter for an LLM endpoint.
func ResolveAdapter(ctx context.Context, client *http.Client, llm snapshot.LLM) Adapter {
	adapter := AdapterForRuntime(llm.Runtime)
	if adapter == nil {
		adapter = DetectAdapter(ctx, client, llm.Endpoint)
	}
	return adapter
}

// SmokeLLM runs a one-hit smoke probe against a single LLM.
func SmokeLLM(ctx context.Context, client *http.Client, llm *snapshot.LLM, model, prompt string) snapshot.ProbeResult {
	if llm == nil {
		return snapshot.ProbeResult{OK: false, Error: "no endpoint"}
	}
	adapter := ResolveAdapter(ctx, client, *llm)
	if adapter == nil {
		return snapshot.ProbeResult{OK: false, Error: "unknown runtime"}
	}
	return Smoke(ctx, client, *llm, adapter, model, prompt)
}

// SmokeAll runs smoke probes against every discovered LLM.
func SmokeAll(ctx context.Context, client *http.Client, llms []snapshot.LLM, model, prompt string) []snapshot.LLM {
	out := append([]snapshot.LLM(nil), llms...)
	for i := range out {
		result := SmokeLLM(ctx, client, &out[i], model, prompt)
		out[i].Probe = &result
	}
	return out
}
