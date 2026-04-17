#!/usr/bin/env bash
# =============================================================================
# 01-db-to-qa: From Database to Intelligent Q&A
#
# End-to-end flow: MySQL → Datasource → Knowledge Network → Semantic Search → Agent Chat
# =============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# ── CLI flags ────────────────────────────────────────────────────────────────
SEED_ONLY=0
usage() {
    cat <<EOF
Usage: $(basename "$0") [options]

Options:
  -s, --seed-only   Run Step 0 only: import seed.sql into MySQL, then exit.
                    Skips datasource / KN / agent steps. Useful for resetting
                    the demo database without touching the platform.
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

# ── Debug helpers (set DEBUG=1 or DEBUG=true in .env) ───────────────────────
# Never logs DB_PASS.
DEBUG="${DEBUG:-0}"
debug() {
    if [ "$DEBUG" = "1" ] || [ "$DEBUG" = "true" ]; then
        echo "[debug] $*" >&2
    fi
}

debug_dump_json() {
    local label="$1"
    local payload="$2"
    if [ "$DEBUG" = "1" ] || [ "$DEBUG" = "true" ]; then
        echo "[debug] --- ${label} (raw) ---" >&2
        echo "$payload" >&2
        echo "[debug] --- end ${label} ---" >&2
    fi
}

# ── Load config ──────────────────────────────────────────────────────────────
if [ -f "$SCRIPT_DIR/.env" ]; then
    # shellcheck disable=SC1091
    source "$SCRIPT_DIR/.env"
fi

DB_HOST="${DB_HOST:?Set DB_HOST in .env}"
# Optional: host for Step 0 only (mysql on your PC). Use public IP / VPN-reachable address if your laptop
# cannot reach DB_HOST (e.g. DB_HOST is a cloud internal IP like 172.x for kweaver ds connect).
DB_HOST_SEED="${DB_HOST_SEED:-$DB_HOST}"
DB_PORT="${DB_PORT:-3306}"
DB_NAME="${DB_NAME:?Set DB_NAME in .env}"
DB_USER="${DB_USER:?Set DB_USER in .env}"
DB_PASS="${DB_PASS:?Set DB_PASS in .env}"

# MySQL client binary (must be installed locally; only `kweaver` talks to the platform)
MYSQL_BIN="${MYSQL_BIN:-mysql}"
if ! command -v "$MYSQL_BIN" >/dev/null 2>&1; then
    # Default name not on PATH: common Homebrew layouts (Intel / Apple Silicon) without requiring PATH
    if [ "$MYSQL_BIN" = "mysql" ]; then
        _brew_mysql="$(brew --prefix mysql-client 2>/dev/null)/bin/mysql"
        for _p in "$_brew_mysql" /opt/homebrew/opt/mysql-client/bin/mysql /usr/local/opt/mysql-client/bin/mysql; do
            if [ -x "$_p" ]; then
                MYSQL_BIN="$_p"
                break
            fi
        done
    fi
fi
if ! command -v "$MYSQL_BIN" >/dev/null 2>&1; then
    echo "Error: MySQL client not found (${MYSQL_BIN}). Step 0 runs mysql on your machine to import seed.sql."
    echo "  macOS:  brew install mysql-client"
    echo "          export PATH=\"\$(brew --prefix mysql-client)/bin:\$PATH\""
    echo "  Ubuntu: sudo apt install -y mysql-client"
    echo "  Or set MYSQL_BIN in .env to the full path of the mysql executable."
    exit 1
fi

debug "script: $SCRIPT_DIR/run.sh"
debug "host: $(hostname 2>/dev/null || true) date: $(date -Iseconds 2>/dev/null || date)"
debug "MYSQL_BIN=$MYSQL_BIN"
debug "kweaver=$(command -v kweaver 2>/dev/null || echo 'not found')"
if command -v kweaver >/dev/null 2>&1; then
    debug "kweaver version: $(kweaver --version 2>&1 || true)"
fi
if [ "$DEBUG" = "1" ] || [ "$DEBUG" = "true" ]; then
    echo "[debug] DB_HOST=$DB_HOST (kweaver ds connect / platform) DB_HOST_SEED=$DB_HOST_SEED (local mysql Step 0) DB_PORT=$DB_PORT DB_NAME=$DB_NAME DB_USER=$DB_USER DB_PASS=***" >&2
    echo "[debug] kweaver config (first lines):" >&2
    kweaver config show 2>&1 | head -25 >&2 || true
fi

TIMESTAMP=$(date +%s)
DS_NAME="example_ds_${TIMESTAMP}"
KN_NAME="example_kn_${TIMESTAMP}"

# Track created resources for cleanup
DS_ID=""
KN_ID=""

cleanup() {
    if [ -z "$KN_ID" ] && [ -z "$DS_ID" ]; then
        return 0
    fi
    echo ""
    echo "=== Cleanup ==="
    [ -n "$KN_ID" ] && kweaver bkn delete "$KN_ID" -y 2>/dev/null && echo "  Deleted KN $KN_ID"
    [ -n "$DS_ID" ] && kweaver ds delete "$DS_ID" -y 2>/dev/null && echo "  Deleted datasource $DS_ID"
    echo "Done."
}
trap cleanup EXIT

# ── Step 0: Seed the database ───────────────────────────────────────────────
echo "=== Step 0: Seed sample data into MySQL ==="
if [ "$DB_HOST_SEED" != "$DB_HOST" ]; then
    echo "  (from this PC: $DB_HOST_SEED:$DB_PORT — platform will use $DB_HOST in Step 1)"
fi
debug "Step 0: mysql client imports seed.sql into default database $DB_NAME"
debug "Step 0: note — this runs on YOUR machine; platform pods use DB_HOST from Step 1."
debug "Step 0: mysql args (password hidden): -h $DB_HOST_SEED -P $DB_PORT -u $DB_USER -p*** $DB_NAME < seed.sql"
# Pass DB_NAME as the default database (seed.sql has no USE — works with schema-only users)
"$MYSQL_BIN" -h "$DB_HOST_SEED" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" < "$SCRIPT_DIR/seed.sql"
echo "  Imported seed.sql → ${DB_NAME} (erp_material_bom, erp_purchase_order)"

if [ "$SEED_ONLY" = "1" ]; then
    echo ""
    echo "=== Seed-only mode: stopping after Step 0 ==="
    echo "  Database '${DB_NAME}' on ${DB_HOST_SEED}:${DB_PORT} is ready."
    echo "  Re-run without --seed-only to continue with datasource / KN / agent steps."
    exit 0
fi

# ── Step 1: Connect datasource ──────────────────────────────────────────────
echo ""
echo "=== Step 1: Connect MySQL datasource ==="
echo "  Host: $DB_HOST:$DB_PORT  Database: $DB_NAME"

debug "Step 1: kweaver ds connect mysql $DB_HOST $DB_PORT $DB_NAME --name $DS_NAME"
DS_JSON=$(kweaver ds connect mysql "$DB_HOST" "$DB_PORT" "$DB_NAME" \
    --account "$DB_USER" --password "$DB_PASS" --name "$DS_NAME")
debug_dump_json "ds connect response" "$DS_JSON"

DS_ID=$(echo "$DS_JSON" | python3 -c "import sys,json; print(json.load(sys.stdin).get('datasource_id',''))" 2>/dev/null || true)
if [ -z "$DS_ID" ]; then
    echo "Error: could not parse datasource_id from kweaver ds connect output." >&2
    debug_dump_json "ds connect (parse failed)" "$DS_JSON"
    exit 1
fi
echo "  Datasource created: $DS_ID"

# ── Step 2: Create Knowledge Network ────────────────────────────────────────
echo ""
echo "=== Step 2: Create Knowledge Network from datasource ==="

debug "Step 2: kweaver bkn create-from-ds $DS_ID --name $KN_NAME --build"
KN_JSON=$(kweaver bkn create-from-ds "$DS_ID" --name "$KN_NAME" --build)
debug_dump_json "create-from-ds response" "$KN_JSON"

KN_ID=$(echo "$KN_JSON" | python3 -c "import sys,json; print(json.load(sys.stdin).get('kn_id',''))" 2>/dev/null || true)
if [ -z "$KN_ID" ]; then
    echo "Error: could not parse kn_id from kweaver bkn create-from-ds output." >&2
    debug_dump_json "create-from-ds (parse failed)" "$KN_JSON"
    exit 1
fi
echo "  Knowledge Network created: $KN_ID"

# Show auto-discovered object types
OT_COUNT=$(echo "$KN_JSON" | python3 -c "
import sys,json
d = json.load(sys.stdin)
ots = d.get('object_types', [])
print(len(ots))
for ot in ots[:5]:
    print(f\"    - {ot.get('name', '?')}\")
if len(ots) > 5:
    print(f'    ... and {len(ots)-5} more')
")
echo "  Auto-discovered object types:"
echo "$OT_COUNT"

# ── Step 3: Explore schema ──────────────────────────────────────────────────
echo ""
echo "=== Step 3: Explore Knowledge Network schema ==="

OT_LIST=$(kweaver bkn object-type list "$KN_ID")
echo "$OT_LIST" | python3 -c "
import sys, json
data = json.load(sys.stdin)
entries = data.get('entries', data) if isinstance(data, dict) else data
if isinstance(entries, list):
    print(f'  Object types ({len(entries)}):')
    for e in entries[:8]:
        name = e.get('name', '?')
        props = e.get('data_properties', [])
        print(f'    - {name} ({len(props)} properties)')
"

# Pick the first OT and show its properties
FIRST_OT_ID=$(echo "$OT_LIST" | python3 -c "
import sys, json
data = json.load(sys.stdin)
entries = data.get('entries', data) if isinstance(data, dict) else data
if isinstance(entries, list) and entries:
    print(entries[0].get('id', ''))
")

if [ -n "$FIRST_OT_ID" ]; then
    echo ""
    echo "  Properties of first object type:"
    kweaver bkn object-type get "$KN_ID" "$FIRST_OT_ID" --pretty 2>/dev/null | python3 -c "
import sys, json
data = json.load(sys.stdin)
entry = data.get('entries', [data])[0] if isinstance(data, dict) else data
if isinstance(entry, dict):
    for p in (entry.get('data_properties') or [])[:10]:
        print(f\"    - {p.get('name', '?')} ({p.get('type', '?')})\")
" 2>/dev/null || true
fi

# ── Step 4: Semantic search via context-loader ──────────────────────────────
echo ""
echo "=== Step 4: Semantic search ==="

kweaver context-loader config set --kn-id "$KN_ID" --name example-e2e >/dev/null 2>&1
echo "  Context-loader configured for KN $KN_ID"

echo "  Searching schema: \"采购订单 物料\""
SCHEMA_RAW=$(kweaver context-loader kn-search "采购订单 物料" --only-schema 2>/dev/null \
    | python3 -c "import sys,json; print(json.load(sys.stdin).get('raw',''))" 2>/dev/null || true)

echo "$SCHEMA_RAW" | head -10 | sed 's/^/    /'
[ "$(echo "$SCHEMA_RAW" | wc -l)" -gt 10 ] && echo "    ..."

# ── Step 5: Chat with Agent ─────────────────────────────────────────────────
echo ""
echo "=== Step 5: Chat with Agent ==="

# Use provided AGENT_ID or find the first available agent
if [ -z "${AGENT_ID:-}" ]; then
    AGENT_ID=$(kweaver agent list --limit 1 2>/dev/null | python3 -c "
import sys, json
data = json.load(sys.stdin)
if isinstance(data, list) and data:
    print(data[0].get('id', ''))
" 2>/dev/null || true)
fi

if [ -z "$AGENT_ID" ]; then
    echo "  No agent available. Set AGENT_ID in .env or create one."
    echo "  Skipping chat step."
else
    echo "  Agent: $AGENT_ID"

    QUESTION="这个数据库里有哪些核心的业务表？它们之间是什么关系？"

    # Inject schema context so the agent answers with real table info
    if [ -n "$SCHEMA_RAW" ]; then
        PROMPT="以下是通过知识网络检索到的数据库 schema 信息：

${SCHEMA_RAW}

基于以上 schema，请回答：${QUESTION}"
    else
        PROMPT="$QUESTION"
    fi

    echo "  Question: $QUESTION"
    echo ""

    debug "Step 5: kweaver agent chat $AGENT_ID --stream"
    echo "  Agent response:"
    if [ "$DEBUG" = "1" ] || [ "$DEBUG" = "true" ]; then
        kweaver agent chat "$AGENT_ID" \
            -m "$PROMPT" \
            --stream --verbose 2>&1 | sed 's/^/    /'
    else
        kweaver agent chat "$AGENT_ID" \
            -m "$PROMPT" \
            --stream 2>/dev/null | sed 's/^/    /'
    fi
fi

echo ""
echo "=== Example complete ==="
