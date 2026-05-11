#!/usr/bin/env bash
# =============================================================================
# 06-world-cup: Fjelstul World Cup CSVs → Knowledge Network → Agent Q&A
#
# Data: https://github.com/jfjelstul/worldcup (CC-BY-SA 4.0) — run download.sh first.
# =============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

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

DEBUG="${DEBUG:-0}"
debug_json() {
    local label="$1" payload="$2"
    if [ "$DEBUG" = "1" ] || [ "$DEBUG" = "true" ]; then
        echo "[debug] --- ${label} ---" >&2
        echo "$payload" >&2
        echo "[debug] --- end ${label} ---" >&2
    fi
}

[ -f "$SCRIPT_DIR/.env" ] && source "$SCRIPT_DIR/.env"

DB_HOST="${DB_HOST:?Set DB_HOST in .env}"
DB_PORT="${DB_PORT:-3306}"
DB_NAME="${DB_NAME:?Set DB_NAME in .env}"
DB_USER="${DB_USER:?Set DB_USER in .env}"
DB_PASS="${DB_PASS:?Set DB_PASS in .env}"

DATA_DIR="$SCRIPT_DIR/data"
# shellcheck source=/dev/null
source "$SCRIPT_DIR/scripts/worldcup_dataset_stems.inc.sh"
EXPECTED_CSV=${#WORLD_CUP_DATASET_STEMS[@]}
count_csvs() {
    shopt -s nullglob
    local a=("$DATA_DIR"/*.csv)
    echo "${#a[@]}"
}

# ── Ensure CSVs exist ───────────────────────────────────────────────────────
if [ ! -d "$DATA_DIR" ] || [ "$(count_csvs)" -lt "$EXPECTED_CSV" ]; then
    echo "=== Step 0: Download World Cup CSVs (${EXPECTED_CSV} files) ==="
    "$SCRIPT_DIR/download.sh"
fi
actual=$(count_csvs)
if [ "$actual" -lt "$EXPECTED_CSV" ]; then
    echo "Error: expected at least ${EXPECTED_CSV} CSV files in $DATA_DIR, found $actual." >&2
    exit 1
fi

SLIM_WIDE_CSV_FOR_MYSQL="${SLIM_WIDE_CSV_FOR_MYSQL:-1}"
if [ "$SLIM_WIDE_CSV_FOR_MYSQL" = 1 ] || [ "$SLIM_WIDE_CSV_FOR_MYSQL" = true ]; then
    echo "=== Slim wide tables for MySQL (avoid Error 1118 row size) ==="
    python3 "$SCRIPT_DIR/scripts/slim_wide_tables_for_mysql_import.py" --data-dir "$DATA_DIR"
fi

TIMESTAMP=$(date +%s)
DS_NAME="worldcup_ds_${TIMESTAMP}"
KN_NAME="worldcup_kn_${TIMESTAMP}"

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

# ── Step 2: Import CSVs + build KN ──────────────────────────────────────────
echo ""
echo "=== Step 2: Import CSVs and build Knowledge Network ==="
echo "  Files: $actual CSV files under data/"

KN_JSON=$(kweaver bkn create-from-csv "$DS_ID" \
    --files "$DATA_DIR/*.csv" \
    --name "$KN_NAME" \
    --table-prefix wc_ \
    --build \
    --timeout 600)
debug_json "create-from-csv" "$KN_JSON"

KN_ID=$(echo "$KN_JSON" | python3 -c \
    "import sys,json; d=json.load(sys.stdin); \
     print(d.get('kn_id') or d.get('id',''))" 2>/dev/null || true)

[ -z "$KN_ID" ] && { echo "Error: create-from-csv failed." >&2; exit 1; }
echo "  Knowledge Network: $KN_ID"

echo "$KN_JSON" | python3 -c "
import sys, json
d = json.load(sys.stdin)
ots = d.get('object_types', [])
if ots:
    print(f'  Auto-discovered object types ({len(ots)}):')
    for ot in ots[:35]:
        print(f'    - {ot.get(\"name\",\"?\")}')
    if len(ots) > 35:
        print(f'    ... and {len(ots) - 35} more')
" 2>/dev/null || true

# ── Step 3: List object types ─────────────────────────────────────────────────
echo ""
echo "=== Step 3: Explore schema ==="

OT_LIST=$(kweaver bkn object-type list "$KN_ID")
debug_json "object-type list" "$OT_LIST"

echo "$OT_LIST" | python3 -c "
import sys, json
data = json.load(sys.stdin)
entries = data.get('entries', data) if isinstance(data, dict) else data
if not isinstance(entries, list):
    entries = []
print(f'  Object types ({len(entries)}):')
for e in sorted(entries, key=lambda x: (x.get('name') or '')):
    name = e.get('name', '?')
    ot_id = e.get('id') or e.get('ot_id') or ''
    print(f'    - {name}  id={ot_id}')
" 2>/dev/null || echo "    (could not parse object-type list)"

# Resolve OT IDs by canonical table basename (handles wc_ prefix on logical names).
resolve_ot_id() {
    local want="$1"
    echo "$OT_LIST" | python3 -c "
import sys, json
want = sys.argv[1].lower().strip()
data = json.load(sys.stdin)
entries = data.get('entries', data) if isinstance(data, dict) else data
if not isinstance(entries, list):
    entries = []
def base(n):
    n = (n or '').lower()
    if n.startswith('wc_'):
        n = n[3:]
    return n
best = ''
for e in entries:
    n = base(e.get('name', ''))
    if n == want:
        print(e.get('id') or e.get('ot_id') or '')
        sys.exit(0)
for e in entries:
    if base(e.get('name', '')).endswith(want):
        print(e.get('id') or e.get('ot_id') or '')
        sys.exit(0)
sys.exit(1)
" "$want" 2>/dev/null || true
}

OT_MATCHES=$(resolve_ot_id matches)
OT_PLAYERS=$(resolve_ot_id players)
OT_GOALS=$(resolve_ot_id goals)

# ── Step 4: Sample instance queries ───────────────────────────────────────────
echo ""
echo "=== Step 4: Query instances (samples) ==="

query_ot() {
    local label="$1" oid="$2"
    [ -z "$oid" ] && echo "  ($label: OT not found — skip)" && return 0
    echo "  --- $label (ot=$oid) ---"
    kweaver bkn object-type query "$KN_ID" "$oid" --limit 5 2>/dev/null | python3 -c "
import sys, json
data = json.load(sys.stdin)
rows = data.get('datas') or data.get('instances') or data.get('entries') or []
for r in rows[:5]:
    items = [(k,v) for k,v in r.items() if not k.startswith('_')][:6]
    print('    ' + '  '.join(f'{k}={v}' for k,v in items))
" 2>/dev/null || echo "    (query failed or empty)"
}

query_ot matches "$OT_MATCHES"
query_ot players "$OT_PLAYERS"
query_ot goals "$OT_GOALS"

# ── Step 5: Semantic schema search (optional) ─────────────────────────────────
echo ""
echo "=== Step 5: Semantic search (context-loader, optional) ==="

kweaver context-loader config set --kn-id "$KN_ID" >/dev/null 2>&1 || true

SEARCH_RESULT=$(kweaver context-loader kn-search "goals tournaments players matches" \
    --kn-id "$KN_ID" --only-schema 2>/dev/null || true)
SCHEMA_RAW=""
if [ -n "$SEARCH_RESULT" ]; then
    SCHEMA_RAW=$(echo "$SEARCH_RESULT" | python3 -c \
        "import sys,json; print(json.load(sys.stdin).get('raw',''))" 2>/dev/null || true)
    echo "  Schema context (first 12 lines):"
    echo "$SCHEMA_RAW" | head -12 | sed 's/^/    /'
    [ "$(echo "$SCHEMA_RAW" | wc -l | tr -d ' ')" -gt 12 ] 2>/dev/null && echo "    ..."
else
    echo "  (context-loader not available on this deployment — continuing)"
fi

# ── Step 6: Agent chat ────────────────────────────────────────────────────────
echo ""
echo "=== Step 6: Agent Q&A ==="

AGENT_ID="${AGENT_ID:-}"
if [ -z "$AGENT_ID" ]; then
    AGENT_ID=$(kweaver agent list --limit 1 2>/dev/null | python3 -c \
        "import sys,json; d=json.load(sys.stdin); \
         print(d[0].get('id','') if isinstance(d,list) and d else '')" 2>/dev/null || true)
fi

if [ -z "$AGENT_ID" ]; then
    echo "  No agent available — set AGENT_ID in .env or import one from help/*/examples/sample-agent.import.json."
    echo "  Skipping chat step."
else
    QUESTION="Across all men's and women's tournaments in this dataset: which player has scored the highest number of goals? Answer with player name(s) only if inferable from the goals and players tables; say if ambiguous."

    GOALS_ROWS=""
    PLAYERS_ROWS=""
    if [ -n "$OT_GOALS" ]; then
        GOALS_ROWS=$(kweaver bkn object-type query "$KN_ID" "$OT_GOALS" --limit 30 2>/dev/null | python3 -c "
import sys, json
data = json.load(sys.stdin)
rows = data.get('datas') or data.get('instances') or data.get('entries') or []
lines = []
for r in rows:
    lines.append('  ' + ', '.join(f'{k}={v}' for k,v in sorted(r.items()) if not str(k).startswith('_')))
print('\n'.join(lines[:30]))
" 2>/dev/null || true)
    fi
    if [ -n "$OT_PLAYERS" ]; then
        PLAYERS_ROWS=$(kweaver bkn object-type query "$KN_ID" "$OT_PLAYERS" --limit 15 2>/dev/null | python3 -c "
import sys, json
data = json.load(sys.stdin)
rows = data.get('datas') or data.get('instances') or data.get('entries') or []
lines = []
for r in rows[:15]:
    lines.append('  ' + ', '.join(f'{k}={v}' for k,v in sorted(r.items()) if not str(k).startswith('_')))
print('\n'.join(lines[:15]))
" 2>/dev/null || true)
    fi

    PROMPT=""
    if [ -n "${SCHEMA_RAW:-}" ]; then
        PROMPT="Schema context from knowledge network search:

${SCHEMA_RAW}

"
    fi
    if [ -n "${GOALS_ROWS:-}" ]; then
        PROMPT="${PROMPT}Sample rows from the goals object type (not exhaustive):
${GOALS_ROWS}

"
    fi
    if [ -n "${PLAYERS_ROWS:-}" ]; then
        PROMPT="${PROMPT}Sample rows from the players object type (not exhaustive):
${PLAYERS_ROWS}

"
    fi
    PROMPT="${PROMPT}Question: ${QUESTION}"

    echo "  Agent: $AGENT_ID"
    echo "  Question (summary): top goal scorer in dataset"
    echo ""
    echo "  Response:"
    kweaver agent chat "$AGENT_ID" \
        -m "$PROMPT" \
        --stream 2>/dev/null | sed 's/^/    /' || \
    kweaver agent chat "$AGENT_ID" \
        -m "$PROMPT" \
        --no-stream 2>/dev/null | sed 's/^/    /' || \
    echo "    (agent chat failed)"
fi

echo ""
echo "=== Example complete ==="
echo "  KN '$KN_NAME' ($KN_ID) and datasource '$DS_NAME' ($DS_ID) will be deleted on exit."
