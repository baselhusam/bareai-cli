package probe

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

type openAIAdapterBase struct {
	runtimeName string
	metricsPath string
}

func (b *openAIAdapterBase) Runtime() string { return b.runtimeName }

func (b *openAIAdapterBase) MetricsPath() string { return b.metricsPath }

func (b *openAIAdapterBase) Detect(_ string, resp *http.Response, body []byte) bool {
	if resp == nil {
		return false
	}
	path := resp.Request.URL.Path
	if strings.HasPrefix(path, "/v1/") {
		return true
	}
	var models struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if json.Unmarshal(body, &models) == nil && len(models.Data) > 0 {
		return true
	}
	var errBody struct {
		Object string `json:"object"`
	}
	if json.Unmarshal(body, &errBody) == nil && errBody.Object == "list" {
		return true
	}
	return false
}

func (b *openAIAdapterBase) Health(ctx context.Context, client *http.Client, baseURL string) snapshot.ProbeResult {
	if resp, _, latency, err := get(ctx, client, joinURL(baseURL, "/health")); err == nil && resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return okResult(latency, resp.StatusCode, "health ok")
	}
	resp, body, latency, err := get(ctx, client, joinURL(baseURL, "/v1/models"))
	if err != nil {
		return failResult(latency, 0, err)
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return okResult(latency, resp.StatusCode, "models reachable")
	}
	return readError(resp.StatusCode, body, latency)
}

func (b *openAIAdapterBase) ListModels(ctx context.Context, client *http.Client, baseURL string) ([]snapshot.LLMModel, error) {
	resp, body, _, err := get(ctx, client, joinURL(baseURL, "/v1/models"))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("openai models: HTTP %d", resp.StatusCode)
	}
	var parsed struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	out := make([]snapshot.LLMModel, 0, len(parsed.Data))
	for _, m := range parsed.Data {
		out = append(out, snapshot.LLMModel{ID: m.ID, Name: m.ID})
	}
	return out, nil
}

func (b *openAIAdapterBase) Smoke(ctx context.Context, client *http.Client, baseURL, model, prompt string) snapshot.ProbeResult {
	if model == "" {
		model = "default"
	}
	payload := `{"model":"` + escapeJSON(model) + `","messages":[{"role":"user","content":"` + escapeJSON(prompt) + `"}],"max_tokens":1}`
	resp, body, latency, err := postJSON(ctx, client, joinURL(baseURL, "/v1/chat/completions"), payload)
	if err != nil {
		return failResult(latency, 0, err)
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return okResult(latency, resp.StatusCode, "chat ok")
	}
	return readError(resp.StatusCode, body, latency)
}

type VLLMAdapter struct {
	openAIAdapterBase
}

func NewVLLMAdapter() *VLLMAdapter {
	return &VLLMAdapter{openAIAdapterBase: openAIAdapterBase{runtimeName: RuntimeVLLM, metricsPath: "/metrics"}}
}

type SGLangAdapter struct {
	openAIAdapterBase
}

func NewSGLangAdapter() *SGLangAdapter {
	return &SGLangAdapter{openAIAdapterBase: openAIAdapterBase{runtimeName: RuntimeSGLang, metricsPath: "/metrics"}}
}

type OpenAICompatAdapter struct {
	openAIAdapterBase
}

func NewOpenAICompatAdapter() *OpenAICompatAdapter {
	return &OpenAICompatAdapter{openAIAdapterBase: openAIAdapterBase{runtimeName: RuntimeOpenAICompat, metricsPath: "/metrics"}}
}

// OpenAICompatAdapter only matches generic /v1/models without vendor-specific hints.
func (a *OpenAICompatAdapter) Detect(baseURL string, resp *http.Response, body []byte) bool {
	if a.openAIAdapterBase.Detect(baseURL, resp, body) {
		// Avoid stealing Ollama/Triton paths.
		if resp != nil && strings.HasPrefix(resp.Request.URL.Path, "/v2/") {
			return false
		}
		return true
	}
	return false
}
