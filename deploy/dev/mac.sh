#!/usr/bin/env bash
# KWeaver — macOS dev helper (kind + Helm). Does NOT run Linux preflight/k3s/kubeadm.
#
# Typical order (from repo deploy/: cd deploy):
#   1. doctor                 — optional; check docker / kind / kubectl / helm / node
#   2. doctor --fix           — optional; install missing CLIs via Homebrew (prompts; -y skips)
#   3. cluster up             — kind + ingress-nginx; context becomes kind-<KIND_CLUSTER_NAME>
#   4. data-services install  — MariaDB / Redis / Kafka / Zookeeper / OpenSearch (required before Core on mac)
#   5. kweaver-core download  — optional; cache charts locally (mac.sh defaults --minimum unless --full)
#   6. kweaver-core install   — Helm install Core (mac.sh adds --minimum unless you pass --full)
#   7. isf / etrino (vega)    — optional; same Helm path as Linux when cluster + config are ready
#   8. onboard                — optional; needs kweaver CLI + Core up (add -y for non-interactive)
#   Teardown: cluster down
#   Full write-up: deploy/dev/README.md
#
# Usage:
#   bash deploy/dev/mac.sh doctor
#   bash deploy/dev/mac.sh doctor --fix
#   bash deploy/dev/mac.sh cluster up
#   bash deploy/dev/mac.sh cluster down
#   bash deploy/dev/mac.sh cluster status
#   bash deploy/dev/mac.sh data-services install
#   bash deploy/dev/mac.sh kweaver-core install
#   bash deploy/dev/mac.sh kweaver-core download
#   bash deploy/dev/mac.sh isf install
#   bash deploy/dev/mac.sh etrino install   # Vega stack (alias: vega)
#   bash deploy/dev/mac.sh onboard
#
# Global flags (same as deploy.sh; must come first):
#   -y, --yes | --force-upgrade | --distro=k3s|k8s|kubeadm
#
# Commands that delegate to deploy.sh run Helm chart logic only on mac
# (no host k3s / bundled data-service bootstrap) unless you run **data-services install**
# (or individual mariadb/redis/… via deploy.sh). See deploy/dev/README.md.
#
# doctor --fix prompts before running brew unless you pass -y / --yes (globally before doctor, or after --fix).
#
set -euo pipefail

SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SELF_PATH="${SELF_DIR}/$(basename "${BASH_SOURCE[0]}")"
# Render examples with the actual invocation: prefer how the user called the script
# (so copy-paste matches their muscle memory), but fall back to absolute path so
# relative invocations from another cwd still produce runnable lines.
if [[ "${0}" == /* ]]; then
    INVOKE_CMD="bash ${0}"
elif [[ "${0}" == */* ]]; then
    INVOKE_CMD="bash ${SELF_PATH}"
else
    INVOKE_CMD="bash ${SELF_PATH}"
fi
# shellcheck source=lib/mac_common.sh
source "${SELF_DIR}/lib/mac_common.sh"

ASSUME_YES="${ASSUME_YES:-false}"

usage() {
    local cmd="${INVOKE_CMD}"
    local readme="${SELF_DIR}/README.md"
    local mac_cfg="${SELF_DIR}/conf/mac-config.yaml"
    cat <<EOF
KWeaver mac dev (kind) — thin wrapper around deploy/onboard.

Typical order (shortest path: doctor? → cluster up → data-services install → kweaver-core install):
  1) doctor                     optional toolchain check
  2) cluster up                 kind + ingress; kubectl context kind-<name>
  3) data-services install      MariaDB, Redis, Kafka, Zookeeper, OpenSearch (mac default skips second ingress)
  4) kweaver-core download      optional; charts cache only (minimum profile by default)
  5) kweaver-core install ...   Helm install (--minimum implied by mac.sh; use --full to opt out)
  6) isf / etrino|vega          optional; deploy.sh modules (cluster + config must be ready)
  7) onboard                    optional; after Core is up
  cluster down                  delete kind cluster
  See ${readme}

Commands:
  doctor [--fix] [-y|--yes]        Check toolchain; --fix runs brew after confirm (use -y to skip prompt)
  cluster up|down|status           kind cluster + ingress-nginx (kind manifest)
  data-services install|uninstall  Platform data layer (install before Core on mac); uninstall tears down bundled charts
  kweaver-core|core <action> ...   Delegates to deploy.sh (see deploy.sh help)
  isf <action> ...                 ISF via deploy.sh (install|download|uninstall|status)
  etrino|vega <action> ...         Vega charts (vega-hdfs/calculate/metadata) via deploy.sh; vega = alias of etrino
  onboard [args ...]               Runs deploy/onboard.sh

Examples:
  ${cmd} doctor
  ${cmd} doctor --fix              # confirm before brew
  ${cmd} -y doctor --fix           # no prompt (same as deploy.sh global -y)
  ${cmd} doctor --fix -y
  ${cmd} cluster up
  ${cmd} data-services install
  ${cmd} kweaver-core install --full   # full manifest / ISF when manifest says so
  ${cmd} kweaver-core install
  ${cmd} kweaver-core download
  ${cmd} isf install
  ${cmd} vega status
  ${cmd} onboard

Not wired on mac.sh: kweaver-dip|dip (use Linux deploy.sh).

Environment:
  KIND_CLUSTER_NAME       Default: kweaver-dev
  CONFIG_YAML_PATH        Default: ${mac_cfg} when unset (kweaver-core|core|isf|etrino|vega|data-services)

Note: data-services install runs deploy.sh data-services (Helm charts into the current kube context). Other deploy.sh modules on mac still skip host k3s bootstrap unless you install infra yourself. See ${readme}.

Mac default: kweaver-core / core add --minimum unless you pass --minimum / --min already or opt out with --full.

EOF
}

main() {
    local -a global_flags=()
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -y | --yes)
                ASSUME_YES="true"
                global_flags+=(-y)
                shift
                ;;
            --force-upgrade)
                global_flags+=(--force-upgrade)
                shift
                ;;
            --distro=k3s | --distro=k8s | --distro=kubeadm)
                global_flags+=("$1")
                shift
                ;;
            --distro)
                global_flags+=("$1" "$2")
                shift 2
                ;;
            -h | --help)
                usage
                exit 0
                ;;
            -*)
                mac_log_error "Unknown flag: $1"
                usage >&2
                exit 1
                ;;
            *) break ;;
        esac
    done

    if [[ $# -lt 1 ]]; then
        usage >&2
        exit 1
    fi

    local cmd="$1"
    shift

    export ASSUME_YES
    mac_common_init

    export MAC_DOCTOR_FIX_CMD="bash ${SELF_DIR}/mac.sh doctor --fix"
    export MAC_DOCTOR_FIX_CMD_AUTO="bash ${SELF_DIR}/mac.sh -y doctor --fix"

    case "${cmd}" in
        doctor)
            export MAC_DOCTOR_HINT_NEXT_STEPS=true
            mac_require_darwin
            local want_fix=false
            while [[ $# -gt 0 ]]; do
                case "${1:-}" in
                    --fix)
                        want_fix=true
                        shift
                        ;;
                    -y | --yes)
                        ASSUME_YES="true"
                        shift
                        ;;
                    *)
                        mac_log_error "Unknown doctor argument: $1"
                        exit 1
                        ;;
                esac
            done
            if [[ "${want_fix}" == "true" ]]; then
                if mac_doctor; then
                    exit 0
                fi
                if [[ "${MAC_DOCTOR_DOCKER_DAEMON_DOWN:-0}" == "1" && "${MAC_DOCTOR_BREW_FIX_USEFUL:-0}" != "1" ]]; then
                    mac_log_error "Docker engine is not running. Open Docker Desktop, wait until it is ready, then run doctor again."
                    mac_log_error "doctor --fix only installs CLIs via Homebrew; it does not start Docker."
                    exit 1
                fi
                mac_log_info "---"
                if ! mac_doctor_confirm_fix; then
                    exit 1
                fi
                if ! mac_doctor_apply_fixes; then
                    exit 1
                fi
                mac_log_info "Re-running doctor after --fix..."
                mac_doctor
            else
                mac_doctor
            fi
            ;;
        cluster)
            mac_require_darwin
            if ! mac_doctor; then
                exit 1
            fi
            # shellcheck source=lib/mac_cluster.sh
            source "${SELF_DIR}/lib/mac_cluster.sh"
            mac_cluster_dispatch "$@"
            ;;
        data-services)
            mac_require_darwin
            if ! mac_doctor; then
                exit 1
            fi
            if ! mac_kube_context_guard; then
                exit 1
            fi
            if [[ -z "${CONFIG_YAML_PATH:-}" ]]; then
                export CONFIG_YAML_PATH="${MAC_DEV_ROOT}/conf/mac-config.yaml"
            fi
            # kind path already installs ingress-nginx; avoid a second controller from ensure_data_services.
            export AUTO_INSTALL_INGRESS_NGINX="${AUTO_INSTALL_INGRESS_NGINX:-false}"
            export AUTO_INSTALL_LOCALPV="${AUTO_INSTALL_LOCALPV:-true}"
            if [[ ${#global_flags[@]} -gt 0 ]]; then
                exec bash "${DEPLOY_ROOT}/deploy.sh" "${global_flags[@]}" data-services "$@"
            fi
            exec bash "${DEPLOY_ROOT}/deploy.sh" data-services "$@"
            ;;
        kweaver-core | core)
            mac_require_darwin
            if ! mac_doctor; then
                exit 1
            fi
            if ! mac_kube_context_guard; then
                exit 1
            fi
            if [[ -z "${CONFIG_YAML_PATH:-}" ]]; then
                export CONFIG_YAML_PATH="${MAC_DEV_ROOT}/conf/mac-config.yaml"
            fi
            export KWEAVER_SKIP_PLATFORM_BOOTSTRAP="${KWEAVER_SKIP_PLATFORM_BOOTSTRAP:-true}"
            # Default --minimum for Mac dev (skips ISF chart download/install when manifest ties ISF to auth).
            # IMPORTANT: deploy.sh parses argv as  module  action  [flags...]  — never put --minimum before
            # the action (e.g. download|install), or action becomes --minimum and flags are never parsed.
            local -a _kw_pos=()
            local _a _kw_saw_min=false _kw_saw_full=false
            for _a in "$@"; do
                case "${_a}" in
                    --minimum | --min)
                        _kw_saw_min=true
                        _kw_pos+=("${_a}")
                        ;;
                    --full)
                        _kw_saw_full=true
                        ;;
                    *)
                        _kw_pos+=("${_a}")
                        ;;
                esac
            done
            local -a _kw_final=()
            if [[ "${_kw_saw_full}" != "true" ]] && [[ "${_kw_saw_min}" != "true" ]]; then
                if [[ ${#_kw_pos[@]} -eq 0 ]]; then
                    mac_log_error "kweaver-core|core needs an action (e.g. download, install, status)."
                    exit 1
                fi
                # Insert --minimum after deploy action (index 0), bash 3.2–safe (no ${arr[@]:1}).
                _kw_final=("${_kw_pos[0]}" --minimum)
                local _kw_i
                for ((_kw_i = 1; _kw_i < ${#_kw_pos[@]}; _kw_i++)); do
                    _kw_final+=("${_kw_pos[_kw_i]}")
                done
            else
                _kw_final=("${_kw_pos[@]}")
            fi
            if [[ ${#global_flags[@]} -gt 0 ]]; then
                exec bash "${DEPLOY_ROOT}/deploy.sh" "${global_flags[@]}" "${cmd}" "${_kw_final[@]}"
            fi
            exec bash "${DEPLOY_ROOT}/deploy.sh" "${cmd}" "${_kw_final[@]}"
            ;;
        isf)
            mac_require_darwin
            if ! mac_doctor; then
                exit 1
            fi
            if ! mac_kube_context_guard; then
                exit 1
            fi
            if [[ -z "${CONFIG_YAML_PATH:-}" ]]; then
                export CONFIG_YAML_PATH="${MAC_DEV_ROOT}/conf/mac-config.yaml"
            fi
            export KWEAVER_SKIP_PLATFORM_BOOTSTRAP="${KWEAVER_SKIP_PLATFORM_BOOTSTRAP:-true}"
            if [[ ${#global_flags[@]} -gt 0 ]]; then
                exec bash "${DEPLOY_ROOT}/deploy.sh" "${global_flags[@]}" isf "$@"
            fi
            exec bash "${DEPLOY_ROOT}/deploy.sh" isf "$@"
            ;;
        etrino | vega)
            mac_require_darwin
            if ! mac_doctor; then
                exit 1
            fi
            if ! mac_kube_context_guard; then
                exit 1
            fi
            if [[ -z "${CONFIG_YAML_PATH:-}" ]]; then
                export CONFIG_YAML_PATH="${MAC_DEV_ROOT}/conf/mac-config.yaml"
            fi
            export KWEAVER_SKIP_PLATFORM_BOOTSTRAP="${KWEAVER_SKIP_PLATFORM_BOOTSTRAP:-true}"
            if [[ ${#global_flags[@]} -gt 0 ]]; then
                exec bash "${DEPLOY_ROOT}/deploy.sh" "${global_flags[@]}" etrino "$@"
            fi
            exec bash "${DEPLOY_ROOT}/deploy.sh" etrino "$@"
            ;;
        onboard)
            mac_require_darwin
            if ! mac_doctor; then
                exit 1
            fi
            if ! mac_kube_context_guard; then
                exit 1
            fi
            export NAMESPACE="${NAMESPACE:-kweaver}"
            exec bash "${DEPLOY_ROOT}/onboard.sh" "$@"
            ;;
        kweaver-dip | dip)
            mac_log_error "Module '${cmd}' is not wired in mac.sh. Use deploy.sh on Linux."
            exit 1
            ;;
        *)
            mac_log_error "Unknown command: ${cmd}"
            usage >&2
            exit 1
            ;;
    esac
}

main "$@"
