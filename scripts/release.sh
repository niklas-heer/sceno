#!/usr/bin/env bash
# Release sceno: suggest semver, run CI, bump VERSION, update CHANGELOG, tag, push.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

# mask injects flags as env vars when set
YES="${yes:-}"
DRY_RUN="${dry_run:-}"
SKIP_CI="${skip_ci:-}"
FORCE="${force:-}"
OVERRIDE_VERSION="${version:-}"

VERSION_FILE="${ROOT}/internal/version/VERSION"
CHANGELOG="${ROOT}/CHANGELOG.md"
SVU="${SVU:-github.com/caarlos0/svu/v3@v3.4.0}"

current_version() {
  tr -d '[:space:]' < "$VERSION_FILE"
}

latest_tag() {
  local current="$1"
  if git rev-parse "v${current}" >/dev/null 2>&1; then
    echo "v${current}"
    return
  fi
  git tag -l 'v*.*.*' --sort=-v:refname | head -1
}

repo_url() {
  local url
  url="$(git remote get-url origin 2>/dev/null || true)"
  url="${url%.git}"
  url="${url#git@github.com:}"
  url="${url#ssh://git@github.com/}"
  url="${url#https://github.com/}"
  url="${url#ssh://github.com/}"
  if [[ -n "$url" ]]; then
    echo "https://github.com/${url}"
  else
    echo "https://github.com/niklas-heer/sceno"
  fi
}

svu_next() {
  go run "$SVU" next --v0 --tag.prefix=""
}

svu_reason() {
  go run "$SVU" next --v0 --verbose --tag.prefix="" 2>&1 | sed '/^[0-9]/d' || true
}

validate_version() {
  local v="$1"
  if [[ ! "$v" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Invalid version: $v (expected X.Y.Z)" >&2
    exit 1
  fi
}

preflight() {
  if ! git rev-parse --git-dir >/dev/null 2>&1; then
    echo "Not a git repository." >&2
    exit 1
  fi

  if [[ -z "$FORCE" && "$(git branch --show-current)" != "main" ]]; then
    echo "Release from main only (current: $(git branch --show-current))." >&2
    echo "Use --force to override." >&2
    exit 1
  fi

  if [[ -z "$DRY_RUN" ]]; then
    if ! git diff --quiet || ! git diff --cached --quiet; then
      echo "Working tree is not clean. Commit or stash changes first." >&2
      exit 1
    fi
  fi

  if [[ -z "$(git remote get-url origin 2>/dev/null || true)" ]]; then
    echo "No git remote 'origin' configured." >&2
    exit 1
  fi
}

changelog_section() {
  local next="$1"
  local tag="$2"
  local base="$3"
  ./scripts/release-notes.sh section "$base" "$next" "$tag"
}

prepend_changelog() {
  local next="$1"
  local tag="$2"
  local base="$3"
  local section tmp

  section="$(changelog_section "$next" "$tag" "$base")"
  tmp="$(mktemp)"

  if [[ ! -f "$CHANGELOG" ]]; then
    cat > "$CHANGELOG" <<EOF
# Changelog

All notable changes to this project are documented in this file.

${section}
EOF
    return
  fi

  {
    head -n 6 "$CHANGELOG"
    echo
    echo "$section"
    tail -n +7 "$CHANGELOG"
  } > "$tmp"
  mv "$tmp" "$CHANGELOG"
}

confirm() {
  local next="$1"
  if [[ -n "$YES" ]]; then
    return 0
  fi
  local reply
  read -r -p "Release v${next}? [Y/n] " reply
  case "${reply:-Y}" in
    y|Y|yes|Yes|YES) return 0 ;;
    *) echo "Aborted."; exit 1 ;;
  esac
}

main() {
  preflight

  local current next tag base reason
  current="$(current_version)"
  tag="$(latest_tag "$current")"
  base="${tag:-}"

  if [[ -n "$OVERRIDE_VERSION" ]]; then
    next="$OVERRIDE_VERSION"
    echo "Using override version: ${next}"
  else
    next="$(svu_next)"
    reason="$(svu_reason)"
  fi

  validate_version "$next"

  if [[ "$next" == "$current" && -z "$OVERRIDE_VERSION" ]]; then
    echo "Nothing to release."
    echo "Current version is ${current} and conventional commits since ${base:-the first commit} do not warrant a bump."
    echo "Use --version X.Y.Z to force a release."
    exit 1
  fi

  echo "Current:  v${current} (${base:-no prior tag})"
  echo "Suggested: v${next}"
  if [[ -n "$reason" ]]; then
    echo
    echo "$reason" | sed '/^$/d'
  fi
  echo
  echo "Commits since ${base:-start}:"
  git log "${base}..HEAD" --oneline --no-merges 2>/dev/null | sed 's/^/  /' || true
  echo

  if [[ -n "$DRY_RUN" ]]; then
    echo "Dry run — no changes made."
    echo "Would: bump VERSION → ${next}, update CHANGELOG, run CI, commit, tag v${next}, push."
    echo
    echo "Release notes preview:"
    echo "---"
    ./scripts/release-notes.sh github "$base" "$next" "v${next}"
    echo "---"
    exit 0
  fi

  confirm "$next"

  if [[ -z "$SKIP_CI" ]]; then
    echo "Running CI..."
    mask ci
  else
    echo "Skipping CI (--skip-ci)."
  fi

  echo "$next" > "$VERSION_FILE"
  prepend_changelog "$next" "v${next}" "$base"

  git add "$VERSION_FILE" "$CHANGELOG"
  git commit -m "chore: release v${next}"
  git tag -a "v${next}" -m "Release v${next}"

  echo
  echo "Pushing main and tag v${next}..."
  git push origin main
  git push origin "v${next}"

  echo
  echo "Released v${next}."
  echo "GitHub Actions will build and publish binaries:"
  echo "  $(repo_url)/actions"
  echo "  $(repo_url)/releases/tag/v${next}"
}

main "$@"
