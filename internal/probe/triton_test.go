package probe

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTritonAdapterHealth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/health/ready" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	a := &TritonAdapter{}
	result := a.Health(context.Background(), srv.Client(), srv.URL)
	if !result.OK {
		t.Fatalf("expected ok health, got %+v", result)
	}
}
