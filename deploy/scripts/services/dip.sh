
# Default DIP namespace
DIP_NAMESPACE="${DIP_NAMESPACE:-kweaver-ai}"

# Default local DIP charts directory (relative to deploy root)
DIP_LOCAL_CHARTS_DIR="${DIP_LOCAL_CHARTS_DIR:-}"
DIP_VERSION_MANIFEST_FILE="${DIP_VERSION_MANIFEST_FILE:-}"
DIP_CONFIRM_MISSING_OPENCLAW_PATHS="${DIP_CONFIRM_MISSING_OPENCLAW_PATHS:-false}"

# --set values forwarded to DIP helm releases (and used to seed core defaults
# when dip auto-installs missing core releases).
declare -a DIP_SET_VALUES=()

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
            --version_file=*)
                DIP_VERSION_MANIFEST_FILE="${1#*=}"
                shift
                ;;
            --version_file)
                DIP_VERSION_MANIFEST_FILE="$2"
                shift 2
                ;;
            --force-refresh)
                FORCE_REFRESH_CHARTS="true"
                shift
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
            --confirm-missing-openclaw-paths)
                DIP_CONFIRM_MISSING_OPENCLAW_PATHS="true"
                shift
                ;;
            --set=*)
                DIP_SET_VALUES+=("${1#*=}")
                shift
                ;;
            --set)
                DIP_SET_VALUES+=("$2")
                shift 2
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

# Check if a single Helm release is deployed in the given namespace
_dip_helm_release_exists() {
    local release="$1"
    local ns="$2"
    helm list -n "${ns}" --short 2>/dev/null | grep -q "^${release}$"
}

# Ensure kweaver-core modules are installed.
_dip_ensure_kweaver_core() {
    _dip_require_version_manifest || return 1

    local namespace
    namespace=$(grep "^namespace:" "${CONFIG_YAML_PATH}" 2>/dev/null | head -1 | awk '{print $2}' | tr -d "'\"")
    namespace="${namespace:-kweaver-ai}"

    log_info "Checking kweaver-core dependencies for KWeaver DIP..."

    local missing_isf=false
    local missing_core=false
    local release_name
    local original_chart_version="${HELM_CHART_VERSION:-}"
    local original_core_manifest="${CORE_VERSION_MANIFEST_FILE:-}"
    local original_isf_manifest="${ISF_VERSION_MANIFEST_FILE:-}"
    local core_dependency_manifest=""
    local core_dependency_version="${HELM_CHART_VERSION:-}"
    local isf_dependency_manifest=""
    local isf_dependency_version="${HELM_CHART_VERSION:-}"

    core_dependency_manifest="$(_dip_resolve_core_dependency_manifest)"
    core_dependency_version="$(_dip_resolve_core_dependency_version)"
    isf_dependency_manifest="$(_dip_resolve_isf_dependency_manifest)"
    isf_dependency_version="$(_dip_resolve_isf_dependency_version)"

    local -a isf_release_names=()
    mapfile -t isf_release_names < <(_dip_list_manifest_release_names "isf" "${isf_dependency_manifest}" "${isf_dependency_version}")
    local -a core_release_names=()
    mapfile -t core_release_names < <(_dip_list_manifest_release_names "kweaver-core" "${core_dependency_manifest}" "${core_dependency_version}")

    for release_name in "${isf_release_names[@]}"; do
        if _dip_helm_release_exists "${release_name}" "${namespace}"; then
            log_info "  ✓ ISF release already installed (${release_name})"
        else
            log_info "  ✗ ISF release not installed (${release_name})"
            missing_isf=true
        fi
    done

    for release_name in "${core_release_names[@]}"; do
        if _dip_helm_release_exists "${release_name}" "${namespace}"; then
            log_info "  ✓ Core release already installed (${release_name})"
        else
            log_info "  ✗ Core release not installed (${release_name})"
            missing_core=true
        fi
    done

    if [[ "${missing_isf}" == "true" ]]; then
        if [[ -n "${isf_dependency_manifest}" ]]; then
            ISF_VERSION_MANIFEST_FILE="${isf_dependency_manifest}"
            HELM_CHART_VERSION="${isf_dependency_version}"
        fi
        if ! install_isf; then
            HELM_CHART_VERSION="${original_chart_version}"
            CORE_VERSION_MANIFEST_FILE="${original_core_manifest}"
            ISF_VERSION_MANIFEST_FILE="${original_isf_manifest}"
            log_error "Failed to install kweaver-core module: ISF"
            return 1
        fi
        HELM_CHART_VERSION="${original_chart_version}"
        CORE_VERSION_MANIFEST_FILE="${original_core_manifest}"
        ISF_VERSION_MANIFEST_FILE="${original_isf_manifest}"
    fi

    if [[ "${missing_core}" == "true" ]]; then
        log_info "Installing missing KWeaver Core releases..."
        if [[ -n "${core_dependency_manifest}" ]]; then
            CORE_VERSION_MANIFEST_FILE="${core_dependency_manifest}"
            HELM_CHART_VERSION="${core_dependency_version}"
        fi
        if ! install_core; then
            HELM_CHART_VERSION="${original_chart_version}"
            CORE_VERSION_MANIFEST_FILE="${original_core_manifest}"
            ISF_VERSION_MANIFEST_FILE="${original_isf_manifest}"
            log_error "Failed to install missing KWeaver Core releases"
            return 1
        fi
        HELM_CHART_VERSION="${original_chart_version}"
        CORE_VERSION_MANIFEST_FILE="${original_core_manifest}"
        ISF_VERSION_MANIFEST_FILE="${original_isf_manifest}"
    fi

    if [[ "${missing_isf}" == "false" && "${missing_core}" == "false" ]]; then
        log_info "All kweaver-core dependencies are satisfied."
    else
        log_info "kweaver-core dependency installation completed."
    fi
}

# Resolve the local charts directory for install-time local chart usage.
# Only an explicit --charts_dir opt-in enables local chart installs.
_dip_resolve_charts_dir() {
    if [[ -n "${DIP_LOCAL_CHARTS_DIR}" ]]; then
        if [[ -d "${DIP_LOCAL_CHARTS_DIR}" ]]; then
            echo "${DIP_LOCAL_CHARTS_DIR}"
        fi
    fi
}

_dip_download_charts_dir() {
    if [[ -n "${DIP_LOCAL_CHARTS_DIR}" ]]; then
        ensure_charts_dir "${DIP_LOCAL_CHARTS_DIR}"
        return 0
    fi

    ensure_charts_dir "$(resolve_shared_charts_dir)"
}

_dip_auto_resolve_version_manifest() {
    if [[ -n "${DIP_VERSION_MANIFEST_FILE:-}" ]]; then
        return 0
    fi

    local embedded_manifest
    if [[ -n "${HELM_CHART_VERSION:-}" ]]; then
        embedded_manifest="$(resolve_embedded_release_manifest "kweaver-dip" "${HELM_CHART_VERSION}")"
    else
        embedded_manifest="$(resolve_latest_embedded_release_manifest "kweaver-dip")"
    fi
    if [[ -n "${embedded_manifest}" ]]; then
        DIP_VERSION_MANIFEST_FILE="${embedded_manifest}"
    fi
}

_dip_require_version_manifest() {
    _dip_auto_resolve_version_manifest

    if [[ -z "${DIP_VERSION_MANIFEST_FILE:-}" ]]; then
        log_error "No release manifest found for kweaver-dip. Provide --version or --version_file."
        return 1
    fi
}

_dip_resolve_release_version() {
    local release_name="$1"
    _dip_require_version_manifest || return 1
    resolve_release_chart_version "${DIP_VERSION_MANIFEST_FILE:-}" "kweaver-dip" "${HELM_CHART_VERSION:-}" "${release_name}" "${HELM_CHART_VERSION:-}"
}

_dip_resolve_chart_name() {
    local release_name="$1"
    _dip_require_version_manifest || return 1
    resolve_release_chart_name "${DIP_VERSION_MANIFEST_FILE:-}" "kweaver-dip" "${HELM_CHART_VERSION:-}" "${release_name}" "${release_name}"
}

_dip_resolve_release_stage() {
    local release_name="$1"
    _dip_require_version_manifest || return 1
    get_release_manifest_release_stage "${DIP_VERSION_MANIFEST_FILE}" "kweaver-dip" "${HELM_CHART_VERSION:-}" "${release_name}"
}

_dip_release_names() {
    _dip_require_version_manifest || return 1

    local -a manifest_release_names=()
    mapfile -t manifest_release_names < <(get_release_manifest_release_names "${DIP_VERSION_MANIFEST_FILE}" "kweaver-dip" "${HELM_CHART_VERSION:-}")

    local stage
    local release_name
    local release_stage
    for stage in pre main post; do
        for release_name in "${manifest_release_names[@]}"; do
            release_stage="$(_dip_resolve_release_stage "${release_name}")" || return 1
            if [[ "${release_stage}" == "${stage}" ]]; then
                printf '%s\n' "${release_name}"
            fi
        done
    done
}

_dip_release_names_reverse() {
    local -a ordered_release_names=()
    mapfile -t ordered_release_names < <(_dip_release_names)

    local i
    for ((i=${#ordered_release_names[@]}-1; i>=0; i--)); do
        printf '%s\n' "${ordered_release_names[$i]}"
    done
}

_dip_resolve_core_dependency_version() {
    _dip_require_version_manifest || return 1
    get_release_manifest_dependency_version "${DIP_VERSION_MANIFEST_FILE}" "kweaver-core"
}

_dip_resolve_core_dependency_manifest() {
    _dip_require_version_manifest || return 1
    get_release_manifest_dependency_manifest "${DIP_VERSION_MANIFEST_FILE}" "kweaver-core"
}

_dip_has_direct_dependency() {
    local dependency_product="$1"

    _dip_require_version_manifest || return 1

    [[ -n "$(_manifest_strip_quotes "$(_manifest_read_dependency_field "${DIP_VERSION_MANIFEST_FILE}" "${dependency_product}" "version")")" ]]
}

_dip_resolve_isf_dependency_version() {
    _dip_require_version_manifest || return 1

    if _dip_has_direct_dependency "isf"; then
        get_release_manifest_dependency_version "${DIP_VERSION_MANIFEST_FILE}" "isf"
        return 0
    fi

    local core_manifest
    core_manifest="$(_dip_resolve_core_dependency_manifest)"
    if [[ -n "${core_manifest}" ]]; then
        get_release_manifest_dependency_version "${core_manifest}" "isf"
        return 0
    fi

    echo "${HELM_CHART_VERSION:-}"
}

_dip_resolve_isf_dependency_manifest() {
    _dip_require_version_manifest || return 1

    if _dip_has_direct_dependency "isf"; then
        get_release_manifest_dependency_manifest "${DIP_VERSION_MANIFEST_FILE}" "isf"
        return 0
    fi

    local core_manifest
    core_manifest="$(_dip_resolve_core_dependency_manifest)"
    if [[ -n "${core_manifest}" ]]; then
        get_release_manifest_dependency_manifest "${core_manifest}" "isf"
    fi
}

_dip_list_manifest_release_names() {
    local product="$1"
    local manifest_file="${2:-}"
    local aggregate_version="${3:-}"

    if [[ -z "${manifest_file}" ]]; then
        if [[ -n "${aggregate_version}" ]]; then
            manifest_file="$(resolve_embedded_release_manifest "${product}" "${aggregate_version}")"
        else
            manifest_file="$(resolve_latest_embedded_release_manifest "${product}")"
        fi
    fi

    if [[ -z "${manifest_file}" ]]; then
        log_error "No release manifest found for dependency product: ${product}"
        return 1
    fi

    get_release_manifest_release_names "${manifest_file}" "${product}" "${aggregate_version}"
}

# Find a local tgz for a chart name inside a directory (picks the first match)
_dip_find_local_chart() {
    local charts_dir="$1"
    local chart_name="$2"
    find_cached_chart_tgz "${charts_dir}" "${chart_name}"
}

_dip_show_access_hints() {
    local base_url
    base_url="$(get_access_address_base_url)"
    if [[ -z "${base_url}" ]]; then
        return 0
    fi

    log_info "Access KWeaver deploy console: ${base_url}/deploy"
    log_info "Access KWeaver studio: ${base_url}/studio"
    log_info "Access KWeaver dip-hub: ${base_url}/dip-hub"
}

_dip_confirm_missing_openclaw_paths() {
    local -a missing_messages=("$@")

    if [[ "${DIP_CONFIRM_MISSING_OPENCLAW_PATHS}" == "true" ]]; then
        local message
        for message in "${missing_messages[@]}"; do
            log_warn "${message}"
        done
        log_warn "Continuing because --confirm-missing-openclaw-paths was provided."
        return 0
    fi

    local message
    for message in "${missing_messages[@]}"; do
        log_warn "${message}"
    done

    if [[ "${ASSUME_YES}" == "true" ]]; then
        log_info "Continuing with configured paths (-y). Ensure they will be created before dip-studio starts."
        return 0
    fi

    echo ""
    echo "Options:"
    echo "  1) Re-enter the paths"
    echo "  2) Continue with these paths anyway"
    echo "  3) Cancel installation"
    echo ""
    
    local choice
    read -r -p "Choose an option [1/2/3]: " choice </dev/tty || {
        log_error "Failed to read input. Rerun with --confirm-missing-openclaw-paths to skip this check."
        return 1
    }

    case "${choice}" in
        1)
            # Return special code to indicate re-entry is needed
            return 2
            ;;
        2)
            log_info "Continuing with configured paths. Ensure they will be created before dip-studio starts."
            return 0
            ;;
        3|*)
            log_error "Installation cancelled. You can update dipStudio.openClaw paths in ${CONFIG_YAML_PATH} or rerun with --confirm-missing-openclaw-paths."
            return 1
            ;;
    esac
}

_dip_prompt_openclaw_config() {
    local host_path="$1"
    local gateway_token="$2"
    local gateway_host="$3"
    local gateway_port="$4"

    echo ""
    log_warn "studio.openclaw configuration is missing or incomplete in ${CONFIG_YAML_PATH}"
    log_info "dip-studio requires OpenClaw configuration:"
    log_info "  - hostPath: Directory for OpenClaw data (config and workspace)"
    log_info "  - gatewayToken: Authentication token for OpenClaw gateway"
    log_info "  - gatewayHost: OpenClaw gateway host address"
    log_info "  - gatewayPort: OpenClaw gateway port (default: 18789)"
    echo ""

    # Get default gateway host (same as accessAddress)
    local default_gateway_host
    default_gateway_host="$(get_access_address_field "host")"
    
    local input_host_path=""
    local input_gateway_token=""
    local input_gateway_host=""
    local input_gateway_port=""

    if [[ "${ASSUME_YES}" == "true" ]]; then
        host_path="${host_path:-${HOME}}"
        gateway_host="${gateway_host:-${default_gateway_host}}"
        gateway_port="${gateway_port:-18789}"
        log_info "Using defaults (-y): hostPath=${host_path}, gatewayHost=${gateway_host}, gatewayPort=${gateway_port}"
        if [[ -z "${gateway_token}" ]]; then
            log_error "gatewayToken is required. Pre-configure studio.openclaw.gatewayToken in ${CONFIG_YAML_PATH}, or drop -y and run interactively."
            return 1
        fi
    else
        if [[ -z "${host_path}" ]]; then
            read -r -p "Enter hostPath: " input_host_path </dev/tty || {
                log_error "Failed to read input. Please run in interactive mode or pre-configure studio.openclaw in ${CONFIG_YAML_PATH}"
                return 1
            }
            host_path="${input_host_path}"
        fi

        if [[ -z "${gateway_token}" ]]; then
            read -r -p "Enter gatewayToken: " input_gateway_token </dev/tty || {
                log_error "Failed to read input. Please run in interactive mode or pre-configure studio.openclaw in ${CONFIG_YAML_PATH}"
                return 1
            }
            gateway_token="${input_gateway_token}"
        fi

        if [[ -z "${gateway_host}" ]]; then
            read -r -p "Enter gatewayHost [${default_gateway_host}]: " input_gateway_host </dev/tty || {
                log_error "Failed to read input. Please run in interactive mode or pre-configure studio.openclaw in ${CONFIG_YAML_PATH}"
                return 1
            }
            gateway_host="${input_gateway_host:-${default_gateway_host}}"
        fi

        if [[ -z "${gateway_port}" ]]; then
            read -r -p "Enter gatewayPort [18789]: " input_gateway_port </dev/tty || {
                log_error "Failed to read input. Please run in interactive mode or pre-configure studio.openclaw in ${CONFIG_YAML_PATH}"
                return 1
            }
            gateway_port="${input_gateway_port:-18789}"
        fi
    fi

    # Validate that required fields are not empty
    if [[ -z "${host_path}" || -z "${gateway_token}" || -z "${gateway_host}" || -z "${gateway_port}" ]]; then
        log_error "OpenClaw configuration cannot have empty fields. Please provide all required values."
        return 1
    fi

    echo ""
    log_info "Will use:"
    log_info "  hostPath: ${host_path}"
    log_info "  gatewayToken: ${gateway_token}"
    log_info "  gatewayHost: ${gateway_host}"
    log_info "  gatewayPort: ${gateway_port}"
    echo ""

    # Return the values via global variables
    DIP_OPENCLAW_HOST_PATH="${host_path}"
    DIP_OPENCLAW_GATEWAY_TOKEN="${gateway_token}"
    DIP_OPENCLAW_GATEWAY_HOST="${gateway_host}"
    DIP_OPENCLAW_GATEWAY_PORT="${gateway_port}"
}

_dip_update_config_with_openclaw_config() {
    local host_path="$1"
    local gateway_token="$2"
    local gateway_host="$3"
    local gateway_port="$4"

    if [[ ! -f "${CONFIG_YAML_PATH}" ]]; then
        log_error "Config file does not exist: ${CONFIG_YAML_PATH}"
        return 1
    fi

    local tmp_file
    tmp_file="$(mktemp)"

    # If all fields are empty, remove the studio.openclaw section
    if [[ -z "${host_path}" && -z "${gateway_token}" && -z "${gateway_host}" && -z "${gateway_port}" ]]; then
        awk '
            BEGIN {
                in_studio=0
                in_openclaw=0
            }
            {
                if ($1=="studio:") {
                    in_studio=1
                    next
                }
                if (in_studio && $1=="openclaw:") {
                    in_openclaw=1
                    next
                }
                if (in_studio && in_openclaw) {
                    if ($0 ~ /^  [^ ]/) {
                        in_openclaw=0
                        in_studio=0
                        print $0
                    } else if ($0 ~ /^[^ ]/) {
                        in_studio=0
                        in_openclaw=0
                        print $0
                    }
                    next
                }
                if (in_studio && $0 ~ /^[^ ]/) {
                    in_studio=0
                }
                if (!in_studio) {
                    print $0
                }
            }
        ' "${CONFIG_YAML_PATH}" > "${tmp_file}"
    else
        # Use awk to update or insert the studio.openclaw section
        awk -v hp="${host_path}" -v gw_token="${gateway_token}" \
            -v gw_host="${gateway_host}" -v gw_port="${gateway_port}" '
            BEGIN {
                in_studio=0
                in_openclaw=0
                studio_found=0
                openclaw_updated=0
            }
            {
                if ($1=="studio:") {
                    print $0
                    in_studio=1
                    studio_found=1
                    next
                }
                if (in_studio && $1=="openclaw:") {
                    print "  openclaw:"
                    print "    hostPath: " hp
                    print "    gatewayToken: " gw_token
                    print "    gatewayHost: " gw_host
                    print "    gatewayPort: " gw_port
                    in_openclaw=1
                    openclaw_updated=1
                    next
                }
                if (in_studio && in_openclaw) {
                    if ($0 ~ /^  [^ ]/) {
                        in_openclaw=0
                        print $0
                    } else if ($0 ~ /^[^ ]/) {
                        in_studio=0
                        in_openclaw=0
                        print $0
                    }
                    next
                }
                if (in_studio && $0 ~ /^[^ ]/) {
                    if (openclaw_updated==0) {
                        print "  openclaw:"
                        print "    hostPath: " hp
                        print "    gatewayToken: " gw_token
                        print "    gatewayHost: " gw_host
                        print "    gatewayPort: " gw_port
                        openclaw_updated=1
                    }
                    in_studio=0
                    print $0
                    next
                }
                print $0
            }
            END {
                if (studio_found==0) {
                    print "studio:"
                    print "  openclaw:"
                    print "    hostPath: " hp
                    print "    gatewayToken: " gw_token
                    print "    gatewayHost: " gw_host
                    print "    gatewayPort: " gw_port
                } else if (in_studio==1 && openclaw_updated==0) {
                    print "  openclaw:"
                    print "    hostPath: " hp
                    print "    gatewayToken: " gw_token
                    print "    gatewayHost: " gw_host
                    print "    gatewayPort: " gw_port
                }
            }
        ' "${CONFIG_YAML_PATH}" > "${tmp_file}"
    fi

    mv "${tmp_file}" "${CONFIG_YAML_PATH}"
    if [[ -z "${host_path}" && -z "${gateway_token}" && -z "${gateway_host}" && -z "${gateway_port}" ]]; then
        log_info "Removed studio.openclaw configuration from ${CONFIG_YAML_PATH}"
    else
        log_info "Updated ${CONFIG_YAML_PATH} with studio.openclaw configuration"
    fi
}

_dip_append_release_extra_helm_args() {
    local release_name="$1"
    local target_array_name="$2"

    if [[ "${release_name}" != "dip-studio" ]]; then
        return 0
    fi

    # Check if config file exists
    if [[ ! -f "${CONFIG_YAML_PATH}" ]]; then
        log_error "Configuration file not found: ${CONFIG_YAML_PATH}"
        log_error "Please run 'config generate' first or ensure the config file exists."
        return 1
    fi

    local host_path
    local gateway_token
    local gateway_host
    local gateway_port
    local need_reprompt=true

    # Loop to allow re-entering configuration if validation fails
    while [[ "${need_reprompt}" == "true" ]]; do
        host_path="$(get_dip_studio_openclaw_field "hostPath")"
        gateway_token="$(get_dip_studio_openclaw_field "gatewayToken")"
        gateway_host="$(get_dip_studio_openclaw_field "gatewayHost")"
        gateway_port="$(get_dip_studio_openclaw_field "gatewayPort")"

        # If any required field is missing, prompt for input
        if [[ -z "${host_path}" || -z "${gateway_token}" || -z "${gateway_host}" || -z "${gateway_port}" ]]; then
            _dip_prompt_openclaw_config "${host_path}" "${gateway_token}" "${gateway_host}" "${gateway_port}" || return 1
            host_path="${DIP_OPENCLAW_HOST_PATH}"
            gateway_token="${DIP_OPENCLAW_GATEWAY_TOKEN}"
            gateway_host="${DIP_OPENCLAW_GATEWAY_HOST}"
            gateway_port="${DIP_OPENCLAW_GATEWAY_PORT}"

            # Update config file with the new configuration
            _dip_update_config_with_openclaw_config "${host_path}" "${gateway_token}" "${gateway_host}" "${gateway_port}" || return 1
        fi

        # Validate that all fields are not empty strings after prompting
        if [[ -z "${host_path}" || -z "${gateway_token}" || -z "${gateway_host}" || -z "${gateway_port}" ]]; then
            log_error "OpenClaw configuration fields cannot be empty."
            return 1
        fi

        # Check if path exists on disk
        local -a missing_messages=()
        if [[ ! -e "${host_path}" ]]; then
            missing_messages+=("Configured OpenClaw host path does not exist: ${host_path}")
        fi

        if [[ ${#missing_messages[@]} -gt 0 ]]; then
            _dip_confirm_missing_openclaw_paths "${missing_messages[@]}"
            local confirm_result=$?
            if [[ ${confirm_result} -eq 2 ]]; then
                # User chose to re-enter configuration, clear the config and loop again
                log_info "Re-entering OpenClaw configuration..."
                _dip_update_config_with_openclaw_config "" "" "" "" || return 1
                continue
            elif [[ ${confirm_result} -ne 0 ]]; then
                # User cancelled or error occurred
                return 1
            fi
        fi

        # Paths are valid or user confirmed to continue
        need_reprompt=false
    done

    # Use eval to append to the array (Bash 4.2 compatible)
    eval "${target_array_name}+=(
        \"--set-string\" \"studio.openclaw.hostPath=\${host_path}\"
        \"--set-string\" \"studio.openclaw.gatewayToken=\${gateway_token}\"
        \"--set-string\" \"studio.openclaw.gatewayHost=\${gateway_host}\"
        \"--set-string\" \"studio.openclaw.gatewayPort=\${gateway_port}\"
    )"
}

init_dip_database() {
    local sql_dir
    sql_dir="$(resolve_versioned_sql_dir "kweaver-dip" "${HELM_CHART_VERSION:-}")"

    if ! is_rds_internal; then
        warn_external_rds_sql_required "KWeaver DIP" "${sql_dir}"
        log_warn "Skipping automatic KWeaver DIP database initialization (external RDS)"
        return 0
    fi

    init_module_database_if_present "kweaver-dip" "${sql_dir}" "KWeaver DIP"
}

# Inject DIP-side defaults: businessDomain.enabled=true unless overridden by user.
# Also propagate the same default into core (CORE_SET_VALUES) so that when DIP
# triggers install_core for missing releases, core gets businessDomain.enabled=true
# instead of core's own default (false).
_dip_apply_default_set_values() {
    if ! get_set_value "businessDomain.enabled" "${DIP_SET_VALUES[@]}" >/dev/null 2>&1; then
        DIP_SET_VALUES+=("businessDomain.enabled=true")
        log_info "Default applied for DIP: --set businessDomain.enabled=true (override with --set businessDomain.enabled=false)"
    fi
    if ! get_set_value "businessDomain.enabled" "${CORE_SET_VALUES[@]}" >/dev/null 2>&1; then
        local dip_bd
        dip_bd="$(get_set_value "businessDomain.enabled" "${DIP_SET_VALUES[@]}")"
        CORE_SET_VALUES+=("businessDomain.enabled=${dip_bd}")
        log_info "Propagated to core for DIP install: --set businessDomain.enabled=${dip_bd}"
    fi
}

# Install DIP services via Helm
install_dip() {
    log_info "Installing KWeaver DIP services via Helm..."
    _dip_require_version_manifest || return 1
    _dip_apply_default_set_values

    local charts_dir
    charts_dir="$(_dip_resolve_charts_dir)"

    local use_local=false
    if [[ -n "${charts_dir}" && -d "${charts_dir}" ]]; then
        use_local=true
        CORE_LOCAL_CHARTS_DIR="${charts_dir}"
        ISF_LOCAL_CHARTS_DIR="${charts_dir}"
        log_info "Using local DIP charts from: ${charts_dir}"
    fi

    if ! ensure_platform_prerequisites; then
        log_error "Failed to ensure platform prerequisites for KWeaver DIP"
        return 1
    fi

    # Ensure kweaver-core dependencies are installed
    if ! _dip_ensure_kweaver_core; then
        log_error "kweaver-core dependency check/installation failed"
        return 1
    fi

    if ! init_dip_database; then
        log_error "Failed to initialize KWeaver DIP database"
        return 1
    fi

    # Resolve namespace
    local namespace
    namespace=$(grep "^namespace:" "${CONFIG_YAML_PATH}" 2>/dev/null | head -1 | awk '{print $2}' | tr -d "'\"")
    namespace="${namespace:-${DIP_NAMESPACE}}"

    # Create namespace if not exists
    kubectl create namespace "${namespace}" 2>/dev/null || true

    # Resolve chart source: local directory takes priority over remote repo
    if [[ "${use_local}" != "true" ]]; then
        if [[ -z "${HELM_CHART_REPO_NAME}" ]]; then
            HELM_CHART_REPO_NAME="kweaver"
        fi
        log_info "No explicit local DIP charts directory provided, using Helm repo."
        log_info "  Version:   ${HELM_CHART_VERSION:-latest}"
        if [[ -n "${DIP_VERSION_MANIFEST_FILE:-}" ]]; then
            log_info "  Version Manifest: ${DIP_VERSION_MANIFEST_FILE}"
        fi
        log_info "  Helm Repo: ${HELM_CHART_REPO_NAME} -> ${HELM_CHART_REPO_URL}"
        ensure_helm_repo "${HELM_CHART_REPO_NAME}" "${HELM_CHART_REPO_URL}"
    fi

    log_info "Target namespace: ${namespace}"

    local release_name
    local release_version
    local chart_name
    while IFS= read -r release_name; do
        [[ -n "${release_name}" ]] || continue
        release_version="$(_dip_resolve_release_version "${release_name}")"
        chart_name="$(_dip_resolve_chart_name "${release_name}")"
        if [[ "${use_local}" == "true" ]]; then
            _install_dip_release_local "${release_name}" "${charts_dir}" "${namespace}"
        else
            _install_dip_release_repo "${release_name}" "${chart_name}" "${namespace}" "${HELM_CHART_REPO_NAME}" "${release_version}"
        fi
    done < <(_dip_release_names)

    log_info "KWeaver DIP services installation completed."
    
    # Install Etrino services (vega-hdfs, vega-calculate, vega-metadata)
    log_info "Installing Etrino services..."
    local etrino_script="${SCRIPT_DIR}/scripts/services/etrino.sh"
    if [[ -f "${etrino_script}" ]]; then
        NAMESPACE="${namespace}" CONFIG_FILE="${CONFIG_YAML_PATH}" bash "${etrino_script}" || {
            log_warn "Etrino installation encountered issues, but continuing..."
        }
    else
        log_warn "Etrino installation script not found at ${etrino_script}, skipping..."
    fi
    
    _dip_show_access_hints
}

download_dip() {
    log_info "Downloading KWeaver DIP charts..."
    ensure_helm_available
    _dip_require_version_manifest || return 1

    HELM_CHART_REPO_NAME="${HELM_CHART_REPO_NAME:-kweaver}"
    HELM_CHART_REPO_URL="${HELM_CHART_REPO_URL:-https://kweaver-ai.github.io/helm-repo/}"

    local charts_dir
    charts_dir="$(_dip_download_charts_dir)"

    local original_core_charts_dir="${CORE_LOCAL_CHARTS_DIR:-}"
    local original_isf_charts_dir="${ISF_LOCAL_CHARTS_DIR:-}"
    local original_enable_isf="${ENABLE_ISF:-}"
    local original_chart_version="${HELM_CHART_VERSION:-}"
    local original_core_manifest="${CORE_VERSION_MANIFEST_FILE:-}"
    CORE_LOCAL_CHARTS_DIR="${charts_dir}"
    ISF_LOCAL_CHARTS_DIR="${charts_dir}"
    ENABLE_ISF="true"

    ensure_helm_repo "${HELM_CHART_REPO_NAME}" "${HELM_CHART_REPO_URL}"
    if [[ -n "${DIP_VERSION_MANIFEST_FILE:-}" ]]; then
        CORE_VERSION_MANIFEST_FILE="$(_dip_resolve_core_dependency_manifest)"
        HELM_CHART_VERSION="$(_dip_resolve_core_dependency_version)"
    fi
    download_core
    HELM_CHART_VERSION="${original_chart_version}"
    CORE_VERSION_MANIFEST_FILE="${original_core_manifest}"

    local release_name
    while IFS= read -r release_name; do
        [[ -n "${release_name}" ]] || continue
        local release_version
        local chart_name
        release_version="$(_dip_resolve_release_version "${release_name}")"
        chart_name="$(_dip_resolve_chart_name "${release_name}")"
        download_chart_to_cache "${charts_dir}" "${HELM_CHART_REPO_NAME}" "${chart_name}" "${release_version}" "${FORCE_REFRESH_CHARTS:-false}"
    done < <(_dip_release_names)

    CORE_LOCAL_CHARTS_DIR="${original_core_charts_dir}"
    ISF_LOCAL_CHARTS_DIR="${original_isf_charts_dir}"
    ENABLE_ISF="${original_enable_isf}"
}

# Install a single DIP release from a local .tgz chart file
_install_dip_release_local() {
    local release_name="$1"
    local charts_dir="$2"
    local namespace="$3"
    local requested_version
    local chart_name

    requested_version="$(_dip_resolve_release_version "${release_name}")"
    chart_name="$(_dip_resolve_chart_name "${release_name}")"

    local chart_tgz=""
    if [[ -n "${requested_version}" ]]; then
        chart_tgz="$(find_cached_chart_tgz_by_version "${charts_dir}" "${chart_name}" "${requested_version}" || true)"
    fi
    if [[ -z "${chart_tgz}" ]]; then
        chart_tgz="$(_dip_find_local_chart "${charts_dir}" "${chart_name}")"
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
        "upgrade" "--install" "${release_name}"
        "${chart_tgz}"
        "--namespace" "${namespace}"
        "-f" "${CONFIG_YAML_PATH}"
        "--wait" "--timeout=600s"
    )

    local set_value
    for set_value in "${DIP_SET_VALUES[@]}"; do
        helm_args+=("--set" "${set_value}")
    done

    _dip_append_release_extra_helm_args "${release_name}" helm_args || return 1

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
    local chart_name="$2"
    local namespace="$3"
    local helm_repo_name="$4"
    local release_version="$5"

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

    log_info "Installing ${release_name} from repo..."

    local chart_ref="${helm_repo_name}/${chart_name}"

    local -a helm_args=(
        "upgrade" "--install" "${release_name}"
        "${chart_ref}"
        "--namespace" "${namespace}"
        "-f" "${CONFIG_YAML_PATH}"
    )

    local set_value
    for set_value in "${DIP_SET_VALUES[@]}"; do
        helm_args+=("--set" "${set_value}")
    done

    _dip_append_release_extra_helm_args "${release_name}" helm_args || return 1

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
    _dip_require_version_manifest || return 1

    local namespace
    namespace=$(grep "^namespace:" "${CONFIG_YAML_PATH}" 2>/dev/null | head -1 | awk '{print $2}' | tr -d "'\"")
    namespace="${namespace:-${DIP_NAMESPACE}}"

    # Uninstall in reverse install order
    local release_name
    while IFS= read -r release_name; do
        [[ -n "${release_name}" ]] || continue
        log_info "Uninstalling ${release_name}..."
        if helm uninstall "${release_name}" -n "${namespace}" 2>/dev/null; then
            log_info "✓ ${release_name} uninstalled successfully"
        else
            log_warn "⚠ ${release_name} not found or already uninstalled"
        fi
    done < <(_dip_release_names_reverse)

    log_info "KWeaver DIP services uninstallation completed."
}

# Show DIP services status
show_dip_status() {
    log_info "KWeaver DIP services status:"
    _dip_require_version_manifest || return 1

    local namespace
    namespace=$(grep "^namespace:" "${CONFIG_YAML_PATH}" 2>/dev/null | head -1 | awk '{print $2}' | tr -d "'\"")
    namespace="${namespace:-${DIP_NAMESPACE}}"

    log_info "Namespace: ${namespace}"
    log_info ""

    local release_name
    while IFS= read -r release_name; do
        [[ -n "${release_name}" ]] || continue
        if helm status "${release_name}" -n "${namespace}" >/dev/null 2>&1; then
            local status
            status=$(helm status "${release_name}" -n "${namespace}" -o json 2>/dev/null | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
            log_info "  ✓ ${release_name}: ${status}"
        else
            log_info "  ✗ ${release_name}: not installed"
        fi
    done < <(_dip_release_names)
}
