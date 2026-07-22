#!/usr/bin/env bash
# Smoke-test bareai mcp over stdio (no Cursor config required).
#
# Default: spawns its own short-lived bareai mcp subprocess (logs in THIS terminal).
# --attach DIR: talk to a server already running on FIFOs (logs in THAT terminal).
#
# Two-terminal demo:
#   terminal 1: ./scripts/mcp-server-fifo.sh
#   terminal 2: ./scripts/mcp-smoke.sh --attach /tmp/bareai-mcp
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BIN="${BAREAI_BIN:-$ROOT/bareai}"
ATTACH_DIR=""

usage() {
  cat <<EOF
Usage: $(basename "$0") [--attach DIR]

  (default)  Start embedded bareai mcp and run smoke checks (logs here)
  --attach   Connect to FIFO server at DIR/in and DIR/out (logs on server terminal)

Example two-terminal test:
  ./scripts/mcp-server-fifo.sh
  ./scripts/mcp-smoke.sh --attach /tmp/bareai-mcp
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --attach)
      ATTACH_DIR="${2:-}"
      if [[ -z "$ATTACH_DIR" ]]; then
        echo "mcp-smoke: --attach requires a directory" >&2
        exit 1
      fi
      shift 2
      ;;
    -h | --help)
      usage
      exit 0
      ;;
    *)
      echo "mcp-smoke: unknown argument: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if [[ ! -x "$BIN" ]]; then
  echo "building bareai..."
  (cd "$ROOT" && go build -o bareai ./cmd/bareai)
fi

if ! command -v python3 >/dev/null 2>&1; then
  echo "mcp-smoke: python3 is required" >&2
  exit 1
fi

if [[ -n "$ATTACH_DIR" ]]; then
  if [[ ! -p "$ATTACH_DIR/in" || ! -p "$ATTACH_DIR/out" ]]; then
    echo "mcp-smoke: expected FIFOs at $ATTACH_DIR/in and $ATTACH_DIR/out" >&2
    echo "Start the server first: ./scripts/mcp-server-fifo.sh $ATTACH_DIR" >&2
    exit 1
  fi
  echo "mcp-smoke: attaching to FIFO server at $ATTACH_DIR"
  echo "mcp-smoke: activity logs appear on the server terminal, not here"
else
  echo "mcp-smoke: using embedded server (activity logs appear below)"
  echo "mcp-smoke: to hit a server in another terminal, use --attach (see --help)"
fi

export BAREAI_BIN="$BIN"
export MCP_ATTACH_DIR="$ATTACH_DIR"

python3 <<'PY'
import json
import os
import select
import subprocess
import sys
import threading

attach_dir = os.environ.get("MCP_ATTACH_DIR", "")
bin_path = os.environ["BAREAI_BIN"]

proc = None
stdin_wr = None
stdout_rd = None
stderr_lines: list[str] = []


def drain_stderr() -> None:
    if proc is None or proc.stderr is None:
        return
    for line in proc.stderr:
        stderr_lines.append(line.rstrip("\n"))
        sys.stderr.write(line)


if attach_dir:
    in_path = os.path.join(attach_dir, "in")
    out_path = os.path.join(attach_dir, "out")
    in_fd = os.open(in_path, os.O_RDWR)
    out_fd = os.open(out_path, os.O_RDWR)
    stdin_wr = os.fdopen(in_fd, "w", buffering=1)
    stdout_rd = os.fdopen(out_fd, "r", buffering=1)
else:
    proc = subprocess.Popen(
        [bin_path, "mcp"],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        bufsize=1,
    )
    stdin_wr = proc.stdin
    stdout_rd = proc.stdout
    threading.Thread(target=drain_stderr, daemon=True).start()


def send(msg: dict) -> None:
    assert stdin_wr is not None
    stdin_wr.write(json.dumps(msg, separators=(",", ":")) + "\n")
    stdin_wr.flush()


def recv(timeout: float = 60) -> dict:
    assert stdout_rd is not None
    if not select.select([stdout_rd], [], [], timeout)[0]:
        raise TimeoutError(f"no MCP response within {timeout}s")
    return json.loads(stdout_rd.readline())


def call_tool(name: str, arguments: dict | None = None) -> dict:
    send(
        {
            "jsonrpc": "2.0",
            "id": name,
            "method": "tools/call",
            "params": {"name": name, "arguments": arguments or {}},
        }
    )
    reply = recv(90)
    if reply.get("error"):
        raise RuntimeError(f"{name}: {reply['error']}")
    content = reply.get("result", {}).get("content") or []
    if not content:
        raise RuntimeError(f"{name}: empty tool result")
    return json.loads(content[0]["text"])


results: list[str] = []

try:
    send(
        {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "initialize",
            "params": {
                "protocolVersion": "2024-11-05",
                "capabilities": {},
                "clientInfo": {"name": "mcp-smoke", "version": "1"},
            },
        }
    )
    init = recv()
    info = init.get("result", {}).get("serverInfo") or {}
    if info.get("name") != "bareai":
        raise RuntimeError(f"unexpected serverInfo: {info}")
    results.append(f"initialize: server={info}")

    send({"jsonrpc": "2.0", "method": "notifications/initialized"})

    send({"jsonrpc": "2.0", "id": 2, "method": "tools/list", "params": {}})
    tools = recv()
    names = sorted(t["name"] for t in tools["result"]["tools"])
    expected = sorted(
        [
            "bareai_snapshot",
            "bareai_correlations",
            "bareai_llms",
            "bareai_databases",
            "bareai_doctor",
            "bareai_probe",
        ]
    )
    if names != expected:
        raise RuntimeError(f"tools/list mismatch: got {names}, want {expected}")
    results.append(f"tools/list: {len(names)} tools OK")

    corr = call_tool("bareai_correlations", {"timeout_seconds": 15})
    for key in ("schema_version", "bareai_version", "collected_at", "data"):
        if key not in corr:
            raise RuntimeError(f"bareai_correlations missing {key}")
    data = corr.get("data") or {}
    corrs = data.get("correlations") or []
    skipped = data.get("skipped") or []
    results.append(
        f"bareai_correlations: schema={corr['schema_version']} "
        f"correlations={len(corrs)} skipped={len(skipped)}"
    )

    doc = call_tool("bareai_doctor", {"min_severity": "info", "timeout_seconds": 20})
    doc_data = doc.get("data") or {}
    findings = doc_data.get("findings") or []
    counts = doc_data.get("counts") or {}
    results.append(f"bareai_doctor: findings={len(findings)} counts={counts}")

finally:
    if proc is not None:
        proc.terminate()
        try:
            proc.wait(timeout=3)
        except subprocess.TimeoutExpired:
            proc.kill()
    else:
        if stdin_wr is not None:
            stdin_wr.close()
        if stdout_rd is not None:
            stdout_rd.close()

print("\n=== MCP smoke results ===")
for line in results:
    print("✓", line)

if stderr_lines:
    print("\n=== MCP server stderr (embedded mode) ===")
    for line in stderr_lines:
        print(line)

print("\nmcp-smoke: passed")
PY
