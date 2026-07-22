package render

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func TestTerminalWidthFromColumns(t *testing.T) {
	t.Setenv("COLUMNS", "100")
	if got := TerminalWidth(os.Stdout); got != 100 {
		t.Fatalf("TerminalWidth() = %d, want 100", got)
	}
}

func TestWriteInspect(t *testing.T) {
	gpuIndex := 0
	ok := true
	snap := &snapshot.Snapshot{
		CollectedAt: time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC),
		Host:        &snapshot.Host{Hostname: "ai-box"},
		GPUs:        []snapshot.GPU{{Index: 0, Name: "NVIDIA A100", Vendor: "nvidia"}},
		Docker:      &snapshot.Docker{Available: true, NVIDIARuntime: true},
		Correlations: []snapshot.Correlation{{
			Endpoint:      "http://127.0.0.1:11434",
			Runtime:       "ollama",
			ContainerName: "ollama",
			PID:           100,
			GPUIndex:      &gpuIndex,
			VRAMBytes:     2 * giB,
			Models:        []string{"llama3.2"},
			HealthOK:      &ok,
		}},
		LLMs: []snapshot.LLM{{
			Name:     "Ollama",
			Runtime:  "ollama",
			Endpoint: "http://127.0.0.1:11434",
			Health:   &snapshot.ProbeResult{OK: true, LatencyMS: 42},
		}},
		Findings: []snapshot.Finding{{
			ID: "llm.multiple_runtimes", Severity: "info", Summary: "3 LLM runtimes discovered",
		}},
	}

	var buf bytes.Buffer
	if err := WriteInspect(&buf, snap, InspectOptions{NoColor: true, Width: 120}); err != nil {
		t.Fatalf("WriteInspect failed: %v", err)
	}

	out := buf.String()
	for _, want := range []string{
		"bareai inspect",
		"Overview",
		"ai-box",
		"Correlation",
		"KIND",
		"ollama",
		"2.0 GiB",
		"llama3.2",
		"Databases",
		"LLM runtimes",
		"Findings",
		"llm.multiple_runtimes",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}

func TestWriteInspectNarrow(t *testing.T) {
	snap := &snapshot.Snapshot{
		CollectedAt: time.Now().UTC(),
		Correlations: []snapshot.Correlation{{
			Endpoint: "http://127.0.0.1:11434",
			Runtime:  "ollama",
			Models:   []string{"llama3.2"},
		}},
	}
	var buf bytes.Buffer
	if err := WriteInspect(&buf, snap, InspectOptions{NoColor: true, Width: 72}); err != nil {
		t.Fatalf("WriteInspect failed: %v", err)
	}
	if strings.Contains(buf.String(), "ENDPOINT") {
		t.Fatalf("narrow layout should use theater columns: %s", buf.String())
	}
}
