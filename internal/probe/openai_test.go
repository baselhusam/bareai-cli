package probe

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAIAdapterListModels(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"meta-llama/Llama-3.1-8B"}]}`))
	}))
	defer srv.Close()

	a := NewVLLMAdapter()
	models, err := a.ListModels(context.Background(), srv.Client(), srv.URL)
	if err != nil {
		t.Fatalf("ListModels failed: %v", err)
	}
	if len(models) != 1 || models[0].ID != "meta-llama/Llama-3.1-8B" {
		t.Fatalf("unexpected models: %+v", models)
	}
}

func TestParsePrometheusMetrics(t *testing.T) {
	text := "# comment\nvllm:num_requests_running 2\nother_metric 9\n"
	got := parsePrometheusMetrics(text, []string{"vllm:num_requests_running"})
	if got["vllm:num_requests_running"] != 2 {
		t.Fatalf("unexpected metrics: %+v", got)
	}
}
