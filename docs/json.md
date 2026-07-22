# JSON output and snapshot model

Every inspect command supports `--json` (`-j`). Output is a single **`Snapshot`** object:

```json
{
  "collected_at": "2026-07-22T12:00:00Z",
  "host": { "hostname": "...", "mem_total_bytes": 0, "load1": 0.0 },
  "gpus": [{ "index": 0, "vendor": "nvidia", "utilization_pct": 45.0, "processes": [] }],
  "docker": { "available": true, "containers": [], "images": [], "volumes": [] },
  "llms": [{ "runtime": "ollama", "endpoint": "...", "health": { "ok": true, "latency_ms": 42 } }],
  "databases": [{ "engine": "redis", "address": "127.0.0.1:6379", "health": { "ok": true } }],
  "correlations": [{ "kind": "llm", "endpoint": "...", "models": ["llama3.2"], "gpu_index": 0, "health_ok": true }],
  "findings": [{ "id": "...", "severity": "info", "summary": "...", "why": "...", "try": "..." }],
  "skipped": [{ "component": "docker", "reason": "..." }]
}
```

## Top-level fields

| Field | Present in | Description |
|-------|------------|-------------|
| `collected_at` | All commands | UTC timestamp of collection |
| `host` | status, inspect, doctor, … | Host inventory |
| `gpus` | gpu, status, inspect, … | Accelerator list |
| `docker` | docker, status, inspect, … | Engine inventory |
| `llms` | llm, probe, inspect, … | Discovered inference servers |
| `databases` | db, status, inspect, … | Local database instances |
| `correlations` | inspect, doctor (after enrich) | Joined LLM/DB ↔ resource rows (`kind`: `llm` or `db`) |
| `findings` | inspect, doctor | Diagnostic findings |
| `skipped` | All | Collectors that failed with reason |

## Scripting examples

```bash
# Fail CI if any LLM health check failed
bareai llm --json | jq -e '.llms | all(.health.ok == true)'

# Export GPU util for monitoring
bareai gpu --json | jq '.gpus[] | {index, util: .utilization_pct, mem: .memory_used_bytes}'
```

Types are defined in [`internal/snapshot/snapshot.go`](../internal/snapshot/snapshot.go).
