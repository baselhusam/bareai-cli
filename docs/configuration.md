# Configuration

Optional YAML config file. CLI flags always override config values.

| Location | Path |
|----------|------|
| Linux / macOS (default) | `~/.config/bareai/config.yaml` |
| Windows (default) | `%AppData%\bareai\config.yaml` |
| Override | Set `BAREAI_CONFIG` to any file path |

Show the resolved path:

```bash
bareai config path
```

## Example config

See [`config.example.yaml`](../config.example.yaml):

```yaml
defaults:
  timeout: 10s      # probe/API timeout (also --timeout)
  refresh: 3s       # TUI refresh interval (also watch --refresh)
  no_color: false   # also --no-color

probe:
  prompt: "Hello"   # default smoke-test prompt (also probe --prompt)
  model: ""         # default model; empty = auto-pick first available

discovery:
  ports:
    - 11434         # Ollama
    - 8000          # vLLM
    - 30000         # SGLang
  endpoints: []     # explicit URLs to always probe

doctor:
  min_severity: info   # info | warn | critical (also doctor --severity)

output:
  json_indent: true      # pretty-print JSON snapshots
```

## Global flags

These flags apply to all commands (persistent on the root command):

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--json` | `-j` | `false` | Emit machine-readable JSON instead of human tables |
| `--timeout` | | `10s` | Timeout for HTTP probes, Docker API, and collection |
| `--no-color` | | `false` | Disable ANSI colors (CLI and TUI) |

```bash
bareai status --json
bareai llm --timeout 30s
bareai inspect --no-color
```

## Environment variables

| Variable | Description |
|----------|-------------|
| `BAREAI_CONFIG` | Override path to config YAML |
| `XDG_CONFIG_HOME` | Base directory for config when `BAREAI_CONFIG` is unset |
| `DOCKER_HOST` | Docker Engine address (standard Docker CLI variable) |
| `COLUMNS` | Terminal width hint for `inspect` / `doctor` human layout |
