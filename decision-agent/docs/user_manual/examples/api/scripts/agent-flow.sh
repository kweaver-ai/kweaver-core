#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"
require_jq

: "${KWEAVER_LLM_ID:?Please set KWEAVER_LLM_ID before running make flow.}"
: "${KWEAVER_LLM_NAME:?Please set KWEAVER_LLM_NAME before running make flow.}"

WORK_DIR="$(mktemp -d "${TMPDIR:-/tmp}/kweaver-api-example.XXXXXX")"
AGENT_ID=""

cleanup() {
  if [[ -n "$AGENT_ID" ]]; then
    af_curl -X PUT "$AF_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID/unpublish" >/dev/null 2>&1 || true
    af_curl -X DELETE "$AF_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID" >/dev/null 2>&1 || true
  fi
  rm -rf "$WORK_DIR"
}
trap cleanup EXIT

jq -n \
  --arg llm_id "$KWEAVER_LLM_ID" \
  --arg llm_name "$KWEAVER_LLM_NAME" \
  '{
    input: { fields: [{ name: "query", type: "string" }] },
    output: { default_format: "markdown", variables: { answer_var: "answer" } },
    llms: [{
      is_default: true,
      llm_config: {
        id: $llm_id,
        name: $llm_name,
        model_type: "llm",
        temperature: 1,
        top_p: 1,
        top_k: 1,
        frequency_penalty: 0,
        presence_penalty: 0,
        max_tokens: 1000,
        retrieval_max_tokens: 32
      }
    }],
    memory: { is_enabled: false },
    related_question: { is_enabled: false },
    plan_mode: { is_enabled: false }
  }' > "$WORK_DIR/agent-config.json"

AGENT_NAME="example_api_agent_$(date +%Y%m%d%H%M%S)"

jq -n \
  --arg name "$AGENT_NAME" \
  --argjson config "$(cat "$WORK_DIR/agent-config.json")" \
  '{
    name: $name,
    profile: "Created by docs/user_manual/examples/api",
    avatar_type: 1,
    avatar: "icon-dip-agent-default",
    product_key: "DIP",
    config: $config
  }' > "$WORK_DIR/create-agent.json"

af_curl -X POST "$AF_BASE_URL/api/agent-factory/v3/agent" \
  -H "Content-Type: application/json" \
  -d @"$WORK_DIR/create-agent.json" \
  | tee "$WORK_DIR/create-response.json" \
  | jq .

AGENT_ID="$(jq -r '.id // empty' "$WORK_DIR/create-response.json")"
test -n "$AGENT_ID"

af_curl "$AF_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID" | jq '{id, name, key, version}'

af_curl -X POST "$AF_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID/publish" \
  -H "Content-Type: application/json" \
  -d '{"category_ids":[],"description":"Published by docs example","publish_to_where":["square"],"publish_to_bes":["skill_agent"],"pms_control":null}' \
  | jq .

af_curl -X PUT "$AF_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID/unpublish"

af_curl -X DELETE "$AF_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID"
AGENT_ID=""
echo
echo "API flow finished."
