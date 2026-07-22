# Commands

## Overview

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

Global flags (`--json`, `--timeout`, `--no-color`) are documented in [configuration.md](configuration.md).

---

## bareai (default)

```bash
bareai
bareai --no-color
bareai --timeout 15s
```

**Behavior:**

- **TTY (interactive terminal):** Opens the Bubble Tea dashboard (same app as `bareai watch`, refresh from config `defaults.refresh`, default `3s`).
- **Non-TTY (pipe, CI, redirect):** Prints help and exits.

Use `bareai watch` when you want an explicit subcommand or `--refresh` on the command line. See [tui.md](tui.md).

---

## status

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

## gpu

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

## docker

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

## llm

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

## probe

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

## inspect

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

## doctor

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

## watch

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

See [tui.md](tui.md) for keyboard reference.

---

## config

```bash
bareai config path
```

Prints the resolved config file path (respects `BAREAI_CONFIG`, `XDG_CONFIG_HOME`, and platform defaults). Does not create the file.

---

## version

```bash
bareai version
bareai version --json
```

Prints build metadata: version, git commit, build date, `GOOS`, `GOARCH`.

---

## completion

```bash
bareai completion bash
bareai completion zsh
bareai completion fish
bareai completion powershell
```

Generates shell completion scripts. See [install.md](install.md#shell-completions).
