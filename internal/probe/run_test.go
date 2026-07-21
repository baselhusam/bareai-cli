package probe

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func TestSmokeLLMUnknownRuntime(t *testing.T) {
	ctx := context.Background()
	client := NewClient(ctx)
	llm := snapshot.LLM{Endpoint: "http://127.0.0.1:1", Runtime: "unknown"}
	result := SmokeLLM(ctx, client, &llm, "", "Hello")
	if result.OK {
		t.Fatal("expected probe failure for unknown runtime")
	}
}

func TestResolveAdapterOllama(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"version":"0.1"}`))
	}))
	defer srv.Close()

	ctx := context.Background()
	client := NewClient(ctx)
	llm := snapshot.LLM{Endpoint: srv.URL, Runtime: RuntimeOllama}
	adapter := ResolveAdapter(ctx, client, llm)
	if adapter == nil || adapter.Runtime() != RuntimeOllama {
		t.Fatalf("expected ollama adapter, got %v", adapter)
	}
}

func TestSmokeAllSetsProbe(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/tags":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"models":[{"name":"llama3"}]}`))
		case "/api/generate":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"response":"hi"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	client := NewClient(ctx)
	llms := []snapshot.LLM{{Endpoint: srv.URL, Runtime: RuntimeOllama, Name: "ollama"}}
	out := SmokeAll(ctx, client, llms, "", "Hello")
	if len(out) != 1 || out[0].Probe == nil {
		t.Fatal("expected probe result on LLM")
	}
}
