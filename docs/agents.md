# Agent integration (MCP)

Coding agents should use **bareai** as ground truth for a bare-metal AI box instead of ad-hoc `nvidia-smi`, `docker ps`, and curl probes.

## When to use bareai

| Question | Use |
|----------|-----|
| What is on this box? | `bareai_snapshot` or `bareai inspect --json` |
| Which model runs on which GPU? | `bareai_correlations` |
| What's wrong / what to try? | `bareai_doctor` |
| Is inference healthy? | `bareai_probe` |
| LLM or DB details only | `bareai_llms` / `bareai_databases` |

## MCP setup

Start the server (stdio — for agent subprocess wiring):

```bash
bareai mcp
```

When run in a terminal, **stderr** shows activity so you know the server is alive:

```
bareai mcp: stdio server ready (version 0.x.y, 6 tools)
bareai mcp: stdout=MCP protocol · stderr=logs · waiting for client on stdin…
bareai mcp: ← initialize client="cursor"
bareai mcp: → initialize ok (1ms)
bareai mcp: ← tools/call tool="bareai_snapshot"
bareai mcp: → tools/call ok (420ms)
```

Stdout stays reserved for MCP JSON-RPC only.

### Local smoke test (no Cursor)

**Default** — self-contained (logs in the smoke terminal):

```bash
./scripts/mcp-smoke.sh
```

This starts its **own** `bareai mcp` subprocess. It does **not** talk to a separate `./bareai mcp` you already have running.

**Two-terminal demo** — activity logs on the server terminal:

```bash
# terminal 1
./scripts/mcp-server-fifo.sh

# terminal 2
./scripts/mcp-smoke.sh --attach /tmp/bareai-mcp
```

Terminal 1 shows `← initialize`, `← tools/call`, etc. when terminal 2 runs the smoke test.

### Cursor

Copy [`examples/cursor-mcp.json`](../examples/cursor-mcp.json) into your Cursor MCP settings, or:

```json
{
  "mcpServers": {
    "bareai": {
      "command": "bareai",
      "args": ["mcp"]
    }
  }
}
```

### Claude Desktop

Merge [`examples/claude-desktop-config.json`](../examples/claude-desktop-config.json) into your Claude Desktop MCP config (same shape as Cursor).

## MCP tools

All tools return JSON text with a versioned envelope:

```json
{
  "schema_version": "1.0",
  "bareai_version": "0.x.y",
  "collected_at": "2026-07-22T12:00:00Z",
  "data": {}
}
```

| Tool | Input | Output `data` |
|------|-------|---------------|
| `bareai_snapshot` | `{ "light": bool, "timeout_seconds": int }` | Full enriched `Snapshot` |
| `bareai_correlations` | `{ "timeout_seconds": int }` | `{ "correlations": [...], "skipped": [...] }` |
| `bareai_llms` | `{ "list_models": bool, "timeout_seconds": int }` | `{ "llms": [...], "skipped": [...] }` |
| `bareai_databases` | `{ "timeout_seconds": int }` | `{ "databases": [...], "skipped": [...] }` |
| `bareai_doctor` | `{ "min_severity": "info\|warn\|critical", "timeout_seconds": int }` | `{ "findings": [...], "counts": {...}, "skipped": [...] }` |
| `bareai_probe` | `{ "endpoint", "runtime", "model", "prompt", "timeout_seconds" }` | `{ "llms": [...], "skipped": [...] }` |

Optional `timeout_seconds` is capped at **120**; default comes from `~/.config/bareai/config.yaml`.

## Shell fallback (no MCP)

See [`examples/agent-shell.sh`](../examples/agent-shell.sh):

```bash
bareai inspect --json
bareai doctor --json
bareai probe --json
```

`bareai inspect --json` includes additive `schema_version` alongside the snapshot fields.

## Schema stability

**Stable (do not rename without bumping `schema_version`):**

- Snapshot top-level fields (`host`, `gpus`, `docker`, `llms`, `databases`, `correlations`, `findings`, `skipped`)
- `correlations[].kind` (`llm` | `db`)
- Finding fields: `id`, `severity`, `summary`, `why`, `try`

**May grow:** new optional JSON fields only.

**Schema changelog**

| Version | Notes |
|---------|-------|
| `1.0` | Initial agent envelope; MCP tools; `correlations[].kind`; databases pane |

## Agent prompt snippet

> Before suggesting GPU, Docker, or inference changes, call `bareai_snapshot` or `bareai_doctor` on this machine. Prefer bareai MCP tools over raw shell diagnostics.
