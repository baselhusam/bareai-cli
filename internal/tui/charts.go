package tui

import (
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	defaultBarWidth       = 10
	defaultSparkWidth     = 16
	thresholdOK           = 70.0
	thresholdWarn         = 90.0
	tempThresholdWarn     = 75.0
	tempThresholdFail     = 85.0
)

func pctUsed(used, total uint64) float64 {
	if total == 0 {
		return 0
	}
	return float64(used) / float64(total) * 100
}

func loadPct(load1 float64, cores int) float64 {
	if cores <= 0 {
		return math.Min(load1*100, 100)
	}
	return math.Min(load1/float64(cores)*100, 100)
}

func pressureStyle(s styles, pct float64) lipgloss.Style {
	switch {
	case pct >= thresholdWarn:
		return s.fail
	case pct >= thresholdOK:
		return s.warn
	default:
		return s.ok
	}
}

func tempStyle(s styles, temp float64) lipgloss.Style {
	switch {
	case temp >= tempThresholdFail:
		return s.fail
	case temp >= tempThresholdWarn:
		return s.warn
	default:
		return s.ok
	}
}

func renderBar(s styles, pct float64, width int) string {
	if width <= 0 {
		width = defaultBarWidth
	}
	if width < 1 {
		width = 1
	}
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	filled := int(math.Round(pct / 100 * float64(width)))
	if filled > width {
		filled = width
	}
	empty := width - filled
	barStyle := pressureStyle(s, pct)
	filledStr := barStyle.Render(strings.Repeat("█", filled))
	emptyStr := s.barEmpty.Render(strings.Repeat("░", empty))
	return filledStr + emptyStr
}

func renderSparkline(s styles, samples []float64, width int) string {
	if width <= 0 {
		width = defaultSparkWidth
	}
	if len(samples) == 0 {
		return s.muted.Render(strings.Repeat("·", width))
	}

	start := 0
	if len(samples) > width {
		start = len(samples) - width
	}
	window := samples[start:]

	minV, maxV := window[0], window[0]
	for _, v := range window {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	span := maxV - minV
	if span < 1 {
		span = 1
	}

	const blocks = "▁▂▃▄▅▆▇█"
	out := make([]rune, len(window))
	for i, v := range window {
		norm := (v - minV) / span
		idx := int(math.Round(norm * float64(len(blocks)-1)))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(blocks) {
			idx = len(blocks) - 1
		}
		out[i] = rune(blocks[idx])
	}
	return s.spark.Render(string(out))
}

func barWidthForTerminal(termWidth int) int {
	if termWidth < 60 {
		return 6
	}
	if termWidth < 80 {
		return 8
	}
	return defaultBarWidth
}

func sparkWidthForTerminal(termWidth int) int {
	if termWidth < 60 {
		return 10
	}
	if termWidth < 80 {
		return 12
	}
	return defaultSparkWidth
}
