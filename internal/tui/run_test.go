package tui

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestRunFallbackStatus(t *testing.T) {
	var buf bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RunFallbackStatus(ctx, &buf, 5*time.Second, true); err != nil {
		t.Fatalf("RunFallbackStatus failed: %v", err)
	}
	if !strings.Contains(buf.String(), "bareai status") {
		t.Fatalf("expected status output, got: %s", buf.String())
	}
}

func TestRunNonTTYFallback(t *testing.T) {
	// stdout in tests is not a TTY; Run should fall back without hanging.
	err := Run(context.Background(), Options{
		Timeout: 5 * time.Second,
		Refresh: time.Second,
		NoColor: true,
	})
	if err != nil {
		t.Fatalf("Run fallback failed: %v", err)
	}
}
