#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"
require_jq

ACTION="${1:-all}"

run_category() {
  echo "== category =="
  af_curl "$AF_BASE_URL/api/agent-factory/v3/category" | jq .
}

run_product() {
  echo "== product =="
  af_curl "$AF_BASE_URL/api/agent-factory/v3/product" | jq .
}

run_file_ext_map() {
  echo "== temp-zone file ext map =="
  af_curl "$AF_BASE_URL/api/agent-factory/v3/agent/temp-zone/file-ext-map" | jq .
}

run_permission() {
  echo "== permission =="
  af_curl "$AF_BASE_URL/api/agent-factory/v3/agent-permission/management/user-status" | jq .
}

case "$ACTION" in
  category) run_category ;;
  product) run_product ;;
  file-ext-map) run_file_ext_map ;;
  permission) run_permission ;;
  all)
    run_category
    run_product
    run_file_ext_map
    ;;
  *)
    echo "Unknown action: $ACTION" >&2
    exit 2
    ;;
esac
