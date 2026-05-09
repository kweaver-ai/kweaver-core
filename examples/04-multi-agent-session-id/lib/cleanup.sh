#!/usr/bin/env bash
# Demo teardown: unpublish + delete the three agents and the skill.
# Only invoked when the user runs `./run.sh --cleanup`. Default run.sh
# behavior never calls this.

source "$(dirname "${BASH_SOURCE[0]}")/common.sh"

cleanup_all() {
  for name in "$EXP_FATHER_NAME" "$EXP_SON1_NAME" "$EXP_SON2_NAME"; do
    local id
    id=$(resolve_agent_id "$name")
    if [ -n "$id" ]; then
      log "unpublish + delete agent $name ($id)"
      kweaver agent unpublish "$id" 2>&1 | tail -1 >&2 || true
      kweaver agent delete "$id" -y 2>&1 | tail -1 >&2 || warn "delete agent $id failed"
    else
      log "agent $name not present, skip"
    fi
  done

  local sid
  sid=$(resolve_skill_id "$EXP_SKILL_NAME")
  if [ -n "$sid" ]; then
    log "delete skill $EXP_SKILL_NAME ($sid)"
    kweaver skill delete "$sid" -y 2>&1 | tail -1 >&2 || warn "delete skill $sid failed"
  else
    log "skill $EXP_SKILL_NAME not present, skip"
  fi
}
