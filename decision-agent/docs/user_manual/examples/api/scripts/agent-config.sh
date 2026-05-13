#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
EXAMPLE_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
source "$SCRIPT_DIR/common.sh"
require_jq

OUT_DIR="$EXAMPLE_DIR/.tmp"
mkdir -p "$OUT_DIR"

MINIMAL_CONFIG="$OUT_DIR/agent-config-minimal.json"
ADVANCED_CONFIG="$OUT_DIR/agent-config-advanced.json"

jq -n \
  --arg llm_id "${KWEAVER_LLM_ID:-<llm-id>}" \
  --arg llm_name "${KWEAVER_LLM_NAME:-<llm-name>}" \
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
        max_tokens: 1000,
        retrieval_max_tokens: 32
      }
    }]
  }' > "$MINIMAL_CONFIG"

jq -n \
  --arg llm_id "${KWEAVER_LLM_ID:-<llm-id>}" \
  --arg llm_name "${KWEAVER_LLM_NAME:-<llm-name>}" \
  --arg kn_id "${KN_ID:-<knowledge-network-id>}" \
  '{
    input: {
      fields: [
        { name: "query", type: "string" },
        { name: "history", type: "array" }
      ]
    },
    output: { default_format: "markdown", variables: { answer_var: "answer" } },
    llms: [{
      is_default: true,
      llm_config: {
        id: $llm_id,
        name: $llm_name,
        model_type: "llm",
        temperature: 0.7,
        top_p: 0.9,
        top_k: 40,
        max_tokens: 2000,
        retrieval_max_tokens: 128
      }
    }],
    data_source: {
      knowledge_network: [
        { knowledge_network_id: $kn_id }
      ]
    },
    memory: { is_enabled: false },
    related_question: { is_enabled: false },
    plan_mode: { is_enabled: false }
  }' > "$ADVANCED_CONFIG"

jq . "$MINIMAL_CONFIG" >/dev/null
jq . "$ADVANCED_CONFIG" >/dev/null

echo "Generated:"
echo "  .tmp/agent-config-minimal.json"
echo "  .tmp/agent-config-advanced.json"
