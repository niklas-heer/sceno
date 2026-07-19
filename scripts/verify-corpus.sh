#!/usr/bin/env bash
set -euo pipefail

binary="${SCENO_BIN:-./sceno}"
output="${SCENO_VERIFY_OUT:-$(mktemp -d)}"

if [[ ! -x "$binary" ]]; then
  echo "verify-corpus: binary not executable: $binary" >&2
  exit 2
fi

mkdir -p "$output"
count=0
while IFS= read -r file; do
	count=$((count + 1))
  name="${file#examples/}"
  name="${name%.kdl}"
  name="${name//\//-}"
  case_dir="$output/$name"
  mkdir -p "$case_dir"

  "$binary" validate -i "$file" --json >"$case_dir/validate.json"
  grep -q '"ok": true' "$case_dir/validate.json"

	"$binary" advise -i "$file" --json >"$case_dir/advise.json"
	grep -q '"scene_stack"' "$case_dir/advise.json"
	visual_score=$(sed -n 's/^[[:space:]]*"visual_score": \([0-9][0-9]*\),*$/\1/p' "$case_dir/advise.json" | head -n 1)
	if [[ -z "$visual_score" ]] || (( visual_score < 60 )); then
		echo "verify-corpus: visual score for $file is ${visual_score:-missing}, expected at least 60" >&2
		exit 1
	fi

  "$binary" describe -i "$file" --json >"$case_dir/describe.json"
  grep -q '"render_ready": true' "$case_dir/describe.json"
  grep -q '"scene_stack"' "$case_dir/describe.json"

  "$binary" render -i "$file" -o "$case_dir/render" --all >"$case_dir/render.log"
  artifact_count=$(find "$case_dir" -maxdepth 1 -type f -name 'render*' -size +99c | wc -l | tr -d '[:space:]')
  if (( artifact_count < 5 )); then
    echo "verify-corpus: expected at least 5 render artifacts for $file, got $artifact_count" >&2
    exit 1
  fi
done < <(find examples -type f -name '*.kdl' | sort)

if (( count == 0 )); then
	echo "verify-corpus: no KDL examples found" >&2
	exit 2
fi

printf 'verified %d KDL files through validate, advise, describe, and all exports (%s)\n' "$count" "$output"
