package tui

import "time"

// Options configures the TUI runtime.
type Options struct {
	Timeout time.Duration
	Refresh time.Duration
	NoColor bool
	Force   bool // launch TUI even when stdout is not a TTY
}
