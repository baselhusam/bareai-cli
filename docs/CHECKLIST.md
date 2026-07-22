# bareai cross-platform checklist

Manual smoke scenarios to verify bareai on Linux, macOS, and Windows. All commands should exit **0** unless noted.

## All platforms

| Step | Command | Expected |
|------|---------|----------|
| Host JSON | `bareai status --json` | Exit 0; `host.hostname` populated |
| No GPU | `bareai gpu` | Exit 0; "none detected" or GPU list |
| Docker stopped | `bareai docker` (daemon not running) | Exit 0; "not available" or skip reason |
| LLM empty | `bareai llm` (no runtimes) | Exit 0; "none discovered" |
| Inspect | `bareai inspect --json` | Exit 0; snapshot object |
| Probe fail | `bareai probe --endpoint http://127.0.0.1:59999 --runtime ollama` | Exit 0; fail result in output |
| Watch pipe | `bareai watch \| cat` | Exit 0; status-style fallback (not TUI) |

## Linux (primary AIOps path)

Full collector fidelity: NVIDIA/AMD GPUs, unix Docker socket, process + port LLM discovery.

| Step | Command | Expected |
|------|---------|----------|
| Load avg | `bareai status` | Load line with three values (when supported) |
| NVIDIA | `bareai gpu` | GPU util/VRAM/temp when `nvidia-smi` present |
| AMD | `bareai gpu` | sysfs/ROCm metrics; `rocm-smi --json` util/name when available |
| Docker | `bareai docker` | Engine info + containers via `/var/run/docker.sock` |
| Ollama container | `bareai llm` | Discovers `http://127.0.0.1:11434` from Docker heuristics |
| Correlation | `bareai inspect` | Endpoint → container → PID → GPU table when workloads running |

## macOS

| Step | Command | Expected |
|------|---------|----------|
| Apple GPU | `bareai gpu` | Chip name + unified memory; richer identity via system_profiler; no util/temp/power |
| Docker Desktop | `bareai docker` | Connects via Docker Desktop socket |
| Ollama via Docker | `bareai llm` | Discovers Ollama container when running |
| Native Ollama | `bareai llm` | Discovers via port `:11434` or process `ollama` |
| TTY watch | `bareai watch` | TUI launches in Terminal/iTerm |

**Notes:** Apple GPUs do not report util/temp/power/process lists. Process scanning may require permissions for some PIDs.

## Windows

| Step | Command | Expected |
|------|---------|----------|
| Load avg | `bareai status` | `Load: n/a (not available on this platform)` |
| NVIDIA | `bareai gpu` | Works when `nvidia-smi` in PATH |
| Docker Desktop | `bareai docker` | Connects via `npipe:////./pipe/docker_engine` |
| Docker stopped | Stop Docker Desktop, run `bareai docker` | Exit 0; clear unavailable message |
| Process `.exe` | `bareai llm` | Discovers `ollama.exe` / cmdline heuristics |
| TTY watch | `bareai watch` in Windows Terminal | TUI launches without panic |

**Notes:** Docker Desktop required for container inventory. WSL2-internal Docker engine is not discovered from Windows host unless `DOCKER_HOST` points to it. Load averages are not available via gopsutil on Windows.

## Phase 9 — Manual checks

| Step | Command | Expected |
|------|---------|----------|
| Config path | `bareai config path` | Prints resolved config file location |
| Doctor | `bareai doctor` | Ranked findings with Why/Try hints |
| Doctor JSON | `bareai doctor --json` | Valid JSON with `findings` array |
| Custom ports | Set `discovery.ports` in config, run `bareai llm` | Discovers configured endpoint |
| TUI refresh | `bareai watch --refresh 5s` | Periodic refresh without heavy model re-list |
| Man pages | `man bareai-doctor` (after `.deb` install) | Man page renders |

## Automated scripts

```bash
# Linux / macOS
make smoke

# Windows (PowerShell)
make smoke-windows
```

Or run integration tests:

```bash
make test-integration
```
