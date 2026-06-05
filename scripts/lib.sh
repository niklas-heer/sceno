#!/usr/bin/env bash
# Shared build metadata for mask tasks and release scripts.
set -euo pipefail

sceno_root() {
  local src="${BASH_SOURCE[1]:-${BASH_SOURCE[0]}}"
  cd "$(dirname "$src")/.." && pwd
}

sceno_env() {
  ROOT="$(sceno_root)"
  cd "$ROOT"
  BINARY="${BINARY:-sceno}"
  CMD="${CMD:-./cmd/sceno}"
  VERSION_FILE="${ROOT}/internal/version/VERSION"
  VERSION="$(tr -d '[:space:]' < "$VERSION_FILE")"
  COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"
  DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  LDFLAGS="-s -w \
    -X github.com/niklas-heer/sceno/internal/version.Version=${VERSION} \
    -X github.com/niklas-heer/sceno/internal/version.Commit=${COMMIT} \
    -X github.com/niklas-heer/sceno/internal/version.Date=${DATE}"
  export ROOT BINARY CMD VERSION_FILE VERSION COMMIT DATE LDFLAGS
}
