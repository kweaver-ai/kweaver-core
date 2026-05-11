#!/usr/bin/env bash
# SSH to REMOTE_HOST (default DB_HOST), run mysql locally on that VM to DROP wc_matches / wc_team_appearances.
# Sources ../.env for DB_* and optional REMOTE_HOST.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=/dev/null
set -a
source "$ROOT/.env"
set +a

: "${DB_HOST:?}"
: "${DB_USER:?}"
: "${DB_PASS:?}"
: "${DB_NAME:?}"
DB_PORT="${DB_PORT:-3306}"
RH="${REMOTE_HOST:-$DB_HOST}"

PWQ=$(printf '%q' "$DB_PASS")

echo "=== SSH ubuntu@${RH}: DROP wc_matches, wc_team_appearances in ${DB_NAME} ==="
# Remote expands PWQ from literal we pass — pass password as base64 to avoid quote hell
B64=$(printf '%s' "$DB_PASS" | base64 | tr -d '\n')
ssh -o BatchMode=yes -o ConnectTimeout=25 "ubuntu@${RH}" bash -s -- "$DB_PORT" "$DB_USER" "$DB_NAME" "$B64" <<'EOS'
set -euo pipefail
DB_PORT="$1"
DB_USER="$2"
DB_NAME="$3"
B64="$4"
export MYSQL_PWD="$(printf '%s' "$B64" | base64 -d)"
mysql -h 127.0.0.1 -P "$DB_PORT" -u "$DB_USER" "$DB_NAME" -e "
DROP TABLE IF EXISTS wc_matches;
DROP TABLE IF EXISTS wc_team_appearances;
"
echo "Dropped (if existed)."
EOS

echo "=== Next on laptop: slim CSV then re-import ==="
echo "    python3 $ROOT/scripts/slim_wide_tables_for_mysql_import.py --data-dir $ROOT/data"
echo "    cd $ROOT && kweaver ds import-csv <DS_ID> --files \"./data/*.csv\" --table-prefix wc_ --recreate"
echo "    (or run ./run.sh which slims + create-from-csv)"
