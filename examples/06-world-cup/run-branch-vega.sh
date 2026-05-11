#!/usr/bin/env bash
# =============================================================================
# Vega branch helper: Vega Catalog → discover --wait (--dry-run exits after plan).
# Does NOT bulk-create BKN OTs (PK/display-key mapping is per-table).
# Reuses DB_* from .env for mysql connector-config when VEGA_* not set separately.
# Requires: jq; kweaver CLI (`npm i -g @kweaver-ai/kweaver-sdk`; put Node kweaver before any broken `/usr/local/bin/kweaver` Python stub).
# Vega backend must reach the MySQL host in connector-config from its own network (often not identical to your laptop's DB_HOST).
# =============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

usage() {
    printf 'Usage: %s [options]\n' "$(basename "$0")"
    cat <<'EOF'
Options:
  -h, --help     Show help.
  --dry-run      Print planned steps only.

Env (.env via env.sample Vega section):
  VEGA_CATALOG_NAME   Catalog display name (required for create).
  VEGA_SKIP_CREATE=1 Skip catalog create if you already set VEGA_CATALOG_ID.
  VEGA_CATALOG_ID     Existing catalog UUID (skip create).

  MYSQL connector defaults from DB_HOST DB_PORT DB_NAME DB_USER DB_PASS unless:
  VEGA_MYSQL_HOST VEGA_MYSQL_PORT VEGA_MYSQL_USER VEGA_MYSQL_PASS VEGA_MYSQL_DATABASES
    (JSON array string, default ["DB_NAME"])

After success: run manual steps from WORKFLOW-BRANCH-VEGA.zh.md / vega catalog resources / bkn object-type create …
EOF
}

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

DRY_RUN="${DRY_RUN:-}"

if [ -z "${KWEAVER_BASE_URL:-}" ]; then
    KWEAV=(kweaver)
else
    KWEAV=(kweaver --base-url "${KWEAVER_BASE_URL}")
fi

if [ "${VEGA_SKIP_CREATE:-0}" = "1" ]; then
    if [ -z "${VEGA_CATALOG_ID:-}" ]; then
        echo "VEGA_SKIP_CREATE=1 requires VEGA_CATALOG_ID" >&2
        exit 2
    fi
elif [ -z "${VEGA_CATALOG_NAME:-}" ]; then
    echo "Missing VEGA_CATALOG_NAME (or set VEGA_SKIP_CREATE=1 and VEGA_CATALOG_ID)" >&2
    exit 2
fi

plan() {
    if [ "${VEGA_SKIP_CREATE:-0}" = "1" ]; then
        s1="1) Vega catalog: skip create (VEGA_SKIP_CREATE=1)"
    else
        s1="1) vega catalog create --name \"$VEGA_CATALOG_NAME\""
    fi
    cat <<EOF
Plan:
  ${s1}
  2) vega catalog discover <id> --wait
  3) Then: WORKFLOW-BRANCH-VEGA.zh.md §3–§4 (resources + bkn object-type create …)
EOF
}

if [ -n "${DRY_RUN:-}" ]; then
    plan
    echo "Dry-run only."
    exit 0
fi

CONN_JSON=""
if [ "${VEGA_SKIP_CREATE:-0}" != "1" ]; then
    VEGA_MYSQL_HOST="${VEGA_MYSQL_HOST:-${DB_HOST:?Set DB_HOST in .env}}"
    VEGA_MYSQL_PORT="${VEGA_MYSQL_PORT:-${DB_PORT:-3306}}"
    VEGA_MYSQL_USER="${VEGA_MYSQL_USER:-${DB_USER:?Set DB_USER in .env}}"
    VEGA_MYSQL_PASS="${VEGA_MYSQL_PASS:-${DB_PASS:?Set DB_PASS in .env}}"
    VEGA_DATABASES_JSON="${VEGA_MYSQL_DATABASES:-}"
    if [ -z "$VEGA_DATABASES_JSON" ]; then
        VEGA_DATABASES_JSON="$(jq -cn --arg d "${DB_NAME:?Set DB_NAME in .env}" '[$d]')"
    fi
    MYSQL_PORT_NUM=$((10#${VEGA_MYSQL_PORT:-3306}))
    CONN_JSON="$(jq -cn \
        --arg host "$VEGA_MYSQL_HOST" \
        --argjson port "$MYSQL_PORT_NUM" \
        --arg username "$VEGA_MYSQL_USER" \
        --arg password "$VEGA_MYSQL_PASS" \
        --argjson databases "$VEGA_DATABASES_JSON" \
        '{host:$host,port:$port,username:$username,password:$password,databases:$databases}')"
fi

resolve_catalog_id() {
    local raw
    raw="$("${KWEAV[@]}" vega catalog list --limit 200 2>/dev/null || true)"
    printf '%s' "$raw" | jq -r --arg n "$VEGA_CATALOG_NAME" '
        (.entries // [])[] | select(.name == $n) | .id' | head -1
}

CATALOG_ID="${VEGA_CATALOG_ID:-}"

_create_catalog() {
    local stderr_log out rc
    stderr_log="$(mktemp)"
    rc=0
    out="$("${KWEAV[@]}" vega catalog create \
        --name "$VEGA_CATALOG_NAME" \
        --connector-type mysql \
        --connector-config "$CONN_JSON" 2>"$stderr_log")" || rc=$?
    if [ "$rc" -ne 0 ] || [ -z "$out" ]; then
        if [ -s "$stderr_log" ]; then
            echo "vega catalog create stderr:" >&2
            sed 's/^/  /' "$stderr_log" >&2
        fi
        if [ -n "$out" ]; then
            echo "vega catalog create stdout:" >&2
            sed 's/^/  /' <<<"$out" >&2
        fi
    fi
    rm -f "$stderr_log"
    printf '%s' "$out"
}

if [ "${VEGA_SKIP_CREATE:-0}" != "1" ]; then
    echo "Creating Vega catalog: $VEGA_CATALOG_NAME" >&2
    CREATE_OUT="$(_create_catalog)"
    CATALOG_ID="$(printf '%s' "$CREATE_OUT" | jq -r '.id // .data.id // empty' 2>/dev/null | head -1)"
    if [ -z "$CATALOG_ID" ]; then
        CATALOG_ID="$(resolve_catalog_id)"
        if [ -n "$CATALOG_ID" ]; then
            echo "Create returned no id (name may already exist) — using existing catalog: $CATALOG_ID" >&2
        fi
    fi
fi

if [ -z "$CATALOG_ID" ]; then
    echo "Could not determine catalog id. Set VEGA_CATALOG_ID or check catalog list." >&2
    exit 1
fi

echo "Using catalog id: $CATALOG_ID" >&2
echo "Running discover (--wait)…" >&2
if ! "${KWEAV[@]}" vega catalog discover "$CATALOG_ID" --wait; then
    echo "Discover failed — see WORKFLOW-BRANCH-VEGA.zh.md for manual resource create." >&2
    exit 1
fi

cat <<EOF

Next commands (replace placeholders):

  ${KWEAV[*]} vega catalog resources "$CATALOG_ID" --category table --limit 200
  ${KWEAV[*]} bkn create --name "<kn-name>"
  ${KWEAV[*]} bkn object-type create <kn_id> --name matches --dataview-id <resource-uuid-for-wc_matches> --primary-key <pk> --display-key <dk>
  … repeat per table stem in scripts/worldcup_dataset_stems.inc.sh …
EOF
