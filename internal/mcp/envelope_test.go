package mcp

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
	"github.com/baselhusam/bareai-cli/internal/version"
)

func TestWrap(t *testing.T) {
	at := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	env := Wrap(map[string]string{"ok": "yes"}, at)
	if env.SchemaVersion != SchemaVersion {
		t.Fatalf("schema_version = %q", env.SchemaVersion)
	}
	if env.BareaiVersion != version.Version {
		t.Fatalf("bareai_version = %q", env.BareaiVersion)
	}
	if !env.CollectedAt.Equal(at) {
		t.Fatalf("collected_at = %v", env.CollectedAt)
	}
}

func TestMarshalJSON(t *testing.T) {
	at := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	text, err := MarshalJSON(map[string]int{"n": 1}, at)
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"schema_version", "bareai_version", "collected_at", "data"} {
		if _, ok := parsed[key]; !ok {
			t.Fatalf("missing key %q in %s", key, text)
		}
	}
}

func TestCountSeverities(t *testing.T) {
	counts := countSeverities([]snapshot.Finding{
		{Severity: "warn"},
		{Severity: "info"},
		{Severity: ""},
	})
	if counts["warn"] != 1 || counts["info"] != 2 {
		t.Fatalf("counts = %+v", counts)
	}
}

func TestWithToolTimeoutCap(t *testing.T) {
	ctx, cancel := WithToolTimeout(t.Context(), 9999)
	defer cancel()
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected deadline")
	}
	if time.Until(deadline) > maxTimeoutSeconds*time.Second+time.Second {
		t.Fatalf("timeout not capped: %v", time.Until(deadline))
	}
}
