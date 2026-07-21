package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadMissingFileUsesDefaults(t *testing.T) {
	t.Setenv("BAREAI_CONFIG", filepath.Join(t.TempDir(), "missing.yaml"))
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Defaults.Timeout != 10*time.Second {
		t.Fatalf("timeout = %v", cfg.Defaults.Timeout)
	}
	if len(cfg.Discovery.Ports) != 3 {
		t.Fatalf("ports = %v", cfg.Discovery.Ports)
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(`
defaults:
  refresh: 5s
probe:
  prompt: "ping"
discovery:
  ports: [9999]
doctor:
  min_severity: warn
`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BAREAI_CONFIG", path)
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Defaults.Refresh != 5*time.Second {
		t.Fatalf("refresh = %v", cfg.Defaults.Refresh)
	}
	if cfg.Probe.Prompt != "ping" {
		t.Fatalf("prompt = %q", cfg.Probe.Prompt)
	}
	if cfg.Discovery.Ports[0] != 9999 {
		t.Fatalf("ports = %v", cfg.Discovery.Ports)
	}
	if cfg.Doctor.MinSeverity != "warn" {
		t.Fatalf("min severity = %q", cfg.Doctor.MinSeverity)
	}
}

func TestPathRespectsEnv(t *testing.T) {
	t.Setenv("BAREAI_CONFIG", "/tmp/custom.yaml")
	if got := Path(); got != "/tmp/custom.yaml" {
		t.Fatalf("path = %q", got)
	}
}
