#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"
require_kweaver
require_jq

: "${KWEAVER_LLM_ID:?Please set KWEAVER_LLM_ID before running make flow.}"
: "${KWEAVER_LLM_NAME:?Please set KWEAVER_LLM_NAME before running make flow.}"

WORK_DIR="$(mktemp -d "${TMPDIR:-/tmp}/kweaver-cli-example.XXXXXX")"
AGENT_ID=""

cleanup() {
  if [[ -n "$AGENT_ID" ]]; then
    kweaver agent unpublish "$AGENT_ID" >/dev/null 2>&1 || true
    kweaver agent delete "$AGENT_ID" -y >/dev/null 2>&1 || true
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

AGENT_NAME="example_cli_agent_$(date +%Y%m%d%H%M%S)"

kweaver agent create \
  --name "$AGENT_NAME" \
  --profile "Created by docs/user_manual/examples/cli" \
  --product-key dip \
  --config "$WORK_DIR/agent-config.json" \
  --pretty \
  | tee "$WORK_DIR/create-response.json"

AGENT_ID="$(jq -r '.id // empty' "$WORK_DIR/create-response.json")"
test -n "$AGENT_ID"

kweaver agent get "$AGENT_ID" --pretty
kweaver agent publish "$AGENT_ID"
kweaver agent unpublish "$AGENT_ID"
kweaver agent delete "$AGENT_ID" -y
AGENT_ID=""
echo "CLI flow finished."
