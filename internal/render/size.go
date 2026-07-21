package render

import "fmt"

const (
	giB = 1024 * 1024 * 1024
	miB = 1024 * 1024
)

// formatBytes returns a human-readable binary size string.
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
