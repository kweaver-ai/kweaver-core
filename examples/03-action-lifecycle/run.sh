#!/usr/bin/env bash
# =============================================================================
# 03-action-lifecycle: Self-Evolving Knowledge Network
#
# Flow: MySQL → Knowledge Network → Register Action Tool →
#       Define Action → Schedule → Execute → Audit Log
# =============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# ── CLI flags ─────────────────────────────────────────────────────────────────
SEED_ONLY=0
usage() {
    cat <<EOF
Usage: $(basename "$0") [options]

Options:
  -s, --seed-only   Import seed.sql only, then exit.
  -h, --help        Show this help.

Environment variables are read from .env (see env.sample).
EOF
}
while [ $# -gt 0 ]; do
    case "$1" in
        -s|--seed-only) SEED_ONLY=1 ;;
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
DB_HOST_SEED="${DB_HOST_SEED:-$DB_HOST}"
DB_PORT="${DB_PORT:-3306}"
DB_NAME="${DB_NAME:?Set DB_NAME in .env}"
DB_USER="${DB_USER:?Set DB_USER in .env}"
DB_PASS="${DB_PASS:?Set DB_PASS in .env}"

MYSQL_BIN="${MYSQL_BIN:-mysql}"
if ! command -v "$MYSQL_BIN" >/dev/null 2>&1; then
    if [ "$MYSQL_BIN" = "mysql" ]; then
        for _p in \
            "$(brew --prefix mysql-client 2>/dev/null)/bin/mysql" \
            /opt/homebrew/opt/mysql-client/bin/mysql \
            /usr/local/opt/mysql-client/bin/mysql; do
            if [ -x "$_p" ]; then MYSQL_BIN="$_p"; break; fi
        done
    fi
fi
if ! command -v "$MYSQL_BIN" >/dev/null 2>&1; then
    echo "Error: MySQL client not found (${MYSQL_BIN})." >&2
    echo "  macOS:  brew install mysql-client" >&2
    echo "  Ubuntu: sudo apt install -y mysql-client" >&2
    exit 1
fi

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

# ── Step 0: Seed the database ─────────────────────────────────────────────────
echo "=== Step 0: Seed sample data into MySQL ==="
if [ "$DB_HOST_SEED" != "$DB_HOST" ]; then
    echo "  (from this PC: $DB_HOST_SEED:$DB_PORT — platform will use $DB_HOST in Step 1)"
fi
"$MYSQL_BIN" -h "$DB_HOST_SEED" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" \
    < "$SCRIPT_DIR/seed.sql"
echo "  Imported: eval_suppliers (5 rows), eval_purchase_orders (10 rows, 4 at-risk)"

if [ "$SEED_ONLY" = "1" ]; then
    echo "Seed-only mode: done."
    exit 0
fi

# ── Step 1: Connect datasource ────────────────────────────────────────────────
echo ""
echo "=== Step 1: Connect MySQL datasource ==="
DS_JSON=$(kweaver ds connect mysql "$DB_HOST" "$DB_PORT" "$DB_NAME" \
    --account "$DB_USER" --password "$DB_PASS" --name "$DS_NAME")
debug_dump_json "ds connect" "$DS_JSON"
DS_ID=$(echo "$DS_JSON" | python3 -c \
    "import sys,json; d=json.load(sys.stdin); print(d.get('datasource_id') or d.get('id',''))")
[ -z "$DS_ID" ] && { echo "Error: no datasource_id in response" >&2; exit 1; }
echo "  Datasource: $DS_ID"

# ── Step 2: Build Knowledge Network ──────────────────────────────────────────
echo ""
echo "=== Step 2: Build Knowledge Network ==="
KN_JSON=$(kweaver bkn create-from-ds "$DS_ID" --name "$KN_NAME" --build)
debug_dump_json "create-from-ds" "$KN_JSON"
KN_ID=$(echo "$KN_JSON" | python3 -c \
    "import sys,json; d=json.load(sys.stdin); print(d.get('kn_id') or d.get('id',''))")
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
