#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
EXAMPLE_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
source "$SCRIPT_DIR/common.sh"
require_jq

if [[ -z "${AGENT_ID:-}" ]]; then
  echo "Skip import/export example: set AGENT_ID before running this target."
  exit 0
fi

OUT_DIR="$EXAMPLE_DIR/.tmp"
mkdir -p "$OUT_DIR"

jq -n --arg agent_id "$AGENT_ID" '{agent_ids: [$agent_id]}' > "$OUT_DIR/agent-export.json"

af_curl -X POST "$AF_BASE_URL/api/agent-factory/v3/agent-inout/export" \
  -H "Content-Type: application/json" \
  -d @"$OUT_DIR/agent-export.json" \
  -o "$OUT_DIR/agent-export-result.json"

echo "Exported to .tmp/agent-export-result.json"

af_curl -X POST "$AF_BASE_URL/api/agent-factory/v3/agent-inout/import" \
  -F "import_type=${IMPORT_TYPE:-upsert}" \
  -F "file=@$OUT_DIR/agent-export-result.json;type=application/json" \
  -o "$OUT_DIR/agent-import-result.json"

jq . "$OUT_DIR/agent-import-result.json"

IMPORTED_AGENT_ID="$(jq -r '.id // .data.id // .agent_id // .data.agent_id // empty' "$OUT_DIR/agent-import-result.json")"
if [[ "${CLEANUP_IMPORTED_AGENT:-1}" == "1" && -n "$IMPORTED_AGENT_ID" && "$IMPORTED_AGENT_ID" != "$AGENT_ID" ]]; then
  echo "Cleaning imported agent: $IMPORTED_AGENT_ID"
  af_curl -X DELETE "$AF_BASE_URL/api/agent-factory/v3/agent/$IMPORTED_AGENT_ID" | jq .
fi
