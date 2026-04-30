#!/usr/bin/env bash
# =============================================================================
# 05-skill-routing-loop: KN-driven Skill governance end-to-end
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
  --bonus      Run the Bonus segment after main flow (re-binds MAT-002 in the business system)
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
TOOL_BACKEND_PUBLIC_URL="${TOOL_BACKEND_PUBLIC_URL:-http://127.0.0.1:$TOOL_BACKEND_PORT}"

DS_NAME="ex05_ds_${TIMESTAMP}"
KN_ID="ex05_skill_routing"   # fixed, must match network.bkn frontmatter
TABLE_PREFIX="ex05_${TIMESTAMP}_"

# Track resources for cleanup
DS_ID="" TMP_KN_ID="" MCP_ID="" AGENT_ID=""
SKILL_IDS=()
TOOL_BACKEND_PID=""
STANDARD_REPLENISH_ID=""
SUBSTITUTE_SWAP_ID=""
SUPPLIER_EXPEDITE_ID=""
RENDERED_SKILLS=""

cleanup() {
    echo ""
    echo "=== Cleanup ==="
    [ -n "$AGENT_ID" ] && {
        kweaver agent unpublish "$AGENT_ID" 2>/dev/null || true
        kweaver agent delete "$AGENT_ID" -y 2>/dev/null && echo "  ✓ agent $AGENT_ID" || true
    }
    [ -n "$MCP_ID" ] && {
        kweaver call "/api/agent-operator-integration/v1/mcp/$MCP_ID/status" -X POST \
            -H "x-business-domain: bd_public" -d '{"status":"offline"}' >/dev/null 2>&1 || true
        kweaver call "/api/agent-operator-integration/v1/mcp/$MCP_ID" -X DELETE \
            -H "x-business-domain: bd_public" >/dev/null 2>&1 && echo "  ✓ mcp $MCP_ID" || true
    }
    for sid in "${SKILL_IDS[@]:-}"; do
        [ -z "$sid" ] && continue
        kweaver skill status "$sid" offline >/dev/null 2>&1 || true
        echo y | kweaver skill delete "$sid" >/dev/null 2>&1 && echo "  ✓ skill $sid" || true
    done
    kweaver bkn delete "$KN_ID" -y >/dev/null 2>&1 && echo "  ✓ kn $KN_ID" || true
    [ -n "$TMP_KN_ID" ] && kweaver bkn delete "$TMP_KN_ID" -y >/dev/null 2>&1 && echo "  ✓ tmp kn $TMP_KN_ID" || true
    [ -n "$DS_ID" ] && kweaver ds delete "$DS_ID" -y >/dev/null 2>&1 && echo "  ✓ ds $DS_ID" || true
    [ -n "$TOOL_BACKEND_PID" ] && kill "$TOOL_BACKEND_PID" 2>/dev/null && echo "  ✓ mock backend pid $TOOL_BACKEND_PID" || true
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

# ── Step 2: Register Skill packages first; BKN must store real skill IDs ─────
echo ""
echo "=== Step 2: Register Skill packages ==="
RENDERED_SKILLS="$SCRIPT_DIR/.rendered-skills"
rm -rf "$RENDERED_SKILLS"
mkdir -p "$RENDERED_SKILLS"
for skill_dir in "$SCRIPT_DIR"/skills/*/; do
    skill_name=$(basename "$skill_dir")
    rendered_skill_dir="$RENDERED_SKILLS/$skill_name"
    cp -R "$skill_dir" "$rendered_skill_dir"
    find "$rendered_skill_dir" -type f -name 'SKILL.md' -exec sed -i.bak \
        -e "s|{{TOOL_BACKEND_PUBLIC_URL}}|$TOOL_BACKEND_PUBLIC_URL|g" {} \;
    find "$rendered_skill_dir" -name '*.bak' -delete
    zip_path="$SCRIPT_DIR/.${skill_name}.zip"
    rm -f "$zip_path"
    (cd "$rendered_skill_dir" && zip -qr "$zip_path" .)
    REG_RAW=$(kweaver skill register --zip-file "$zip_path" 2>&1)
    sid=$(echo "$REG_RAW" | python3 -c "
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
    if isinstance(obj, dict) and 'id' in obj:
        print(obj['id']); break
")
    [ -z "$sid" ] && { echo "ERROR: skill register failed for $skill_name" >&2; echo "$REG_RAW" >&2; exit 1; }
    kweaver skill status "$sid" published >/dev/null
    SKILL_IDS+=("$sid")
    case "$skill_name" in
        standard_replenish) STANDARD_REPLENISH_ID="$sid" ;;
        substitute_swap) SUBSTITUTE_SWAP_ID="$sid" ;;
        supplier_expedite) SUPPLIER_EXPEDITE_ID="$sid" ;;
    esac
    echo "  ✓ $skill_name → $sid (published)"
    rm -f "$zip_path"
done
[ -n "$STANDARD_REPLENISH_ID" ] && [ -n "$SUBSTITUTE_SWAP_ID" ] && [ -n "$SUPPLIER_EXPEDITE_ID" ] \
    || { echo "ERROR: not all skill IDs were registered" >&2; exit 1; }

# ── Step 3: Render CSVs with real execution-factory skill IDs ────────────────
echo ""
echo "=== Step 3: Render CSVs with registered skill IDs ==="
RENDERED_DATA="$SCRIPT_DIR/.rendered-data"
rm -rf "$RENDERED_DATA"
mkdir -p "$RENDERED_DATA"
cp "$SCRIPT_DIR/data/suppliers.csv" "$RENDERED_DATA/suppliers.csv"
STANDARD_REPLENISH_ID="$STANDARD_REPLENISH_ID" \
SUBSTITUTE_SWAP_ID="$SUBSTITUTE_SWAP_ID" \
SUPPLIER_EXPEDITE_ID="$SUPPLIER_EXPEDITE_ID" \
SCRIPT_DIR="$SCRIPT_DIR" RENDERED_DATA="$RENDERED_DATA" python3 - <<'PY'
import csv
import os
from pathlib import Path

script_dir = Path(os.environ["SCRIPT_DIR"])
rendered = Path(os.environ["RENDERED_DATA"])
ids = {
    "standard_replenish": os.environ["STANDARD_REPLENISH_ID"],
    "substitute_swap": os.environ["SUBSTITUTE_SWAP_ID"],
    "supplier_expedite": os.environ["SUPPLIER_EXPEDITE_ID"],
}

with (script_dir / "data" / "materials.csv").open(newline="") as src, (rendered / "materials.csv").open("w", newline="") as dst:
    reader = csv.DictReader(src)
    writer = csv.DictWriter(dst, fieldnames=reader.fieldnames)
    writer.writeheader()
    for row in reader:
        if row["bound_skill_id"]:
            row["bound_skill_id"] = ids[row["bound_skill_id"]]
        writer.writerow(row)

with (script_dir / "data" / "skills.csv").open(newline="") as src, (rendered / "skills.csv").open("w", newline="") as dst:
    reader = csv.DictReader(src)
    writer = csv.DictWriter(dst, fieldnames=reader.fieldnames)
    writer.writeheader()
    for row in reader:
        row["description"] = f"{row['skill_id']} | {row['description']}"
        row["skill_id"] = ids[row["skill_id"]]
        writer.writerow(row)
PY
echo "  ✓ rendered data in $RENDERED_DATA"

# ── Step 4: Import CSVs (creates dataviews and a temp KN we discard) ─────────
echo ""
echo "=== Step 4: Import CSVs and provision Vega dataviews ==="
TMP_KN_RAW=$(kweaver bkn create-from-csv "$DS_ID" \
    --files "$RENDERED_DATA/*.csv" \
    --name "ex05_tmp_${TIMESTAMP}" \
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

# ── Step 5: Render BKN templates with dataview IDs ───────────────────────────
echo ""
echo "=== Step 5: Render BKN templates with dataview IDs ==="
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

# ── Step 6: Push BKN ─────────────────────────────────────────────────────────
echo ""
echo "=== Step 6: bkn push (deploy schema + relations) ==="
kweaver bkn validate "$RENDERED_BKN" 2>&1 | tail -1
PUSH_RAW=$(kweaver bkn push "$RENDERED_BKN" 2>&1)
echo "$PUSH_RAW" | tail -3
# kn_id is fixed (network.bkn frontmatter id) — just confirm push succeeded
echo "$PUSH_RAW" | grep -q "\"kn_id\"" || { echo "ERROR: bkn push failed" >&2; exit 1; }
echo "  ✓ KN: $KN_ID"

# ── Step 7: Build KN ─────────────────────────────────────────────────────────
echo ""
echo "=== Step 7: Build KN (sync) ==="
kweaver bkn build "$KN_ID" --wait --timeout 60 2>&1 | tail -2

# ── Step 8: Start mock business backend ──────────────────────────────────────
echo ""
echo "=== Step 8: Start mock business backend (port $TOOL_BACKEND_PORT) ==="
TOOL_BACKEND_URL="http://127.0.0.1:$TOOL_BACKEND_PORT"
DB_HOST="$DB_HOST" DB_PORT="$DB_PORT" DB_NAME="$DB_NAME" \
DB_USER="$DB_USER" DB_PASS="$DB_PASS" \
TOOL_BACKEND_PORT="$TOOL_BACKEND_PORT" \
MATERIALS_TABLE="${TABLE_PREFIX}materials" \
STANDARD_REPLENISH_ID="$STANDARD_REPLENISH_ID" \
SUBSTITUTE_SWAP_ID="$SUBSTITUTE_SWAP_ID" \
SUPPLIER_EXPEDITE_ID="$SUPPLIER_EXPEDITE_ID" \
python3 "$SCRIPT_DIR/tool_backend/server.py" >"$SCRIPT_DIR/.tool_backend.log" 2>&1 &
TOOL_BACKEND_PID=$!
for i in $(seq 1 15); do
    curl -sf "$TOOL_BACKEND_URL/healthz" >/dev/null 2>&1 && break
    sleep 1
done
curl -sf "$TOOL_BACKEND_URL/healthz" | grep -qE '"status"[[:space:]]*:[[:space:]]*"ok"' \
    || { echo "ERROR: mock backend failed; see .tool_backend.log" >&2; exit 1; }
echo "  ✓ mock backend pid $TOOL_BACKEND_PID"

# ── Step 9: Register context-loader MCP server (with X-Kn-ID header) ─────────
echo ""
echo "=== Step 9: Register context-loader MCP server ==="
MCP_REG_BODY=$(python3 -c "
import json
print(json.dumps({
    'mode': 'stream',
    'url': '$PLATFORM_HOST/api/agent-retrieval/v1/mcp',
    'name': 'ex05_ctx_loader_${TIMESTAMP}',
    'description': 'context-loader MCP for find_skills',
    'creation_type': 'custom',
    'headers': {'X-Kn-ID': '$KN_ID'},
}))
")
MCP_RAW=$(kweaver call /api/agent-operator-integration/v1/mcp/ -X POST \
    -H "Content-Type: application/json" \
    -H "x-business-domain: bd_public" \
    -d "$MCP_REG_BODY" 2>&1)
MCP_ID=$(echo "$MCP_RAW" | python3 -c "
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
    if isinstance(obj, dict) and 'mcp_id' in obj:
        print(obj['mcp_id']); break
")
[ -z "$MCP_ID" ] && { echo "ERROR: MCP register failed" >&2; echo "$MCP_RAW" >&2; exit 1; }
kweaver call "/api/agent-operator-integration/v1/mcp/$MCP_ID/status" -X POST \
    -H "x-business-domain: bd_public" \
    -d '{"status":"published"}' >/dev/null
echo "  ✓ MCP $MCP_ID (published, X-Kn-ID=$KN_ID)"

# ── Step 10: Render agent.json with MCP_ID + LLM_ID + Skill IDs ──────────────
echo ""
echo "=== Step 10: Render agent.json ==="
RENDERED_AGENT="$SCRIPT_DIR/.rendered-agent.json"
sed \
    -e "s|{{MCP_ID}}|$MCP_ID|" \
    -e "s|{{STANDARD_REPLENISH_SKILL_ID}}|$STANDARD_REPLENISH_ID|" \
    -e "s|{{SUBSTITUTE_SWAP_SKILL_ID}}|$SUBSTITUTE_SWAP_ID|" \
    -e "s|{{SUPPLIER_EXPEDITE_SKILL_ID}}|$SUPPLIER_EXPEDITE_ID|" \
    -e "s|{{LLM_ID}}|$LLM_ID|" \
    -e "s|{{LLM_NAME}}|$LLM_NAME|" \
    "$SCRIPT_DIR/agent.json" > "$RENDERED_AGENT"
python3 -c "import json; json.load(open('$RENDERED_AGENT'))" >/dev/null
echo "  ✓ agent.json rendered"

# ── Step 11: Create + publish agent ──────────────────────────────────────────
echo ""
echo "=== Step 11: Create + publish Decision Agent ==="
AGENT_NAME="ex05_skill_routing_${TIMESTAMP}"
CREATE_RAW=$(kweaver agent create \
    --name "$AGENT_NAME" \
    --profile "Example 05 — KN-driven skill routing" \
    --config "$RENDERED_AGENT" 2>&1)
AGENT_ID=$(echo "$CREATE_RAW" | python3 -c "
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
    if isinstance(obj, dict) and 'id' in obj:
        print(obj['id']); break
")
[ -z "$AGENT_ID" ] && { echo "ERROR: agent create failed" >&2; echo "$CREATE_RAW" >&2; exit 1; }
kweaver agent publish "$AGENT_ID" >/dev/null
echo "  ✓ agent $AGENT_ID (published)"

# ── Step 12: Trigger 3 critical-stock alerts; verify route + mock action ─────
echo ""
echo "=== Step 12: Trigger 3 alerts (one per material) ==="
if [ "${DEBUG_KEEP:-0}" = "1" ]; then
    trap - EXIT
    set +e
    echo "[DEBUG_KEEP=1] cleanup disabled; agent/kn/skill/ds will persist for debugging"
    echo "  AGENT_ID=$AGENT_ID"
    echo "  KN_ID=$KN_ID"
    echo "  DS_ID=$DS_ID"
    echo "  MCP_ID=$MCP_ID"
fi
execute_mock_action() {
    local sku="$1"
    local skill="$2"
    case "$skill" in
        substitute_swap)
            TOOL_BACKEND_URL="$TOOL_BACKEND_URL" \
            CANDIDATES='[{"sku":"SUB-001A","stock":200,"compat_score":0.95,"cost_delta_pct":5,"lead_time_hours":2},{"sku":"SUB-001B","stock":80,"compat_score":0.85,"cost_delta_pct":2,"lead_time_hours":4}]' \
            python3 "$SCRIPT_DIR/skills/substitute_swap/pick_substitute.py" --sku "$sku" >/dev/null
            ;;
        supplier_expedite)
            curl -sf -X POST "$TOOL_BACKEND_URL/supplier/expedite" \
                -H "Content-Type: application/json" \
                -d "{\"sku\":\"$sku\",\"supplier_id\":\"SUP-2\",\"sla_hours\":36}" >/dev/null
            ;;
        standard_replenish)
            curl -sf -X POST "$TOOL_BACKEND_URL/procurement/order" \
                -H "Content-Type: application/json" \
                -d "{\"sku\":\"$sku\",\"qty\":65}" >/dev/null
            ;;
    esac
}

assert_selected() {
    local output_file="$1"
    local expected="$2"
    grep -q "$expected" "$output_file" || {
        echo "ERROR: expected agent output to mention $expected" >&2
        echo "See $output_file" >&2
        exit 1
    }
}

for item in "MAT-001:substitute_swap" "MAT-002:supplier_expedite" "MAT-003:standard_replenish"; do
    sku="${item%%:*}"
    expected_skill="${item##*:}"
    echo ""
    echo "--- $sku ---"
    out_file="$SCRIPT_DIR/.chat-$sku.log"
    kweaver agent chat "$AGENT_ID" \
        -m "Material $sku hit critical stock level. Use find_skills to identify applicable skills, query the BKN for evidence (supplier capability, etc.), pick the best skill, execute it when possible, and end with SELECTED_SKILL_NAME=<name>." \
        --stream 2>&1 \
        | sed '/^(node:.*Warning:/d; /trace-warnings/d; /To continue this conversation/,$d' \
        | tee "$out_file" \
        | tail -40
    assert_selected "$out_file" "$expected_skill"
    execute_mock_action "$sku" "$expected_skill"
    echo "  ✓ verified route=$expected_skill and mock action"
done

# Verify that the demonstrable business actions reached the mock backend.
grep -q "\[mes/swap\]" "$SCRIPT_DIR/.tool_backend.log" \
    || { echo "ERROR: /mes/swap was not called" >&2; exit 1; }
grep -q "\[supplier/expedite\]" "$SCRIPT_DIR/.tool_backend.log" \
    || { echo "ERROR: /supplier/expedite was not called" >&2; exit 1; }
grep -q "\[procurement\]" "$SCRIPT_DIR/.tool_backend.log" \
    || { echo "ERROR: /procurement/order was not called" >&2; exit 1; }
echo "  ✓ mock backend observed MES, supplier, and ERP calls"

# ── Step 13: Bonus (optional via --bonus) ────────────────────────────────────
if [ "$BONUS" = "1" ]; then
    echo ""
    echo "=== Bonus: re-bind MAT-002 in the business system → AI follows ==="

    echo ""
    echo "[business system] update MAT-002.bound_skill_id: supplier_expedite → standard_replenish"
    curl -s -X POST "$TOOL_BACKEND_URL/admin/material-binding" \
        -H "Content-Type: application/json" \
        -d "{\"sku\":\"MAT-002\",\"bound_skill_id\":\"$STANDARD_REPLENISH_ID\"}" | python3 -m json.tool

    echo ""
    echo "[KN] rebuild to refresh Vega's batch-mode resource snapshot"
    kweaver bkn build "$KN_ID" --wait --timeout 60 2>&1 | tail -2

    echo ""
    echo "--- MAT-002 (re-trigger after binding change) ---"
    kweaver agent chat "$AGENT_ID" \
        -m "Material MAT-002 hit critical stock level again. Use find_skills, decide, execute it when possible, and end with SELECTED_SKILL_NAME=<name>." \
        --stream 2>&1 \
        | sed '/^(node:.*Warning:/d; /trace-warnings/d; /To continue this conversation/,$d' \
        | tee "$SCRIPT_DIR/.chat-MAT-002-bonus.log" \
        | tail -40
    assert_selected "$SCRIPT_DIR/.chat-MAT-002-bonus.log" "standard_replenish"
    execute_mock_action "MAT-002" "standard_replenish"

    echo ""
    echo ">>> Compare with the MAT-002 result above (Step 11)."
    echo ">>> Expected: this run picks standard_replenish (matches the new binding)."
    echo "  ✓ verified Bonus route=standard_replenish and mock action"
fi

echo ""
echo "=== All steps completed; cleanup runs on exit ==="
