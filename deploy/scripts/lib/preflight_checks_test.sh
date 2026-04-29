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
# shellcheck source=../services/k3s.sh
source "${SCRIPT_DIR}/scripts/services/k3s.sh"
# shellcheck source=./preflight_checks.sh
source "${SCRIPT_DIR}/scripts/lib/preflight_checks.sh"

# --- kweaver_normalize_kube_distro (from common.sh) ---
[[ "$(kweaver_normalize_kube_distro k8s)" == "kubeadm" ]] || fail "k8s -> kubeadm"
[[ "$(kweaver_normalize_kube_distro kubeadm)" == "kubeadm" ]] || fail "kubeadm alias"
[[ "$(kweaver_normalize_kube_distro k3s)" == "k3s" ]] || fail "k3s"
[[ "$(kweaver_normalize_kube_distro "")" == "k3s" ]] || fail "empty -> k3s default"

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

# --- Onboard helpers: byte-compile (syntax-only; matches CPython 3.6+ subset) ---
ONBOARD_PY=(
  "${SCRIPT_DIR}/scripts/lib/onboard_patch_bkn_cm.py"
  "${SCRIPT_DIR}/scripts/lib/onboard_apply_config.py"
)
_onboard_py_compile() {
  local py="$1"
  echo "== onboard py_compile: ${py} ($("${py}" -V 2>&1)) =="
  "${py}" -m py_compile "${ONBOARD_PY[@]}" || fail "py_compile failed for ${py}"
}
if command -v python3 &>/dev/null; then
  _onboard_py_compile python3
  for py in ${EXTRA_PYTHONS:-}; do
    [[ -z "${py}" ]] && continue
    if command -v "${py}" &>/dev/null; then
      _onboard_py_compile "${py}"
    else
      echo "(skip: ${py} not on PATH)" >&2
    fi
  done
else
  echo "skip: onboard py_compile (python3 not in PATH)"
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
