package doctor

import (
	"testing"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func TestAnalyzeSortsByRank(t *testing.T) {
	snap := &snapshot.Snapshot{
		LLMs: []snapshot.LLM{{
			Name:     "Ollama",
			Endpoint: "http://127.0.0.1:11434",
			Health:   &snapshot.ProbeResult{OK: false},
		}},
		GPUs: []snapshot.GPU{{
			Index:       0,
			MemoryTotal: 10,
			MemoryUsed:  95,
		}},
	}
	findings := Analyze(snap, Options{})
	if len(findings) < 2 {
		t.Fatalf("findings = %+v", findings)
	}
	if findings[0].ID != "llm.unreachable" {
		t.Fatalf("expected llm.unreachable first, got %s", findings[0].ID)
	}
}

func TestAnalyzeSeverityFilter(t *testing.T) {
	snap := &snapshot.Snapshot{
		LLMs: []snapshot.LLM{
			{Name: "A", Endpoint: "http://127.0.0.1:1", Health: &snapshot.ProbeResult{OK: false}},
			{Name: "B", Endpoint: "http://127.0.0.1:2", Health: &snapshot.ProbeResult{OK: true}},
			{Name: "C", Endpoint: "http://127.0.0.1:3", Health: &snapshot.ProbeResult{OK: true}},
		},
	}
	findings := Analyze(snap, Options{MinSeverity: "warn"})
	for _, f := range findings {
		if f.Severity == "info" {
			t.Fatalf("info finding leaked through warn filter: %+v", f)
		}
	}
}
