#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"
require_jq

if [[ -z "${AGENT_KEY:-}" ]]; then
  echo "Skip conversation examples: set AGENT_KEY before running this target."
  exit 0
fi

echo "== conversation list =="
af_curl "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/conversation?page=1&size=${LIMIT:-10}" | jq .

if [[ -z "${CONVERSATION_ID:-}" ]]; then
  echo "Skip conversation detail and mark_read examples: set CONVERSATION_ID."
  exit 0
fi

echo "== conversation detail =="
af_curl "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/conversation/$CONVERSATION_ID" | jq .

echo "== mark read =="
af_curl -X PUT "$AF_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/conversation/$CONVERSATION_ID/mark_read" \
  -H "Content-Type: application/json" \
  -d '{"latest_read_index":'"${LATEST_READ_INDEX:-2}"'}' | jq .
