package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func TestModelUpdateSnapshot(t *testing.T) {
	m := newModel(context.Background(), Options{Timeout: time.Second, Refresh: time.Second})
	m.ready = true
	m.width = 100
	m.height = 30

	snap := snapshot.New()
	snap.Host = &snapshot.Host{Hostname: "testbox"}
	snap.GPUs = []snapshot.GPU{{Index: 0, Name: "Test GPU", Vendor: "nvidia"}}

	updated, _ := m.Update(snapshotMsg{gen: 1, snap: snap})
	model := updated.(*Model)
	if model.snap == nil || model.snap.Host.Hostname != "testbox" {
		t.Fatalf("expected snapshot with hostname testbox")
	}
	if model.loading {
		t.Fatal("expected loading false after snapshot")
	}
}

func TestModelUpdateStaleSnapshot(t *testing.T) {
	m := newModel(context.Background(), Options{Timeout: time.Second, Refresh: time.Second})
	m.gen = 2

	updated, _ := m.Update(snapshotMsg{gen: 1, snap: snapshot.New()})
	model := updated.(*Model)
	if model.snap != nil {
		t.Fatal("expected stale snapshot to be ignored")
	}
}

func TestModelTabSwitch(t *testing.T) {
	m := newModel(context.Background(), Options{Timeout: time.Second, Refresh: time.Second})
	m.ready = true
	m.width = 100
	m.height = 30

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	model := updated.(*Model)
	if model.tab != TabLLM {
		t.Fatalf("expected LLM tab, got %v", model.tab)
	}
}

func TestModelViewContainsTabs(t *testing.T) {
	m := newModel(context.Background(), Options{Timeout: time.Second, Refresh: time.Second})
	m.ready = true
	m.width = 100
	m.height = 30
	m.snap = snapshot.New()

	view := m.View()
	for _, want := range []string{"Overview", "GPUs", "LLMs", "Docker", "Probe"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing tab label %q", want)
		}
	}
}

func TestOverviewText(t *testing.T) {
	snap := snapshot.New()
	snap.Host = &snapshot.Host{Hostname: "node1", OS: "linux", Platform: "ubuntu", CPUCores: 8, MemTotal: 16 * 1024 * 1024 * 1024, MemUsed: 8 * 1024 * 1024 * 1024}
	snap.GPUs = []snapshot.GPU{{Index: 0, Name: "GPU0", Vendor: "nvidia"}}
	text := overviewText(snap, 120)
	if !strings.Contains(text, "node1") {
		t.Fatalf("overview missing hostname: %s", text)
	}
	if !strings.Contains(text, "GPUs: 1") {
		t.Fatalf("overview missing gpu count: %s", text)
	}
}

func TestGPUListTitle(t *testing.T) {
	util := 42.0
	s := newStyles(false)
	title := gpuListTitle(snapshot.GPU{Index: 0, Name: "NVIDIA A100", Utilization: &util, MemoryTotal: 40 * 1024 * 1024 * 1024, MemoryUsed: 10 * 1024 * 1024 * 1024}, s, 10)
	if !strings.Contains(title, "42%") {
		t.Fatalf("unexpected title: %s", title)
	}
	if !strings.Contains(title, "█") {
		t.Fatalf("expected bar in title: %s", title)
	}
}
