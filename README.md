# bareai

CLI and TUI for solo AI engineers inspecting bare-metal AI infrastructure: host resources, GPUs (NVIDIA / AMD / Apple), Docker, and local LLM runtimes (Ollama, vLLM, SGLang, Triton, …).

**Status:** Phase 7 complete — cross-platform hardening with full capability matrix. See [ROADMAP.md](ROADMAP.md).

**Repository:** [github.com/baselhusam/bareai-cli](https://github.com/baselhusam/bareai-cli)

## Requirements

- Go 1.23+

## Build

```bash
make build
# or
go build -o bareai ./cmd/bareai
```

## Usage

```bash
./bareai --help
./bareai status
./bareai status --json
./bareai gpu
./bareai gpu --json
./bareai docker
./bareai docker --all --images --volumes
./bareai docker --json
./bareai llm
./bareai llm --json
./bareai probe
./bareai probe --endpoint http://127.0.0.1:11434 --runtime ollama
./bareai probe --json
./bareai inspect
./bareai inspect --json
./bareai watch
./bareai watch --refresh 5s
```

### Example

```text
$ bareai gpu
bareai gpu
Collected: 2026-07-22T12:00:00Z

GPU 0 (nvidia)
  Name:      NVIDIA A100-SXM4-80GB
  UUID:      GPU-abc-123
  Driver:    535.54
  Memory:    4.0 GiB / 80.0 GiB
  Util:      45%
  Temp:      55 C
  Power:     250 W / 300 W
  Processes:
    pid 1234  python  2.0 GiB
```

Use `--json` for machine-readable output (scripts, agents, CI).

### Platform capability matrix

**Primary path:** Linux bare-metal is the reference AIOps environment with full collector fidelity.

| Feature | Linux | macOS | Windows |
|---------|-------|-------|---------|
| **Host** (`status`) | Full: CPU, RAM, disk, load | Full; load may be 0 on some configs | Full; load averages show as n/a |
| **GPU** | NVIDIA via `nvidia-smi`; AMD via sysfs/ROCm | Apple Silicon chip name; NVIDIA if driver installed | NVIDIA if `nvidia-smi` installed |
| **Docker** | Full via unix socket | Full via Docker Desktop socket | Full via named pipe (`docker_engine`) |
| **LLM discovery** | Docker + process + port scan | Docker Desktop; native Ollama on `:11434` | Docker Desktop; process `.exe` names; port scan |
| **Probe / inspect / TUI** | Full | Full | Full; TUI via Windows Terminal |

#### GPU detail

| Platform | NVIDIA | AMD | Apple |
|----------|--------|-----|-------|
| Linux | Full via `nvidia-smi` (util, VRAM, temp, power, processes) | sysfs VRAM + temp; `rocm-smi` JSON deferred to sysfs fallback | n/a |
| macOS | If `nvidia-smi` installed | n/a | Chip name only; unified memory (no util/temp/power) |
| Windows | If `nvidia-smi` installed | n/a | n/a |

#### Platform notes

- **Linux:** NVIDIA GPU↔process correlation via `nvidia-smi`; AMD metrics from `/sys/class/drm`
- **macOS:** Ollama commonly via Docker Desktop or native app; process scan may need permissions
- **Windows:** Docker Desktop required for containers; WSL2 engine not visible unless `DOCKER_HOST` is set; load averages unavailable

When collectors are unavailable, commands exit `0` with clear skip messages. See [docs/CHECKLIST.md](docs/CHECKLIST.md) for per-OS manual verification steps.

### Docker

`bareai docker` connects to the local Docker Engine via `DOCKER_HOST` (unix socket on Linux/macOS, named pipe on Windows with Docker Desktop). When the daemon is absent, the command exits `0` with a clear message.

Human output shows running containers by default; use `--all` for stopped containers, `--images` and `--volumes` for detail lists. `--json` always returns full inventory.

```text
$ bareai docker
bareai docker
Collected: 2026-07-22T12:00:00Z

Engine: Docker 27.5.1  api 1.47  linux/amd64
Runtime: runc (default)  nvidia: yes

Containers (2 running / 4 total)
  NAME            IMAGE              STATE    PORTS                  GPU
  ollama          ollama/ollama        running  11434->11434/tcp       no
  vllm            vllm/vllm-openai     running  8000->8000/tcp         yes

Images: 12  (pass --images for detail)
Volumes: 4  (pass --volumes for detail)
```

### LLM discovery and probe

`bareai llm` discovers local inference servers via Docker heuristics, process names, and well-known ports, then probes health and lists models. `bareai probe` runs one-hit smoke tests (Ollama generate or OpenAI-compatible chat completion).

GPU correlation uses NVIDIA process lists when available; AMD/Apple GPU join is best-effort.

```text
$ bareai llm
bareai llm
Collected: 2026-07-22T12:00:00Z

Ollama  http://127.0.0.1:11434  (docker: ollama)
  Health: ok  42ms  tags reachable
  Models: llama3.2, qwen2.5:7b
  GPU: 0

vLLM  http://127.0.0.1:8000  (process pid 1234)
  Health: ok  88ms  models reachable
  Models: meta-llama/Llama-3.1-8B
```

Probe flags: `--endpoint`, `--runtime`, `--model`, `--prompt`. Probe failures are reported as results; the command exits `0`.

### Inspect

`bareai inspect` is the full correlated report: overview, correlation table (endpoint → container → PID → GPU → VRAM), GPU/LLM/Docker sections, and informational findings. Human output adapts to terminal width (`COLUMNS` or TTY size); `--json` includes `correlations` and `findings` arrays.

```text
$ bareai inspect
bareai inspect
Collected: 2026-07-22T12:00:00Z

Overview
  Host: ai-box   GPUs: 1   Docker: 2 running   LLMs: 2

Correlation
  ENDPOINT                      RUNTIME  CONTAINER  PID   GPU  VRAM     MODELS
  http://127.0.0.1:11434        ollama   ollama     100   0    2.0 GiB  llama3.2

Findings
  [info] llm.multiple_runtimes: 2 LLM runtimes discovered on this host
```

Findings are informational only (no mutating suggestions until Phase 9 doctor).

### Interactive TUI

On a TTY, bare `bareai` launches the live dashboard. Use `bareai watch` explicitly with `--refresh` to control the snapshot interval (default `3s`).

When stdout is not a terminal (pipes, CI), `bareai watch` falls back to `bareai status`; bare `bareai` shows help.

| Key | Action |
|-----|--------|
| `1`–`5`, `Tab` | Switch tab (Overview · GPUs · LLMs · Docker · Probe) |
| `↑`/`↓`, `j`/`k` | Move selection in list tabs |
| `Enter` | Focus detail pane (scroll with arrows) |
| `Esc` | Return focus to list |
| `r` | Force refresh |
| `p` | Run smoke probe on selected LLM (LLMs / Probe tabs) |
| `q`, `Ctrl+C` | Quit |
| `?` | Toggle key help |

The TUI reuses the same collectors and probe logic as the CLI; it is a view over `snapshot.Snapshot` only.

### Commands

| Command   | Description                                      | Status   |
|-----------|--------------------------------------------------|----------|
| `status`  | Host and infrastructure summary                  | Phase 1–5 |
| `gpu`     | GPU and accelerator details                      | Phase 2  |
| `docker`  | Docker containers, images, and volumes           | Phase 3  |
| `llm`     | Discovered LLM runtimes and models               | Phase 4  |
| `probe`   | One-hit smoke tests against discovered LLMs      | Phase 4  |
| `inspect` | Full correlated infrastructure report            | Phase 5  |
| `watch`   | Live TUI monitoring dashboard                    | Phase 6  |

### Global flags

| Flag          | Short | Default | Description                          |
|---------------|-------|---------|--------------------------------------|
| `--json`      | `-j`  | false   | Output in JSON format                |
| `--timeout`   |       | `10s`   | Timeout for probes and API calls     |
| `--no-color`  |       | false   | Disable colored output               |

`bareai watch` also accepts `--refresh` (default `3s`) for the live snapshot interval.

## Development

```bash
make test
make test-integration   # subprocess smoke tests (requires built binary)
make smoke              # Linux/macOS checklist script
make lint    # requires golangci-lint
make run ARGS="gpu"
make clean
```

Cross-platform manual checklist: [docs/CHECKLIST.md](docs/CHECKLIST.md)

Install `golangci-lint`:

```bash
brew install golangci-lint
# or see https://golangci-lint.run/welcome/install/
```

## Install (coming soon)

Package distribution is planned for Phase 8:

- Homebrew (`brew install …`)
- APT (`apt install …`)
- winget (Windows)

For now, build from source.

## License

MIT — see [LICENSE](LICENSE).
