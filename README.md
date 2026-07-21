# bareai

CLI and TUI for solo AI engineers inspecting bare-metal AI infrastructure: host resources, GPUs (NVIDIA / AMD / Apple), Docker, and local LLM runtimes (Ollama, vLLM, SGLang, Triton, …).

**Status:** Phase 0 — project skeleton. Commands are stubbed; see [ROADMAP.md](ROADMAP.md) for the full plan.

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
./bareai watch
```

### Commands

| Command   | Description                                      | Phase |
|-----------|--------------------------------------------------|-------|
| `status`  | Host and infrastructure summary                  | 1     |
| `gpu`     | GPU and accelerator details                      | 2     |
| `docker`  | Docker containers, images, and volumes           | 3     |
| `llm`     | Discovered LLM runtimes and models               | 4     |
| `probe`   | One-hit smoke tests against discovered LLMs      | 4     |
| `inspect` | Full correlated infrastructure report            | 5     |
| `watch`   | Live TUI monitoring dashboard                    | 6     |

All commands are currently stubs that print a not-implemented message.

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
make run ARGS="--help"
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
