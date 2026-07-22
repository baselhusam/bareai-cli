package render

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func TestWriteDoctorShare(t *testing.T) {
	ok := true
	snap := &snapshot.Snapshot{
		CollectedAt: time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC),
		Host:        &snapshot.Host{Hostname: "ai-box", OS: "darwin", Arch: "arm64", Platform: "macOS"},
		GPUs:        []snapshot.GPU{{Index: 0, Name: "Apple M2", Vendor: "apple"}},
		Correlations: []snapshot.Correlation{{
			Kind:     snapshot.CorrelationKindLLM,
			Runtime:  "ollama",
			Models:   []string{"llama3.2"},
			HealthOK: &ok,
		}},
		Findings: []snapshot.Finding{{
			ID: "host.empty_box", Severity: "info", Summary: "idle box", Why: "nothing running", Try: "start ollama",
		}},
	}

	var buf bytes.Buffer
	if err := WriteDoctorShare(&buf, snap, "0.1.0"); err != nil {
		t.Fatalf("WriteDoctorShare failed: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		"bareai doctor 0.1.0",
		"Host: ai-box",
		"Correlation",
		"llm",
		"Findings",
		"host.empty_box",
		"Why:",
		"Try:",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}
