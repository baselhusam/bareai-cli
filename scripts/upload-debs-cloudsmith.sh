#!/usr/bin/env bash
# Upload .deb packages from a GitHub Release to Cloudsmith APT.
# Maintainer helper when CI upload did not run.
set -euo pipefail

REPO="${REPO:-baselhusam/bareai-cli}"
CLOUDSMITH_REPO="${CLOUDSMITH_REPO:-baselhusam/bareai}"
VERSION="${VERSION:-v0.1.0}"

if [[ -z "${CLOUDSMITH_API_KEY:-}" ]]; then
  echo "Set CLOUDSMITH_API_KEY to your Cloudsmith API key (csa_...)." >&2
  exit 1
fi

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Required command not found: $1" >&2
    exit 1
  }
}

need_cmd curl
need_cmd tar

if ! command -v cloudsmith >/dev/null 2>&1; then
  need_cmd pip3
  pip3 install --user cloudsmith-cli
  export PATH="${HOME}/.local/bin:${PATH}"
fi

ver="${VERSION#v}"
tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

base="https://github.com/${REPO}/releases/download/${VERSION}"
for arch in amd64 arm64; do
  file="bareai_${ver}_linux_${arch}.deb"
  echo "Downloading ${file}..."
  curl -fsSL "${base}/${file}" -o "${tmpdir}/${file}"
done

for file in "${tmpdir}"/*.deb; do
  echo "Uploading $(basename "$file") to ${CLOUDSMITH_REPO}..."
  cloudsmith push deb "${CLOUDSMITH_REPO}/any-distro/any-version" "$file" \
    --republish --no-wait-for-sync
done

echo "Done. Check packages at: https://cloudsmith.io/~baselhusam/repos/bareai/packages/"
