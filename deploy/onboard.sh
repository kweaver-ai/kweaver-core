#!/usr/bin/env bash
# KWeaver Core — onboard: register models, BKN, rollout (run after `deploy.sh` install)
# Requires: kweaver, kubectl, python3, PyYAML (pip3 install pyyaml) for --config; interactive is lighter.
# Run from the deploy/ directory (symmetric with preflight.sh).

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/lib/common.sh
source "${SCRIPT_DIR}/scripts/lib/common.sh"
# shellcheck source=scripts/lib/onboard_models.sh
source "${SCRIPT_DIR}/scripts/lib/onboard_models.sh"

NAMESPACE="${NAMESPACE:-kweaver}"
CONFIG_FILE=""
BKN_NAME=""
ENABLE_BKN_ONLY="false"
SKIP_BKN="false"
INTERACTIVE="true"

usage() {
    echo "Usage: $0 [options]"
    echo "  --config=PATH            YAML: deploy/conf/models.yaml.example (needs PyYAML)"
    echo "  --namespace=NS           (default: kweaver; or key 'namespace' in yaml)"
    echo "  --enable-bkn-search      Only patch bkn/ontology ConfigMaps and rollout"
    echo "  --bkn-embedding-name=X   Required with --enable-bkn-search (registered model_name)"
    echo "  --skip-bkn               With --config: register models but skip BKN + rollout"
    echo "  -h, --help"
}

if ! command -v kweaver &>/dev/null; then
    log_error "kweaver not found. Install: npm i -g @kweaver-ai/kweaver-sdk (needs Node 22+; npm may warn on older Node)"
    log_error "  sudo: on many Linux hosts the global prefix is under /usr → use sudo npm i -g @kweaver-ai/kweaver-sdk, or get EACCES without it."
    log_error "  no sudo: use nvm/fnm, or npm config set prefix \"\$HOME/.local\" and put ~/.local/bin on PATH, then npm i -g without sudo."
    exit 1
fi
if ! command -v kubectl &>/dev/null; then
    log_error "kubectl not found"
    exit 1
fi
if ! command -v python3 &>/dev/null; then
    log_error "python3 not found"
    exit 1
fi

while [[ $# -gt 0 ]]; do
    case "$1" in
        -h | --help)
            usage
            exit 0
            ;;
        --config=*)
            CONFIG_FILE="${1#*=}"
            INTERACTIVE="false"
            shift
            ;;
        --namespace=*)
            NAMESPACE="${1#*=}"
            shift
            ;;
        --bkn-embedding-name=*)
            BKN_NAME="${1#*=}"
            shift
            ;;
        --enable-bkn-search) ENABLE_BKN_ONLY="true"; shift ;;
        --skip-bkn)          SKIP_BKN="true"; shift ;;
        *)
            log_error "Unknown: $1"
            usage
            exit 1
            ;;
    esac
done

onboard_probe() {
    if ! kweaver bkn list &>/dev/null; then
        onboard_log_err "kweaver bkn list failed. Run: kweaver auth login <https://access-url> -k"
        exit 1
    fi
    if ! kubectl get "namespace/${NAMESPACE}" &>/dev/null; then
        onboard_log_err "kubectl: namespace ${NAMESPACE} not found (export NAMESPACE=...)"
        exit 1
    fi
    onboard_log_info "OK: kweaver + kubectl (ns=${NAMESPACE})"
    onboard_recommend_admin_cli
}

# Detect ISF (full install) and recommend kweaver-admin when present.
onboard_recommend_admin_cli() {
    local has_isf="false" isf_releases=""
    if command -v helm &>/dev/null; then
        isf_releases="$(helm list -A 2>/dev/null \
            | awk 'NR>1 {print $1}' \
            | grep -E '^(authentication|hydra|user-management|eacp|isfweb|isf-data-migrator|policy-management|audit-log|authorization|sharemgnt|oauth2-ui|ingress-informationsecurityfabric)$' \
            | paste -sd ',' - || true)"
        [[ -n "${isf_releases}" ]] && has_isf="true"
    fi
    if [[ "${has_isf}" != "true" ]]; then
        if kubectl get ns 2>/dev/null | awk '{print $1}' | grep -qiE '^(isf|information-security-fabric)$'; then
            has_isf="true"
        fi
    fi

    if [[ "${has_isf}" == "true" ]]; then
        onboard_log_info "Detected ISF (full install) on this cluster${isf_releases:+ — releases: ${isf_releases}}"
        if command -v kweaver-admin &>/dev/null; then
            onboard_log_info "kweaver-admin: $(kweaver-admin --version 2>/dev/null | head -n1)"
        else
            onboard_log_warn "kweaver-admin not in PATH. Full install needs it for user/org/role/model/audit ops."
            onboard_log_warn "  Install:  npm i -g @kweaver-ai/kweaver-admin"
            onboard_log_warn "  Login:    kweaver-admin auth login <https://access-url> -k"
        fi
    else
        onboard_log_info "No ISF releases detected — minimum install. kweaver-sdk (this CLI) is enough; kweaver-admin not required."
    fi
}

onboard_do_bkn_bash() {
    local emb_name="$1"
    onboard_upsert_cm_embedded_yaml "${NAMESPACE}" "bkn-backend-cm" "${emb_name}" || return 1
    onboard_upsert_cm_embedded_yaml "${NAMESPACE}" "ontology-query-cm" "${emb_name}" || return 1
    onboard_bkn_rollout "${NAMESPACE}" || return 1
}

# ---- main --------------------------------------------------------------------
if [[ "${ENABLE_BKN_ONLY}" == "true" ]]; then
    if [[ -z "${BKN_NAME}" ]]; then
        onboard_log_err "Use --bkn-embedding-name=<model_name> with --enable-bkn-search"
        exit 1
    fi
    # Prefer Python (same as full config) if PyYAML available; else bash+yq
    onboard_probe
    export KWE_POST_NS="${NAMESPACE}" KWE_POST_BKN="${BKN_NAME}"
    if python3 -c "import yaml" 2>/dev/null && PYTHONPATH="${SCRIPT_DIR}/scripts/lib" python3 -c "import onboard_apply_config" 2>/dev/null; then
        PYTHONPATH="${SCRIPT_DIR}/scripts/lib" python3 -c "
import sys
from onboard_apply_config import patch_bkn_cms_and_rollout
import os
sys.exit(patch_bkn_cms_and_rollout(os.environ['KWE_POST_NS'], os.environ['KWE_POST_BKN']))
"
    else
        onboard_log_warn "PyYAML or module missing; using bash path (needs PyYAML in onboard_upsert)"
        onboard_do_bkn_bash "${BKN_NAME}"
    fi
    onboard_log_info "Done (BKN only)."
    exit 0
fi

onboard_probe

if [[ -n "${CONFIG_FILE}" ]]; then
    if [[ ! -f "${CONFIG_FILE}" ]]; then
        onboard_log_err "Config not found: ${CONFIG_FILE}"
        exit 1
    fi
    if ! python3 -c "import yaml" 2>/dev/null; then
        onboard_log_err "For --config, install PyYAML: pip3 install pyyaml"
        exit 1
    fi
    exec python3 "${SCRIPT_DIR}/scripts/lib/onboard_apply_config.py" \
        "${CONFIG_FILE}" \
        "${NAMESPACE}" \
        "${SKIP_BKN}"
fi

# Interactive (bash path for registration; BKN uses upsert in models.sh)
if [[ "${INTERACTIVE}" == "true" ]]; then
    if ! python3 -c "import yaml" 2>/dev/null; then
        onboard_log_warn "PyYAML not installed: BKN ConfigMap patch will fail. pip3 install pyyaml"
    fi
    onboard_log_info "Interactive model registration (empty line skips section)"
    read -r -p "Namespace [${NAMESPACE}]: " ns
    [[ -n "${ns}" ]] && NAMESPACE="${ns}"
    if ! kubectl get "namespace/${NAMESPACE}" &>/dev/null; then
        onboard_log_err "namespace ${NAMESPACE} not found"
        exit 1
    fi

    llm_n=""
    read -r -p "LLM model_name: " llm_n
    if [[ -n "${llm_n}" ]]; then
        read -r -p "LLM model_series (e.g. deepseek) [others]: " llm_s
        read -r -p "max_model_len [8192]: " llm_ml
        read -r -p "api_key: " -s llm_key
        echo
        read -r -p "api_model: " llm_am
        read -r -p "api_url: " llm_url
        _POSTI_EXISTING_LLM="$(onboard_get_existing_llm_names)"
        _POSTI_EXISTING_SM="$(onboard_get_existing_small_model_names)"
        onboard_ensure_llm "${llm_n}" "${llm_s:-others}" "${llm_ml:-8192}" "${llm_key}" "${llm_am}" "${llm_url}" "llm"
    fi
    _POSTI_EXISTING_LLM="$(onboard_get_existing_llm_names)"
    _POSTI_EXISTING_SM="$(onboard_get_existing_small_model_names)"

    em_n=""
    bkn_default_name=""
    read -r -p "Embedding model_name: " em_n
    if [[ -n "${em_n}" ]]; then
        read -r -p "api_key: " -s em_key
        echo
        read -r -p "api_model: " em_am
        read -r -p "api_url: " em_url
        read -r -p "embedding_dim [1024]: " em_dim
        read -r -p "Set as BKN default? [Y/n]: " em_bkn
        onboard_ensure_small_model "${em_n}" "embedding" "${em_key}" "${em_url}" "${em_am}" 32 512 "${em_dim:-1024}"
        if [[ ! "${em_bkn}" =~ ^[Nn] ]]; then
            bkn_default_name="${em_n}"
        fi
    fi
    if [[ -n "${llm_n:-}" ]]; then
        onboard_test_llm "$(onboard_get_id_for_llm "${llm_n}")"
    fi
    if [[ -n "${em_n:-}" ]]; then
        onboard_test_small "$(onboard_get_id_for_small "${em_n}")"
    fi
    if [[ "${SKIP_BKN}" == "true" ]]; then
        onboard_log_info "Done (skip BKN not used in interactive; omit model to skip BKN patch)."
    elif [[ -n "${bkn_default_name}" ]]; then
        read -r -p "Patch BKN ConfigMaps and restart bkn-backend / ontology-query? [Y/n]: " do_b
        if [[ ! "${do_b}" =~ ^[Nn] ]]; then
            onboard_do_bkn_bash "${bkn_default_name}" || exit 1
        fi
    else
        onboard_log_info "No BKN default embedding; skipped ConfigMap."
    fi
    onboard_log_info "Done."
    exit 0
fi

usage
exit 1
