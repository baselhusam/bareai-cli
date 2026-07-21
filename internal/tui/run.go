package tui

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	"github.com/baselhusam/bareai-cli/internal/collect"
	"github.com/baselhusam/bareai-cli/internal/render"
)

// IsTTY reports whether stdout is an interactive terminal.
func IsTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// RunFallbackStatus prints a non-interactive status report when the TUI cannot run.
func RunFallbackStatus(ctx context.Context, w io.Writer, timeout time.Duration, noColor bool) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	snap := collect.Snapshot(ctx)
	return render.WriteStatus(w, snap, noColor)
}

// Run launches the Bubble Tea dashboard.
func Run(ctx context.Context, opts Options) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if opts.Refresh <= 0 {
		opts.Refresh = 3 * time.Second
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 10 * time.Second
	}

	if !opts.Force && !IsTTY() {
		fmt.Fprintln(os.Stderr, "bareai: stdout is not a terminal; showing status instead")
		return RunFallbackStatus(ctx, os.Stdout, opts.Timeout, opts.NoColor)
	}

	m := newModel(ctx, opts)
	p := tea.NewProgram(&m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
