package llm

import (
	"testing"

	"github.com/baselhusam/bareai-cli/internal/probe"
)

func TestMatchProcessRuntime(t *testing.T) {
	tests := []struct {
		name    string
		proc    string
		cmdline string
		want    string
	}{
		{
			name: "ollama exe",
			proc: "ollama.exe",
			want: probe.RuntimeOllama,
		},
		{
			name: "Ollama exe mixed case",
			proc: "Ollama.exe",
			want: probe.RuntimeOllama,
		},
		{
			name:    "python exe vllm cmdline",
			proc:    "python.exe",
			cmdline: "python -m vllm.entrypoints.openai.api_server",
			want:    probe.RuntimeVLLM,
		},
		{
			name:    "tritonserver",
			proc:    "tritonserver",
			want:    probe.RuntimeTriton,
		},
		{
			name: "unknown",
			proc: "nginx.exe",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchProcessRuntime(tt.proc, tt.cmdline)
			if got != tt.want {
				t.Fatalf("matchProcessRuntime(%q, %q) = %q, want %q", tt.proc, tt.cmdline, got, tt.want)
			}
		})
	}
}
