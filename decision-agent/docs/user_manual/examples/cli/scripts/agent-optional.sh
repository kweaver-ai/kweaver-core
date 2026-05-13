#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"
require_kweaver

ACTION="${1:-details}"

run_details() {
  if [[ -n "${AGENT_ID:-}" ]]; then
    kweaver agent get "$AGENT_ID" --pretty
  else
    echo "Skip agent get: set AGENT_ID."
  fi

  if [[ -n "${AGENT_KEY:-}" ]]; then
    kweaver agent get-by-key "$AGENT_KEY"
  else
    echo "Skip agent get-by-key: set AGENT_KEY."
  fi

  if [[ -n "${TEMPLATE_ID:-}" ]]; then
    kweaver agent template-get "$TEMPLATE_ID" --pretty
  else
    echo "Skip template-get: set TEMPLATE_ID."
  fi
}

run_chat() {
  if [[ -z "${AGENT_ID:-}" ]]; then
    echo "Skip chat examples: set AGENT_ID."
    return 0
  fi

  kweaver agent chat "$AGENT_ID" \
    --version "${AGENT_VERSION:-v0}" \
    --message "${CHAT_MESSAGE:-请用一句话介绍你自己}" \
    --no-stream \
    --verbose

  kweaver agent sessions "$AGENT_ID" --limit "${LIMIT:-10}" --pretty

  if [[ -n "${CONVERSATION_ID:-}" ]]; then
    kweaver agent history "$AGENT_ID" "$CONVERSATION_ID" --pretty
    if ! kweaver agent trace "$AGENT_ID" "$CONVERSATION_ID" --pretty; then
      echo "Skip trace output: current backend may not expose the trace route (or trace-ai service is unavailable) used by this CLI version."
    fi
  else
    echo "Skip history and trace: set CONVERSATION_ID."
  fi
}

run_skill() {
  if [[ -z "${AGENT_ID:-}" ]]; then
    echo "Skip skill examples: set AGENT_ID."
    return 0
  fi

  kweaver agent skill list "$AGENT_ID" --pretty

  if [[ -z "${SKILL_ID:-}" ]]; then
    echo "Skip skill add/remove: set SKILL_ID."
    return 0
  fi

  kweaver agent skill add "$AGENT_ID" "$SKILL_ID"
  kweaver agent skill remove "$AGENT_ID" "$SKILL_ID"
}

run_update_knowledge_network() {
  if [[ -z "${AGENT_ID:-}" ]]; then
    echo "Skip knowledge network update: set AGENT_ID."
    return 0
  fi
  if [[ -z "${KN_ID:-}" ]]; then
    echo "Skip knowledge network update: set KN_ID."
    return 0
  fi

  kweaver agent update "$AGENT_ID" --knowledge-network-id "$KN_ID" --pretty
}

case "$ACTION" in
  details) run_details ;;
  chat) run_chat ;;
  skill) run_skill ;;
  update-knowledge-network) run_update_knowledge_network ;;
  *)
    echo "Unknown action: $ACTION" >&2
    exit 2
    ;;
esac
