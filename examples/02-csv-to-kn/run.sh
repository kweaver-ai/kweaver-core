#!/usr/bin/env bash
# =============================================================================
# 02-csv-to-kn: From CSV Files to Knowledge Network
#
# Import local CSV files → Knowledge Network → Graph Exploration → Agent Q&A
# =============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# ── CLI flags ────────────────────────────────────────────────────────────────
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
    [ "$DEBUG" = "1" ] || [ "$DEBUG" = "true" ] && echo "[debug] $*" >&2 || true
}
debug_json() {
    local label="$1" payload="$2"
    if [ "$DEBUG" = "1" ] || [ "$DEBUG" = "true" ]; then
        echo "[debug] --- ${label} ---" >&2
        echo "$payload" >&2
        echo "[debug] --- end ${label} ---" >&2
    fi
}

# ── Load config ──────────────────────────────────────────────────────────────
[ -f "$SCRIPT_DIR/.env" ] && source "$SCRIPT_DIR/.env"

DB_HOST="${DB_HOST:?Set DB_HOST in .env}"
DB_PORT="${DB_PORT:-3306}"
DB_NAME="${DB_NAME:?Set DB_NAME in .env}"
DB_USER="${DB_USER:?Set DB_USER in .env}"
DB_PASS="${DB_PASS:?Set DB_PASS in .env}"

TIMESTAMP=$(date +%s)
DS_NAME="csv_example_ds_${TIMESTAMP}"
KN_NAME="csv_example_kn_${TIMESTAMP}"

DS_ID=""
KN_ID=""

cleanup() {
    [ -z "$KN_ID" ] && [ -z "$DS_ID" ] && return 0
    echo ""
    echo "=== Cleanup ==="
    [ -n "$KN_ID" ] && kweaver bkn delete "$KN_ID" -y 2>/dev/null && echo "  Deleted KN $KN_ID"
    [ -n "$DS_ID" ] && kweaver ds delete "$DS_ID" -y 2>/dev/null && echo "  Deleted datasource $DS_ID"
    echo "Done."
}
trap cleanup EXIT

# ── Step 1: Connect datasource ────────────────────────────────────────────────
echo "=== Step 1: Connect datasource ==="
echo "  Host: $DB_HOST:$DB_PORT  Database: $DB_NAME"

DS_JSON=$(kweaver ds connect mysql "$DB_HOST" "$DB_PORT" "$DB_NAME" \
    --account "$DB_USER" --password "$DB_PASS" --name "$DS_NAME")
debug_json "ds connect" "$DS_JSON"

DS_ID=$(echo "$DS_JSON" | python3 -c \
    "import sys,json; d=json.load(sys.stdin); \
     r=d[0] if isinstance(d,list) else d; \
     print(r.get('datasource_id') or r.get('id',''))" 2>/dev/null || true)

[ -z "$DS_ID" ] && { echo "Error: ds connect failed — check credentials." >&2; exit 1; }
echo "  Datasource: $DS_ID"

# ── Step 2: Import CSVs + build Knowledge Network ─────────────────────────────
echo ""
echo "=== Step 2: Import CSVs and build Knowledge Network ==="
echo "  Files: $(ls "$SCRIPT_DIR/data/"*.csv | xargs -n1 basename | tr '\n' ' ')"

KN_JSON=$(kweaver bkn create-from-csv "$DS_ID" \
    --files "$SCRIPT_DIR/data/*.csv" \
    --name "$KN_NAME" \
    --build \
    --timeout 300)
debug_json "create-from-csv" "$KN_JSON"

KN_ID=$(echo "$KN_JSON" | python3 -c \
    "import sys,json; d=json.load(sys.stdin); \
     print(d.get('kn_id') or d.get('id',''))" 2>/dev/null || true)

[ -z "$KN_ID" ] && { echo "Error: create-from-csv failed." >&2; exit 1; }
echo "  Knowledge Network: $KN_ID"

# Show auto-discovered object types
echo "$KN_JSON" | python3 -c "
import sys, json
d = json.load(sys.stdin)
ots = d.get('object_types', [])
if ots:
    print(f'  Auto-discovered object types ({len(ots)}):')
    for ot in ots:
        print(f'    - {ot.get(\"name\",\"?\")}')
" 2>/dev/null || true

# ── Step 3: Explore schema ────────────────────────────────────────────────────
echo ""
echo "=== Step 3: Explore schema ==="

OT_LIST=$(kweaver bkn object-type list "$KN_ID")
debug_json "object-type list" "$OT_LIST"

OT_IDS=$(echo "$OT_LIST" | python3 -c "
import sys, json
data = json.load(sys.stdin)
entries = data.get('entries', data) if isinstance(data, dict) else data
print(f'  Object types ({len(entries)}):')
ids = []
for e in entries:
    name = e.get('name','?')
    props = e.get('data_properties', [])
    ot_id = e.get('id') or e.get('ot_id','')
    print(f'    - {name}  ({len(props)} properties)  id={ot_id}')
    ids.append(ot_id)
# Print first two IDs for later steps
print('__IDS__:' + ','.join(ids[:2]))
" 2>/dev/null || true)

echo "$OT_IDS" | grep -v "^__IDS__"
FIRST_OT=$(echo "$OT_IDS" | grep "^__IDS__" | cut -d: -f2 | cut -d, -f1)
SECOND_OT=$(echo "$OT_IDS" | grep "^__IDS__" | cut -d: -f2 | cut -d, -f2)

if [ -n "$FIRST_OT" ]; then
    echo ""
    echo "  Properties of '$(echo "$OT_LIST" | python3 -c \
        "import sys,json; d=json.load(sys.stdin); \
         e=d.get('entries',d) if isinstance(d,dict) else d; \
         print(e[0].get('name','?') if e else '?')" 2>/dev/null)':"
    kweaver bkn object-type get "$KN_ID" "$FIRST_OT" --pretty 2>/dev/null | python3 -c "
import sys, json
data = json.load(sys.stdin)
entry = data
if isinstance(data, dict) and 'entries' in data:
    entry = data['entries'][0] if data['entries'] else data
for p in (entry.get('data_properties') or [])[:8]:
    print(f'    - {p.get(\"name\",\"?\")}  ({p.get(\"type\",\"?\")})')
" 2>/dev/null || true
fi

# ── Step 4: Query instances ───────────────────────────────────────────────────
echo ""
echo "=== Step 4: Query instances ==="

if [ -n "$FIRST_OT" ]; then
    echo "  Querying first 5 records from object type '$FIRST_OT':"
    kweaver bkn object-type query "$KN_ID" "$FIRST_OT" --limit 5 2>/dev/null | python3 -c "
import sys, json
data = json.load(sys.stdin)
instances = data.get('datas') or data.get('instances') or data.get('entries') or (data if isinstance(data, list) else [])
for inst in instances[:5]:
    items = [(k,v) for k,v in inst.items() if not k.startswith('_')][:4]
    print('    ' + '  '.join(f'{k}={v}' for k,v in items))
" 2>/dev/null || echo "    (no instances returned)"
fi

# ── Step 5: Subgraph traversal ────────────────────────────────────────────────
echo ""
echo "=== Step 5: Subgraph traversal ==="

# Get an instance ID from the first object type
if [ -n "$FIRST_OT" ]; then
    INSTANCE_ID=$(kweaver bkn object-type query "$KN_ID" "$FIRST_OT" --limit 1 2>/dev/null | python3 -c "
import sys, json
data = json.load(sys.stdin)
instances = data.get('datas') or data.get('instances') or data.get('entries') or (data if isinstance(data,list) else [])
if instances:
    print(instances[0].get('_instance_id') or instances[0].get('id') or instances[0].get('rid',''))
" 2>/dev/null || true)
fi

if [ -n "${INSTANCE_ID:-}" ]; then
    echo "  Starting from instance: $INSTANCE_ID"
    SUBGRAPH=$(kweaver bkn subgraph "$KN_ID" "$INSTANCE_ID" --depth 2 2>/dev/null \
        || kweaver bkn subgraph "$KN_ID" "$INSTANCE_ID" 2>/dev/null || true)
    if [ -n "$SUBGRAPH" ]; then
        echo "$SUBGRAPH" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    nodes = data.get('nodes') or data.get('vertices') or []
    edges = data.get('edges') or data.get('relations') or []
    print(f'  Graph: {len(nodes)} nodes, {len(edges)} edges')
except Exception:
    pass
" 2>/dev/null || true
    fi
else
    echo "  (no instances available for subgraph — skipping)"
fi

# ── Step 6: Semantic search ───────────────────────────────────────────────────
echo ""
echo "=== Step 6: Semantic search via context-loader ==="

kweaver context-loader config set --kn-id "$KN_ID" >/dev/null 2>&1 || true

SEARCH_RESULT=$(kweaver context-loader kn-search "部门 预算 人员" \
    --kn-id "$KN_ID" --only-schema 2>/dev/null || true)

if [ -n "$SEARCH_RESULT" ]; then
    SCHEMA_RAW=$(echo "$SEARCH_RESULT" | python3 -c \
        "import sys,json; print(json.load(sys.stdin).get('raw',''))" 2>/dev/null || true)
    echo "  Schema context (first 8 lines):"
    echo "$SCHEMA_RAW" | head -8 | sed 's/^/    /'
    [ "$(echo "$SCHEMA_RAW" | wc -l)" -gt 8 ] && echo "    ..."
else
    echo "  (context-loader not available on this deployment)"
    SCHEMA_RAW=""
fi

# ── Step 7: Export KN ────────────────────────────────────────────────────────
echo ""
echo "=== Step 7: Export Knowledge Network ==="

EXPORT=$(kweaver bkn export "$KN_ID" 2>/dev/null || true)
if [ -n "$EXPORT" ]; then
    echo "$EXPORT" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    ots = data.get('object_types', [])
    rts = data.get('relation_types', [])
    print(f'  Exported: {len(ots)} object types, {len(rts)} relation types')
except Exception:
    pass
" 2>/dev/null || true
fi

# ── Step 8: Agent Q&A ─────────────────────────────────────────────────────────
echo ""
echo "=== Step 8: Agent Q&A ==="

AGENT_ID="${AGENT_ID:-}"
if [ -z "$AGENT_ID" ]; then
    AGENT_ID=$(kweaver agent list --limit 1 2>/dev/null | python3 -c \
        "import sys,json; d=json.load(sys.stdin); \
         print(d[0].get('id','') if isinstance(d,list) and d else '')" 2>/dev/null || true)
fi

if [ -z "$AGENT_ID" ]; then
    echo "  No agent available — set AGENT_ID in .env or create one with \`kweaver agent create\`."
    echo "  Skipping Q&A step."
else
    QUESTION="这份数据里，哪个部门的预算最高？Engineering 部门有多少员工？"

    # Query actual instance data for departments and employees
    DEPT_DATA=""
    EMP_DATA=""
    if [ -n "$FIRST_OT" ]; then
        DEPT_DATA=$(kweaver bkn object-type query "$KN_ID" "$FIRST_OT" --limit 20 2>/dev/null | python3 -c "
import sys, json
data = json.load(sys.stdin)
rows = data.get('datas') or data.get('instances') or data.get('entries') or []
lines = []
for r in rows:
    lines.append('  ' + ', '.join(f'{k}={v}' for k,v in r.items() if not k.startswith('_')))
print('\n'.join(lines))
" 2>/dev/null || true)
    fi
    if [ -n "$SECOND_OT" ]; then
        EMP_DATA=$(kweaver bkn object-type query "$KN_ID" "$SECOND_OT" --limit 20 2>/dev/null | python3 -c "
import sys, json
data = json.load(sys.stdin)
rows = data.get('datas') or data.get('instances') or data.get('entries') or []
lines = []
for r in rows:
    lines.append('  ' + ', '.join(f'{k}={v}' for k,v in r.items() if not k.startswith('_')))
print('\n'.join(lines))
" 2>/dev/null || true)
    fi

    PROMPT=""
    if [ -n "${SCHEMA_RAW:-}" ]; then
        PROMPT="以下是通过知识网络检索到的 schema 信息：

${SCHEMA_RAW}
"
    fi
    if [ -n "${DEPT_DATA:-}" ]; then
        PROMPT="${PROMPT}
departments 表实际数据：
${DEPT_DATA}
"
    fi
    if [ -n "${EMP_DATA:-}" ]; then
        PROMPT="${PROMPT}
employees 表实际数据：
${EMP_DATA}
"
    fi
    PROMPT="${PROMPT}
请基于以上数据回答：${QUESTION}"

    echo "  Agent: $AGENT_ID"
    echo "  Question: $QUESTION"
    echo ""
    echo "  Response:"
    kweaver agent chat "$AGENT_ID" \
        -m "$PROMPT" \
        --stream 2>/dev/null | sed 's/^/    /' || \
    kweaver agent chat "$AGENT_ID" \
        -m "$PROMPT" \
        --no-stream 2>/dev/null | sed 's/^/    /' || \
    echo "    (agent unavailable)"
fi

echo ""
echo "=== Example complete ==="
echo "  KN '$KN_NAME' ($KN_ID) will be deleted on exit."
