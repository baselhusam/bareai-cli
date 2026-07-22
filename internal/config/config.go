package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds user settings for bareai.
type Config struct {
	Defaults  Defaults  `yaml:"defaults"`
	Probe     Probe     `yaml:"probe"`
	Discovery Discovery `yaml:"discovery"`
	Doctor    Doctor    `yaml:"doctor"`
	Actions   Actions   `yaml:"actions"`
	Output    Output    `yaml:"output"`
}

// Defaults holds global CLI/TUI defaults.
type Defaults struct {
	Timeout  time.Duration `yaml:"timeout"`
	Refresh  time.Duration `yaml:"refresh"`
	NoColor  bool          `yaml:"no_color"`
}

// Probe holds smoke-test defaults.
type Probe struct {
	Prompt string `yaml:"prompt"`
	Model  string `yaml:"model"`
}

// Discovery holds LLM discovery overrides.
type Discovery struct {
	Ports     []int    `yaml:"ports"`
	Endpoints []string `yaml:"endpoints"`
}

// Doctor holds doctor command defaults.
type Doctor struct {
	MinSeverity string `yaml:"min_severity"`
}

// Actions holds bareai do defaults.
type Actions struct {
	Confirm      bool `yaml:"confirm"`
	AutoReprobe  bool `yaml:"auto_reprobe"`
	LogTail      int  `yaml:"log_tail"`
	LogMaxBytes  int  `yaml:"log_max_bytes"`
}

// Output holds output formatting options.
type Output struct {
	JSONIndent bool `yaml:"json_indent"`
}

// Default returns built-in defaults.
func Default() Config {
	return Config{
		Defaults: Defaults{
			Timeout: 10 * time.Second,
			Refresh: 3 * time.Second,
			NoColor: false,
		},
		Probe: Probe{
			Prompt: "Hello",
			Model:  "",
		},
		Discovery: Discovery{
			Ports: []int{11434, 8000, 30000},
		},
		Doctor: Doctor{
			MinSeverity: "info",
		},
		Actions: Actions{
			Confirm:     true,
			AutoReprobe: true,
			LogTail:     100,
			LogMaxBytes: 262144,
		},
		Output: Output{
			JSONIndent: true,
		},
	}
}

// Path returns the resolved config file path (may not exist).
func Path() string {
	if p := os.Getenv("BAREAI_CONFIG"); p != "" {
		return p
	}
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("AppData"); appData != "" {
			return filepath.Join(appData, "bareai", "config.yaml")
		}
	}
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "bareai", "config.yaml")
}

// Load reads config from disk or returns defaults if missing.
func Load() (Config, error) {
	cfg := Default()
	path := Path()
	if path == "" {
		return cfg, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	if cfg.Defaults.Timeout <= 0 {
		cfg.Defaults.Timeout = Default().Defaults.Timeout
	}
	if cfg.Defaults.Refresh <= 0 {
		cfg.Defaults.Refresh = Default().Defaults.Refresh
	}
	if cfg.Probe.Prompt == "" {
		cfg.Probe.Prompt = Default().Probe.Prompt
	}
	if len(cfg.Discovery.Ports) == 0 {
		cfg.Discovery.Ports = Default().Discovery.Ports
	}
	if cfg.Doctor.MinSeverity == "" {
		cfg.Doctor.MinSeverity = Default().Doctor.MinSeverity
	}
	if cfg.Actions.LogTail <= 0 {
		cfg.Actions.LogTail = Default().Actions.LogTail
	}
	if cfg.Actions.LogMaxBytes <= 0 {
		cfg.Actions.LogMaxBytes = Default().Actions.LogMaxBytes
	}
	return cfg, nil
}

// SetGlobal stores the loaded config for package consumers.
var global Config

// Init loads config and stores it globally.
func Init() error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	global = cfg
	return nil
}

// Global returns the loaded config.
func Global() Config {
	return global
}

// SetDiscoveryForTest overrides discovery settings (tests only).
func SetDiscoveryForTest(d Discovery) {
	global.Discovery = d
}

// DiscoveryPorts returns configured discovery ports.
func DiscoveryPorts() []int {
	return global.Discovery.Ports
}

// DiscoveryEndpoints returns configured explicit endpoints.
func DiscoveryEndpoints() []string {
	return global.Discovery.Endpoints
}
