#!/usr/bin/env bash
# =============================================================================
# Branch flow: ds connect → import CSV into MySQL (ds import-csv) → create-from-ds
#            → bkn pull → validate / push → optional Agent / chat.
# Requires: kweaver CLI、本机 ./data 下 27 份 CSV、平台可达的 MySQL + 平台 API。
# 「灌库」= 本条线的 import-csv（不是 rsync 文件）；需单独备份 CSV 时用 ./upload-data.sh。
# =============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DATA_DIR="$SCRIPT_DIR/data"

usage() {
    cat <<EOF
Usage: $(basename "$0") [options]

Options:
  -h, --help     Show help.
  --dry-run      Print planned steps only (no kweaver side effects beyond --help).

Env:
  Single file: ./.env (see env.sample — DB + optional branch fields)

Typical:
  cp env.sample .env && vim .env
  ./download.sh
  ./run-branch-bkn.sh
EOF
}

while [ $# -gt 0 ]; do
    case "$1" in
        -h|--help) usage; exit 0 ;;
        --dry-run)
            cat <<'PLAN'
Planned steps (no kweaver calls):
  1. ensure ./data has 27 CSVs (./download.sh)
  2. slim matches.csv + team_appearances.csv unless SLIM_WIDE_CSV_FOR_MYSQL=0 (MySQL Error 1118 / VARCHAR-per-column limit)
  3. kweaver ds connect mysql → DS_ID
  4. kweaver ds tables DS_ID (optional snapshot before import)
  5. kweaver ds import-csv DS_ID --files ./data/*.csv → MySQL (--table-prefix wc_ --recreate)
  6. kweaver ds tables DS_ID again after import
  7. kweaver bkn create-from-ds … --tables "<DB>.wc_<each_stem>,…" --build
  8. kweaver bkn pull KN_ID → BKN_EXPORT_DIR
  9. kweaver bkn validate + optional bkn push
  10. optional: agent create → update --knowledge-network-id → publish → chat

See env.sample for DB_* / BKN_* / AGENT_* (REMOTE_* only if you separately run ./upload-data.sh).
PLAN
            exit 0
            ;;
        *) echo "Unknown: $1" >&2; usage >&2; exit 2 ;;
    esac
done

[ -f "$SCRIPT_DIR/.env" ] && source "$SCRIPT_DIR/.env"

KN_NAME_BRANCH="${KN_NAME_BRANCH:-worldcup_bkn_branch}"
BKN_EXPORT_DIR="${BKN_EXPORT_DIR:-$SCRIPT_DIR/bkn-export/worldcup}"
DO_BKN_PUSH="${DO_BKN_PUSH:-true}"
SKIP_AGENT="${SKIP_AGENT:-false}"
AGENT_NAME="${AGENT_NAME:-世界杯数据分析助手}"
AGENT_PROFILE="${AGENT_PROFILE:-基于世界杯知识网络回答问题}"
BRANCH_CHAT_QUESTION="${BRANCH_CHAT_QUESTION:-哪位球员累计进球最多？}"

DB_HOST="${DB_HOST:?Set DB_HOST in .env}"
DB_PORT="${DB_PORT:-3306}"
DB_NAME="${DB_NAME:?Set DB_NAME in .env}"
DB_USER="${DB_USER:?Set DB_USER in .env}"
DB_PASS="${DB_PASS:?Set DB_PASS in .env}"

build_qualified_tables() {
    local db="$1"
    local out="" sep=""
    local f base
    shopt -s nullglob
    for f in "$DATA_DIR"/*.csv; do
        base=$(basename "$f" .csv)
        out+="${sep}${db}.wc_${base}"
        sep=","
    done
    shopt -u nullglob
    if [ -z "$out" ]; then
        echo "No CSV files in $DATA_DIR" >&2
        return 1
    fi
    printf '%s' "$out"
}

EXPECTED_CSV=27
count_csvs() {
    shopt -s nullglob
    local a=("$DATA_DIR"/*.csv)
    echo "${#a[@]}"
}

echo "=== Prerequisites: CSVs ==="
if [ ! -d "$DATA_DIR" ] || [ "$(count_csvs)" -lt "$EXPECTED_CSV" ]; then
    echo "  Running download.sh (need $EXPECTED_CSV CSVs)."
    "$SCRIPT_DIR/download.sh"
fi
if [ "$(count_csvs)" -lt "$EXPECTED_CSV" ]; then
    echo "Error: need at least $EXPECTED_CSV CSVs in $DATA_DIR" >&2
    exit 1
fi

SLIM_WIDE_CSV_FOR_MYSQL="${SLIM_WIDE_CSV_FOR_MYSQL:-1}"
if [ "$SLIM_WIDE_CSV_FOR_MYSQL" = 1 ] || [ "$SLIM_WIDE_CSV_FOR_MYSQL" = true ]; then
    echo "=== Slim wide tables for MySQL (import-csv uses VARCHAR per column → Error 1118) ==="
    python3 "$SCRIPT_DIR/scripts/slim_wide_tables_for_mysql_import.py" --data-dir "$DATA_DIR"
else
    echo "  SLIM_WIDE_CSV_FOR_MYSQL=0 — skipping slim (may fail wc_matches / wc_team_appearances)."
fi

# ── 1. Connect DS ─────────────────────────────────────────────────────────────
TIMESTAMP=$(date +%s)
DS_NAME="worldcup_branch_ds_${TIMESTAMP}"

echo ""
echo "=== Step 1: Connect datasource ($DS_NAME) ==="
echo "  $DB_HOST:$DB_PORT/$DB_NAME"

KN_ID=""
DS_ID=""
cleanup() {
    [ -z "$KN_ID" ] && [ -z "$DS_ID" ] && return 0
    echo ""
    echo "=== Teardown (TEARDOWN_BRANCH=1) ==="
    [ -n "$KN_ID" ] && kweaver bkn delete "$KN_ID" -y 2>/dev/null && echo "  Deleted KN $KN_ID" || true
    [ -n "$DS_ID" ] && kweaver ds delete "$DS_ID" -y 2>/dev/null && echo "  Deleted DS $DS_ID" || true
}
if [ "${TEARDOWN_BRANCH:-}" = 1 ]; then
    trap cleanup EXIT
fi

DS_JSON=$(kweaver ds connect mysql "$DB_HOST" "$DB_PORT" "$DB_NAME" \
    --account "$DB_USER" --password "$DB_PASS" --name "$DS_NAME")

DS_ID=$(echo "$DS_JSON" | python3 -c \
    "import sys,json; d=json.load(sys.stdin); r=d[0] if isinstance(d,list) else d; print(r.get('datasource_id') or r.get('id',''))" 2>/dev/null || true)
[ -z "$DS_ID" ] && { echo "Error: ds connect failed." >&2; exit 1; }
echo "  DS_ID=$DS_ID"

# ── 2. Scan tables (after import; first pass shows pre-import state) ───────────
echo ""
echo "=== Step 2: Scan datasource (ds tables — before CSV import snapshot) ==="
kweaver ds tables "$DS_ID" 2>/dev/null | head -40 || true
echo "  (full list may be long; re-run 'kweaver ds tables $DS_ID' after import)"

# ── 3. Import CSVs (platform writes into remote MySQL) ─────────────────────────
echo ""
echo "=== Step 3: Import CSVs into datasource (wc_ prefix) ==="
kweaver ds import-csv "$DS_ID" --files "$DATA_DIR/*.csv" --table-prefix wc_ --recreate

echo ""
echo "=== Step 2b: Scan again (after import) ==="
kweaver ds tables "$DS_ID" --keyword wc_ 2>/dev/null | head -50 || kweaver ds tables "$DS_ID" | head -50

# ── 4. create-from-ds + pull → local .bkn files ───────────────────────────────
TABLES_LIST=$(build_qualified_tables "$DB_NAME")
echo ""
echo "=== Step 4: create-from-ds (all $EXPECTED_CSV logical tables) ==="
echo "  --tables length: ${#TABLES_LIST} chars (prefix list worldcup.wc_*)"

KN_JSON=$(kweaver bkn create-from-ds "$DS_ID" \
    --name "$KN_NAME_BRANCH" \
    --tables "$TABLES_LIST" \
    --build \
    --timeout 600)

KN_ID=$(echo "$KN_JSON" | python3 -c \
    "import sys,json; d=json.load(sys.stdin); print(d.get('kn_id') or d.get('id',''))" 2>/dev/null || true)
[ -z "$KN_ID" ] && { echo "Error: create-from-ds failed." >&2; exit 1; }
echo "  KN_ID=$KN_ID"

echo ""
echo "=== Step 5: Pull BKN files to $BKN_EXPORT_DIR ==="
mkdir -p "$BKN_EXPORT_DIR"
kweaver bkn pull "$KN_ID" "$BKN_EXPORT_DIR"

echo ""
echo "=== Step 6: Validate (and optional push) exported BKN ==="
kweaver bkn validate "$BKN_EXPORT_DIR"
if [ "$DO_BKN_PUSH" = true ] || [ "$DO_BKN_PUSH" = 1 ]; then
    echo "  bkn push $BKN_EXPORT_DIR"
    kweaver bkn push "$BKN_EXPORT_DIR"
else
    echo "  DO_BKN_PUSH=false — skipped push."
fi

# ── 7–8: Agent ─────────────────────────────────────────────────────────────────
echo ""
echo "=== Step 7–8: Agent (optional) ==="
AGENT_ID=""
if [ "$SKIP_AGENT" = true ] || [ "$SKIP_AGENT" = 1 ]; then
    echo "  SKIP_AGENT=true — skipping."
elif [ -z "${AGENT_LLM_ID:-}" ]; then
    echo "  Set AGENT_LLM_ID in .env to create/publish an agent; skipped."
else
    AGENT_JSON=$(kweaver agent create \
        --name "$AGENT_NAME" \
        --profile "$AGENT_PROFILE" \
        --llm-id "$AGENT_LLM_ID")
    AGENT_ID=$(echo "$AGENT_JSON" | python3 -c \
        "import sys,json; d=json.load(sys.stdin); print(d.get('id') or d.get('agent_id',''))" 2>/dev/null || true)
    [ -z "$AGENT_ID" ] && { echo "Error: agent create failed." >&2; exit 1; }
    echo "  AGENT_ID=$AGENT_ID"
    kweaver agent update "$AGENT_ID" --knowledge-network-id "$KN_ID"
    kweaver agent publish "$AGENT_ID"
    echo ""
    echo "=== Sample chat ==="
    kweaver agent chat "$AGENT_ID" -m "$BRANCH_CHAT_QUESTION" --no-stream 2>/dev/null | sed 's/^/  /' || \
    kweaver agent chat "$AGENT_ID" -m "$BRANCH_CHAT_QUESTION" --stream 2>/dev/null | sed 's/^/  /' || true
fi

echo ""
echo "=== Branch flow complete ==="
echo "  KN_ID=$KN_ID"
echo "  BKN export: $BKN_EXPORT_DIR"
echo "  DS_ID=$DS_ID"
echo "  (默认保留 KN/DS；脚本退出时若要自动删除 KN+DS，可加 TEARDOWN_BRANCH=1)"
if [ -n "$AGENT_ID" ]; then
    echo "  AGENT_ID=$AGENT_ID"
fi
