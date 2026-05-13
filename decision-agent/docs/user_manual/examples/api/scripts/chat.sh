#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

ACTION="${1:-all}"

if [[ -z "${AGENT_ID:-}" || -z "${AGENT_KEY:-}" ]]; then
  echo "Skip chat examples: set AGENT_ID and AGENT_KEY from a published Agent before running this target."
  exit 0
fi

AGENT_VERSION="${AGENT_VERSION:-v1}"

post_chat() {
  local name="$1"
  local path="$2"
  local body="$3"

  echo "== $name =="
  af_curl -X POST "$AF_BASE_URL$path" \
    -H "Content-Type: application/json" \
    -d "$body"
  echo
}

run_non_stream() {
  post_chat "non-stream chat" \
    "/api/agent-factory/v1/app/$AGENT_KEY/chat/completion" \
    '{"agent_id":"'"$AGENT_ID"'","agent_key":"'"$AGENT_KEY"'","agent_version":"'"$AGENT_VERSION"'","query":"请用一句话介绍你自己","stream":false}'
}

run_stream() {
  post_chat "stream chat" \
    "/api/agent-factory/v1/app/$AGENT_KEY/chat/completion" \
    '{"agent_id":"'"$AGENT_ID"'","agent_key":"'"$AGENT_KEY"'","agent_version":"'"$AGENT_VERSION"'","query":"请用一句话介绍你自己","stream":true,"inc_stream":false}'
}

run_incremental_stream() {
  post_chat "incremental stream chat" \
    "/api/agent-factory/v1/app/$AGENT_KEY/chat/completion" \
    '{"agent_id":"'"$AGENT_ID"'","agent_key":"'"$AGENT_KEY"'","agent_version":"'"$AGENT_VERSION"'","query":"请用一句话介绍你自己","stream":true,"inc_stream":true}'
}

run_debug() {
  post_chat "debug chat" \
    "/api/agent-factory/v1/app/$AGENT_KEY/debug/completion" \
    '{"agent_id":"'"$AGENT_ID"'","agent_version":"'"$AGENT_VERSION"'","input":{"query":"测试 debug completion","history":[],"custom_querys":{}},"stream":false}'
}

case "$ACTION" in
  non-stream) run_non_stream ;;
  stream) run_stream ;;
  incremental-stream) run_incremental_stream ;;
  debug) run_debug ;;
  all)
    run_non_stream
    run_stream
    run_incremental_stream
    run_debug
    ;;
  *)
    echo "Unknown action: $ACTION" >&2
    exit 2
    ;;
esac
