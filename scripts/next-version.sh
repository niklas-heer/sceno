#!/usr/bin/env bash
# Compute next semver from conventional commits since the last git tag.
# Uses github.com/caarlos0/svu (same rules as Release Please).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

SVU="${SVU:-github.com/caarlos0/svu/v3@v3.4.0}"
SVU_V0=(--v0)
SVU_PLAIN=(--tag.prefix="")

cmd="${1:-next}"

case "$cmd" in
  next)
    go run "$SVU" next "${SVU_V0[@]}" "${SVU_PLAIN[@]}"
    ;;
  current)
    go run "$SVU" current "${SVU_PLAIN[@]}"
    ;;
  tag)
    go run "$SVU" next "${SVU_V0[@]}"
    ;;
  *)
    echo "Usage: next-version.sh [next|current|tag]" >&2
    exit 1
    ;;
esac
