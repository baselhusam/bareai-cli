package probe

import (
	"context"
	"net/http"
	"strings"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

const (
	RuntimeOllama       = "ollama"
	RuntimeVLLM         = "vllm"
	RuntimeSGLang       = "sglang"
	RuntimeTriton       = "triton"
	RuntimeOpenAICompat = "openai_compat"
)

// Adapter describes vendor-specific HTTP probe behavior.
type Adapter interface {
	Runtime() string
	Detect(baseURL string, resp *http.Response, body []byte) bool
	Health(ctx context.Context, client *http.Client, baseURL string) snapshot.ProbeResult
	ListModels(ctx context.Context, client *http.Client, baseURL string) ([]snapshot.LLMModel, error)
	Smoke(ctx context.Context, client *http.Client, baseURL, model, prompt string) snapshot.ProbeResult
	MetricsPath() string
}

// Adapters returns runtime adapters in detection priority order.
func Adapters() []Adapter {
	return []Adapter{
		&OllamaAdapter{},
		&TritonAdapter{},
		NewVLLMAdapter(),
		NewSGLangAdapter(),
		NewOpenAICompatAdapter(),
	}
}

// AdapterForRuntime returns the adapter for a runtime name, or nil.
func AdapterForRuntime(runtime string) Adapter {
	for _, a := range Adapters() {
		if strings.EqualFold(a.Runtime(), runtime) {
			return a
		}
	}
	return nil
}

// DetectAdapter probes baseURL and returns the matching adapter, or nil.
func DetectAdapter(ctx context.Context, client *http.Client, baseURL string) Adapter {
	paths := []string{"/", "/api/tags", "/v1/models", "/v2/health/ready", "/health"}
	seen := make(map[string]bool)
	for _, path := range paths {
		if seen[path] {
			continue
		}
		seen[path] = true
		resp, body, _, err := get(ctx, client, joinURL(baseURL, path))
		if err != nil || resp == nil {
			continue
		}
		for _, a := range Adapters() {
			if a.Detect(baseURL, resp, body) {
				return a
			}
		}
	}
	return nil
}
