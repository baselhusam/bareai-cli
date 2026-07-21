package tui

import "github.com/baselhusam/bareai-cli/internal/snapshot"

type snapshotMsg struct {
	gen  uint64
	snap *snapshot.Snapshot
}

type probeResultMsg struct {
	gen    uint64
	index  int
	result snapshot.ProbeResult
}

type tickMsg struct{}
