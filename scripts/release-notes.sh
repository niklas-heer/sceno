#!/usr/bin/env bash
# Generate and extract release notes from conventional commits.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CHANGELOG="${ROOT}/CHANGELOG.md"

repo_url() {
  local url
  url="$(git -C "$ROOT" remote get-url origin 2>/dev/null || true)"
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

raw_install_script_url() {
  local slug
  slug="$(repo_url)"
  slug="${slug#https://github.com/}"
  echo "https://raw.githubusercontent.com/${slug}/main/scripts/install.sh"
}

commit_link() {
  local hash="$1"
  local short="${hash:0:7}"
  echo "([${short}]($(repo_url)/commit/${hash}))"
}

# Strip conventional commit prefix; returns "scope|message" or "|message"
parse_subject() {
  local subject="$1"
  local line

  line="$(printf '%s' "$subject" | sed -E -n \
    's/^([a-zA-Z]+)(\(([^)]+)\))?!?:[[:space:]]*(.+)$/\1|\3|\4/p')"
  if [[ -n "$line" ]]; then
    printf '%s\n' "$line"
    return 0
  fi

  line="$(printf '%s' "$subject" | sed -E -n \
    's/^([a-zA-Z]+):[[:space:]]*(.+)$/\1||\2/p')"
  if [[ -n "$line" ]]; then
    printf '%s\n' "$line"
    return 0
  fi

  printf '|%s\n' "$subject"
}

section_for_type() {
  case "$1" in
    feat) echo "Features" ;;
    fix) echo "Bug Fixes" ;;
    perf) echo "Performance" ;;
    refactor) echo "Refactoring" ;;
    docs) echo "Documentation" ;;
    test) echo "Tests" ;;
    build) echo "Build System" ;;
    ci) echo "CI/CD" ;;
    style) echo "Style" ;;
    revert) echo "Reverts" ;;
    chore) echo "Chores" ;;
    *) echo "Other" ;;
  esac
}

should_skip_commit() {
  local subject="$1"
  case "$subject" in
    chore:\ release\ v*)
      return 0
      ;;
  esac
  if [[ "$subject" == "chore(release):"* ]]; then
    return 0
  fi
  return 1
}

format_entry() {
  local hash="$1"
  local scope="$2"
  local message="$3"
  local link
  link="$(commit_link "$hash")"

  if [[ -n "$scope" ]]; then
    printf '* **%s**: %s %s\n' "$scope" "$message" "$link"
  else
    printf '* %s %s\n' "$message" "$link"
  fi
}

is_breaking() {
  local hash="$1"
  local subject="$2"
  local body

  if [[ "$subject" == *"!":* ]]; then
    return 0
  fi

  body="$(git -C "$ROOT" log -1 --format=%B "$hash" 2>/dev/null || true)"
  if grep -qi 'BREAKING CHANGE:' <<< "$body"; then
    return 0
  fi
  return 1
}

breaking_summary() {
  local hash="$1"
  local body line
  body="$(git -C "$ROOT" log -1 --format=%B "$hash" 2>/dev/null || true)"
  while IFS= read -r line; do
    if [[ "$line" =~ ^BREAKING[[:space:]]CHANGE:[[:space:]]*(.*)$ ]]; then
      echo "${BASH_REMATCH[1]}"
      return 0
    fi
  done <<< "$body"
  git -C "$ROOT" log -1 --format=%s "$hash" | sed -E 's/^[a-zA-Z]+(\([^)]+\))?!?: //'
}

append_bucket() {
  local section="$1"
  local entry="$2"
  case "$section" in
    Features) _b_features+=$'\n'"$entry" ;;
    "Bug Fixes") _b_fixes+=$'\n'"$entry" ;;
    Performance) _b_perf+=$'\n'"$entry" ;;
    Refactoring) _b_refactor+=$'\n'"$entry" ;;
    Documentation) _b_docs+=$'\n'"$entry" ;;
    "CI/CD") _b_ci+=$'\n'"$entry" ;;
    "Build System") _b_build+=$'\n'"$entry" ;;
    Tests) _b_tests+=$'\n'"$entry" ;;
    Style) _b_style+=$'\n'"$entry" ;;
    Reverts) _b_reverts+=$'\n'"$entry" ;;
    Other) _b_other+=$'\n'"$entry" ;;
  esac
}

emit_bucket() {
  local title="$1"
  local content="$2"
  [[ -n "${content//$'\n'/}" ]] || return 0
  echo "### ${title}"
  echo "$content" | sed '/^$/d'
  echo
}

generate_section() {
  local version="$1"
  local tag="$2"
  local base="$3"
  local url date range

  url="$(repo_url)"
  date="$(date +%Y-%m-%d)"
  range="${base}..HEAD"
  if [[ -z "$base" ]]; then
    range="HEAD"
  fi

  local breaking=""
  local _b_features="" _b_fixes="" _b_perf="" _b_refactor="" _b_docs=""
  local _b_ci="" _b_build="" _b_tests="" _b_style="" _b_reverts="" _b_other=""
  local hash subject parsed type scope message section entry

  while IFS='|' read -r hash subject; do
    [[ -z "$hash" ]] && continue
    should_skip_commit "$subject" && continue

    parsed="$(parse_subject "$subject")"
    IFS='|' read -r type scope message <<< "$parsed"
    [[ -z "$message" && -n "$scope" && -z "$type" ]] && message="$scope" && scope=""

    if is_breaking "$hash" "$subject"; then
      if [[ -n "$scope" ]]; then
        breaking+=$'\n'"* **${scope}**: $(breaking_summary "$hash") $(commit_link "$hash")"
      else
        breaking+=$'\n'"* $(breaking_summary "$hash") $(commit_link "$hash")"
      fi
      continue
    fi

    if [[ -z "$type" ]]; then
      section="Other"
    else
      section="$(section_for_type "$type")"
      [[ "$section" == "Chores" ]] && continue
    fi

    entry="$(format_entry "$hash" "$scope" "${message:-$subject}")"
    append_bucket "$section" "$entry"
  done < <(git -C "$ROOT" log "$range" --pretty=format:'%H|%s' --no-merges 2>/dev/null; echo)

  {
    echo "## [${version}](${url}/releases/tag/${tag}) (${date})"
    echo

    if [[ -n "${breaking//$'\n'/}" ]]; then
      echo "### ⚠️ Breaking Changes"
      echo "$breaking" | sed '/^$/d'
      echo
    fi

    emit_bucket "Features" "$_b_features"
    emit_bucket "Bug Fixes" "$_b_fixes"
    emit_bucket "Performance" "$_b_perf"
    emit_bucket "Refactoring" "$_b_refactor"
    emit_bucket "Documentation" "$_b_docs"
    emit_bucket "CI/CD" "$_b_ci"
    emit_bucket "Build System" "$_b_build"
    emit_bucket "Tests" "$_b_tests"
    emit_bucket "Style" "$_b_style"
    emit_bucket "Reverts" "$_b_reverts"
    emit_bucket "Other" "$_b_other"
  }
}

generate_github_body() {
  local version="$1"
  local tag="$2"
  local base="$3"
  local url section

  url="$(repo_url)"
  section="$(generate_section "$version" "$tag" "$base")"

  {
    echo "# Sceno ${tag}"
    echo
    echo "$section" | tail -n +3 | sed '/^$/d'
    echo
    echo "---"
    echo
    echo "## Install"
    echo
    echo '```bash'
    echo "curl -fsSL $(raw_install_script_url) | bash"
    echo '```'
    echo
    echo "Pin this release:"
    echo
    echo '```bash'
    echo "curl -fsSL $(raw_install_script_url) | bash -s -- --version ${tag}"
    echo '```'
    echo
    echo "Or download \`${tag}\` assets from [GitHub Releases](${url}/releases/tag/${tag})."
  }
}

extract_from_changelog() {
  local version="$1"
  if [[ ! -f "$CHANGELOG" ]]; then
    echo "CHANGELOG.md not found" >&2
    exit 1
  fi
  awk -v ver="$version" '
    $0 ~ "^## \\[" ver "\\]" { found=1; print; next }
    found && /^## \[/ { exit }
    found { print }
  ' "$CHANGELOG"
}

extract_github_body() {
  local version="$1"
  local tag="v${version}"
  local url section

  url="$(repo_url)"
  section="$(extract_from_changelog "$version")"
  if [[ -z "${section//$'\n'/}" ]]; then
    echo "No changelog section for version ${version}" >&2
    exit 1
  fi

  {
    echo "# Sceno ${tag}"
    echo
    echo "$section" | tail -n +2 | sed '/^$/d'
    echo
    echo "---"
    echo
    echo "## Install"
    echo
    echo '```bash'
    echo "curl -fsSL $(raw_install_script_url) | bash"
    echo '```'
    echo
    echo "Pin this release:"
    echo
    echo '```bash'
    echo "curl -fsSL $(raw_install_script_url) | bash -s -- --version ${tag}"
    echo '```'
    echo
    echo "Or download \`${tag}\` assets from [GitHub Releases](${url}/releases/tag/${tag})."
  }
}

usage() {
  cat <<EOF
Usage:
  release-notes.sh section <base_ref> <version> <tag>   # CHANGELOG section
  release-notes.sh github <base_ref> <version> <tag>    # GitHub release body (pre-release)
  release-notes.sh extract <version>                    # GitHub body from CHANGELOG (CI)
EOF
}

cmd="${1:-}"
shift || true

case "$cmd" in
  section)
    generate_section "${2:?version}" "${3:?tag}" "${1:-}"
    ;;
  github)
    generate_github_body "${2:?version}" "${3:?tag}" "${1:-}"
    ;;
  extract)
    extract_github_body "${1:?version}"
    ;;
  *)
    usage >&2
    exit 1
    ;;
esac
