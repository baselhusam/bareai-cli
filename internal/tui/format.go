package tui

import (
	"fmt"
	"strings"
)

const (
	giB = 1024 * 1024 * 1024
	miB = 1024 * 1024
)

func formatBytes(n uint64) string {
	switch {
	case n >= giB:
		return fmt.Sprintf("%.1f GiB", float64(n)/float64(giB))
	case n >= miB:
		return fmt.Sprintf("%.1f MiB", float64(n)/float64(miB))
	default:
		return fmt.Sprintf("%d B", n)
	}
}

func formatOptionalFloat(v *float64, fallback string) string {
	if v == nil {
		return fallback
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", *v), "0"), ".")
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}
