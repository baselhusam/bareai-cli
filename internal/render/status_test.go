package render

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func TestWriteStatus(t *testing.T) {
	snap := &snapshot.Snapshot{
		CollectedAt: time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC),
		Host: &snapshot.Host{
			Hostname:   "test-box",
			OS:         "darwin",
			Platform:   "darwin",
			PlatformVer: "15.0",
			Arch:       "arm64",
			Uptime:     2*time.Hour + 30*time.Minute,
			CPUModel:   "Apple M1",
			CPUCores:   8,
			CPULogical: 8,
			Load1:      1.5,
			MemTotal:   16 * giB,
			MemUsed:    8 * giB,
			MemAvailable: 8 * giB,
			Disks: []snapshot.Disk{
				{Mount: "/", FSType: "apfs", Total: 500 * giB, Used: 100 * giB, Free: 400 * giB},
			},
		},
	}

	var buf bytes.Buffer
	if err := WriteStatus(&buf, snap, true); err != nil {
		t.Fatalf("WriteStatus failed: %v", err)
	}

	out := buf.String()
	for _, want := range []string{
		"bareai status",
		"test-box",
		"Apple M1",
		"none detected",
		"none discovered",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}

func TestWriteStatusWithGPUs(t *testing.T) {
	snap := &snapshot.Snapshot{
		CollectedAt: time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC),
		GPUs: []snapshot.GPU{{
			Index:  0,
			Vendor: "apple",
			Name:   "Apple M3 Max",
		}},
	}

	var buf bytes.Buffer
	if err := WriteStatus(&buf, snap, true); err != nil {
		t.Fatalf("WriteStatus failed: %v", err)
	}
	if !strings.Contains(buf.String(), "Apple M3 Max") {
		t.Fatalf("expected GPU in status output: %s", buf.String())
	}
}

func TestWriteJSON(t *testing.T) {
	snap := snapshot.New()
	snap.Host = &snapshot.Host{Hostname: "json-host"}

	var buf bytes.Buffer
	if err := WriteJSON(&buf, snap); err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"hostname": "json-host"`) {
		t.Errorf("expected hostname in JSON, got:\n%s", out)
	}
}

func TestFormatBytes(t *testing.T) {
	if got := formatBytes(2 * giB); got != "2.0 GiB" {
		t.Errorf("formatBytes(2 GiB) = %q", got)
	}
}
