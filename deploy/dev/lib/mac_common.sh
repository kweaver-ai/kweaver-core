#!/usr/bin/env bash
# Shared helpers for deploy/dev/mac.sh (sourced; not executed directly).
# Requires: mac_common_init was run (sets DEPLOY_ROOT, SCRIPT_DIR, etc.)

mac_log_info() {
    echo "[mac] $*"
}

mac_log_warn() {
    echo "[WARNING] $*" >&2
}

mac_log_error() {
    echo "[FAIL] $*" >&2
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
        mac_doctor_theme
        printf '%b[WARNING]%b Expected macOS (Darwin); uname=%s (continuing anyway).\n' "${MAC_D_WARN}" "${MAC_D_RESET}" "${os}" >&2
    fi
}

mac_check_cmd() {
    local name
    name="${1:-}"
    [[ -n "${name}" ]] || return 1
    command -v "${name}" >/dev/null 2>&1
}

# ANSI highlights in a terminal (stdout or stderr).
mac_doctor_theme() {
    if [[ -t 1 || -t 2 ]]; then
        MAC_D_OK=$'\033[32m'
        MAC_D_WARN=$'\033[33m'
        MAC_D_BAD=$'\033[31m'
        MAC_D_DIM=$'\033[2m'
        MAC_D_BOLD=$'\033[1m'
        MAC_D_RESET=$'\033[0m'
    else
        MAC_D_OK=""
        MAC_D_WARN=""
        MAC_D_BAD=""
        MAC_D_DIM=""
        MAC_D_BOLD=""
        MAC_D_RESET=""
    fi
}

mac_doctor_hint_for_tool() {
    local t="${1:-}"
    case "${t}" in
        docker) printf '%s\n' "brew install --cask docker" ;;
        kind) printf '%s\n' "brew install kind" ;;
        kubectl) printf '%s\n' "brew install kubectl" ;;
        helm) printf '%s\n' "brew install helm" ;;
        node) printf '%s\n' "brew install node@22" ;;
        *) printf '%s\n' "(no packaged hint)" ;;
    esac
}

# Minimum Helm 3.x for Bitnami-embedded charts (Kafka data-services path).
MAC_HELM_MIN_SEMVER="${MAC_HELM_MIN_SEMVER:-3.10.0}"

mac_helm_client_semver() {
    local raw
    raw="$(helm version 2>/dev/null || true)"
    printf '%s' "${raw}" | grep -oE 'Version:"v[0-9]+\.[0-9]+\.[0-9]+"' | head -1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' || true
}

mac_helm_version_ok_for_charts() {
    local cur="$1"
    [[ -n "${cur}" ]] || return 1
    [[ "$(printf '%s\n' "${MAC_HELM_MIN_SEMVER}" "${cur}" | sort -V | head -1)" == "${MAC_HELM_MIN_SEMVER}" ]]
}

# True if node is usable (present and major >= 22).
mac_doctor_node_ok() {
    if ! command -v node >/dev/null 2>&1; then
        return 1
    fi
    local major
    major="$(node -p "process.versions.node.split('.')[0]" 2>/dev/null || echo 0)"
    [[ "${major}" -ge 22 ]]
}

# Confirm before doctor --fix runs brew (skip if ASSUME_YES=true or non-interactive -y).
mac_doctor_confirm_fix() {
    if [[ "${ASSUME_YES:-false}" == "true" ]]; then
        mac_log_info "Assume yes (-y): skipping confirmation before brew install."
        return 0
    fi
    if [[ ! -t 0 ]]; then
        mac_log_error "doctor --fix needs a TTY to confirm, or pass -y / --yes (e.g. bash .../mac.sh -y doctor --fix)."
        return 1
    fi
    local ans
    read -r -p "Install missing tools with Homebrew now? [y/N] " ans || true
    case "${ans}" in
        y | Y | yes | YES)
            return 0
            ;;
        *)
            mac_log_warn "Aborted; no Homebrew installs were run."
            return 1
            ;;
    esac
}

# Install missing CLI tools via Homebrew (mac doctor --fix). Idempotent per package.
mac_doctor_apply_fixes() {
    if ! command -v brew >/dev/null 2>&1; then
        mac_log_error "Homebrew (brew) not found. Install from https://brew.sh then retry doctor --fix."
        return 1
    fi

    mac_log_info "doctor --fix: installing missing packages via brew..."
    local c
    local ec=0

    for c in docker kind kubectl helm; do
        if ! mac_check_cmd "${c}"; then
            mac_log_info "brew: installing ${c}..."
            case "${c}" in
                docker)
                    if ! brew install --cask docker; then ec=1; fi
                    ;;
                kind)
                    if ! brew install kind; then ec=1; fi
                    ;;
                kubectl)
                    if ! brew install kubectl; then ec=1; fi
                    ;;
                helm)
                    if ! brew install helm; then ec=1; fi
                    ;;
            esac
        fi
    done

    if ! mac_doctor_node_ok; then
        mac_log_info "brew: installing node@22..."
        if ! brew install node@22; then ec=1; fi
        # Caveats: keg-only; common paths for Apple Silicon vs Intel Homebrew
        if [[ -d /opt/homebrew/opt/node@22/bin ]]; then
            mac_log_warn "If \`node -v\` is still < 22, add to PATH: export PATH=\"/opt/homebrew/opt/node@22/bin:\$PATH\""
        elif [[ -d /usr/local/opt/node@22/bin ]]; then
            mac_log_warn "If \`node -v\` is still < 22, add to PATH: export PATH=\"/usr/local/opt/node@22/bin:\$PATH\""
        fi
    fi

    if [[ "${ec}" -ne 0 ]]; then
        mac_log_warn "Some brew commands failed; re-run doctor to see what is still missing."
    fi
    return 0
}

# Pretty status lines; missing/warn rows print only the fix that applies.
# On failure, prints MAC_DOCTOR_FIX_CMD hint when set (e.g. by mac.sh).
mac_doctor() {
    local fail=0
    local brew_fix_useful=0
    local docker_daemon_down=0
    local c
    local hint

    export MAC_DOCTOR_DOCKER_DAEMON_DOWN=0
    export MAC_DOCTOR_BREW_FIX_USEFUL=0

    mac_doctor_theme
    mac_log_info "Checking local toolchain (kind dev cluster on Mac)..."

    for c in docker kind kubectl helm; do
        if [[ "${c}" == "docker" ]]; then
            if ! mac_check_cmd docker; then
                brew_fix_useful=1
                hint="$(mac_doctor_hint_for_tool docker)"
                printf '%b[FAIL]%b docker (CLI missing)\n' "${MAC_D_BAD}" "${MAC_D_RESET}"
                printf '  %bto fix:%b %b%s%b\n' "${MAC_D_DIM}" "${MAC_D_RESET}" "${MAC_D_BOLD}" "${hint}" "${MAC_D_RESET}"
                fail=1
            elif ! docker info >/dev/null 2>&1; then
                docker_daemon_down=1
                printf '%b[FAIL]%b docker (CLI ok, engine not reachable — kind needs a running daemon)\n' "${MAC_D_BAD}" "${MAC_D_RESET}"
                printf '  %bto fix:%b Open %bDocker Desktop%b, wait until it is running (whale icon), then run: %bdocker info%b\n' \
                    "${MAC_D_DIM}" "${MAC_D_RESET}" "${MAC_D_BOLD}" "${MAC_D_RESET}" "${MAC_D_BOLD}" "${MAC_D_RESET}"
                printf '  %bNote:%b %bdoctor --fix%b cannot start the Docker engine (Homebrew only installs the %bdocker%b CLI/cask).\n' \
                    "${MAC_D_DIM}" "${MAC_D_RESET}" "${MAC_D_BOLD}" "${MAC_D_RESET}" "${MAC_D_BOLD}" "${MAC_D_RESET}"
                fail=1
            else
                printf '%b[OK]%b docker\n' "${MAC_D_OK}" "${MAC_D_RESET}"
            fi
            continue
        fi
        if mac_check_cmd "${c}"; then
            if [[ "${c}" == "helm" ]]; then
                local hv
                hv="$(mac_helm_client_semver)"
                if [[ -n "${hv}" ]] && ! mac_helm_version_ok_for_charts "${hv}"; then
                    printf '%b[FAIL]%b %s (version %s; need >= %s for Kafka/Bitnami charts — e.g. %bbrew install helm%b or re-run %bget-helm-3%b with HELM_VERSION=v3.19.0)\n' \
                        "${MAC_D_BAD}" "${MAC_D_RESET}" "${c}" "${hv}" "${MAC_HELM_MIN_SEMVER}" "${MAC_D_BOLD}" "${MAC_D_RESET}" "${MAC_D_BOLD}" "${MAC_D_RESET}"
                    fail=1
                else
                    printf '%b[OK]%b %s' "${MAC_D_OK}" "${MAC_D_RESET}" "${c}"
                    [[ -n "${hv:-}" ]] && printf ' %s' "${hv}"
                    printf '\n'
                fi
            else
                printf '%b[OK]%b %s\n' "${MAC_D_OK}" "${MAC_D_RESET}" "${c}"
            fi
        else
            brew_fix_useful=1
            hint="$(mac_doctor_hint_for_tool "${c}")"
            printf '%b[FAIL]%b %s\n' "${MAC_D_BAD}" "${MAC_D_RESET}" "${c}"
            printf '  %bto fix:%b %b%s%b\n' "${MAC_D_DIM}" "${MAC_D_RESET}" "${MAC_D_BOLD}" "${hint}" "${MAC_D_RESET}"
            fail=1
        fi
    done

    if mac_doctor_node_ok; then
        printf '%b[OK]%b node %s\n' "${MAC_D_OK}" "${MAC_D_RESET}" "$(node -v 2>/dev/null || true)"
    elif command -v node >/dev/null 2>&1; then
        brew_fix_useful=1
        hint="$(mac_doctor_hint_for_tool node)"
        printf '%b[WARNING]%b node (need major >= 22, found %s)\n' "${MAC_D_WARN}" "${MAC_D_RESET}" "$(node -v 2>/dev/null || true)"
        printf '  %bto fix:%b %b%s%b\n' "${MAC_D_DIM}" "${MAC_D_RESET}" "${MAC_D_BOLD}" "${hint}" "${MAC_D_RESET}"
        fail=1
    else
        brew_fix_useful=1
        hint="$(mac_doctor_hint_for_tool node)"
        printf '%b[FAIL]%b node\n' "${MAC_D_BAD}" "${MAC_D_RESET}"
        printf '  %bto fix:%b %b%s%b\n' "${MAC_D_DIM}" "${MAC_D_RESET}" "${MAC_D_BOLD}" "${hint}" "${MAC_D_RESET}"
        fail=1
    fi

    export MAC_DOCTOR_DOCKER_DAEMON_DOWN="${docker_daemon_down}"
    export MAC_DOCTOR_BREW_FIX_USEFUL="${brew_fix_useful}"

    if [[ "${fail}" -ne 0 ]]; then
        printf '\n' >&2
        printf '%b[FAIL]%b doctor: toolchain not ready.\n' "${MAC_D_BAD}" "${MAC_D_RESET}" >&2
        printf '        %bNote:%b execute each %bto fix:%b command listed above.\n' \
            "${MAC_D_DIM}" "${MAC_D_RESET}" "${MAC_D_BOLD}" "${MAC_D_RESET}" >&2
        if [[ "${brew_fix_useful}" == "1" ]] && [[ -n "${MAC_DOCTOR_FIX_CMD:-}" ]]; then
            printf '  %bOr run:%b %s\n' "${MAC_D_BOLD}" "${MAC_D_RESET}" "${MAC_DOCTOR_FIX_CMD}" >&2
            if [[ -n "${MAC_DOCTOR_FIX_CMD_AUTO:-}" ]]; then
                printf '           %b(no prompt:%b %s)\n' "${MAC_D_DIM}" "${MAC_D_RESET}" "${MAC_DOCTOR_FIX_CMD_AUTO}" >&2
            fi
        elif [[ "${brew_fix_useful}" == "1" ]] && [[ -z "${MAC_DOCTOR_FIX_CMD:-}" ]]; then
            printf '  %bOr run:%b %bdoctor --fix%b (Homebrew installs what is missing).\n' "${MAC_D_BOLD}" "${MAC_D_RESET}" "${MAC_D_BOLD}" "${MAC_D_RESET}" >&2
        elif [[ "${docker_daemon_down}" == "1" ]]; then
            printf '  %bThen:%b run %bdoctor%b again (no need for %b--fix%b if nothing else is missing).\n' \
                "${MAC_D_DIM}" "${MAC_D_RESET}" "${MAC_D_BOLD}" "${MAC_D_RESET}" "${MAC_D_BOLD}" "${MAC_D_RESET}" >&2
        fi
        return 1
    fi
    printf '%b[OK]%b doctor: toolchain ready.\n' "${MAC_D_OK}" "${MAC_D_RESET}"
    if [[ "${MAC_DOCTOR_HINT_NEXT_STEPS:-false}" == "true" ]]; then
        printf '\n'
        printf '  %bNext:%b run from your %bdeploy/%b directory:\n' "${MAC_D_DIM}" "${MAC_D_RESET}" "${MAC_D_BOLD}" "${MAC_D_RESET}"
        printf '    bash ./dev/mac.sh cluster up\n'
        printf '    bash ./dev/mac.sh kweaver-core install\n'
        printf '  %bOptional:%b bash ./dev/mac.sh onboard -y\n' "${MAC_D_DIM}" "${MAC_D_RESET}"
        printf '  %bGuide:%b deploy/dev/README.md\n' "${MAC_D_DIM}" "${MAC_D_RESET}"
    fi
    return 0
}

mac_kube_context_guard() {
    local expected="kind-${KIND_CLUSTER_NAME}"
    local cur
    local cur_disp
    cur="$(kubectl config current-context 2>/dev/null || true)"
    cur_disp="${cur:-}"
    if [[ -z "${cur_disp}" ]]; then
        cur_disp="(none)"
    fi
    if [[ "${cur}" != "${expected}" ]]; then
        mac_doctor_theme
        printf '%b[WARNING]%b kubectl context is %s, expected %s.\n' "${MAC_D_WARN}" "${MAC_D_RESET}" "'${cur_disp}'" "'${expected}'" >&2
        printf '%b[WARNING]%b Run: kubectl config use-context %s\n' "${MAC_D_WARN}" "${MAC_D_RESET}" "${expected}" >&2
        return 1
    fi
    return 0
}
