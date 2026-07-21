package gpu

import (
	"testing"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func TestParseNVIDIAGPUsCSV(t *testing.T) {
	data := `0, NVIDIA A100-SXM4-80GB, GPU-abc-123, 535.54, 81920, 4096, 45, 55, 250.00, 300.00
1, NVIDIA A100-SXM4-80GB, GPU-def-456, 535.54, 81920, 1024, 0, [N/A], [N/A], [N/A]
`

	gpus, err := parseNVIDIAGPUsCSV(data)
	if err != nil {
		t.Fatalf("parseNVIDIAGPUsCSV failed: %v", err)
	}
	if len(gpus) != 2 {
		t.Fatalf("expected 2 GPUs, got %d", len(gpus))
	}

	g0 := gpus[0]
	if g0.Vendor != "nvidia" || g0.Index != 0 {
		t.Errorf("unexpected gpu0: %+v", g0)
	}
	if g0.MemoryTotal != 81920*1024*1024 {
		t.Errorf("memory total bytes = %d", g0.MemoryTotal)
	}
	if g0.Utilization == nil || *g0.Utilization != 45 {
		t.Errorf("expected utilization 45, got %+v", g0.Utilization)
	}
	if g0.PowerDrawW == nil || *g0.PowerDrawW != 250 {
		t.Errorf("expected power draw 250, got %+v", g0.PowerDrawW)
	}

	g1 := gpus[1]
	if g1.Temperature != nil || g1.PowerDrawW != nil {
		t.Errorf("expected nil optional metrics for gpu1, got %+v", g1)
	}
}

func TestParseNVIDIAProcessesCSV(t *testing.T) {
	data := `GPU-abc-123, 1234, python, 2048
GPU-def-456, 5678, tritonserver, 1024
`

	procs, err := parseNVIDIAProcessesCSV(data)
	if err != nil {
		t.Fatalf("parseNVIDIAProcessesCSV failed: %v", err)
	}
	if len(procs) != 2 {
		t.Fatalf("expected 2 processes, got %d", len(procs))
	}
	if procs[0].Proc.PID != 1234 || procs[0].Proc.MemoryUsed != 2048*1024*1024 {
		t.Errorf("unexpected process: %+v", procs[0])
	}
}

func TestAttachNVIDIAProcesses(t *testing.T) {
	gpus := []snapshot.GPU{
		{UUID: "GPU-abc-123", Name: "GPU0"},
		{UUID: "GPU-def-456", Name: "GPU1"},
	}
	procs := []nvidiaProcess{
		{GPUUUID: "GPU-abc-123", Proc: snapshot.GPUProcess{PID: 1, Name: "python"}},
	}

	attachNVIDIAProcesses(gpus, procs)
	if len(gpus[0].Processes) != 1 || gpus[0].Processes[0].PID != 1 {
		t.Errorf("process not attached: %+v", gpus[0].Processes)
	}
	if len(gpus[1].Processes) != 0 {
		t.Errorf("unexpected processes on gpu1: %+v", gpus[1].Processes)
	}
}

func TestParseOptionalFloat(t *testing.T) {
	if parseOptionalFloat("[N/A]") != nil {
		t.Error("expected nil for [N/A]")
	}
	f := parseOptionalFloat("42.5")
	if f == nil || *f != 42.5 {
		t.Errorf("expected 42.5, got %+v", f)
	}
}
