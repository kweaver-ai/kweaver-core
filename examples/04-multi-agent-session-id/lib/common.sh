#!/usr/bin/env bash
# Shared variables and helpers for the exp multi-agent demo.

set -euo pipefail

EXP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EXP_SKILL_NAME="exp_session_echo"
EXP_FATHER_NAME="exp_father"
EXP_SON1_NAME="exp_son1"
EXP_SON2_NAME="exp_son2"

log()  { printf '\033[1;36m[exp]\033[0m %s\n' "$*" >&2; }
warn() { printf '\033[1;33m[exp WARN]\033[0m %s\n' "$*" >&2; }
fail() { printf '\033[1;31m[exp FAIL]\033[0m %s\n' "$*" >&2; exit 1; }

# Resolve agent id by exact name (personal-list). Empty if not found.
resolve_agent_id() {
  local name="$1"
  kweaver agent personal-list --name "$name" 2>/dev/null \
    | jq -r --arg n "$name" '.[]? | select(.name == $n) | .id' \
    | head -1
}

# Resolve skill id by exact name (skill list). Empty if not found.
resolve_skill_id() {
  local name="$1"
  kweaver skill list --name "$name" 2>/dev/null \
    | jq -r --arg n "$name" '.[]? | select(.name == $n) | .id' \
    | head -1
}

# Get agent key + version. Writes to caller's AGENT_KEY / AGENT_VER.
# After publish, version becomes "v1"; pre-publish .version is null and we
# default to "v0".
fetch_agent_keyver() {
  local id="$1"
  local raw
  raw=$(kweaver agent get "$id" --verbose 2>/dev/null)
  AGENT_KEY=$(echo "$raw" | jq -r '.key // .agent_key // empty')
  AGENT_VER=$(echo "$raw" | jq -r '.version // "v1"')
  [ -n "$AGENT_KEY" ] || fail "agent $id has no key in get response"
}

# Pull the platform default LLM id from base.config.json. The template
# stores it at .llms[0].llm_config.id (NOT .llms[0].id).
get_default_llm_id() {
  jq -r '.llms[0].llm_config.id // empty' "$EXP_DIR/configs/base.config.json"
}
