#!/usr/bin/env bash
# Bump VERSION file (semantic versioning).
# Usage: scripts/bump-version.sh [patch|minor|major]
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
VERSION_FILE="${ROOT}/internal/version/VERSION"
KIND="${1:-patch}"

if [[ ! -f "$VERSION_FILE" ]]; then
  echo "VERSION file not found at $VERSION_FILE" >&2
  exit 1
fi

CURRENT="$(tr -d '[:space:]' < "$VERSION_FILE")"
IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT"

case "$KIND" in
  patch) PATCH=$((PATCH + 1)) ;;
  minor) MINOR=$((MINOR + 1)); PATCH=0 ;;
  major) MAJOR=$((MAJOR + 1)); MINOR=0; PATCH=0 ;;
  *)
    echo "Usage: bump-version.sh [patch|minor|major]" >&2
    exit 1
    ;;
esac

NEXT="${MAJOR}.${MINOR}.${PATCH}"
echo "$NEXT" > "$VERSION_FILE"
echo "Bumped version: ${CURRENT} → ${NEXT}"
echo "Prefer: mask release (automated semver + CI + tag + push)"
