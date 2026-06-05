#!/usr/bin/env bash
# Install sceno from GitHub releases.
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/niklas-heer/sceno/main/scripts/install.sh | bash
#   curl -fsSL ... | bash -s -- --version v0.2.0   # optional pin
#   curl -fsSL ... | bash -s -- --dir ~/.local/bin
set -euo pipefail

REPO="niklas-heer/sceno"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
VERSION=""
VERIFY=1

usage() {
  cat <<EOF
Usage: install.sh [options]

Install sceno from GitHub releases.

Options:
  --version VER   Pin a specific release tag (default: latest published release)
  --dir PATH      Install directory (default: /usr/local/bin)
  --no-verify     Skip SHA256 checksum verification
  -h, --help      Show this help

Environment:
  INSTALL_DIR     Same as --dir
  SCENO_VERSION   Same as --version (without v prefix ok)

Examples:
  curl -fsSL https://raw.githubusercontent.com/${REPO}/main/scripts/install.sh | bash
  curl -fsSL ... | bash -s -- --dir ~/.local/bin
  curl -fsSL ... | bash -s -- --version v0.2.0
EOF
}

fetch_latest_version() {
  local json tag
  json="$(curl -fsSL -H "Accept: application/vnd.github+json" \
    "https://api.github.com/repos/${REPO}/releases/latest")"
  if command -v python3 >/dev/null 2>&1; then
    tag="$(python3 -c 'import json,sys; print(json.load(sys.stdin)["tag_name"])' <<< "$json")"
  elif command -v jq >/dev/null 2>&1; then
    tag="$(jq -r .tag_name <<< "$json")"
  else
    tag="$(printf '%s' "$json" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\(v[^"]*\)".*/\1/p' | head -1)"
  fi
  tag="${tag#v}"
  if [[ -z "$tag" || "$tag" == "null" ]]; then
    echo "Could not determine latest release version." >&2
    exit 1
  fi
  echo "$tag"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      VERSION="${2#v}"
      shift 2
      ;;
    --dir)
      INSTALL_DIR="$2"
      shift 2
      ;;
    --no-verify)
      VERIFY=0
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if [[ -n "${SCENO_VERSION:-}" && -z "$VERSION" ]]; then
  VERSION="${SCENO_VERSION#v}"
fi

case "$(uname -s)" in
  Darwin)  OS=darwin ;;
  Linux)   OS=linux ;;
  *)
    echo "Unsupported OS: $(uname -s). Use go install github.com/niklas-heer/sceno/cmd/sceno@latest" >&2
    exit 1
    ;;
esac

case "$(uname -m)" in
  x86_64|amd64)  ARCH=amd64 ;;
  arm64|aarch64) ARCH=arm64 ;;
  *)
    echo "Unsupported architecture: $(uname -m)" >&2
    exit 1
    ;;
esac

ARCHIVE="sceno_${OS}_${ARCH}.tar.gz"
BASE_URL="https://github.com/${REPO}/releases/download"

if [[ -z "$VERSION" ]]; then
  echo "Fetching latest ${REPO} release..."
  VERSION="$(fetch_latest_version)"
  echo "Latest release: v${VERSION}"
fi

TAG="v${VERSION}"
URL="${BASE_URL}/${TAG}/${ARCHIVE}"
CHECKSUMS_URL="${BASE_URL}/${TAG}/SHA256SUMS"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "Downloading sceno ${TAG} for ${OS}/${ARCH}..."
curl -fsSL "$URL" -o "${TMP}/${ARCHIVE}"

if [[ "$VERIFY" -eq 1 ]]; then
  echo "Verifying checksum..."
  curl -fsSL "$CHECKSUMS_URL" -o "${TMP}/SHA256SUMS"
  EXPECTED="$(grep " ${ARCHIVE}$" "${TMP}/SHA256SUMS" | awk '{print $1}')"
  if [[ -z "$EXPECTED" ]]; then
    echo "Checksum for ${ARCHIVE} not found in SHA256SUMS" >&2
    exit 1
  fi
  if command -v shasum >/dev/null 2>&1; then
    ACTUAL="$(shasum -a 256 "${TMP}/${ARCHIVE}" | awk '{print $1}')"
  else
    ACTUAL="$(sha256sum "${TMP}/${ARCHIVE}" | awk '{print $1}')"
  fi
  if [[ "$ACTUAL" != "$EXPECTED" ]]; then
    echo "Checksum mismatch for ${ARCHIVE}" >&2
    exit 1
  fi
fi

tar -xzf "${TMP}/${ARCHIVE}" -C "$TMP"

DEFAULT_INSTALL_DIR="/usr/local/bin"
install_dir_writable() {
  local dir="$1"
  mkdir -p "$dir" 2>/dev/null || return 1
  [[ -w "$dir" ]]
}

if [[ "$INSTALL_DIR" == "$DEFAULT_INSTALL_DIR" ]] && ! install_dir_writable "$INSTALL_DIR"; then
  FALLBACK="${HOME}/.local/bin"
  echo "Cannot write to ${INSTALL_DIR} (permission denied)."
  echo "Installing to ${FALLBACK} instead."
  echo "Re-run with --dir PATH or sudo if you need ${DEFAULT_INSTALL_DIR}."
  INSTALL_DIR="$FALLBACK"
fi

if ! install_dir_writable "$INSTALL_DIR"; then
  echo "Cannot write to ${INSTALL_DIR} (permission denied)." >&2
  echo "Try: curl -fsSL ... | bash -s -- --dir ~/.local/bin" >&2
  exit 1
fi

install -m 755 "${TMP}/sceno" "${INSTALL_DIR}/sceno"

echo ""
echo "Installed sceno ${VERSION} to ${INSTALL_DIR}/sceno"
if ! command -v sceno >/dev/null 2>&1; then
  echo "Add ${INSTALL_DIR} to your PATH if needed."
fi
"${INSTALL_DIR}/sceno" version
