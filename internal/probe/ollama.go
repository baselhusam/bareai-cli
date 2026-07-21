package probe

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

type OllamaAdapter struct{}

func (a *OllamaAdapter) Runtime() string { return RuntimeOllama }

func (a *OllamaAdapter) MetricsPath() string { return "" }

func (a *OllamaAdapter) Detect(_ string, resp *http.Response, body []byte) bool {
	if resp == nil {
		return false
	}
	path := resp.Request.URL.Path
	if strings.HasPrefix(path, "/api/") {
		return true
	}
	var tags struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if json.Unmarshal(body, &tags) == nil && len(tags.Models) > 0 {
		return true
	}
	return strings.Contains(strings.ToLower(string(body)), "ollama")
}

func (a *OllamaAdapter) Health(ctx context.Context, client *http.Client, baseURL string) snapshot.ProbeResult {
	resp, body, latency, err := get(ctx, client, joinURL(baseURL, "/api/tags"))
	if err != nil {
		return failResult(latency, 0, err)
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return okResult(latency, resp.StatusCode, "tags reachable")
	}
	return readError(resp.StatusCode, body, latency)
}

func (a *OllamaAdapter) ListModels(ctx context.Context, client *http.Client, baseURL string) ([]snapshot.LLMModel, error) {
	resp, body, _, err := get(ctx, client, joinURL(baseURL, "/api/tags"))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("ollama tags: HTTP %d", resp.StatusCode)
	}
	var parsed struct {
		Models []struct {
			Name   string `json:"name"`
			Size   int64  `json:"size"`
			Model  string `json:"model"`
		} `json:"models"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	out := make([]snapshot.LLMModel, 0, len(parsed.Models))
	for _, m := range parsed.Models {
		id := m.Name
		if id == "" {
			id = m.Model
		}
		out = append(out, snapshot.LLMModel{
			ID:   id,
			Name: id,
			Size: uint64(m.Size),
		})
	}
	return out, nil
}

func (a *OllamaAdapter) Smoke(ctx context.Context, client *http.Client, baseURL, model, prompt string) snapshot.ProbeResult {
	if model == "" {
		model = "llama3.2"
	}
	payload := `{"model":"` + escapeJSON(model) + `","prompt":"` + escapeJSON(prompt) + `","stream":false}`
	resp, body, latency, err := postJSON(ctx, client, joinURL(baseURL, "/api/generate"), payload)
	if err != nil {
		return failResult(latency, 0, err)
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return okResult(latency, resp.StatusCode, "generate ok")
	}
	return readError(resp.StatusCode, body, latency)
}
