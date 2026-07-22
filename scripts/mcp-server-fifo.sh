#!/usr/bin/env bash
# Run bareai mcp on named pipes so another terminal can attach with mcp-smoke.sh.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BIN="${BAREAI_BIN:-$ROOT/bareai}"
DIR="${1:-/tmp/bareai-mcp}"

if [[ ! -x "$BIN" ]]; then
  echo "building bareai..."
  (cd "$ROOT" && go build -o bareai ./cmd/bareai)
fi

mkdir -p "$DIR"
if [[ ! -p "$DIR/in" ]]; then
  mkfifo "$DIR/in"
fi
if [[ ! -p "$DIR/out" ]]; then
  mkfifo "$DIR/out"
fi

echo "FIFO MCP server on $DIR"
echo "In another terminal: ./scripts/mcp-smoke.sh --attach $DIR"
echo "Activity logs appear HERE when the client connects."
echo

exec "$BIN" mcp --fifo-dir "$DIR"
