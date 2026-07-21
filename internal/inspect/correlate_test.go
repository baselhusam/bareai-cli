package inspect

import (
	"testing"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func TestBuildCorrelationsVRAMJoin(t *testing.T) {
	gpuIndex := 0
	llms := []snapshot.LLM{{
		Endpoint:      "http://127.0.0.1:11434",
		Runtime:       "ollama",
		ContainerName: "ollama",
		PID:           100,
		GPUIndex:      &gpuIndex,
		Models:        []snapshot.LLMModel{{ID: "llama3.2"}},
		Health:        &snapshot.ProbeResult{OK: true},
	}}
	gpus := []snapshot.GPU{{
		Index: 0,
		Processes: []snapshot.GPUProcess{{
			PID:        100,
			MemoryUsed: 2 * 1024 * 1024 * 1024,
		}},
	}}

	rows := BuildCorrelations(llms, gpus)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].VRAMBytes != 2*1024*1024*1024 {
		t.Fatalf("vram = %d", rows[0].VRAMBytes)
	}
	if rows[0].HealthOK == nil || !*rows[0].HealthOK {
		t.Fatal("expected health ok")
	}
	if len(rows[0].Models) != 1 || rows[0].Models[0] != "llama3.2" {
		t.Fatalf("models = %+v", rows[0].Models)
	}
}

func TestBuildCorrelationsEmpty(t *testing.T) {
	if rows := BuildCorrelations(nil, nil); rows != nil {
		t.Fatalf("expected nil, got %+v", rows)
	}
}

func TestAnalyzeFindingsUnreachable(t *testing.T) {
	snap := &snapshot.Snapshot{
		LLMs: []snapshot.LLM{{
			Name:     "Ollama",
			Endpoint: "http://127.0.0.1:11434",
			Health:   &snapshot.ProbeResult{OK: false},
		}},
	}
	findings := AnalyzeFindings(snap)
	if len(findings) == 0 || findings[0].ID != "llm.unreachable" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestAnalyzeFindingsNVIDIARuntime(t *testing.T) {
	snap := &snapshot.Snapshot{
		Docker: &snapshot.Docker{
			Available:     true,
			NVIDIARuntime: false,
			Containers: []snapshot.DockerContainer{{
				Name:         "vllm",
				State:        "running",
				GPURequested: true,
			}},
		},
	}
	findings := AnalyzeFindings(snap)
	found := false
	for _, f := range findings {
		if f.ID == "docker.no_nvidia_runtime" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected nvidia runtime finding, got %+v", findings)
	}
}

func TestAnalyzeFindingsVRAMHigh(t *testing.T) {
	snap := &snapshot.Snapshot{
		GPUs: []snapshot.GPU{{
			Index:       0,
			MemoryTotal: 10 * 1024 * 1024 * 1024,
			MemoryUsed:  95 * 1024 * 1024 * 1024 / 10,
		}},
	}
	findings := AnalyzeFindings(snap)
	if len(findings) != 1 || findings[0].ID != "gpu.vram_high" {
		t.Fatalf("findings = %+v", findings)
	}
}
