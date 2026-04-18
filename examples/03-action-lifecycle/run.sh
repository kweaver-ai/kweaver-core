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
