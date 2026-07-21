# bareai

CLI and TUI for solo AI engineers inspecting bare-metal AI infrastructure: host resources, GPUs (NVIDIA / AMD / Apple), Docker, and local LLM runtimes (Ollama, vLLM, SGLang, Triton, …).

**Status:** Phase 3 complete — `bareai status`, `bareai gpu`, and `bareai docker` report host, accelerator, and Docker inventory. LLM commands are still stubbed. See [ROADMAP.md](ROADMAP.md).

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

### GPU capability matrix

| Platform | NVIDIA | AMD | Apple |
|----------|--------|-----|-------|
| Linux | Full via `nvidia-smi` (util, VRAM, temp, power, processes) | ROCm / sysfs VRAM + temp when present | n/a |
| macOS | If `nvidia-smi` installed | n/a | Best-effort chip name; unified memory (no discrete VRAM) |
| Windows | If `nvidia-smi` installed | n/a | n/a |

When no accelerators are found, commands exit `0` with a clear message.

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

### Commands

| Command   | Description                                      | Status   |
|-----------|--------------------------------------------------|----------|
| `status`  | Host and infrastructure summary                  | Phase 1–3 |
| `gpu`     | GPU and accelerator details                      | Phase 2  |
| `docker`  | Docker containers, images, and volumes           | Phase 3  |
| `llm`     | Discovered LLM runtimes and models               | stub     |
| `probe`   | One-hit smoke tests against discovered LLMs      | stub     |
| `inspect` | Full correlated infrastructure report            | stub     |
| `watch`   | Live TUI monitoring dashboard                    | stub     |

### Global flags

| Flag          | Short | Default | Description                          |
|---------------|-------|---------|--------------------------------------|
| `--json`      | `-j`  | false   | Output in JSON format                |
| `--timeout`   |       | `10s`   | Timeout for probes and API calls     |
| `--no-color`  |       | false   | Disable colored output               |

## Development

```bash
make test
make lint    # requires golangci-lint
make run ARGS="gpu"
make clean
```

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
