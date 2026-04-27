#!/usr/bin/env bash
# Context Loader toolset (ADP) import — run from onboard.sh with kweaver CLI (kweaver-sdk) auth.
# Replaces deploy-time curl/port-forward; uses: kweaver call ... -F (multipart), same as manual impex.
#
# Who can import (token from  kweaver auth login  , stored under ~/.kweaver):
#   - Full install (ISF present): use a platform admin. Default admin login is often
#     admin@eisoo.com / eisoo.com (same as deploy/auto_cofig/README.md console: admin / eisoo.com).
#   - Minimum (no ISF, kweaver-core --minimum): only kweaver-sdk is needed;  kweaver auth login
#     then  kweaver call  impex; kweaver-admin is not required.
# shellcheck source=/dev/null

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
    if type onboard_isf_full_install &>/dev/null && onboard_isf_full_install 2>/dev/null; then
        log_info "Context Loader: full/ISF — ensure  kweaver auth login  uses a platform admin (e.g. admin@eisoo.com / eisoo.com; see deploy/auto_cofig/README.md)."
    else
        log_info "Context Loader: minimum (no ISF) —  kweaver auth login  is enough; kweaver-admin not required."
    fi
    if [[ ! -f "${adp}" ]]; then
        log_warn "Context Loader: ADP not found: ${adp}"
        return 1
    fi
    if ! kubectl get deploy agent-operator-integration -n "${ns}" &>/dev/null; then
        log_warn "Context Loader: no deploy/agent-operator-integration in ${ns}; skip import."
        return 1
    fi
    log_info "Context Loader: waiting for agent-operator-integration (up to 120s)…"
    kubectl rollout status deploy/agent-operator-integration -n "${ns}" --timeout=120s &>/dev/null || true
    log_info "Importing Context Loader toolset via kweaver call (impex upsert)…"
    if kweaver call "/api/agent-operator-integration/v1/impex/import/toolbox" -X POST \
        -F "data=@${adp}" \
        -F "mode=upsert" \
        -bd "${bd}" --pretty; then
        log_info "Context Loader toolset import finished (check output above for HTTP errors in body)."
        return 0
    fi
    log_warn "Context Loader: kweaver call failed. Re-login:  kweaver auth login <url> -k"
    if type onboard_isf_full_install &>/dev/null && onboard_isf_full_install 2>/dev/null; then
        log_warn "  (ISF/full) use a platform admin token, e.g. admin@eisoo.com / eisoo.com (see deploy/auto_cofig/README.md)."
    else
        log_warn "  (minimum) ensure a valid  kweaver auth login  for this cluster."
    fi
    log_warn "  Or import manually. ADP: ${adp}"
    return 1
}

# After kweaver auth; optional Y/n; -y to auto-run. Skips if no operator or ADP.
onboard_offer_context_loader_toolset() {
    if [[ "${ONBOARD_SKIP_CONTEXT_LOADER:-false}" == "true" ]]; then
        return 0
    fi
    if [[ "${IMPORT_CONTEXT_LOADER_TOOLSET:-true}" == "false" ]]; then
        return 0
    fi
    if ! command -v kweaver &>/dev/null; then
        return 0
    fi
    local ns
    ns="${NAMESPACE:-kweaver}"
    if ! kubectl get deploy agent-operator-integration -n "${ns}" &>/dev/null; then
        return 0
    fi
    local adp
    adp="$(onboard_context_loader_adp_path)"
    if [[ ! -f "${adp}" ]]; then
        log_warn "Context Loader: ADP missing (${adp}). Clone repo or set CONTEXT_LOADER_TOOLSET_ADP_PATH."
        return 0
    fi

    if [[ "${ONBOARD_ASSUME_YES}" == "true" ]]; then
        log_info "Context Loader: importing toolset (-y)…"
        onboard_context_loader_import_via_kweaver || true
        return 0
    fi
    if ! onboard_is_bootstrap_tty; then
        log_info "Context Loader: not a TTY — skipping import. Re-run: ./onboard.sh -y  or run manually (see help)."
        return 0
    fi
    echo ""
    read -r -p "Import Context Loader toolset (ADP) now via kweaver (uses your kweaver login)? [Y/n]: " _clm
    if [[ "${_clm}" =~ ^[Nn] ]]; then
        log_info "Skipped. Manual: kweaver call '/api/agent-operator-integration/v1/impex/import/toolbox' -X POST -F data=@'${adp}' -F mode=upsert -bd ${DEPLOY_BUSINESS_DOMAIN:-bd_public}"
        return 0
    fi
    onboard_context_loader_import_via_kweaver || true
}
