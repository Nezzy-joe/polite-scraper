#!/usr/bin/env bash
# Measure both approaches uniformly with /usr/bin/time -v (wall clock + peak RSS)
# so the only difference between them is "plain HTTP" vs "headless browser".
set -euo pipefail
cd "$(dirname "$0")"
source .venv/bin/activate

run() {                       # $1 = label, $2 = script
  local log; log="$(mktemp)"
  /usr/bin/time -v python "$2" 2> "$log"
  local wall rss
  wall=$(grep 'wall clock' "$log" | grep -oE '[0-9:.]+$')
  rss=$(grep 'Maximum resident set size' "$log" | grep -oE '[0-9]+$')
  printf '%-9s wall=%-10s peak_rss=%s MB\n' "$1" "$wall" "$(( rss / 1024 ))"
  rm -f "$log"
}

echo "=== PLAIN HTTP ==="
run "plain" plain.py
echo
echo "=== HEADLESS BROWSER (Playwright/Chromium) ==="
run "browser" browser.py
