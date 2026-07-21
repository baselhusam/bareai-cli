# bareai Roadmap

CLI + TUI for solo AI engineers inspecting bare-metal AI boxes: host, GPUs (NVIDIA / AMD / Apple), Docker, and local LLM runtimes (Ollama, vLLM, SGLang, Triton, …).

**Stack:** Go · Cobra · Bubble Tea · GoReleaser  
**Mode:** inspect / probe only (no mutate)  
**Persona:** one engineer, one machine

---

## Phase 0 — Project skeleton

**Goal:** Compilable binary, empty command tree, CI-ready layout.

- [x] Init Go module (`bareai`), `cmd/bareai`, Cobra root command
- [x] Placeholder subcommands: `status`, `gpu`, `llm`, `docker`, `inspect`, `probe`, `watch`
- [x] Global flags: `--json`, `--timeout`, `--no-color`
- [x] Makefile / `task` targets: `build`, `test`, `lint`
- [x] Basic GitHub Actions: build + test on Linux / macOS / Windows
- [x] README with vision, install placeholders, command overview

**Exit:** `go build` produces `bareai`; `--help` documents the surface.

---

## Phase 1 — Host + snapshot core

**Goal:** Reliable host inventory and a shared snapshot model.

- [x] `internal/snapshot` types (Host, GPU, Docker, LLM, Findings)
- [x] Host collector: OS, arch, CPU count/model, load, RAM, disk, uptime
- [x] Platform adapters (Linux / macOS / Windows) with build tags
- [x] `bareai status` human table + `--json`
- [x] Graceful partial snapshots when a collector fails

**Exit:** `bareai status` works on Mac/Linux/Windows without GPU or Docker.

---

## Phase 2 — Multi-vendor GPU

**Goal:** Accelerator inventory that degrades cleanly per platform.

- [x] NVIDIA: NVML and/or `nvidia-smi` fallback (util, VRAM, temp, power, processes)
- [x] AMD: ROCm / sysfs-style metrics where available
- [x] Apple Silicon: best-effort GPU/memory reporting (document limits)
- [x] Multi-GPU listing + process ↔ device correlation when possible
- [x] `bareai gpu` command + status section

**Exit:** Correct vendor path on at least one real NVIDIA Linux box; stubs/clear messages elsewhere.

---

## Phase 3 — Docker inspection

**Goal:** Read-only Docker Engine view useful for AI workloads.

- [x] Connect via Docker API (socket / named pipe); skip if daemon absent
- [x] List containers (status, image, ports, GPU device requests when present)
- [x] List images and volumes (summary + optional detail flags)
- [x] Detect NVIDIA Container Toolkit / runtime presence (informational)
- [x] `bareai docker` + include in `inspect` / status

**Exit:** Useful Docker summary on a typical AI box; clean “Docker not available” elsewhere.

---

## Phase 4 — LLM discovery + probe

**Goal:** Find local inference servers and smoke-test them.

- [x] Discovery: process names, listening ports, Docker image/name heuristics
- [x] Adapters: Ollama, vLLM, SGLang, Triton (common interface)
- [x] HTTP probes: health, `/v1/models` (or vendor equivalent), Prometheus `/metrics` when present
- [x] Correlate model/server ↔ container ↔ GPU when possible
- [x] `bareai llm` — list servers, models, endpoints, basic load metrics if exposed
- [x] `bareai probe` — one-hit request (chat/completions or vendor API) with latency/status
- [x] Timeouts, no hanging; failures reported as probe results (not crashes)

**Exit:** Discovers at least Ollama + one OpenAI-compatible server; `probe` returns clear pass/fail.

---

## Phase 5 — `inspect` + correlation

**Goal:** One full picture for “what is this box doing?”

- [x] `bareai inspect` aggregates host + GPU + Docker + LLM into one report
- [x] Correlation table: endpoint → container → PID → GPU index → VRAM
- [x] Human layout optimized for SSH (width-aware); full fidelity in `--json`
- [x] Optional light findings (informational only): multiple heavy models, no GPU runtime, unreachable endpoint

**Exit:** Single command answers host/GPU/containers/models for a solo engineer’s box.

---

## Phase 6 — Interactive TUI

**Goal:** Default day-to-day UX: browse, select, monitor.

- [x] Bubble Tea app (`bareai` / `bareai watch`)
- [x] Panels or tabs: Overview · GPUs · LLMs · Docker · Probe
- [x] Keyboard navigation, selection, detail panes
- [x] Live refresh for metrics (configurable interval)
- [x] Trigger one-hit probe from TUI (same probe package as CLI)
- [x] Reuse collectors; TUI is a view, not a second implementation
- [x] Fallback: if not a TTY, print CLI help or `status`

**Exit:** Usable live TUI over SSH; selections and probe from the UI.

---

## Phase 7 — Cross-platform hardening

**Goal:** First-class Mac / Windows / Linux behavior.

- [x] Document per-OS capability matrix (what works where)
- [x] Windows: Docker Desktop pipe, NVIDIA where present
- [x] macOS: host + Docker + Ollama path; honest Apple GPU limits
- [x] Linux: primary AIOps path (full collectors)
- [x] Integration smoke tests / manual checklist per OS

**Exit:** Capability matrix in README; no surprise crashes on unsupported features.

---

## Phase 8 — Packaging & distribution

**Goal:** Install via brew, Windows package managers, and APT.

- [x] GoReleaser: multi-arch binaries (linux/darwin/windows, amd64/arm64)
- [x] Homebrew tap formula
- [x] winget manifest (Scoop optional)
- [x] `.deb` via nFPM + APT repo (Cloudsmith, GemFury, or self-hosted)
- [x] Optional `curl | sh` / PowerShell install scripts from GitHub Releases
- [x] Shell completions (bash/zsh/fish/powershell)
- [x] Signed checksums on releases

**Exit:** Documented one-liner installs for macOS, Windows, and Debian/Ubuntu.

---

## Phase 9 — Polish & doctor (still inspect-only)

**Goal:** Faster debugging without mutating the system.

- [ ] `bareai doctor` — ranked findings with “what/why/try” (read-only suggestions)
- [ ] Richer LLM metrics when exporters exist (KV cache, queue, tok/s)
- [ ] Config file (`~/.config/bareai/config.yaml`) for default probe prompts, ports, refresh
- [ ] Man pages / better `--help` examples
- [ ] Performance pass: low overhead on refresh loops

**Exit:** Doctor + config make daily AIOps inspection feel complete for one box.

---

## Later (backlog — not scheduled)

- Mutating actions (restart container, reload model) behind explicit confirmations
- Multi-host / SSH remote snapshot
- Prometheus metrics export
- MCP server so coding agents can call bareai
- Kubernetes / pod awareness (secondary to bare metal)
- Official distro packages (Debian/Ubuntu archives)

---

## Name notes

**Chosen:** `bareai` — short, signals bare-metal + AI.

**Alternatives** (if you want to bike-shed later): `barectl`, `metalai`, `boxai`, `aibox`, `inferbox`, `gpuctl`.

Stick with `bareai` unless packaging/trademark conflict appears.
