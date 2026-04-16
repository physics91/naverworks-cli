#!/usr/bin/env bash
set -euo pipefail

ROOT="${1:-.}"
cd "$ROOT"

need() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "missing required command: $1" >&2
    exit 1
  }
}

need awk
need sort
need head
need python
need grep

count_matches() {
  local pattern="$1"
  if command -v rg >/dev/null 2>&1; then
    rg -c "$pattern" cmd/*.go | awk -F: '{s+=$2} END {print s+0}'
    return
  fi
  grep -cH -E "$pattern" cmd/*.go | awk -F: '{s+=$2} END {print s+0}'
}

hotspot_matches() {
  local pattern="$1"
  if command -v rg >/dev/null 2>&1; then
    rg -c "$pattern" cmd/*.go
    return
  fi
  grep -cH -E "$pattern" cmd/*.go
}

waiver_status() {
  local mode="$1"
  local path="docs/reuse/waivers.yaml"
  if [[ ! -f "$path" ]]; then
    echo "n/a"
    return
  fi

  python - "$path" "$mode" <<'PY'
import datetime
import pathlib
import re
import sys

path = pathlib.Path(sys.argv[1])
mode = sys.argv[2]
today = datetime.date.today()
count = 0

for line in path.read_text().splitlines():
    match = re.search(r'expires_on:\s*([0-9]{4}-[0-9]{2}-[0-9]{2})', line)
    if not match:
        continue
    try:
        value = datetime.date.fromisoformat(match.group(1))
    except ValueError:
        continue
    if mode == "overdue" and value < today:
        count += 1
    if mode == "expiring" and 0 <= (value - today).days <= 14:
        count += 1

print(count)
PY
}

printf "# Reuse Scorecard\n\n"
printf "## Helper Adoption\n\n"
printf "| Metric | Count |\n"
printf "|---|---:|\n"
printf "| newSvc call sites | %s |\n" "$(count_matches 'newSvc\(')"
printf "| runListCmd call sites | %s |\n" "$(count_matches 'runListCmd\(')"
printf "| fetchAndPrint call sites | %s |\n" "$(count_matches 'fetchAndPrint\(')"
printf "| readJSONFlagRaw call sites | %s |\n" "$(count_matches 'readJSONFlagRaw\(')"
printf "| printBody call sites | %s |\n" "$(count_matches 'printBody\(')"
printf "| manual cursor parsing sites | %s |\n" "$(count_matches 'GetString\(\"cursor\"\)')"
printf "| manual count parsing sites | %s |\n" "$(count_matches 'GetInt\(\"count\"\)')"
printf "| overdue waivers | %s |\n" "$(waiver_status overdue)"
printf "| waivers expiring within 14d | %s |\n" "$(waiver_status expiring)"

printf "\n## Hotspots\n\n"
printf "| File | Helper Touches |\n"
printf "|---|---:|\n"
hotspot_matches 'newSvc\(|runListCmd\(|fetchAndPrint\(|readJSONFlagRaw\(|printBody\(' \
  | sort -t: -k2,2nr \
  | head -10 \
  | awk -F: '{printf "| `%s` | %s |\n", $1, $2}'
