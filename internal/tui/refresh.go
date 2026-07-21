package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/baselhusam/bareai-cli/internal/collect"
	"github.com/baselhusam/bareai-cli/internal/inspect"
)

func tickCmd(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func collectSnapshotCmd(parent context.Context, timeout time.Duration, gen uint64, full bool) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(parent, timeout)
		defer cancel()

		opts := collect.LightRefreshOptions()
		if full {
			opts = collect.FullOptions()
		}
		snap := collect.SnapshotWithOptions(ctx, opts)
		inspect.Enrich(snap)
		return snapshotMsg{gen: gen, snap: snap}
	}
}
