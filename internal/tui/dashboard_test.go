package tui

import (
	"strings"
	"testing"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func TestOverviewFocusMove(t *testing.T) {
	snap := snapshot.New()
	snap.Host = &snapshot.Host{Hostname: "node"}
	snap.GPUs = []snapshot.GPU{{Index: 0}, {Index: 1}}
	snap.LLMs = []snapshot.LLM{{Runtime: "ollama"}}

	f := overviewFocus{section: sectionHost, row: 0}
	f = f.moveDown(snap)
	if f.section != sectionGPU || f.row != 0 {
		t.Fatalf("moveDown host -> gpu: %+v", f)
	}
	f = f.moveUp(snap)
	if f.section != sectionHost {
		t.Fatalf("moveUp gpu -> host: %+v", f)
	}
}

func TestOverviewDiveTargetGPU(t *testing.T) {
	snap := snapshot.New()
	snap.GPUs = []snapshot.GPU{{Index: 0}, {Index: 1}}
	target, ok := overviewFocus{section: sectionGPU, row: 1}.diveTarget(snap)
	if !ok || target.tab != TabGPU || target.index != 1 {
		t.Fatalf("dive target = %+v ok=%v", target, ok)
	}
}

func TestRenderDashboardContainsBars(t *testing.T) {
	util := 42.0
	snap := snapshot.New()
	snap.Host = &snapshot.Host{
		Hostname: "node1",
		CPUCores: 4,
		Load1:    1.5,
		MemTotal: 16 * 1024 * 1024 * 1024,
		MemUsed:  8 * 1024 * 1024 * 1024,
	}
	snap.GPUs = []snapshot.GPU{
		{Index: 0, Name: "GPU0", Utilization: &util, MemoryTotal: 1000, MemoryUsed: 500},
	}
	snap.LLMs = []snapshot.LLM{{Runtime: "ollama", Name: "local", Endpoint: "http://127.0.0.1:11434", PID: 99}}

	h := newMetricHistory(5)
	h.record(snap)
	out := renderDashboard(snap, h, overviewFocus{section: sectionHost, row: 0}, 100, newStyles(false))
	for _, want := range []string{"Host", "GPUs", "Providers", "GiB", "ollama", "pid"} {
		if !strings.Contains(out, want) {
			t.Fatalf("dashboard missing %q in:\n%s", want, out)
		}
	}
}
