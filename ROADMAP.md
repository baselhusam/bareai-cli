# bareai Roadmap

CLI + TUI for solo AI engineers on bare-metal AI boxes: host, GPUs (NVIDIA / AMD / Apple), Docker, local LLM runtimes (Ollama, vLLM, SGLang, Triton, …), and related local services.

**Stack:** Go · Cobra · Bubble Tea · GoReleaser  
**Mode:** inspect / probe first; confirm-gated mutate only after the cockpit is addictive (Phase 12)  
**Persona:** one engineer, one machine (multi-host only after single-box is religious)

**North star:** the pane you never close → the sensor every agent trusts → carefully fix what you found.

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

## Phase 6.1 — Dashboard richness

**Goal:** Default `./bareai` feels like a live AIOps dashboard (hybrid layout + trends).

- [x] Overview dashboard: host/GPU/LLM/correlation panels with focus navigation
- [x] Unicode bars + in-memory sparklines (load, RAM, GPU util/VRAM)
- [x] Keyboard dive from Overview into GPU / LLM / Docker tabs with selection sync
- [x] Fuzzy search (`/`) on GPU, LLM, and Docker lists (PID, provider, endpoint, …)
- [x] Colored health/state, richer list titles and detail panes
- [x] Docker tab: toggle all vs running containers (`a`)

**Exit:** `./bareai` shows correlated infrastructure at a glance with trends; deep tabs searchable over SSH.

---

## Phase 7 — Cross-platform hardening

**Goal:** First-class Mac / Windows / Linux behavior.

- [x] Document per-OS capability matrix (what works where)
- [x] Windows: Docker Desktop pipe, NVIDIA where present
- [x] macOS: host + Docker + Ollama path; honest Apple GPU limits
- [x] Linux: primary AIOps path (full collectors)
- [x] Integration smoke tests / manual checklist per OS

**Exit:** Capability matrix in docs/platforms.md; no surprise crashes on unsupported features.

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

- [x] `bareai doctor` — ranked findings with “what/why/try” (read-only suggestions)
- [x] Richer LLM metrics when exporters exist (KV cache, queue, tok/s)
- [x] Config file (`~/.config/bareai/config.yaml`) for default probe prompts, ports, refresh
- [x] Man pages / better `--help` examples
- [x] Performance pass: low overhead on refresh loops

**Exit:** Doctor + config make daily AIOps inspection feel complete for one box.

---

## Phase 10 — Cockpit love (always-open pane)

**Goal:** Make `bareai` / `bareai watch` muscle memory — the correlated AI-box cockpit people leave open like `btop`, not a utility they open only when something breaks.

- [x] Finish local DB discovery as a first-class pane of the same story (Postgres, Redis, MongoDB, MySQL, Qdrant, Elasticsearch, …): docs, doctor rules, CLI/TUI/JSON, and correlation into the main join (not a silo)
- [x] Correlation as theater: Overview answers *which model → which container → which GPU → VRAM → health* in one glance; join graph is the product
- [x] Residency UX for `watch`: cheaper refresh paths, livelier sparklines/trends, fuzzy dive that feels inevitable over SSH/tmux
- [x] Empty-box magic: delightful first run when Docker/GPU/LLM/DB are absent (“nothing running yet — here’s what I’d look for”); never a dead dashboard
- [x] Apple Silicon + AMD GPU dignity: richer identity/metrics where the platform allows; honest limits elsewhere (Mac persona must not feel second-class)
- [x] Shareable doctor: paste-friendly report (`doctor --share` / gist-ready text or JSON) for Discord, GitHub issues, Slack
- [x] Protect identity: every new collector must join the correlation graph or stay out

**Exit:** First 10 seconds feel unfairly good; people leave `watch` open in a pane; pasteable doctor becomes how the community debugs a box.

---

## Phase 11 — Agent eyes (dependency)

**Goal:** Coding agents and scripts treat bareai as ground truth for the machine — not `nvidia-smi` + `docker ps` + curl. `--json` got us to the door; a first-class agent contract walks through it.

- [ ] MCP server exposing stable tools: list endpoints/models, inspect correlation, probe latency/health, surface doctor findings
- [ ] Documented agent contract on top of the shared `Snapshot` schema (versioned, predictable field names, skip reasons)
- [ ] Agent-oriented examples: Cursor / Claude / scripts calling bareai instead of ad-hoc shell soup
- [ ] Keep CLI/TUI/MCP on one brain — no parallel discovery implementations
- [ ] Optional light Prometheus export of bareai’s own snapshot signals (for stacks that already scrape)

**Exit:** An agent can answer “what’s on this box and is inference healthy?” via bareai alone; uninstall starts to hurt.

---

## Phase 12 — Careful power (close the loop)

**Goal:** After diagnosis, close the loop with a tiny set of confirm-gated actions — a reward after trust, not a mutate platform.

- [ ] Explicit confirmations (and dry-run where useful) for ~5 daily actions, e.g.:
  - restart / stop a discovered container or endpoint tenant
  - free a stuck GPU tenant (documented, safe path)
  - re-probe after a change
  - open / tail relevant logs (read path first; mutate only where clearly scoped)
- [ ] Every mutate surfaces what will change and why (doctor-style what/why/try → do)
- [ ] Audit-friendly output (`--json` result of the action); never silent side effects
- [ ] Stay out of general Docker/systemd wrappers — only actions that follow a bareai finding

**Exit:** Diagnose → confirm → fix for the common hung-box cases without leaving bareai; still not a fleet orchestrator.

---

## Phase 13 — Reach (multi-host, only when ready)

**Goal:** Laptop → GPU box without diluting the single-pane identity. Ship only after Phases 10–12 feel religious on one machine.

- [ ] SSH / remote snapshot of one other host (same collectors, same Snapshot)
- [ ] UX that still feels like one cockpit (context switch / host picker), not a fleet console
- [ ] Auth and timeout behavior that degrade cleanly when the remote is down

**Exit:** Solo engineer can inspect laptop + one AI box with the same mental model.

---

## Later (backlog — not scheduled)

- Kubernetes / pod awareness (secondary to bare metal)
- Broader fleet / multi-box beyond one remote
- Official distro packages (Debian/Ubuntu archives)
- Scoop formula (winget already covered)
- Deeper vendor GPU parity beyond Phase 10 dignity pass

**Do not chase:** becoming kubectl, Datadog, or a generic sysadmin suite. Stay the cockpit for an AI box.

---

## Name notes

**Chosen:** `bareai` — short, signals bare-metal + AI.

**Alternatives** (if you want to bike-shed later): `barectl`, `metalai`, `boxai`, `aibox`, `inferbox`, `gpuctl`.

Stick with `bareai` unless packaging/trademark conflict appears.
