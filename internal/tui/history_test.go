package tui

import (
	"testing"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func TestMetricHistoryRecordHost(t *testing.T) {
	h := newMetricHistory(5)
	snap := snapshot.New()
	snap.Host = &snapshot.Host{
		CPUCores: 4,
		Load1:    2,
		MemTotal: 100,
		MemUsed:  40,
	}
	h.record(snap)
	if len(h.hostLoad()) != 1 || h.hostLoad()[0] != 50 {
		t.Fatalf("load sample = %v", h.hostLoad())
	}
	if len(h.hostMem()) != 1 || h.hostMem()[0] != 40 {
		t.Fatalf("mem sample = %v", h.hostMem())
	}
}

func TestMetricHistoryRecordGPU(t *testing.T) {
	h := newMetricHistory(5)
	util := 55.0
	snap := snapshot.New()
	snap.GPUs = []snapshot.GPU{
		{Index: 0, Utilization: &util, MemoryTotal: 1000, MemoryUsed: 500},
	}
	h.record(snap)
	if got := h.gpuUtil(0); len(got) != 1 || got[0] != 55 {
		t.Fatalf("gpu util = %v", got)
	}
	if got := h.gpuMem(0); len(got) != 1 || got[0] != 50 {
		t.Fatalf("gpu mem = %v", got)
	}
}

func TestMetricHistoryTruncates(t *testing.T) {
	h := newMetricHistory(3)
	snap := snapshot.New()
	snap.Host = &snapshot.Host{CPUCores: 1, Load1: 1, MemTotal: 100, MemUsed: 10}
	for i := 0; i < 5; i++ {
		h.record(snap)
	}
	if len(h.hostLoad()) > defaultHistoryMax {
		t.Fatalf("history not truncated: len=%d", len(h.hostLoad()))
	}
}

func TestMetricHistoryRemovesStaleGPU(t *testing.T) {
	h := newMetricHistory(5)
	util := 10.0
	h.record(&snapshot.Snapshot{
		GPUs: []snapshot.GPU{{Index: 0, Utilization: &util}},
	})
	h.record(&snapshot.Snapshot{GPUs: nil})
	if len(h.gpuUtil(0)) != 0 {
		t.Fatal("expected stale gpu history removed")
	}
}
