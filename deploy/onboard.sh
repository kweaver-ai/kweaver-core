#!/usr/bin/env bash
# KWeaver Core — onboard: register models, BKN, rollout (run after `deploy.sh` install)
# Requires: kweaver, kubectl, python3, PyYAML (pip3 install pyyaml) for --config; interactive is lighter.
# Run from the deploy/ directory (symmetric with preflight.sh).

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/lib/common.sh
source "${SCRIPT_DIR}/scripts/lib/common.sh"

# Prefer Mac dev values when running onboard directly: common.sh defaults CONFIG_YAML_PATH to
# deploy/conf/config.yaml (often no accessAddress). If still on that default, use mac-config.yaml.
_onboard_default_conf="${SCRIPT_DIR}/conf/config.yaml"
_onboard_default_home="${HOME}/.kweaver-ai/config.yaml"
_onboard_mac_cfg="${SCRIPT_DIR}/dev/conf/mac-config.yaml"
if [[ "$(uname -s 2>/dev/null || true)" == "Darwin" ]] && [[ -f "${_onboard_mac_cfg}" ]]; then
    if [[ "${CONFIG_YAML_PATH:-}" == "${_onboard_default_conf}" ]] || [[ "${CONFIG_YAML_PATH:-}" == "${_onboard_default_home}" ]]; then
        export CONFIG_YAML_PATH="${_onboard_mac_cfg}"
    fi
fi

# shellcheck source=scripts/lib/onboard_models.sh
source "${SCRIPT_DIR}/scripts/lib/onboard_models.sh"

NAMESPACE="${NAMESPACE:-kweaver}"
CONFIG_FILE=""
BKN_NAME=""
ENABLE_BKN_ONLY="false"
SKIP_BKN="false"
INTERACTIVE="true"
ONBOARD_ASSUME_YES="false"
ONBOARD_SKIP_ISF_TEST_USER="${ONBOARD_SKIP_ISF_TEST_USER:-false}"
ONBOARD_SKIP_CONTEXT_LOADER="${ONBOARD_SKIP_CONTEXT_LOADER:-false}"

# kweaver auth: HTTP sign-in defaults (ISF / full install). Console account is usually  admin  /  eisoo.com  if not changed.
# Override in CI. Used when you press Enter at username/password prompts.
: "${ONBOARD_DEFAULT_KWEAVER_USER:=admin}"
: "${ONBOARD_DEFAULT_KWEAVER_PASSWORD:=eisoo.com}"

# ISF: first business user  test  (after  kweaver-admin user create ) — platform default is 123456 until  reset-password;
# we set this for onboard default /  -y  / empty Enter. Override: ONBOARD_TEST_USER_PASSWORD; rename default: ONBOARD_DEFAULT_TEST_USER_PASSWORD
: "${ONBOARD_DEFAULT_TEST_USER_PASSWORD:=111111}"

# Same requirement as @kweaver-ai/kweaver-sdk on npm (node >= 22). https://www.npmjs.com/package/@kweaver-ai/kweaver-sdk
ONBOARD_MIN_NODE_MAJOR="${ONBOARD_MIN_NODE_MAJOR:-22}"

# Node/kweaver bootstrap prompts use the terminal (even with --config); -y skips those prompts; no TTY + no -y = error.
onboard_is_bootstrap_tty() {
    [[ -t 0 && -t 1 ]]
}

# shellcheck source=scripts/lib/onboard_isf_test_user.sh
source "${SCRIPT_DIR}/scripts/lib/onboard_isf_test_user.sh"
# shellcheck source=scripts/lib/onboard_context_loader.sh
source "${SCRIPT_DIR}/scripts/lib/onboard_context_loader.sh"
# shellcheck source=scripts/lib/onboard_report.sh
source "${SCRIPT_DIR}/scripts/lib/onboard_report.sh"

# Primary IPv4 of this host (for default KWeaver access URL). Override: ONBOARD_DEFAULT_ACCESS_IP=...
onboard_default_local_ipv4() {
    if [[ -n "${ONBOARD_DEFAULT_ACCESS_IP:-}" ]]; then
        echo "${ONBOARD_DEFAULT_ACCESS_IP}"
        return
    fi
    python3 -c "
import re
import socket
import subprocess
import sys

def main() -> None:
    for remote in ('8.8.8.8', '1.1.1.1'):
        try:
            s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
            s.settimeout(0.4)
            s.connect((remote, 80))
            print(s.getsockname()[0])
            s.close()
            return
        except Exception:
            pass
    try:
        out = subprocess.check_output(
            ['ip', '-4', 'route', 'get', '1.1.1.1'], text=True, stderr=subprocess.DEVNULL, timeout=2
        )
        m = re.search(r'\\bsrc (\\d+\\.\\d+\\.\\d+\\.\\d+)', out)
        if m:
            print(m.group(1))
            return
    except Exception:
        pass
    try:
        out = subprocess.check_output(
            ['ipconfig', 'getifaddr', 'en0'], text=True, stderr=subprocess.DEVNULL, timeout=2
        ).strip()
        if out and out != '0.0.0.0':
            print(out)
            return
    except Exception:
        pass
    print('127.0.0.1')

if __name__ == '__main__':
    main()
" 2>/dev/null || echo "127.0.0.1"
}

# Default access base for  kweaver auth login  (this machine, HTTPS + primary IPv4 unless overridden).
# Set ONBOARD_DEFAULT_ACCESS_BASE to a full URL to skip auto IP.
# Otherwise, when ONBOARD_SKIP_CONFIG_ACCESS_URL!=true, prefers accessAddress.host|port|scheme|path
# from CONFIG_YAML_PATH (same file as deploy.sh / Helm values; mac.sh sets CONFIG_YAML_PATH to
# deploy/dev/conf/mac-config.yaml).
onboard_default_access_base_url() {
    if [[ -n "${ONBOARD_DEFAULT_ACCESS_BASE:-}" ]]; then
        echo "${ONBOARD_DEFAULT_ACCESS_BASE%/}"
        return
    fi
    local _from_cfg=""
    if [[ "${ONBOARD_SKIP_CONFIG_ACCESS_URL:-false}" != "true" ]] && type get_access_address_base_url &>/dev/null; then
        _from_cfg="$(get_access_address_base_url 2>/dev/null || true)"
    fi
    if [[ -n "${_from_cfg}" ]]; then
        echo "${_from_cfg%/}"
        return
    fi
    local ip _scheme _port
    ip="$(onboard_default_local_ipv4)"
    _scheme="${ONBOARD_DEFAULT_ACCESS_SCHEME:-https}"
    _port="${ONBOARD_DEFAULT_ACCESS_PORT:-}"
    if [[ -n "${_port}" ]]; then
        echo "${_scheme}://${ip}:${_port}"
    else
        echo "${_scheme}://${ip}"
    fi
}

# kweaver / kweaver-admin: append --insecure/-k only for HTTPS URLs (self-signed dev certs).
# For plain http:// bases, unconditional -k has been observed to break some login flows (--no-auth)
# against HTTP-only ingress backends (404 Not Found).
# Override: ONBOARD_FORCE_INSECURE_LOGIN=true forces -k even for HTTP (rare; debugging only).
onboard_kweaver_tls_insecure_args() {
    local _base="$1"
    if [[ "${ONBOARD_FORCE_INSECURE_LOGIN:-false}" == "true" ]]; then
        printf '%s' "-k "
        return
    fi
    case "${_base}" in
        https://*|HTTPS://*)
            printf '%s' "-k "
            ;;
        *)
            return
            ;;
    esac
}

usage() {
    echo "Usage: $0 [options]"
    echo "  Requires: Node 22+ (see @kweaver-ai/kweaver-sdk on npm), kweaver, kubectl, python3; run from deploy/"
    echo "  (no flags)                Interactive: nvm+Node 22 and npm -g (Y/n) in your terminal, then models/BKN"
    echo "  -y, --yes                 Auto nvm+Node 22, npm -g, context-loader import, ISF [test] user+roles (no Y/n)"
    echo "  --config=PATH            YAML: deploy/conf/models.yaml.example; model prompts off, but nvm/kweaver still Y/n in a TTY (use -y to skip those asks)"
    echo "  --skip-isf-test-user     Do not offer: kweaver-admin user test + all roles (full install only)"
    echo "  --skip-context-loader   Do not offer Context Loader ADP import (kweaver call impex); same as ONBOARD_SKIP_CONTEXT_LOADER=true"
    echo ""
    echo "  Context Loader (impex):  kweaver call  uses ~/.kweaver from  kweaver auth login ."
    echo "    - ISF (full):  kweaver-admin  / console  admin  for user ops. ADP impex uses user  test  with all  role list"
    echo "      roles (typically three business admins), then  kweaver auth  as  test .  -y  uses password  ${ONBOARD_DEFAULT_TEST_USER_PASSWORD:-111111}  (override: ONBOARD_TEST_USER_PASSWORD) ."
    echo "    - Minimum (no ISF):  kweaver auth login  only; kweaver-admin is not required."
    echo "  --namespace=NS           (default: kweaver; or key 'namespace' in yaml)"
    echo "  --enable-bkn-search      Only patch bkn/ontology ConfigMaps and rollout"
    echo "  --bkn-embedding-name=X   Required with --enable-bkn-search (registered model_name)"
    echo "  --skip-bkn               With --config: register models but skip BKN + rollout"
    echo "  -h, --help"
    echo ""
    echo "  Environment: ONBOARD_SKIP_NODE_INSTALL=true  skip nvm in onboard (fail if Node < ${ONBOARD_MIN_NODE_MAJOR})"
    echo "                ONBOARD_SKIP_KWEAVER_INSTALL=true  never run npm -g for kweaver in onboard"
    echo "                ONBOARD_SKIP_KWEAVER_ADMIN_INSTALL=true  on ISF: do not auto/offer  npm -g  kweaver-admin  ( -y  也不会装 )"
    echo "                ONBOARD_SKIP_ISF_TEST_USER=true  same as --skip-isf-test-user"
    echo "                ONBOARD_SKIP_CONTEXT_LOADER=true  same as --skip-context-loader"
    echo "                IMPORT_CONTEXT_LOADER_TOOLSET=false  skip Context Loader (legacy name; same effect)"
    echo "                CONTEXT_LOADER_TOOLSET_ADP_PATH=...  default ADP under repo adp/context-loader/.../context_loader_toolset.adp"
    echo "                ONBOARD_TEST_USER_PASSWORD=...  override default password for  test  (ISF; default: ONBOARD_DEFAULT_TEST_USER_PASSWORD, built-in 111111)"
    echo "                ONBOARD_DEFAULT_TEST_USER_PASSWORD=...  first-user  test  password (default 111111;  -y  non-interactive)"
    echo "                ONBOARD_KWEAVER_IMPEX_NO_RELLOGIN=1  skip  kweaver auth  as  test  before impex (use current kweaver session)"
    echo "                ONBOARD_NO_COMPLETION_REPORT=1  do not print the English completion report at the end"
    echo "                ONBOARD_FORCE_INSECURE_LOGIN=true  always pass -k (--insecure) to kweaver/kweaver-admin auth login (even for http:// bases; default false)"
    echo "                ONBOARD_SKIP_CONFIG_ACCESS_URL=true  do not derive default URL from CONFIG_YAML_PATH accessAddress"
    echo "  Default KWeaver access URL (kweaver auth): accessAddress in CONFIG_YAML_PATH when present;"
    echo "                on macOS, if CONFIG_YAML_PATH is still deploy/conf/config.yaml or ~/.kweaver-ai/config.yaml,"
    echo "                onboard uses deploy/dev/conf/mac-config.yaml when that file exists (same as mac.sh)."
    echo "                Else host primary IPv4 + ONBOARD_DEFAULT_ACCESS_SCHEME (https by default)."
    echo "                Set ONBOARD_DEFAULT_ACCESS_BASE to force a URL; ONBOARD_DEFAULT_ACCESS_PORT / SCHEME override fallback IP path."
    echo "  kweaver auth: you confirm URL. ISF+full: HTTP defaults user=admin pass=eisoo.com (if still default); override with ONBOARD_DEFAULT_KWEAVER_USER / _PASSWORD. Enter keeps defaults. Minimum: default --no-auth; Enter to accept."
    echo "  kweaver-admin auth (ISF): use  auth login <url> -u admin -p <pass>  (append -k for https:// + self-signed) or browser-only flow; kweaver-admin has no --http-signin flag. Then kweaver re-logs in as user test for impex and model steps."
    echo "  Node: onboard is not a login shell — it auto-loads nvm/fnm/asdf/Volta and Homebrew paths so an already-configured Node 22+ is found without re-asking. ONBOARD_SKIP_NVM_INIT=true skips that; ONBOARD_NVM_VERSION=22 (default) is used after  nvm.sh  load."
    echo "  (preflight on the server: sudo preflight --fix still optional; this script can install Node in your *user* account via nvm.)"
}

for _ob_arg in "$@"; do
    case "${_ob_arg}" in
        -h | --help) usage; exit 0 ;;
        --config=*) INTERACTIVE="false" ;;
        -y | --yes) ONBOARD_ASSUME_YES="true" ;;
        --skip-isf-test-user) ONBOARD_SKIP_ISF_TEST_USER="true" ;;
        --skip-context-loader) ONBOARD_SKIP_CONTEXT_LOADER="true" ;;
    esac
done

onboard_node_major() {
    if ! command -v node &>/dev/null; then
        echo 0
        return
    fi
    local v
    v="$(node -v 2>/dev/null)"
    v="${v#v}"
    v="${v%%.*}"
    if [[ "${v}" =~ ^[0-9]+$ ]]; then
        echo "${v}"
    else
        echo 0
    fi
}

# This script is not a login shell: ~/.zshrc / .bashrc are not sourced, so nvm's node is often missing
# from PATH even when the user already "configured" it in a terminal. Load common version managers
# and standard locations before we decide to prompt for nvm install.
onboard_bootstrap_node_path() {
    if [[ "${ONBOARD_SKIP_NVM_INIT:-false}" == "true" ]]; then
        return 0
    fi
    # Volta
    if [[ -d "${HOME}/.volta/bin" ]]; then
        case ":${PATH}:" in *":${HOME}/.volta/bin:"*) ;; *) export PATH="${HOME}/.volta/bin:${PATH}" ;; esac
    fi
    # asdf
    if [[ -f "${HOME}/.asdf/asdf.sh" ]]; then
        # shellcheck source=/dev/null
        . "${HOME}/.asdf/asdf.sh" 2>/dev/null && hash -r 2>/dev/null || true
    fi
    # fnm
    if command -v fnm &>/dev/null; then
        # shellcheck disable=SC1091
        eval "$(fnm env 2>/dev/null)" && hash -r 2>/dev/null || true
    fi
    # nvm (most common)
    export NVM_DIR="${NVM_DIR:-$HOME/.nvm}"
    if [[ -s "${NVM_DIR}/nvm.sh" ]]; then
        # shellcheck source=/dev/null
        if . "${NVM_DIR}/nvm.sh" 2>/dev/null; then
            nvm use "${ONBOARD_NVM_VERSION:-22}" 2>/dev/null || nvm use default 2>/dev/null || nvm use node 2>/dev/null || true
            hash -r 2>/dev/null || true
        fi
    fi
    # Homebrew (macOS): node@22 is often in PATH once these dirs are prepended
    for _nd in /opt/homebrew/bin /usr/local/bin; do
        if [[ -x "${_nd}/node" ]]; then
            case ":${PATH}:" in *":${_nd}:"*) ;; *) export PATH="${_nd}:${PATH}" ;; esac
        fi
    done
}

# Install nvm + Node 22 in the current user (no sudo; same idea as preflight's node-22 fix, user-local).
onboard_install_node22_nvm() {
    if ! command -v curl &>/dev/null; then
        log_error "curl is required to install nvm. Install curl, or install Node ${ONBOARD_MIN_NODE_MAJOR}+ from https://nodejs.org/"
        return 1
    fi
    if ! command -v bash &>/dev/null; then
        log_error "bash is required to run the nvm installer."
        return 1
    fi
    export NVM_DIR="${NVM_DIR:-$HOME/.nvm}"
    if [[ ! -s "${NVM_DIR}/nvm.sh" ]]; then
        log_info "Installing nvm into ${NVM_DIR}…"
        if ! curl -fsSL https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.1/install.sh | bash; then
            log_error "nvm install.sh failed (network, proxy, or missing deps). See https://github.com/nvm-sh/nvm"
            return 1
        fi
    fi
    # shellcheck source=/dev/null
    if ! . "${NVM_DIR}/nvm.sh"; then
        log_error "Could not source \${NVM_DIR}/nvm.sh (${NVM_DIR})"
        return 1
    fi
    if ! nvm install 22; then
        log_error "nvm install 22 failed"
        return 1
    fi
    nvm use 22
    nvm alias default 22 2>/dev/null || true
    hash -r 2>/dev/null || true
    return 0
}

# If not sudo preflight --fix, still help: offer (or with -y, run) nvm+Node 22 in this user.
onboard_ensure_node_22() {
    local mj
    onboard_bootstrap_node_path
    mj="$(onboard_node_major)"
    if command -v node &>/dev/null && [[ -n "${mj}" && $(( 10#${mj} )) -ge ${ONBOARD_MIN_NODE_MAJOR} ]]; then
        log_info "Using $(node -v) ($(command -v node))"
        return 0
    fi

    if [[ "${ONBOARD_SKIP_NODE_INSTALL:-false}" == "true" ]]; then
        log_error "Node is $(node -v 2>/dev/null || echo missing) but Node.js ${ONBOARD_MIN_NODE_MAJOR}+ is required. Unset ONBOARD_SKIP_NODE_INSTALL or install Node manually."
        exit 1
    fi

    # Interactive on a TTY: always ask, including when --config is set. No TTY: must pass -y to auto-install.
    if [[ "${ONBOARD_ASSUME_YES}" == "true" ]]; then
        log_info "Node ${ONBOARD_MIN_NODE_MAJOR}+ not active; installing via nvm (-y)…"
    elif onboard_is_bootstrap_tty; then
        echo ""
        read -r -p "Node.js ${ONBOARD_MIN_NODE_MAJOR}+ is required for kweaver/onboard. Install nvm and Node 22 in this user account now? [Y/n]: " _obn
        if [[ "${_obn}" =~ ^[Nn] ]]; then
            log_error "Install Node ${ONBOARD_MIN_NODE_MAJOR}+ (e.g. nvm install 22), or use another machine with Node 22+ on PATH, or run: sudo preflight --fix on the host where you need system-wide Node."
            exit 1
        fi
    else
        log_error "Node ${ONBOARD_MIN_NODE_MAJOR}+ required (or missing). In a real terminal you get a Y/n prompt; without a TTY pass  $0 -y  (e.g. CI), or install Node / nvm first. Or: sudo preflight --fix (onboard-tooling) on a server."
        exit 1
    fi

    if ! onboard_install_node22_nvm; then
        exit 1
    fi
    mj="$(onboard_node_major)"
    if ! command -v node &>/dev/null || [[ -z "${mj}" || $(( 10#${mj} )) -lt ${ONBOARD_MIN_NODE_MAJOR} ]]; then
        log_error "Node is still < ${ONBOARD_MIN_NODE_MAJOR} in this process. In a new terminal run:  source \"\$NVM_DIR/nvm.sh\" && nvm use 22  then:  $0  again."
        exit 1
    fi
    # After a fresh nvm install in this function, the success message is similar
    if command -v node &>/dev/null; then
        log_info "Using Node $(node -v) ($(command -v node))"
    fi
}

onboard_ensure_kweaver_cli() {
    if command -v kweaver &>/dev/null; then
        return 0
    fi
    if ! command -v npm &>/dev/null; then
        log_error "kweaver not in PATH and npm not found. With nvm+Node, npm should exist; re-open a shell and re-run."
        exit 1
    fi
    if [[ "${ONBOARD_SKIP_KWEAVER_INSTALL:-false}" == "true" ]]; then
        log_error "kweaver not in PATH. Install: npm i -g @kweaver-ai/kweaver-sdk  (or unset ONBOARD_SKIP_KWEAVER_INSTALL to allow this script to run npm -g.)"
        exit 1
    fi
    if [[ "${ONBOARD_ASSUME_YES}" == "true" ]]; then
        log_info "Installing @kweaver-ai/kweaver-sdk globally (-y)…"
    elif onboard_is_bootstrap_tty; then
        echo ""
        read -r -p "kweaver CLI not in PATH. Install @kweaver-ai/kweaver-sdk globally now? (npm i -g) [Y/n]: " _obk
        if [[ "${_obk}" =~ ^[Nn] ]]; then
            log_error "kweaver is required. Run:  npm i -g @kweaver-ai/kweaver-sdk"
            exit 1
        fi
    else
        log_error "kweaver not in PATH. In a TTY you get a Y/n prompt; without a TTY use  $0 -y  or install: npm i -g @kweaver-ai/kweaver-sdk"
        exit 1
    fi
    if ! npm i -g @kweaver-ai/kweaver-sdk; then
        log_error "npm i -g @kweaver-ai/kweaver-sdk failed. Check registry/proxy, or EACCES (avoid sudo; use nvm user prefix.)"
        exit 1
    fi
    hash -r 2>/dev/null || true
    if ! command -v kweaver &>/dev/null; then
        log_error "kweaver still not on PATH. Add npm global bin to PATH, e.g.:  export PATH=\"\$(npm config get prefix 2>/dev/null)/bin:\$PATH\""
        exit 1
    fi
    log_info "kweaver: $(kweaver --version 2>/dev/null | head -1)"
}

# Same shell as nvm/node: global CLIs (kweaver, kweaver-admin) live under $(npm config get prefix)/bin — prepend so a just-installed -g is visible.
onboard_prepend_npm_global_bin_to_path() {
    local pfx
    pfx="$(npm config get prefix 2>/dev/null)" || true
    if [[ -n "${pfx}" && -d "${pfx}/bin" ]]; then
        case ":${PATH}:" in
            *":${pfx}/bin:"*) ;;
            *) export PATH="${pfx}/bin${PATH:+:${PATH}}" ;;
        esac
    fi
    hash -r 2>/dev/null || true
}

onboard_ensure_node_22
onboard_ensure_kweaver_cli
if ! command -v kubectl &>/dev/null; then
    log_error "kubectl not found"
    exit 1
fi
if ! command -v python3 &>/dev/null; then
    log_error "python3 not found"
    exit 1
fi
# Verify (and best-effort install) yq OR PyYAML now, so we never get stuck
# halfway through onboard at the BKN ConfigMap patch step. Skip when the
# operator has explicitly opted out of BKN with --skip-bkn.
__onboard_skip_bkn_early=false
for __arg in "$@"; do
    case "${__arg}" in
        --skip-bkn) __onboard_skip_bkn_early=true ;;
    esac
done
if [[ "${__onboard_skip_bkn_early}" != "true" ]]; then
    if ! onboard_ensure_yaml_dep; then
        log_error "onboard.sh needs PyYAML or yq to patch the BKN ConfigMap. Install one of the commands above (or pass --skip-bkn if you really want to onboard without touching BKN), then re-run."
        exit 1
    fi
fi
unset __onboard_skip_bkn_early __arg

while [[ $# -gt 0 ]]; do
    case "$1" in
        -h | --help)
            usage
            exit 0
            ;;
        -y | --yes)
            ONBOARD_ASSUME_YES="true"
            shift
            ;;
        --skip-isf-test-user)
            ONBOARD_SKIP_ISF_TEST_USER="true"
            shift
            ;;
        --skip-context-loader)
            ONBOARD_SKIP_CONTEXT_LOADER="true"
            shift
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

onboard_kweaver_auth_login_echo_cmd() {
    local _cmd="kweaver auth login \"$1\""
    shift
    local _arg
    for _arg in "$@"; do
        _cmd="${_cmd} ${_arg}"
    done
    onboard_log_info "Running: ${_cmd}"
}

# After access URL is chosen: ISF → HTTP sign-in (defaults admin / eisoo.com if unchanged) or browser; no ISF → --no-auth (Enter) or HTTP.
# Env: ONBOARD_DEFAULT_KWEAVER_USER, ONBOARD_DEFAULT_KWEAVER_PASSWORD, ONBOARD_ASSUME_YES (non-interactive: ISF=HTTP+defaults, min=--no-auth).
onboard_kweaver_auth_login_for_url() {
    local _kurl="$1"
    local _tls
    local _u _p _duser _dpass
    _duser="${ONBOARD_DEFAULT_KWEAVER_USER:-admin}"
    _dpass="${ONBOARD_DEFAULT_KWEAVER_PASSWORD:-eisoo.com}"
    _tls="$(onboard_kweaver_tls_insecure_args "${_kurl}")"
    onboard_log_info "kweaver CLI: $(command -v kweaver 2>/dev/null || echo missing) ($(kweaver --version 2>/dev/null || echo '?')) CONFIG_YAML_PATH=${CONFIG_YAML_PATH:-unset}"

    if type onboard_isf_full_install &>/dev/null && onboard_isf_full_install 2>/dev/null; then
        if [[ "${ONBOARD_ASSUME_YES}" == "true" ]]; then
            onboard_log_info "kweaver auth: ISF detected — HTTP sign-in (defaults, -y): ${_duser}"
            onboard_kweaver_auth_login_echo_cmd "${_kurl}" -u "${_duser}" -p '***' --http-signin ${_tls}
            if ! kweaver auth login "${_kurl}" -u "${_duser}" -p "${_dpass}" --http-signin ${_tls} ; then
                return 1
            fi
            return 0
        fi
        echo ""
        read -r -p "ISF (full) install: HTTP sign-in (user/password; recommended) [Y/n] (Enter = Y): " _htt
        if [[ -z "${_htt}" || ! "${_htt}" =~ ^[Nn] ]]; then
            read -r -p "  Username [Enter = ${_duser}]: " _u
            _u="${_u:-${_duser}}"
            read -r -s -p "  Password [Enter = ${_dpass} if default unchanged on console] " _p
            echo
            _p="${_p:-${_dpass}}"
            onboard_kweaver_auth_login_echo_cmd "${_kurl}" -u "${_u}" -p '***' --http-signin ${_tls}
            if ! kweaver auth login "${_kurl}" -u "${_u}" -p "${_p}" --http-signin ${_tls} ; then
                return 1
            fi
            return 0
        fi
        onboard_log_info "Using browser / device flow: kweaver auth login \"${_kurl}\" ${_tls}"
        onboard_kweaver_auth_login_echo_cmd "${_kurl}" ${_tls}
        if ! kweaver auth login "${_kurl}" ${_tls} ; then
            return 1
        fi
        return 0
    fi

    if [[ "${ONBOARD_ASSUME_YES}" == "true" ]]; then
        onboard_log_info "kweaver auth: no ISF — --no-auth (default, -y)"
        onboard_kweaver_auth_login_echo_cmd "${_kurl}" --no-auth ${_tls}
        if ! kweaver auth login "${_kurl}" --no-auth ${_tls} ; then
            return 1
        fi
        return 0
    fi
    echo ""
    read -r -p "No ISF (minimum install): use --no-auth (typical) [Y/n] (Enter = Y): " _mna
    if [[ -z "${_mna}" || ! "${_mna}" =~ ^[Nn] ]]; then
        onboard_kweaver_auth_login_echo_cmd "${_kurl}" --no-auth ${_tls}
        if ! kweaver auth login "${_kurl}" --no-auth ${_tls} ; then
            return 1
        fi
        return 0
    fi
    read -r -p "  Username [Enter = ${_duser}]: " _u
    _u="${_u:-${_duser}}"
    read -r -s -p "  Password [Enter = ${_dpass} if default unchanged on console] " _p
    echo
    _p="${_p:-${_dpass}}"
    onboard_kweaver_auth_login_echo_cmd "${_kurl}" -u "${_u}" -p '***' --http-signin ${_tls}
    if ! kweaver auth login "${_kurl}" -u "${_u}" -p "${_p}" --http-signin ${_tls} ; then
        return 1
    fi
    return 0
}

# kweaver-admin: -u/-p use HTTP /oauth2/signin (no --http-signin flag; unlike kweaver-sdk). Same defaults as kweaver. See ONBOARD_DEFAULT_KWEAVER_*.
onboard_kweaver_admin_auth_login_for_url() {
    local _kurl="$1"
    local _tls
    local _u _p _duser _dpass
    _duser="${ONBOARD_DEFAULT_KWEAVER_USER:-admin}"
    _dpass="${ONBOARD_DEFAULT_KWEAVER_PASSWORD:-eisoo.com}"
    _tls="$(onboard_kweaver_tls_insecure_args "${_kurl}")"

    if [[ "${ONBOARD_ASSUME_YES}" == "true" ]]; then
        onboard_log_info "kweaver-admin auth: ISF — HTTP sign-in (defaults, -y): ${_duser}"
        onboard_log_info "Running: kweaver-admin auth login \"${_kurl}\" -u ${_duser} -p *** ${_tls}"
        if ! kweaver-admin auth login "${_kurl}" -u "${_duser}" -p "${_dpass}" ${_tls} ; then
            return 1
        fi
        return 0
    fi
    echo ""
    read -r -p "  Username [Enter = ${_duser}]: " _u
    _u="${_u:-${_duser}}"
    read -r -s -p "  Password [Enter = ${_dpass} if default unchanged on console] " _p
    echo
    _p="${_p:-${_dpass}}"
    onboard_log_info "kweaver-admin: HTTP sign-in…"
    onboard_log_info "Running: kweaver-admin auth login \"${_kurl}\" -u ${_u} -p *** ${_tls}"
    if ! kweaver-admin auth login "${_kurl}" -u "${_u}" -p "${_p}" ${_tls} ; then
        return 1
    fi
    return 0
}

# When kweaver bkn list fails, interactively let the user log in or retry; non-interactive (or -y) exits.
onboard_ensure_kweaver_auth() {
    while true; do
        if kweaver bkn list &>/dev/null; then
            return 0
        fi
        if [[ "${INTERACTIVE}" != "true" ]]; then
            _durl="$(onboard_default_access_base_url)"
            onboard_log_err "kweaver bkn list failed. Run: kweaver auth login ${_durl} $(onboard_kweaver_tls_insecure_args "${_durl}") (or set ONBOARD_DEFAULT_ACCESS_BASE=...)"
            exit 1
        fi
        if [[ "${ONBOARD_ASSUME_YES}" == "true" ]]; then
            _durl="$(onboard_default_access_base_url)"
            onboard_log_warn "kweaver bkn list failed (-y, non-interactive). Trying  kweaver auth login ${_durl}  with defaults…"
            if ! onboard_kweaver_auth_login_for_url "${_durl}"; then
                onboard_log_err "kweaver auth login failed. Set ONBOARD_DEFAULT_ACCESS_BASE / ONBOARD_DEFAULT_KWEAVER_USER / _PASSWORD, or run kweaver auth login manually, then re-run: $0 -y"
                exit 1
            fi
            if ! kweaver bkn list &>/dev/null; then
                onboard_log_err "kweaver bkn list still fails after auth login. Check the platform, then re-run: $0 -y"
                exit 1
            fi
            return 0
        fi
        onboard_log_warn "kweaver bkn list failed (not logged in or platform unreachable)."
        echo ""
        echo "Choose:"
        echo "  1) Run login: URL (Enter = this host IP), then ISF/HTTP or minimum/--no-auth — see -h for defaults"
        echo "  2) Retry (after you ran login in another terminal)"
        echo "  3) Quit"
        read -r -p "Select [1-3] (default: 1): " _kwa
        _kwa="${_kwa:-1}"
        case "${_kwa}" in
            1)
                _def_url="$(onboard_default_access_base_url)"
                read -r -p "Access base URL [Enter = ${_def_url}]: " _kurl
                _kurl="${_kurl:-${_def_url}}"
                if ! onboard_kweaver_auth_login_for_url "${_kurl}"; then
                    onboard_log_warn "kweaver auth login failed. If you saw engine, SyntaxError, or RegExp issues under node_modules, upgrade Node (see npm @kweaver-ai/kweaver-sdk engines), then reinstall the CLI."
                    onboard_log_warn "Otherwise: set ONBOARD_DEFAULT_ACCESS_* or run login manually, then choose 2 to retry."
                fi
                ;;
            2) : ;;
            3) exit 1 ;;
            *) onboard_log_warn "Invalid choice, try again." ;;
        esac
    done
}

# ISF 全量：无 kweaver-admin 则无法  user create / assign-role ；可选安装（交互 Y/n 或  -y  自动、或跳过）
onboard_ensure_kweaver_admin_for_isf() {
    if ! (type onboard_isf_full_install &>/dev/null && onboard_isf_full_install 2>/dev/null); then
        return 0
    fi
    onboard_prepend_npm_global_bin_to_path
    command -v kweaver-admin &>/dev/null && return 0
    if [[ "${ONBOARD_SKIP_KWEAVER_ADMIN_INSTALL:-false}" == "true" ]]; then
        onboard_log_info "kweaver-admin: skip npm install (ONBOARD_SKIP_KWEAVER_ADMIN_INSTALL=true)."
        return 0
    fi
    if ! command -v npm &>/dev/null; then
        onboard_log_warn "kweaver-admin not in PATH and npm is missing; cannot offer install. Install Node/npm first."
        return 0
    fi
    if [[ "${ONBOARD_ASSUME_YES}" == "true" ]]; then
        onboard_log_info "ISF: installing @kweaver-ai/kweaver-admin (-y)…"
        if ! npm i -g @kweaver-ai/kweaver-admin; then
            onboard_log_warn "npm i -g @kweaver-ai/kweaver-admin failed; install manually, then: kweaver-admin auth login <url> -u admin -p '<password>'  (-k only for https:// + self-signed; kweaver-admin: HTTP sign-in, no --http-signin flag)"
        fi
        hash -r 2>/dev/null || true
        onboard_prepend_npm_global_bin_to_path
        if command -v kweaver-admin &>/dev/null; then
            onboard_log_info "kweaver-admin: $(kweaver-admin --version 2>/dev/null | head -n1)"
        fi
        return 0
    fi
    if ! onboard_is_bootstrap_tty; then
        return 0
    fi
    echo ""
    read -r -p "ISF: run  npm i -g @kweaver-ai/kweaver-admin  to create user [test] (re-run is OK if already installed) [Y/n]: " _kadm
    if [[ -n "${_kadm}" && "${_kadm}" =~ ^[Nn] ]]; then
        onboard_log_warn "kweaver-admin not installed: user test will not be created this run. Install later: npm i -g @kweaver-ai/kweaver-admin"
        return 0
    fi
    if ! npm i -g @kweaver-ai/kweaver-admin; then
        onboard_log_warn "npm i -g @kweaver-ai/kweaver-admin failed (registry, proxy, or EACCES)."
        return 0
    fi
    onboard_prepend_npm_global_bin_to_path
    if command -v kweaver-admin &>/dev/null; then
        onboard_log_info "kweaver-admin: $(kweaver-admin --version 2>/dev/null | head -n1)"
    else
        onboard_log_warn "kweaver-admin still not on PATH. In this shell:  export PATH=\"\$(npm config get prefix 2>/dev/null)/bin:\$PATH\""
    fi
}

# Full ISF: kweaver (SDK) and kweaver-admin are separate logins. After kweaver auth, ensure admin CLI can list users before [test] + Context Loader.
# On ISF, kweaver-admin must be on PATH and authenticated; otherwise the rest of onboard cannot succeed — exit 1 (no skip).
onboard_ensure_kweaver_admin_auth_for_isf() {
    if ! (type onboard_isf_full_install &>/dev/null && onboard_isf_full_install 2>/dev/null); then
        return 0
    fi
    onboard_prepend_npm_global_bin_to_path
    if ! command -v kweaver-admin &>/dev/null; then
        onboard_log_err "ISF (full) install: kweaver-admin is not on PATH. Install: npm i -g @kweaver-ai/kweaver-admin, add npm global bin to PATH, then re-run. (Unset ONBOARD_SKIP_KWEAVER_ADMIN_INSTALL if that blocked the install step.)"
        exit 1
    fi
    if kweaver-admin --json user list --limit 1 &>/dev/null; then
        onboard_log_info "kweaver-admin: authenticated (user list ok)."
        return 0
    fi
    onboard_log_warn "ISF (full install):  kweaver  and  kweaver-admin  are two different logins — two different saved sessions. The sign-in you just did only applies to  kweaver  (the SDK), not to  kweaver-admin  (user/role management)."
    onboard_log_warn "Next, sign in to  kweaver-admin  the same way as  kweaver  (HTTP). User:  ${ONBOARD_DEFAULT_KWEAVER_USER:-admin} ; password: the same as the web console (factory default is often  ${ONBOARD_DEFAULT_KWEAVER_PASSWORD:-eisoo.com}  if you did not change it). After that, this script can create  test  and re-login  kweaver  as  test  for Context Loader / ADP import."
    local _url _defu _go
    if [[ "${ONBOARD_ASSUME_YES}" == "true" ]]; then
        _defu="$(onboard_default_access_base_url 2>/dev/null || true)"
        if [[ -z "${_defu}" ]]; then
            onboard_log_err "ISF: set ONBOARD_DEFAULT_ACCESS_BASE=... to your platform URL, or re-run in a TTY. kweaver-admin sign-in is required; cannot continue (-y, non-interactive)."
            exit 1
        fi
        onboard_log_info "kweaver-admin: ISF — HTTP sign-in (same defaults as kweaver: ${ONBOARD_DEFAULT_KWEAVER_USER:-admin})…"
        if ! onboard_kweaver_admin_auth_login_for_url "${_defu}"; then
            onboard_log_err "kweaver-admin: HTTP sign-in failed. Check URL, user ${ONBOARD_DEFAULT_KWEAVER_USER:-admin}, and password, then re-run: $0"
            exit 1
        fi
        if ! kweaver-admin --json user list --limit 1 &>/dev/null; then
            onboard_log_err "kweaver-admin: sign-in did not work (user list still fails). Fix credentials or platform, then re-run: $0"
            exit 1
        fi
        onboard_log_info "kweaver-admin: authenticated (user list ok, -y)."
        return 0
    fi
    if ! onboard_is_bootstrap_tty; then
        onboard_log_err "ISF: kweaver-admin is not signed in, and this is not a TTY — cannot prompt. Run kweaver-admin auth in this shell, or: $0 -y  (set ONBOARD_DEFAULT_ACCESS_BASE=... for HTTP). Cannot continue."
        exit 1
    fi
    _defu="$(onboard_default_access_base_url 2>/dev/null || true)"
    echo ""
    read -r -p "kweaver-admin access base URL [Enter = ${_defu}]: " _url
    _url="${_url:-${_defu}}"
    if [[ -z "${_url}" ]]; then
        onboard_log_err "ISF: kweaver-admin sign-in needs a non-empty access base URL. Re-run: $0"
        exit 1
    fi
    if ! onboard_kweaver_admin_auth_login_for_url "${_url}"; then
        onboard_log_err "kweaver-admin: sign-in failed. Fix the error above, then re-run: $0"
        exit 1
    fi
    if ! kweaver-admin --json user list --limit 1 &>/dev/null; then
        onboard_log_err "kweaver-admin: sign-in did not work (user list still fails). Re-check, then re-run: $0"
        exit 1
    fi
    onboard_log_info "kweaver-admin: login OK — next: user [test], then kweaver CLI as test, then Context Loader."
    return 0
}

onboard_probe() {
    onboard_ensure_kweaver_auth
    if [[ "${INTERACTIVE}" == "true" ]]; then
        if ! kubectl get ns &>/dev/null; then
            onboard_log_err "kubectl: cannot list namespaces (check KUBECONFIG / cluster access)"
            exit 1
        fi
    else
        if ! kubectl get "namespace/${NAMESPACE}" &>/dev/null; then
            onboard_log_err "kubectl: namespace ${NAMESPACE} not found (export NAMESPACE=...)"
            exit 1
        fi
    fi
    if [[ "${INTERACTIVE}" == "true" ]]; then
        onboard_log_info "OK: kweaver + kubectl (enter namespace in the next prompt if not ${NAMESPACE})"
    else
        onboard_log_info "OK: kweaver + kubectl (ns=${NAMESPACE})"
    fi
    onboard_prepend_npm_global_bin_to_path
    onboard_recommend_admin_cli
    onboard_ensure_kweaver_admin_for_isf
    onboard_prepend_npm_global_bin_to_path
    onboard_ensure_kweaver_admin_auth_for_isf
    onboard_offer_isf_test_user
    onboard_offer_context_loader_toolset
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
            onboard_log_info "kweaver-admin: not on initial PATH; prepended npm global bin. If still missing, the next step may install or show hints. For full install user ops:  kweaver-admin auth login <url> -u admin -p '<password>' (-k only for https:// self-signed)."
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
    ONBOARD_REPORT_MAIN_MODE="bkn-only"
    onboard_log_info "Done (BKN only)."
    onboard_print_completion_report
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
    if [[ "${ONBOARD_ASSUME_YES}" == "true" ]]; then
        _y_llm_count="$(onboard_get_existing_llm_names | grep -c . || true)"
        _y_sm_count="$(onboard_get_existing_small_model_names | grep -c . || true)"
        _y_bkn_default="$(onboard_bkn_cm_current_default_name "${NAMESPACE}" 2>/dev/null || true)"
        if [[ -n "${_y_bkn_default}" ]]; then
            ONBOARD_REPORT_BKN_CM="skipped — already patched (defaultSmallModelName=${_y_bkn_default})"
        else
            ONBOARD_REPORT_BKN_CM="skipped (-y; not applied without explicit --config)"
        fi
        ONBOARD_REPORT_MODELS="skipped (-y); on platform: ${_y_llm_count} LLM, ${_y_sm_count} small/embedding"
        onboard_log_info "Skipping interactive model registration (-y). On platform: ${_y_llm_count} LLM, ${_y_sm_count} small/embedding. For non-interactive registration, use --config=models.yaml (see -h) or --enable-bkn-search."
        ONBOARD_REPORT_MAIN_MODE="interactive"
        onboard_log_info "Done."
        onboard_print_completion_report
        exit 0
    fi
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

    _POSTI_EXISTING_LLM="$(onboard_get_existing_llm_names)"
    _POSTI_EXISTING_SM="$(onboard_get_existing_small_model_names)"
    _existing_llm_count="$(printf '%s\n' "${_POSTI_EXISTING_LLM}" | grep -c . || true)"
    _existing_sm_count="$(printf '%s\n' "${_POSTI_EXISTING_SM}" | grep -c . || true)"
    _bkn_current_default="$(onboard_bkn_cm_current_default_name "${NAMESPACE}" 2>/dev/null || true)"

    # LLM section: ask whether to add another if any already exist; otherwise prompt directly.
    llm_n=""
    _add_llm="true"
    if [[ "${_existing_llm_count}" -gt 0 ]]; then
        onboard_log_info "LLM already registered on this platform: ${_existing_llm_count}."
        read -r -p "Register another LLM now? [y/N]: " _add_llm_ans
        if [[ ! "${_add_llm_ans}" =~ ^[Yy] ]]; then
            _add_llm="false"
        fi
    fi
    if [[ "${_add_llm}" == "true" ]]; then
        read -r -p "LLM model_name (Enter to skip): " llm_n
        if [[ -n "${llm_n}" ]]; then
            read -r -p "LLM model_series (e.g. deepseek) [others]: " llm_s
            read -r -p "max_model_len [8192]: " llm_ml
            read -r -p "api_key: " -s llm_key
            echo
            read -r -p "api_model: " llm_am
            read -r -p "api_url: " llm_url
            onboard_ensure_llm "${llm_n}" "${llm_s:-others}" "${llm_ml:-8192}" "${llm_key}" "${llm_am}" "${llm_url}" "llm"
        fi
        _POSTI_EXISTING_LLM="$(onboard_get_existing_llm_names)"
        _POSTI_EXISTING_SM="$(onboard_get_existing_small_model_names)"
    fi

    # Embedding / small-model section: same pattern. If a new one is added, ask whether to set as BKN default.
    em_n=""
    bkn_default_name=""
    _new_em_set_default="false"
    _add_em="true"
    if [[ "${_existing_sm_count}" -gt 0 ]]; then
        if [[ -n "${_bkn_current_default}" ]]; then
            onboard_log_info "Embedding / small models already registered: ${_existing_sm_count}. BKN default is currently [${_bkn_current_default}] (defaultSmallModelEnabled=true)."
        else
            onboard_log_info "Embedding / small models already registered: ${_existing_sm_count}. BKN default is not set yet."
        fi
        read -r -p "Register another embedding / small model now? [y/N]: " _add_em_ans
        if [[ ! "${_add_em_ans}" =~ ^[Yy] ]]; then
            _add_em="false"
        fi
    fi
    if [[ "${_add_em}" == "true" ]]; then
        read -r -p "Embedding model_name (Enter to skip): " em_n
        if [[ -n "${em_n}" ]]; then
            read -r -p "api_key: " -s em_key
            echo
            read -r -p "api_model: " em_am
            read -r -p "api_url: " em_url
            read -r -p "embedding_dim [1024]: " em_dim
            if [[ -n "${_bkn_current_default}" ]]; then
                read -r -p "BKN default is currently [${_bkn_current_default}]. Set [${em_n}] as the new BKN default (will patch ConfigMaps and restart bkn-backend / ontology-query)? [y/N]: " em_bkn
                if [[ "${em_bkn}" =~ ^[Yy] ]]; then
                    _new_em_set_default="true"
                fi
            else
                read -r -p "Set [${em_n}] as the BKN default (no default is set yet)? [Y/n]: " em_bkn
                if [[ ! "${em_bkn}" =~ ^[Nn] ]]; then
                    _new_em_set_default="true"
                fi
            fi
            onboard_ensure_small_model "${em_n}" "embedding" "${em_key}" "${em_url}" "${em_am}" 32 512 "${em_dim:-1024}"
            if [[ "${_new_em_set_default}" == "true" ]]; then
                bkn_default_name="${em_n}"
            fi
        fi
    fi
    if [[ -n "${llm_n:-}" ]]; then
        onboard_test_llm "$(onboard_get_id_for_llm "${llm_n}")"
    fi
    if [[ -n "${em_n:-}" ]]; then
        onboard_test_small "$(onboard_get_id_for_small "${em_n}")"
    fi

    _models_status_parts=()
    if [[ -n "${llm_n:-}" ]]; then
        _models_status_parts+=("registered LLM ${llm_n}")
    elif [[ "${_existing_llm_count}" -gt 0 ]]; then
        _models_status_parts+=("LLM unchanged (${_existing_llm_count} on platform)")
    else
        _models_status_parts+=("no LLM entered")
    fi
    if [[ -n "${em_n:-}" ]]; then
        if [[ "${_new_em_set_default}" == "true" ]]; then
            _models_status_parts+=("registered embedding ${em_n} (set as new BKN default)")
        else
            _models_status_parts+=("registered embedding ${em_n}")
        fi
    elif [[ "${_existing_sm_count}" -gt 0 ]]; then
        _models_status_parts+=("embedding/small unchanged (${_existing_sm_count} on platform)")
    else
        _models_status_parts+=("no embedding entered")
    fi
    ONBOARD_REPORT_MODELS=""
    for _p in "${_models_status_parts[@]}"; do
        if [[ -z "${ONBOARD_REPORT_MODELS}" ]]; then
            ONBOARD_REPORT_MODELS="${_p}"
        else
            ONBOARD_REPORT_MODELS="${ONBOARD_REPORT_MODELS}; ${_p}"
        fi
    done

    # If the user did NOT register a new embedding but embeddings already exist on the platform,
    # offer to (re)apply one of them as the BKN default. This covers two cases:
    #   1) BKN default is unset — first-time wiring of an already-registered embedding.
    #   2) BKN default is set — operator wants to switch to a different existing embedding.
    if [[ "${SKIP_BKN}" != "true" && -z "${bkn_default_name}" && "${_existing_sm_count}" -gt 0 ]]; then
        if [[ -z "${_bkn_current_default}" ]]; then
            read -r -p "Use one of the already-registered embeddings as the BKN default (will patch ConfigMaps and restart bkn-backend / ontology-query)? [Y/n]: " _use_existing_em
            _use_existing_em_apply="true"
            if [[ "${_use_existing_em}" =~ ^[Nn] ]]; then
                _use_existing_em_apply="false"
            fi
        else
            read -r -p "BKN default is currently [${_bkn_current_default}]. Switch to a different already-registered embedding? [y/N]: " _use_existing_em
            _use_existing_em_apply="false"
            if [[ "${_use_existing_em}" =~ ^[Yy] ]]; then
                _use_existing_em_apply="true"
            fi
        fi
        if [[ "${_use_existing_em_apply}" == "true" ]]; then
            _existing_em_names="$(onboard_get_existing_small_model_names | sed '/^$/d')"
            if [[ -z "${_existing_em_names}" ]]; then
                onboard_log_info "Could not list existing embedding/small models — skipping ConfigMap patch."
            else
                _default_pick="$(printf '%s\n' "${_existing_em_names}" | head -n1)"
                echo "Available embedding / small models:"
                printf '  - %s\n' ${_existing_em_names}
                read -r -p "Embedding model_name to set as BKN default [${_default_pick}]: " _pick_em
                _pick_em="${_pick_em:-${_default_pick}}"
                if printf '%s\n' "${_existing_em_names}" | grep -Fxq "${_pick_em}"; then
                    bkn_default_name="${_pick_em}"
                else
                    onboard_log_info "[${_pick_em}] not in the existing embedding list — skipping ConfigMap patch."
                fi
            fi
        fi
    fi

    if [[ "${SKIP_BKN}" == "true" ]]; then
        onboard_log_info "Done (skip BKN not used in interactive; omit model to skip BKN patch)."
        ONBOARD_REPORT_BKN_CM="skipped (--skip-bkn)"
    elif [[ -n "${bkn_default_name}" ]]; then
        # User explicitly chose a (new) default — patch and restart, even if CMs were already patched (the name changed).
        onboard_do_bkn_bash "${bkn_default_name}" || exit 1
        if [[ -n "${_bkn_current_default}" && "${_bkn_current_default}" != "${bkn_default_name}" ]]; then
            ONBOARD_REPORT_BKN_CM="re-patched (defaultSmallModelName ${_bkn_current_default} -> ${bkn_default_name}); bkn-backend & ontology-query restarted"
        else
            ONBOARD_REPORT_BKN_CM="patched (defaultSmallModelName=${bkn_default_name}); bkn-backend & ontology-query restarted"
        fi
    elif [[ -n "${_bkn_current_default}" ]]; then
        onboard_log_info "BKN ConfigMaps unchanged — already patched (defaultSmallModelName=${_bkn_current_default}). No restart needed."
        ONBOARD_REPORT_BKN_CM="unchanged — already patched (defaultSmallModelName=${_bkn_current_default})"
    else
        onboard_log_info "No BKN default embedding selected; ConfigMap not patched."
        ONBOARD_REPORT_BKN_CM="skipped (no default embedding selected)"
    fi
    ONBOARD_REPORT_MAIN_MODE="interactive"
    onboard_log_info "Done."
    onboard_print_completion_report
    exit 0
fi

usage
exit 1
