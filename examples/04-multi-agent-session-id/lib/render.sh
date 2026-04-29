#!/usr/bin/env bash
# Render son/father agent configs from jq templates.

source "$(dirname "${BASH_SOURCE[0]}")/common.sh"

# Render a son config (exp_son1 or exp_son2).
# Args:
#   $1 agent_name (e.g. "exp_son1")
#   $2 skill_id (exp_session_echo's UUID)
#   $3 output file path
render_son_config() {
  local agent_name="$1" skill_id="$2" out="$3"
  jq -f "$EXP_DIR/configs/son.config.template.jq" \
     --arg agent_name "$agent_name" \
     --arg skill_id   "$skill_id" \
     "$EXP_DIR/configs/base.config.json" > "$out"
}

# Render father config.
# Args:
#   $1 son1_key, $2 son1_version
#   $3 son2_key, $4 son2_version
#   $5 output file path
render_father_config() {
  local son1_key="$1" son1_ver="$2" son2_key="$3" son2_ver="$4" out="$5"
  jq -f "$EXP_DIR/configs/father.config.template.jq" \
     --arg son1_key     "$son1_key" \
     --arg son1_version "$son1_ver" \
     --arg son2_key     "$son2_key" \
     --arg son2_version "$son2_ver" \
     "$EXP_DIR/configs/base.config.json" > "$out"
}
