# Confirm-gated actions (`bareai do`)

Phase 12 closes the diagnose → confirm → fix loop with a **tiny** set of finding-scoped actions. Docker-only mutate in v1; MCP stays read-only.

## Safety model

1. Every **mutating** verb requires `--finding <id>` tied to a live doctor finding.
2. Targets must resolve from the current snapshot (container ID/name validated against discovery).
3. Default path: `bareai do plan <verb>` or `--dry-run` → review → `--yes` or interactive `Proceed? [y/N]`.
4. All verbs emit audit JSON with `--json` (`ActionResult` schema).

**Out of scope:** systemd/brew service control, generic `docker run/exec`, fleet orchestration, MCP mutate tools.

## Verbs

| Verb | Mutates? | Typical findings |
|------|----------|------------------|
| `logs` | No | `llm.unreachable`, `db.unreachable` |
| `reprobe` | No | `llm.unreachable`, `llm.no_models` |
| `restart` | Yes | `llm.unreachable`, `db.unreachable` |
| `stop` | Yes | `gpu.vram_high` |
| `free-gpu` | Yes | `gpu.vram_high`, `gpu.idle_while_llm` |

`free-gpu` stops the correlated inference container, waits, restarts, and optionally reprobes — it does **not** `kill -9` arbitrary PIDs.

## Commands

```bash
bareai do list
bareai do plan restart --finding llm.unreachable --container ollama
bareai do restart --finding llm.unreachable --container ollama --yes
bareai do logs --finding llm.unreachable --tail 200
bareai do reprobe --finding llm.no_models --endpoint http://127.0.0.1:11434
bareai do free-gpu --finding gpu.vram_high --container vllm --yes
```

Doctor human output includes `Do:` lines with suggested `bareai do plan …` commands when a docker target exists.

## ActionResult JSON

```json
{
  "schema_version": "1.0",
  "bareai_version": "0.x.y",
  "executed_at": "2026-07-22T12:00:00Z",
  "verb": "restart",
  "finding_id": "llm.unreachable",
  "target": { "kind": "container", "id": "abc", "name": "ollama" },
  "dry_run": false,
  "confirmed": true,
  "before": { "state": "running" },
  "after": { "state": "running" },
  "ok": true,
  "steps": []
}
```

Read-only `logs` responses include an `output` field (truncated by config `actions.log_max_bytes`).

## Config (`~/.config/bareai/config.yaml`)

```yaml
actions:
  confirm: true
  auto_reprobe: true
  log_tail: 100
  log_max_bytes: 262144
```

## Workflow

See [`examples/do-workflows.sh`](../examples/do-workflows.sh):

```bash
bareai doctor
bareai do list
bareai do plan restart --finding llm.unreachable --container ollama
bareai do restart --finding llm.unreachable --container ollama --yes
bareai do reprobe --finding llm.unreachable --endpoint http://127.0.0.1:11434
```
