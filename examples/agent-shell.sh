#!/usr/bin/env bash
# Shell fallback when MCP is unavailable — same brain, no subprocess soup.
set -euo pipefail

echo "=== Full snapshot ==="
bareai inspect --json | jq '{schema_version, host: .hostname, gpus: (.gpus|length), llms: (.llms|length), dbs: (.databases|length)}'

echo "=== Correlations ==="
bareai inspect --json | jq '.correlations[]'

echo "=== Doctor (warn+) ==="
bareai doctor --severity warn --json | jq '.findings[]'

echo "=== Probe all LLMs ==="
bareai probe --json | jq '.llms[] | {endpoint, probe: .probe.ok}'
