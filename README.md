# bareai

CLI and TUI for solo AI engineers inspecting bare-metal AI infrastructure: host resources, GPUs (NVIDIA / AMD / Apple), Docker, and local LLM runtimes (Ollama, vLLM, SGLang, Triton, …).

**Status:** Phase 1 complete — `bareai status` reports host inventory (CPU, RAM, disk, uptime). Other commands are still stubbed. See [ROADMAP.md](ROADMAP.md).

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
./bareai gpu          # still a stub
```

### Example

```text
$ bareai status
bareai status
Collected: 2026-07-22T12:00:00Z

Host
  Hostname:  ai-box
  OS:        linux 6.8.0 (linux)
  Arch:      amd64
  Uptime:    5d 3h 12m
  CPU:       AMD EPYC 7763 (64 cores, 128 logical)
  Load:      2.10 / 1.85 / 1.60 (1/5/15 min)
  Memory:    96.0 GiB / 256.0 GiB (38% used, 160.0 GiB available)
  Disks:
    / (ext4)                 420.0 GiB / 1.0 TiB (42% used)

GPUs:        not collected yet (Phase 2)
Docker:      not collected yet (Phase 3)
LLM runtimes: not collected yet (Phase 4)
```

Use `--json` for machine-readable output (scripts, agents, CI).

### Commands

| Command   | Description                                      | Status   |
|-----------|--------------------------------------------------|----------|
| `status`  | Host and infrastructure summary                  | Phase 1  |
| `gpu`     | GPU and accelerator details                      | stub     |
| `docker`  | Docker containers, images, and volumes           | stub     |
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
make run ARGS="status"
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
