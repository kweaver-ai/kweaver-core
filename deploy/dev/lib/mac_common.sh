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

    # Mac dev (kind on Docker Desktop) is memory-tight; cap data-services to small footprints
    # so the whole stack fits in the default Docker memory budget. Chart defaults are kept for
    # k8s/kubeadm. Users can still override any of these via env before invoking mac.sh.
    # redis (chart default 4GB / 512Mi req)
    export REDIS_MAXMEMORY="${REDIS_MAXMEMORY:-512mb}"
    export REDIS_MEMORY_REQUEST="${REDIS_MEMORY_REQUEST:-128Mi}"
    export REDIS_MEMORY_LIMIT="${REDIS_MEMORY_LIMIT:-512Mi}"
    export REDIS_CPU_REQUEST="${REDIS_CPU_REQUEST:-50m}"
    # opensearch (k8s default: req=512Mi, lim=2048Mi)
    export OPENSEARCH_MEMORY_REQUEST="${OPENSEARCH_MEMORY_REQUEST:-512Mi}"
    export OPENSEARCH_MEMORY_LIMIT="${OPENSEARCH_MEMORY_LIMIT:-1024Mi}"
    # zookeeper (k8s default: req cpu=500m mem=1Gi, lim cpu=1 mem=2Gi, jvm 500m)
    # Note: chart hard-codes a zookeeper-exporter sidecar at req/lim 100m/100Mi (no knob).
    export ZOOKEEPER_RESOURCES_REQUESTS_CPU="${ZOOKEEPER_RESOURCES_REQUESTS_CPU:-50m}"
    export ZOOKEEPER_RESOURCES_REQUESTS_MEMORY="${ZOOKEEPER_RESOURCES_REQUESTS_MEMORY:-64Mi}"
    export ZOOKEEPER_RESOURCES_LIMITS_CPU="${ZOOKEEPER_RESOURCES_LIMITS_CPU:-300m}"
    export ZOOKEEPER_RESOURCES_LIMITS_MEMORY="${ZOOKEEPER_RESOURCES_LIMITS_MEMORY:-256Mi}"
    export ZOOKEEPER_JVMFLAGS="${ZOOKEEPER_JVMFLAGS:--Xms64m -Xmx128m}"
    # kweaver-core app services (chart defaults: limits=4-8Gi, mostly request=0).
    # Tiny request keeps QoS=Burstable (not BestEffort) without hogging scheduling budget;
    # generous 2Gi limit so heavier services (agent-retrieval, ontology-query) don't OOM
    # during normal dev exercises.
    export KWEAVER_CORE_REQ_CPU="${KWEAVER_CORE_REQ_CPU:-50m}"
    export KWEAVER_CORE_REQ_MEM="${KWEAVER_CORE_REQ_MEM:-64Mi}"
    export KWEAVER_CORE_LIM_CPU="${KWEAVER_CORE_LIM_CPU:-2}"
    export KWEAVER_CORE_LIM_MEM="${KWEAVER_CORE_LIM_MEM:-2Gi}"
    # ISF (chart defaults: limits 1-8Gi). Symmetric with core; only used when --auth.enabled=true.
    export KWEAVER_ISF_REQ_CPU="${KWEAVER_ISF_REQ_CPU:-50m}"
    export KWEAVER_ISF_REQ_MEM="${KWEAVER_ISF_REQ_MEM:-64Mi}"
    export KWEAVER_ISF_LIM_CPU="${KWEAVER_ISF_LIM_CPU:-2}"
    export KWEAVER_ISF_LIM_MEM="${KWEAVER_ISF_LIM_MEM:-2Gi}"
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

# Inspect Docker engine memory budget and warn when it is too low for KWeaver Core
# + bundled data services (mariadb/redis/kafka/zookeeper/opensearch). Warning only —
# does NOT set fail=1, since the user can still proceed (just slower / OOM-prone).
# Threshold defaults are tuned for --minimum profile + data-services on kind:
#   < MIN  -> WARNING (highly likely to OOM-loop, e.g. redis crash-restart)
#   < REC -> WARNING + hint (below recommended budget for full Core + data-services)
#   >= REC -> OK
# Override via MAC_DOCTOR_MIN_MEM_GB / MAC_DOCTOR_REC_MEM_GB.
mac_doctor_check_docker_memory() {
    local mem_bytes mem_gb min_gb rec_gb
    mem_bytes="$(docker info --format '{{.MemTotal}}' 2>/dev/null || true)"
    if [[ -z "${mem_bytes}" ]] || ! [[ "${mem_bytes}" =~ ^[0-9]+$ ]] || [[ "${mem_bytes}" -eq 0 ]]; then
        return 0
    fi
    mem_gb="$(awk -v b="${mem_bytes}" 'BEGIN{printf "%.1f", b/1024/1024/1024}')"
    min_gb="${MAC_DOCTOR_MIN_MEM_GB:-12}"
    rec_gb="${MAC_DOCTOR_REC_MEM_GB:-16}"
    if (( $(awk -v m="${mem_gb}" -v t="${min_gb}" 'BEGIN{print (m+0 < t+0)}') )); then
        printf '%b[WARNING]%b docker memory %s GB < %s GB minimum (KWeaver Core + data-services likely to OOM; redis/bkn-backend will crash-restart)\n' \
            "${MAC_D_WARN}" "${MAC_D_RESET}" "${mem_gb}" "${min_gb}"
        printf '  %bto fix:%b Docker Desktop → Settings → Resources → %bMemory ≥ %s GB%b → Apply & restart\n' \
            "${MAC_D_DIM}" "${MAC_D_RESET}" "${MAC_D_BOLD}" "${rec_gb}" "${MAC_D_RESET}"
    elif (( $(awk -v m="${mem_gb}" -v t="${rec_gb}" 'BEGIN{print (m+0 < t+0)}') )); then
        printf '%b[WARNING]%b docker memory %s GB — recommend ≥ %s GB for full Core + data-services\n' \
            "${MAC_D_WARN}" "${MAC_D_RESET}" "${mem_gb}" "${rec_gb}"
        printf '  %bto fix:%b Docker Desktop → Settings → Resources → %bMemory ≥ %s GB%b → Apply & restart\n' \
            "${MAC_D_DIM}" "${MAC_D_RESET}" "${MAC_D_BOLD}" "${rec_gb}" "${MAC_D_RESET}"
    else
        printf '%b[OK]%b docker memory %s GB\n' "${MAC_D_OK}" "${MAC_D_RESET}" "${mem_gb}"
    fi
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
                mac_doctor_check_docker_memory
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

# Mac-only: prepare HTTPS for ISF install.
#   1. flip mac-config.yaml accessAddress to scheme=https / port=443
#   2. generate a self-signed TLS cert (CN/SAN = accessAddress.host) and
#      apply it as Secret kweaver-ingress-tls in the kweaver namespace
# Idempotent: re-running just rotates the cert and re-applies the Secret.
# After ISF install completes, run mac_isf_patch_ingress_tls to wire it in.
mac_prepare_isf_https() {
    local cfg="${CONFIG_YAML_PATH:-${MAC_DEV_ROOT}/conf/mac-config.yaml}"
    [[ -f "${cfg}" ]] || { mac_log_error "mac-config not found: ${cfg}"; return 1; }
    local host
    host="$(awk '/^accessAddress:/{f=1;next} f&&/^  host:/{print $2;exit}' "${cfg}" | tr -d "'\"")"
    host="${host:-localhost}"
    local ns="${MAC_ISF_NAMESPACE:-$(awk '/^namespace:/{print $2;exit}' "${cfg}" | tr -d "'\"")}"
    ns="${ns:-kweaver}"
    local secret="${MAC_ISF_TLS_SECRET:-kweaver-ingress-tls}"

    mac_log_info "Switching ${cfg} accessAddress to https/443 (host=${host})"
    awk '
        /^accessAddress:/ {in_a=1; print; next}
        in_a && /^[a-zA-Z]/ {in_a=0}
        in_a && /^[[:space:]]*scheme:/ {sub(/scheme:.*/, "scheme: https"); print; next}
        in_a && /^[[:space:]]*port:/   {sub(/port:.*/,   "port: 443");     print; next}
        {print}
    ' "${cfg}" > "${cfg}.tmp" && mv "${cfg}.tmp" "${cfg}"

    mac_log_info "Generating self-signed TLS cert for ${host}"
    local tmp; tmp="$(mktemp -d)"
    openssl req -x509 -nodes -newkey rsa:2048 -days 825 \
        -keyout "${tmp}/tls.key" -out "${tmp}/tls.crt" \
        -subj "/CN=${host}" -addext "subjectAltName=DNS:${host},DNS:localhost,IP:127.0.0.1" \
        >/dev/null 2>&1 || { mac_log_error "openssl failed (install via brew install openssl)"; return 1; }
    kubectl create namespace "${ns}" --dry-run=client -o yaml | kubectl apply -f - >/dev/null
    kubectl create secret tls "${secret}" --cert="${tmp}/tls.crt" --key="${tmp}/tls.key" \
        -n "${ns}" --dry-run=client -o yaml | kubectl apply -f - >/dev/null
    rm -rf "${tmp}"
    mac_log_info "TLS Secret ${ns}/${secret} ready (cert valid 825 days, CN=${host})"

    # If kweaver-core releases already exist they were rendered with the old http
    # accessAddress; refresh them so the in-cluster URLs/issuers match the new https.
    if helm list -n "${ns}" -q 2>/dev/null | grep -qE '.'; then
        mac_log_info "kweaver-core releases already in ${ns}; running 'kweaver-core install --minimum' to refresh accessAddress (https/443)"
        bash "${DEPLOY_ROOT}/deploy.sh" kweaver-core install --minimum || \
            mac_log_warn "kweaver-core refresh failed; you may need to re-run 'mac.sh kweaver-core install --minimum' manually"
    fi
}

# Patch the ISF ingress to attach the TLS Secret. Run AFTER `deploy.sh isf install`.
mac_isf_patch_ingress_tls() {
    local cfg="${CONFIG_YAML_PATH:-${MAC_DEV_ROOT}/conf/mac-config.yaml}"
    local ns="${MAC_ISF_NAMESPACE:-$(awk '/^namespace:/{print $2;exit}' "${cfg}" | tr -d "'\"")}"
    ns="${ns:-kweaver}"
    local secret="${MAC_ISF_TLS_SECRET:-kweaver-ingress-tls}"
    local ing
    ing="$(kubectl get ingress -n "${ns}" -o jsonpath='{.items[?(@.metadata.name=="ingress-informationsecurityfabric")].metadata.name}' 2>/dev/null)"
    [[ -n "${ing}" ]] || { mac_log_warn "ISF ingress not found in ${ns}; skip TLS patch"; return 0; }
    local host
    host="$(awk '/^accessAddress:/{f=1;next} f&&/^  host:/{print $2;exit}' "${CONFIG_YAML_PATH:-${MAC_DEV_ROOT}/conf/mac-config.yaml}" | tr -d "'\"")"
    host="${host:-localhost}"
    kubectl patch ingress "${ing}" -n "${ns}" --type=merge -p \
        "{\"spec\":{\"tls\":[{\"hosts\":[\"${host}\"],\"secretName\":\"${secret}\"}]}}" >/dev/null \
        && mac_log_info "✓ ISF ingress patched with TLS (host=${host}, secret=${secret})"
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
