package mcp

import (
	"context"
	"time"

	"github.com/baselhusam/bareai-cli/internal/config"
)

const maxTimeoutSeconds = 120

// WithToolTimeout applies a per-tool timeout capped at maxTimeoutSeconds.
// When seconds is zero, config defaults are used.
func WithToolTimeout(ctx context.Context, seconds int) (context.Context, context.CancelFunc) {
	if seconds <= 0 {
		cfg := config.Global()
		d := cfg.Defaults.Timeout
		if d <= 0 {
			d = config.Default().Defaults.Timeout
		}
		return context.WithTimeout(ctx, d)
	}
	if seconds > maxTimeoutSeconds {
		seconds = maxTimeoutSeconds
	}
	return context.WithTimeout(ctx, time.Duration(seconds)*time.Second)
}
