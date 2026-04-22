#!/usr/bin/env bash
# gen-mock-data.sh
# Generates realistic mock typing stats for the typr visualize demo.
# Places CSV files in ~/.local/share/typr/results/

set -euo pipefail

RESULTS_DIR="${XDG_DATA_HOME:-$HOME/.local/share}/typr/results"
mkdir -p "$RESULTS_DIR"

WORDS_CSV="$RESULTS_DIR/words-stats.csv"

echo "timestamp,wpm,cpm,accuracy,file,n" > "$WORDS_CSV"

# 20 days of data with gradual WPM improvement (40 -> 68 WPM)
# 3-5 entries per day, slight variance per session
# macOS-compatible: date -v-Nd +%s

declare -a DAY_OFFSETS=(20 19 18 17 16 15 14 13 12 11 10 9 8 7 6 5 4 3 2 1)

# Base WPM per day: starts ~40, ends ~68 over 20 days
declare -a BASE_WPM=(40 41 43 42 45 44 47 48 46 50 51 53 52 55 57 56 59 62 65 68)

for i in "${!DAY_OFFSETS[@]}"; do
  offset="${DAY_OFFSETS[$i]}"
  base="${BASE_WPM[$i]}"
  day_ts=$(date -v-${offset}d +%s 2>/dev/null || date -d "${offset} days ago" +%s)

  # 3-5 sessions per day at different hours
  sessions=(9 12 15 18 21)
  count=$(( (i % 3) + 3 ))  # 3, 4, or 5 sessions

  for s in $(seq 0 $(( count - 1 ))); do
    hour="${sessions[$s]}"
    ts=$(( day_ts + hour * 3600 + (s * 137) ))  # small offset per session

    # Variance: +/- 4 WPM within a day
    variance=$(( (s * 3 - 4) ))
    wpm=$(( base + variance ))
    cpm=$(( wpm * 5 ))

    # Accuracy degrades slightly at start of day, improves by end
    acc=$(echo "scale=2; 95 + $s * 0.8 - 1.2" | bc)

    echo "${ts},${wpm},${cpm},${acc},words-demo.txt,10" >> "$WORDS_CSV"
  done
done

echo "Mock data written to $WORDS_CSV"
echo "Rows: $(wc -l < "$WORDS_CSV")"
