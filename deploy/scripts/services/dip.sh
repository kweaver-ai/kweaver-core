
# KWeaver DIP (Data Intelligence Platform) releases list
# Chart names correspond to the tgz files in docs/kweaver-dip/charts/
declare -a DIP_RELEASES=(
    "anyfabric-frontend"
    "auth-service"
    "basic-search"
    "configuration-center"
    "data-application-gateway"
    "data-application-service"
    "data-catalog"
    "data-exploration-service"
    "data-semantic"
    "data-subject"
    "data-view"
    "sailor"
    "sailor-agent"
    "sailor-service"
    "session"
    "standardization"
    "task-center"
)

# Default DIP namespace
DIP_NAMESPACE="${DIP_NAMESPACE:-kweaver-ai}"

# Default local DIP charts directory (relative to deploy root)
DIP_LOCAL_CHARTS_DIR="${DIP_LOCAL_CHARTS_DIR:-}"

# Parse dip command arguments
parse_dip_args() {
    local action="$1"
    shift

    # Parse arguments to override defaults
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
                DIP_LOCAL_CHARTS_DIR="${1#*=}"
                shift
                ;;
            --charts_dir)
                DIP_LOCAL_CHARTS_DIR="$2"
                shift 2
                ;;
            --namespace=*)
                DIP_NAMESPACE="${1#*=}"
                shift
                ;;
            --namespace)
                DIP_NAMESPACE="$2"
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
            *)
                log_error "Unknown argument: $1"
                return 1
                ;;
        esac
    done
}

# Check if a single Helm release is deployed in the given namespace
_dip_helm_release_exists() {
    local release="$1"
    local ns="$2"
    helm list -n "${ns}" --short 2>/dev/null | grep -q "^${release}$"
}

# Ensure kweaver-core modules are installed.
# Core = ISF + KWEAVER_CORE_RELEASES (defined in core.sh)
# Use available installer entrypoints: install_isf and install_core.
_dip_ensure_kweaver_core() {
    local namespace
    namespace=$(grep "^namespace:" "${CONFIG_YAML_PATH}" 2>/dev/null | head -1 | awk '{print $2}' | tr -d "'\"")
    namespace="${namespace:-kweaver-ai}"

    log_info "Checking kweaver-core dependencies for KWeaver DIP..."

    local missing_isf=false
    local missing_core=false

    if _dip_helm_release_exists "hydra" "${namespace}"; then
        log_info "  ✓ ISF already installed (hydra found)"
    else
        log_info "  ✗ ISF not installed — installing now..."
        missing_isf=true
    fi

    local release_name
    for release_name in "${KWEAVER_CORE_RELEASES[@]}"; do
        if _dip_helm_release_exists "${release_name}" "${namespace}"; then
            log_info "  ✓ Core release already installed (${release_name})"
        else
            log_info "  ✗ Core release not installed (${release_name})"
            missing_core=true
        fi
    done

    if [[ "${missing_isf}" == "true" ]]; then
        if ! install_isf; then
            log_error "Failed to install kweaver-core module: ISF"
            return 1
        fi
    fi

    if [[ "${missing_core}" == "true" ]]; then
        log_info "Installing missing KWeaver Core releases..."
        if ! install_core; then
            log_error "Failed to install missing KWeaver Core releases"
            return 1
        fi
    fi

    if [[ "${missing_isf}" == "false" && "${missing_core}" == "false" ]]; then
        log_info "All kweaver-core dependencies are satisfied."
    else
        log_info "kweaver-core dependency installation completed."
    fi
}

# Detect whether K8s is already running and accessible
_dip_k8s_is_running() {
    if ! command -v kubectl >/dev/null 2>&1; then
        return 1
    fi
    kubectl get nodes >/dev/null 2>&1
}

# Ensure K8s is up; install it if not present
_dip_ensure_k8s() {
    if _dip_k8s_is_running; then
        log_info "Kubernetes cluster detected, skipping K8s installation."
        return 0
    fi

    log_info "No running Kubernetes cluster detected. Installing K8s first..."
    detect_package_manager
    install_containerd
    install_kubernetes
    install_helm

    check_prerequisites
    init_k8s_master
    allow_master_scheduling
    install_cni
    wait_for_dns

    if [[ "${AUTO_INSTALL_LOCALPV}" == "true" ]]; then
        if [[ -z "$(kubectl get storageclass --no-headers 2>/dev/null)" ]]; then
            install_localpv
        fi
    fi

    if [[ "${AUTO_INSTALL_INGRESS_NGINX}" == "true" ]]; then
        install_ingress_nginx
    fi

    log_info "K8s installation completed."
}

# Ensure data services are installed for DIP runtime dependencies.
# This keeps `kweaver-dip install` behavior aligned with infra/full flows.
_dip_ensure_data_services() {
    log_info "Ensuring data services for KWeaver DIP (MariaDB/Redis/Kafka/Zookeeper/OpenSearch)..."

    install_mariadb
    install_redis
    install_kafka
    install_zookeeper
    # install_mongodb  # MongoDB disabled
    if [[ "${AUTO_INSTALL_INGRESS_NGINX}" == "true" ]]; then
        install_ingress_nginx
    fi
    install_opensearch

    if [[ "${AUTO_GENERATE_CONFIG}" == "true" ]]; then
        generate_config_yaml
    fi
}

# Resolve the local charts directory: use explicit override, or auto-discover
# relative to SCRIPT_DIR (deploy root).
_dip_resolve_charts_dir() {
    if [[ -n "${DIP_LOCAL_CHARTS_DIR}" ]]; then
        echo "${DIP_LOCAL_CHARTS_DIR}"
        return 0
    fi

    # Auto-discover: look for charts next to deploy/ (sibling docs/ dir)
    local candidates=(
        "${SCRIPT_DIR}/../docs/kweaver-dip/charts"
        "${SCRIPT_DIR}/charts/kweaver-dip"
        "${LOCAL_CHARTS_DIR}/kweaver-dip"
    )
    for candidate in "${candidates[@]}"; do
        if [[ -d "${candidate}" ]]; then
            echo "$(cd "${candidate}" && pwd)"
            return 0
        fi
    done

    echo ""
    # No local charts found: caller should fallback to Helm repo mode.
    return 0
}

# Find a local tgz for a chart name inside a directory (picks the first match)
_dip_find_local_chart() {
    local charts_dir="$1"
    local chart_name="$2"
    local tgz
    tgz=$(find "${charts_dir}" -maxdepth 1 -name "${chart_name}-*.tgz" 2>/dev/null | sort -V | tail -1)
    echo "${tgz}"
}

# Install DIP services via Helm
install_dip() {
    log_info "Installing KWeaver DIP services via Helm..."

    # Ensure K8s is running (install if absent)
    _dip_ensure_k8s

    # Ensure data services are ready before installing core and DIP services
    _dip_ensure_data_services

    # Ensure kweaver-core dependencies are installed
    if ! _dip_ensure_kweaver_core; then
        log_error "kweaver-core dependency check/installation failed"
        return 1
    fi

    # Resolve namespace
    local namespace
    namespace=$(grep "^namespace:" "${CONFIG_YAML_PATH}" 2>/dev/null | head -1 | awk '{print $2}' | tr -d "'\"")
    namespace="${namespace:-${DIP_NAMESPACE}}"

    # Create namespace if not exists
    kubectl create namespace "${namespace}" 2>/dev/null || true

    # Resolve chart source: local directory takes priority over remote repo
    local charts_dir
    charts_dir="$(_dip_resolve_charts_dir)"

    local use_local=false
    if [[ -n "${charts_dir}" && -d "${charts_dir}" ]]; then
        use_local=true
        log_info "Using local DIP charts from: ${charts_dir}"
    else
        # Auto-derive repo name from URL last path segment if not explicitly set
        if [[ "${HELM_CHART_REPO_NAME}" == "kweaver" && -n "${HELM_CHART_REPO_URL}" ]]; then
            HELM_CHART_REPO_NAME="${HELM_CHART_REPO_URL##*/}"
        fi
        log_info "No local DIP charts directory found, using Helm repo."
        log_info "  Version:   ${HELM_CHART_VERSION:-latest}"
        log_info "  Helm Repo: ${HELM_CHART_REPO_NAME} -> ${HELM_CHART_REPO_URL}"
        helm repo add --force-update "${HELM_CHART_REPO_NAME}" "${HELM_CHART_REPO_URL}" || true
        helm repo update "${HELM_CHART_REPO_NAME}" || true
    fi

    log_info "Target namespace: ${namespace}"

    # Install each release
    for release_name in "${DIP_RELEASES[@]}"; do
        if [[ "${use_local}" == "true" ]]; then
            _install_dip_release_local "${release_name}" "${charts_dir}" "${namespace}"
        else
            _install_dip_release_repo "${release_name}" "${namespace}" "${HELM_CHART_REPO_NAME}" "${HELM_CHART_VERSION}"
        fi
    done

    log_info "KWeaver DIP services installation completed."
}

# Install a single DIP release from a local .tgz chart file
_install_dip_release_local() {
    local release_name="$1"
    local charts_dir="$2"
    local namespace="$3"

    local chart_tgz
    chart_tgz="$(_dip_find_local_chart "${charts_dir}" "${release_name}")"

    if [[ -z "${chart_tgz}" ]]; then
        log_error "✗ Local chart not found for ${release_name} in ${charts_dir}"
        return 1
    fi

    log_info "Installing ${release_name} from local chart: $(basename "${chart_tgz}")..."

    local -a helm_args=(
        "upgrade" "--install" "${release_name}"
        "${chart_tgz}"
        "--namespace" "${namespace}"
        "-f" "${CONFIG_YAML_PATH}"
        "--wait" "--timeout=600s"
    )

    if helm "${helm_args[@]}"; then
        log_info "✓ ${release_name} installed successfully"
    else
        log_error "✗ Failed to install ${release_name}"
        return 1
    fi
}

# Install a single DIP release from a Helm repository
_install_dip_release_repo() {
    local release_name="$1"
    local namespace="$2"
    local helm_repo_name="$3"
    local release_version="$4"

    # Clean up any pending state before installing
    local current_status
    current_status=$(helm status "${release_name}" -n "${namespace}" -o json 2>/dev/null | grep -o '"status":"[^"]*"' | head -1 | cut -d'"' -f4)
    if [[ -n "${current_status}" && "${current_status}" != "deployed" && "${current_status}" != "failed" ]]; then
        log_info "Cleaning up ${release_name} (status: ${current_status})..."
        helm uninstall "${release_name}" -n "${namespace}" 2>/dev/null || true
    fi

    log_info "Installing ${release_name} from repo..."

    local chart_ref="${helm_repo_name}/${release_name}"

    local -a helm_args=(
        "upgrade" "--install" "${release_name}"
        "${chart_ref}"
        "--namespace" "${namespace}"
        "-f" "${CONFIG_YAML_PATH}"
    )

    if [[ -n "${release_version}" ]]; then
        helm_args+=("--version" "${release_version}")
    fi

    helm_args+=("--devel" "--wait" "--timeout=600s")

    if helm "${helm_args[@]}"; then
        log_info "✓ ${release_name} installed successfully"
    else
        log_error "✗ Failed to install ${release_name}"
        return 1
    fi
}

# Uninstall DIP services
uninstall_dip() {
    log_info "Uninstalling KWeaver DIP services..."

    local namespace
    namespace=$(grep "^namespace:" "${CONFIG_YAML_PATH}" 2>/dev/null | head -1 | awk '{print $2}' | tr -d "'\"")
    namespace="${namespace:-${DIP_NAMESPACE}}"

    # Uninstall in reverse order
    for ((i=${#DIP_RELEASES[@]}-1; i>=0; i--)); do
        local release_name="${DIP_RELEASES[$i]}"
        log_info "Uninstalling ${release_name}..."
        if helm uninstall "${release_name}" -n "${namespace}" 2>/dev/null; then
            log_info "✓ ${release_name} uninstalled successfully"
        else
            log_warn "⚠ ${release_name} not found or already uninstalled"
        fi
    done

    log_info "KWeaver DIP services uninstallation completed."
}

# Show DIP services status
show_dip_status() {
    log_info "KWeaver DIP services status:"

    local namespace
    namespace=$(grep "^namespace:" "${CONFIG_YAML_PATH}" 2>/dev/null | head -1 | awk '{print $2}' | tr -d "'\"")
    namespace="${namespace:-${DIP_NAMESPACE}}"

    log_info "Namespace: ${namespace}"
    log_info ""

    for release_name in "${DIP_RELEASES[@]}"; do
        if helm status "${release_name}" -n "${namespace}" >/dev/null 2>&1; then
            local status
            status=$(helm status "${release_name}" -n "${namespace}" -o json 2>/dev/null | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
            log_info "  ✓ ${release_name}: ${status}"
        else
            log_info "  ✗ ${release_name}: not installed"
        fi
    done
}
