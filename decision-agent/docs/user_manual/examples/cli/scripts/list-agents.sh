#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"
require_kweaver

ACTION="${1:-list}"

case "$ACTION" in
  list)
    kweaver agent list --limit "${LIMIT:-1}" --pretty
    ;;
  personal-list)
    kweaver agent personal-list --size "${LIMIT:-1}" --pretty
    ;;
  category-list)
    kweaver agent category-list --pretty
    ;;
  template-list)
    kweaver agent template-list --size "${LIMIT:-1}" --pretty
    ;;
  *)
    echo "Unknown action: $ACTION" >&2
    exit 2
    ;;
esac
