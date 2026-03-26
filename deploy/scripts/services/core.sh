
# KWeaver Core releases list
# Merged from: studio, ontology, agentoperator, dataagent, decisionagent, flowautomation, sandboxruntime
# Note: ISF releases are managed separately by isf.sh
declare -a KWEAVER_CORE_RELEASES=(
    # studio
    "deploy-web"
    "studio-web"
    "business-system-frontend"
    "business-system-service"
    "mf-model-manager-nginx"
    "mf-model-manager"
    "mf-model-api"
    # ontology
    "vega-backend"
    "bkn-backend"
    "ontology-query"
    "vega-web"
    "data-connection"
    "vega-gateway"
    "vega-gateway-pro"
    "mdl-data-model"
    "mdl-uniquery"
    "mdl-data-model-job"
    # agentoperator
    "agent-operator-integration"
    "operator-web"
    "agent-retrieval"
    "data-retrieval"
    # dataagent
    "agent-backend"
    "agent-web"
    # flowautomation
    "flow-web"
    "dataflow"
    "coderunner"
    # sandboxruntime
    "sandbox"
)

# Default kweaver-core namespace
CORE_NAMESPACE="${CORE_NAMESPACE:-kweaver-ai}"

# release name -> chart name mapping (when chart name differs from release name)
declare -A CORE_CHART_NAME_MAP=(
)

# Default local charts directory
CORE_LOCAL_CHARTS_DIR="${CORE_LOCAL_CHARTS_DIR:-}"

# Parse kweaver-core command arguments
parse_core_args() {
    local action="$1"
    shift

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --version=*)
                HELM_CHART_VERSION="${1#*=}"
                shift
                ;;
            --version)
                HELM_CHART_VERSION="$2"
                shift 2
                ;;
            --helm_repo=*)
                HELM_CHART_REPO_URL="${1#*=}"
                shift
                ;;
            --helm_repo)
                HELM_CHART_REPO_URL="$2"
                shift 2
                ;;
            --helm_repo_name=*)
                HELM_CHART_REPO_NAME="${1#*=}"
                shift
                ;;
            --helm_repo_name)
                HELM_CHART_REPO_NAME="$2"
                shift 2
                ;;
            --charts_dir=*)
                CORE_LOCAL_CHARTS_DIR="${1#*=}"
                shift
                ;;
            --charts_dir)
                CORE_LOCAL_CHARTS_DIR="$2"
                shift 2
                ;;
            --namespace=*)
                CORE_NAMESPACE="${1#*=}"
                shift
                ;;
            --namespace)
                CORE_NAMESPACE="$2"
                shift 2
                ;;
            --config=*)
                CONFIG_YAML_PATH="${1#*=}"
                shift
                ;;
            --config)
                CONFIG_YAML_PATH="$2"
                shift 2
                ;;
            --enable-isf=*)
                ENABLE_ISF="${1#*=}"
                shift
                ;;
            --enable-isf)
                ENABLE_ISF="$2"
                shift 2
                ;;
            *)
                log_error "Unknown argument: $1"
                return 1
                ;;
        esac
    done
}

# Resolve local charts directory for kweaver-core
_core_resolve_charts_dir() {
    if [[ -n "${CORE_LOCAL_CHARTS_DIR}" ]]; then
        # --charts_dir was explicitly set; only use it if it exists
        if [[ -d "${CORE_LOCAL_CHARTS_DIR}" ]]; then
            echo "${CORE_LOCAL_CHARTS_DIR}"
        fi
        return
    fi
    echo ""
}

# Find local chart tgz for a given release name
_core_find_local_chart() {
    local charts_dir="$1"
    local release_name="$2"
    find "${charts_dir}" -maxdepth 1 -name "${release_name}-*.tgz" 2>/dev/null | sort -V | tail -1
}

# Install a single kweaver-core release from a local .tgz
_install_core_release_local() {
    local release_name="$1"
    local charts_dir="$2"
    local namespace="$3"

    local chart_tgz
    chart_tgz="$(_core_find_local_chart "${charts_dir}" "${release_name}")"

    if [[ -z "${chart_tgz}" ]]; then
        log_error "✗ Local chart not found for ${release_name} in ${charts_dir}"
        return 1
    fi

    local target_version
    target_version=$(get_local_chart_version "${chart_tgz}")
    if should_skip_upgrade_same_chart_version "${release_name}" "${namespace}" "${release_name}" "${target_version}"; then
        return 0
    fi

    log_info "Installing ${release_name} from local chart: $(basename "${chart_tgz}")..."

    if helm upgrade --install "${release_name}" "${chart_tgz}" \
            --namespace "${namespace}" \
            -f "${CONFIG_YAML_PATH}" \
            --wait --timeout=600s; then
        log_info "✓ ${release_name} installed successfully"
    else
        log_error "✗ Failed to install ${release_name}"
        return 1
    fi
}

# Install a single kweaver-core release from a Helm repository
_install_core_release_repo() {
    local release_name="$1"
    local namespace="$2"
    local helm_repo_name="$3"
    local release_version="$4"

    # Resolve actual chart name (may differ from release name)
    local chart_name="${release_name}"
    if [[ -n "${CORE_CHART_NAME_MAP[${release_name}]+_}" ]]; then
        chart_name="${CORE_CHART_NAME_MAP[${release_name}]}"
    fi

    local chart_ref="${helm_repo_name}/${chart_name}"

    local target_version="${release_version}"
    if [[ -z "${target_version}" ]]; then
        target_version=$(get_repo_chart_latest_version "${helm_repo_name}" "${chart_name}")
    fi

    if should_skip_upgrade_same_chart_version "${release_name}" "${namespace}" "${chart_name}" "${target_version}"; then
        return 0
    fi

    # Clean up any pending state before installing
    local current_status
    current_status=$(helm status "${release_name}" -n "${namespace}" -o json 2>/dev/null | grep -o '"status":"[^"]*"' | head -1 | cut -d'"' -f4)
    if [[ -n "${current_status}" && "${current_status}" != "deployed" && "${current_status}" != "failed" ]]; then
        log_info "Cleaning up ${release_name} (status: ${current_status})..."
        helm uninstall "${release_name}" -n "${namespace}" 2>/dev/null || true
    fi

    log_info "Installing ${release_name} from ${chart_ref}..."

    local -a helm_args=(
        "upgrade" "--install" "${release_name}"
        "${chart_ref}"
        "--namespace" "${namespace}"
        "-f" "${CONFIG_YAML_PATH}"
    )

    if [[ -n "${release_version}" ]]; then
        helm_args+=("--version" "${release_version}")
    fi

    helm_args+=("--devel")

    if helm "${helm_args[@]}"; then
        log_info "✓ ${release_name} installed successfully"
    else
        log_error "✗ Failed to install ${release_name}"
        return 1
    fi
}

# Install KWeaver Core services via Helm
install_core() {
    log_info "Installing KWeaver Core services via Helm..."

    local namespace
    namespace=$(grep "^namespace:" "${CONFIG_YAML_PATH}" 2>/dev/null | head -1 | awk '{print $2}' | tr -d "'\"")
    namespace="${namespace:-${CORE_NAMESPACE}}"

    kubectl create namespace "${namespace}" 2>/dev/null || true

    local charts_dir
    charts_dir="$(_core_resolve_charts_dir)"

    local use_local=false
    if [[ -n "${charts_dir}" && -d "${charts_dir}" ]]; then
        use_local=true
        log_info "Using local Core charts from: ${charts_dir}"
    else
        log_info "No local Core charts directory found, using Helm repo."
        log_info "  Version:   ${HELM_CHART_VERSION}"
        log_info "  Helm Repo: ${HELM_CHART_REPO_NAME:-kweaver} -> ${HELM_CHART_REPO_URL:-https://kweaver-ai.github.io/helm-repo/}"
        HELM_CHART_REPO_NAME="${HELM_CHART_REPO_NAME:-kweaver}"
        HELM_CHART_REPO_URL="${HELM_CHART_REPO_URL:-https://kweaver-ai.github.io/helm-repo/}"
        helm repo add --force-update "${HELM_CHART_REPO_NAME}" "${HELM_CHART_REPO_URL}" || true
        helm repo update "${HELM_CHART_REPO_NAME}" || true
    fi

    log_info "Target namespace: ${namespace}"

    # Check if ISF should be enabled (default: install ISF)
    if [[ "${ENABLE_ISF}" == "false" ]]; then
        log_info "ISF installation disabled via --enable-isf=false"
    else
        log_info "Installing ISF services (default, use --enable-isf=false to skip)"
        install_isf
    fi

    for release_name in "${KWEAVER_CORE_RELEASES[@]}"; do
        if [[ "${use_local}" == "true" ]]; then
            _install_core_release_local "${release_name}" "${charts_dir}" "${namespace}"
        else
            _install_core_release_repo "${release_name}" "${namespace}" "${HELM_CHART_REPO_NAME}" "${HELM_CHART_VERSION}"
        fi
    done

    log_info "KWeaver Core services installation completed."
}

# Uninstall KWeaver Core services
uninstall_core() {
    log_info "Uninstalling KWeaver Core services..."

    local namespace
    namespace=$(grep "^namespace:" "${CONFIG_YAML_PATH}" 2>/dev/null | head -1 | awk '{print $2}' | tr -d "'\"")
    namespace="${namespace:-${CORE_NAMESPACE}}"

    for ((i=${#KWEAVER_CORE_RELEASES[@]}-1; i>=0; i--)); do
        local release_name="${KWEAVER_CORE_RELEASES[$i]}"
        log_info "Uninstalling ${release_name}..."
        if helm uninstall "${release_name}" -n "${namespace}" 2>/dev/null; then
            log_info "✓ ${release_name} uninstalled"
        else
            log_warn "⚠ ${release_name} not found or already uninstalled"
        fi
    done

    log_info "KWeaver Core services uninstallation completed."
}

# Show KWeaver Core services status
show_core_status() {
    log_info "KWeaver Core services status:"

    local namespace
    namespace=$(grep "^namespace:" "${CONFIG_YAML_PATH}" 2>/dev/null | head -1 | awk '{print $2}' | tr -d "'\"")
    namespace="${namespace:-${CORE_NAMESPACE}}"

    log_info "Namespace: ${namespace}"
    log_info ""

    for release_name in "${KWEAVER_CORE_RELEASES[@]}"; do
        if helm status "${release_name}" -n "${namespace}" >/dev/null 2>&1; then
            local status
            status=$(helm status "${release_name}" -n "${namespace}" -o json 2>/dev/null \
                | grep -o '"status":"[^"]*"' | head -1 | cut -d'"' -f4)
            log_info "  ✓ ${release_name}: ${status}"
        else
            log_info "  ✗ ${release_name}: not installed"
        fi
    done
}
