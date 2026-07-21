//go:build integration

package smoke

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func bareaiBin(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("abs root: %v", err)
	}
	name := "bareai"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	bin := filepath.Join(root, name)
	if _, err := os.Stat(bin); err != nil {
		t.Fatalf("binary not found at %s: build with `go build -o %s ./cmd/bareai`", bin, name)
	}
	return bin
}

func runBareai(t *testing.T, args ...string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, bareaiBin(t), args...)
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bareai %v failed: %v\n%s", args, err, out)
	}
}

func TestAllCommandsExitZero(t *testing.T) {
	t.Parallel()
	commands := [][]string{
		{"status", "--json"},
		{"gpu", "--json"},
		{"docker", "--json"},
		{"llm", "--json"},
		{"inspect", "--json"},
		{"version", "--json"},
		{"probe", "--endpoint", "http://127.0.0.1:59999", "--runtime", "ollama", "--json"},
	}
	for _, args := range commands {
		args := args
		t.Run(args[0], func(t *testing.T) {
			t.Parallel()
			runBareai(t, args...)
		})
	}
}

func TestWatchNonTTYFallback(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, bareaiBin(t), "watch")
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	if err := cmd.Run(); err != nil {
		t.Fatalf("bareai watch fallback failed: %v", err)
	}
}
