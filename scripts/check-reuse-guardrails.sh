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

need python
need sed
need grep

search_cursor_sites() {
  if command -v rg >/dev/null 2>&1; then
    rg -n 'GetString\("cursor"\)' cmd/*.go
    return
  fi
  grep -nH 'GetString("cursor")' cmd/*.go
}

fail=0

while IFS=: read -r file line _; do
  [[ -z "${file:-}" ]] && continue
  if [[ "$file" == "cmd/helpers.go" ]]; then
    continue
  fi

  start=$((line - 3))
  if (( start < 1 )); then
    start=1
  fi

  context=$(sed -n "${start},${line}p" "$file")
  if ! grep -q "reuse-guardrail: allow-manual-pagination" <<<"$context"; then
    echo "unexpected manual cursor parsing: ${file}:${line}" >&2
    fail=1
  fi
done < <(search_cursor_sites)

if [[ -f docs/reuse/waivers.yaml ]]; then
  if ! python - docs/reuse/waivers.yaml <<'PY'
import datetime
import pathlib
import re
import sys

path = pathlib.Path(sys.argv[1])
today = datetime.date.today()
overdue = []

for line_no, line in enumerate(path.read_text().splitlines(), start=1):
    match = re.search(r'expires_on:\s*([0-9]{4}-[0-9]{2}-[0-9]{2})', line)
    if not match:
        continue
    try:
        value = datetime.date.fromisoformat(match.group(1))
    except ValueError:
        continue
    if value < today:
        overdue.append(f"{path}:{line_no}: overdue waiver expired on {value.isoformat()}")

if overdue:
    for item in overdue:
        print(item, file=sys.stderr)
    raise SystemExit(1)
PY
  then
    fail=1
  fi
fi

if [[ $fail -ne 0 ]]; then
  exit 1
fi

echo "reuse guardrails: OK"
