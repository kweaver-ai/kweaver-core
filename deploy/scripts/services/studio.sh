
# Studio releases list
declare -a STUDIO_RELEASES=(
    "deploy-web"
    "studio-web"
    "business-system-frontend"
    "business-system-service"
    "mf-model-manager-nginx"
    "mf-model-manager"
    "mf-model-api"
)

# Parse studio command arguments
parse_studio_args() {
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
            *)
                log_error "Unknown argument: $1"
                return 1
                ;;
        esac
    done
}

# Initialize Studio database using common database initialization function
init_studio_database() {
    local sql_dir="${SCRIPT_DIR}/scripts/sql/studio"
    
    # Only initialize database if RDS is internal (MariaDB installed in cluster)
    if ! is_rds_internal; then
        warn_external_rds_sql_required "Studio" "${sql_dir}"
        log_warn "Skipping automatic Studio database initialization (external RDS)"
        return 0
    fi
    
    init_module_database "studio" "${sql_dir}"
}

# Install Studio services via Helm
install_studio() {
    log_info "Installing Studio services via Helm..."
    log_info "  Version: ${HELM_CHART_VERSION}"
    log_info "  Helm Repo: ${HELM_CHART_REPO_NAME:-kweaver} -> ${HELM_CHART_REPO_URL:-https://kweaver-ai.github.io/helm-repo/}"

    # Get namespace from config.yaml
    local namespace=$(grep "^namespace:" "${CONFIG_YAML_PATH}" 2>/dev/null | head -1 | awk '{print $2}' | tr -d "'\"")
    namespace="${namespace:-kweaver-ai}"
    
    # Create namespace if not exists
    kubectl create namespace "${namespace}" 2>/dev/null || true
    
    # Add Helm repo with retry
    helm_repo_add_with_retry "${HELM_CHART_REPO_NAME}" "${HELM_CHART_REPO_URL}"
    
    # Initialize database first
    if ! init_studio_database; then
        log_error "Failed to initialize Studio database"
        return 1
    fi
    
    log_info "Target namespace: ${namespace}"
    
    # Install each release
    for release_name in "${STUDIO_RELEASES[@]}"; do
        install_studio_release "${release_name}" "${release_name}" "${namespace}" "${HELM_CHART_REPO_NAME}" "${HELM_CHART_VERSION}"
    done
    
    log_info "Studio services installation completed"
}

# Install a single Studio release
install_studio_release() {
    local release_name="$1"
    local chart_name="$2"
    local namespace="$3"
    local helm_repo_name="$4"
    local default_version="$5"
    
    log_info "Installing ${release_name}..."
    
    # Build Helm chart reference
    local chart_ref="${helm_repo_name}/${chart_name}"
    
    # Build Helm command
    local -a helm_args=(
        "upgrade" "--install" "${release_name}"
        "${chart_ref}"
        "--namespace" "${namespace}"
        "-f" "${SCRIPT_DIR}/conf/config.yaml"
    )
    
    # Add version parameter only if specified
    if [[ -n "${default_version}" ]]; then
        helm_args+=("--version" "${default_version}")
    fi
    
    helm_args+=("--devel" "--wait" "--timeout=600s")
    
    # Execute Helm install/upgrade
    if helm "${helm_args[@]}"; then
        log_info "✓ ${release_name} installed successfully"
    else
        log_error "✗ Failed to install ${release_name}"
        return 1
    fi
}

# Uninstall Studio services
uninstall_studio() {
    log_info "Uninstalling Studio services..."
    
    # Get namespace from config.yaml
    local namespace=$(grep "^namespace:" "${CONFIG_YAML_PATH}" 2>/dev/null | head -1 | awk '{print $2}' | tr -d "'\"")
    namespace="${namespace:-kweaver-ai}"
    
    # Uninstall in reverse order
    for ((i=${#STUDIO_RELEASES[@]}-1; i>=0; i--)); do
        local release_name="${STUDIO_RELEASES[$i]}"
        log_info "Uninstalling ${release_name}..."
        if helm uninstall "${release_name}" -n "${namespace}" 2>/dev/null; then
            log_info "✓ ${release_name} uninstalled successfully"
        else
            log_warn "⚠ ${release_name} not found or already uninstalled"
        fi
    done
    
    log_info "Studio services uninstallation completed"
}

# Show Studio services status
show_studio_status() {
    log_info "Studio services status:"
    
    # Get namespace from config.yaml
    local namespace=$(grep "^namespace:" "${CONFIG_YAML_PATH}" 2>/dev/null | head -1 | awk '{print $2}' | tr -d "'\"")
    namespace="${namespace:-kweaver-ai}"
    
    log_info "Namespace: ${namespace}"
    log_info ""
    
    # Check each release
    for release_name in "${STUDIO_RELEASES[@]}"; do
        if helm status "${release_name}" -n "${namespace}" >/dev/null 2>&1; then
            local status=$(helm status "${release_name}" -n "${namespace}" -o json 2>/dev/null | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
            log_info "  ✓ ${release_name}: ${status}"
        else
            log_info "  ✗ ${release_name}: not installed"
        fi
    done
}
