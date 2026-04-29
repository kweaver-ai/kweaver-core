#!/usr/bin/env bash
# Shared helpers for deploy/dev/mac.sh (sourced; not executed directly).
# Requires: mac_common_init was run (sets DEPLOY_ROOT, SCRIPT_DIR, etc.)

mac_log_info() {
    echo "[mac] $*"
}

mac_log_warn() {
    echo "[mac] WARN: $*" >&2
}

mac_log_error() {
    echo "[mac] ERROR: $*" >&2
}

# Caller sets DEPLOY_ROOT, CONF_DIR, CONFIG_YAML_PATH, and sources common.sh after this.
mac_common_init() {
    local here
    here="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
    export MAC_DEV_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
    export DEPLOY_ROOT="${here}"
    export SCRIPT_DIR="${DEPLOY_ROOT}"
    export CONF_DIR="${DEPLOY_ROOT}/conf"
    export FLANNEL_MANIFEST_PATH="${DEPLOY_ROOT}/conf/kube-flannel.yml"
    export LOCALPV_MANIFEST_PATH="${DEPLOY_ROOT}/conf/local-path-storage.yaml"
    export HELM_INSTALL_SCRIPT_PATH="${DEPLOY_ROOT}/conf/get-helm-3"

    export KIND_CLUSTER_NAME="${KIND_CLUSTER_NAME:-kweaver-dev}"
    export KWEAVER_SKIP_PLATFORM_BOOTSTRAP="${KWEAVER_SKIP_PLATFORM_BOOTSTRAP:-true}"
}

mac_require_darwin() {
    local os
    os="$(uname -s 2>/dev/null || true)"
    if [[ "${os}" != "Darwin" ]]; then
        mac_log_warn "Expected macOS (Darwin); uname='${os}'. Continuing anyway."
    fi
}

mac_check_cmd() {
    local name
    name="${1:-}"
    [[ -n "${name}" ]] || return 1
    command -v "${name}" >/dev/null 2>&1
}

# Prints lines like "ok:kubectl" or "missing:helm"
mac_doctor() {
    local fail=0
    local c

    mac_log_info "Checking local toolchain (kind dev cluster on Mac)..."

    for c in docker kind kubectl helm; do
        if mac_check_cmd "${c}"; then
            echo "ok:${c}"
        else
            echo "missing:${c}"
            fail=1
        fi
    done

    if command -v node >/dev/null 2>&1; then
        local major
        major="$(node -p "process.versions.node.split('.')[0]" 2>/dev/null || echo 0)"
        if [[ "${major}" -ge 22 ]]; then
            echo "ok:node (>=22)"
        else
            echo "warn:node (need major >= 22 for kweaver CLI / onboard; found $(node -v 2>/dev/null || true))"
            fail=1
        fi
    else
        echo "missing:node (install Node 22+ for onboard / kweaver CLI)"
        fail=1
    fi

    if [[ "${fail}" -ne 0 ]]; then
        mac_log_warn "Install hints: brew install docker kind kubectl helm node@22"
        return 1
    fi
    return 0
}

mac_kube_context_guard() {
    local expected="kind-${KIND_CLUSTER_NAME}"
    local cur
    cur="$(kubectl config current-context 2>/dev/null || true)"
    if [[ "${cur}" != "${expected}" ]]; then
        mac_log_warn "kubectl context is '${cur:-(none)}', expected '${expected}'."
        mac_log_warn "Run: kubectl config use-context ${expected}"
        return 1
    fi
    return 0
}
