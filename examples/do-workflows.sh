#!/usr/bin/env bash
# Diagnose → plan → confirm → verify workflow for bareai do.
set -euo pipefail

echo "=== Doctor findings ==="
bareai doctor --severity warn

echo "=== Available actions ==="
bareai do list

echo "=== Example plan (dry-run) ==="
echo "bareai do plan restart --finding llm.unreachable --container ollama"
echo "bareai do restart --finding llm.unreachable --container ollama --yes"
echo "bareai do reprobe --finding llm.unreachable --endpoint http://127.0.0.1:11434"
