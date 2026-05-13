#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
EXAMPLE_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

export PATH="$EXAMPLE_DIR/node_modules/.bin:$PATH"
export KWEAVER_BASE_URL="${KWEAVER_BASE_URL:-http://127.0.0.1:13020}"
export KWEAVER_NO_AUTH="${KWEAVER_NO_AUTH:-1}"

require_kweaver() {
  if ! command -v kweaver >/dev/null 2>&1; then
    echo "kweaver command is required. Run make install in this directory first." >&2
    exit 1
  fi
}

require_jq() {
  if ! command -v jq >/dev/null 2>&1; then
    echo "jq is required for this example." >&2
    exit 1
  fi
}

skip_if_empty() {
  local var_name="$1"
  local message="$2"
  if [[ -z "${!var_name:-}" ]]; then
    echo "Skip: $message"
    exit 0
  fi
}
