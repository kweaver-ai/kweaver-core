#!/usr/bin/env bash
# End-to-end demo: verify session_id propagates user -> exp_father ->
# exp_son1 / exp_son2 -> exp_session_echo SKILL.
#
# Usage:
#   ./run.sh                       # ensure artifacts + invoke + verify; keep on platform
#   ./run.sh --session-id MY-XX    # override the session_id (default: DEMO-2026-XXXXXX)
#   ./run.sh --cleanup             # unpublish + delete the three agents and the skill

set -euo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
source "$HERE/lib/common.sh"
source "$HERE/lib/render.sh"
source "$HERE/lib/verify.sh"
source "$HERE/lib/cleanup.sh"

SID=""
DO_CLEANUP=0
while [ $# -gt 0 ]; do
  case "$1" in
    --session-id) SID="$2"; shift 2 ;;
    --cleanup)    DO_CLEANUP=1; shift ;;
    -h|--help)    sed -n '1,12p' "$0" >&2; exit 0 ;;
    *) fail "unknown arg: $1" ;;
  esac
done

if [ "$DO_CLEANUP" = "1" ]; then
  cleanup_all
  log "cleanup done"
  exit 0
fi

[ -z "$SID" ] && SID="DEMO-2026-$(LC_ALL=C tr -dc 'A-Z0-9' </dev/urandom | head -c 6)"
log "session_id=$SID"

# === ensure SKILL ===
SKILL_ID=$(resolve_skill_id "$EXP_SKILL_NAME")
if [ -z "$SKILL_ID" ]; then
  log "registering skill $EXP_SKILL_NAME"
  resp=$(kweaver skill register --content-file "$HERE/skills/exp_session_echo/SKILL.md" 2>/dev/null)
  SKILL_ID=$(echo "$resp" | jq -r '.id // empty')
  [ -n "$SKILL_ID" ] || fail "skill register failed: $resp"
  kweaver skill status "$SKILL_ID" published >/dev/null 2>&1 || fail "skill publish failed"
fi
log "skill_id=$SKILL_ID (reused if existed)"

LLM_ID=$(get_default_llm_id)
[ -n "$LLM_ID" ] || fail "no LLM id in base.config.json (.llms[0].llm_config.id is empty)"

# === ensure son1 + son2 ===
ensure_son() {
  local name="$1" out_var="$2"
  local id
  id=$(resolve_agent_id "$name")
  if [ -z "$id" ]; then
    log "creating agent $name"
    local render_path="/tmp/exp_${name}_render.json"
    render_son_config "$name" "$SKILL_ID" "$render_path"
    local resp
    resp=$(kweaver agent create --name "$name" \
            --profile "exp demo $name" \
            --llm-id "$LLM_ID" \
            --config "$render_path" 2>/dev/null)
    id=$(echo "$resp" | jq -r '.id // empty')
    [ -n "$id" ] || fail "create $name failed: $resp"
    kweaver agent publish "$id" >/dev/null 2>&1 || fail "publish $name failed"
  else
    log "agent $name already exists ($id), reusing as-is"
  fi
  printf -v "$out_var" '%s' "$id"
}

SON1_ID=""
SON2_ID=""
ensure_son "$EXP_SON1_NAME" SON1_ID
ensure_son "$EXP_SON2_NAME" SON2_ID

fetch_agent_keyver "$SON1_ID"; SON1_KEY="$AGENT_KEY"; SON1_VER="$AGENT_VER"
fetch_agent_keyver "$SON2_ID"; SON2_KEY="$AGENT_KEY"; SON2_VER="$AGENT_VER"
log "son1: id=$SON1_ID key=$SON1_KEY ver=$SON1_VER"
log "son2: id=$SON2_ID key=$SON2_KEY ver=$SON2_VER"

# === ensure father ===
FATHER_ID=$(resolve_agent_id "$EXP_FATHER_NAME")
if [ -z "$FATHER_ID" ]; then
  log "creating agent $EXP_FATHER_NAME"
  render_father_config "$SON1_KEY" "$SON1_VER" "$SON2_KEY" "$SON2_VER" /tmp/exp_father_render.json
  resp=$(kweaver agent create --name "$EXP_FATHER_NAME" \
          --profile "exp demo father" \
          --llm-id "$LLM_ID" \
          --config /tmp/exp_father_render.json 2>/dev/null)
  FATHER_ID=$(echo "$resp" | jq -r '.id // empty')
  [ -n "$FATHER_ID" ] || fail "create father failed: $resp"
  kweaver agent publish "$FATHER_ID" >/dev/null 2>&1 || fail "publish father failed"
else
  log "agent $EXP_FATHER_NAME already exists ($FATHER_ID), reusing as-is"
fi
fetch_agent_keyver "$FATHER_ID"; FATHER_KEY="$AGENT_KEY"; FATHER_VER="$AGENT_VER"
log "father: id=$FATHER_ID key=$FATHER_KEY ver=$FATHER_VER"

# === invoke father ===
log "invoking father with custom_querys.session_id=$SID..."
BODY=$(jq -nc \
  --arg id  "$FATHER_ID" \
  --arg key "$FATHER_KEY" \
  --arg ver "$FATHER_VER" \
  --arg sid "$SID" '{
    agent_id:      $id,
    agent_key:     $key,
    agent_version: $ver,
    query:         "请回显 session_id 验证透传",
    custom_querys: { session_id: $sid },
    stream:        false
  }')

RESP_FILE=/tmp/exp_run_resp.json
START=$(date +%s)
kweaver call -X POST "/api/agent-factory/v1/app/$FATHER_KEY/chat/completion" \
  -H "content-type: application/json" -d "$BODY" 2>/dev/null > "$RESP_FILE"
END=$(date +%s)

CONV_ID=$(jq -r '.conversation_id // empty' "$RESP_FILE")
if [ -z "$CONV_ID" ]; then
  warn "no conversation_id in response (the call may have failed)"
  log "first 600 bytes of response:"
  head -c 600 "$RESP_FILE" >&2
  fail "father call did not return a conversation_id"
fi
log "conversation_id=$CONV_ID, elapsed=$((END-START))s"

log "son1 answer (first 200 chars):"
jq -r '.message.content.final_answer.answer_type_other.res_1.answer.answer // "<missing>"' "$RESP_FILE" | head -c 200 >&2
echo "" >&2
log "son2 answer (first 200 chars):"
jq -r '.message.content.final_answer.answer_type_other.res_2.answer.answer // "<missing>"' "$RESP_FILE" | head -c 200 >&2
echo "" >&2

# === verify ===
ok=1
assert_literal_in_answer        "$RESP_FILE" "$SID" || ok=0
assert_session_id_in_son_input  "$RESP_FILE" "$SID" || ok=0

if [ "$ok" = "1" ]; then
  log "ALL ASSERTIONS PASSED. session_id=$SID, conversation=$CONV_ID"
  log "(artifacts kept on platform; run with --cleanup to remove)"
  exit 0
else
  fail "assertions failed; see warnings above. Response saved at $RESP_FILE"
fi
