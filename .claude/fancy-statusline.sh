#!/usr/bin/env bash

# Read JSON from stdin
input=$(cat)

# Extract core fields using jq with fallback defaults
MODEL=$(echo "$input" | jq -r '.model.display_name // .model.id // "Claude"')
DIR=$(echo "$input" | jq -r '.workspace.current_dir // ""')
DIR_NAME="${DIR##*/}"
[ -z "$DIR_NAME" ] && DIR_NAME="workspace"

SESSION_ID=$(echo "$input" | jq -r '.session_id // "default"')
COST=$(echo "$input" | jq -r '.cost.total_cost_usd // 0')
PCT=$(echo "$input" | jq -r '.context_window.used_percentage // 0' | cut -d. -f1)
DURATION_MS=$(echo "$input" | jq -r '.cost.total_duration_ms // 0')

# Extract cache and token details from current_usage with fallback
CACHE_READ=$(echo "$input" | jq -r '.context_window.current_usage.cache_read_input_tokens // 0')
CACHE_WRITE=$(echo "$input" | jq -r '.context_window.current_usage.cache_creation_input_tokens // 0')
INPUT_TOKENS=$(echo "$input" | jq -r '.context_window.current_usage.input_tokens // 0')

# Git status cache to prevent terminal lag
CACHE_FILE="/tmp/statusline-git-cache-$SESSION_ID"
CACHE_MAX_AGE=5 # seconds

cache_is_stale() {
  [ ! -f "$CACHE_FILE" ] || \
  [ $(( $(date +%s) - $(stat -c %Y "$CACHE_FILE" 2>/dev/null || stat -f %m "$CACHE_FILE" 2>/dev/null || echo 0) )) -gt $CACHE_MAX_AGE ]
}

if cache_is_stale; then
  if git rev-parse --git-dir >/dev/null 2>&1; then
    BRANCH=$(git branch --show-current 2>/dev/null)
    STAGED=$(git diff --cached --numstat 2>/dev/null | wc -l | tr -d ' ')
    MODIFIED=$(git diff --numstat 2>/dev/null | wc -l | tr -d ' ')
    echo "$BRANCH|$STAGED|$MODIFIED" > "$CACHE_FILE"
  else
    echo "||" > "$CACHE_FILE"
  fi
fi

if [ -f "$CACHE_FILE" ]; then
  IFS='|' read -r BRANCH STAGED MODIFIED < "$CACHE_FILE"
else
  BRANCH=""
  STAGED=0
  MODIFIED=0
fi

# Colors (ANSI escape codes)
CYAN='\033[36m'
GREEN='\033[32m'
YELLOW='\033[33m'
RED='\033[31m'
MAGENTA='\033[35m'
BLUE='\033[34m'
RESET='\033[0m'
BOLD='\033[1m'

# Build Git indicator
GIT_STR=""
if [ -n "$BRANCH" ]; then
  GIT_COLOR="$GREEN"
  if [ "$MODIFIED" -gt 0 ] || [ "$STAGED" -gt 0 ]; then
    GIT_COLOR="$YELLOW"
  fi
  GIT_STR=" 🌿 ${GIT_COLOR}${BRANCH}${RESET}"
  if [ "$STAGED" -gt 0 ]; then
    GIT_STR="${GIT_STR} ${GREEN}+${STAGED}${RESET}"
  fi
  if [ "$MODIFIED" -gt 0 ]; then
    GIT_STR="${GIT_STR} ${YELLOW}~${MODIFIED}${RESET}"
  fi
fi

# Build Context Progress Bar (10 chars width)
if [ "$PCT" -ge 90 ]; then
  BAR_COLOR="$RED"
elif [ "$PCT" -ge 70 ]; then
  BAR_COLOR="$YELLOW"
else
  BAR_COLOR="$GREEN"
fi

FILLED=$((PCT / 10))
EMPTY=$((10 - FILLED))
# Ensure values are within boundaries
[ "$FILLED" -gt 10 ] && FILLED=10
[ "$FILLED" -lt 0 ] && FILLED=0
[ "$EMPTY" -gt 10 ] && EMPTY=10
[ "$EMPTY" -lt 0 ] && EMPTY=0

printf -v FILL "%${FILLED}s"
printf -v PAD "%${EMPTY}s"
BAR="${FILL// /█}${PAD// /░}"

# Calculate duration format
MINS=$((DURATION_MS / 60000))
SECS=$(((DURATION_MS % 60000) / 1000))
TIME_STR="${MINS}m ${SECS}s"

# Calculate cost format
COST_FMT=$(printf '$%.2f' "$COST")

# Calculate Cache Hit Performance Metric
TOTAL_INPUT=$((INPUT_TOKENS + CACHE_READ + CACHE_WRITE))
CACHE_HIT_PCT=0
if [ "$TOTAL_INPUT" -gt 0 ]; then
  CACHE_HIT_PCT=$((CACHE_READ * 100 / TOTAL_INPUT))
fi

# Convert token counts to readable short formats (e.g. 15k)
format_tokens() {
  local num=$1
  if [ "$num" -ge 1000 ]; then
    echo "$((num / 1000))k"
  else
    echo "$num"
  fi
}

READ_FMT=$(format_tokens "$CACHE_READ")
WRITE_FMT=$(format_tokens "$CACHE_WRITE")

# Format Cache Stats
CACHE_STR=""
if [ "$TOTAL_INPUT" -gt 0 ]; then
  CACHE_STR=" | ⚡ Hit: ${MAGENTA}${CACHE_HIT_PCT}%${RESET} (Read: ${GREEN}${READ_FMT}${RESET}, Write: ${YELLOW}${WRITE_FMT}${RESET})"
else
  CACHE_STR=" | ⚡ Hit: ${MAGENTA}0%${RESET} (Read: 0, Write: 0)"
fi

# Build the status lines
echo -e " ${CYAN}[$MODEL]${RESET} 📁 ${BOLD}${DIR_NAME}${RESET}${GIT_STR}"
echo -e " ${BAR_COLOR}${BAR}${RESET} ${PCT}% | 💰 ${YELLOW}${COST_FMT}${RESET} | ⏱️ ${TIME_STR}${CACHE_STR}"
