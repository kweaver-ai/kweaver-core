#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"
require_jq

LIMIT="${LIMIT:-1}"

jq -n \
  --argjson limit "$LIMIT" \
  '{
    offset: 0,
    limit: $limit,
    category_id: "",
    name: "",
    custom_space_id: "",
    is_to_square: 1
  }' \
  | af_curl -X POST "$AF_BASE_URL/api/agent-factory/v3/published/agent" \
      -H "Content-Type: text/plain;charset=UTF-8" \
      --data-binary @- \
  | jq .

