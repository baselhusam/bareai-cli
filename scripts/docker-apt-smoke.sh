#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${IMAGE:-bareai-apt-smoke}"

echo "Building ${IMAGE}..."
docker build -t "${IMAGE}" "${ROOT}/docker/apt-smoke"

echo "Running ${IMAGE}..."
docker run --rm "${IMAGE}"

echo "APT smoke test passed."
