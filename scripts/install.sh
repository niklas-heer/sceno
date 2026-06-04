#!/usr/bin/env bash
# Install sceno from GitHub releases.
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/niklas-heer/sceno/main/scripts/install.sh | bash
#   curl -fsSL ... | bash -s -- --version v0.1.0
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
  --version VER   Install a specific tag (default: latest release)
  --dir PATH      Install directory (default: /usr/local/bin)
  --no-verify     Skip SHA256 checksum verification
  -h, --help      Show this help

Environment:
  INSTALL_DIR     Same as --dir
  SCENO_VERSION   Same as --version (without v prefix ok)

Examples:
  curl -fsSL https://raw.githubusercontent.com/${REPO}/main/scripts/install.sh | bash
  curl -fsSL ... | bash -s -- --version v0.1.0 --dir ~/.local/bin
EOF
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
  VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | sed -n 's/.*"tag_name": *"v\(.*\)".*/\1/p' | head -1)"
  if [[ -z "$VERSION" ]]; then
    echo "Could not determine latest release version." >&2
    exit 1
  fi
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
mkdir -p "$INSTALL_DIR"
install -m 755 "${TMP}/sceno" "${INSTALL_DIR}/sceno"

echo ""
echo "Installed sceno ${VERSION} to ${INSTALL_DIR}/sceno"
if ! command -v sceno >/dev/null 2>&1; then
  echo "Add ${INSTALL_DIR} to your PATH if needed."
fi
"${INSTALL_DIR}/sceno" version
