package tui

import (
	"strings"
	"testing"
)

func TestRenderBarWidths(t *testing.T) {
	s := newStyles(false)
	tests := []struct {
		name    string
		pct     float64
		filled  int
	}{
		{"zero", 0, 0},
		{"half", 50, 5},
		{"full", 100, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderBar(s, tt.pct, 10)
			plain := stripANSI(got)
			if strings.Count(plain, "█") != tt.filled {
				t.Fatalf("filled=%d want %d: %q", strings.Count(plain, "█"), tt.filled, plain)
			}
			if strings.Count(plain, "█")+strings.Count(plain, "░") != 10 {
				t.Fatalf("bar width not 10: %q", plain)
			}
		})
	}
}

func TestRenderBarThresholdColors(t *testing.T) {
	s := newStyles(false)
	low := renderBar(s, 30, 10)
	mid := renderBar(s, 80, 10)
	high := renderBar(s, 95, 10)
	if low == mid || mid == high {
		t.Fatalf("expected different styles for thresholds")
	}
}

func TestRenderSparklineEmpty(t *testing.T) {
	s := newStyles(false)
	got := renderSparkline(s, nil, 8)
	if !strings.Contains(got, "·") {
		t.Fatalf("expected muted placeholder, got %q", got)
	}
}

func TestRenderSparklineSamples(t *testing.T) {
	s := newStyles(false)
	samples := []float64{10, 20, 30, 40, 50, 60, 70, 80}
	got := renderSparkline(s, samples, 8)
	plain := stripANSI(got)
	if len([]rune(plain)) != 8 {
		t.Fatalf("sparkline runes=%d want 8: %q", len([]rune(plain)), plain)
	}
}

func TestPctUsed(t *testing.T) {
	if pct := pctUsed(50, 100); pct != 50 {
		t.Fatalf("pctUsed = %v", pct)
	}
	if pct := pctUsed(1, 0); pct != 0 {
		t.Fatalf("pctUsed zero total = %v", pct)
	}
}

func TestLoadPct(t *testing.T) {
	if pct := loadPct(2, 4); pct != 50 {
		t.Fatalf("loadPct = %v", pct)
	}
	if pct := loadPct(2, 0); pct != 100 {
		t.Fatalf("loadPct no cores capped = %v", pct)
	}
	if pct := loadPct(5, 4); pct != 100 {
		t.Fatalf("loadPct capped = %v", pct)
	}
}

func stripANSI(s string) string {
	var b strings.Builder
	esc := false
	for _, r := range s {
		if esc {
			if r == 'm' {
				esc = false
			}
			continue
		}
		if r == '\x1b' {
			esc = true
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
