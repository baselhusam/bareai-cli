package render

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func TestWriteLLM(t *testing.T) {
	snap := &snapshot.Snapshot{
		CollectedAt: time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC),
		LLMs: []snapshot.LLM{{
			Name:     "Ollama",
			Runtime:  "ollama",
			Endpoint: "http://127.0.0.1:11434",
			Source:   "docker",
			ContainerName: "ollama",
			Health:   &snapshot.ProbeResult{OK: true, LatencyMS: 42, Message: "tags reachable"},
			Models:   []snapshot.LLMModel{{ID: "llama3.2"}},
			GPUIndex: intPtr(0),
		}},
	}

	var buf bytes.Buffer
	if err := WriteLLM(&buf, snap, true); err != nil {
		t.Fatalf("WriteLLM failed: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"bareai llm", "Ollama", "11434", "Health: ok", "llama3.2", "GPU: 0"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestWriteProbe(t *testing.T) {
	snap := &snapshot.Snapshot{
		CollectedAt: time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC),
		LLMs: []snapshot.LLM{{
			Runtime:  "ollama",
			Endpoint: "http://127.0.0.1:11434",
			Probe:    &snapshot.ProbeResult{OK: true, LatencyMS: 55, Message: "generate ok"},
		}},
	}
	var buf bytes.Buffer
	if err := WriteProbe(&buf, snap, true); err != nil {
		t.Fatalf("WriteProbe failed: %v", err)
	}
	if !strings.Contains(buf.String(), "pass") {
		t.Fatalf("expected pass in output: %s", buf.String())
	}
}

func TestWriteLLMSummary(t *testing.T) {
	var buf bytes.Buffer
	if err := writeLLMSummary(&buf, []snapshot.LLM{{
		Runtime: "ollama",
		Health:  &snapshot.ProbeResult{OK: true},
	}}); err != nil {
		t.Fatalf("writeLLMSummary failed: %v", err)
	}
	if !strings.Contains(buf.String(), "1 runtimes (ollama)") {
		t.Fatalf("unexpected summary: %s", buf.String())
	}
}

func intPtr(v int) *int { return &v }
