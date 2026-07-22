# Common workflows

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
