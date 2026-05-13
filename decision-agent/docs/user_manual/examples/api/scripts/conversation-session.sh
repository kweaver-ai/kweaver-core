#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"
require_jq

if [[ -z "${CONVERSATION_ID:-}" ]]; then
  echo "Skip conversation session example: set CONVERSATION_ID before running this target."
  exit 0
fi

if [[ -z "${AGENT_ID:-}" ]]; then
  echo "Skip conversation session example: set AGENT_ID before running this target."
  exit 0
fi

jq -n \
  --arg action "${SESSION_ACTION:-get_info_or_create}" \
  --arg agent_id "$AGENT_ID" \
  --arg agent_version "${AGENT_VERSION:-v1}" \
  '{
    action: $action,
    agent_id: $agent_id,
    agent_version: $agent_version
  }' \
  | af_curl -X PUT "$AF_BASE_URL/api/agent-factory/v1/conversation/session/$CONVERSATION_ID" \
      -H "Content-Type: application/json" \
      --data-binary @- \
  | jq .
