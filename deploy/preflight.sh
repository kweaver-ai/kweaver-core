#!/usr/bin/env bash
# KWeaver Core — pre-install environment check and safe fixes
# See help/zh/install.md. Run on the target Linux host (often as root for fixes).

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export PREFLIGHT_ROOT="${PREFLIGHT_ROOT:-${SCRIPT_DIR}}"
# shellcheck source=scripts/lib/common.sh
source "${SCRIPT_DIR}/scripts/lib/common.sh"
# shellcheck source=scripts/services/k8s.sh
source "${SCRIPT_DIR}/scripts/services/k8s.sh"
# shellcheck source=scripts/lib/preflight_checks.sh
source "${SCRIPT_DIR}/scripts/lib/preflight_checks.sh"

PREFLIGHT_CHECK_ONLY="false"
PREFLIGHT_REPORT_FILE=""
PREFLIGHT_SKIP_SET="|"
PREFLIGHT_ASSUME_YES="false"
PREFLIGHT_ASSUME_NO="false"
PREFLIGHT_FIX_ALLOW=""
PREFLIGHT_OUTPUT_JSON="false"
PREFLIGHT_ROLE="both"
PREFLIGHT_LIST_FIXES_ONLY="false"
PREFLIGHT_NO_RECHECK="false"

usage() {
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  -h, --help           Show this help"
    echo "  --check-only         Only run checks, do not modify the system (no root required for partial checks)"
    echo "  --fix                Check + interactively apply fixes (default; requires root for fixes)"
    echo "  -y, --yes            Auto-approve every fix (skip per-fix y/N prompt)"
    echo "  -n, --no             Auto-decline every fix (preview risk text, change nothing)"
    echo "  --fix-allow=LIST     Comma-separated fix names to auto-approve (others are skipped)."
    echo "                       Names: k3s-uninstall,kubeadm-reset,k8s-apt-source,containerd-install,helm-v3,"
    echo "                       chrony,firewalld,ufw,selinux,system-tuning,bridge-sysctl,kernel-limits,iptables-legacy,etc-hosts"
    echo "  --list-fixes         Run checks then list fixes that would be offered (no changes; non-root OK)"
    echo "  --output=json        Emit JSON summary to stdout (human logs to stderr); requires python3"
    echo "  --role=target|admin|both  Target = kubectl/helm only; admin = kweaver/node; both = all (default)"
    echo "  --no-recheck         Do not re-run full checks after applying fixes"
    echo "  --report=PATH        Append full log to a file"
    echo "  --skip=LIST          Comma-separated check names to skip (see source: preflight_checks.sh preflight_skip)"
    echo ""
    echo "Environment: PREFLIGHT_ROOT=path/to/deploy  PREFLIGHT_CONFIG_YAML=...  PREFLIGHT_K8S_APT_MINOR=vX.YY"
    echo ""
    echo "Exit codes: 0 = OK, 1 = FAIL present, 2 = only WARN (no FAIL)"
    echo ""
    echo "Examples:"
    echo "  sudo $0"
    echo "  $0 --check-only"
    echo "  $0 --list-fixes"
    echo "  $0 --skip=network --report=/tmp/preflight.txt"
}

# Parse args
while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)
            usage
            exit 0
            ;;
        --check-only)
            PREFLIGHT_CHECK_ONLY="true"
            shift
            ;;
        --fix)
            PREFLIGHT_CHECK_ONLY="false"
            shift
            ;;
        -y|--yes)
            PREFLIGHT_ASSUME_YES="true"
            shift
            ;;
        -n|--no)
            PREFLIGHT_ASSUME_NO="true"
            shift
            ;;
        --list-fixes)
            PREFLIGHT_LIST_FIXES_ONLY="true"
            shift
            ;;
        --output=json)
            PREFLIGHT_OUTPUT_JSON="true"
            shift
            ;;
        --no-recheck)
            PREFLIGHT_NO_RECHECK="true"
            shift
            ;;
        --role=*)
            PREFLIGHT_ROLE="${1#*=}"
            shift
            ;;
        --fix-allow=*)
            IFS=',' read -r -a _fa <<< "${1#*=}"
            PREFLIGHT_FIX_ALLOW="|"
            for s in "${_fa[@]}"; do
                s="${s#"${s%%[![:space:]]*}"}"
                s="${s%"${s##*[![:space:]]}"}"
                s="$(printf '%s' "$s" | tr '[:upper:]' '[:lower:]')"
                [[ -n "${s}" ]] && PREFLIGHT_FIX_ALLOW+="${s}|"
            done
            shift
            ;;
        --report=*)
            PREFLIGHT_REPORT_FILE="${1#*=}"
            shift
            ;;
        --skip=*)
            IFS=',' read -r -a _sk <<< "${1#*=}"
            for s in "${_sk[@]}"; do
                s="${s#"${s%%[![:space:]]*}"}"
                s="${s%"${s##*[![:space:]]}"}"
                s="$(printf '%s' "$s" | tr '[:upper:]' '[:lower:]')"
                PREFLIGHT_SKIP_SET+="${s}|"
            done
            shift
            ;;
        *)
            log_error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

export PREFLIGHT_CHECK_ONLY PREFLIGHT_REPORT_FILE PREFLIGHT_SKIP_SET
export PREFLIGHT_ASSUME_YES PREFLIGHT_ASSUME_NO PREFLIGHT_FIX_ALLOW
export PREFLIGHT_OUTPUT_JSON PREFLIGHT_ROLE PREFLIGHT_LIST_FIXES_ONLY PREFLIGHT_NO_RECHECK PREFLIGHT_ROOT

if [[ -n "${PREFLIGHT_REPORT_FILE}" ]]; then
    mkdir -p "$(dirname "${PREFLIGHT_REPORT_FILE}")" 2>/dev/null || true
    {
        echo "=== KWeaver preflight $(date -Iseconds) ==="
    } > "${PREFLIGHT_REPORT_FILE}"
fi

if [[ "${PREFLIGHT_CHECK_ONLY}" != "true" && "${PREFLIGHT_LIST_FIXES_ONLY}" != "true" ]]; then
    if [[ "${EUID}" -ne 0 ]]; then
        log_error "For automatic fixes, run as root: sudo $0 (or use --check-only / --list-fixes)"
        log_info "Falling back to read-only check (--check-only) …"
        PREFLIGHT_CHECK_ONLY="true"
    fi
fi

if [[ "${PREFLIGHT_CHECK_ONLY}" != "true" && "${PREFLIGHT_LIST_FIXES_ONLY}" != "true" ]]; then
    check_root
fi

preflight_reset_counters

if [[ "${PREFLIGHT_OUTPUT_JSON}" == "true" ]]; then
    log_info "========== KWeaver preflight checks ==========" >&2
else
    log_info "========== KWeaver preflight checks =========="
fi
preflight_run_all_checks
PREFLIGHT_FAIL_COUNT_INITIAL="${PREFLIGHT_FAIL_COUNT}"
export PREFLIGHT_FAIL_COUNT_INITIAL

if [[ "${PREFLIGHT_CHECK_ONLY}" != "true" || "${PREFLIGHT_LIST_FIXES_ONLY}" == "true" ]]; then
    if [[ "${PREFLIGHT_OUTPUT_JSON}" == "true" ]]; then
        log_info "========== Safe fixes ==========" >&2
    else
        log_info "========== Safe fixes =========="
    fi
    preflight_apply_safe_fixes
    if [[ "${PREFLIGHT_NO_RECHECK}" != "true" && "${PREFLIGHT_CHECK_ONLY}" != "true" && "${PREFLIGHT_LIST_FIXES_ONLY}" != "true" ]]; then
        preflight_recheck_after_fixes
    fi
fi

if [[ "${PREFLIGHT_OUTPUT_JSON}" == "true" ]]; then
    log_info "========== Summary ==========" >&2
    echo "  [OK]    ${PREFLIGHT_OK_COUNT}" >&2
    echo "  [WARN]  ${PREFLIGHT_WARN_COUNT}" >&2
    echo "  [FAIL]  ${PREFLIGHT_FAIL_COUNT}" >&2
    echo "  [FIXED] ${PREFLIGHT_FIXED_COUNT}" >&2
    emit_preflight_json
else
    log_info "========== Summary =========="
    echo "  [OK]    ${PREFLIGHT_OK_COUNT}"
    echo "  [WARN]  ${PREFLIGHT_WARN_COUNT}"
    echo "  [FAIL]  ${PREFLIGHT_FAIL_COUNT}"
    echo "  [FIXED] ${PREFLIGHT_FIXED_COUNT}"
    if [[ -n "${PREFLIGHT_FAIL_COUNT_INITIAL}" ]]; then
        echo "  (initial [FAIL] before fix phase: ${PREFLIGHT_FAIL_COUNT_INITIAL})"
    fi
    if [[ -n "${PREFLIGHT_REPORT_FILE}" ]]; then
        {
            echo "--- summary ---"
            echo "OK=${PREFLIGHT_OK_COUNT} WARN=${PREFLIGHT_WARN_COUNT} FAIL=${PREFLIGHT_FAIL_COUNT} FIXED=${PREFLIGHT_FIXED_COUNT} FAIL_INITIAL=${PREFLIGHT_FAIL_COUNT_INITIAL:-0}"
        } >> "${PREFLIGHT_REPORT_FILE}"
    fi
fi

preflight_compute_exit_code
exit_code=$?

if [[ "${PREFLIGHT_OUTPUT_JSON}" != "true" ]]; then
    if [[ ${exit_code} -eq 1 ]]; then
        log_error "Preflight failed (see [FAIL] lines above)."
    elif [[ ${exit_code} -eq 2 ]]; then
        log_warn "Preflight completed with warnings only."
    else
        log_info "Preflight passed."
    fi
fi

exit "${exit_code}"
