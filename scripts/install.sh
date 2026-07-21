#!/usr/bin/env bash
set -euo pipefail

REPO="${REPO:-baselhusam/bareai-cli}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
VERSION="${VERSION:-}"

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "install.sh: required command not found: $1" >&2
    exit 1
  }
}

detect_platform() {
  local os arch
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "$os" in
    linux) os="linux" ;;
    darwin) os="darwin" ;;
    *)
      echo "install.sh: unsupported OS: $os" >&2
      exit 1
      ;;
  esac

  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64) arch="amd64" ;;
    aarch64|arm64) arch="arm64" ;;
    *)
      echo "install.sh: unsupported architecture: $arch" >&2
      exit 1
      ;;
  esac

  echo "${os} ${arch}"
}

latest_version() {
  curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' \
    | head -n1 \
    | cut -d '"' -f4
}

verify_checksum() {
  local file="$1"
  local checksums="$2"
  local expected actual
  expected="$(grep "  $(basename "$file")$" "$checksums" | awk '{print $1}')"
  if [[ -z "$expected" ]]; then
    echo "install.sh: checksum entry not found for $(basename "$file")" >&2
    exit 1
  fi
  if command -v sha256sum >/dev/null 2>&1; then
    actual="$(sha256sum "$file" | awk '{print $1}')"
  else
    actual="$(shasum -a 256 "$file" | awk '{print $1}')"
  fi
  if [[ "$expected" != "$actual" ]]; then
    echo "install.sh: checksum mismatch" >&2
    exit 1
  fi
}

main() {
  need_cmd curl
  need_cmd tar
  need_cmd grep

  read -r os arch <<<"$(detect_platform)"

  if [[ -z "$VERSION" ]]; then
    VERSION="$(latest_version)"
  fi
  if [[ -z "$VERSION" ]]; then
    echo "install.sh: could not determine release version" >&2
    exit 1
  fi

  ver="${VERSION#v}"
  archive="bareai_${ver}_${os}_${arch}.tar.gz"
  base="https://github.com/${REPO}/releases/download/${VERSION}"
  tmpdir="$(mktemp -d)"
  trap 'rm -rf "$tmpdir"' EXIT

  echo "Installing bareai ${VERSION} for ${os}/${arch}..."
  curl -fsSL "${base}/checksums.txt" -o "${tmpdir}/checksums.txt"
  curl -fsSL "${base}/${archive}" -o "${tmpdir}/${archive}"
  verify_checksum "${tmpdir}/${archive}" "${tmpdir}/checksums.txt"

  tar -xzf "${tmpdir}/${archive}" -C "${tmpdir}" bareai
  if [[ ! -f "${tmpdir}/bareai" ]]; then
    echo "install.sh: bareai binary not found in archive" >&2
    exit 1
  fi

  if [[ -w "$INSTALL_DIR" ]]; then
    install -m 0755 "${tmpdir}/bareai" "${INSTALL_DIR}/bareai"
  else
    need_cmd sudo
    sudo install -m 0755 "${tmpdir}/bareai" "${INSTALL_DIR}/bareai"
  fi

  echo "Installed to ${INSTALL_DIR}/bareai"
  "${INSTALL_DIR}/bareai" version
}

main "$@"
