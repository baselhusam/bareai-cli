# Common workflows

## Use bareai from an agent (MCP)

```bash
# Add to Cursor MCP settings (see examples/cursor-mcp.json)
bareai mcp

# Or shell fallback without MCP
bareai inspect --json | jq '.correlations'
bareai doctor --json | jq '.findings[]'
```

Full guide: [agents.md](agents.md)

## Diagnose → fix (confirm-gated)

```bash
bareai doctor
bareai do list
bareai do plan restart --finding llm.unreachable --container ollama
bareai do restart --finding llm.unreachable --container ollama --yes
bareai do reprobe --finding llm.unreachable --endpoint http://127.0.0.1:11434
```

Full guide: [actions.md](actions.md)

## Paste a doctor report into an issue

```bash
bareai doctor --share > box-report.txt
# paste into GitHub issue, Discord, or a gist
```

## First run on an empty box

```bash
bareai              # Overview shows hints when nothing is running yet
bareai doctor       # includes host.empty_box info finding when idle
```

## SSH into an AI box and see everything

```bash
ssh ai-box
bareai                    # live dashboard
# or
bareai inspect            # one-shot full report
```

## Debug an unreachable LLM

```bash
bareai llm
bareai probe --endpoint http://127.0.0.1:11434 --runtime ollama
bareai doctor --severity warn
```

## Check GPU usage before starting a job

```bash
bareai gpu
bareai gpu --json | jq '.gpus[] | select(.utilization_pct > 80)'
```

## Scriptable health check

```bash
#!/usr/bin/env bash
set -euo pipefail
bareai probe --json | jq -e '.llms | length > 0 and all(.probe.ok // .health.ok)'
echo "All LLM probes passed"
```

## Monitor in TUI with slower refresh

```bash
bareai watch --refresh 10s
```
