#!/usr/bin/env bash
set -euo pipefail

REPO="${REPO:-baselhusam/bareai-cli}"
INSTALL_DIR="${INSTALL_DIR:-}"
VERSION="${VERSION:-}"
INSTALL_SYSTEM="${INSTALL_SYSTEM:-}"

usage() {
  cat <<EOF
Usage: install.sh [options]

Install bareai from GitHub Releases. No sudo required by default.

Options:
  --dir PATH       Install directory (default: ~/.local/bin)
  --system         Install to /usr/local/bin (may prompt for sudo)
  --version TAG    Release tag (default: latest)
  -h, --help       Show this help

Examples:
  curl -fsSL .../install.sh | bash
  curl -fsSL .../install.sh | bash -s -- --version v0.1.0
EOF
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "install.sh: required command not found: $1" >&2
    exit 1
  }
}

parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --dir)
        INSTALL_DIR="$2"
        shift 2
        ;;
      --system)
        INSTALL_SYSTEM=1
        shift
        ;;
      --version)
        VERSION="$2"
        shift 2
        ;;
      -h | --help)
        usage
        exit 0
        ;;
      *)
        echo "install.sh: unknown argument: $1" >&2
        usage >&2
        exit 1
        ;;
    esac
  done
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
    x86_64 | amd64) arch="amd64" ;;
    aarch64 | arm64) arch="arm64" ;;
    *)
      echo "install.sh: unsupported architecture: $arch" >&2
      exit 1
      ;;
  esac

  echo "${os} ${arch}"
}

default_install_dir() {
  if [[ "$INSTALL_SYSTEM" == "1" ]]; then
    echo "/usr/local/bin"
    return
  fi
  echo "${HOME}/.local/bin"
}

shell_rc() {
  case "$(basename "${SHELL:-}")" in
    zsh) echo "${HOME}/.zshrc" ;;
    bash) echo "${HOME}/.bashrc" ;;
    *) echo "${HOME}/.profile" ;;
  esac
}

ensure_path() {
  local dir="$1"
  if [[ ":$PATH:" == *":${dir}:"* ]]; then
    return 0
  fi

  export PATH="${dir}:${PATH}"

  local rc line marker updated=0
  rc="$(shell_rc)"
  line="export PATH=\"${dir}:\$PATH\""
  marker="# bareai"

  if [[ ! -f "$rc" ]]; then
    touch "$rc"
  fi

  if ! grep -qF "$dir" "$rc" 2>/dev/null; then
    {
      echo ""
      echo "$marker"
      echo "$line"
    } >>"$rc"
    echo "Added ${dir} to PATH in ${rc}"
    updated=1
  fi

  if [[ "$updated" -eq 1 ]]; then
    echo "Restart your shell, or run: source ${rc}"
  fi
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
  parse_args "$@"

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

  if [[ -z "$INSTALL_DIR" ]]; then
    INSTALL_DIR="$(default_install_dir)"
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

  mkdir -p "$INSTALL_DIR"
  if [[ -w "$INSTALL_DIR" ]]; then
    install -m 0755 "${tmpdir}/bareai" "${INSTALL_DIR}/bareai"
  elif [[ "$INSTALL_SYSTEM" == "1" ]]; then
    need_cmd sudo
    sudo install -m 0755 "${tmpdir}/bareai" "${INSTALL_DIR}/bareai"
  else
    echo "install.sh: cannot write to ${INSTALL_DIR}" >&2
    echo "install.sh: pick another directory with --dir PATH" >&2
    exit 1
  fi

  ensure_path "$INSTALL_DIR"

  echo "Installed to ${INSTALL_DIR}/bareai"
  "${INSTALL_DIR}/bareai" version
}

main "$@"
