#!/usr/bin/env bash
# KWeaver Core — pre-install environment check and safe fixes
# See help/zh/install.md. Must run on the target Linux host as root (sudo), except -h/--help.

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
    echo "  --check-only         Only run checks, do not modify the system (default; still requires root)"
    echo "  --fix                Check + apply fixes (K8s/sysctl/etc.); also offers optional Node ${PREFLIGHT_KWEAVER_MIN_NODE_MAJOR}+ + kweaver CLIs (each ask y/N unless -y)"
    echo "  -y, --yes            Auto-approve every fix (skip per-fix y/N prompt)"
    echo "  -n, --no             Auto-decline every fix (preview risk text, change nothing)"
    echo "  --fix-allow=LIST     Comma-separated fix names to auto-approve (others are skipped)."
    echo "                       Names: k3s-uninstall,kubeadm-reset,k8s-pkgs-repo,k8s-bins,containerd-install,helm-v3,"
    echo "                       chrony,firewalld,ufw,selinux,system-tuning,bridge-sysctl,kernel-limits,nofile-limits,iptables-legacy,etc-hosts,"
    echo "                       onboard-tooling,nodejs-npm,node-22,kweaver-sdk,kweaver-admin"
    echo "  --list-fixes         Run checks then list fixes that would be offered (no changes; requires root)"
    echo "  --output=json        Emit JSON summary to stdout (human logs to stderr); requires python3"
    echo "  --role=target|admin|both  Target = kubectl/helm only; admin = kweaver/node/npm; both = all (default)"
    echo "                              kweaver CLIs need Node.js ${PREFLIGHT_KWEAVER_MIN_NODE_MAJOR}+ (per @kweaver-ai/kweaver-sdk on npm; help/zh/install.md)"
    echo "  --no-recheck         Do not re-run full checks after applying fixes"
    echo "  --lenient            Downgrade install-blocking [FAIL] items (sysctl, ip_forward, kernel modules,"
    echo "                       containerd, kubectl, helm, swap, broken apt sources, missing k8s/containerd"
    echo "                       install candidate, ulimit, inotify, vm.max_map_count, overlay) back to [WARN]."
    echo "                       Same as PREFLIGHT_STRICT=false PREFLIGHT_STRICT_SOURCES=false."
    echo "  --forget-decisions   Wipe remembered \"no\" answers under /var/lib/kweaver/preflight-decline-* before"
    echo "                       running. Use after you change your mind about onboard-tooling / node-22 etc."
    echo "                       Same as PREFLIGHT_FORGET_DECISIONS=true."
    echo "  --report=PATH        Append full log to a file"
    echo "  --skip=LIST          Comma-separated check names to skip (see source: preflight_checks.sh preflight_skip)"
    echo ""
    echo "Environment:"
    echo "  PREFLIGHT_ROOT=path/to/deploy            override deploy root (defaults to script dir)"
    echo "  PREFLIGHT_CONFIG_YAML=path/to/config     override config.yaml"
    echo "  PREFLIGHT_K8S_APT_MINOR=vX.YY            pin pkgs.k8s.io minor (default v1.28 / detected from kubeadm)"
    echo "  PREFLIGHT_STRICT=true|false              [default true] install-blocking items as [FAIL] not [WARN]"
    echo "  PREFLIGHT_STRICT_SOURCES=true|false      [default true] verify apt/yum can fetch kubeadm + containerd"
    echo "  PREFLIGHT_REMEMBER_DECISIONS=true|false  [default true] persist 'no' to onboard-tooling / node-22"
    echo "  PREFLIGHT_FORGET_DECISIONS=true          wipe remembered decisions before this run (one-shot)"
    echo "  PREFLIGHT_DECISION_DIR=/path             where decision sentinels live (default /var/lib/kweaver)"
    echo ""
    echo "Exit codes: 0 = OK, 1 = FAIL present, 2 = only WARN (no FAIL)"
    echo ""
    echo "Examples:"
    echo "  sudo $0                            # check-only (default)"
    echo "  sudo $0 --fix                      # check + interactive fixes"
    echo "  sudo $0 --list-fixes"
    echo "  sudo $0 --skip=network --report=/tmp/preflight.txt"
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
        --lenient)
            PREFLIGHT_STRICT="false"
            PREFLIGHT_STRICT_SOURCES="false"
            shift
            ;;
        --forget-decisions)
            PREFLIGHT_FORGET_DECISIONS="true"
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
export PREFLIGHT_STRICT PREFLIGHT_STRICT_SOURCES
export PREFLIGHT_REMEMBER_DECISIONS PREFLIGHT_FORGET_DECISIONS PREFLIGHT_DECISION_DIR

# Wipe remembered "no" answers (onboard-tooling / node-22) before the run.
if [[ "${PREFLIGHT_FORGET_DECISIONS:-false}" == "true" ]]; then
    preflight_forget_decisions
fi

if [[ -n "${PREFLIGHT_REPORT_FILE}" ]]; then
    mkdir -p "$(dirname "${PREFLIGHT_REPORT_FILE}")" 2>/dev/null || true
    {
        echo "=== KWeaver preflight $(date -Iseconds) ==="
    } > "${PREFLIGHT_REPORT_FILE}"
fi

if [[ "${EUID}" -ne 0 ]]; then
    log_error "Preflight must be run as root: sudo $0 [options]  (only -h / --help works without root)"
    exit 1
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
        if [[ "${PREFLIGHT_STRICT:-true}" == "true" && "${PREFLIGHT_CHECK_ONLY}" == "true" ]]; then
            log_info "Hint: most install-blocking [FAIL] items are auto-fixable — re-run: sudo $0 --fix"
            log_info "      Need to bypass strict severity (e.g. low-spec lab box)? sudo $0 --check-only --lenient"
        fi
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
            echo "    - Node/kweaver on an admin host: default preflight is check-only; run sudo preflight --fix to opt in to help installing Node ${PREFLIGHT_KWEAVER_MIN_NODE_MAJOR}+ and CLIs (y/N per step)"
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
            echo "  No KWeaver releases detected. Environment looks ready for a first-time install:"
            echo "    sudo ./deploy.sh kweaver-core install --minimum    # try first / for evaluation"
            echo "    sudo ./deploy.sh kweaver-core install              # full install (auth + business-domain)"
            echo ""
            echo "  After deploy: from this repo's deploy/ directory run ./onboard.sh (needs Node ${PREFLIGHT_KWEAVER_MIN_NODE_MAJOR}+ + kweaver CLI on that host)."
            echo "  If this host still lacks Node/CLIs: sudo $0 --fix"
        else
            echo "  No KWeaver releases detected, but preflight above is NOT all clear — fix that before treating deploy as ready."
            echo "  Typical loop:"
            echo "    sudo $0 --fix          # applies safe fixes / opt-in tooling (y/N unless -y)"
            echo "    sudo $0 --check-only   # re-check until blocking [FAIL] items are addressed (or sudo $0 --check-only --lenient if you accept the caveats)"
            echo "  Only then install:"
            echo "    sudo ./deploy.sh kweaver-core install --minimum    # try first / for evaluation"
            echo "    sudo ./deploy.sh kweaver-core install              # full install (auth + business-domain)"
            echo "  Finally: ./onboard.sh from deploy/ (Node ${PREFLIGHT_KWEAVER_MIN_NODE_MAJOR}+ + kweaver on PATH; sudo $0 --fix helps install tooling on this machine)."
        fi
    fi
    echo "${_PF_BAR}"
fi

exit "${exit_code}"
