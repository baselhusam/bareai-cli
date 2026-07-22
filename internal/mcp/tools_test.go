package mcp

import (
	"encoding/json"
	"testing"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func TestTextResultEnvelope(t *testing.T) {
	snap := &snapshot.Snapshot{
		CollectedAt: time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC),
		Correlations: []snapshot.Correlation{{
			Kind:     snapshot.CorrelationKindLLM,
			Runtime:  "ollama",
			Endpoint: "http://127.0.0.1:11434",
		}},
	}
	data := correlationsData{Correlations: snap.Correlations}
	result, _, err := textResult(snap, data)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Content) != 1 {
		t.Fatalf("content len = %d", len(result.Content))
	}
	tc, ok := result.Content[0].(*sdkmcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	var env AgentEnvelope
	if err := json.Unmarshal([]byte(tc.Text), &env); err != nil {
		t.Fatal(err)
	}
	if env.SchemaVersion != SchemaVersion {
		t.Fatalf("schema = %q", env.SchemaVersion)
	}
}
