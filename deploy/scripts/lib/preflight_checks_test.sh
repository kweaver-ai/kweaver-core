#!/usr/bin/env bash
# Lightweight tests for preflight helpers (no cluster required)
set -euo pipefail
ONE_FAILED=0
fail() { echo "FAIL: $*"; ONE_FAILED=1; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
# shellcheck source=../lib/common.sh
source "${SCRIPT_DIR}/scripts/lib/common.sh"
# shellcheck source=../services/k8s.sh
source "${SCRIPT_DIR}/scripts/services/k8s.sh"
# shellcheck source=./preflight_checks.sh
source "${SCRIPT_DIR}/scripts/lib/preflight_checks.sh"

# --- Test resolve minor ---
out="$(PREFLIGHT_K8S_APT_MINOR=  preflight_resolve_k8s_apt_minor)"
[[ "${out}" =~ ^v[0-9]+\.[0-9]+$ ]] || fail "resolve default should be vM.N, got ${out}"

# --- Test JSON emit (empty) ---
PREFLIGHT_JSON_OK=()
PREFLIGHT_JSON_WARN=()
PREFLIGHT_JSON_FAIL=()
PREFLIGHT_JSON_FIXED=()
PREFLIGHT_JSON_DECLINED=()
json="$(emit_preflight_json 2>/dev/null)"
if command -v python3 &>/dev/null; then
  echo "$json" | python3 -c "import json,sys; d=json.load(sys.stdin); assert not d['ok'] and not d['warn'] and not d['fail']" || fail "empty json roundtrip"
else
  echo "skip: python3 not in PATH"
fi

# --- Test confirm: assume yes ---
PREFLIGHT_ASSUME_YES=true PREFLIGHT_ASSUME_NO=false preflight_confirm_fix "t" "a" "r" && true || fail "expect yes with ASSUME_YES"
PREFLIGHT_ASSUME_YES=false
PREFLIGHT_ASSUME_NO=true preflight_confirm_fix "t" "a" "r" && fail "expect no with ASSUME_NO" || true

# --- Test PREFLIGHT_FIX_ALLOW allowlist ---
PREFLIGHT_ASSUME_NO=false
PREFLIGHT_FIX_ALLOW="|t|"
PREFLIGHT_ASSUME_YES=false preflight_confirm_fix "t" "a" "r" && true || fail "allowlist t"
PREFLIGHT_FIX_ALLOW="|other|"
PREFLIGHT_ASSUME_NO=false PREFLIGHT_ASSUME_YES=false preflight_confirm_fix "t" "a" "r" && fail "not in allowlist" || true
PREFLIGHT_FIX_ALLOW=""

if [[ ${ONE_FAILED} -ne 0 ]]; then
  exit 1
fi
echo "OK preflight_checks_test.sh"
