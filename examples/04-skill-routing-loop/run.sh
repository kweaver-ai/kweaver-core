#!/usr/bin/env bash
# =============================================================================
# 04-skill-routing-loop: KN-driven Skill governance end-to-end
#
# Flow: business DB → Vega → BKN → context-loader find_skills →
#       Decision Agent → Skill execute → mock business backend → audit log
# =============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TIMESTAMP=$(date +%s)

# ── CLI flags ────────────────────────────────────────────────────────────────
BONUS=0
usage() {
    cat <<USAGE
Usage: $(basename "$0") [options]

Options:
  --bonus      Run the Bonus segment after main flow (changes SUP-2 capability)
  -h, --help   Show this help
USAGE
}
while [ $# -gt 0 ]; do
    case "$1" in
        --bonus) BONUS=1 ;;
        -h|--help) usage; exit 0 ;;
        *) echo "Unknown argument: $1" >&2; usage >&2; exit 2 ;;
    esac
    shift
done

# ── Debug helper ─────────────────────────────────────────────────────────────
DEBUG="${DEBUG:-0}"
debug() { [ "$DEBUG" = "1" ] && echo "[debug] $*" >&2 || true; }

# ── Step 0: Load .env ────────────────────────────────────────────────────────
[ -f "$SCRIPT_DIR/.env" ] && source "$SCRIPT_DIR/.env"

PLATFORM_HOST="${PLATFORM_HOST:?Set PLATFORM_HOST in .env}"
LLM_ID="${LLM_ID:?Set LLM_ID in .env (use: kweaver call /api/mf-model-manager/v1/llm/list)}"
LLM_NAME="${LLM_NAME:-deepseek-v3.2}"
DB_HOST="${DB_HOST:?Set DB_HOST in .env}"
DB_PORT="${DB_PORT:-3306}"
DB_NAME="${DB_NAME:?Set DB_NAME in .env}"
DB_USER="${DB_USER:?Set DB_USER in .env}"
DB_PASS="${DB_PASS:?Set DB_PASS in .env}"
TOOL_BACKEND_PORT="${TOOL_BACKEND_PORT:-8765}"

DS_NAME="ex04_ds_${TIMESTAMP}"
KN_ID="ex04_skill_routing"   # fixed, must match network.bkn frontmatter
TABLE_PREFIX="ex04_${TIMESTAMP}_"

# Track resources for cleanup
DS_ID="" TMP_KN_ID="" MCP_ID="" AGENT_ID=""
SKILL_IDS=()
TOOL_BACKEND_PID=""

cleanup() {
    echo ""
    echo "=== Cleanup ==="
    [ -n "$AGENT_ID" ] && {
        kweaver agent unpublish "$AGENT_ID" 2>/dev/null || true
        kweaver agent delete "$AGENT_ID" -y 2>/dev/null && echo "  ✓ agent $AGENT_ID"
    }
    [ -n "$MCP_ID" ] && {
        kweaver call "/api/agent-operator-integration/v1/mcp/$MCP_ID/status" -X POST \
            -H "x-business-domain: bd_public" -d '{"status":"offline"}' >/dev/null 2>&1 || true
        kweaver call "/api/agent-operator-integration/v1/mcp/$MCP_ID" -X DELETE \
            -H "x-business-domain: bd_public" >/dev/null 2>&1 && echo "  ✓ mcp $MCP_ID"
    }
    for sid in "${SKILL_IDS[@]:-}"; do
        [ -z "$sid" ] && continue
        kweaver skill status "$sid" offline >/dev/null 2>&1 || true
        echo y | kweaver skill delete "$sid" >/dev/null 2>&1 && echo "  ✓ skill $sid"
    done
    kweaver bkn delete "$KN_ID" -y >/dev/null 2>&1 && echo "  ✓ kn $KN_ID" || true
    [ -n "$TMP_KN_ID" ] && kweaver bkn delete "$TMP_KN_ID" -y >/dev/null 2>&1 && echo "  ✓ tmp kn $TMP_KN_ID" || true
    [ -n "$DS_ID" ] && kweaver ds delete "$DS_ID" -y >/dev/null 2>&1 && echo "  ✓ ds $DS_ID"
    [ -n "$TOOL_BACKEND_PID" ] && kill "$TOOL_BACKEND_PID" 2>/dev/null && echo "  ✓ mock backend pid $TOOL_BACKEND_PID"
}
trap cleanup EXIT

# ── Step 1: Connect MySQL datasource ─────────────────────────────────────────
echo "=== Step 1: Connect MySQL datasource ==="
DS_RAW=$(kweaver ds connect mysql "$DB_HOST" "$DB_PORT" "$DB_NAME" \
    --account "$DB_USER" --password "$DB_PASS" --name "$DS_NAME" 2>&1)
DS_ID=$(echo "$DS_RAW" | python3 -c "
import sys, json
raw = sys.stdin.read()
def find_objs(s):
    depth = 0; start = -1
    for i, ch in enumerate(s):
        if ch == '{':
            if depth == 0: start = i
            depth += 1
        elif ch == '}':
            depth -= 1
            if depth == 0 and start >= 0:
                yield s[start:i+1]; start = -1
for chunk in find_objs(raw):
    try: obj = json.loads(chunk)
    except Exception: continue
    if isinstance(obj, dict) and 'datasource_id' in obj:
        print(obj['datasource_id']); break
")
[ -z "$DS_ID" ] && { echo "ERROR: ds connect failed" >&2; exit 1; }
echo "  Datasource: $DS_ID"

# ── Step 2: Import CSVs (creates dataviews and a temp KN we discard) ─────────
echo ""
echo "=== Step 2: Import CSVs and provision Vega dataviews ==="
TMP_KN_RAW=$(kweaver bkn create-from-csv "$DS_ID" \
    --files "$SCRIPT_DIR/data/*.csv" \
    --name "ex04_tmp_${TIMESTAMP}" \
    --table-prefix "$TABLE_PREFIX" \
    --build --timeout 180 2>&1)
TMP_KN_ID=$(echo "$TMP_KN_RAW" | python3 -c "
import sys, json, re
raw = sys.stdin.read()
# Find balanced JSON objects; pick the last one containing kn_id at top level
def find_objs(s):
    depth = 0
    start = -1
    for i, ch in enumerate(s):
        if ch == '{':
            if depth == 0:
                start = i
            depth += 1
        elif ch == '}':
            depth -= 1
            if depth == 0 and start >= 0:
                yield s[start:i+1]
                start = -1
for chunk in reversed(list(find_objs(raw))):
    try:
        obj = json.loads(chunk)
    except Exception:
        continue
    if isinstance(obj, dict) and 'kn_id' in obj:
        print(obj['kn_id'])
        break
")
[ -z "$TMP_KN_ID" ] && { echo "ERROR: failed to parse kn_id from create-from-csv" >&2; exit 1; }
echo "  Tmp KN (will discard, keep dataviews): $TMP_KN_ID"

# Look up dataview IDs by name
DV_LIST=$(kweaver dataview list --datasource-id "$DS_ID" --limit 50 2>&1)
get_dv_id() {
    local table="$1"
    echo "$DV_LIST" | python3 -c "
import sys, json
raw = sys.stdin.read()
i = raw.find('[')
for v in json.loads(raw[i:]):
    if v.get('name') == '$table':
        print(v.get('id'))
        break
"
}
MATERIALS_DV_ID=$(get_dv_id "${TABLE_PREFIX}materials")
SUPPLIERS_DV_ID=$(get_dv_id "${TABLE_PREFIX}suppliers")
SKILLS_DV_ID=$(get_dv_id "${TABLE_PREFIX}skills")
echo "  Dataview IDs: materials=$MATERIALS_DV_ID, suppliers=$SUPPLIERS_DV_ID, skills=$SKILLS_DV_ID"

# Discard the auto-generated KN; we'll push our own with controlled OT IDs
kweaver bkn delete "$TMP_KN_ID" -y >/dev/null
TMP_KN_ID=""

# ── Step 3: Render BKN templates with dataview IDs ───────────────────────────
echo ""
echo "=== Step 3: Render BKN templates with dataview IDs ==="
RENDERED_BKN="$SCRIPT_DIR/.rendered-bkn"
rm -rf "$RENDERED_BKN"
cp -r "$SCRIPT_DIR/bkn" "$RENDERED_BKN"
sed -i.bak \
    -e "s|{{MATERIALS_DV_ID}}|$MATERIALS_DV_ID|" \
    -e "s|{{MATERIALS_DV_NAME}}|${TABLE_PREFIX}materials|" \
    "$RENDERED_BKN/object_types/material.bkn"
sed -i.bak \
    -e "s|{{SUPPLIERS_DV_ID}}|$SUPPLIERS_DV_ID|" \
    -e "s|{{SUPPLIERS_DV_NAME}}|${TABLE_PREFIX}suppliers|" \
    "$RENDERED_BKN/object_types/supplier.bkn"
sed -i.bak \
    -e "s|{{SKILLS_DV_ID}}|$SKILLS_DV_ID|" \
    -e "s|{{SKILLS_DV_NAME}}|${TABLE_PREFIX}skills|" \
    "$RENDERED_BKN/object_types/skills.bkn"
find "$RENDERED_BKN" -name '*.bak' -delete
echo "  ✓ rendered .bkn files"

# ── Step 4: Push BKN ─────────────────────────────────────────────────────────
echo ""
echo "=== Step 4: bkn push (deploy schema + relations) ==="
kweaver bkn validate "$RENDERED_BKN" 2>&1 | tail -1
PUSH_RAW=$(kweaver bkn push "$RENDERED_BKN" 2>&1)
echo "$PUSH_RAW" | tail -3
# kn_id is fixed (network.bkn frontmatter id) — just confirm push succeeded
echo "$PUSH_RAW" | grep -q "\"kn_id\"" || { echo "ERROR: bkn push failed" >&2; exit 1; }
echo "  ✓ KN: $KN_ID"

# ── Step 5: Build KN ─────────────────────────────────────────────────────────
echo ""
echo "=== Step 5: Build KN (sync) ==="
kweaver bkn build "$KN_ID" --wait --timeout 60 2>&1 | tail -2

