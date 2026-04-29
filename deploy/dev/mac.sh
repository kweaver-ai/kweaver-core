#!/usr/bin/env bash
# KWeaver — macOS dev helper (kind + Helm). Does NOT run Linux preflight/k3s/kubeadm.
#
# Usage:
#   bash deploy/dev/mac.sh doctor
#   bash deploy/dev/mac.sh cluster up
#   bash deploy/dev/mac.sh cluster down
#   bash deploy/dev/mac.sh cluster status
#   bash deploy/dev/mac.sh -y kweaver-core install --minimum
#   bash deploy/dev/mac.sh onboard -y
#
# Global flags (same as deploy.sh; must come first):
#   -y, --yes | --force-upgrade | --distro=k3s|k8s|kubeadm
#
# Uses CONFIG_YAML_PATH=deploy/dev/conf/mac-config.yaml and KWEAVER_SKIP_PLATFORM_BOOTSTRAP=true
# so deploy.sh only runs Helm chart logic (no host bootstrap).
#
set -euo pipefail

SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/mac_common.sh
source "${SELF_DIR}/lib/mac_common.sh"

usage() {
    cat <<'EOF'
KWeaver mac dev (kind) — thin wrapper around deploy/onboard.

Commands:
  doctor                           Check docker, kind, kubectl, helm, node>=22
  cluster up|down|status           kind cluster + ingress-nginx (kind manifest)
  kweaver-core|core <action> ...   Delegates to deploy.sh (see deploy.sh help)
  onboard [args ...]               Runs deploy/onboard.sh

Examples:
  bash deploy/dev/mac.sh doctor
  bash deploy/dev/mac.sh cluster up
  bash deploy/dev/mac.sh -y kweaver-core install --minimum
  bash deploy/dev/mac.sh kweaver-core download
  bash deploy/dev/mac.sh onboard -y

Not in v1 (use Linux deploy.sh): isf, kweaver-dip.

Environment:
  KIND_CLUSTER_NAME       Default: kweaver-dev
  CONFIG_YAML_PATH        Overridden for kweaver-core unless already set (see below)
  KWEAVER_SKIP_PLATFORM_BOOTSTRAP  Default true for kweaver-core via this script

EOF
}

main() {
    local -a global_flags=()
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -y | --yes)
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

    mac_common_init

    case "${cmd}" in
        doctor)
            mac_require_darwin
            mac_doctor
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
            exec bash "${DEPLOY_ROOT}/deploy.sh" "${global_flags[@]}" "${cmd}" "$@"
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
        isf | kweaver-dip | dip)
            mac_log_error "Module '${cmd}' is not wired in mac.sh v1. Use deploy.sh on Linux or extend deploy/dev/mac.sh."
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
