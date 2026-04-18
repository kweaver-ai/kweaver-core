#!/usr/bin/env bash
# =============================================================================
# 03-action-lifecycle: Self-Evolving Knowledge Network
#
# Flow: CSV Files → Knowledge Network → Register Action Tool →
#       Define Action → Schedule → Execute → Audit Log
#
# No local MySQL client needed — CSV files are uploaded to the platform.
# =============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# ── CLI flags ─────────────────────────────────────────────────────────────────
usage() {
    cat <<EOF
Usage: $(basename "$0") [options]

Options:
  -h, --help   Show this help.

Environment variables are read from .env (see env.sample).
EOF
}
while [ $# -gt 0 ]; do
    case "$1" in
        -h|--help) usage; exit 0 ;;
        *) echo "Unknown argument: $1" >&2; usage >&2; exit 2 ;;
    esac
    shift
done

# ── Debug helpers ─────────────────────────────────────────────────────────────
DEBUG="${DEBUG:-0}"
debug() {
    if [ "$DEBUG" = "1" ] || [ "$DEBUG" = "true" ]; then
        echo "[debug] $*" >&2
    fi
}
debug_dump_json() {
    local label="$1" payload="$2"
    if [ "$DEBUG" = "1" ] || [ "$DEBUG" = "true" ]; then
        echo "[debug] --- ${label} ---" >&2
        echo "$payload" >&2
    fi
}

# ── Load config ───────────────────────────────────────────────────────────────
if [ -f "$SCRIPT_DIR/.env" ]; then
    # shellcheck disable=SC1091
    source "$SCRIPT_DIR/.env"
fi

DB_HOST="${DB_HOST:?Set DB_HOST in .env}"
DB_PORT="${DB_PORT:-3306}"
DB_NAME="${DB_NAME:?Set DB_NAME in .env}"
DB_USER="${DB_USER:?Set DB_USER in .env}"
DB_PASS="${DB_PASS:?Set DB_PASS in .env}"

TIMESTAMP=$(date +%s)
DS_NAME="example_action_ds_${TIMESTAMP}"
KN_NAME="example_action_kn_${TIMESTAMP}"

# Track all created resources for cleanup
DS_ID=""
KN_ID=""
BOX_ID=""
AT_ID=""
SCHED_ID=""

cleanup() {
    echo ""
    echo "=== Cleanup ==="
    [ -n "$SCHED_ID" ] && kweaver bkn action-schedule delete "$KN_ID" "$SCHED_ID" -y 2>/dev/null \
        && echo "  Deleted action-schedule $SCHED_ID" || true
    [ -n "$AT_ID" ] && kweaver bkn action-type delete "$KN_ID" "$AT_ID" -y 2>/dev/null \
        && echo "  Deleted action-type $AT_ID" || true
    [ -n "$BOX_ID" ] && kweaver call "/api/agent-operator-integration/v1/tool-box/$BOX_ID" \
        -X DELETE 2>/dev/null && echo "  Deleted toolbox $BOX_ID" || true
    [ -n "$KN_ID" ] && kweaver bkn delete "$KN_ID" -y 2>/dev/null \
        && echo "  Deleted KN $KN_ID" || true
    [ -n "$DS_ID" ] && kweaver ds delete "$DS_ID" -y 2>/dev/null \
        && echo "  Deleted datasource $DS_ID" || true
    echo "Done."
}
trap cleanup EXIT

# ── Step 1: Connect datasource ────────────────────────────────────────────────
echo "=== Step 1: Connect MySQL datasource ==="
_DS_RAW=$(kweaver ds connect mysql "$DB_HOST" "$DB_PORT" "$DB_NAME" \
    --account "$DB_USER" --password "$DB_PASS" --name "$DS_NAME" 2>&1)
debug_dump_json "ds connect" "$_DS_RAW"
DS_ID=$(echo "$_DS_RAW" | python3 -c "
import sys,json
txt=sys.stdin.read()
d=json.loads(txt[txt.index('{'):])
print(d.get('datasource_id') or d.get('id',''))
")
[ -z "$DS_ID" ] && { echo "Error: no datasource_id in response" >&2; exit 1; }
echo "  Datasource: $DS_ID"

# ── Step 2: Import CSVs and build Knowledge Network ──────────────────────────
echo ""
echo "=== Step 2: Import CSVs → Build Knowledge Network ==="
echo "  Files: $(ls "$SCRIPT_DIR/data/"*.csv | xargs -n1 basename | tr '\n' ' ')"
KN_JSON=$(kweaver bkn create-from-csv "$DS_ID" \
    --files "$SCRIPT_DIR/data/*.csv" \
    --name "$KN_NAME" \
    --build --timeout 300 2>&1)
debug_dump_json "create-from-csv" "$KN_JSON"
KN_ID=$(echo "$KN_JSON" | python3 -c "
import sys,json
txt=sys.stdin.read()
d=json.loads(txt[txt.index('{'):])
print(d.get('kn_id') or d.get('id',''))
")
[ -z "$KN_ID" ] && { echo "Error: no kn_id in response" >&2; exit 1; }
echo "  Knowledge Network: $KN_ID"

# Discover the purchase-order object type ID (needed for action condition)
OT_LIST=$(kweaver bkn object-type list "$KN_ID")
PO_OT_ID=$(echo "$OT_LIST" | python3 -c "
import sys, json
entries = json.load(sys.stdin)
if isinstance(entries, dict):
    entries = entries.get('entries', [])
for e in entries:
    name = (e.get('name','') + e.get('id','')).lower()
    if 'purchase_order' in name or 'eval_purchase' in name:
        print(e.get('id',''))
        break
")
[ -z "$PO_OT_ID" ] && { echo "Error: could not find purchase_orders object type" >&2; exit 1; }
echo "  PO object type: $PO_OT_ID"

# ── Step 3: Register demo action toolbox ──────────────────────────────────────
echo ""
echo "=== Step 3: Register action tool backend ==="
BOX_JSON=$(kweaver call "/api/agent-operator-integration/v1/tool-box" \
    -X POST \
    -d "{\"metadata_type\":\"openapi\",\"box_name\":\"eval_action_toolbox_${TIMESTAMP}\",\"box_desc\":\"Demo toolbox for action-lifecycle example\",\"box_svc_url\":\"http://ontology-manager-svc:13014\",\"source\":\"custom\"}")
debug_dump_json "create toolbox" "$BOX_JSON"
BOX_ID=$(echo "$BOX_JSON" | python3 -c \
    "import sys,json; print(json.load(sys.stdin).get('box_id',''))")
[ -z "$BOX_ID" ] && { echo "Error: no box_id in toolbox response" >&2; exit 1; }
echo "  Toolbox: $BOX_ID"

# ── Step 4: Register a tool in the toolbox ────────────────────────────────────
echo ""
echo "=== Step 4: Register tool (OpenAPI spec) ==="

_OPENAPI_TMP=$(mktemp /tmp/eval_tool_openapi_XXXXXX.json)
trap 'rm -f "$_OPENAPI_TMP"; cleanup' EXIT
cat > "$_OPENAPI_TMP" <<'OPENAPI'
{
  "openapi": "3.0.0",
  "info": {"title": "采购单风险跟进", "version": "1.0.0"},
  "servers": [{"url": "http://ontology-manager-svc:13014"}],
  "paths": {
    "/api/ontology-manager/in/v1/health": {
      "get": {
        "summary": "采购单风险跟进",
        "operationId": "follow_up_at_risk_po",
        "responses": {"200": {"description": "ok"}}
      }
    }
  }
}
OPENAPI

_TOKEN=$(kweaver token 2>/dev/null)
_PLATFORM=$(kweaver config show 2>/dev/null | python3 -c \
    "import sys; [print(l.split()[-1]) for l in sys.stdin if 'Platform' in l]" | head -1)
_BD=$(kweaver config show 2>/dev/null | python3 -c \
    "import sys; [print(l.split()[2]) for l in sys.stdin if 'Business Domain' in l]" | head -1)

_TOOL_RESP=$(curl -s \
    -H "Authorization: Bearer $_TOKEN" \
    -H "x-business-domain: $_BD" \
    -F "metadata_type=openapi" \
    -F "data=@${_OPENAPI_TMP}" \
    "${_PLATFORM}/api/agent-operator-integration/v1/tool-box/$BOX_ID/tool")
rm -f "$_OPENAPI_TMP"
debug_dump_json "create tool" "$_TOOL_RESP"

TOOL_ID=$(echo "$_TOOL_RESP" | python3 -c \
    "import sys,json; d=json.load(sys.stdin); ids=d.get('success_ids',[]); print(ids[0] if ids else '')")
[ -z "$TOOL_ID" ] && { echo "Error: no tool_id in response" >&2; exit 1; }
echo "  Tool: $TOOL_ID"

# ── Step 5: Publish toolbox and enable tool ───────────────────────────────────
echo ""
echo "=== Step 5: Publish toolbox ==="
kweaver call "/api/agent-operator-integration/v1/tool-box/$BOX_ID/status" \
    -X POST -d '{"status":"published"}' > /dev/null
kweaver call "/api/agent-operator-integration/v1/tool-box/$BOX_ID/tools/status" \
    -X POST -d "[{\"tool_id\":\"$TOOL_ID\",\"status\":\"enabled\"}]" > /dev/null
echo "  Toolbox published, tool enabled"

# ── Step 6: Define action type ────────────────────────────────────────────────
echo ""
echo "=== Step 6: Define action type — 采购单风险跟进 ==="

AT_BODY=$(python3 -c "
import json
body = {
    'name': '采购单风险跟进',
    'action_type': 'modify',
    'object_type_id': '$PO_OT_ID',
    'tags': ['采购', '风险管理'],
    'comment': '发现供应商状态异常的采购单，自动触发跟进行动',
    'action_source': {
        'type': 'tool',
        'box_id': '$BOX_ID',
        'tool_id': '$TOOL_ID'
    },
    'condition': {
        'object_type_id': '$PO_OT_ID',
        'field': 'supplier_status',
        'operation': '==',
        'value': 'at_risk'
    }
}
print(json.dumps(body))
")

AT_JSON=$(kweaver bkn action-type create "$KN_ID" "$AT_BODY")
debug_dump_json "action-type create" "$AT_JSON"
AT_ID=$(echo "$AT_JSON" | python3 -c "
import sys,json
d=json.load(sys.stdin)
if isinstance(d, list): d = d[0]
print(d.get('id',''))
")
[ -z "$AT_ID" ] && { echo "Error: no action-type id in response" >&2; exit 1; }
echo "  Action type: $AT_ID"

# ── Step 7: Query — verify affected instances ─────────────────────────────────
echo ""
echo "=== Step 7: Query — which POs have at-risk suppliers? ==="
QUERY_JSON=$(kweaver bkn action-type query "$KN_ID" "$AT_ID" '{}' 2>&1 || true)
debug_dump_json "action-type query" "$QUERY_JSON"
AFFECTED=$(echo "$QUERY_JSON" | python3 -c \
    "import sys,json; d=json.load(sys.stdin); print(d.get('total_count',0))" 2>/dev/null || echo "unknown")
echo "  Found $AFFECTED purchase order(s) linked to at-risk suppliers"
echo "  (The knowledge network identified these via supplier relationship context)"

# Capture first identity for use in Step 10
FIRST_IDENTITY=$(echo "$QUERY_JSON" | python3 -c "
import sys,json
d=json.load(sys.stdin)
actions=d.get('actions',[])
if actions:
    print(json.dumps(actions[0].get('_instance_identity',{})))
else:
    print('{}')
" 2>/dev/null || echo "{}")

# ── Step 8: Create action schedule ────────────────────────────────────────────
echo ""
echo "=== Step 8: Schedule — every day at 08:00 ==="

SCHED_BODY=$(python3 -c "
import json
print(json.dumps({
    'name': '采购单风险每日巡检',
    'cron_expression': '0 8 * * *',
    'action_type_id': '$AT_ID',
    '_instance_identities': [{}]
}))
")
SCHED_JSON=$(kweaver bkn action-schedule create "$KN_ID" "$SCHED_BODY")
debug_dump_json "action-schedule create" "$SCHED_JSON"
SCHED_ID=$(echo "$SCHED_JSON" | python3 -c \
    "import sys,json; d=json.load(sys.stdin); print(d.get('id',''))")
[ -z "$SCHED_ID" ] && { echo "Error: no schedule id in response" >&2; exit 1; }
echo "  Schedule: $SCHED_ID (cron: 0 8 * * *)"

# ── Step 9: Confirm schedule is active ────────────────────────────────────────
echo ""
echo "=== Step 9: Confirm schedule active ==="
SCHED_DETAIL=$(kweaver bkn action-schedule get "$KN_ID" "$SCHED_ID")
SCHED_STATUS=$(echo "$SCHED_DETAIL" | python3 -c \
    "import sys,json; print(json.load(sys.stdin).get('status','unknown'))")
if [ "$SCHED_STATUS" = "inactive" ]; then
    kweaver bkn action-schedule set-status "$KN_ID" "$SCHED_ID" active > /dev/null
    SCHED_STATUS="active"
fi
echo "  Schedule status: $SCHED_STATUS"
echo "  The knowledge network will act autonomously every morning at 08:00."

# ── Step 10: Trigger action now ───────────────────────────────────────────────
echo ""
echo "=== Step 10: Trigger action — first run ==="
echo "  (In production this runs automatically at 08:00.)"
echo "  Executing now so you can see results immediately..."

EXEC_BODY=$(python3 -c "import json,sys; print(json.dumps({'_instance_identities': [json.loads('$FIRST_IDENTITY')]}))" 2>/dev/null \
    || python3 -c "import json; print(json.dumps({'_instance_identities': [{}]}))")
EXEC_JSON=$(kweaver bkn action-type execute "$KN_ID" "$AT_ID" "$EXEC_BODY" \
    --timeout 60 2>&1 || true)
debug_dump_json "action-type execute" "$EXEC_JSON"

EXEC_ID=$(echo "$EXEC_JSON" | python3 -c \
    "import sys,json; print(json.load(sys.stdin).get('id',''))" 2>/dev/null || true)
EXEC_STATUS=$(echo "$EXEC_JSON" | python3 -c \
    "import sys,json; print(json.load(sys.stdin).get('status','unknown'))" 2>/dev/null || true)
EXEC_TOTAL=$(echo "$EXEC_JSON" | python3 -c \
    "import sys,json; print(json.load(sys.stdin).get('total_count',0))" 2>/dev/null || true)

echo "  Execution ID : ${EXEC_ID:-n/a}"
echo "  Instances    : $EXEC_TOTAL"
echo "  Status       : $EXEC_STATUS"
echo "  (demo tool has no real backend — the execution record is what matters)"

# ── Step 11: Audit log ────────────────────────────────────────────────────────
echo ""
echo "=== Step 11: Audit log — what the knowledge network has done ==="
LOG_JSON=$(kweaver bkn action-log list "$KN_ID" 2>&1 || true)
debug_dump_json "action-log list" "$LOG_JSON"
LOG_COUNT=$(echo "$LOG_JSON" | python3 -c "
import sys,json
d=json.load(sys.stdin)
entries=d.get('entries') or []
print(max(d.get('total_count',0), len(entries)))
" 2>/dev/null || echo 0)
echo "  Total executions recorded: $LOG_COUNT"
echo ""
echo "$LOG_JSON" | python3 -c "
import sys, json, datetime
try:
    d = json.load(sys.stdin)
    entries = d.get('entries') or []
    for e in (entries[:5] if entries else []):
        ts = e.get('create_time', 0)
        t = datetime.datetime.fromtimestamp(ts/1000).strftime('%Y-%m-%d %H:%M') if ts else 'n/a'
        print(f'  [{t}]  {e.get(\"action_type_name\",\"?\")} → {e.get(\"status\",\"?\")}  (id: {e.get(\"id\",\"?\")[:8]}...)')
except Exception:
    pass
" 2>/dev/null || true

echo ""
echo "======================================================"
echo "  Knowledge network is now self-acting."
echo "  Every morning at 08:00 it will:"
echo "    1. Identify POs linked to at-risk suppliers"
echo "    2. Trigger the follow-up action"
echo "    3. Record the result in the audit log"
echo ""
echo "  Check the log anytime:"
echo "    kweaver bkn action-log list $KN_ID"
echo "======================================================"
