
# DecisionAgent releases list
declare -a DECISIONAGENT_RELEASES=(
    "agent-backend"
    "agent-web"
)

# Parse decisionagent command arguments
parse_decisionagent_args() {
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

# Initialize DecisionAgent database using common database initialization function
init_decisionagent_database() {
    local sql_dir="${SCRIPT_DIR}/scripts/sql/decisionagent"

    # Only initialize database if RDS is internal (MariaDB installed in cluster)
    if ! is_rds_internal; then
        warn_external_rds_sql_required "DecisionAgent" "${sql_dir}"
        log_warn "Skipping automatic DecisionAgent database initialization (external RDS)"
        return 0
    fi

    init_module_database "decisionagent" "${sql_dir}"
}

# Install DecisionAgent services via Helm
install_decisionagent() {
    log_info "Installing DecisionAgent services via Helm..."
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
    if ! init_decisionagent_database; then
        log_error "Failed to initialize DecisionAgent database"
        return 1
    fi

    log_info "Target namespace: ${namespace}"

    # Install each release
    for release_name in "${DECISIONAGENT_RELEASES[@]}"; do
        install_decisionagent_release "${release_name}" "${release_name}" "${namespace}" "${HELM_CHART_REPO_NAME}" "${HELM_CHART_VERSION}"
    done

    log_info "DecisionAgent services installation completed"
}

# Install a single DecisionAgent release
install_decisionagent_release() {
    local release_name="$1"
    local chart_name="$2"
    local namespace="$3"
    local helm_repo_name="$4"
    local release_version="$5"
    local values_file="${6:-${SCRIPT_DIR}/conf/config.yaml}"

    # Build Helm chart reference
    local chart_ref="${helm_repo_name}/${chart_name}"

    # Build Helm command args
    local -a helm_args=(
        "${release_name}"
        "${chart_ref}"
        "--namespace" "${namespace}"
        "-f" "${values_file}"
    )

    # Add version parameter only if specified
    if [[ -n "${release_version}" ]]; then
        helm_args+=("--version" "${release_version}")
    fi

    helm_args+=("--devel")

    helm_install_with_retry "${release_name}" "${helm_args[@]}"
}

# Uninstall DecisionAgent services
uninstall_decisionagent() {
    log_info "Uninstalling DecisionAgent services..."

    # Get namespace from config.yaml
    local namespace=$(grep "^namespace:" "${CONFIG_YAML_PATH}" 2>/dev/null | head -1 | awk '{print $2}' | tr -d "'\"")
    namespace="${namespace:-kweaver-ai}"

    # Uninstall in reverse order (with timeout protection per release)
    for ((i=${#DECISIONAGENT_RELEASES[@]}-1; i>=0; i--)); do
        local release_name="${DECISIONAGENT_RELEASES[$i]}"
        helm_uninstall_safe "${release_name}" "${namespace}"
    done

    log_info "DecisionAgent services uninstallation completed"
}

# Show DecisionAgent services status
show_decisionagent_status() {
    log_info "DecisionAgent services status:"

    # Get namespace from config.yaml
    local namespace=$(grep "^namespace:" "${CONFIG_YAML_PATH}" 2>/dev/null | head -1 | awk '{print $2}' | tr -d "'\"")
    namespace="${namespace:-kweaver-ai}"

    log_info "Namespace: ${namespace}"
    log_info ""

    # Check each release
    for release_name in "${DECISIONAGENT_RELEASES[@]}"; do
        if helm status "${release_name}" -n "${namespace}" >/dev/null 2>&1; then
            local status=$(helm status "${release_name}" -n "${namespace}" -o json 2>/dev/null | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
            log_info "  ✓ ${release_name}: ${status}"
        else
            log_info "  ✗ ${release_name}: not installed"
        fi
    done
}
