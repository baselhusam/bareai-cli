package cli

import (
	"context"
	"testing"
	"time"
)

func TestRootCommand(t *testing.T) {
	if rootCmd.Use != "bareai" {
		t.Fatalf("expected Use bareai, got %q", rootCmd.Use)
	}

	expected := map[string]bool{
		"status":  false,
		"gpu":     false,
		"llm":     false,
		"docker":  false,
		"inspect": false,
		"probe":   false,
		"watch":   false,
	}

	for _, cmd := range rootCmd.Commands() {
		if _, ok := expected[cmd.Name()]; ok {
			expected[cmd.Name()] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("missing subcommand: %s", name)
		}
	}
}

func TestWatchNonTTYFallback(t *testing.T) {
	watchCmd.SetContext(context.Background())
	if err := watchCmd.RunE(watchCmd, nil); err != nil {
		t.Fatalf("watch non-TTY fallback failed: %v", err)
	}
}

func TestWatchRefreshFlagDefault(t *testing.T) {
	if watchRefresh != 3*time.Second {
		t.Fatalf("expected default refresh 3s, got %v", watchRefresh)
	}
}
