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

PREFLIGHT_CHECK_ONLY="true"
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
    echo "  --check-only         Only run checks, do not modify the system (default; no root required for partial checks)"
    echo "  --fix                Check + interactively apply fixes (requires root)"
    echo "                       On the install/K8s node use sudo for accurate checks (admin.conf, sysctl) and fixes"
    echo "  -y, --yes            Auto-approve every fix (skip per-fix y/N prompt)"
    echo "  -n, --no             Auto-decline every fix (preview risk text, change nothing)"
    echo "  --fix-allow=LIST     Comma-separated fix names to auto-approve (others are skipped)."
    echo "                       Names: k3s-uninstall,kubeadm-reset,k8s-apt-source,containerd-install,helm-v3,"
    echo "                       chrony,firewalld,ufw,selinux,system-tuning,bridge-sysctl,kernel-limits,iptables-legacy,etc-hosts,"
    echo "                       nodejs-npm,kweaver-sdk,kweaver-admin"
    echo "  --list-fixes         Run checks then list fixes that would be offered (no changes; non-root OK)"
    echo "  --output=json        Emit JSON summary to stdout (human logs to stderr); requires python3"
    echo "  --role=target|admin|both  Target = kubectl/helm only; admin = kweaver/node/npm; both = all (default)"
    echo "                              kweaver CLIs need Node.js 22+ (kweaver-sdk engines; help/zh/install.md)"
    echo "  --no-recheck         Do not re-run full checks after applying fixes"
    echo "  --report=PATH        Append full log to a file"
    echo "  --skip=LIST          Comma-separated check names to skip (see source: preflight_checks.sh preflight_skip)"
    echo ""
    echo "Environment: PREFLIGHT_ROOT=path/to/deploy  PREFLIGHT_CONFIG_YAML=...  PREFLIGHT_K8S_APT_MINOR=vX.YY"
    echo ""
    echo "Exit codes: 0 = OK, 1 = FAIL present, 2 = only WARN (no FAIL)"
    echo ""
    echo "Examples:"
    echo "  $0                                # check-only (default)"
    echo "  sudo $0 --fix                     # check + interactive fixes"
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
        log_warn "Automatic fixes require root. Re-run as: sudo $0 --fix (or use default check-only without --fix)"
        log_info "Falling back to read-only check (--check-only) …"
        PREFLIGHT_CHECK_ONLY="true"
    fi
fi

if [[ "${PREFLIGHT_CHECK_ONLY}" != "true" && "${PREFLIGHT_LIST_FIXES_ONLY}" != "true" ]]; then
    check_root
fi

preflight_reset_counters

_PF_BAR="================================================================"
_pf_section() {
    if [[ "${PREFLIGHT_OUTPUT_JSON}" == "true" ]]; then
        printf '\n%s\n%s\n%s\n' "${_PF_BAR}" "  $1" "${_PF_BAR}" >&2
    else
        printf '\n%s\n%s\n%s\n' "${_PF_BAR}" "  $1" "${_PF_BAR}"
    fi
}

_pf_section "KWeaver preflight checks"
preflight_run_all_checks
PREFLIGHT_FAIL_COUNT_INITIAL="${PREFLIGHT_FAIL_COUNT}"
export PREFLIGHT_FAIL_COUNT_INITIAL

if [[ "${PREFLIGHT_CHECK_ONLY}" != "true" || "${PREFLIGHT_LIST_FIXES_ONLY}" == "true" ]]; then
    _pf_section "Safe fixes"
    preflight_apply_safe_fixes
    if [[ "${PREFLIGHT_NO_RECHECK}" != "true" && "${PREFLIGHT_CHECK_ONLY}" != "true" && "${PREFLIGHT_LIST_FIXES_ONLY}" != "true" ]]; then
        preflight_recheck_after_fixes
    fi
fi

if [[ "${PREFLIGHT_OUTPUT_JSON}" == "true" ]]; then
    _pf_section "Summary"
    echo "  [OK]    ${PREFLIGHT_OK_COUNT}" >&2
    echo "  [WARN]  ${PREFLIGHT_WARN_COUNT}" >&2
    echo "  [FAIL]  ${PREFLIGHT_FAIL_COUNT}" >&2
    echo "  [FIXED] ${PREFLIGHT_FIXED_COUNT}" >&2
    emit_preflight_json
else
    _pf_section "Summary"
    echo "  [OK]    ${PREFLIGHT_OK_COUNT}"
    echo "  [WARN]  ${PREFLIGHT_WARN_COUNT}"
    echo "  [FAIL]  ${PREFLIGHT_FAIL_COUNT}"
    echo "  [FIXED] ${PREFLIGHT_FIXED_COUNT}"
    if [[ -n "${PREFLIGHT_FAIL_COUNT_INITIAL}" ]]; then
        echo "  (initial [FAIL] before fix phase: ${PREFLIGHT_FAIL_COUNT_INITIAL})"
    fi
    if [[ "${PREFLIGHT_FAIL_COUNT}" -gt 0 && ${#PREFLIGHT_FAIL_SNAPSHOT[@]} -gt 0 ]]; then
        echo ""
        echo "  Outstanding [FAIL] items:"
        _pfi=1
        for _pfl in "${PREFLIGHT_FAIL_SNAPSHOT[@]}"; do
            echo "    ${_pfi}. ${_pfl}"
            _pfi=$((_pfi + 1))
        done
    fi
    if [[ -n "${PREFLIGHT_REPORT_FILE}" ]]; then
        {
            echo "--- summary ---"
            echo "OK=${PREFLIGHT_OK_COUNT} WARN=${PREFLIGHT_WARN_COUNT} FAIL=${PREFLIGHT_FAIL_COUNT} FIXED=${PREFLIGHT_FIXED_COUNT} FAIL_INITIAL=${PREFLIGHT_FAIL_COUNT_INITIAL:-0}"
            if [[ "${PREFLIGHT_FAIL_COUNT}" -gt 0 && ${#PREFLIGHT_FAIL_SNAPSHOT[@]} -gt 0 ]]; then
                echo "--- outstanding fails ---"
                for _pfl in "${PREFLIGHT_FAIL_SNAPSHOT[@]}"; do
                    echo "[FAIL] ${_pfl}"
                done
            fi
        } >> "${PREFLIGHT_REPORT_FILE}"
    fi
fi

exit_code=0
preflight_compute_exit_code || exit_code=$?

if [[ "${PREFLIGHT_OUTPUT_JSON}" != "true" ]]; then
    if [[ ${exit_code} -eq 1 ]]; then
        log_error "Preflight failed (see [FAIL] lines above)."
    elif [[ ${exit_code} -eq 2 ]]; then
        log_warn "Preflight completed with warnings only."
    else
        log_info "Preflight passed."
    fi

    _pf_section "Conclusion"
    _pf_total="${PREFLIGHT_KWEAVER_RELEASE_TOTAL:-0}"
    _pf_bad="${PREFLIGHT_KWEAVER_RELEASE_BAD:-0}"
    if [[ "${_pf_total}" -gt 0 ]]; then
        echo "  KWeaver appears INSTALLED on this cluster (${_pf_total} helm release(s))."
        echo "  You probably do NOT need to run a fresh install — the components below already exist:"
        if [[ -n "${PREFLIGHT_KWEAVER_RELEASE_NAMES:-}" ]]; then
            _pf_first=true
            _pf_line=""
            IFS=',' read -r -a _pf_names <<< "${PREFLIGHT_KWEAVER_RELEASE_NAMES}"
            for _pf_n in "${_pf_names[@]}"; do
                if [[ -z "${_pf_line}" ]]; then
                    _pf_line="    ${_pf_n}"
                elif [[ ${#_pf_line} -gt 72 ]]; then
                    echo "${_pf_line},"
                    _pf_line="    ${_pf_n}"
                else
                    _pf_line="${_pf_line}, ${_pf_n}"
                fi
            done
            [[ -n "${_pf_line}" ]] && echo "${_pf_line}"
        fi
        if [[ "${_pf_bad}" -eq 0 ]]; then
            echo ""
            echo "  Suggested next step (skip install, just configure / verify):"
            echo "    - Configure models / BKN search:    ./onboard.sh"
            echo "    - Check status:                     ./deploy.sh kweaver-core status"
            echo "    - Only if you really want to upgrade: ./deploy.sh kweaver-core install --force-upgrade"
        else
            echo ""
            echo "  However, ${_pf_bad}/${_pf_total} release(s) are NOT in 'deployed' state."
            echo "  Suggested next step:"
            echo "    - Inspect:  helm list -A | grep -iE 'kweaver|isf|dip'"
            echo "    - Repair:   ./deploy.sh kweaver-core install --force-upgrade"
        fi
    else
        if [[ ${exit_code} -eq 0 ]]; then
            echo "  No KWeaver releases detected. Environment is READY for a fresh install:"
        else
            echo "  No KWeaver releases detected. After resolving the items above, install with:"
        fi
        echo "    sudo ./deploy.sh kweaver-core install --minimum    # try first / for evaluation"
        echo "    sudo ./deploy.sh kweaver-core install              # full install (auth + business-domain)"
        echo "  Then run ./onboard.sh to register models and enable BKN semantic search."
    fi
    echo "${_PF_BAR}"
fi

exit "${exit_code}"
