
ISF_LOCAL_CHARTS_DIR="${ISF_LOCAL_CHARTS_DIR:-}"
ISF_VERSION_MANIFEST_FILE="${ISF_VERSION_MANIFEST_FILE:-}"

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
            --charts_dir=*)
                ISF_LOCAL_CHARTS_DIR="${1#*=}"
                shift
                ;;
            --charts_dir)
                ISF_LOCAL_CHARTS_DIR="$2"
                shift 2
                ;;
            --version_file=*)
                ISF_VERSION_MANIFEST_FILE="${1#*=}"
                shift
                ;;
            --version_file)
                ISF_VERSION_MANIFEST_FILE="$2"
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
            --force-refresh)
                FORCE_REFRESH_CHARTS="true"
                shift
                ;;
            -y|--yes)
                ASSUME_YES="true"
                shift
                ;;
            --force-upgrade)
                FORCE_UPGRADE="true"
                shift
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
    local sql_dir
    sql_dir="$(resolve_versioned_sql_dir "isf" "${HELM_CHART_VERSION:-}")"

    # Only initialize database if RDS is internal (MariaDB installed in cluster)
    if ! is_rds_internal; then
        warn_external_rds_sql_required "ISF" "${sql_dir}"
        log_warn "Skipping automatic ISF database initialization (external RDS)"
        return 0
    fi

    # Check if ISF manifest has pre-stage data-migrator (0.6.0+)
    # If so, skip SQL initialization - the data-migrator chart will handle it
    _isf_require_version_manifest || return 1
    if should_skip_db_init_for_manifest "${ISF_VERSION_MANIFEST_FILE}"; then
        log_info "ISF manifest ${ISF_VERSION_MANIFEST_FILE} has pre-stage data-migrator (0.6.0+), skipping SQL initialization"
        return 0
    fi

    init_module_database_if_present "isf" "${sql_dir}" "ISF"
}

_isf_resolve_charts_dir() {
    if [[ -n "${ISF_LOCAL_CHARTS_DIR}" ]]; then
        if [[ -d "${ISF_LOCAL_CHARTS_DIR}" ]]; then
            echo "${ISF_LOCAL_CHARTS_DIR}"
        fi
    fi
}

_isf_download_charts_dir() {
    if [[ -n "${ISF_LOCAL_CHARTS_DIR}" ]]; then
        ensure_charts_dir "${ISF_LOCAL_CHARTS_DIR}"
        return 0
    fi

    ensure_charts_dir "$(resolve_shared_charts_dir)"
}

_isf_auto_resolve_version_manifest() {
    if [[ -n "${ISF_VERSION_MANIFEST_FILE:-}" ]]; then
        return 0
    fi

    local embedded_manifest
    if [[ -n "${HELM_CHART_VERSION:-}" ]]; then
        embedded_manifest="$(resolve_embedded_release_manifest "isf" "${HELM_CHART_VERSION}")"
    else
        embedded_manifest="$(resolve_latest_embedded_release_manifest "isf")"
    fi
    if [[ -n "${embedded_manifest}" ]]; then
        ISF_VERSION_MANIFEST_FILE="${embedded_manifest}"
    fi
}

_isf_require_version_manifest() {
    _isf_auto_resolve_version_manifest

    if [[ -z "${ISF_VERSION_MANIFEST_FILE:-}" ]]; then
        log_error "No release manifest found for isf. Provide --version or --version_file."
        return 1
    fi
}

_isf_resolve_release_version() {
    local release_name="$1"
    _isf_require_version_manifest || return 1
    resolve_release_chart_version "${ISF_VERSION_MANIFEST_FILE:-}" "isf" "${HELM_CHART_VERSION:-}" "${release_name}" "${HELM_CHART_VERSION:-}"
}

_isf_resolve_chart_name() {
    local release_name="$1"
    _isf_require_version_manifest || return 1
    resolve_release_chart_name "${ISF_VERSION_MANIFEST_FILE:-}" "isf" "${HELM_CHART_VERSION:-}" "${release_name}" "${release_name}"
}

_isf_release_names() {
    _isf_require_version_manifest || return 1
    get_release_manifest_release_names "${ISF_VERSION_MANIFEST_FILE}" "isf" "${HELM_CHART_VERSION:-}"
}

# Install ISF services via Helm
install_isf() {
    log_info "Installing ISF services via Helm..."
    _isf_require_version_manifest || return 1
    log_info "  Version: ${HELM_CHART_VERSION}"
    if [[ -n "${ISF_VERSION_MANIFEST_FILE:-}" ]]; then
        log_info "  Version Manifest: ${ISF_VERSION_MANIFEST_FILE}"
    fi
    log_info "  Helm Repo: ${HELM_CHART_REPO_NAME:-kweaver} -> ${HELM_CHART_REPO_URL:-https://kweaver-ai.github.io/helm-repo/}"

    if ! ensure_platform_prerequisites; then
        log_error "Failed to ensure platform prerequisites for ISF"
        return 1
    fi

    # Get namespace from config.yaml
    local namespace=$(grep "^namespace:" "${CONFIG_YAML_PATH}" 2>/dev/null | head -1 | awk '{print $2}' | tr -d "'\"")
    namespace="${namespace:-kweaver-ai}"
    
    # Create namespace if not exists
    kubectl create namespace "${namespace}" 2>/dev/null || true
    
    local charts_dir
    charts_dir="$(_isf_resolve_charts_dir)"

    local use_local=false
    if [[ -n "${charts_dir}" && -d "${charts_dir}" ]]; then
        use_local=true
        log_info "Using local ISF charts from: ${charts_dir}"
    else
        log_info "No explicit local ISF charts directory provided, using Helm repo."
        log_info "Adding Helm repo: ${HELM_CHART_REPO_NAME} -> ${HELM_CHART_REPO_URL}"
        ensure_helm_repo "${HELM_CHART_REPO_NAME}" "${HELM_CHART_REPO_URL}"
    fi
    
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
    local -a release_names=()
    mapfile -t release_names < <(_isf_release_names)
    local release_name
    local release_version
    local chart_name
    for release_name in "${release_names[@]}"; do
        release_version="$(_isf_resolve_release_version "${release_name}")"
        chart_name="$(_isf_resolve_chart_name "${release_name}")"
        if [[ "${use_local}" == "true" ]]; then
            if ! _install_isf_release_local "${release_name}" "${charts_dir}" "${namespace}" "${temp_config}"; then
                install_failed=1
                break
            fi
        elif ! install_isf_release "${release_name}" "${chart_name}" "${namespace}" "${HELM_CHART_REPO_NAME}" "${release_version}" "${temp_config}"; then
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

_install_isf_release_local() {
    local release_name="$1"
    local charts_dir="$2"
    local namespace="$3"
    local values_file="$4"
    local requested_version
    local chart_name

    requested_version="$(_isf_resolve_release_version "${release_name}")"
    chart_name="$(_isf_resolve_chart_name "${release_name}")"

    local chart_tgz=""
    if [[ -n "${requested_version}" ]]; then
        chart_tgz="$(find_cached_chart_tgz_by_version "${charts_dir}" "${chart_name}" "${requested_version}" || true)"
    fi
    if [[ -z "${chart_tgz}" ]]; then
        chart_tgz="$(find_cached_chart_tgz "${charts_dir}" "${chart_name}")"
    fi
    if [[ -z "${chart_tgz}" ]]; then
        log_error "✗ Local chart not found for ${release_name} (${chart_name}) in ${charts_dir}"
        return 1
    fi

    local target_version
    target_version="${requested_version}"
    if [[ -z "${target_version}" ]]; then
        target_version="$(get_local_chart_version "${chart_tgz}")"
    fi
    if [[ -z "${target_version}" ]]; then
        target_version="$(get_chart_version_from_filename "${chart_tgz}" "${chart_name}")"
    fi

    if should_skip_upgrade_same_chart_version "${release_name}" "${namespace}" "${chart_name}" "${target_version}"; then
        return 0
    fi

    log_info "Installing ${release_name} from local chart: $(basename "${chart_tgz}")..."
    helm upgrade --install "${release_name}" "${chart_tgz}" \
        --namespace "${namespace}" \
        -f "${values_file}" \
        --devel --wait --timeout=600s
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

    local target_version="${release_version}"
    if [[ -z "${target_version}" ]]; then
        target_version=$(get_repo_chart_latest_version "${helm_repo_name}" "${chart_name}")
    fi

    if should_skip_upgrade_same_chart_version "${release_name}" "${namespace}" "${chart_name}" "${target_version}"; then
        return 0
    fi
    
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

download_isf() {
    log_info "Downloading ISF charts..."
    ensure_helm_available
    _isf_require_version_manifest || return 1

    HELM_CHART_REPO_NAME="${HELM_CHART_REPO_NAME:-kweaver}"
    HELM_CHART_REPO_URL="${HELM_CHART_REPO_URL:-https://kweaver-ai.github.io/helm-repo/}"

    local charts_dir
    charts_dir="$(_isf_download_charts_dir)"

    ensure_helm_repo "${HELM_CHART_REPO_NAME}" "${HELM_CHART_REPO_URL}"

    local -a release_names=()
    mapfile -t release_names < <(_isf_release_names)
    local release_name
    local release_version
    local chart_name
    for release_name in "${release_names[@]}"; do
        release_version="$(_isf_resolve_release_version "${release_name}")"
        chart_name="$(_isf_resolve_chart_name "${release_name}")"
        download_chart_to_cache "${charts_dir}" "${HELM_CHART_REPO_NAME}" "${chart_name}" "${release_version}" "${FORCE_REFRESH_CHARTS:-false}"
    done
}

# Uninstall ISF services
uninstall_isf() {
    log_info "Uninstalling ISF services..."
    _isf_require_version_manifest || return 1
    
    # Get namespace from config.yaml
    local namespace=$(grep "^namespace:" "${CONFIG_YAML_PATH}" 2>/dev/null | head -1 | awk '{print $2}' | tr -d "'\"")
    namespace="${namespace:-kweaver-ai}"
    
    # Uninstall in reverse order
    local -a release_names=()
    mapfile -t release_names < <(_isf_release_names)
    for ((i=${#release_names[@]}-1; i>=0; i--)); do
        local release_name="${release_names[$i]}"
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
    _isf_require_version_manifest || return 1
    
    # Get namespace from config.yaml
    local namespace=$(grep "^namespace:" "${CONFIG_YAML_PATH}" 2>/dev/null | head -1 | awk '{print $2}' | tr -d "'\"")
    namespace="${namespace:-kweaver-ai}"
    
    log_info "Namespace: ${namespace}"
    log_info ""
    
    # Check each release
    local -a release_names=()
    mapfile -t release_names < <(_isf_release_names)
    local release_name
    for release_name in "${release_names[@]}"; do
        if helm status "${release_name}" -n "${namespace}" >/dev/null 2>&1; then
            local status=$(helm status "${release_name}" -n "${namespace}" -o json 2>/dev/null | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
            log_info "  ✓ ${release_name}: ${status}"
        else
            log_info "  ✗ ${release_name}: not installed"
        fi
    done
}
