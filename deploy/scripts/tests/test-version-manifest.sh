#!/bin/bash
set -euo pipefail

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="$(cd "${TEST_DIR}/../.." && pwd)"
EMBEDDED_DIR="${DEPLOY_DIR}/release-manifests"

source "${DEPLOY_DIR}/scripts/lib/common.sh"
source "${DEPLOY_DIR}/scripts/services/isf.sh"
source "${DEPLOY_DIR}/scripts/services/core.sh"
source "${DEPLOY_DIR}/scripts/services/dip.sh"

assert_eq() {
    local actual="$1"
    local expected="$2"
    local message="$3"
    if [[ "${actual}" != "${expected}" ]]; then
        echo "ASSERT_EQ failed: ${message}" >&2
        echo "  expected: ${expected}" >&2
        echo "  actual:   ${actual}" >&2
        exit 1
    fi
}

test_manifest_release_lookup() {
    local actual
    python3() {
        echo "python3 should not be called" >&2
        return 99
    }
    actual="$(get_release_manifest_release_version "${EMBEDDED_DIR}/0.4.0/kweaver-core.yaml" "kweaver-core" "0.4.0" "deploy-web")"
    assert_eq "${actual}" "0.3.0" "core manifest should resolve release version"
}

test_manifest_dependency_lookup() {
    local actual
    python3() {
        echo "python3 should not be called" >&2
        return 99
    }
    actual="$(get_release_manifest_dependency_manifest "${EMBEDDED_DIR}/0.4.0/kweaver-core.yaml" "isf")"
    assert_eq "${actual}" "${EMBEDDED_DIR}/0.4.0/isf.yaml" "core manifest should resolve dependency manifest file"
}

test_core_manifest_release_names() {
    local actual
    actual="$(get_release_manifest_release_names "${EMBEDDED_DIR}/0.4.0/kweaver-core.yaml" "kweaver-core" "0.4.0" | paste -sd ',' -)"
    assert_eq "${actual}" "deploy-web,studio-web,business-system-frontend,business-system-service,mf-model-manager-nginx,mf-model-manager,mf-model-api,bkn-backend,ontology-query,vega-backend,vega-web,data-connection,vega-gateway,vega-gateway-pro,mdl-data-model,mdl-uniquery,mdl-data-model-job,agent-operator-integration,operator-web,agent-retrieval,data-retrieval,agent-backend,agent-web,flow-web,dataflow,coderunner,sandbox" "core manifest should expose the real 0.4.0 release list"
}

test_embedded_manifest_path_lookup() {
    local actual
    actual="$(resolve_embedded_release_manifest "kweaver-core" "0.4.0")"
    assert_eq "${actual}" "${EMBEDDED_DIR}/0.4.0/kweaver-core.yaml" "embedded manifest path should resolve from version/product"
}

test_dip_manifest_direct_isf_dependency_lookup() {
    local actual_manifest
    local actual_version
    DIP_VERSION_MANIFEST_FILE="${EMBEDDED_DIR}/0.4.0/kweaver-dip.yaml"
    actual_manifest="$(_dip_resolve_isf_dependency_manifest)"
    actual_version="$(_dip_resolve_isf_dependency_version)"
    assert_eq "${actual_manifest}" "${EMBEDDED_DIR}/0.4.0/isf.yaml" "dip manifest should expose direct isf dependency manifest"
    assert_eq "${actual_version}" "0.4.0" "dip manifest should expose direct isf dependency version"
}

test_sql_dir_resolution() {
    local actual
    actual="$(resolve_versioned_sql_dir "isf" "0.4.0")"
    assert_eq "${actual}" "${DEPLOY_DIR}/scripts/sql/0.4.0/isf" "sql dir should resolve from version/product"
}

test_core_sql_modules_for_0_4_0() {
    local actual
    actual="$(list_versioned_sql_modules "kweaver-core" "0.4.0" | paste -sd ',' -)"
    assert_eq "${actual}" "agentoperator,dataagent,decisionagent,flowautomation,ontology,sandbox,studio" "core 0.4.0 should only list existing sql modules"
}

test_init_isf_database_uses_versioned_sql_dir() {
    (
        local tmp_dir
        tmp_dir="$(mktemp -d)"
        local calls_file="${tmp_dir}/calls.log"

        is_rds_internal() { return 0; }
        init_module_database() {
            printf '%s|%s\n' "$1" "$2" >> "${calls_file}"
        }

        HELM_CHART_VERSION="0.4.0"
        init_isf_database

        local actual
        actual="$(cat "${calls_file}")"
        assert_eq "${actual}" "isf|${DEPLOY_DIR}/scripts/sql/0.4.0/isf" "isf should initialize versioned sql directory"
        rm -rf "${tmp_dir}"
    )
}

test_init_core_databases_uses_existing_versioned_modules() {
    (
        local tmp_dir
        tmp_dir="$(mktemp -d)"
        local calls_file="${tmp_dir}/calls.log"

        is_rds_internal() { return 0; }
        init_module_database() {
            printf '%s|%s\n' "$1" "$2" >> "${calls_file}"
        }

        HELM_CHART_VERSION="0.4.0"
        init_core_databases

        local actual
        actual="$(paste -sd ';' "${calls_file}")"
        assert_eq "${actual}" "agentoperator|${DEPLOY_DIR}/scripts/sql/0.4.0/kweaver-core/agentoperator;dataagent|${DEPLOY_DIR}/scripts/sql/0.4.0/kweaver-core/dataagent;decisionagent|${DEPLOY_DIR}/scripts/sql/0.4.0/kweaver-core/decisionagent;flowautomation|${DEPLOY_DIR}/scripts/sql/0.4.0/kweaver-core/flowautomation;ontology|${DEPLOY_DIR}/scripts/sql/0.4.0/kweaver-core/ontology;sandbox|${DEPLOY_DIR}/scripts/sql/0.4.0/kweaver-core/sandbox;studio|${DEPLOY_DIR}/scripts/sql/0.4.0/kweaver-core/studio" "core should initialize only versioned sql modules that exist"
        rm -rf "${tmp_dir}"
    )
}

test_init_dip_database_skips_when_sql_missing() {
    (
        local tmp_dir
        tmp_dir="$(mktemp -d)"
        local calls_file="${tmp_dir}/calls.log"

        is_rds_internal() { return 0; }
        init_module_database() {
            printf '%s|%s\n' "$1" "$2" >> "${calls_file}"
        }

        HELM_CHART_VERSION="0.4.0"
        init_dip_database

        local actual=""
        if [[ -f "${calls_file}" ]]; then
            actual="$(cat "${calls_file}")"
        fi
        assert_eq "${actual}" "" "dip should skip sql initialization when versioned sql directory is missing"
        rm -rf "${tmp_dir}"
    )
}

test_download_core_uses_manifest_versions() {
    local tmp_dir
    tmp_dir="$(mktemp -d)"
    local calls_file="${tmp_dir}/calls.log"

    ensure_helm_available() { :; }
    ensure_helm_repo() { :; }
    _core_download_charts_dir() { echo "${tmp_dir}"; }
    download_chart_to_cache() {
        printf '%s|%s|%s|%s\n' "$2" "$3" "$4" "$5" >> "${calls_file}"
    }
    download_isf() {
        printf 'isf|%s|%s\n' "${ISF_VERSION_MANIFEST_FILE:-}" "${HELM_CHART_VERSION:-}" >> "${calls_file}"
    }

    local old_core_releases=("${KWEAVER_CORE_RELEASES[@]}")
    KWEAVER_CORE_RELEASES=("deploy-web" "studio-web")

    HELM_CHART_REPO_NAME="kweaver"
    HELM_CHART_REPO_URL="https://example.invalid/repo"
    HELM_CHART_VERSION="0.4.0"
    CORE_VERSION_MANIFEST_FILE=""
    ENABLE_ISF="true"

    download_core

    mapfile -t calls < "${calls_file}"
    assert_eq "${CORE_VERSION_MANIFEST_FILE}" "${EMBEDDED_DIR}/0.4.0/kweaver-core.yaml" "core download should auto-resolve embedded manifest"
    assert_eq "${calls[0]}" "isf|${EMBEDDED_DIR}/0.4.0/isf.yaml|0.4.0" "core download should pass dependency manifest to isf"
    assert_eq "${calls[1]}" "kweaver|deploy-web|0.3.0|false" "core download should use manifest release version for deploy-web"
    assert_eq "${calls[2]}" "kweaver|studio-web|0.3.0|false" "core download should use manifest release version for studio-web"

    KWEAVER_CORE_RELEASES=("${old_core_releases[@]}")
    rm -rf "${tmp_dir}"
}

test_download_dip_auto_resolves_embedded_manifests() {
    local tmp_dir
    tmp_dir="$(mktemp -d)"
    local calls_file="${tmp_dir}/calls.log"

    ensure_helm_available() { :; }
    ensure_helm_repo() { :; }
    _dip_download_charts_dir() { echo "${tmp_dir}"; }
    download_core() {
        printf 'core|%s|%s\n' "${CORE_VERSION_MANIFEST_FILE:-}" "${HELM_CHART_VERSION:-}" >> "${calls_file}"
    }
    download_chart_to_cache() {
        printf 'dip|%s|%s|%s|%s\n' "$2" "$3" "$4" "$5" >> "${calls_file}"
    }

    local old_dip_prereleases=("${DIP_PRERELEASES[@]}")
    local old_dip_releases=("${DIP_RELEASES[@]}")
    DIP_PRERELEASES=("dip-data-migrator")
    DIP_RELEASES=("data-catalog")

    HELM_CHART_REPO_NAME="kweaver"
    HELM_CHART_REPO_URL="https://example.invalid/repo"
    HELM_CHART_VERSION="0.4.0"
    DIP_VERSION_MANIFEST_FILE=""

    download_dip

    mapfile -t calls < "${calls_file}"
    assert_eq "${DIP_VERSION_MANIFEST_FILE}" "${EMBEDDED_DIR}/0.4.0/kweaver-dip.yaml" "dip download should auto-resolve embedded manifest"
    assert_eq "${calls[0]}" "core|${EMBEDDED_DIR}/0.4.0/kweaver-core.yaml|0.4.0" "dip download should pass embedded core manifest to core download"
    assert_eq "${calls[1]}" "dip|kweaver|dip-data-migrator|0.4.0-main.20260327.5.fac488b|false" "dip prerelease should use embedded manifest version"
    assert_eq "${calls[2]}" "dip|kweaver|data-catalog|0.4.0-main.20260327.1|false" "dip release should use embedded manifest version"

    DIP_PRERELEASES=("${old_dip_prereleases[@]}")
    DIP_RELEASES=("${old_dip_releases[@]}")
    rm -rf "${tmp_dir}"
}

test_manifest_release_lookup
test_manifest_dependency_lookup
test_core_manifest_release_names
test_embedded_manifest_path_lookup
test_dip_manifest_direct_isf_dependency_lookup
test_sql_dir_resolution
test_core_sql_modules_for_0_4_0
test_init_isf_database_uses_versioned_sql_dir
test_init_core_databases_uses_existing_versioned_modules
test_init_dip_database_skips_when_sql_missing
test_download_core_uses_manifest_versions
test_download_dip_auto_resolves_embedded_manifests

echo "PASS"
