# Interactive TUI

The TUI is the default experience on a TTY. It reuses the same collectors and probe logic as the CLI — it is a **view** over `snapshot.Snapshot`, not a second implementation.

```bash
bareai              # default on TTY
bareai watch        # explicit, with --refresh
bareai watch --refresh 5s --no-color
```

## Layout

| Tab | Key | Content |
|-----|-----|---------|
| **Overview** | `1` | Dense live dashboard (see below) |
| **GPUs** | `2` | Searchable GPU list + detail pane |
| **LLMs** | `3` | Providers/runtimes, PIDs, health, models |
| **Docker** | `4` | Container list + detail (running by default) |
| **DBs** | `5` | Database list + detail (Postgres, Redis, …) |
| **Probe** | `6` | Smoke-probe selected LLM |

## Overview dashboard

The Overview tab shows everything at a glance:

- **Host** — CPU load, RAM, primary disk with colored Unicode bars and sparklines
- **GPUs** — per-device util/VRAM bars, sparklines, temperature coloring (Apple shows unified memory honestly)
- **Providers / LLMs** — runtime, endpoint, PID, GPU#, health, latency sparkline
- **Databases** — engine, address, version, health
- **Correlation** — theater rows: model/engine → container → GPU/VRAM or address → health
- **Findings** — top diagnostics
- **Skipped** — collectors that failed (muted)

When nothing is running yet, panels show actionable hints (start Ollama, Docker Desktop, local DBs on default ports, etc.) instead of dead empty states.

Sparklines track the last **~40 refresh samples** (load, RAM, GPU util/VRAM, LLM health latency).

Bar/sparkline color thresholds: **70% / 90%** util/mem; temperature warn **≥75°C**, fail **≥85°C**.

## Keyboard reference

| Key | Overview | List tabs (GPU / LLM / Docker / DBs) | Detail pane |
|-----|----------|--------------------------------------|-------------|
| `1`–`6`, `Tab`, `Shift+Tab` | Switch tab | Switch tab | Switch tab |
| `↑` / `↓`, `j` / `k` | Move section / row | Move selection | Scroll |
| `Enter`, `l` | Dive into matching tab with selection synced | Focus detail pane | — |
| `/` | Jump to LLMs tab + open filter | Fuzzy filter list | — |
| `Esc` | — | Clear filter / back to list | Back to list |
| `a` | — | Docker: toggle all vs running containers | — |
| `r` | Force full refresh | Force full refresh | Force full refresh |
| `p` | — | Probe selected LLM (LLMs tab) | — |
| `q`, `Ctrl+C` | Quit | Quit | Quit |
| `?` | Toggle context-sensitive help | Toggle help | Toggle help |

**Overview dive:** Press `Enter` on a GPU, LLM, DB, or correlation row to jump to the matching tab with that item selected. Correlation rows dive to DBs, GPU (when indexed), Docker (when container matches), or LLMs.

**Refresh behavior:** Background ticks use a **light** snapshot (skips model listing, Docker images/volumes, DB version probes) to stay fast; `r` forces a full collect.
