
# =============================================================================
# Kubernetes Infrastructure Initialization Script
# =============================================================================
# Features:
#   1. Initialize K8s master node with scheduling enabled
#   2. Auto-install CNI (Calico) and DNS (CoreDNS)
#   3. Install Helm 3
#   4. Install single-node MariaDB 11 via Helm
#   5. Install single-node Redis 7 via Helm
# =============================================================================

# =============================================================================
# Global Configuration Variables
# =============================================================================
# Script directory (used for local chart paths)
SCRIPT_DIR="${SCRIPT_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"

# Local config/manifest directory (vendored files to avoid runtime fetching)
CONF_DIR="${CONF_DIR:-${SCRIPT_DIR}/conf}"

CONFIG_YAML_PATH="${CONFIG_YAML_PATH:-${CONF_DIR}/config.yaml}"

AUTO_GENERATE_CONFIG="${AUTO_GENERATE_CONFIG:-true}"
DEFAULT_SQL_VERSION="${DEFAULT_SQL_VERSION:-0.5.0}"

# Local Helm charts directory
LOCAL_CHARTS_DIR="${LOCAL_CHARTS_DIR:-${SCRIPT_DIR}/charts}"
SHARED_CHARTS_DIR="${SHARED_CHARTS_DIR:-${SCRIPT_DIR}/.tmp/charts}"

# Default namespace for infrastructure components (MariaDB/Redis/Kafka/OpenSearch, etc.)
RESOURCE_NAMESPACE="${RESOURCE_NAMESPACE:-resource}"

# Generate a random password
generate_random_password() {
    local length="${1:-16}"
    # Use tr to get only alphanumeric characters and some safe special characters
    cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w "${length}" | head -n 1
}

# Check if component is already installed in Helm
is_helm_installed() {
    local release="$1"
    local ns="$2"
    helm list -n "${ns}" --short | grep -q "^${release}$"
}

# Get currently installed chart version for a release.
# Args: <release_name> <namespace> [chart_name]
get_installed_chart_version() {
    local release_name="$1"
    local namespace="$2"
    local chart_name="${3:-}"

    local installed_chart
    installed_chart=$(helm list -n "${namespace}" --filter "^${release_name}$" -o json 2>/dev/null \
        | grep -o '"chart":"[^"]*"' | head -1 | sed -e 's/^"chart":"//' -e 's/"$//')

    if [[ -z "${installed_chart}" ]]; then
        return 0
    fi

    if [[ -n "${chart_name}" && "${installed_chart}" == "${chart_name}-"* ]]; then
        echo "${installed_chart#${chart_name}-}"
        return 0
    fi

    # Fallback format: chart string is usually <chartName>-<version>
    echo "${installed_chart##*-}"
}

# Get latest chart version from Helm repo metadata.
# Args: <repo_name> <chart_name>
get_repo_chart_latest_version() {
    local repo_name="$1"
    local chart_name="$2"
    helm search repo "${repo_name}/${chart_name}" --devel -l 2>/dev/null | awk 'NR==2 {print $2}'
}

# Resolve the shared local cache directory for downloaded application charts.
resolve_shared_charts_dir() {
    echo "${SHARED_CHARTS_DIR}"
}

# Remove the default shared chart cache before an install that does not use an
# explicit local charts directory.
# Args: [explicit_charts_dir]
clear_shared_charts_cache_for_install() {
    local explicit_charts_dir="${1:-}"
    if [[ -n "${explicit_charts_dir}" ]]; then
        return 0
    fi

    local shared_dir
    shared_dir="$(resolve_shared_charts_dir)"
    if [[ -d "${shared_dir}" ]]; then
        rm -rf "${shared_dir}"
    fi
}

# Ensure a chart directory exists and print its absolute path.
# Args: <charts_dir>
ensure_charts_dir() {
    local charts_dir="$1"
    mkdir -p "${charts_dir}"
    (
        cd "${charts_dir}" >/dev/null 2>&1
        pwd
    )
}

# List cached chart tarballs whose filenames share the requested chart prefix.
# Args: <charts_dir> <chart_name>
list_cached_chart_candidates() {
    local charts_dir="$1"
    local chart_name="$2"
    find "${charts_dir}" -maxdepth 1 -name "${chart_name}-*.tgz" 2>/dev/null | sort -V
}

# Read the embedded chart name from a local .tgz package.
# Args: <chart_tgz_path>
get_local_chart_name() {
    local chart_tgz="$1"
    helm show chart "${chart_tgz}" 2>/dev/null | awk '/^name:[[:space:]]/ {sub(/^name:[[:space:]]*/, "", $0); print; exit}'
}

# Find the newest cached chart tarball for a chart name.
# Args: <charts_dir> <chart_name>
find_cached_chart_tgz() {
    local charts_dir="$1"
    local chart_name="$2"
    local chart_tgz
    local resolved_chart_name
    local latest_match=""

    while IFS= read -r chart_tgz; do
        [[ -n "${chart_tgz}" ]] || continue
        resolved_chart_name="$(get_local_chart_name "${chart_tgz}")"
        if [[ "${resolved_chart_name}" == "${chart_name}" ]]; then
            latest_match="${chart_tgz}"
        fi
    done < <(list_cached_chart_candidates "${charts_dir}" "${chart_name}")

    echo "${latest_match}"
}

# Extract chart version from a chart tarball filename.
# Args: <chart_tgz_path> <chart_name>
get_chart_version_from_filename() {
    local chart_tgz="$1"
    local chart_name="$2"
    local filename
    filename="$(basename "${chart_tgz}")"
    filename="${filename%.tgz}"
    filename="${filename#${chart_name}-}"
    echo "${filename}"
}

# Get the latest cached chart version from a directory.
# Args: <charts_dir> <chart_name>
get_cached_chart_latest_version() {
    local charts_dir="$1"
    local chart_name="$2"
    local chart_tgz
    chart_tgz="$(find_cached_chart_tgz "${charts_dir}" "${chart_name}")"
    if [[ -z "${chart_tgz}" ]]; then
        return 0
    fi

    local chart_version
    chart_version="$(get_local_chart_version "${chart_tgz}")"
    if [[ -n "${chart_version}" ]]; then
        echo "${chart_version}"
        return 0
    fi

    get_chart_version_from_filename "${chart_tgz}" "${chart_name}"
}

# Compare semantic-like versions using sort -V.
# Return 0 when the first version is newer than the second.
# Args: <lhs_version> <rhs_version>
version_gt() {
    local lhs="$1"
    local rhs="$2"

    if [[ -z "${lhs}" ]]; then
        return 1
    fi
    if [[ -z "${rhs}" ]]; then
        return 0
    fi
    [[ "$(printf '%s\n%s\n' "${lhs}" "${rhs}" | sort -V | tail -1)" == "${lhs}" && "${lhs}" != "${rhs}" ]]
}

# Download a chart to the local cache if needed.
# Args: <charts_dir> <repo_name> <chart_name> [chart_version] [force_refresh]
download_chart_to_cache() {
    local charts_dir="$1"
    local repo_name="$2"
    local chart_name="$3"
    local requested_version="${4:-}"
    local force_refresh="${5:-false}"

    charts_dir="$(ensure_charts_dir "${charts_dir}")"

    local target_version="${requested_version}"
    if [[ -z "${target_version}" ]]; then
        target_version="$(get_repo_chart_latest_version "${repo_name}" "${chart_name}")"
        if [[ -z "${target_version}" ]]; then
            log_error "Failed to resolve latest chart version for ${repo_name}/${chart_name}"
            return 1
        fi
    fi

    local cached_version
    cached_version="$(get_cached_chart_latest_version "${charts_dir}" "${chart_name}")"

    if [[ "${force_refresh}" != "true" ]]; then
        if [[ -n "${requested_version}" ]]; then
            if [[ "${cached_version}" == "${requested_version}" ]] || [[ -n "$(find "${charts_dir}" -maxdepth 1 -name "${chart_name}-${requested_version}.tgz" -print -quit 2>/dev/null)" ]]; then
                log_info "Skip download ${chart_name}: cached version ${requested_version} already exists."
                return 0
            fi
        elif [[ -n "${cached_version}" ]] && ! version_gt "${target_version}" "${cached_version}"; then
            log_info "Skip download ${chart_name}: cached version ${cached_version} is current."
            return 0
        fi
    fi

    log_info "Downloading ${repo_name}/${chart_name} ${target_version} to ${charts_dir}..."
    helm pull "${repo_name}/${chart_name}" \
        --version "${target_version}" \
        --devel \
        --destination "${charts_dir}"
}

# Ensure a Helm repo is registered and refreshed.
# Args: <repo_name> <repo_url>
ensure_helm_repo() {
    local repo_name="$1"
    local repo_url="$2"
    helm repo add --force-update "${repo_name}" "${repo_url}" || true
    helm repo update "${repo_name}" || true
}

# Ensure helm is available before running chart download logic.
ensure_helm_available() {
    if type -P helm >/dev/null 2>&1; then
        return 0
    fi

    log_info "Helm not found; installing it before continuing..."
    install_helm
}

# Get chart version from local .tgz package.
# Args: <chart_tgz_path>
get_local_chart_version() {
    local chart_tgz="$1"
    helm show chart "${chart_tgz}" 2>/dev/null | awk '$1=="version:" {print $2; exit}'
}

# Find a cached chart tarball for an exact chart version.
# Args: <charts_dir> <chart_name> <chart_version>
find_cached_chart_tgz_by_version() {
    local charts_dir="$1"
    local chart_name="$2"
    local chart_version="$3"
    local chart_tgz
    local resolved_chart_name
    local resolved_chart_version

    while IFS= read -r chart_tgz; do
        [[ -n "${chart_tgz}" ]] || continue
        resolved_chart_name="$(get_local_chart_name "${chart_tgz}")"
        resolved_chart_version="$(get_local_chart_version "${chart_tgz}")"
        if [[ "${resolved_chart_name}" == "${chart_name}" && "${resolved_chart_version}" == "${chart_version}" ]]; then
            echo "${chart_tgz}"
            return 0
        fi
    done < <(list_cached_chart_candidates "${charts_dir}" "${chart_name}")

    return 1
}

_manifest_fail() {
    echo "$1" >&2
    return 1
}

_manifest_strip_quotes() {
    local value="${1:-}"
    value="${value%\"}"
    value="${value#\"}"
    value="${value%\'}"
    value="${value#\'}"
    echo "${value}"
}

_manifest_read_top_level_value() {
    local manifest_file="$1"
    local key="$2"

    awk -F': ' -v key="${key}" '
        $1 == key { print $2; exit }
    ' "${manifest_file}" | sed 's/[[:space:]]*$//'
}

_manifest_validate_identity() {
    local manifest_file="$1"
    local expected_product="${2:-}"
    local expected_version="${3:-}"

    [[ -f "${manifest_file}" ]] || _manifest_fail "Manifest file not found: ${manifest_file}" || return 1

    local actual_product actual_version
    actual_product="$(_manifest_strip_quotes "$(_manifest_read_top_level_value "${manifest_file}" "product")")"
    actual_version="$(_manifest_strip_quotes "$(_manifest_read_top_level_value "${manifest_file}" "version")")"

    if [[ -n "${expected_product}" && "${actual_product}" != "${expected_product}" ]]; then
        _manifest_fail "Manifest product mismatch for ${manifest_file}: expected ${expected_product}, got ${actual_product:-<empty>}"
        return 1
    fi

    if [[ -n "${expected_version}" && "${actual_version}" != "${expected_version}" ]]; then
        _manifest_fail "Manifest version mismatch for ${manifest_file}: expected ${expected_version}, got ${actual_version:-<empty>}"
        return 1
    fi
}

_manifest_read_release_field() {
    local manifest_file="$1"
    local release_name="$2"
    local field_name="$3"

    awk -v release="${release_name}" -v field="${field_name}" '
        BEGIN {
            in_releases = 0
            in_target = 0
        }
        /^releases:/ {
            in_releases = 1
            next
        }
        in_releases && /^[A-Za-z0-9_-]+:/ {
            in_releases = 0
        }
        !in_releases { next }
        $0 == "  " release ":" {
            in_target = 1
            next
        }
        in_target && $0 ~ /^  [^[:space:]][^:]*:/ {
            in_target = 0
        }
        in_target && $1 == field ":" {
            print $2
            exit
        }
    ' "${manifest_file}" | sed 's/[[:space:]]*$//'
}

_manifest_list_release_names() {
    local manifest_file="$1"

    awk '
        BEGIN {
            in_releases = 0
        }
        /^releases:/ {
            in_releases = 1
            next
        }
        in_releases && /^[A-Za-z0-9_-]+:/ {
            in_releases = 0
        }
        !in_releases { next }
        /^  [^[:space:]][^:]*:/ {
            line = $0
            sub(/^  /, "", line)
            sub(/:.*/, "", line)
            print line
        }
    ' "${manifest_file}"
}

_manifest_read_dependency_field() {
    local manifest_file="$1"
    local dependency_product="$2"
    local field_name="$3"

    awk -v dependency="${dependency_product}" -v field="${field_name}" '
        BEGIN {
            in_dependencies = 0
            in_target = 0
        }
        /^dependencies:/ {
            in_dependencies = 1
            next
        }
        in_dependencies && /^[A-Za-z0-9_-]+:/ {
            in_dependencies = 0
        }
        !in_dependencies { next }
        $1 == "-" && $2 == "product:" {
            in_target = ($3 == dependency)
            next
        }
        in_target && $1 == field ":" {
            print $2
            exit
        }
    ' "${manifest_file}" | sed 's/[[:space:]]*$//'
}

# Resolve the embedded release manifest path for one aggregate product version.
# Args: <product> <version>
resolve_embedded_release_manifest() {
    local product="$1"
    local version="${2:-}"

    if [[ -z "${product}" || -z "${version}" ]]; then
        return 0
    fi

    local candidate="${RELEASE_MANIFESTS_DIR}/${version}/${product}.yaml"
    if [[ -f "${candidate}" ]]; then
        echo "${candidate}"
    fi
}

# Resolve the latest embedded release manifest path for one aggregate product.
# Args: <product>
resolve_latest_embedded_release_manifest() {
    local product="$1"

    if [[ -z "${product}" ]] || [[ ! -d "${RELEASE_MANIFESTS_DIR}" ]]; then
        return 0
    fi

    find "${RELEASE_MANIFESTS_DIR}" -mindepth 2 -maxdepth 2 -type f -name "${product}.yaml" 2>/dev/null \
        | sort -V \
        | tail -1
}

# Resolve the exact chart version for one aggregate release.
# Args: <manifest_file> <expected_product> <aggregate_version> <release_name> [fallback_version]
resolve_release_chart_version() {
    local manifest_file="${1:-}"
    local expected_product="$2"
    local aggregate_version="${3:-}"
    local release_name="$4"
    local fallback_version="${5:-}"

    if [[ -z "${manifest_file}" ]]; then
        echo "${fallback_version}"
        return 0
    fi

    get_release_manifest_release_version "${manifest_file}" "${expected_product}" "${aggregate_version}" "${release_name}"
}

# Resolve the chart name for one aggregate release.
# Args: <manifest_file> <expected_product> <aggregate_version> <release_name> [fallback_chart_name]
resolve_release_chart_name() {
    local manifest_file="${1:-}"
    local expected_product="$2"
    local aggregate_version="${3:-}"
    local release_name="$4"
    local fallback_chart_name="${5:-${release_name}}"

    if [[ -z "${manifest_file}" ]]; then
        echo "${fallback_chart_name}"
        return 0
    fi

    get_release_manifest_release_chart_name "${manifest_file}" "${expected_product}" "${aggregate_version}" "${release_name}"
}

# Get one release's exact chart version from a release manifest.
# Args: <manifest_file> <expected_product> <aggregate_version> <release_name>
get_release_manifest_release_version() {
    local manifest_file="$1"
    local expected_product="$2"
    local aggregate_version="${3:-}"
    local release_name="$4"

    _manifest_validate_identity "${manifest_file}" "${expected_product}" "${aggregate_version}" || return 1

    local value
    value="$(_manifest_strip_quotes "$(_manifest_read_release_field "${manifest_file}" "${release_name}" "version")")"
    if [[ -z "${value}" ]]; then
        _manifest_fail "Release version missing in manifest: ${release_name}"
        return 1
    fi
    echo "${value}"
}

# Get one release's chart name from a release manifest.
# Args: <manifest_file> <expected_product> <aggregate_version> <release_name>
get_release_manifest_release_chart_name() {
    local manifest_file="$1"
    local expected_product="$2"
    local aggregate_version="${3:-}"
    local release_name="$4"

    _manifest_validate_identity "${manifest_file}" "${expected_product}" "${aggregate_version}" || return 1

    local value
    value="$(_manifest_strip_quotes "$(_manifest_read_release_field "${manifest_file}" "${release_name}" "chart")")"
    if [[ -z "${value}" ]]; then
        echo "${release_name}"
        return 0
    fi

    echo "${value}"
}

# List release names from a release manifest in manifest order.
# Args: <manifest_file> <expected_product> <aggregate_version>
get_release_manifest_release_names() {
    local manifest_file="$1"
    local expected_product="$2"
    local aggregate_version="${3:-}"

    _manifest_validate_identity "${manifest_file}" "${expected_product}" "${aggregate_version}" || return 1
    _manifest_list_release_names "${manifest_file}"
}

# Get one release's install stage from a release manifest.
# Args: <manifest_file> <expected_product> <aggregate_version> <release_name>
get_release_manifest_release_stage() {
    local manifest_file="$1"
    local expected_product="$2"
    local aggregate_version="${3:-}"
    local release_name="$4"

    _manifest_validate_identity "${manifest_file}" "${expected_product}" "${aggregate_version}" || return 1

    local value
    value="$(_manifest_strip_quotes "$(_manifest_read_release_field "${manifest_file}" "${release_name}" "stage")")"
    if [[ -z "${value}" ]]; then
        echo "main"
        return 0
    fi

    case "${value}" in
        pre|main|post)
            echo "${value}"
            ;;
        *)
            _manifest_fail "Unsupported release stage in manifest for ${release_name}: ${value} (expected pre, main, or post)"
            return 1
            ;;
    esac
}

# Get one dependency's aggregate version from a release manifest.
# Args: <manifest_file> <dependency_product>
get_release_manifest_dependency_version() {
    local manifest_file="$1"
    local dependency_product="$2"

    [[ -f "${manifest_file}" ]] || _manifest_fail "Manifest file not found: ${manifest_file}" || return 1

    local value
    value="$(_manifest_strip_quotes "$(_manifest_read_dependency_field "${manifest_file}" "${dependency_product}" "version")")"
    if [[ -z "${value}" ]]; then
        _manifest_fail "Dependency version missing in manifest: ${dependency_product}"
        return 1
    fi
    echo "${value}"
}

# Get one dependency's manifest file from a release manifest.
# Args: <manifest_file> <dependency_product>
get_release_manifest_dependency_manifest() {
    local manifest_file="$1"
    local dependency_product="$2"
    local manifest_dir

    [[ -f "${manifest_file}" ]] || _manifest_fail "Manifest file not found: ${manifest_file}" || return 1

    local value
    value="$(_manifest_strip_quotes "$(_manifest_read_dependency_field "${manifest_file}" "${dependency_product}" "manifest")")"
    if [[ -z "${value}" ]]; then
        _manifest_fail "Dependency manifest missing in manifest: ${dependency_product}"
        return 1
    fi

    if [[ "${value}" == /* ]]; then
        echo "${value}"
        return 0
    fi

    manifest_dir="$(cd "$(dirname "${manifest_file}")" && pwd)"
    echo "$(cd "${manifest_dir}" && cd "$(dirname "${value}")" && pwd)/$(basename "${value}")"
}

# Get one dependency's aggregate version from a release manifest (optional, returns empty if not found).
# Args: <manifest_file> <dependency_product>
get_release_manifest_dependency_version_optional() {
    local manifest_file="$1"
    local dependency_product="$2"

    [[ -f "${manifest_file}" ]] || return 0

    local value
    value="$(_manifest_strip_quotes "$(_manifest_read_dependency_field "${manifest_file}" "${dependency_product}" "version")")"
    echo "${value}"
}

# Get one dependency's manifest file from a release manifest (optional, returns empty if not found).
# Args: <manifest_file> <dependency_product>
get_release_manifest_dependency_manifest_optional() {
    local manifest_file="$1"
    local dependency_product="$2"
    local manifest_dir

    [[ -f "${manifest_file}" ]] || return 0

    local value
    value="$(_manifest_strip_quotes "$(_manifest_read_dependency_field "${manifest_file}" "${dependency_product}" "manifest")")"
    if [[ -z "${value}" ]]; then
        return 0
    fi

    if [[ "${value}" == /* ]]; then
        echo "${value}"
        return 0
    fi

    manifest_dir="$(cd "$(dirname "${manifest_file}")" && pwd)"
    echo "$(cd "${manifest_dir}" && cd "$(dirname "${value}")" && pwd)/$(basename "${value}")"
}

# Decide whether upgrade can be skipped when installed chart version equals target version.
# Return 0 => skip upgrade, Return 1 => continue upgrade.
# Args: <release_name> <namespace> <chart_name> <target_version>
should_skip_upgrade_same_chart_version() {
    local release_name="$1"
    local namespace="$2"
    local chart_name="$3"
    local target_version="$4"

    if [[ -z "${target_version}" ]]; then
        return 1
    fi

    local current_status
    current_status=$(helm status "${release_name}" -n "${namespace}" -o json 2>/dev/null \
        | grep -o '"status":"[^"]*"' | head -1 | cut -d'"' -f4)
    if [[ "${current_status}" != "deployed" ]]; then
        return 1
    fi

    local installed_version
    installed_version=$(get_installed_chart_version "${release_name}" "${namespace}" "${chart_name}")
    if [[ -n "${installed_version}" && "${installed_version}" == "${target_version}" ]]; then
        log_info "Skip ${release_name}: installed chart version ${installed_version} equals target ${target_version}."
        return 0
    fi

    return 1
}

# Get existing password from config.yaml if it exists
get_existing_password() {
    local key="$1"
    if [[ -f "${CONFIG_YAML_PATH}" ]]; then
        grep "${key}:" "${CONFIG_YAML_PATH}" | awk '{print $2}' | tr -d '"'\'' '
    fi
}

resolve_sql_version() {
    local requested_version="${1:-}"
    if [[ -n "${requested_version}" ]]; then
        echo "${requested_version}"
        return 0
    fi

    echo "${DEFAULT_SQL_VERSION}"
}

# Resolve the SQL base directory for one product/version pair.
# Args: <product> [version]
resolve_versioned_sql_dir() {
    local product="$1"
    local version="${2:-}"
    local resolved_version

    if [[ -z "${product}" ]]; then
        return 0
    fi

    resolved_version="$(resolve_sql_version "${version}")"
    echo "${SCRIPT_DIR}/scripts/sql/${resolved_version}/${product}"
}

# Return 0 when a directory exists and contains at least one .sql file.
# Args: <sql_dir>
sql_dir_has_files() {
    local sql_dir="$1"
    [[ -d "${sql_dir}" ]] || return 1
    find "${sql_dir}" -type f -name "*.sql" -print -quit 2>/dev/null | grep -q .
}

# List module subdirectories under one product/version SQL directory.
# Args: <product> [version]
list_versioned_sql_modules() {
    local product="$1"
    local version="${2:-}"
    local sql_base_dir

    sql_base_dir="$(resolve_versioned_sql_dir "${product}" "${version}")"
    [[ -d "${sql_base_dir}" ]] || return 0

    find "${sql_base_dir}" -mindepth 1 -maxdepth 1 -type d -printf '%f\n' 2>/dev/null | sort
}

# Execute one SQL directory only when it exists and contains SQL files.
# Args: <module_name> <sql_dir> [display_name]
init_module_database_if_present() {
    local module_name="$1"
    local sql_dir="$2"
    local display_name="${3:-${module_name}}"

    if ! sql_dir_has_files "${sql_dir}"; then
        log_info "Skipping ${display_name} database initialization: no SQL files found in ${sql_dir}"
        return 0
    fi

    init_module_database "${module_name}" "${sql_dir}"
}

# Check if RDS is internal (MariaDB installed in cluster)
is_rds_internal() {
    if [[ ! -f "${CONFIG_YAML_PATH}" ]]; then
        return 1
    fi
    # Check if rds section has source_type: internal
    grep -A 10 "^  rds:" "${CONFIG_YAML_PATH}" | grep -q "source_type: internal"
}

# Show prominent warning when RDS is external and manual SQL import is required
warn_external_rds_sql_required() {
    local module_name="$1"
    local sql_dir="$2"
    
    echo ""
    echo "╔════════════════════════════════════════════════════════════════════════════╗"
    echo "║                                                                            ║"
    echo "║  ⚠️  WARNING: EXTERNAL DATABASE - MANUAL SQL INITIALIZATION REQUIRED  ⚠️   ║"
    echo "║                                                                            ║"
    echo "╠════════════════════════════════════════════════════════════════════════════╣"
    echo "║                                                                            ║"
    echo "║  RDS source_type is set to 'external' in config.yaml.                      ║"
    echo "║  You MUST manually execute SQL scripts to initialize the database.         ║"
    echo "║                                                                            ║"
    echo "║  Module: ${module_name}"
    echo "║  SQL Directory: ${sql_dir}"
    echo "║                                                                            ║"
    echo "║  Steps:                                                                    ║"
    echo "║    1. Connect to your external database server                             ║"
    echo "║    2. Execute all .sql files in the directory above                        ║"
    echo "║    3. Ensure all required databases and tables are created                 ║"
    echo "║                                                                            ║"
    echo "╚════════════════════════════════════════════════════════════════════════════╝"
    echo ""
}

# Image registry prefix loaded from conf/config.yaml (image.registry) or env
IMAGE_REGISTRY="${IMAGE_REGISTRY:-}"

# Kubernetes Network Configuration
POD_CIDR="${POD_CIDR:-192.169.0.0/16}"
SERVICE_CIDR="${SERVICE_CIDR:-10.96.0.0/12}"

# Kubernetes API Server Configuration
API_SERVER_ADVERTISE_ADDRESS="${API_SERVER_ADVERTISE_ADDRESS:-}"

# Kubernetes Image Repository Configuration
IMAGE_REPOSITORY="${IMAGE_REPOSITORY:-registry.aliyuncs.com/google_containers}"

# Kubernetes yum repo (Aliyun mirror) for kubeadm/kubelet/kubectl/cri-tools
K8S_RPM_REPO_BASEURL="${K8S_RPM_REPO_BASEURL:-https://mirrors.aliyun.com/kubernetes-new/core/stable/v1.28/rpm/}"
K8S_RPM_REPO_GPGKEY="${K8S_RPM_REPO_GPGKEY:-https://mirrors.aliyun.com/kubernetes-new/core/stable/v1.28/rpm/repodata/repomd.xml.key}"

# Flannel CNI Image Repository Configuration
FLANNEL_IMAGE_REPO="${FLANNEL_IMAGE_REPO:-swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/}"
FLANNEL_MANIFEST_PATH="${FLANNEL_MANIFEST_PATH:-${CONF_DIR}/kube-flannel.yml}"
FLANNEL_MANIFEST_URL="${FLANNEL_MANIFEST_URL:-https://gitee.com/mirrors/flannel/raw/main/Documentation/kube-flannel.yml}"


# Helm Configuration
HELM_REPO_BITNAMI="${HELM_REPO_BITNAMI:-https://charts.bitnami.com/bitnami}"
HELM_REPO_INGRESS_NGINX="${HELM_REPO_INGRESS_NGINX:-https://kubernetes.github.io/ingress-nginx}"
HELM_REPO_OPENSEARCH="${HELM_REPO_OPENSEARCH:-https://opensearch-project.github.io/helm-charts}"
HELM_INSTALL_SCRIPT_PATH="${HELM_INSTALL_SCRIPT_PATH:-${CONF_DIR}/get-helm-3}"
HELM_INSTALL_SCRIPT_URL="${HELM_INSTALL_SCRIPT_URL:-https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3}"
HELM_VERSION="${HELM_VERSION:-v3.19.0}"
HELM_TARBALL_BASEURL="${HELM_TARBALL_BASEURL:-https://repo.huaweicloud.com/helm/${HELM_VERSION}/}"

# Global Helm Chart Configuration (for Studio, BKN, and other modules)
HELM_CHART_VERSION="${HELM_CHART_VERSION:-}"
HELM_CHART_REPO_URL="${HELM_CHART_REPO_URL:-https://kweaver-ai.github.io/helm-repo/}"
HELM_CHART_REPO_NAME="${HELM_CHART_REPO_NAME:-kweaver}"
RELEASE_MANIFESTS_DIR="${RELEASE_MANIFESTS_DIR:-${VERSION_MANIFESTS_DIR:-${SCRIPT_DIR}/release-manifests}}"

DOCKER_IO_MIRROR_PREFIX="${DOCKER_IO_MIRROR_PREFIX:-swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/}"
DOCKER_CE_REPO_URL="${DOCKER_CE_REPO_URL:-http://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo}"
LOCALPV_PROVISIONER_IMAGE="${LOCALPV_PROVISIONER_IMAGE:-swr.cn-east-3.myhuaweicloud.com/kweaver-ai/rancher/local-path-provisioner:v0.0.32}"
LOCALPV_HELPER_IMAGE="${LOCALPV_HELPER_IMAGE:-swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/busybox:1.36.1}"
LOCALPV_MANIFEST_PATH="${LOCALPV_MANIFEST_PATH:-${CONF_DIR}/local-path-storage.yaml}"
LOCALPV_MANIFEST_URL="${LOCALPV_MANIFEST_URL:-https://raw.githubusercontent.com/rancher/local-path-provisioner/v0.0.32/deploy/local-path-storage.yaml}"
LOCALPV_BASE_PATH="${LOCALPV_BASE_PATH:-/opt/local-path-provisioner}"
LOCALPV_SET_DEFAULT="${LOCALPV_SET_DEFAULT:-true}"
AUTO_INSTALL_LOCALPV="${AUTO_INSTALL_LOCALPV:-true}"
STORAGE_STORAGE_CLASS_NAME="${STORAGE_STORAGE_CLASS_NAME:-}"

# MariaDB Configuration
MARIADB_NAMESPACE="${MARIADB_NAMESPACE:-${RESOURCE_NAMESPACE}}"
MARIADB_IMAGE="${MARIADB_IMAGE:-}"
MARIADB_IMAGE_REPOSITORY="${MARIADB_IMAGE_REPOSITORY:-mariadb}"
MARIADB_IMAGE_TAG="${MARIADB_IMAGE_TAG:-11.4.7}"
MARIADB_IMAGE_FALLBACK="${MARIADB_IMAGE_FALLBACK:-mariadb:11.4.7}"
MARIADB_VERSION="${MARIADB_VERSION:-11.4}"
MARIADB_CHART_VERSION="${MARIADB_CHART_VERSION:-1.0.0}"
MARIADB_CHART_TGZ="${MARIADB_CHART_TGZ:-${SCRIPT_DIR}/charts/proton-mariadb-${MARIADB_CHART_VERSION}.tgz}"
MARIADB_PERSISTENCE_ENABLED="${MARIADB_PERSISTENCE_ENABLED:-true}"
MARIADB_STORAGE_CLASS="${MARIADB_STORAGE_CLASS:-}"
MARIADB_PURGE_PVC="${MARIADB_PURGE_PVC:-false}"
MARIADB_ROOT_PASSWORD="${MARIADB_ROOT_PASSWORD:-}"
MARIADB_DATABASE="${MARIADB_DATABASE:-kweaver}"
MARIADB_USER="${MARIADB_USER:-kweaver}"
MARIADB_PASSWORD="${MARIADB_PASSWORD:-kweaver}"
MARIADB_STORAGE_SIZE="${MARIADB_STORAGE_SIZE:-10Gi}"
MARIADB_MAX_CONNECTIONS="${MARIADB_MAX_CONNECTIONS:-5000}"

# Redis Configuration
REDIS_NAMESPACE="${REDIS_NAMESPACE:-${RESOURCE_NAMESPACE}}"
REDIS_VERSION="${REDIS_VERSION:-7.4}"
REDIS_CHART_VERSION="${REDIS_CHART_VERSION:-1.11.2}"
REDIS_CHART_TGZ="${REDIS_CHART_TGZ:-${SCRIPT_DIR}/charts/proton-redis-${REDIS_CHART_VERSION}.tgz}"
REDIS_LOCAL_CHART_DIR="${REDIS_LOCAL_CHART_DIR:-${SCRIPT_DIR}/charts/proton-redis}"
REDIS_ARCHITECTURE="${REDIS_ARCHITECTURE:-sentinel}"  # standalone or sentinel
REDIS_IMAGE="${REDIS_IMAGE:-}"
REDIS_IMAGE_REGISTRY="${REDIS_IMAGE_REGISTRY:-}"
REDIS_IMAGE_REPOSITORY="${REDIS_IMAGE_REPOSITORY:-proton/proton-redis}"
REDIS_IMAGE_TAG="${REDIS_IMAGE_TAG:-1.11.2-20251029.2.169ac3c0}"
REDIS_PERSISTENCE_ENABLED="${REDIS_PERSISTENCE_ENABLED:-true}"
REDIS_STORAGE_CLASS="${REDIS_STORAGE_CLASS:-}"
REDIS_PURGE_PVC="${REDIS_PURGE_PVC:-true}"
REDIS_PASSWORD="${REDIS_PASSWORD:-}"
REDIS_STORAGE_SIZE="${REDIS_STORAGE_SIZE:-5Gi}"
REDIS_MASTER_GROUP_NAME="${REDIS_MASTER_GROUP_NAME:-mymaster}"
REDIS_REPLICA_COUNT="${REDIS_REPLICA_COUNT:-1}"
REDIS_SENTINEL_QUORUM="${REDIS_SENTINEL_QUORUM:-1}"

# Kafka Configuration
KAFKA_NAMESPACE="${KAFKA_NAMESPACE:-${RESOURCE_NAMESPACE}}"
KAFKA_RELEASE_NAME="${KAFKA_RELEASE_NAME:-kafka}"
KAFKA_CHART_VERSION="${KAFKA_CHART_VERSION:-32.4.3}"
KAFKA_CHART_TGZ="${KAFKA_CHART_TGZ:-${SCRIPT_DIR}/charts/kafka-${KAFKA_CHART_VERSION}.tgz}"
# NOTE: Bitnami Kafka chart expects Bitnami Kafka images (/opt/bitnami/kafka/*).
# NOTE: Kafka 4.0 drops support for some older client protocol versions. Some apps (e.g. older Go clients)
# may still send JoinGroup v1 and will fail with:
#   UnsupportedVersionException: Received request for api with key 11 (JoinGroup) and unsupported version 1
# Default to a Kafka 3.x image for broader client compatibility; you can override via KAFKA_IMAGE/KAFKA_IMAGE_TAG.
# Use an SWR mirror by default to improve pull reliability in restricted networks.
KAFKA_IMAGE="${KAFKA_IMAGE:-swr.cn-east-3.myhuaweicloud.com/kweaver-ai/bitnami/kafka:3.9.0-debian-12-r10}"
KAFKA_IMAGE_REPOSITORY="${KAFKA_IMAGE_REPOSITORY:-bitnami/kafka}"
KAFKA_IMAGE_TAG="${KAFKA_IMAGE_TAG:-3.9.0-debian-12-r10}"
KAFKA_IMAGE_FALLBACK="${KAFKA_IMAGE_FALLBACK:-swr.cn-east-3.myhuaweicloud.com/kweaver-ai/bitnami/kafka:3.9.0-debian-12-r10}"
KAFKA_HELM_TIMEOUT="${KAFKA_HELM_TIMEOUT:-1800s}"
# NOTE: --atomic will auto-uninstall on failure, which makes debugging hard. Default to false.
KAFKA_HELM_ATOMIC="${KAFKA_HELM_ATOMIC:-false}"
KAFKA_READY_TIMEOUT="${KAFKA_READY_TIMEOUT:-600s}"
KAFKA_HEAP_OPTS="${KAFKA_HEAP_OPTS:--Xms256m -Xmx256m}"
KAFKA_MEMORY_REQUEST="${KAFKA_MEMORY_REQUEST:-256Mi}"
KAFKA_MEMORY_LIMIT="${KAFKA_MEMORY_LIMIT:-512Mi}"
KAFKA_PERSISTENCE_ENABLED="${KAFKA_PERSISTENCE_ENABLED:-true}"
KAFKA_STORAGE_CLASS="${KAFKA_STORAGE_CLASS:-}"
KAFKA_STORAGE_SIZE="${KAFKA_STORAGE_SIZE:-8Gi}"
# Delete Kafka PVCs by default on uninstall (set false to retain data)
KAFKA_PURGE_PVC="${KAFKA_PURGE_PVC:-true}"
KAFKA_AUTH_ENABLED="${KAFKA_AUTH_ENABLED:-true}"
KAFKA_PROTOCOL="${KAFKA_PROTOCOL:-SASL_PLAINTEXT}"
KAFKA_SASL_MECHANISM="${KAFKA_SASL_MECHANISM:-PLAIN}"
KAFKA_CLIENT_USER="${KAFKA_CLIENT_USER:-kafkauser}"
KAFKA_CLIENT_PASSWORD="${KAFKA_CLIENT_PASSWORD:-}"
KAFKA_INTERBROKER_USER="${KAFKA_INTERBROKER_USER:-inter_broker_user}"
KAFKA_INTERBROKER_PASSWORD="${KAFKA_INTERBROKER_PASSWORD:-}"
KAFKA_CONTROLLER_USER="${KAFKA_CONTROLLER_USER:-controller_user}"
KAFKA_CONTROLLER_PASSWORD="${KAFKA_CONTROLLER_PASSWORD:-}"
KAFKA_SASL_SECRET_NAME="${KAFKA_SASL_SECRET_NAME:-${KAFKA_RELEASE_NAME}-sasl}"
KAFKA_REPLICAS="${KAFKA_REPLICAS:-1}"
KAFKA_AUTO_CREATE_TOPICS_ENABLE="${KAFKA_AUTO_CREATE_TOPICS_ENABLE:-true}"

# OpenSearch Configuration
LOCAL_OPENSEARCH_CHARTS_DIR="${LOCAL_OPENSEARCH_CHARTS_DIR:-${SCRIPT_DIR}/charts/opensearch}"
OPENSEARCH_NAMESPACE="${OPENSEARCH_NAMESPACE:-${RESOURCE_NAMESPACE}}"
OPENSEARCH_RELEASE_NAME="${OPENSEARCH_RELEASE_NAME:-opensearch}"
OPENSEARCH_CLUSTER_NAME="${OPENSEARCH_CLUSTER_NAME:-opensearch-cluster}"
OPENSEARCH_NODE_GROUP="${OPENSEARCH_NODE_GROUP:-master}"
OPENSEARCH_CHART_VERSION="${OPENSEARCH_CHART_VERSION:-2.36.0}"
OPENSEARCH_CHART_TGZ="${OPENSEARCH_CHART_TGZ:-${SCRIPT_DIR}/charts/opensearch-${OPENSEARCH_CHART_VERSION}.tgz}"
OPENSEARCH_IMAGE="${OPENSEARCH_IMAGE:-swr.cn-east-3.myhuaweicloud.com/kweaver-ai/opensearchproject/opensearch:2.19.4}"
OPENSEARCH_IMAGE_REPOSITORY="${OPENSEARCH_IMAGE_REPOSITORY:-swr.cn-east-3.myhuaweicloud.com/kweaver-ai/opensearchproject/opensearch}"
OPENSEARCH_IMAGE_TAG="${OPENSEARCH_IMAGE_TAG:-2.19.4}"
OPENSEARCH_IMAGE_FALLBACK="${OPENSEARCH_IMAGE_FALLBACK:-swr.cn-east-3.myhuaweicloud.com/kweaver-ai/opensearchproject/opensearch:2.19.4}"
# OpenSearch chart uses busybox initContainers (fsgroup-volume/sysctl); use a dedicated SWR mirror by default.
OPENSEARCH_INIT_IMAGE="${OPENSEARCH_INIT_IMAGE:-swr.cn-east-3.myhuaweicloud.com/kweaver-ai/busybox:1.36.1}"
OPENSEARCH_JAVA_OPTS="${OPENSEARCH_JAVA_OPTS:--Xms512m -Xmx512m -XX:MaxDirectMemorySize=128m}"
OPENSEARCH_MEMORY_REQUEST="${OPENSEARCH_MEMORY_REQUEST:-512Mi}"
# NOTE: OpenSearch uses heap + direct memory + native overhead. 768Mi is too tight for -Xmx512m.
# Increased to 2Gi to support plugin installation (IK analyzer, etc.)
OPENSEARCH_MEMORY_LIMIT="${OPENSEARCH_MEMORY_LIMIT:-2048Mi}"
OPENSEARCH_PROTOCOL="${OPENSEARCH_PROTOCOL:-http}" # http (default) or https (requires enabling security)
OPENSEARCH_DISABLE_SECURITY="${OPENSEARCH_DISABLE_SECURITY:-}"
OPENSEARCH_SINGLE_NODE="${OPENSEARCH_SINGLE_NODE:-true}"
OPENSEARCH_HELM_ATOMIC="${OPENSEARCH_HELM_ATOMIC:-false}"
OPENSEARCH_PERSISTENCE_ENABLED="${OPENSEARCH_PERSISTENCE_ENABLED:-true}"
OPENSEARCH_STORAGE_CLASS="${OPENSEARCH_STORAGE_CLASS:-}"
OPENSEARCH_STORAGE_SIZE="${OPENSEARCH_STORAGE_SIZE:-8Gi}"
OPENSEARCH_PURGE_PVC="${OPENSEARCH_PURGE_PVC:-false}"
OPENSEARCH_INITIAL_ADMIN_PASSWORD="${OPENSEARCH_INITIAL_ADMIN_PASSWORD:-OpenSearch@123456}"
OPENSEARCH_SYSCTL_INIT_ENABLED="${OPENSEARCH_SYSCTL_INIT_ENABLED:-true}"
OPENSEARCH_SYSCTL_VM_MAX_MAP_COUNT="${OPENSEARCH_SYSCTL_VM_MAX_MAP_COUNT:-262144}"

# MongoDB Configuration
LOCAL_MONGODB_CHARTS_DIR="${LOCAL_MONGODB_CHARTS_DIR:-${SCRIPT_DIR}/charts/mongodb}"
MONGODB_CHART_TGZ="${MONGODB_CHART_TGZ:-${SCRIPT_DIR}/charts/mongodb-1.0.0.tgz}"
MONGODB_NAMESPACE="${MONGODB_NAMESPACE:-${RESOURCE_NAMESPACE}}"
MONGODB_RELEASE_NAME="${MONGODB_RELEASE_NAME:-mongodb}"
MONGODB_IMAGE="${MONGODB_IMAGE:-}"
MONGODB_IMAGE_REPOSITORY="${MONGODB_IMAGE_REPOSITORY:-swr.cn-east-3.myhuaweicloud.com/kweaver-ai/proton/proton-mongo}"
MONGODB_IMAGE_TAG="${MONGODB_IMAGE_TAG:-2.1.0-feature-mongo-4.4.30}"
MONGODB_REPLICAS="${MONGODB_REPLICAS:-1}"
MONGODB_REPLSET_ENABLED="${MONGODB_REPLSET_ENABLED:-true}"  # Default: single-node replica set mode (requires keyfile)
MONGODB_REPLSET_NAME="${MONGODB_REPLSET_NAME:-rs0}"
MONGODB_SERVICE_TYPE="${MONGODB_SERVICE_TYPE:-ClusterIP}"
MONGODB_SERVICE_PORT="${MONGODB_SERVICE_PORT:-30280}"
MONGODB_WIRED_TIGER_CACHE_SIZE_GB="${MONGODB_WIRED_TIGER_CACHE_SIZE_GB:-4}"
MONGODB_STORAGE_CLASS="${MONGODB_STORAGE_CLASS:-}"
MONGODB_STORAGE_SIZE="${MONGODB_STORAGE_SIZE:-10Gi}"
MONGODB_SECRET_NAME="${MONGODB_SECRET_NAME:-mongodb-secret}"
MONGODB_SECRET_USERNAME="${MONGODB_SECRET_USERNAME:-admin}"
MONGODB_SECRET_PASSWORD="${MONGODB_SECRET_PASSWORD:-}"
MONGODB_RESOURCES_REQUESTS_CPU="${MONGODB_RESOURCES_REQUESTS_CPU:-100m}"
MONGODB_RESOURCES_REQUESTS_MEMORY="${MONGODB_RESOURCES_REQUESTS_MEMORY:-128Mi}"
MONGODB_RESOURCES_LIMITS_CPU="${MONGODB_RESOURCES_LIMITS_CPU:-1}"
MONGODB_RESOURCES_LIMITS_MEMORY="${MONGODB_RESOURCES_LIMITS_MEMORY:-1Gi}"

# Zookeeper Configuration
LOCAL_ZOOKEEPER_CHARTS_DIR="${LOCAL_ZOOKEEPER_CHARTS_DIR:-${SCRIPT_DIR}/charts/zookeeper}"
ZOOKEEPER_CHART_TGZ="${ZOOKEEPER_CHART_TGZ:-${SCRIPT_DIR}/charts/proton-zookeeper-5.6.0.tgz}"
ZOOKEEPER_NAMESPACE="${ZOOKEEPER_NAMESPACE:-${RESOURCE_NAMESPACE}}"
ZOOKEEPER_RELEASE_NAME="${ZOOKEEPER_RELEASE_NAME:-zookeeper}"
ZOOKEEPER_CHART_REF="${ZOOKEEPER_CHART_REF:-}"  # e.g., "dip/zookeeper" for remote repo, or local path
ZOOKEEPER_CHART_VERSION="${ZOOKEEPER_CHART_VERSION:-}"  # Chart version (--version)
ZOOKEEPER_CHART_DEVEL="${ZOOKEEPER_CHART_DEVEL:-false}"  # Use --devel flag
ZOOKEEPER_VALUES_FILE="${ZOOKEEPER_VALUES_FILE:-}"  # Additional values file (e.g., conf/config.yaml)
ZOOKEEPER_REPLICAS="${ZOOKEEPER_REPLICAS:-1}"
ZOOKEEPER_IMAGE_REGISTRY="${ZOOKEEPER_IMAGE_REGISTRY:-swr.cn-east-3.myhuaweicloud.com/kweaver-ai}"
ZOOKEEPER_IMAGE_REPOSITORY="${ZOOKEEPER_IMAGE_REPOSITORY:-proton/proton-zookeeper}"
ZOOKEEPER_IMAGE_TAG="${ZOOKEEPER_IMAGE_TAG:-5.6.0-20250625.2.138fb9}"
ZOOKEEPER_EXPORTER_IMAGE_REPOSITORY="${ZOOKEEPER_EXPORTER_IMAGE_REPOSITORY:-proton/proton-zookeeper-exporter}"
ZOOKEEPER_EXPORTER_IMAGE_TAG="${ZOOKEEPER_EXPORTER_IMAGE_TAG:-5.6.0-20250625.2.138fb9}"
ZOOKEEPER_SERVICE_PORT="${ZOOKEEPER_SERVICE_PORT:-2181}"
ZOOKEEPER_EXPORTER_PORT="${ZOOKEEPER_EXPORTER_PORT:-9101}"
ZOOKEEPER_JMX_EXPORTER_PORT="${ZOOKEEPER_JMX_EXPORTER_PORT:-9995}"
ZOOKEEPER_STORAGE_CLASS="${ZOOKEEPER_STORAGE_CLASS:-}"
ZOOKEEPER_STORAGE_SIZE="${ZOOKEEPER_STORAGE_SIZE:-1Gi}"
ZOOKEEPER_PURGE_PVC="${ZOOKEEPER_PURGE_PVC:-true}"
ZOOKEEPER_RESOURCES_REQUESTS_CPU="${ZOOKEEPER_RESOURCES_REQUESTS_CPU:-500m}"
ZOOKEEPER_RESOURCES_REQUESTS_MEMORY="${ZOOKEEPER_RESOURCES_REQUESTS_MEMORY:-1Gi}"
ZOOKEEPER_RESOURCES_LIMITS_CPU="${ZOOKEEPER_RESOURCES_LIMITS_CPU:-1000m}"
ZOOKEEPER_RESOURCES_LIMITS_MEMORY="${ZOOKEEPER_RESOURCES_LIMITS_MEMORY:-2Gi}"
ZOOKEEPER_JVMFLAGS="${ZOOKEEPER_JVMFLAGS:--Xms500m -Xmx500m}"
ZOOKEEPER_SASL_ENABLED="${ZOOKEEPER_SASL_ENABLED:-true}"
ZOOKEEPER_SASL_USER="${ZOOKEEPER_SASL_USER:-kafka}"
ZOOKEEPER_SASL_PASSWORD="${ZOOKEEPER_SASL_PASSWORD:-}"
ZOOKEEPER_EXTRA_SET_VALUES="${ZOOKEEPER_EXTRA_SET_VALUES:-}"  # Additional --set values (space-separated, e.g., "image.registry=xxx key2=value2")

# Ingress-Nginx Configuration
INGRESS_NGINX_HTTP_PORT="${INGRESS_NGINX_HTTP_PORT:-80}"
INGRESS_NGINX_HTTPS_PORT="${INGRESS_NGINX_HTTPS_PORT:-443}"
INGRESS_NGINX_CLASS="${INGRESS_NGINX_CLASS:-class-443}"
INGRESS_NGINX_CONTROLLER_IMAGE="${INGRESS_NGINX_CONTROLLER_IMAGE:-swr.cn-east-3.myhuaweicloud.com/kweaver-ai/ingress-nginx/controller:v1.14.1}"
INGRESS_NGINX_CONTROLLER_IMAGE_REPOSITORY="${INGRESS_NGINX_CONTROLLER_IMAGE_REPOSITORY:-swr.cn-east-3.myhuaweicloud.com/kweaver-ai/ingress-nginx/controller}"
INGRESS_NGINX_CONTROLLER_IMAGE_TAG="${INGRESS_NGINX_CONTROLLER_IMAGE_TAG:-v1.14.1}"
INGRESS_NGINX_CHART_VERSION="${INGRESS_NGINX_CHART_VERSION:-4.13.1}"
INGRESS_NGINX_CHART_TGZ="${INGRESS_NGINX_CHART_TGZ:-${SCRIPT_DIR}/charts/ingress-nginx-${INGRESS_NGINX_CHART_VERSION}.tgz}"
INGRESS_NGINX_WEBHOOK_CERTGEN_IMAGE="${INGRESS_NGINX_WEBHOOK_CERTGEN_IMAGE:-swr.cn-east-3.myhuaweicloud.com/kweaver-ai/ingress-nginx/kube-webhook-certgen:v1.6.1}"
INGRESS_NGINX_WEBHOOK_CERTGEN_IMAGE_REPOSITORY="${INGRESS_NGINX_WEBHOOK_CERTGEN_IMAGE_REPOSITORY:-swr.cn-east-3.myhuaweicloud.com/kweaver-ai/ingress-nginx/kube-webhook-certgen}"
INGRESS_NGINX_WEBHOOK_CERTGEN_IMAGE_TAG="${INGRESS_NGINX_WEBHOOK_CERTGEN_IMAGE_TAG:-v1.6.1}"
INGRESS_NGINX_HOSTNETWORK="${INGRESS_NGINX_HOSTNETWORK:-true}"
INGRESS_NGINX_ADMISSION_WEBHOOKS_ENABLED="${INGRESS_NGINX_ADMISSION_WEBHOOKS_ENABLED:-false}"
AUTO_INSTALL_INGRESS_NGINX="${AUTO_INSTALL_INGRESS_NGINX:-true}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

k8s_is_running() {
    if ! command -v kubectl >/dev/null 2>&1; then
        return 1
    fi

    if kubectl get nodes >/dev/null 2>&1; then
        return 0
    fi

    if [[ -f /root/.kube/config ]]; then
        export KUBECONFIG=/root/.kube/config
        if kubectl get nodes >/dev/null 2>&1; then
            return 0
        fi
    fi

    if [[ -f /etc/kubernetes/admin.conf ]]; then
        mkdir -p /root/.kube
        cp -f /etc/kubernetes/admin.conf /root/.kube/config
        chown root:root /root/.kube/config 2>/dev/null || true
        export KUBECONFIG=/root/.kube/config
        if kubectl get nodes >/dev/null 2>&1; then
            log_info "Recovered kubeconfig from /etc/kubernetes/admin.conf"
            return 0
        fi
    fi

    return 1
}

ensure_k8s() {
    if [[ "${KWEAVER_K8S_ENSURED:-false}" == "true" ]]; then
        return 0
    fi

    if k8s_is_running; then
        log_info "Kubernetes cluster detected, skipping K8s installation."
        export KWEAVER_K8S_ENSURED="true"
        return 0
    fi

    log_info "No running Kubernetes cluster detected. Installing K8s first..."
    check_root
    detect_package_manager || return 1
    install_containerd || return 1
    install_kubernetes || return 1
    install_helm || return 1

    check_prerequisites || return 1
    init_k8s_master || return 1
    allow_master_scheduling || return 1
    install_cni || return 1
    wait_for_dns || return 1

    if [[ "${AUTO_INSTALL_LOCALPV}" == "true" ]]; then
        if [[ -z "$(kubectl get storageclass --no-headers 2>/dev/null)" ]]; then
            install_localpv || return 1
        fi
    fi

    if [[ "${AUTO_INSTALL_INGRESS_NGINX}" == "true" ]]; then
        install_ingress_nginx || return 1
    fi

    export KWEAVER_K8S_ENSURED="true"
    log_info "K8s installation completed."
}

ensure_data_services() {
    if [[ "${KWEAVER_DATA_SERVICES_ENSURED:-false}" == "true" ]]; then
        return 0
    fi

    log_info "Ensuring platform data services (MariaDB/Redis/Kafka/Zookeeper/OpenSearch)..."

    install_mariadb || return 1
    install_redis || return 1
    install_kafka || return 1
    install_zookeeper || return 1
    if [[ "${AUTO_INSTALL_INGRESS_NGINX}" == "true" ]]; then
        install_ingress_nginx || return 1
    fi
    install_opensearch || return 1

    if [[ "${AUTO_GENERATE_CONFIG}" == "true" ]]; then
        generate_config_yaml || return 1
    fi

    export KWEAVER_DATA_SERVICES_ENSURED="true"
}

ensure_platform_prerequisites() {
    if [[ "${KWEAVER_PLATFORM_PREREQUISITES_DONE:-false}" == "true" ]]; then
        return 0
    fi

    ensure_k8s || return 1
    ensure_data_services || return 1

    export KWEAVER_PLATFORM_PREREQUISITES_DONE="true"
}

get_access_address_field() {
    local field="$1"
    local cfg="${CONFIG_YAML_PATH}"

    if [[ ! -f "${cfg}" ]]; then
        return 0
    fi

    awk -v key="${field}:" '
        $1=="accessAddress:" {in_block=1; next}
        in_block && $1==key {print $2; exit}
        in_block && $0 ~ /^[^ ]/ {in_block=0}
    ' "${cfg}" 2>/dev/null | sed -e 's/^"//; s/"$//' -e "s/^'//; s/'$//"
}

get_access_address_base_url() {
    local host port path scheme
    host="$(get_access_address_field "host")"
    port="$(get_access_address_field "port")"
    path="$(get_access_address_field "path")"
    scheme="$(get_access_address_field "scheme")"

    if [[ -z "${host}" ]]; then
        return 0
    fi

    scheme="${scheme:-https}"
    path="${path:-/}"
    if [[ "${path}" != /* ]]; then
        path="/${path}"
    fi
    if [[ "${path}" == "/" ]]; then
        path=""
    else
        path="${path%/}"
    fi

    local url="${scheme}://${host}"
    if [[ -n "${port}" ]]; then
        url="${url}:${port}"
    fi
    echo "${url}${path}"
}

get_dip_studio_openclaw_field() {
    local field="$1"
    local cfg="${CONFIG_YAML_PATH}"

    if [[ ! -f "${cfg}" ]]; then
        return 0
    fi

    awk -v key="${field}:" '
        $1=="studio:" {
            in_studio=1
            in_openclaw=0
            next
        }
        in_studio && $1=="openclaw:" {
            in_openclaw=1
            next
        }
        in_studio && in_openclaw && $1==key {
            sub(/^[^:]+:[[:space:]]*/, "", $0)
            print $0
            exit
        }
        in_studio && in_openclaw && $0 ~ /^  [^ ]/ {
            in_openclaw=0
        }
        in_studio && $0 ~ /^[^ ]/ {
            in_studio=0
            in_openclaw=0
        }
    ' "${cfg}" 2>/dev/null | sed -e 's/^"//; s/"$//' -e "s/^'//; s/'$//" | tr -d '\r'
}

random_password() {
    if command -v openssl >/dev/null 2>&1; then
        openssl rand -base64 18 | tr -d '\n'
        return 0
    fi
    head -c 32 /dev/urandom | base64 | tr -d '\n' | head -c 24
}

# Quote a string for YAML single-quoted scalars.
yaml_quote() {
    local s="$1"
    s="${s//\'/\'\'}"
    printf "'%s'" "${s}"
}

get_config_image_registry() {
    local cfg="${CONFIG_YAML_PATH}"
    if [[ ! -f "${cfg}" ]]; then
        return 0
    fi

    awk '
      $1 == "image:" { in_image=1; next }
      in_image && $1 == "registry:" { print $2; exit }
      in_image && $0 ~ /^[^ ]/ { in_image=0 }
    ' "${cfg}" 2>/dev/null | sed -e 's/^["'\'']//; s/["'\'']$//' | tr -d '\r' || true
}

load_image_registry_from_config() {
    if [[ -n "${IMAGE_REGISTRY}" ]]; then
        return 0
    fi
    IMAGE_REGISTRY="$(get_config_image_registry)"
    IMAGE_REGISTRY="${IMAGE_REGISTRY%/}"
    if [[ -z "${IMAGE_REGISTRY}" ]]; then
        IMAGE_REGISTRY="swr.cn-east-3.myhuaweicloud.com/kweaver-ai"
    fi
}

image_from_registry() {
    local repository="$1"
    local tag="$2"
    local fallback="$3"

    load_image_registry_from_config
    if [[ -n "${IMAGE_REGISTRY}" ]]; then
        echo "${IMAGE_REGISTRY}/${repository}:${tag}"
    else
        echo "${fallback}"
    fi
}

get_secret_b64_key() {
    local namespace="$1"
    local name="$2"
    local key="$3"
    local safe_key="${key//\'/\\\'}"
    kubectl -n "${namespace}" get secret "${name}" -o "jsonpath={.data['${safe_key}']}" 2>/dev/null | base64 -d 2>/dev/null || true
}

first_service_with_port() {
    local namespace="$1"
    local selector="$2"
    local port="$3"
    kubectl -n "${namespace}" get svc -l "${selector}" -o jsonpath='{range .items[*]}{.metadata.name}{" "}{range .spec.ports[*]}{.port}{" "}{end}{"\n"}{end}' 2>/dev/null | \
        awk -v want="${port}" '$0 ~ (" " want " ") {print $1; exit}'
}

# Read vendored file if exists; otherwise fetch from URL.
read_or_fetch() {
    local path="$1"
    local url="$2"

    if [[ -n "${path}" && -f "${path}" ]]; then
        cat "${path}"
        return 0
    fi

    if [[ -z "${url}" ]]; then
        log_error "No local file found and no URL provided"
        return 1
    fi

    curl -fsSL "${url}"
}

# Initialize database by connecting to MariaDB pod and executing SQL files
# Usage: init_module_database "module_name" "sql_directory"
# Example: init_module_database "decisionagent" "${SCRIPT_DIR}/scripts/sql/0.5.0/kweaver-core/decisionagent"
init_module_database() {
    local module_name="$1"
    local sql_dir="$2"
    local mariadb_namespace="${MARIADB_NAMESPACE:-resource}"
    
    if [[ -z "${module_name}" || -z "${sql_dir}" ]]; then
        log_error "Usage: init_module_database <module_name> <sql_directory>"
        return 1
    fi
    
    log_info "Initializing ${module_name} database..."
    
    # Check if MariaDB pod is running
    local mariadb_pod=$(kubectl get pods -n "${mariadb_namespace}" -l "app.kubernetes.io/name=proton-mariadb" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    if [[ -z "${mariadb_pod}" ]]; then
        log_error "MariaDB pod not found in namespace ${mariadb_namespace}"
        return 1
    fi
    
    log_info "Found MariaDB pod: ${mariadb_pod}"
    
    # Get MariaDB credentials from config.yaml (under depServices.rds section)
    local mariadb_user=$(grep -A 20 "^  rds:" "${CONFIG_YAML_PATH}" | grep "user:" | head -1 | awk '{print $2}' | tr -d "'\"")
    local mariadb_password=$(grep -A 20 "^  rds:" "${CONFIG_YAML_PATH}" | grep "password:" | head -1 | awk '{print $2}' | tr -d "'\"")

    # Set defaults if not found
    mariadb_user="${mariadb_user:-kweaver}"
    mariadb_password="${mariadb_password:-kweaver}"
    
    log_info "Using MariaDB user: ${mariadb_user}"
    
    # Check if SQL directory exists
    if [[ ! -d "${sql_dir}" ]]; then
        log_error "SQL directory not found: ${sql_dir}"
        return 1
    fi
    
    # Execute all SQL files in the directory in order
    local sql_files=($(find "${sql_dir}" -name "*.sql" -type f | sort))
    if [[ ${#sql_files[@]} -eq 0 ]]; then
        log_error "No SQL files found in ${sql_dir}"
        return 1
    fi
    
    for sql_file in "${sql_files[@]}"; do
        local sql_filename=$(basename "${sql_file}")
        log_info "Executing SQL file: ${sql_filename}"
        
        # Execute SQL in MariaDB pod using cat pipe with mariadb command
        local exec_output
        exec_output=$(cat "${sql_file}" | kubectl exec -i -n "${mariadb_namespace}" "${mariadb_pod}" -- \
            mariadb -u "${mariadb_user}" -p"${mariadb_password}" 2>&1)
        
        if [[ $? -ne 0 ]]; then
            log_error "Failed to execute SQL file ${sql_filename} in MariaDB pod"
            log_error "Error output: ${exec_output}"
            return 1
        fi
        
        log_info "✓ ${sql_filename} executed successfully"
    done
    
    log_info "✓ ${module_name} database initialized successfully"
}

# Create databases without initializing SQL
# Usage: create_databases "database_name1" "database_name2" ...
# Example: create_databases "user_management" "anyshare" "policy_mgnt"
create_databases() {
    local mariadb_namespace="${MARIADB_NAMESPACE:-resource}"
    local db_user=$(grep -A 20 "^  rds:" "${CONFIG_YAML_PATH}" | grep "user:" | head -1 | awk '{print $2}' | tr -d "'\"")
    local root_password=$(grep -A 20 "^  rds:" "${CONFIG_YAML_PATH}" | grep "root_password:" | head -1 | awk '{print $2}' | tr -d "'\"")
    
    # Set defaults if not found
    db_user="${db_user:-adp}"
    root_password="${root_password:-}"
    
    log_info "Creating databases..."
    
    # Check if MariaDB pod is running
    local mariadb_pod=$(kubectl get pods -n "${mariadb_namespace}" -l "app.kubernetes.io/name=proton-mariadb" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    if [[ -z "${mariadb_pod}" ]]; then
        log_error "MariaDB pod not found in namespace ${mariadb_namespace}"
        return 1
    fi
    
    log_info "Found MariaDB pod: ${mariadb_pod}"
    
    # Create each database using root account
    for db_name in "$@"; do
        log_info "Creating database: ${db_name}"
        
        # Create database and grant privileges using root account
        if [[ -n "${root_password}" ]]; then
            kubectl exec -n "${mariadb_namespace}" "${mariadb_pod}" -- mariadb -u root -p"${root_password}" -e "
                CREATE DATABASE IF NOT EXISTS \`${db_name}\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
                GRANT ALL PRIVILEGES ON \`${db_name}\`.* TO '${db_user}'@'%';
                FLUSH PRIVILEGES;
            " 2>/dev/null || log_warn "Failed to create database ${db_name} (may already exist)"
        else
            log_error "root_password not found in config.yaml, cannot create database ${db_name}"
            return 1
        fi
    done
    
    log_info "✓ Databases created successfully"
}

# Show cluster status
show_status() {
    log_info "Cluster Status:"
    echo ""
    kubectl get nodes -o wide
    echo ""
    kubectl get pods -A
}

# =============================================================================
# Release Manifest Database Initialization Detection
# =============================================================================

# Check if a release manifest has a stage="pre" data-migrator release.
# This indicates the manifest handles database initialization via Helm chart.
# Args: <manifest_file>
# Returns: 0 if pre-stage data-migrator found, 1 otherwise
manifest_has_pre_stage_db_init() {
    local manifest_file="$1"

    if [[ -z "${manifest_file}" || ! -f "${manifest_file}" ]]; then
        return 1
    fi

    # Look for any release with stage="pre" and chart name containing "data-migrator"
    local release_name
    for release_name in $(_manifest_list_release_names "${manifest_file}"); do
        local stage
        stage="$(_manifest_strip_quotes "$(_manifest_read_release_field "${manifest_file}" "${release_name}" "stage")")"
        if [[ "${stage}" == "pre" && "${release_name}" == *"data-migrator"* ]]; then
            return 0
        fi
    done

    return 1
}

# Check if the manifest version is 0.6.0 or higher (semver comparison).
# Args: <manifest_file>
# Returns: 0 if version >= 0.6.0, 1 otherwise
manifest_version_gte_060() {
    local manifest_file="$1"

    if [[ -z "${manifest_file}" || ! -f "${manifest_file}" ]]; then
        return 1
    fi

    local version
    version="$(_manifest_strip_quotes "$(_manifest_read_top_level_field "${manifest_file}" "version")")"

    if [[ -z "${version}" ]]; then
        return 1
    fi

    # Extract major.minor.patch
    local major minor
    major="$(echo "${version}" | cut -d. -f1)"
    minor="$(echo "${version}" | cut -d. -f2)"

    # Compare: >= 0.6.0
    if [[ "${major}" -gt 0 ]] || [[ "${major}" -eq 0 && "${minor}" -ge 6 ]]; then
        return 0
    fi

    return 1
}

# Check if database initialization should be skipped for this manifest.
# Returns true (0) if:
#   1. Manifest version is >= 0.6.0 AND
#   2. Manifest has a stage="pre" data-migrator release
# Args: <manifest_file>
# Returns: 0 if DB init should be skipped, 1 otherwise
should_skip_db_init_for_manifest() {
    local manifest_file="$1"

    if [[ -z "${manifest_file}" || ! -f "${manifest_file}" ]]; then
        return 1
    fi

    # Check version >= 0.6.0 and has pre-stage data-migrator
    if manifest_version_gte_060 "${manifest_file}" && manifest_has_pre_stage_db_init "${manifest_file}"; then
        return 0
    fi

    return 1
}

# Read a top-level field from a YAML manifest file.
# Args: <manifest_file> <field_name>
_manifest_read_top_level_field() {
    local manifest_file="$1"
    local field_name="$2"

    awk -v field="${field_name}" '
        BEGIN { in_releases = 0 }
        /^releases:/ || /^dependencies:/ { in_releases = 1; next }
        in_releases && /^[A-Za-z0-9_-]+:/ { in_releases = 0 }
        !in_releases && $1 == field ":" {
            sub(/^[^:]+:[[:space:]]*/, "", $0)
            print $0
            exit
        }
    ' "${manifest_file}" | sed 's/[[:space:]]*$//'
}

# List all release names from a manifest file.
# Args: <manifest_file>
_manifest_list_release_names() {
    local manifest_file="$1"

    awk '
        BEGIN { in_releases = 0 }
        /^releases:/ { in_releases = 1; next }
        in_releases && /^  [A-Za-z0-9_-]+:/ {
            sub(/:.*$/, "", $0)
            sub(/^  /, "", $0)
            print $0
        }
    ' "${manifest_file}"
}
