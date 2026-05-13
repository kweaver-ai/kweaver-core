#!/usr/bin/env bash
set -euo pipefail

AF_BASE_URL="${AF_BASE_URL:-${KWEAVER_BASE_URL:-http://127.0.0.1:13020}}"
AF_BD="${AF_BD:-${KWEAVER_BUSINESS_DOMAIN:-bd_public}}"
AF_TOKEN="${AF_TOKEN:-${KWEAVER_TOKEN:-__NO_AUTH__}}"

af_curl() {
  if [[ "${KWEAVER_NO_AUTH:-}" == "1" || "$AF_TOKEN" == "__NO_AUTH__" ]]; then
    curl -fsS -H "X-Business-Domain: $AF_BD" -H "X-Language: zh-cn" "$@"
  else
    curl -fsS \
      -H "Authorization: Bearer $AF_TOKEN" \
      -H "Token: $AF_TOKEN" \
      -H "X-Business-Domain: $AF_BD" \
      -H "X-Language: zh-cn" \
      "$@"
  fi
}

require_jq() {
  if ! command -v jq >/dev/null 2>&1; then
    echo "jq is required for this example." >&2
    exit 1
  fi
}

