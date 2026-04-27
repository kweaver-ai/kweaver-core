#!/usr/bin/env bash
# Context Loader toolset (ADP) import — run from onboard.sh with kweaver CLI (kweaver-sdk) auth.
# Replaces deploy-time curl/port-forward; uses: kweaver call ... -F (multipart), same as manual impex.
#
# Who can import (token from  kweaver auth login  , stored under ~/.kweaver):
#   - Full install (ISF): prefer user  test  with all roles from  kweaver-admin role list  (typically
#     three business admin roles). Console  admin  is often not allowed to add ADP resources (HTTP 403).
#     This script syncs role list to  test  and re-runs  kweaver auth login  as  test  before impex.
#   - Minimum (no ISF):  kweaver auth login  only; kweaver-admin is not required.
# shellcheck source=/dev/null
# Report: ONBOARD_REPORT_CONTEXT_LOADER (onboard_report.sh)

# Resolve path to default ADP in repo (override: CONTEXT_LOADER_TOOLSET_ADP_PATH).
onboard_context_loader_adp_path() {
    local repo_root
    repo_root="$(cd "${SCRIPT_DIR}/.." && pwd)"
    printf '%s' "${CONTEXT_LOADER_TOOLSET_ADP_PATH:-${repo_root}/adp/context-loader/agent-retrieval/docs/release/toolset/context_loader_toolset.adp}"
}

# Import via kweaver call (uses ~/.kweaver token from kweaver auth login). Needs permission to impex.
onboard_context_loader_import_via_kweaver() {
    local ns adp bd
    ns="${NAMESPACE:-kweaver}"
    adp="$(onboard_context_loader_adp_path)"
    bd="${DEPLOY_BUSINESS_DOMAIN:-bd_public}"
    if ! (type onboard_isf_full_install &>/dev/null && onboard_isf_full_install 2>/dev/null); then
        log_info "Context Loader: minimum (no ISF) —  kweaver auth login  is enough; kweaver-admin not required."
    fi
    if [[ ! -f "${adp}" ]]; then
        ONBOARD_REPORT_CONTEXT_LOADER="skipped: ADP file missing"
        log_warn "Context Loader: ADP not found: ${adp}"
        return 1
    fi
    if ! kubectl get deploy agent-operator-integration -n "${ns}" &>/dev/null; then
        ONBOARD_REPORT_CONTEXT_LOADER="skipped: no agent-operator-integration deployment (${ns})"
        log_warn "Context Loader: no deploy/agent-operator-integration in ${ns}; skip import."
        return 1
    fi
    log_info "Context Loader: waiting for agent-operator-integration (up to 120s)…"
    kubectl rollout status deploy/agent-operator-integration -n "${ns}" --timeout=120s &>/dev/null || true
    if type onboard_isf_full_install &>/dev/null && onboard_isf_full_install 2>/dev/null; then
        if type onboard_ensure_isf_test_for_kweaver_impex &>/dev/null; then
            if ! onboard_ensure_isf_test_for_kweaver_impex; then
                ONBOARD_REPORT_CONTEXT_LOADER="import_failed: ISF relogin as test / kweaver auth failed (see logs above)"
                return 1
            fi
        else
            log_info "Context Loader: ISF — (onboard_ensure_isf_test_for_kweaver_impex missing) using current kweaver user."
        fi
    fi
    log_info "Importing Context Loader toolset via kweaver call (impex upsert)…"
    if kweaver call "/api/agent-operator-integration/v1/impex/import/toolbox" -X POST \
        -F "data=@${adp}" \
        -F "mode=upsert" \
        -bd "${bd}" --pretty; then
        ONBOARD_REPORT_CONTEXT_LOADER="Imported: imported_ok — confirm above HTTP 2xx and no error in body"
        log_info "Context Loader toolset import finished (check output above for HTTP errors in body)."
        return 0
    fi
    ONBOARD_REPORT_CONTEXT_LOADER="Import failed: kweaver call non-zero or HTTP 4xx/5xx in logs above"
    log_warn "Context Loader: kweaver call failed. Re-login:  kweaver auth login <url> -k"
    if type onboard_isf_full_install &>/dev/null && onboard_isf_full_install 2>/dev/null; then
        log_warn "  (ISF) ensure user  test  has all  role list  roles and: kweaver auth login <url> -u test -p ... --http-signin -k"
        log_warn "  (or set ONBOARD_TEST_USER_PASSWORD; built-in  admin  often cannot import this ADP path)"
    else
        log_warn "  (minimum) ensure a valid  kweaver auth login  for this cluster."
    fi
    log_warn "  Or import manually. ADP: ${adp}"
    return 1
}

# After kweaver auth; optional Y/n; -y to auto-run. Skips if no operator or ADP.
onboard_offer_context_loader_toolset() {
    if [[ "${ONBOARD_SKIP_CONTEXT_LOADER:-false}" == "true" ]]; then
        ONBOARD_REPORT_CONTEXT_LOADER="skipped: ONBOARD_SKIP_CONTEXT_LOADER"
        return 0
    fi
    if [[ "${IMPORT_CONTEXT_LOADER_TOOLSET:-true}" == "false" ]]; then
        ONBOARD_REPORT_CONTEXT_LOADER="skipped: IMPORT_CONTEXT_LOADER_TOOLSET=false"
        return 0
    fi
    if ! command -v kweaver &>/dev/null; then
        ONBOARD_REPORT_CONTEXT_LOADER="skipped: kweaver CLI not found"
        return 0
    fi
    local ns
    ns="${NAMESPACE:-kweaver}"
    if ! kubectl get deploy agent-operator-integration -n "${ns}" &>/dev/null; then
        ONBOARD_REPORT_CONTEXT_LOADER="skipped: no agent-operator deployment (namespace=${ns})"
        return 0
    fi
    local adp
    adp="$(onboard_context_loader_adp_path)"
    if [[ ! -f "${adp}" ]]; then
        ONBOARD_REPORT_CONTEXT_LOADER="skipped: ADP not found (clone repo or set CONTEXT_LOADER_TOOLSET_ADP_PATH)"
        log_warn "Context Loader: ADP missing (${adp}). Clone repo or set CONTEXT_LOADER_TOOLSET_ADP_PATH."
        return 0
    fi

    if [[ "${ONBOARD_ASSUME_YES}" == "true" ]]; then
        log_info "Context Loader: importing toolset (-y)…"
        onboard_context_loader_import_via_kweaver || true
        return 0
    fi
    if ! (type onboard_is_bootstrap_tty &>/dev/null && onboard_is_bootstrap_tty); then
        ONBOARD_REPORT_CONTEXT_LOADER="skipped: not a TTY, no auto-import (use -y or import manually)"
        log_info "Context Loader: not a TTY — skipping import. Re-run: ./onboard.sh -y  or run manually (see help)."
        return 0
    fi
    echo ""
    read -r -p "Import Context Loader (ADP) now (ISF: will re-login  kweaver  as user test for impex) [Y/n]: " _clm
    if [[ "${_clm}" =~ ^[Nn] ]]; then
        ONBOARD_REPORT_CONTEXT_LOADER="skipped: user declined import (run kweaver call impex later)"
        log_info "Skipped. Manual: kweaver call '/api/agent-operator-integration/v1/impex/import/toolbox' -X POST -F data=@'${adp}' -F mode=upsert -bd ${DEPLOY_BUSINESS_DOMAIN:-bd_public}"
        return 0
    fi
    onboard_context_loader_import_via_kweaver || true
}
