#!/usr/bin/env bash
# Context Loader toolset (ADP) import — run from onboard.sh with kweaver CLI (kweaver-sdk) auth.
# Replaces deploy-time curl/port-forward; uses: kweaver call ... -F (multipart), same as manual impex.
#
# Who can import (token from  kweaver auth login  , stored under ~/.kweaver):
#   - Full install (ISF): create user  test  first (onboard offer / kweaver-admin), then this import. Uses
#      test  (role list roles) and  kweaver auth  as  test ; console  admin  often gets HTTP 403 on impex.
#   - Minimum (no ISF):  kweaver auth login  only; kweaver-admin is not required.
# shellcheck source=/dev/null
# Report: ONBOARD_REPORT_CONTEXT_LOADER (onboard_report.sh)

# Resolve path to default ADP in repo (override: CONTEXT_LOADER_TOOLSET_ADP_PATH).
onboard_context_loader_adp_path() {
    local repo_root
    repo_root="$(cd "${SCRIPT_DIR}/.." && pwd)"
    printf '%s' "${CONTEXT_LOADER_TOOLSET_ADP_PATH:-${repo_root}/adp/context-loader/agent-retrieval/docs/release/toolset/context_loader_toolset.adp}"
}

# Extract box_id / box_name from a toolbox ADP file (JSON: toolbox.configs[0]).
# Prints "<box_id>\t<box_name>" on stdout; empty and non-zero on failure.
onboard_context_loader_adp_meta() {
    local adp="$1"
    [[ -f "${adp}" ]] || return 1
    python3 - "${adp}" <<'PY' 2>/dev/null || return 1
import json, sys
try:
    with open(sys.argv[1], "r", encoding="utf-8") as f:
        j = json.load(f)
except Exception:
    sys.exit(1)
configs = (j.get("toolbox") or {}).get("configs") or []
if not configs:
    sys.exit(1)
bid = configs[0].get("box_id") or ""
bname = configs[0].get("box_name") or ""
if not bid:
    sys.exit(1)
print("%s\t%s" % (bid, bname))
PY
}

# Returns 0 if the toolbox box_id is already present on the platform, else 1.
# Uses the same -bd as the importer so the check sees the right business domain.
onboard_context_loader_already_imported() {
    local box_id="$1" bd
    bd="${DEPLOY_BUSINESS_DOMAIN:-bd_public}"
    [[ -z "${box_id}" ]] && return 1
    local out
    if ! out=$(kweaver call "/api/agent-operator-integration/v1/tool-box/${box_id}" -bd "${bd}" 2>/dev/null); then
        return 1
    fi
    # success payload contains the box_id; failures look like {"code":"...","message":"..."}
    # or HTTP 404 wrapped JSON. Be conservative: require the box_id literal AND no "not found" indicator.
    if printf '%s' "${out}" | grep -Fq "${box_id}" \
        && ! printf '%s' "${out}" | grep -qiE '"code"[[:space:]]*:[[:space:]]*"?(4[0-9]{2}|5[0-9]{2})|not[[:space:]]*found|不存在|does not exist'; then
        return 0
    fi
    return 1
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
    # Skip if the toolbox is already on the platform. With -y we never re-import (use --reimport-context-loader=true to force).
    # In TTY mode we ask whether to overwrite via the same upsert endpoint.
    local _adp_meta _adp_box_id _adp_box_name
    if _adp_meta="$(onboard_context_loader_adp_meta "${adp}")" && [[ -n "${_adp_meta}" ]]; then
        _adp_box_id="${_adp_meta%%	*}"
        _adp_box_name="${_adp_meta#*	}"
        if onboard_context_loader_already_imported "${_adp_box_id}"; then
            local _force="${REIMPORT_CONTEXT_LOADER:-false}"
            if [[ "${_force}" != "true" && "${ONBOARD_ASSUME_YES:-false}" == "true" ]]; then
                ONBOARD_REPORT_CONTEXT_LOADER="skipped: already imported (box_id=${_adp_box_id}${_adp_box_name:+, name=${_adp_box_name}}); set REIMPORT_CONTEXT_LOADER=true to overwrite"
                log_info "Context Loader: already imported on this platform (box_id=${_adp_box_id}${_adp_box_name:+, name=${_adp_box_name}}). Skipping. Set REIMPORT_CONTEXT_LOADER=true to force re-import."
                return 0
            fi
            local _is_tty="false"
            if type onboard_is_bootstrap_tty &>/dev/null && onboard_is_bootstrap_tty; then
                _is_tty="true"
            fi
            if [[ "${_force}" != "true" && "${_is_tty}" == "true" ]]; then
                read -r -p "Context Loader is already imported (box_id=${_adp_box_id}${_adp_box_name:+, name=${_adp_box_name}}). Re-import (upsert) anyway? [y/N]: " _reimp
                if [[ ! "${_reimp}" =~ ^[Yy] ]]; then
                    ONBOARD_REPORT_CONTEXT_LOADER="skipped: already imported (box_id=${_adp_box_id}); user declined re-import"
                    log_info "Context Loader: skipped (already imported, user declined re-import)."
                    return 0
                fi
            fi
            log_info "Context Loader: already imported — proceeding with upsert (force=${_force})."
        fi
    else
        log_info "Context Loader: could not extract box_id from ADP — will import unconditionally."
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
# ISF: user [test] must exist before ADP impex (admin kweaver token often gets CommonAddForbidden/403).
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
    if type onboard_prepend_npm_global_bin_to_path &>/dev/null; then
        onboard_prepend_npm_global_bin_to_path
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

    # ISF: create user [test] first, then this step (import uses kweaver as test, not admin).
    if type onboard_isf_full_install &>/dev/null && onboard_isf_full_install 2>/dev/null; then
        if ! command -v kweaver-admin &>/dev/null; then
            ONBOARD_REPORT_CONTEXT_LOADER="error: ISF needs kweaver-admin on PATH for ADP impex"
            onboard_log_err "Context Loader: ISF — kweaver-admin not on PATH. Install it and re-run onboard (earlier steps should not reach here)."
            exit 1
        fi
        if ! kweaver-admin --json user list --limit 1 &>/dev/null; then
            ONBOARD_REPORT_CONTEXT_LOADER="error: kweaver-admin not authenticated"
            onboard_log_err "Context Loader: ISF — kweaver-admin is not signed in. Sign in, then re-run: $0"
            exit 1
        fi
        if type onboard_user_test_exists &>/dev/null && ! onboard_user_test_exists; then
            ONBOARD_REPORT_CONTEXT_LOADER="error: ISF requires user [test] before toolbox import"
            onboard_log_err "Context Loader: ISF — user [test] is missing. Create with kweaver-admin (or re-run onboard without --skip-isf-test-user), then import."
            exit 1
        fi
    fi

    if [[ "${ONBOARD_ASSUME_YES}" == "true" ]]; then
        log_info "Context Loader: importing toolset (-y)…"
        onboard_context_loader_import_via_kweaver
        return 0
    fi
    if ! (type onboard_is_bootstrap_tty &>/dev/null && onboard_is_bootstrap_tty); then
        ONBOARD_REPORT_CONTEXT_LOADER="skipped: not a TTY, no auto-import (use -y or import manually)"
        log_info "Context Loader: not a TTY — skipping import. Re-run: ./onboard.sh -y  or run manually (see help)."
        return 0
    fi
    echo ""
    read -r -p "Import Context Loader (ADP) now? (ISF: user test must already exist; kweaver will sign in as test for impex) [Y/n]: " _clm
    if [[ "${_clm}" =~ ^[Nn] ]]; then
        ONBOARD_REPORT_CONTEXT_LOADER="skipped: user declined import (run kweaver call impex later)"
        log_info "Skipped. Manual: kweaver call '/api/agent-operator-integration/v1/impex/import/toolbox' -X POST -F data=@'${adp}' -F mode=upsert -bd ${DEPLOY_BUSINESS_DOMAIN:-bd_public}"
        return 0
    fi
    onboard_context_loader_import_via_kweaver
}
