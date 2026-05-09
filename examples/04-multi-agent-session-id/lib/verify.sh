#!/usr/bin/env bash
# Two assertions verifying session_id propagation through father -> sons -> skill.
#
# Both assertions read from the chat completion response JSON. The response
# stores each son's output under .message.content.final_answer.answer_type_other.
# (We don't use `kweaver agent trace` because the platform's trace endpoint
# is currently broken — returns 500.)

source "$(dirname "${BASH_SOURCE[0]}")/common.sh"

# Helper: extract son's textual answer from response JSON.
# Args: $1 response file path, $2 res_1 or res_2
_son_answer() {
  local file="$1" key="$2"
  jq -r --arg k "$key" '.message.content.final_answer.answer_type_other[$k].answer.answer // ""' "$file"
}

# Helper: extract son's input_message (which the platform builds from
# father's @<son>(...) call args plus its system prompt).
_son_input_message() {
  local file="$1" key="$2"
  jq -r --arg k "$key" '.message.content.final_answer.answer_type_other[$k].answer.input_message // ""' "$file"
}

# Assertion 1: each son's answer contains the literal echo line.
assert_literal_in_answer() {
  local resp_file="$1" sid="$2"
  local missing=0
  for entry in "res_1:$EXP_SON1_NAME" "res_2:$EXP_SON2_NAME"; do
    local key="${entry%%:*}"
    local son="${entry#*:}"
    local ans
    ans=$(_son_answer "$resp_file" "$key")
    local expected="[exp_session_echo] RECEIVED session_id=$sid from $son"
    if echo "$ans" | grep -qF "$expected"; then
      log "assert_literal: $son — PASS"
    else
      warn "assert_literal: $son — FAIL"
      warn "  expected substring: $expected"
      warn "  actual answer (first 200 chars): ${ans:0:200}"
      missing=1
    fi
  done
  [ "$missing" = "0" ] && return 0 || return 1
}

# Assertion 2: each son's input_message starts with the [input.session_id=<sid>]
# prefix injected by son's dolphin DSL — proving the platform really
# routed father's input.session_id into the son call.
assert_session_id_in_son_input() {
  local resp_file="$1" sid="$2"
  local missing=0
  for entry in "res_1:$EXP_SON1_NAME" "res_2:$EXP_SON2_NAME"; do
    local key="${entry%%:*}"
    local son="${entry#*:}"
    local input_msg
    input_msg=$(_son_input_message "$resp_file" "$key")
    local expected="[input.session_id=$sid]"
    if echo "$input_msg" | grep -qF "$expected"; then
      log "assert_propagation: $son input contains $expected — PASS"
    else
      warn "assert_propagation: $son — FAIL"
      warn "  expected substring: $expected"
      warn "  actual input_message (first 300 chars): ${input_msg:0:300}"
      missing=1
    fi
  done
  [ "$missing" = "0" ] && return 0 || return 1
}
