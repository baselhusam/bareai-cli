#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BIN="$ROOT/bareai"

if [[ ! -x "$BIN" ]]; then
  echo "building bareai..."
  (cd "$ROOT" && go build -o bareai ./cmd/bareai)
fi

run() {
  echo "+ bareai $*"
  NO_COLOR=1 "$BIN" "$@"
}

run status --json >/dev/null
run gpu --json >/dev/null
run docker --json >/dev/null
run llm --json >/dev/null
run inspect --json >/dev/null
run probe --endpoint http://127.0.0.1:59999 --runtime ollama --json >/dev/null

echo "smoke: all commands exited 0"
