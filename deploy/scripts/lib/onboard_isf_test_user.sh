#!/usr/bin/env bash
# Optional: first business test user (login: test) with all platform roles (ISF full install + kweaver-admin).
# shellcheck source=/dev/null

# Match onboard.sh: ISF (full) install present.
onboard_isf_full_install() {
    local has_isf="false"
    if command -v helm &>/dev/null; then
        if helm list -A 2>/dev/null \
            | awk 'NR>1 {print $1}' \
            | grep -qE '^(authentication|hydra|user-management|eacp|isfweb|isf-data-migrator|policy-management|audit-log|authorization|sharemgnt|oauth2-ui|ingress-informationsecurityfabric)$'; then
            has_isf="true"
        fi
    fi
    if [[ "${has_isf}" != "true" ]] && kubectl get ns 2>/dev/null | awk '{print $1}' | grep -qiE '^(isf|information-security-fabric)$'; then
        has_isf="true"
    fi
    [[ "${has_isf}" == "true" ]]
}

# True if a user with account (login) "test" already exists.
onboard_user_test_exists() {
    local _js
    if ! _js="$(kweaver-admin --json user list --keyword test --limit 200 2>/dev/null)"; then
        return 1
    fi
    echo "${_js}" | python3 -c "
import json, sys
try:
    d = json.load(sys.stdin)
except Exception:
    sys.exit(1)
for e in d.get('entries', []):
    if (e.get('account') or '') == 'test':
        sys.exit(0)
sys.exit(1)" || return 1
    return 0
}

# Create login=test and assign every role from role list (POC; production should use least privilege).
onboard_create_test_user_with_all_roles() {
    log_info "Creating user test (default password 123456, change on first login)…"
    local uerr
    uerr="$(mktemp 2>/dev/null || echo /tmp/onboard-uc.$$)"
    if ! kweaver-admin --json user create --login test >/dev/null 2> "${uerr}"; then
        if grep -qiE 'already|exists|exist|重复|已存在' "${uerr}" 2>/dev/null; then
            log_info "User test may already exist; continuing to role assignment…"
        else
            log_warn "kweaver-admin user create failed: $(tr '\n' ' ' < "${uerr}" | head -c 400)"
            rm -f "${uerr}"
            return 1
        fi
    fi
    rm -f "${uerr}" 2>/dev/null || true
    # Collect role ids
    local rjson rids _rid _n _fail
    if ! rjson="$(kweaver-admin --json role list --limit 1000 2>/dev/null)"; then
        log_warn "kweaver-admin role list failed; cannot assign roles to test"
        return 1
    fi
    rids="$(echo "${rjson}" | python3 -c "
import json, sys
try:
    d = json.load(sys.stdin)
except Exception:
    sys.exit(1)
for e in d.get('entries', []):
    i = e.get('id') or e.get('Id')
    if i:
        print(i)
" 2>/dev/null)" || true
    if [[ -z "${rids// }" ]]; then
        log_warn "No roles in role list; test user was created (if new) but has no extra roles"
        return 0
    fi
    _fail=0
    _n=0
    _ok=0
    while IFS= read -r _rid; do
        [[ -n "${_rid}" ]] || continue
        _n=$((_n + 1))
        if kweaver-admin user assign-role test "${_rid}" 2>/dev/null; then
            _ok=$((_ok + 1))
        else
            log_warn "assign-role test <- ${_rid} failed (may already be bound)"
            _fail=$((_fail + 1))
        fi
    done <<< "${rids}"
    log_info "Role assign: succeeded ${_ok}, failed/duplicate ${_fail} (of ${_n} role ids from role list). Check: kweaver-admin user roles test"
    return 0
}

# After kweaver is usable; only when full ISF + kweaver-admin and user chose to run.
onboard_offer_isf_test_user() {
    if [[ "${ONBOARD_SKIP_ISF_TEST_USER:-false}" == "true" ]]; then
        return 0
    fi
    onboard_isf_full_install || return 0
    command -v kweaver-admin &>/dev/null || return 0

    if ! kweaver-admin --json user list --limit 1 &>/dev/null; then
        log_warn "kweaver-admin: cannot list users (run: kweaver-admin auth login <https://access-url> -k). Skipping test user offer."
        return 0
    fi
    if onboard_user_test_exists; then
        log_info "User test already exists; skip creating a first test account. (Re-run after deleting test or set ONBOARD_SKIP_ISF_TEST_USER=true to silence.)"
        return 0
    fi

    if [[ "${ONBOARD_ASSUME_YES}" == "true" ]]; then
        log_info "ONBOARD: creating user test and assigning all roles (-y)…"
        onboard_create_test_user_with_all_roles
        return 0
    fi
    if ! onboard_is_bootstrap_tty; then
        log_info "Not a TTY: skipping interactive offer to create user test. Re-run in a terminal, or use -y, or: kweaver-admin user create --login test && (role list + assign-role)."
        return 0
    fi
    echo ""
    read -r -p "Create the first business test user [test] (password 123456) and grant ALL roles from role list (POC; use least privilege in production) [Y/n]: " _otu
    if [[ "${_otu}" =~ ^[Nn] ]]; then
        log_info "Skipped creating user test. You can create users later: kweaver-admin user create --login test"
        return 0
    fi
    onboard_create_test_user_with_all_roles
}
