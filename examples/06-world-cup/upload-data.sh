#!/usr/bin/env bash
# Optional: rsync ./data/*.csv → REMOTE_* （仅备份/运维；不向 MySQL 灌数据）。
# For loading MySQL, use ./run.sh (create-from-csv) or run kweaver ds import-csv yourself.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DATA_DIR="$SCRIPT_DIR/data"

[ -f "$SCRIPT_DIR/.env" ] && source "$SCRIPT_DIR/.env"
if [ -z "${REMOTE_USER:-}" ] || [ -z "${REMOTE_HOST:-}" ] || [ -z "${REMOTE_DIR:-}" ]; then
    echo "Set REMOTE_USER, REMOTE_HOST, REMOTE_DIR in .env (see env.sample) or export before calling." >&2
    exit 1
fi

if [ ! -d "$DATA_DIR" ] || [ -z "$(ls -A "$DATA_DIR"/*.csv 2>/dev/null || true)" ]; then
    echo "Missing CSVs under $DATA_DIR — run ./download.sh first." >&2
    exit 1
fi

echo "rsync → ${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/"
ssh "${REMOTE_USER}@${REMOTE_HOST}" "mkdir -p '${REMOTE_DIR}'"
rsync -avz --delete --exclude '*.tmp' "${DATA_DIR}/" "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/"
echo Done.
