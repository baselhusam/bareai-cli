# Platform capability matrix

**Primary path:** Linux bare-metal is the reference environment with full collector fidelity.

| Feature | Linux | macOS | Windows |
|---------|-------|-------|---------|
| **Host** (`status`) | Full: CPU, RAM, disk, load | Full; load may be 0 on some configs | Full; load averages show as n/a |
| **GPU** | NVIDIA via `nvidia-smi`; AMD via sysfs/ROCm | Apple Silicon chip name; NVIDIA if driver installed | NVIDIA if `nvidia-smi` installed |
| **Docker** | Full via unix socket | Full via Docker Desktop socket | Full via named pipe (`docker_engine`) |
| **LLM discovery** | Docker + process + port scan | Docker Desktop; native Ollama on `:11434` | Docker Desktop; process `.exe` names; port scan |
| **Probe / inspect / TUI** | Full | Full | Full; TUI via Windows Terminal |

## GPU detail by platform

| Platform | NVIDIA | AMD | Apple |
|----------|--------|-----|-------|
| Linux | NVIDIA via `nvidia-smi`; AMD via sysfs + `rocm-smi` JSON when available | sysfs VRAM + temp; ROCm util/name when present | n/a |
| macOS | If `nvidia-smi` installed | n/a | Chip name + unified memory pool; richer identity via system_profiler; no util/temp/power |
| Windows | If `nvidia-smi` installed | n/a | n/a |

## Platform notes

- **Linux:** NVIDIA GPU↔process correlation via `nvidia-smi`; AMD metrics from `/sys/class/drm`
- **macOS:** Ollama commonly via Docker Desktop or native app; process scan may need permissions
- **Windows:** Docker Desktop required for containers; WSL2 engine not visible unless `DOCKER_HOST` is set; load averages unavailable

When collectors are unavailable, commands exit `0` with clear skip messages. Manual verification checklist: [CHECKLIST.md](CHECKLIST.md).
