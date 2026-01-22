
# ISF (Information Security Fabric) releases list
declare -a ISF_RELEASES=(
    hydra
    sharemgnt-single
    user-management
    sharemgnt
    authentication
    policy-management
    audit-log
    eacp
    thirdparty-message-plugin
    isfwebthrift
    message
    isfweb
    authorization
    news-feed
    ingress-informationsecurityfabric
    eacp-single
    oauth2-ui
)

# ISF databases list
declare -a ISF_DATABASES=(
    "user_management"
    "anyshare"
    "policy_mgnt"
    "privacy"
    "authentication"
    "eofs"
    "deploy"
    "sharemgnt_db"
    "ets"
    "ossmanager"
    "license"
    "nodemgnt"
    "sites"
    "anydata"
    "third_app_mgnt"
    "hydra_v2"
    "thirdparty_message"
)

# Parse isf command arguments
parse_isf_args() {
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

# Initialize ISF database using common database initialization function
init_isf_database() {
    local sql_dir="${SCRIPT_DIR}/scripts/sql/isf"
    
    # Only initialize database if RDS is internal (MariaDB installed in cluster)
    if ! is_rds_internal; then
        warn_external_rds_sql_required "ISF" "${sql_dir}"
        log_warn "Skipping automatic ISF database initialization (external RDS)"
        return 0
    fi
    
    init_module_database "isf" "${sql_dir}"
}

# Install ISF services via Helm
install_isf() {
    log_info "Installing ISF services via Helm..."
    log_info "  Version: ${HELM_CHART_VERSION}"
    log_info "  Helm Repo: ${HELM_CHART_REPO_NAME:-kweaver} -> ${HELM_CHART_REPO_URL:-https://kweaver-ai.github.io/helm-repo/}"

    # Get namespace from config.yaml
    local namespace=$(grep "^namespace:" "${CONFIG_YAML_PATH}" 2>/dev/null | head -1 | awk '{print $2}' | tr -d "'\"")
    namespace="${namespace:-kweaver-ai}"
    
    # Create namespace if not exists
    kubectl create namespace "${namespace}" 2>/dev/null || true
    
    # Add Helm repo
    log_info "Adding Helm repo: ${HELM_CHART_REPO_NAME} -> ${HELM_CHART_REPO_URL}"
    helm repo add --force-update "${HELM_CHART_REPO_NAME}" "${HELM_CHART_REPO_URL}"
    helm repo update
    
    # Initialize database first
    if ! init_isf_database; then
        log_error "Failed to initialize ISF database"
        return 1
    fi
    
    log_info "Target namespace: ${namespace}"
    
    # Create temporary config.yaml without rds.database field for ISF services
    local temp_config="${CONFIG_YAML_PATH}.isf.tmp"
    log_info "Creating temporary config.yaml for ISF services (removing rds.database field)..."
    
    # Copy config.yaml and remove all database: lines (both top-level and nested under rds)
    sed '/database:/d' "${CONFIG_YAML_PATH}" > "${temp_config}"
    
    # Temporarily replace CONFIG_YAML_PATH with temp config
    local original_config="${CONFIG_YAML_PATH}"
    export CONFIG_YAML_PATH="${temp_config}"
    
    # Install each release
    local install_failed=0
    for release_name in "${ISF_RELEASES[@]}"; do
        if ! install_isf_release "${release_name}" "${release_name}" "${namespace}" "${HELM_CHART_REPO_NAME}" "${ISF_RELEASE_VERSIONS[$release_name]}" "${temp_config}"; then
            install_failed=1
            break
        fi
    done
    
    # Restore original config path and clean up temp file
    export CONFIG_YAML_PATH="${original_config}"
    if [[ -f "${temp_config}" ]]; then
        log_info "Cleaning up temporary config.yaml..."
        rm -f "${temp_config}"
    fi
    
    if [[ ${install_failed} -eq 1 ]]; then
        log_error "ISF services installation failed"
        return 1
    fi
    
    log_info "ISF services installation completed"
}

# Install a single ISF release
install_isf_release() {
    local release_name="$1"
    local chart_name="$2"
    local namespace="$3"
    local helm_repo_name="$4"
    local release_version="$5"
    local values_file="${6:-${SCRIPT_DIR}/conf/config.yaml}"
    
    log_info "Installing ${release_name}..."
    
    # Build Helm chart reference
    local chart_ref="${helm_repo_name}/${chart_name}"
    
    # Build Helm command
    local -a helm_args=(
        "upgrade" "--install" "${release_name}"
        "${chart_ref}"
        "--namespace" "${namespace}"
        "-f" "${values_file}"
    )
    
    # Add version parameter only if specified
    if [[ -n "${release_version}" ]]; then
        helm_args+=("--version" "${release_version}")
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

# Uninstall ISF services
uninstall_isf() {
    log_info "Uninstalling ISF services..."
    
    # Get namespace from config.yaml
    local namespace=$(grep "^namespace:" "${CONFIG_YAML_PATH}" 2>/dev/null | head -1 | awk '{print $2}' | tr -d "'\"")
    namespace="${namespace:-kweaver-ai}"
    
    # Uninstall in reverse order
    for ((i=${#ISF_RELEASES[@]}-1; i>=0; i--)); do
        local release_name="${ISF_RELEASES[$i]}"
        log_info "Uninstalling ${release_name}..."
        if helm uninstall "${release_name}" -n "${namespace}" 2>/dev/null; then
            log_info "✓ ${release_name} uninstalled successfully"
        else
            log_warn "⚠ ${release_name} not found or already uninstalled"
        fi
    done
    
    log_info "ISF services uninstallation completed"
}

# Show ISF services status
show_isf_status() {
    log_info "ISF services status:"
    
    # Get namespace from config.yaml
    local namespace=$(grep "^namespace:" "${CONFIG_YAML_PATH}" 2>/dev/null | head -1 | awk '{print $2}' | tr -d "'\"")
    namespace="${namespace:-kweaver-ai}"
    
    log_info "Namespace: ${namespace}"
    log_info ""
    
    # Check each release
    for release_name in "${ISF_RELEASES[@]}"; do
        if helm status "${release_name}" -n "${namespace}" >/dev/null 2>&1; then
            local status=$(helm status "${release_name}" -n "${namespace}" -o json 2>/dev/null | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
            log_info "  ✓ ${release_name}: ${status}"
        else
            log_info "  ✗ ${release_name}: not installed"
        fi
    done
}
