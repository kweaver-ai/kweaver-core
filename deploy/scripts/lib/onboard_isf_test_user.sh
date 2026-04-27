#!/usr/bin/env bash
# Optional: first business test user (login: test) with all roles from kweaver-admin "role list"
# (on typical full stacks this is three business admin roles, e.g. 数据/AI/应用 管理员) + kweaver re-login
# for ADP toolset impex (built-in  admin  often lacks  CommonAdd  on agent-operator;  test  with roles does).
# shellcheck source=/dev/null

# Last password for kweaver HTTP sign-in as test (set in this shell after reset-password or env).
__ONBOARD_TEST_USER_KWEAVER_PASSWORD=""

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

# Assign every role from role list to account (idempotent: duplicate assign may no-op on server).
onboard_assign_all_listed_roles_to_user() {
    local _login="${1:-test}"
    local rjson rids _rid _n _fail _ok
    if ! rjson="$(kweaver-admin --json role list --limit 1000 2>/dev/null)"; then
        log_warn "kweaver-admin role list failed; cannot assign roles to ${_login}"
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
        log_warn "No roles in role list; no roles to assign to ${_login}"
        return 0
    fi
    _fail=0
    _n=0
    _ok=0
    while IFS= read -r _rid; do
        [[ -n "${_rid}" ]] || continue
        _n=$((_n + 1))
        if kweaver-admin user assign-role "${_login}" "${_rid}" 2>/dev/null; then
            _ok=$((_ok + 1))
        else
            log_warn "assign-role ${_login} <- ${_rid} failed (may already be bound)"
            _fail=$((_fail + 1))
        fi
    done <<< "${rids}"
    log_info "Role assign for ${_login}: ok ${_ok}, failed/duplicate ${_fail} (of ${_n} role ids; usually all business admin roles in role list). Check: kweaver-admin user roles ${_login}"
    return 0
}

# Set kweaver-usable password for user  test  (ISF:  user create  leaves platform 123456 until  reset-password ).
# Default password for onboard: ONBOARD_DEFAULT_TEST_USER_PASSWORD (default 111111). Fills __ONBOARD_TEST_USER_KWEAVER_PASSWORD.
onboard_set_test_user_password() {
    local _defp
    _defp="${ONBOARD_DEFAULT_TEST_USER_PASSWORD:-111111}"
    if [[ -n "${ONBOARD_TEST_USER_PASSWORD:-}" ]]; then
        if ! kweaver-admin user reset-password -u test -p "${ONBOARD_TEST_USER_PASSWORD}" -y 2>/dev/null; then
            log_warn "kweaver-admin user reset-password (ONBOARD_TEST_USER_PASSWORD) failed; try: kweaver-admin user reset-password -u test -p '...' -y"
            return 1
        fi
        __ONBOARD_TEST_USER_KWEAVER_PASSWORD="${ONBOARD_TEST_USER_PASSWORD}"
        return 0
    fi
    if [[ "${ONBOARD_ASSUME_YES:-false}" == "true" ]]; then
        if ! kweaver-admin user reset-password -u test -p "${_defp}" -y 2>/dev/null; then
            log_warn "kweaver-admin user reset-password (default ${_defp}) failed; set ONBOARD_TEST_USER_PASSWORD and re-run, or: kweaver-admin user reset-password -u test -p '...' -y"
            return 1
        fi
        __ONBOARD_TEST_USER_KWEAVER_PASSWORD="${_defp}"
        log_info "User test: password set to default ${_defp} (-y, non-interactive). Override: ONBOARD_TEST_USER_PASSWORD=..."
        return 0
    fi
    if ! (type onboard_is_bootstrap_tty &>/dev/null && onboard_is_bootstrap_tty); then
        log_warn "Not a TTY: set ONBOARD_TEST_USER_PASSWORD=... or use  $0 -y  (default password ${_defp} for  test )"
        return 1
    fi
    log_info "Set password for user test (default ${_defp} if you press Enter; same for kweaver impex)…"
    read -r -s -p "  Password for user test (empty = ${_defp}): " __ONBOARD_TEST_USER_KWEAVER_PASSWORD
    echo
    __ONBOARD_TEST_USER_KWEAVER_PASSWORD="${__ONBOARD_TEST_USER_KWEAVER_PASSWORD:-${_defp}}"
    if ! kweaver-admin user reset-password -u test -p "${__ONBOARD_TEST_USER_KWEAVER_PASSWORD}" -y 2>/dev/null; then
        log_warn "kweaver-admin user reset-password failed; set ONBOARD_TEST_USER_PASSWORD=... or: kweaver-admin user reset-password -u test -p '...' -y"
        return 1
    fi
    return 0
}

# Create login=test, set password, assign all roles in role list.
onboard_create_test_user_with_all_roles() {
    log_info "Create user test and assign all roles in kweaver-admin 'role list' (typically three business admin roles)…"
    local uerr
    uerr="$(mktemp 2>/dev/null || echo /tmp/onboard-uc.$$)"
    if ! kweaver-admin --json user create --login test >/dev/null 2> "${uerr}"; then
        if grep -qiE 'already|exists|exist|重复|已存在' "${uerr}" 2>/dev/null; then
            log_info "User test may already exist; continuing…"
        else
            log_warn "kweaver-admin user create failed: $(tr '\n' ' ' < "${uerr}" | head -c 400)"
            rm -f "${uerr}"
            return 1
        fi
    fi
    rm -f "${uerr}" 2>/dev/null || true
    if ! onboard_set_test_user_password; then
        log_warn "Password not set; kweaver re-login for impex may fail until you: kweaver-admin user reset-password -u test -p '...' -y"
        __ONBOARD_TEST_USER_KWEAVER_PASSWORD="${__ONBOARD_TEST_USER_KWEAVER_PASSWORD:-${ONBOARD_TEST_USER_PASSWORD:-${ONBOARD_DEFAULT_TEST_USER_PASSWORD:-111111}}}"
    fi
    if ! onboard_assign_all_listed_roles_to_user test; then
        return 1
    fi
    return 0
}

# kweaver auth as test (HTTP) for impex. Requires access URL; password from __/env or prompt.
# shellcheck disable=SC2120
onboard_kweaver_relogin_isf_test() {
    local kurl="${1:-}"
    if [[ -z "${kurl}" ]] && type onboard_default_access_base_url &>/dev/null; then
        kurl="$(onboard_default_access_base_url)"
    fi
    if [[ -z "${kurl}" ]]; then
        log_warn "kweaver re-login as test: no access URL; set access URL or run from onboard.sh"
        return 1
    fi
    if [[ -n "${ONBOARD_KWEAVER_IMPEX_NO_RELLOGIN:-}" ]]; then
        log_info "ONBOARD_KWEAVER_IMPEX_NO_RELLOGIN set; skipping kweaver auth as test (using existing ~/.kweaver session)"
        return 0
    fi
    local _pw _defp
    _defp="${ONBOARD_DEFAULT_TEST_USER_PASSWORD:-111111}"
    _pw="${__ONBOARD_TEST_USER_KWEAVER_PASSWORD:-${ONBOARD_TEST_USER_PASSWORD:-}}"
    if [[ -z "${_pw}" && -t 0 && -t 1 ]] && (type onboard_is_bootstrap_tty &>/dev/null && onboard_is_bootstrap_tty); then
        read -r -s -p "kweaver: password for user 'test' (ADP impex; empty = ${_defp}): " _pw
        echo
    fi
    _pw="${_pw:-${_defp}}"
    log_info "kweaver auth: signing in as test (HTTP) for impex (token in ~/.kweaver)…"
    if ! kweaver auth login "${kurl}" -u test -p "${_pw}" --http-signin -k; then
        log_warn "kweaver auth as test failed; fix password/URL or re-run: kweaver auth login ${kurl} -k"
        return 1
    fi
    return 0
}

# After kweaver is usable; only when full ISF + kweaver-admin and user chose to run.
onboard_offer_isf_test_user() {
    if [[ "${ONBOARD_SKIP_ISF_TEST_USER:-false}" == "true" ]]; then
        ONBOARD_REPORT_ISF_TEST_USER="已跳过: ONBOARD_SKIP_ISF_TEST_USER / --skip-isf-test-user"
        return 0
    fi
    onboard_isf_full_install || {
        ONBOARD_REPORT_ISF_TEST_USER="不适用: 非 ISF/全量 安装"
        return 0
    }
    command -v kweaver-admin &>/dev/null || {
        ONBOARD_REPORT_ISF_TEST_USER="已跳过: 无 kweaver-admin (需 npm i -g @kweaver-ai/kweaver-admin 后再建 test)"
        log_warn "未检测到 kweaver-admin（可执行  which kweaver-admin ）：已跳过「创建/同步用户 test」。若已  npm -g  装过: 在**当前终端**  export PATH=\"\$(npm config get prefix)/bin:\$PATH\"  后重试本脚本，或:  $0 -y  （ISF 会再装/刷新  PATH ）。"
        log_warn "  首次使用还需:  kweaver-admin auth login <https://平台地址> -k  （再  ./onboard.sh  建  test ）"
        return 0
    }

    if ! kweaver-admin --json user list --limit 1 &>/dev/null; then
        ONBOARD_REPORT_ISF_TEST_USER="已跳过: kweaver-admin 未登录 (auth login 失败或未执行)"
        log_warn "kweaver-admin: cannot list users (run: kweaver-admin auth login <https://access-url> -k). Skipping test user offer."
        return 0
    fi

    if onboard_user_test_exists; then
        log_info "User test already exists. Syncing 'role list' roles to test (for ADP / Context Loader impex)…"
        onboard_assign_all_listed_roles_to_user test || true
        ONBOARD_REPORT_ISF_TEST_USER="已存在 test: 已同步 role list 中全部角色 (kweaver-admin user roles test 可查)"
        log_info "If you reset test's password, export ONBOARD_TEST_USER_PASSWORD for non-interactive kweaver impex, or enter it when asked during Context Loader import."
        return 0
    fi

    if [[ "${ONBOARD_ASSUME_YES}" == "true" ]]; then
        log_info "ONBOARD: creating user test, password/roles (-y)…"
        onboard_create_test_user_with_all_roles
        ONBOARD_REPORT_ISF_TEST_USER="已创建 test 并赋权 (-y: 密码默认 ${ONBOARD_DEFAULT_TEST_USER_PASSWORD:-111111}，可设 ONBOARD_TEST_USER_PASSWORD=...)"
        return 0
    fi
    if ! (type onboard_is_bootstrap_tty &>/dev/null && onboard_is_bootstrap_tty); then
        ONBOARD_REPORT_ISF_TEST_USER="已跳过: 非 TTY (可用 -y 或手动手册)"
        log_info "Not a TTY: skipping interactive offer to create user test. Re-run in a terminal, or use -y, or: kweaver-admin user create --login test && …"
        return 0
    fi
    echo ""
    read -r -p "Create business user [test] (set password) and grant ALL roles from kweaver-admin 'role list' (three admin roles in a typical full install) for ADP import? [Y/n]: " _otu
    if [[ "${_otu}" =~ ^[Nn] ]]; then
        ONBOARD_REPORT_ISF_TEST_USER="已跳过: 用户未创建 test"
        log_info "Skipped. You can: kweaver-admin user create --login test && kweaver-admin user reset-password -u test --prompt-password -y && (assign all role ids from role list)"
        return 0
    fi
    onboard_create_test_user_with_all_roles
    ONBOARD_REPORT_ISF_TEST_USER="已创建 test、设密并挂接 role list 中全部业务角色"
}

# Before kweaver call (ADP impex), ISF: ensure user test + business roles, then kweaver session as test.
# Requires onboard.sh: onboard_default_access_base_url. Returns 0 to proceed (no ISF: always 0).
onboard_ensure_isf_test_for_kweaver_impex() {
    type onboard_isf_full_install &>/dev/null || return 0
    onboard_isf_full_install 2>/dev/null || return 0
    if ! command -v kweaver &>/dev/null; then
        return 0
    fi
    if ! command -v kweaver-admin &>/dev/null; then
        log_info "ISF+impex: kweaver-admin not in PATH; using your current  kweaver  login (may 403; install admin CLI or log in as test.)"
        return 0
    fi
    if ! kweaver-admin --json user list --limit 1 &>/dev/null; then
        log_info "ISF+impex: kweaver-admin not authenticated; using current kweaver session for impex"
        return 0
    fi
    if ! onboard_user_test_exists; then
        log_warn "ISF+impex: user  test  is missing. Create it (kweaver-admin user create / onboard offer) and assign  role list  roles, then re-run this import."
        return 1
    fi
    log_info "ISF+impex: sync role list to user test, then  kweaver auth  as test (built-in  admin  often cannot import toolboxes)…"
    onboard_assign_all_listed_roles_to_user test || true
    local _url=""
    if type onboard_default_access_base_url &>/dev/null; then
        _url="$(onboard_default_access_base_url)"
    fi
    if ! onboard_kweaver_relogin_isf_test "${_url}"; then
        return 1
    fi
    return 0
}
