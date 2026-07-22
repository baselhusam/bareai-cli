<p align="center">
  <img src="branding/readme-header.svg" alt="bareai — a read-only CLI + TUI for solo AI engineers on bare metal" width="100%">
</p>

<p align="center">
  <a href="https://github.com/baselhusam/bareai-cli/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/baselhusam/bareai-cli/ci.yml?branch=main&style=flat-square&label=CI&logo=github&logoColor=eef0e6&labelColor=0b0d08&color=8bd450" alt="CI"></a>
  <a href="https://github.com/baselhusam/bareai-cli/releases/latest"><img src="https://img.shields.io/github/v/release/baselhusam/bareai-cli?style=flat-square&logo=github&logoColor=eef0e6&labelColor=0b0d08&color=8bd450" alt="Release"></a>
  <a href="https://pkg.go.dev/github.com/baselhusam/bareai-cli"><img src="https://img.shields.io/badge/Go-1.25+-8bd450?style=flat-square&logo=go&logoColor=eef0e6&labelColor=0b0d08" alt="Go 1.25+"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-8bd450?style=flat-square&labelColor=0b0d08" alt="License: MIT"></a>
  <img src="https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-5c8a2e?style=flat-square&labelColor=0b0d08" alt="Platforms">
  <img src="https://img.shields.io/badge/mode-inspect%20%2F%20probe%20only-3d5a20?style=flat-square&labelColor=0b0d08" alt="Read-only">
</p>

<p align="center">
  <strong>CLI + TUI for solo AI engineers inspecting bare-metal AI infrastructure.</strong><br>
  <sub>Host · GPU · Docker · LLM runtimes — correlated, read-only, scriptable.</sub>
</p>

<p align="center">
  <a href="#quick-start">Quick start</a> ·
  <a href="#installation">Install</a> ·
  <a href="#commands">Commands</a> ·
  <a href="#interactive-tui">TUI</a> ·
  <a href="ROADMAP.md">Roadmap</a> ·
  <a href="branding/">Branding</a>
</p>

---

`bareai` answers one question on a single machine: *what is this box doing right now?*

It collects host resources, GPUs, Docker, and local LLM runtimes (Ollama, vLLM, SGLang, Triton, …), correlates them, and presents the result as human tables, JSON, or a live terminal dashboard.

| | |
|---|---|
| **Mode** | Inspect / probe only — no stop, restart, deploy, or other mutating ops |
| **Persona** | One engineer, one machine (SSH or local) |
| **Platforms** | Linux, macOS, Windows |
| **GPUs** | NVIDIA, AMD, Apple Silicon (degrades gracefully when absent) |
| **Output** | CLI tables · live TUI · `--json` for scripts and agents |

---

## Table of contents

- [What bareai does](#what-bareai-does)
- [Quick start](#quick-start)
- [Installation](#installation)
- [Configuration](#configuration)
- [Global flags](#global-flags)
- [Commands](#commands)
  - [bareai (default)](#bareai-default)
  - [status](#status)
  - [gpu](#gpu)
  - [docker](#docker)
  - [llm](#llm)
  - [probe](#probe)
  - [inspect](#inspect)
  - [doctor](#doctor)
  - [watch](#watch)
  - [config](#config)
  - [version](#version)
  - [completion](#completion)
- [Interactive TUI](#interactive-tui)
- [JSON output and snapshot model](#json-output-and-snapshot-model)
- [Platform capability matrix](#platform-capability-matrix)
- [Environment variables](#environment-variables)
- [Common workflows](#common-workflows)
- [Development](#development)
- [Branding](#branding)
- [License](#license)

---

## What bareai does

`bareai` is an AIOps inspection tool for AI boxes. It:

1. **Collects** read-only inventory from collectors (each optional; failures degrade gracefully):
   - **Host** — OS, CPU, RAM, disk, load, uptime
   - **GPU** — NVIDIA (`nvidia-smi`), AMD (sysfs/ROCm), Apple Silicon (best-effort)
   - **Docker** — Engine API: containers, images, volumes, NVIDIA runtime detection
   - **LLM** — discovery via Docker heuristics, process names, port scan, and config endpoints; HTTP health and model listing

2. **Correlates** endpoints → containers → PIDs → GPU index → VRAM (when data is available).

3. **Probes** inference servers with lightweight smoke tests (Ollama generate or OpenAI-compatible chat).

4. **Renders** the same underlying `Snapshot` model as:
   - CLI tables (`status`, `gpu`, …)
   - Full report (`inspect`, `doctor`)
   - Live TUI (`bareai`, `bareai watch`)
   - Machine-readable JSON (`--json` on any inspect command)

When a collector cannot run (no Docker, no GPU driver, permission denied), the command still exits **0** and records a `skipped` entry explaining why.

```
┌─────────────┐   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐
│    Host     │   │     GPU     │   │   Docker    │   │     LLM     │
└──────┬──────┘   └──────┬──────┘   └──────┬──────┘   └──────┬──────┘
       │                 │                 │                 │
       └─────────────────┴────────┬────────┴─────────────────┘
                                  ▼
                         ┌────────────────┐
                         │    Snapshot    │
                         └───────┬────────┘
              ┌──────────────────┼──────────────────┐
              ▼                  ▼                  ▼
           CLI / tables      Live TUI           --json
```

---

## Quick start

```bash
# Build from source
make build

# One-screen summary
./bareai status

# Full correlated report
./bareai inspect

# Live dashboard (TTY required)
./bareai

# Machine-readable output for scripts/agents
./bareai inspect --json | jq '.correlations'
```

On a TTY, running `./bareai` with no subcommand launches the interactive dashboard. In pipes or CI, bare `bareai` prints help; `bareai watch` falls back to `bareai status`.

---

## Installation

### macOS / Linux (install script)

```bash
curl -fsSL https://raw.githubusercontent.com/baselhusam/bareai-cli/main/scripts/install.sh | bash
```

Pin a version: `VERSION=v0.1.0 curl -fsSL ... | bash`

### Homebrew

```bash
brew tap baselhusam/tap
brew install bareai
```

### Windows

**winget:**

```powershell
winget install baselhusam.bareai
```

**PowerShell install script:**

```powershell
irm https://raw.githubusercontent.com/baselhusam/bareai-cli/main/scripts/install.ps1 | iex
```

Add to PATH: `.\scripts\install.ps1 -AddToPath`

### Debian / Ubuntu (Cloudsmith APT)

```bash
curl -1sLf 'https://dl.cloudsmith.io/public/baselhusam/bareai/cfg/setup/deb.sh' | sudo bash
sudo apt update && sudo apt install bareai
man bareai-doctor
```

### Manual download

Download archives and `checksums.txt` from [GitHub Releases](https://github.com/baselhusam/bareai-cli/releases). Verify SHA256 before installing.

### Shell completions

```bash
bareai completion bash >> ~/.bashrc
bareai completion zsh >> ~/.zshrc
bareai completion fish > ~/.config/fish/completions/bareai.fish
bareai completion powershell >> $PROFILE
```

### Build from source

**Requirements:** Go 1.25+

```bash
git clone https://github.com/baselhusam/bareai-cli.git
cd bareai-cli
make build
./bareai version
```

Release process for maintainers: [docs/RELEASE.md](docs/RELEASE.md).

---

## Configuration

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

### Example config

See [`config.example.yaml`](config.example.yaml):

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

---

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

---

## Commands

### Command overview

| Command | Purpose |
|---------|---------|
| *(none)* | Launch live TUI dashboard on a TTY; otherwise show help |
| `status` | One-screen host + infrastructure summary |
| `gpu` | GPU/accelerator inventory and metrics |
| `docker` | Docker Engine state (containers, optional images/volumes) |
| `llm` | Discovered inference servers, health, models, metrics |
| `probe` | One-hit smoke tests against discovered or explicit endpoints |
| `inspect` | Full correlated report (overview, correlation table, sections, findings) |
| `doctor` | Ranked diagnostics with read-only what/why/try hints |
| `watch` | Live TUI with configurable refresh interval |
| `config path` | Print resolved config file path |
| `version` | Print version, commit, build date, GOOS/GOARCH |
| `completion` | Generate shell completion scripts |

---

### bareai (default)

```bash
bareai
bareai --no-color
bareai --timeout 15s
```

**Behavior:**

- **TTY (interactive terminal):** Opens the Bubble Tea dashboard (same app as `bareai watch`, refresh from config `defaults.refresh`, default `3s`).
- **Non-TTY (pipe, CI, redirect):** Prints help and exits.

Use `bareai watch` when you want an explicit subcommand or `--refresh` on the command line.

---

### status

```bash
bareai status
bareai status --json
```

**What it collects:** Host, GPU, Docker, and LLM discovery **without** LLM health probes or model listing (fast summary).

**Human output includes:**

- Host: hostname, OS, CPU, RAM, load, disks, uptime
- GPU summary (count, util, VRAM per device)
- Docker summary (engine version, running container count)
- LLM summary (discovered runtimes, count)
- `Skipped` section for failed collectors

**Example:**

```text
$ bareai status
bareai status
Collected: 2026-07-22T12:00:00Z

Host
  Hostname:  ai-box
  OS:        linux 24.04 (ubuntu)
  CPU:       AMD EPYC (64 cores)
  Memory:    128.0 GiB / 512.0 GiB
  Load:      2.10 1.80 1.50

GPUs:
  [0] NVIDIA A100-SXM4-80GB (nvidia)  util=45%  mem=4.0 GiB/80.0 GiB

Docker:      2 running / 4 total  (Engine 27.5.1)
LLM:         2 discovered (ollama, vllm)
```

---

### gpu

```bash
bareai gpu
bareai gpu --json
```

**What it collects:** NVIDIA, AMD, and Apple Silicon accelerators only.

**Human output per GPU:**

- Index, vendor, name, UUID, driver
- Memory used / total (or “unified” on Apple)
- Utilization %, temperature, power draw/limit
- Processes: PID, name, VRAM used

**Example:**

```text
$ bareai gpu
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

GPU ↔ process correlation uses NVIDIA process lists when available; AMD/Apple join is best-effort.

---

### docker

```bash
bareai docker
bareai docker --all
bareai docker --images --volumes
bareai docker --json
```

**What it collects:** Read-only Docker Engine inventory via `DOCKER_HOST` (unix socket on Linux/macOS, named pipe on Windows with Docker Desktop).

| Flag | Short | Description |
|------|-------|-------------|
| `--all` | `-a` | Include stopped/exited containers in human output |
| `--images` | | List local images in human output |
| `--volumes` | | List named volumes in human output |

Human output shows **running containers by default**. `--json` always returns the full inventory (all containers, images, volumes).

**Human output includes:**

- Engine version, API version, OS/arch
- Default runtime and whether NVIDIA Container Toolkit runtime is present
- Container table: name, image, state, ports, GPU requested
- Image/volume counts (detail with flags above)

**Example:**

```text
$ bareai docker
Engine: Docker 27.5.1  api 1.47  linux/amd64
Runtime: runc (default)  nvidia: yes

Containers (2 running / 4 total)
  NAME            IMAGE              STATE    PORTS                  GPU
  ollama          ollama/ollama        running  11434->11434/tcp       no
  vllm            vllm/vllm-openai     running  8000->8000/tcp         yes

Images: 12  (pass --images for detail)
Volumes: 4  (pass --volumes for detail)
```

When the daemon is absent, exits `0` with a clear message in `Skipped`.

---

### llm

```bash
bareai llm
bareai llm --json
```

**What it collects:** Full LLM discovery with health probes, model listing, and exporter metrics when available.

**Discovery sources:**

1. Docker container names/images (Ollama, vLLM, SGLang, Triton heuristics)
2. Running processes (binary names)
3. Port scan on configured ports (`11434`, `8000`, `30000`, …)
4. Explicit `discovery.endpoints` in config

**Human output per runtime:**

- Runtime name, endpoint, source (docker / process / port)
- Health probe result (ok/fail, latency, message)
- Model list
- GPU index when correlated
- Prometheus-style metrics when exposed (KV cache, queue depth, tok/s, etc.)

**Example:**

```text
$ bareai llm
Ollama  http://127.0.0.1:11434  (docker: ollama)
  Health: ok  42ms  tags reachable
  Models: llama3.2, qwen2.5:7b
  GPU: 0

vLLM  http://127.0.0.1:8000  (process pid 1234)
  Health: ok  88ms  models reachable
  Models: meta-llama/Llama-3.1-8B
  Metrics: num_requests_running=2  gpu_cache_usage=0.45
```

---

### probe

```bash
bareai probe
bareai probe --endpoint http://127.0.0.1:11434 --runtime ollama
bareai probe --endpoint http://127.0.0.1:8000 --runtime vllm --model meta-llama/Llama-3.1-8B --prompt "Hi"
bareai probe --json
```

**What it does:** Sends a lightweight smoke request to inference endpoints. Failures are reported as probe results; the command exits **0** (never crashes on unreachable endpoints).

| Flag | Default | Description |
|------|---------|-------------|
| `--endpoint` | *(empty)* | Probe a specific URL instead of auto-discovery |
| `--runtime` | *(empty)* | Adapter when using `--endpoint`: `ollama`, `vllm`, `sglang`, `triton` (auto-detected if omitted) |
| `--model` | config `probe.model` | Model ID for the smoke request; empty picks first available |
| `--prompt` | config `probe.prompt` (`Hello`) | Prompt text for completion/generate |

**Without `--endpoint`:** Discovers all LLMs (same as `llm`) and runs smoke tests on each.

**Probe behavior by runtime:**

- **Ollama:** `/api/generate` or tags health
- **OpenAI-compatible (vLLM, SGLang, …):** `/v1/chat/completions` smoke request

---

### inspect

```bash
bareai inspect
bareai inspect --json
```

**What it collects:** Full snapshot (host, GPU, Docker, LLM with probes/models/metrics) plus correlation and informational findings.

**Human output sections:**

1. **Overview** — hostname, GPU count, Docker running count, LLM count
2. **Correlation table** — endpoint → runtime → container → PID → GPU → VRAM → models
3. **Host / GPU / Docker / LLM** detail sections
4. **Findings** — lightweight informational diagnostics
5. **Skipped** — collectors that could not run

Human layout adapts to terminal width (`COLUMNS` or TTY size). `--json` includes `correlations` and `findings` arrays with full fidelity.

**Example:**

```text
$ bareai inspect
Overview
  Host: ai-box   GPUs: 1   Docker: 2 running   LLMs: 2

Correlation
  ENDPOINT                      RUNTIME  CONTAINER  PID   GPU  VRAM     MODELS
  http://127.0.0.1:11434        ollama   ollama     100   0    2.0 GiB  llama3.2

Findings
  [info] llm.multiple_runtimes: 2 LLM runtimes discovered on this host
```

Use `bareai doctor` for expanded what/why/try remediation hints (still read-only).

---

### doctor

```bash
bareai doctor
bareai doctor --severity warn
bareai doctor --severity critical
bareai doctor --json
```

**What it does:** Runs the full collectors + correlation, then applies ranked diagnostic rules. Reports findings with severity, explanation, and suggested next steps. **Suggestions are read-only** — `bareai` never mutates your system.

| Flag | Default | Description |
|------|---------|-------------|
| `--severity` | config `doctor.min_severity` (`info`) | Minimum severity: `info`, `warn`, or `critical` |

**Severity levels:**

| Level | Meaning |
|-------|---------|
| `info` | Informational (e.g. multiple runtimes, Apple GPU limits) |
| `warn` | Likely issue (unreachable endpoint, high RAM, no GPU runtime) |
| `critical` | Serious misconfiguration or failure |

**Example:**

```text
[warn] llm.unreachable — Ollama (http://127.0.0.1:11434) is unreachable
  Why: Health probe failed; endpoint may be down or blocked.
  Try: curl -s http://127.0.0.1:11434/api/tags  ·  bareai probe --endpoint ...
```

Filter in JSON:

```bash
bareai doctor --json | jq '.findings[] | select(.severity=="warn")'
```

---

### watch

```bash
bareai watch
bareai watch --refresh 5s
bareai watch --no-color
```

**What it does:** Launches the live Bubble Tea TUI dashboard.

| Flag | Default | Description |
|------|---------|-------------|
| `--refresh` | config `defaults.refresh` (`3s`) | Interval between background snapshot refreshes |

**Non-TTY behavior:** Falls back to `bareai status` (unlike bare `bareai`, which prints help).

See [Interactive TUI](#interactive-tui) for keyboard reference.

---

### config

```bash
bareai config path
```

Prints the resolved config file path (respects `BAREAI_CONFIG`, `XDG_CONFIG_HOME`, and platform defaults). Does not create the file.

---

### version

```bash
bareai version
bareai version --json
```

Prints build metadata: version, git commit, build date, `GOOS`, `GOARCH`.

---

### completion

```bash
bareai completion bash
bareai completion zsh
bareai completion fish
bareai completion powershell
```

Generates shell completion scripts. See [Installation → Shell completions](#shell-completions).

---

## Interactive TUI

The TUI is the default experience on a TTY. It reuses the same collectors and probe logic as the CLI — it is a **view** over `snapshot.Snapshot`, not a second implementation.

```bash
bareai              # default on TTY
bareai watch        # explicit, with --refresh
bareai watch --refresh 5s --no-color
```

### Layout

| Tab | Key | Content |
|-----|-----|---------|
| **Overview** | `1` | Dense live dashboard (see below) |
| **GPUs** | `2` | Searchable GPU list + detail pane |
| **LLMs** | `3` | Providers/runtimes, PIDs, health, models |
| **Docker** | `4` | Container list + detail (running by default) |
| **Probe** | `5` | Smoke-probe selected LLM |

### Overview dashboard

The Overview tab shows everything at a glance:

- **Host** — CPU load, RAM, primary disk with colored Unicode bars and sparklines
- **GPUs** — per-device util/VRAM bars, sparklines, temperature coloring
- **Providers / LLMs** — runtime, endpoint, PID, GPU#, health (ok/fail colors)
- **Correlation** — endpoint → container → PID → GPU → VRAM
- **Findings** — top diagnostics
- **Skipped** — collectors that failed (muted)

Sparklines track the last **~40 refresh samples** (load, RAM, GPU util/VRAM).

Bar/sparkline color thresholds: **70% / 90%** util/mem; temperature warn **≥75°C**, fail **≥85°C**.

### Keyboard reference

| Key | Overview | List tabs (GPU / LLM / Docker) | Detail pane |
|-----|----------|----------------------------------|-------------|
| `1`–`5`, `Tab`, `Shift+Tab` | Switch tab | Switch tab | Switch tab |
| `↑` / `↓`, `j` / `k` | Move section / row | Move selection | Scroll |
| `Enter`, `l` | Dive into matching tab with selection synced | Focus detail pane | — |
| `/` | Jump to LLMs tab + open filter | Fuzzy filter list | — |
| `Esc` | — | Clear filter / back to list | Back to list |
| `a` | — | Docker: toggle all vs running containers | — |
| `r` | Force full refresh | Force full refresh | Force full refresh |
| `p` | — | Probe selected LLM (LLMs tab) | — |
| `q`, `Ctrl+C` | Quit | Quit | Quit |
| `?` | Toggle context-sensitive help | Toggle help | Toggle help |

**Overview dive:** Press `Enter` on a GPU, LLM, or correlation row to jump to the matching tab with that item selected. Correlation rows prefer Docker when a container name matches; otherwise LLMs.

**Refresh behavior:** Background ticks use a light snapshot (metrics without re-listing all models) to stay fast; `r` forces a full collect.

---

## JSON output and snapshot model

Every inspect command supports `--json` (`-j`). Output is a single **`Snapshot`** object:

```json
{
  "collected_at": "2026-07-22T12:00:00Z",
  "host": { "hostname": "...", "mem_total_bytes": 0, "load1": 0.0 },
  "gpus": [{ "index": 0, "vendor": "nvidia", "utilization_pct": 45.0, "processes": [] }],
  "docker": { "available": true, "containers": [], "images": [], "volumes": [] },
  "llms": [{ "runtime": "ollama", "endpoint": "...", "health": { "ok": true, "latency_ms": 42 } }],
  "correlations": [{ "endpoint": "...", "pid": 123, "gpu_index": 0, "vram_bytes": 0 }],
  "findings": [{ "id": "...", "severity": "info", "summary": "...", "why": "...", "try": "..." }],
  "skipped": [{ "component": "docker", "reason": "..." }]
}
```

**Top-level fields:**

| Field | Present in | Description |
|-------|------------|-------------|
| `collected_at` | All commands | UTC timestamp of collection |
| `host` | status, inspect, doctor, … | Host inventory |
| `gpus` | gpu, status, inspect, … | Accelerator list |
| `docker` | docker, status, inspect, … | Engine inventory |
| `llms` | llm, probe, inspect, … | Discovered inference servers |
| `correlations` | inspect, doctor (after enrich) | Joined endpoint ↔ resource rows |
| `findings` | inspect, doctor | Diagnostic findings |
| `skipped` | All | Collectors that failed with reason |

Use JSON for scripting, CI checks, and agent tooling:

```bash
# Fail CI if any LLM health check failed
bareai llm --json | jq -e '.llms | all(.health.ok == true)'

# Export GPU util for monitoring
bareai gpu --json | jq '.gpus[] | {index, util: .utilization_pct, mem: .memory_used_bytes}'
```

Types are defined in [`internal/snapshot/snapshot.go`](internal/snapshot/snapshot.go).

---

## Platform capability matrix

**Primary path:** Linux bare-metal is the reference environment with full collector fidelity.

| Feature | Linux | macOS | Windows |
|---------|-------|-------|---------|
| **Host** (`status`) | Full: CPU, RAM, disk, load | Full; load may be 0 on some configs | Full; load averages show as n/a |
| **GPU** | NVIDIA via `nvidia-smi`; AMD via sysfs/ROCm | Apple Silicon chip name; NVIDIA if driver installed | NVIDIA if `nvidia-smi` installed |
| **Docker** | Full via unix socket | Full via Docker Desktop socket | Full via named pipe (`docker_engine`) |
| **LLM discovery** | Docker + process + port scan | Docker Desktop; native Ollama on `:11434` | Docker Desktop; process `.exe` names; port scan |
| **Probe / inspect / TUI** | Full | Full | Full; TUI via Windows Terminal |

### GPU detail by platform

| Platform | NVIDIA | AMD | Apple |
|----------|--------|-----|-------|
| Linux | Full via `nvidia-smi` (util, VRAM, temp, power, processes) | sysfs VRAM + temp; ROCm best-effort | n/a |
| macOS | If `nvidia-smi` installed | n/a | Chip name only; unified memory (no util/temp/power) |
| Windows | If `nvidia-smi` installed | n/a | n/a |

### Platform notes

- **Linux:** NVIDIA GPU↔process correlation via `nvidia-smi`; AMD metrics from `/sys/class/drm`
- **macOS:** Ollama commonly via Docker Desktop or native app; process scan may need permissions
- **Windows:** Docker Desktop required for containers; WSL2 engine not visible unless `DOCKER_HOST` is set; load averages unavailable

When collectors are unavailable, commands exit `0` with clear skip messages. Manual verification checklist: [docs/CHECKLIST.md](docs/CHECKLIST.md).

---

## Environment variables

| Variable | Description |
|----------|-------------|
| `BAREAI_CONFIG` | Override path to config YAML |
| `XDG_CONFIG_HOME` | Base directory for config when `BAREAI_CONFIG` is unset |
| `DOCKER_HOST` | Docker Engine address (standard Docker CLI variable) |
| `COLUMNS` | Terminal width hint for `inspect` / `doctor` human layout |

---

## Common workflows

### SSH into an AI box and see everything

```bash
ssh ai-box
bareai                    # live dashboard
# or
bareai inspect            # one-shot full report
```

### Debug a unreachable LLM

```bash
bareai llm
bareai probe --endpoint http://127.0.0.1:11434 --runtime ollama
bareai doctor --severity warn
```

### Check GPU usage before starting a job

```bash
bareai gpu
bareai gpu --json | jq '.gpus[] | select(.utilization_pct > 80)'
```

### Scriptable health check

```bash
#!/usr/bin/env bash
set -euo pipefail
bareai probe --json | jq -e '.llms | length > 0 and all(.probe.ok // .health.ok)'
echo "All LLM probes passed"
```

### Monitor in TUI with slower refresh

```bash
bareai watch --refresh 10s
```

---

## Development

```bash
make build
make test
make test-integration   # subprocess smoke tests (requires built binary)
make smoke              # Linux/macOS checklist script
make lint               # requires golangci-lint
make man                # generate man pages into docs/man/man1/
make run ARGS="gpu"
make clean
```

Cross-platform manual checklist: [docs/CHECKLIST.md](docs/CHECKLIST.md)

Release and packaging: [docs/RELEASE.md](docs/RELEASE.md)

```bash
make goreleaser-check
make release-snapshot
```

Install `golangci-lint`:

```bash
brew install golangci-lint
# or see https://golangci-lint.run/welcome/install/
```

Architecture overview: [`.cursor/rules/architecture.mdc`](.cursor/rules/architecture.mdc) (collectors → snapshot → CLI/TUI/JSON).

---

## Branding

Assets live in [`branding/`](branding/). Preview locally by opening [`branding/index.html`](branding/index.html).

<p align="center">
  <img src="branding/logo-horizontal.svg" alt="bareai logo" width="280">
</p>

### Color palette

| Token | Hex | Swatch | Use |
|-------|-----|--------|-----|
| Ink | `#0b0d08` | ![#0b0d08](https://img.shields.io/badge/-0b0d08-0b0d08?style=flat-square) | Background / deep canvas |
| Surface | `#12140f` | ![#12140f](https://img.shields.io/badge/-12140f-12140f?style=flat-square) | Page / panel background |
| Panel | `#181b13` | ![#181b13](https://img.shields.io/badge/-181b13-181b13?style=flat-square) | Cards / elevated surfaces |
| Border | `#2a2f22` | ![#2a2f22](https://img.shields.io/badge/-2a2f22-2a2f22?style=flat-square) | Dividers / outlines |
| Text | `#eef0e6` | ![#eef0e6](https://img.shields.io/badge/-eef0e6-eef0e6?style=flat-square) | Primary text / logo bars |
| Muted | `#9aa48a` | ![#9aa48a](https://img.shields.io/badge/-9aa48a-9aa48a?style=flat-square) | Secondary / subtitle text |
| Accent | `#8bd450` | ![#8bd450](https://img.shields.io/badge/-8bd450-8bd450?style=flat-square) | Brand green (`ai` mark) |
| Accent mid | `#5c8a2e` | ![#5c8a2e](https://img.shields.io/badge/-5c8a2e-5c8a2e?style=flat-square) | Mid signal bar |
| Accent deep | `#3d5a20` | ![#3d5a20](https://img.shields.io/badge/-3d5a20-3d5a20?style=flat-square) | Deep signal bar |

### Asset map

| File | Size / role |
|------|-------------|
| [`readme-header.svg`](branding/readme-header.svg) | 1280×260 — GitHub README banner |
| [`banner-social.svg`](branding/banner-social.svg) | 1280×640 — social / OG image |
| [`logo-horizontal.svg`](branding/logo-horizontal.svg) | Dark background |
| [`logo-horizontal-light.svg`](branding/logo-horizontal-light.svg) | Light background |
| [`logo-stacked.svg`](branding/logo-stacked.svg) | Stacked mark + wordmark |
| [`icon.svg`](branding/icon.svg) / mono variants | App / favicon marks |
| [`app-icon.svg`](branding/app-icon.svg) / green | Squircle app icons |
| [`avatar.svg`](branding/avatar.svg) | Profile / org avatar |

---

## License

MIT — see [LICENSE](LICENSE).
