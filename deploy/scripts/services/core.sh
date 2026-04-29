
# Default kweaver-core namespace
CORE_NAMESPACE="${CORE_NAMESPACE:-kweaver-ai}"

# Default local charts directory
CORE_LOCAL_CHARTS_DIR="${CORE_LOCAL_CHARTS_DIR:-}"
CORE_VERSION_MANIFEST_FILE="${CORE_VERSION_MANIFEST_FILE:-}"

# Global --set values array
declare -a CORE_SET_VALUES=()

# Core SQL module directories to initialize before installing Core releases.
declare -a CORE_SQL_MODULES=(
    "studio"
    "bkn"
    "vega"
    "agentoperator"
    "dataagent"
    "decisionagent"
    "flowautomation"
    "sandbox"
)

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
            --version_file=*)
                CORE_VERSION_MANIFEST_FILE="${1#*=}"
                shift
                ;;
            --version_file)
                CORE_VERSION_MANIFEST_FILE="$2"
                shift 2
                ;;
            --force-refresh)
                FORCE_REFRESH_CHARTS="true"
                shift
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
            --set=*)
                CORE_SET_VALUES+=("${1#*=}")
                shift
                ;;
            --set)
                CORE_SET_VALUES+=("$2")
                shift 2
                ;;
            --api_server_address=*)
                API_SERVER_ADVERTISE_ADDRESS="${1#*=}"
                shift
                ;;
            --api_server_address)
                API_SERVER_ADVERTISE_ADDRESS="$2"
                shift 2
                ;;
            --access_address=*)
                KWEAVER_ACCESS_ADDRESS="${1#*=}"
                shift
                ;;
            --access_address)
                KWEAVER_ACCESS_ADDRESS="$2"
                shift 2
                ;;
            --minimum|--min)
                CORE_SET_VALUES+=("auth.enabled=false")
                CORE_SET_VALUES+=("businessDomain.enabled=false")
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

# Resolve local charts directory for kweaver-core
_core_resolve_charts_dir() {
    if [[ -n "${CORE_LOCAL_CHARTS_DIR}" ]]; then
        if [[ -d "${CORE_LOCAL_CHARTS_DIR}" ]]; then
            echo "${CORE_LOCAL_CHARTS_DIR}"
        fi
    fi
}

_core_download_charts_dir() {
    if [[ -n "${CORE_LOCAL_CHARTS_DIR}" ]]; then
        ensure_charts_dir "${CORE_LOCAL_CHARTS_DIR}"
        return 0
    fi

    ensure_charts_dir "$(resolve_shared_charts_dir)"
}

_core_auto_resolve_version_manifest() {
    if [[ -n "${CORE_VERSION_MANIFEST_FILE:-}" ]]; then
        return 0
    fi

    local embedded_manifest
    if [[ -n "${HELM_CHART_VERSION:-}" ]]; then
        embedded_manifest="$(resolve_embedded_release_manifest "kweaver-core" "${HELM_CHART_VERSION}")"
    else
        embedded_manifest="$(resolve_latest_embedded_release_manifest "kweaver-core")"
    fi
    if [[ -n "${embedded_manifest}" ]]; then
        CORE_VERSION_MANIFEST_FILE="${embedded_manifest}"
    fi
}

_core_require_version_manifest() {
    _core_auto_resolve_version_manifest

    if [[ -z "${CORE_VERSION_MANIFEST_FILE:-}" ]]; then
        log_error "No release manifest found for kweaver-core. Provide --version or --version_file."
        return 1
    fi
}

_core_resolve_release_version() {
    local release_name="$1"
    _core_require_version_manifest || return 1
    resolve_release_chart_version "${CORE_VERSION_MANIFEST_FILE:-}" "kweaver-core" "${HELM_CHART_VERSION:-}" "${release_name}" "${HELM_CHART_VERSION:-}"
}

_core_resolve_chart_name() {
    local release_name="$1"
    _core_require_version_manifest || return 1
    resolve_release_chart_name "${CORE_VERSION_MANIFEST_FILE:-}" "kweaver-core" "${HELM_CHART_VERSION:-}" "${release_name}" "${release_name}"
}

_core_release_names() {
    _core_require_version_manifest || return 1
    get_release_manifest_release_names "${CORE_VERSION_MANIFEST_FILE}" "kweaver-core" "${HELM_CHART_VERSION:-}"
}

_core_resolve_isf_dependency_version() {
    if [[ -z "${CORE_VERSION_MANIFEST_FILE:-}" ]]; then
        return 0
    fi

    get_release_manifest_dependency_version_optional "${CORE_VERSION_MANIFEST_FILE}" "isf"
}

_core_resolve_isf_dependency_manifest() {
    if [[ -z "${CORE_VERSION_MANIFEST_FILE:-}" ]]; then
        return 0
    fi

    get_release_manifest_dependency_manifest_optional "${CORE_VERSION_MANIFEST_FILE}" "isf"
}

init_core_databases() {
    local sql_base_dir
    sql_base_dir="$(resolve_versioned_sql_dir "kweaver-core" "${HELM_CHART_VERSION:-}")"

    if ! is_rds_internal; then
        warn_external_rds_sql_required "KWeaver Core" "${sql_base_dir}"
        log_warn "Skipping automatic KWeaver Core database initialization (external RDS)"
        return 0
    fi

    # Check if Core manifest has pre-stage data-migrator (0.6.0+)
    # If so, skip SQL initialization - the data-migrator chart will handle it
    _core_require_version_manifest || return 1
    if should_skip_db_init_for_manifest "${CORE_VERSION_MANIFEST_FILE}"; then
        log_info "Core manifest ${CORE_VERSION_MANIFEST_FILE} has pre-stage data-migrator (0.6.0+), skipping SQL initialization"
        return 0
    fi

    local -a sql_modules=()
    kweaver_mapfile_compat sql_modules list_versioned_sql_modules "kweaver-core" "${HELM_CHART_VERSION:-}"
    if [[ ${#sql_modules[@]} -eq 0 ]]; then
        log_info "Skipping KWeaver Core database initialization: no SQL module directories found in ${sql_base_dir}"
        return 0
    fi

    local module_name
    local sql_dir
    for module_name in "${sql_modules[@]}"; do
        sql_dir="${sql_base_dir}/${module_name}"

        if ! init_module_database_if_present "${module_name}" "${sql_dir}" "${module_name}"; then
            log_error "Failed to initialize database for module: ${module_name}"
            return 1
        fi
    done
}

download_core() {
    log_info "Downloading KWeaver Core charts..."
    ensure_helm_available
    _core_require_version_manifest || return 1

    HELM_CHART_REPO_NAME="${HELM_CHART_REPO_NAME:-kweaver}"
    HELM_CHART_REPO_URL="${HELM_CHART_REPO_URL:-https://kweaver-ai.github.io/helm-repo/}"

    local charts_dir
    charts_dir="$(_core_download_charts_dir)"

    ensure_helm_repo "${HELM_CHART_REPO_NAME}" "${HELM_CHART_REPO_URL}"

    # Check if ISF dependency is declared in manifest
    local isf_dep_version=""
    local isf_dep_manifest=""
    if [[ -n "${CORE_VERSION_MANIFEST_FILE:-}" ]]; then
        isf_dep_version="$(_core_resolve_isf_dependency_version)"
        isf_dep_manifest="$(_core_resolve_isf_dependency_manifest)"
    fi

    if [[ -n "${isf_dep_version}" ]]; then
        if is_dependency_enabled "${CORE_VERSION_MANIFEST_FILE}" "isf" "${CORE_SET_VALUES[@]}"; then
            log_info "ISF dependency found in manifest (version: ${isf_dep_version}), downloading ISF charts"
            local original_isf_charts_dir="${ISF_LOCAL_CHARTS_DIR:-}"
            local original_isf_manifest_file="${ISF_VERSION_MANIFEST_FILE:-}"
            local original_chart_version="${HELM_CHART_VERSION:-}"
            ISF_LOCAL_CHARTS_DIR="${charts_dir}"
            HELM_CHART_VERSION="${isf_dep_version}"
            ISF_VERSION_MANIFEST_FILE="${isf_dep_manifest}"

            download_isf

            ISF_LOCAL_CHARTS_DIR="${original_isf_charts_dir}"
            ISF_VERSION_MANIFEST_FILE="${original_isf_manifest_file}"
            HELM_CHART_VERSION="${original_chart_version}"
        else
            log_info "ISF in manifest but disabled for this profile (e.g. --minimum); skipping ISF chart download"
        fi
    else
        log_info "No ISF dependency declared in manifest, skipping ISF download"
    fi

    local -a release_names=()
    kweaver_mapfile_compat release_names _core_release_names
    local release_name
    local release_version
    local chart_name
    for release_name in "${release_names[@]}"; do
        release_version="$(_core_resolve_release_version "${release_name}")"
        chart_name="$(_core_resolve_chart_name "${release_name}")"
        download_chart_to_cache "${charts_dir}" "${HELM_CHART_REPO_NAME}" "${chart_name}" "${release_version}" "${FORCE_REFRESH_CHARTS:-false}"
    done
}

# Find local chart tgz for a given release name
_core_find_local_chart() {
    local charts_dir="$1"
    local chart_name="$2"
    find_cached_chart_tgz "${charts_dir}" "${chart_name}"
}

# Install a single kweaver-core release from a local .tgz
_install_core_release_local() {
    local release_name="$1"
    local charts_dir="$2"
    local namespace="$3"
    local requested_version
    local chart_name

    requested_version="$(_core_resolve_release_version "${release_name}")"
    chart_name="$(_core_resolve_chart_name "${release_name}")"

    local chart_tgz=""
    if [[ -n "${requested_version}" ]]; then
        chart_tgz="$(find_cached_chart_tgz_by_version "${charts_dir}" "${chart_name}" "${requested_version}" || true)"
    fi
    if [[ -z "${chart_tgz}" ]]; then
        chart_tgz="$(_core_find_local_chart "${charts_dir}" "${chart_name}")"
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
    if should_skip_upgrade_same_chart_version "${release_name}" "${namespace}" "${chart_name}" "${target_version}"; then
        return 0
    fi

    log_info "Installing ${release_name} from local chart: $(basename "${chart_tgz}")..."

    local -a helm_args=(
        "upgrade" "--install" "${release_name}" "${chart_tgz}"
        "--namespace" "${namespace}"
        "-f" "${CONFIG_YAML_PATH}"
        "--wait" "--timeout=600s"
    )

    # Add all --set values
    local set_value
    for set_value in "${CORE_SET_VALUES[@]}"; do
        helm_args+=("--set" "${set_value}")
    done

    if helm "${helm_args[@]}"; then
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
    local chart_name
    chart_name="$(_core_resolve_chart_name "${release_name}")"

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

    # Add all --set values
    local set_value
    for set_value in "${CORE_SET_VALUES[@]}"; do
        helm_args+=("--set" "${set_value}")
    done

    if helm "${helm_args[@]}"; then
        log_info "✓ ${release_name} installed successfully"
    else
        log_error "✗ Failed to install ${release_name}"
        return 1
    fi
}

# Inject default --set values for kweaver-core if user did not override them.
# Currently: businessDomain.enabled defaults to false at install time.
_core_apply_default_set_values() {
    if ! get_set_value "businessDomain.enabled" "${CORE_SET_VALUES[@]}" >/dev/null 2>&1; then
        CORE_SET_VALUES+=("businessDomain.enabled=false")
        log_info "Default applied: --set businessDomain.enabled=false (override with --set businessDomain.enabled=true)"
    fi
}

# Install KWeaver Core services via Helm
install_core() {
    log_info "Installing KWeaver Core services via Helm..."
    _core_require_version_manifest || return 1
    _core_apply_default_set_values

    if ! ensure_platform_prerequisites; then
        log_error "Failed to ensure platform prerequisites for KWeaver Core"
        return 1
    fi

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
        log_info "No explicit local Core charts directory provided, using Helm repo."
        log_info "  Version:   ${HELM_CHART_VERSION}"
        if [[ -n "${CORE_VERSION_MANIFEST_FILE:-}" ]]; then
            log_info "  Version Manifest: ${CORE_VERSION_MANIFEST_FILE}"
        fi
        log_info "  Helm Repo: ${HELM_CHART_REPO_NAME:-kweaver} -> ${HELM_CHART_REPO_URL:-https://kweaver-ai.github.io/helm-repo/}"
        HELM_CHART_REPO_NAME="${HELM_CHART_REPO_NAME:-kweaver}"
        HELM_CHART_REPO_URL="${HELM_CHART_REPO_URL:-https://kweaver-ai.github.io/helm-repo/}"
        ensure_helm_repo "${HELM_CHART_REPO_NAME}" "${HELM_CHART_REPO_URL}"
    fi

    log_info "Target namespace: ${namespace}"

    # Check if ISF dependency is declared in manifest and should be enabled
    local isf_dep_version=""
    local isf_dep_manifest=""
    local should_install_isf=false
    
    if [[ -n "${CORE_VERSION_MANIFEST_FILE:-}" ]]; then
        isf_dep_version="$(_core_resolve_isf_dependency_version)"
        isf_dep_manifest="$(_core_resolve_isf_dependency_manifest)"
        
        if [[ -n "${isf_dep_version}" ]]; then
            # Check if dependency should be enabled based on --set values
            if is_dependency_enabled "${CORE_VERSION_MANIFEST_FILE}" "isf" "${CORE_SET_VALUES[@]}"; then
                should_install_isf=true
                log_info "ISF dependency enabled (version: ${isf_dep_version})"
            else
                log_info "ISF dependency disabled by --set values"
            fi
        fi
    fi

    if [[ "${should_install_isf}" == "true" ]]; then
        log_info "Installing ISF services..."
        local original_isf_charts_dir="${ISF_LOCAL_CHARTS_DIR:-}"
        local original_isf_manifest_file="${ISF_VERSION_MANIFEST_FILE:-}"
        local original_chart_version="${HELM_CHART_VERSION:-}"
        if [[ "${use_local}" == "true" ]]; then
            ISF_LOCAL_CHARTS_DIR="${charts_dir}"
        fi
        HELM_CHART_VERSION="${isf_dep_version}"
        ISF_VERSION_MANIFEST_FILE="${isf_dep_manifest}"
        
        if ! install_isf; then
            ISF_LOCAL_CHARTS_DIR="${original_isf_charts_dir}"
            ISF_VERSION_MANIFEST_FILE="${original_isf_manifest_file}"
            HELM_CHART_VERSION="${original_chart_version}"
            log_error "Failed to install ISF services"
            return 1
        fi
        ISF_LOCAL_CHARTS_DIR="${original_isf_charts_dir}"
        ISF_VERSION_MANIFEST_FILE="${original_isf_manifest_file}"
        HELM_CHART_VERSION="${original_chart_version}"
    fi

    if ! init_core_databases; then
        log_error "Failed to initialize KWeaver Core databases"
        return 1
    fi

    local -a release_names=()
    kweaver_mapfile_compat release_names _core_release_names
    local release_version
    for release_name in "${release_names[@]}"; do
        release_version="$(_core_resolve_release_version "${release_name}")"
        if [[ "${use_local}" == "true" ]]; then
            _install_core_release_local "${release_name}" "${charts_dir}" "${namespace}"
        else
            _install_core_release_repo "${release_name}" "${namespace}" "${HELM_CHART_REPO_NAME}" "${release_version}"
        fi
    done

    log_info "KWeaver Core services installation completed."

    log_info "Context Loader toolset import: run ./onboard.sh after kweaver auth (uses kweaver call impex), or set IMPORT_CONTEXT_LOADER_TOOLSET=false to skip that step in onboard."

    local _host _port _scheme
    _host="$(_read_access_address_field "host" 2>/dev/null || true)"
    _port="$(_read_access_address_field "port" 2>/dev/null || true)"
    _scheme="$(_read_access_address_field "scheme" 2>/dev/null || true)"
    _host="${_host:-localhost}"
    _port="${_port:-443}"
    _scheme="${_scheme:-https}"

    echo ""
    echo "============================================"
    echo "  Verify your installation:"
    echo ""
    if [[ "${_port}" == "443" || "${_port}" == "80" ]]; then
        echo "    curl -k ${_scheme}://${_host}/api/v1/health"
    else
        echo "    curl -k ${_scheme}://${_host}:${_port}/api/v1/health"
    fi
    echo ""
    echo "============================================"
}

# Uninstall KWeaver Core services
uninstall_core() {
    log_info "Uninstalling KWeaver Core services..."

    local namespace
    namespace=$(grep "^namespace:" "${CONFIG_YAML_PATH}" 2>/dev/null | head -1 | awk '{print $2}' | tr -d "'\"")
    namespace="${namespace:-${CORE_NAMESPACE}}"

    local -a release_names=()
    kweaver_mapfile_compat release_names _core_release_names
    for ((i=${#release_names[@]}-1; i>=0; i--)); do
        local release_name="${release_names[$i]}"
        log_info "Uninstalling ${release_name}..."
        if helm uninstall "${release_name}" -n "${namespace}" 2>/dev/null; then
            log_info "✓ ${release_name} uninstalled"
        else
            log_warn "⚠ ${release_name} not found or already uninstalled"
        fi
    done

    # Clean up sandbox session pods created at runtime by sandbox-control-plane.
    # These pods are scheduled via K8s API and are not owned by any Helm release,
    # so `helm uninstall` cannot reclaim them.
    log_warn "Deleting sandbox session pods (label: sandbox-type=execution)"
    kubectl delete pod -n "${namespace}" -l sandbox-type=execution --ignore-not-found >/dev/null 2>&1 || true

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

    local -a release_names=()
    kweaver_mapfile_compat release_names _core_release_names
    for release_name in "${release_names[@]}"; do
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
