#!/usr/bin/env bash
# KWeaver — macOS dev helper (kind + Helm). Does NOT run Linux preflight/k3s/kubeadm.
#
# Typical order (from repo deploy/: cd deploy):
#   1. doctor                 — optional; check docker / kind / kubectl / helm / node
#   2. doctor --fix           — optional; install missing CLIs via Homebrew (prompts; -y skips)
#   3. cluster up             — kind + ingress-nginx; context becomes kind-<KIND_CLUSTER_NAME>
#   4. kweaver-core download  — optional; cache charts locally (no cluster install)
#   5. kweaver-core install --minimum — Helm install Core (prefix -y to skip deploy.sh prompts)
#   6. isf / etrino (vega)  — optional; same Helm path as Linux when cluster + config are ready
#   7. onboard                — optional; needs kweaver CLI + Core up (add -y for non-interactive)
#   Teardown: cluster down
#   Full write-up: deploy/dev/README.md
#
# Usage:
#   bash deploy/dev/mac.sh doctor
#   bash deploy/dev/mac.sh doctor --fix
#   bash deploy/dev/mac.sh cluster up
#   bash deploy/dev/mac.sh cluster down
#   bash deploy/dev/mac.sh cluster status
#   bash deploy/dev/mac.sh kweaver-core install --minimum
#   bash deploy/dev/mac.sh isf install
#   bash deploy/dev/mac.sh etrino install   # Vega stack (alias: vega)
#   bash deploy/dev/mac.sh onboard
#
# Global flags (same as deploy.sh; must come first):
#   -y, --yes | --force-upgrade | --distro=k3s|k8s|kubeadm
#
# For kweaver-core|core|isf|etrino|vega: if CONFIG_YAML_PATH is unset, default is
#   deploy/dev/conf/mac-config.yaml. Those paths call deploy.sh in Helm-only mode on mac
#   (no host k3s / data-service bootstrap). vega is an alias for etrino in mac.sh.
#
# doctor --fix prompts before running brew unless you pass -y / --yes (globally before doctor, or after --fix).
#
set -euo pipefail

SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/mac_common.sh
source "${SELF_DIR}/lib/mac_common.sh"

ASSUME_YES="${ASSUME_YES:-false}"

usage() {
    cat <<'EOF'
KWeaver mac dev (kind) — thin wrapper around deploy/onboard.

Typical order (shortest path: doctor? → cluster up → kweaver-core install --minimum):
  1) doctor                     optional toolchain check
  2) cluster up                 kind + ingress; kubectl context kind-<name>
  3) kweaver-core download      optional; charts cache only (no install)
  4) kweaver-core install ...   Helm install (often --minimum)
  5) isf / etrino|vega          optional; deploy.sh modules (cluster + config must be ready)
  6) onboard                    optional; after Core is up
  cluster down                  delete kind cluster
  See deploy/dev/README.md

Commands:
  doctor [--fix] [-y|--yes]        Check toolchain; --fix runs brew after confirm (use -y to skip prompt)
  cluster up|down|status          kind cluster + ingress-nginx (kind manifest)
  kweaver-core|core <action> ...   Delegates to deploy.sh (see deploy.sh help)
  isf <action> ...                 ISF via deploy.sh (install|download|uninstall|status)
  etrino|vega <action> ...         Vega charts (vega-hdfs/calculate/metadata) via deploy.sh; vega = alias of etrino
  onboard [args ...]               Runs deploy/onboard.sh

Examples:
  bash deploy/dev/mac.sh doctor
  bash deploy/dev/mac.sh doctor --fix              # confirm before brew
  bash deploy/dev/mac.sh -y doctor --fix           # no prompt (same as deploy.sh global -y)
  bash deploy/dev/mac.sh doctor --fix -y
  bash deploy/dev/mac.sh cluster up
  bash deploy/dev/mac.sh kweaver-core install --minimum
  bash deploy/dev/mac.sh kweaver-core download
  bash deploy/dev/mac.sh isf install
  bash deploy/dev/mac.sh vega status
  bash deploy/dev/mac.sh onboard

Not wired on mac.sh: kweaver-dip|dip (use Linux deploy.sh).

Environment:
  KIND_CLUSTER_NAME       Default: kweaver-dev
  CONFIG_YAML_PATH        Default: deploy/dev/conf/mac-config.yaml when unset (kweaver-core|core|isf|etrino|vega)

Note: Commands that delegate to deploy.sh run Helm chart logic only on mac (no host k3s / bundled data-service bootstrap). See deploy/dev/README.md.

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
        kweaver-core | core | isf)
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
            exec bash "${DEPLOY_ROOT}/deploy.sh" "${global_flags[@]}" "${cmd}" "$@"
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
            exec bash "${DEPLOY_ROOT}/deploy.sh" "${global_flags[@]}" etrino "$@"
            ;;
        onboard)
            mac_require_darwin
            if ! mac_doctor; then
                exit 1
            fi
            if ! mac_kube_context_guard; then
                exit 1
            fi
            export NAMESPACE="${NAMESPACE:-kweaver-ai}"
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
