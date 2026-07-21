package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func TestWriteGPU(t *testing.T) {
	util := 55.0
	temp := 62.0
	snap := &snapshot.Snapshot{
		GPUs: []snapshot.GPU{{
			Index:       0,
			Vendor:      "nvidia",
			Name:        "NVIDIA A100",
			UUID:        "GPU-123",
			Driver:      "535.54",
			MemoryTotal: 80 * giB,
			MemoryUsed:  8 * giB,
			Utilization: &util,
			Temperature: &temp,
			Processes: []snapshot.GPUProcess{{
				PID:        42,
				Name:       "python",
				MemoryUsed: 2 * giB,
			}},
		}},
	}

	var buf bytes.Buffer
	if err := WriteGPU(&buf, snap, true); err != nil {
		t.Fatalf("WriteGPU failed: %v", err)
	}

	out := buf.String()
	for _, want := range []string{
		"bareai gpu",
		"NVIDIA A100",
		"GPU-123",
		"pid 42",
		"python",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}

func TestWriteGPUSummary(t *testing.T) {
	util := 10.0
	gpus := []snapshot.GPU{{
		Index:       0,
		Vendor:      "nvidia",
		Name:        "Test GPU",
		MemoryTotal: 16 * giB,
		MemoryUsed:  4 * giB,
		Utilization: &util,
	}}

	var buf bytes.Buffer
	if err := writeGPUSummary(&buf, gpus); err != nil {
		t.Fatalf("writeGPUSummary failed: %v", err)
	}
	if !strings.Contains(buf.String(), "Test GPU") {
		t.Fatalf("expected gpu name in summary: %s", buf.String())
	}
}
