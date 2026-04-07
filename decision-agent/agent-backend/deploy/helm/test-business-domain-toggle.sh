#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
chart_dir="${script_dir}/agent-backend"

render_disable_biz_domain() {
  local enabled="$1"

  helm template agent-factory "${chart_dir}" --set "businessDomain.enabled=${enabled}" \
    | awk '/disable_biz_domain:/ { print $2; exit }'
}

render_sandbox_platform_enable() {
  local enabled="$1"

  helm template agent-factory "${chart_dir}" --set "depServices.sandboxPlatform.enable=${enabled}" \
    | awk '
      /sandbox_platform:/ { in_block=1; next }
      in_block && /enable:/ { print $2; exit }
    '
}

enabled_true_value="$(render_disable_biz_domain true)"
enabled_false_value="$(render_disable_biz_domain false)"
sandbox_enabled_true_value="$(render_sandbox_platform_enable true)"
sandbox_enabled_false_value="$(render_sandbox_platform_enable false)"

if [[ "${enabled_true_value}" != "false" ]]; then
  echo "Expected disable_biz_domain=false when businessDomain.enabled=true, got ${enabled_true_value}"
  exit 1
fi

if [[ "${enabled_false_value}" != "true" ]]; then
  echo "Expected disable_biz_domain=true when businessDomain.enabled=false, got ${enabled_false_value}"
  exit 1
fi

if [[ "${sandbox_enabled_true_value}" != "true" ]]; then
  echo "Expected sandbox_platform.enable=true when depServices.sandboxPlatform.enable=true, got ${sandbox_enabled_true_value}"
  exit 1
fi

if [[ "${sandbox_enabled_false_value}" != "false" ]]; then
  echo "Expected sandbox_platform.enable=false when depServices.sandboxPlatform.enable=false, got ${sandbox_enabled_false_value}"
  exit 1
fi

echo "businessDomain and sandboxPlatform toggles render correctly"
