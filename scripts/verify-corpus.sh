#!/usr/bin/env bash
set -euo pipefail

binary="${SCENO_BIN:-./sceno}"
output="${SCENO_VERIFY_OUT:-$(mktemp -d)}"
minimum_visual_score=80
blocked_visual_codes='collision|edge_collision|edge_detour|edge_hidden|arrow_detached|arrow_hidden|arrowhead_cluster|edge_label_overlap|edge_label_chrome_overlap|edge_label_off_axis|edge_side_mismatch|occluded|text_overflow'

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
	if [[ -z "$visual_score" ]] || (( visual_score < minimum_visual_score )); then
		echo "verify-corpus: visual score for $file is ${visual_score:-missing}, expected at least $minimum_visual_score" >&2
		exit 1
	fi
	if grep -Eq '"code": "('"$blocked_visual_codes"')"' "$case_dir/advise.json"; then
		codes=$(sed -nE 's/^[[:space:]]*"code": "('"$blocked_visual_codes"')",?$/\1/p' "$case_dir/advise.json" | sort -u | tr '\n' ' ')
		echo "verify-corpus: structural visual finding for $file: $codes" >&2
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
