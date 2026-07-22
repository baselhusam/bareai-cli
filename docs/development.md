# Development

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

Cross-platform manual checklist: [CHECKLIST.md](CHECKLIST.md)

Release and packaging: [RELEASE.md](RELEASE.md)

```bash
make goreleaser-check
make release-snapshot
```

Install `golangci-lint`:

```bash
brew install golangci-lint
# or see https://golangci-lint.run/welcome/install/
```

Architecture overview: [`.cursor/rules/architecture.mdc`](../.cursor/rules/architecture.mdc) (collectors → snapshot → CLI/TUI/JSON).
