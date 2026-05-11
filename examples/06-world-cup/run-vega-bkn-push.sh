#!/usr/bin/env bash
# =============================================================================
# Vega path: catalog resources → map wc_* Vega UUIDs → render worldcup-bkn-vega
# → validate → push → bkn build --wait (index build).
#
# Prereqs: jq; Node kweaver CLI; Vega discover already produced table Resources;
#           .env with DB_* and VEGA_CATALOG_NAME or VEGA_CATALOG_ID (see env.sample).
# =============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
MAPPING_TMP="${SCRIPT_DIR}/.vega-bkn-mapping.json"

usage() {
    printf 'Usage: %s [options]\n' "$(basename "$0")"
    cat <<'EOF'
Options:
  -h, --help     Help.
  --dry-run      Only print resolved catalog id and resource count; write mapping JSON preview.

Environment:
  VEGA_CATALOG_ID     Preferred if set (skip name lookup).
  VEGA_CATALOG_NAME   Fallback (default worldcup_mysql_vega in env.sample).

  SKIP_BKN_PUSH=1       Stop after rendering + validate (no push / build).
  DO_BKN_BUILD=0       Vega resource 型 OT 默认无离线索索引任务；bkn push 已成功即完成建模。
                         设为 1 仍会尝试 build（若平台报错 “resource-backed … no indexable” 将提示并以 0 退出）。

  BKN_PUSH_TIMEOUT_BUILD=900   Seconds for optional `kweaver bkn build … --timeout`.
EOF
}

DRY_RUN=0
SKIP_BKN_PUSH="${SKIP_BKN_PUSH:-0}"
BUILD_TIMEOUT="${BKN_PUSH_TIMEOUT_BUILD:-900}"
DO_BKN_BUILD="${DO_BKN_BUILD:-0}"
while [ $# -gt 0 ]; do
    case "$1" in
        -h|--help) usage; exit 0 ;;
        --dry-run) DRY_RUN=1; shift ;;
        *) echo "Unknown: $1" >&2; usage >&2; exit 2 ;;
    esac
done

set -a
[ -f "$SCRIPT_DIR/.env" ] && source "$SCRIPT_DIR/.env"
set +a

if [ -n "${KWEAVER_BASE_URL:-}" ]; then
    KWEAV=(kweaver --base-url "${KWEAVER_BASE_URL}")
else
    KWEAV=(kweaver)
fi

require_jq() {
    command -v jq >/dev/null 2>&1 || {
        echo "Error: jq is required (brew install jq)" >&2
        exit 1
    }
}

# Last complete JSON blob in mixed CLI stdout (warnings may precede).
_extract_cli_json() {
    python3 "$SCRIPT_DIR/scripts/extract_trailing_json.py"
}

kn_id_from_rendered_network() {
    awk '/^id:/ {print $2; exit}' "$SCRIPT_DIR/.rendered-bkn-vega/network.bkn" 2>/dev/null || true
}

resolve_catalog_id() {
    if [ -n "${VEGA_CATALOG_ID:-}" ]; then
        printf '%s' "$VEGA_CATALOG_ID"
        return
    fi
    local name raw
    name="${VEGA_CATALOG_NAME:?Set VEGA_CATALOG_NAME or VEGA_CATALOG_ID in .env}"
    raw="$("${KWEAV[@]}" vega catalog list --limit 200 2>&1)"
    printf '%s' "$raw" | _extract_cli_json | jq -r --arg n "$name" '.entries[]? | select(.name == $n) | .id' | head -1
}

require_jq

CATALOG_ID="$(resolve_catalog_id)"
if [ -z "$CATALOG_ID" ]; then
    echo "Error: empty catalog id (set VEGA_CATALOG_ID or VEGA_CATALOG_NAME)." >&2
    exit 1
fi

echo "=== Catalog ===" >&2
echo "  id=$CATALOG_ID" >&2

echo "=== Fetching table resources (--limit 500) ===" >&2
RAW_RES="$("${KWEAV[@]}" vega catalog resources "$CATALOG_ID" --category table --limit 500 2>&1)"

if ! BODY="$(printf '%s' "$RAW_RES" | _extract_cli_json)"; then
    echo 'Error parsing vega catalog resources output.' >&2
    exit 1
fi

NRES="$(printf '%s' "$BODY" | jq '.entries | length')"
echo "  entries in response: $NRES" >&2

if ! printf '%s' "$BODY" | python3 "$SCRIPT_DIR/scripts/map_vega_table_resources.py" >"$MAPPING_TMP"; then
    echo 'Error: could not map all 27 wc_* stems to Vega resources (see stderr above).' >&2
    exit 1
fi

KEYS="$(jq 'length' "$MAPPING_TMP")"
echo "  mapped placeholders: $KEYS (need 27)" >&2

if [ "$DRY_RUN" = 1 ]; then
    echo "[dry-run] mapping written $MAPPING_TMP" >&2
    head -40 "$MAPPING_TMP" >&2
    exit 0
fi

if [ "${KEYS:-0}" -lt 27 ]; then
    echo "Error: incomplete mapping (<27). Run discover again or raise --limit." >&2
    exit 1
fi
echo "=== Render BKN (.rendered-bkn-vega) ===" >&2
python3 "$SCRIPT_DIR/scripts/render_worldcup_bkn_vega_placeholders.py" \
    --mapping "$MAPPING_TMP" \
    --src "$SCRIPT_DIR/worldcup-bkn-vega" \
    --dst "$SCRIPT_DIR/.rendered-bkn-vega"

RENDERED="$SCRIPT_DIR/.rendered-bkn-vega"
mkdir -p "$SCRIPT_DIR/.tmp"
export TMPDIR="${TMPDIR:-$SCRIPT_DIR/.tmp}"

echo "=== kweaver bkn validate ===" >&2
"${KWEAV[@]}" bkn validate "$RENDERED"

if [ "$SKIP_BKN_PUSH" = 1 ] || [ "$SKIP_BKN_PUSH" = true ]; then
    echo "SKIP_BKN_PUSH=1 — skipping push (and optional build)." >&2
    exit 0
fi

KN_ID="$(kn_id_from_rendered_network)"
[ -z "$KN_ID" ] && { echo "Error: missing id in $RENDERED/network.bkn"; exit 1; }

echo "=== kweaver bkn push → $KN_ID ===" >&2
"${KWEAV[@]}" bkn push "$RENDERED" --branch main

if [ "${DO_BKN_BUILD:-0}" = 1 ] || [ "${DO_BKN_BUILD:-0}" = true ]; then
    echo "=== kweaver bkn build (--wait --timeout ${BUILD_TIMEOUT}s) ===" >&2
    set +e
    BUILD_LOG="$("${KWEAV[@]}" bkn build "$KN_ID" --wait --timeout "$BUILD_TIMEOUT" 2>&1)"
    BUILD_RC=$?
    set -e
    printf '%s\n' "$BUILD_LOG"
    if [ "$BUILD_RC" -ne 0 ]; then
        case "$BUILD_LOG" in *NoneConceptType*|*"no indexable object types"*|*resource-backed*)
            echo "" >&2
            echo "Note: Vega resource-backed KN has no offline index job — push is enough." >&2
            echo "  (Platform: BknBackend.Job.NoneConceptType / indexable OT empty.)" >&2
            ;;
        *)
            exit "$BUILD_RC"
            ;;
        esac
    fi
else
    echo "=== SKIP bkn build (DO_BKN_BUILD=0): resource-backed OTs不在此建模离线索引 ===" >&2
fi

echo "" >&2
echo "Done. KN=$KN_ID  (mapping: $MAPPING_TMP, rendered: $RENDERED)" >&2
