package probe

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOllamaAdapterListModels(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"models":[{"name":"llama3.2","size":123}]}`))
	}))
	defer srv.Close()

	a := &OllamaAdapter{}
	models, err := a.ListModels(context.Background(), srv.Client(), srv.URL)
	if err != nil {
		t.Fatalf("ListModels failed: %v", err)
	}
	if len(models) != 1 || models[0].ID != "llama3.2" {
		t.Fatalf("unexpected models: %+v", models)
	}
}

func TestOllamaAdapterDetect(t *testing.T) {
	a := &OllamaAdapter{}
	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:11434/api/tags", nil)
	resp := &http.Response{Request: req, StatusCode: 200}
	body := []byte(`{"models":[{"name":"llama3.2"}]}`)
	if !a.Detect("", resp, body) {
		t.Fatal("expected detect true")
	}
}
