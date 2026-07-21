package probe

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

type TritonAdapter struct{}

func (a *TritonAdapter) Runtime() string { return RuntimeTriton }

func (a *TritonAdapter) MetricsPath() string { return "/metrics" }

func (a *TritonAdapter) Detect(_ string, resp *http.Response, body []byte) bool {
	if resp == nil {
		return false
	}
	path := resp.Request.URL.Path
	if strings.HasPrefix(path, "/v2/") {
		return true
	}
	return strings.Contains(strings.ToLower(string(body)), "triton")
}

func (a *TritonAdapter) Health(ctx context.Context, client *http.Client, baseURL string) snapshot.ProbeResult {
	resp, body, latency, err := get(ctx, client, joinURL(baseURL, "/v2/health/ready"))
	if err != nil {
		return failResult(latency, 0, err)
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return okResult(latency, resp.StatusCode, "ready")
	}
	return readError(resp.StatusCode, body, latency)
}

func (a *TritonAdapter) ListModels(ctx context.Context, client *http.Client, baseURL string) ([]snapshot.LLMModel, error) {
	resp, body, _, err := get(ctx, client, joinURL(baseURL, "/v2/models"))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, err
	}
	var parsed []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &parsed); err == nil {
		out := make([]snapshot.LLMModel, 0, len(parsed))
		for _, m := range parsed {
			out = append(out, snapshot.LLMModel{ID: m.Name, Name: m.Name})
		}
		return out, nil
	}
	var wrapped struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.Unmarshal(body, &wrapped); err != nil {
		return nil, err
	}
	out := make([]snapshot.LLMModel, 0, len(wrapped.Models))
	for _, m := range wrapped.Models {
		out = append(out, snapshot.LLMModel{ID: m.Name, Name: m.Name})
	}
	return out, nil
}

func (a *TritonAdapter) Smoke(ctx context.Context, client *http.Client, baseURL, _, _ string) snapshot.ProbeResult {
	return a.Health(ctx, client, baseURL)
}
