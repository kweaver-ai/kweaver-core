#!/usr/bin/env bash
# =============================================================================
# KWeaver deploy preflight checks (sourced by deploy/preflight.sh)
# =============================================================================

# shellcheck disable=SC2034
PREFLIGHT_OK_COUNT=0
PREFLIGHT_WARN_COUNT=0
PREFLIGHT_FAIL_COUNT=0
PREFLIGHT_FIXED_COUNT=0
PREFLIGHT_FAIL_COUNT_INITIAL=0
PREFLIGHT_DECLINED_COUNT=0
declare -a PREFLIGHT_JSON_OK=()
declare -a PREFLIGHT_JSON_WARN=()
declare -a PREFLIGHT_JSON_FAIL=()
declare -a PREFLIGHT_JSON_FIXED=()
declare -a PREFLIGHT_JSON_DECLINED=()
declare -a PREFLIGHT_FAIL_SNAPSHOT=()

# --- reporting helpers ---------------------------------------------------------
preflight_reset_counters() {
    PREFLIGHT_OK_COUNT=0
    PREFLIGHT_WARN_COUNT=0
    PREFLIGHT_FAIL_COUNT=0
    PREFLIGHT_FIXED_COUNT=0
    PREFLIGHT_FAIL_COUNT_INITIAL=0
    PREFLIGHT_DECLINED_COUNT=0
    PREFLIGHT_JSON_OK=()
    PREFLIGHT_JSON_WARN=()
    PREFLIGHT_JSON_FAIL=()
    PREFLIGHT_JSON_FIXED=()
    PREFLIGHT_JSON_DECLINED=()
    PREFLIGHT_FAIL_SNAPSHOT=()
}

# Append one line to JSONL capture when PREFLIGHT_OUTPUT_JSON=true
_preflight_json_push() {
    local bucket="$1"
    local line="$2"
    [[ "${PREFLIGHT_OUTPUT_JSON:-false}" == "true" ]] || return 0
    case "${bucket}" in
        ok) PREFLIGHT_JSON_OK+=("${line}") ;;
        warn) PREFLIGHT_JSON_WARN+=("${line}") ;;
        fail) PREFLIGHT_JSON_FAIL+=("${line}") ;;
        fixed) PREFLIGHT_JSON_FIXED+=("${line}") ;;
        declined) PREFLIGHT_JSON_DECLINED+=("${line}") ;;
    esac
}

# Backup a file before mutating (idempotent: copy only if file exists)
preflight_backup_file() {
    local f="$1"
    [[ -f "${f}" ]] || return 0
    local bak="${f}.bak.$(date +%s 2>/dev/null || date +%s)"
    cp -a "${f}" "${bak}" 2>/dev/null && log_info "Backed up ${f} -> ${bak}" || true
}

# Resolve vX.Y for pkgs.k8s.io from installed kubeadm (Debian/.deb), or env default.
preflight_resolve_k8s_apt_minor() {
    local dpkg_ver out
    if command -v dpkg-query &>/dev/null; then
        dpkg_ver="$(dpkg-query -W -f='${Version}' kubeadm 2>/dev/null || true)"
    fi
    if [[ -n "${dpkg_ver}" ]]; then
        out="${dpkg_ver%%[-~]*}"
        out="$(echo "${out}" | cut -d. -f1-2)"
        if [[ "${out}" =~ ^[0-9]+\.[0-9]+$ ]]; then
            echo "v${out}"
            return 0
        fi
    fi
    echo "${PREFLIGHT_K8S_APT_MINOR:-v1.28}"
}

# Write JSON to stdout (python3) from PREFLIGHT_JSON_* arrays; human logs go to stderr in JSON mode.
emit_preflight_json() {
    local f
    f="$(mktemp 2>/dev/null || echo /tmp/preflight-json-$$.txt)"
    {
        echo "###OK###"
        for line in "${PREFLIGHT_JSON_OK[@]:-}"; do printf '%s\n' "$line"; done
        echo "###WARN###"
        for line in "${PREFLIGHT_JSON_WARN[@]:-}"; do printf '%s\n' "$line"; done
        echo "###FAIL###"
        for line in "${PREFLIGHT_JSON_FAIL[@]:-}"; do printf '%s\n' "$line"; done
        echo "###FIXED###"
        for line in "${PREFLIGHT_JSON_FIXED[@]:-}"; do printf '%s\n' "$line"; done
        echo "###DECLINED###"
        for line in "${PREFLIGHT_JSON_DECLINED[@]:-}"; do printf '%s\n' "$line"; done
    } > "$f" 2>/dev/null
    if command -v python3 &>/dev/null; then
        python3 -c '
import json, sys
d = {"ok": [], "warn": [], "fail": [], "fixed": [], "declined": []}
kmap = {"###OK###": "ok", "###WARN###": "warn", "###FAIL###": "fail", "###FIXED###": "fixed", "###DECLINED###": "declined"}
cur = "ok"
path = sys.argv[1]
with open(path, encoding="utf-8", errors="replace") as fh:
    for line in fh:
        line = line.rstrip("\n")
        if line in kmap:
            cur = kmap[line]
        elif line and line not in kmap:
            d[cur].append(line)
print(json.dumps(d, ensure_ascii=False))
' "$f"
    else
        echo '{"error":"python3 required for --output=json"}' >&2
    fi
    rm -f "$f" 2>/dev/null || true
}

preflight_report_append() {
    local line="$1"
    if [[ -n "${PREFLIGHT_REPORT_FILE:-}" ]]; then
        echo "${line}" >> "${PREFLIGHT_REPORT_FILE}" 2>/dev/null || true
    fi
}

_preflight_log_line() {
    if [[ "${PREFLIGHT_OUTPUT_JSON:-false}" == "true" ]]; then
        echo -e "$1" >&2
    else
        echo -e "$1"
    fi
}

preflight_ok() {
    local msg="$1"
    _preflight_log_line "${GREEN}[OK]${NC} ${msg}"
    preflight_report_append "[OK] ${msg}"
    PREFLIGHT_OK_COUNT=$((PREFLIGHT_OK_COUNT + 1))
    _preflight_json_push ok "${msg}"
}

preflight_warn() {
    local msg="$1"
    _preflight_log_line "${YELLOW}[WARN]${NC} ${msg}"
    preflight_report_append "[WARN] ${msg}"
    PREFLIGHT_WARN_COUNT=$((PREFLIGHT_WARN_COUNT + 1))
    _preflight_json_push warn "${msg}"
}

preflight_fail() {
    local msg="$1"
    _preflight_log_line "${RED}[FAIL]${NC} ${msg}"
    preflight_report_append "[FAIL] ${msg}"
    PREFLIGHT_FAIL_COUNT=$((PREFLIGHT_FAIL_COUNT + 1))
    PREFLIGHT_FAIL_SNAPSHOT+=("${msg}")
    _preflight_json_push fail "${msg}"
}

preflight_fixed() {
    local msg="$1"
    _preflight_log_line "${GREEN}[FIXED]${NC} ${msg}"
    preflight_report_append "[FIXED] ${msg}"
    PREFLIGHT_FIXED_COUNT=$((PREFLIGHT_FIXED_COUNT + 1))
    _preflight_json_push fixed "${msg}"
}

preflight_declined() {
    local msg="$1"
    _preflight_log_line "${YELLOW}[DECLINED]${NC} ${msg}"
    preflight_report_append "[DECLINED] ${msg}"
    PREFLIGHT_DECLINED_COUNT=$((PREFLIGHT_DECLINED_COUNT + 1))
    _preflight_json_push declined "${msg}"
}

# --- skip set -----------------------------------------------------------------
preflight_skip() {
    local name="$1"
    [[ "${PREFLIGHT_SKIP_SET:-}" == *"|${name}|"* ]]
}

# --- per-fix confirmation -----------------------------------------------------
# Ask the user before applying a fix. Honors:
#   PREFLIGHT_ASSUME_YES=true   auto-yes for ALL fixes (used by -y / --yes)
#   PREFLIGHT_ASSUME_NO=true    auto-no for ALL fixes (dry-run-style)
#   PREFLIGHT_FIX_ALLOW         pipe-separated allowlist; if non-empty, only
#                               fixes whose name is in the list run automatically.
# Returns 0 if user (or env) approved the fix, 1 otherwise.
# Args: <fix-name> <one-line action description> <one-line risk description>
preflight_confirm_fix() {
    local name="$1"
    local action="$2"
    local risk="$3"

    if [[ "${PREFLIGHT_LIST_FIXES_ONLY:-false}" == "true" ]]; then
        log_info "Would offer fix: ${name} — ${action}"
        return 1
    fi

    if [[ "${PREFLIGHT_OUTPUT_JSON:-false}" == "true" ]]; then
        echo -e "" >&2
        echo -e "${YELLOW}[FIX?]${NC} ${name}" >&2
        echo "  Action: ${action}" >&2
        echo "  Risk:   ${risk}" >&2
    else
        echo ""
        echo -e "${YELLOW}[FIX?]${NC} ${name}"
        echo "  Action: ${action}"
        echo "  Risk:   ${risk}"
    fi
    if [[ -n "${PREFLIGHT_REPORT_FILE:-}" ]]; then
        {
            echo "[FIX?] ${name}"
            echo "  Action: ${action}"
            echo "  Risk: ${risk}"
        } >> "${PREFLIGHT_REPORT_FILE}" 2>/dev/null || true
    fi

    if [[ "${PREFLIGHT_ASSUME_NO:-false}" == "true" ]]; then
        preflight_report_append "[DECLINED?] ${name} (PREFLIGHT_ASSUME_NO)"
        return 1
    fi

    if [[ -n "${PREFLIGHT_FIX_ALLOW:-}" ]]; then
        if [[ "${PREFLIGHT_FIX_ALLOW}" == *"|${name}|"* ]]; then
            preflight_report_append "[APPROVE] ${name} (--fix-allow)"
            return 0
        else
            preflight_report_append "[DECLINED?] ${name} (not in --fix-allow)"
            return 1
        fi
    fi

    if [[ "${PREFLIGHT_ASSUME_YES:-false}" == "true" ]]; then
        preflight_report_append "[APPROVE] ${name} (-y / --yes)"
        return 0
    fi

    if [[ ! -t 0 ]] && [[ ! -e /dev/tty ]]; then
        if [[ "${PREFLIGHT_OUTPUT_JSON:-false}" == "true" ]]; then
            log_warn "  -> no TTY; skipping. Re-run with -y or --fix-allow=${name}." >&2
        else
            log_warn "  -> no TTY for confirmation; skipping. Re-run with -y to apply, or --fix-allow=${name}."
        fi
        return 1
    fi

    set +e
    local reply=""
    if [[ -e /dev/tty ]]; then
        read -r -p "  Apply this fix? [y/N]: " reply </dev/tty
    else
        read -r -p "  Apply this fix? [y/N]: " reply
    fi
    local r=$?
    set -e
    if [[ $r -ne 0 ]]; then
        reply=""
    fi
    case "${reply}" in
        y|Y|yes|YES)
            preflight_report_append "[APPROVE] ${name} (interactive)"
            return 0
            ;;
        *)
            if [[ -n "${PREFLIGHT_REPORT_FILE:-}" ]]; then
                echo "[DECLINED?] ${name} (user)" >> "${PREFLIGHT_REPORT_FILE}" 2>/dev/null || true
            fi
            return 1
            ;;
    esac
}

# --- hardware ----------------------------------------------------------------
preflight_check_hardware() {
    preflight_skip "hardware" && return 0
    log_info "Checking CPU / memory / disk..."

    local cpu nproc_val mem_mb
    nproc_val="$(nproc 2>/dev/null || getconf _NPROCESSORS_ONLN 2>/dev/null || echo 0)"
    cpu="${nproc_val:-0}"
    if [[ "${cpu}" -ge 16 ]]; then
        preflight_ok "CPU cores: ${cpu} (>= 16)"
    else
        preflight_warn "CPU cores: ${cpu} (recommended >= 16; kubeadm ignores NumCPU for init)"
    fi

    if command -v free &>/dev/null; then
        mem_mb="$(free -m 2>/dev/null | awk '/^Mem:/ {print $2}')"
    else
        mem_mb="0"
    fi
    if [[ -n "${mem_mb}" && "${mem_mb}" -ge 47104 ]]; then
        preflight_ok "Memory: ${mem_mb} MB (>= 48GB)"
    elif [[ -n "${mem_mb}" && "${mem_mb}" -ge 1 ]]; then
        preflight_warn "Memory: ${mem_mb} MB (recommended >= 48GB)"
    else
        preflight_warn "Could not read memory (free not available?)"
    fi

    for mp in / /var; do
        local line avail_mib
        if df -P "${mp}" &>/dev/null; then
            line="$(df -Pk "${mp}" 2>/dev/null | tail -1 || true)"
        else
            line="$(df -k "${mp}" 2>/dev/null | tail -1 || true)"
        fi
        # Column 4 = avail K-blocks on GNU & BSD df -k
        avail_mib="$(echo "${line}" | awk '{print int($4/1024)}')"
        if [[ -n "${avail_mib}" && "${avail_mib}" -ge 204800 ]]; then
            preflight_ok "Disk free on ${mp}: ${avail_mib} MiB (>= 200GB)"
        elif [[ -n "${avail_mib}" ]]; then
            preflight_warn "Disk free on ${mp}: ${avail_mib} MiB (recommended >= 200GB free)"
        else
            preflight_warn "Could not parse disk free for ${mp}"
        fi
    done
}

# --- OS / kernel ---------------------------------------------------------------
preflight_check_os() {
    preflight_skip "os" && return 0
    log_info "Checking OS and kernel..."

    if [[ ! -f /etc/os-release ]]; then
        preflight_warn "No /etc/os-release (expected on RHEL/Debian/openEuler; macOS/others: run on Linux target host)"
        return
    fi
    # shellcheck source=/dev/null
    . /etc/os-release

    local id_like="${ID_LIKE:-}"
    local ok_os="no"
    case "${ID:-}" in
        centos|rhel|almalinux|rocky) [[ "${VERSION_ID%%.*}" -ge 8 ]] 2>/dev/null && ok_os="yes" || true ;;
        openeuler) [[ "${VERSION_ID%%.*}" -ge 23 ]] 2>/dev/null && ok_os="yes" || true ;;
        ubuntu) [[ "${VERSION_ID%%.*}" -ge 22 ]] 2>/dev/null && ok_os="yes" || true ;;
        *) true ;;
    esac
    if [[ "${ok_os}" == "yes" ]]; then
        preflight_ok "OS: ${ID:-unknown} ${VERSION_ID:-} (in supported set)"
    else
        preflight_warn "OS: ${ID:-unknown} ${VERSION_ID:-} (expected CentOS 8+ / openEuler 23+ / Ubuntu 22.04+); verify before production"
    fi

    local kver
    kver="$(uname -r 2>/dev/null | cut -d. -f1-2)"
    # 4.18 -> compare as 4.18.0
    local kmajor kminor
    kmajor="$(uname -r | cut -d. -f1)"
    kminor="$(uname -r | cut -d. -f2 | cut -d- -f1)"
    if [[ "${kmajor}" -gt 4 ]] || { [[ "${kmajor}" -eq 4 && "${kminor}" -ge 18 ]]; }; then
        preflight_ok "Kernel: $(uname -r) (>= 4.18)"
    else
        preflight_warn "Kernel: $(uname -r) (recommended >= 4.18 for Kubernetes containerd path)"
    fi
}

# --- hostname / hosts ----------------------------------------------------------
preflight_check_hostname_hosts() {
    preflight_skip "hostname" && return 0
    log_info "Checking hostname and /etc/hosts..."

    local h
    h="$(hostname 2>/dev/null || true)"
    if echo "${h}" | grep -qE '[_A-Z]'; then
        preflight_warn "Hostname contains uppercase or underscore: ${h} (K8s best practice: lowercase, DNS-1123 labels)"
    else
        preflight_ok "Hostname: ${h}"
    fi

    if [[ -f /etc/hosts ]] && grep -qE "127\.0\.0\.1[[:space:]]+${h}" /etc/hosts 2>/dev/null; then
        preflight_ok "/etc/hosts has 127.0.0.1 ${h}"
    elif [[ -f /etc/hosts ]] && grep -qE '127\.0\.0\.1[[:space:]]+localhost' /etc/hosts; then
        preflight_warn "Consider: echo '127.0.0.1 ${h}' >> /etc/hosts (safe-fix may add this)"
    else
        preflight_warn "Review /etc/hosts for 127.0.0.1 and hostname mapping"
    fi
}

# --- swap / selinux (inspect only) --------------------------------------------
preflight_check_swap_selinux() {
    preflight_skip "swap" && return 0
    log_info "Checking swap and SELinux..."

    if swapon --show 2>/dev/null | grep -q .; then
        preflight_warn "Swap is active; deploy will disable (or run with --fix)"
    else
        preflight_ok "No active swap"
    fi

    if command -v getenforce &>/dev/null; then
        local se
        se="$(getenforce 2>/dev/null || true)"
        if [[ "${se}" == "Enforcing" ]]; then
            preflight_warn "SELinux is Enforcing; deploy scripts typically set disabled/permissive for K8s"
        else
            preflight_ok "SELinux: ${se}"
        fi
    else
        preflight_ok "SELinux tools not present (assumed not applicable)"
    fi
}

# --- firewall ------------------------------------------------------------------
preflight_check_firewall() {
    preflight_skip "firewall" && return 0
    log_info "Checking local firewall..."

    if systemctl is-active --quiet firewalld 2>/dev/null; then
        preflight_warn "firewalld is active; recommend stop/disable for one-node install (or open required ports)"
    else
        preflight_ok "firewalld is not active (or not installed)"
    fi

    if command -v ufw &>/dev/null; then
        if ufw status 2>/dev/null | grep -qi "Status: active"; then
            preflight_warn "ufw is active; ensure 6443, 80/443, NodePort range are allowed"
        fi
    fi
}

# --- sysctl / modules (inspect) ----------------------------------------------
preflight_check_sysctl_modules() {
    preflight_skip "sysctl" && return 0
    log_info "Checking IP forward and kernel modules..."

    local ipf
    ipf="$(cat /proc/sys/net/ipv4/ip_forward 2>/dev/null || echo 0)"
    if [[ "${ipf}" == "1" ]]; then
        preflight_ok "net.ipv4.ip_forward=1"
    else
        preflight_warn "net.ipv4.ip_forward is ${ipf} (K8s needs forwarding; configure_system will set)"
    fi

    for mod in br_netfilter overlay; do
        if lsmod 2>/dev/null | awk -v m="${mod}" '$1==m {f=1} END{exit !f}'; then
            preflight_ok "Kernel module loaded: ${mod}"
        else
            preflight_warn "Kernel module not loaded: ${mod} (will be loaded on fix/install)"
        fi
    done
}

# --- chrony / time -------------------------------------------------------------
preflight_check_time_sync() {
    preflight_skip "time" && return 0
    log_info "Checking time sync..."

    if systemctl is-active --quiet chronyd 2>/dev/null; then
        preflight_ok "chronyd is active"
    elif systemctl is-active --quiet ntpd 2>/dev/null; then
        preflight_ok "ntpd is active"
    elif systemctl is-active --quiet systemd-timesyncd 2>/dev/null; then
        preflight_ok "systemd-timesyncd is active"
    else
        preflight_warn "No common time sync service active (recommend chrony/ntp for TLS and logs)"
    fi
}

# --- cgroup version ------------------------------------------------------------
preflight_check_cgroup() {
    preflight_skip "cgroup" && return 0
    log_info "Checking cgroup..."

    if [[ -f /sys/fs/cgroup/cgroup.controllers ]]; then
        preflight_ok "cgroup v2 is present (/sys/fs/cgroup/cgroup.controllers); ensure containerd uses systemd cgroup driver"
    elif [[ -d /sys/fs/cgroup/net_cls ]]; then
        preflight_ok "cgroup v1 layout detected (supported)"
    else
        preflight_warn "Could not determine cgroup version"
    fi
}

# --- P0: architecture ---------------------------------------------------------
preflight_check_arch() {
    preflight_skip "arch" && return 0
    log_info "Checking CPU architecture..."
    local m
    m="$(uname -m 2>/dev/null || echo unknown)"
    case "${m}" in
        x86_64|amd64)
            preflight_ok "Architecture: ${m} (supported)"
            ;;
        aarch64|arm64)
            if [[ "${PREFLIGHT_REQUIRE_AMD64:-false}" == "true" ]]; then
                preflight_fail "Architecture: ${m}; PREFLIGHT_REQUIRE_AMD64=true (need x86_64/amd64 images)"
            else
                preflight_warn "Architecture: ${m} (verify KWeaver image availability for your platform)"
            fi
            ;;
        *)
            preflight_warn "Architecture: ${m} (verify platform support before production)"
            ;;
    esac
}

# --- P0: proxy / no_proxy ------------------------------------------------------
preflight_check_proxy() {
    preflight_skip "proxy" && return 0
    log_info "Checking HTTP(S) proxy and NO_PROXY..."
    local p="${https_proxy:-${HTTPS_PROXY:-}}${http_proxy:-${HTTP_PROXY:-}}"
    if [[ -z "${p}" ]]; then
        preflight_ok "No HTTP/HTTPS proxy environment variables set"
        return
    fi
    local np="${no_proxy:-${NO_PROXY:-}}"
    local need_fail=false
    local need
    for need in 127.0.0.1 localhost .svc .cluster.local; do
        if [[ "${np}" != *"${need}"* ]]; then
            preflight_fail "NO_PROXY is missing required entry '${need}' (current: ${np:-<empty>}). In-cluster and TLS break without it when proxy is set."
            need_fail=true
        fi
    done
    if [[ "${need_fail}" == "false" ]]; then
        preflight_ok "Proxy set and NO_PROXY contains basic Kubernetes exemptions"
    fi
}

# --- P0: DNS (getent) ----------------------------------------------------------
preflight_check_dns() {
    preflight_skip "dns" && return 0
    log_info "Checking DNS name resolution (sample hosts)..."
    if ! command -v getent &>/dev/null; then
        preflight_warn "getent not found; skipping DNS resolution checks"
        return
    fi
    local h okc=0
    for h in swr.cn-east-3.myhuaweicloud.com kweaver-ai.github.io; do
        if getent hosts "${h}" &>/dev/null; then
            preflight_ok "DNS: ${h} resolves"
            okc=$((okc + 1))
        else
            preflight_warn "DNS: ${h} does not resolve (check /etc/resolv.conf and corporate DNS)"
        fi
    done
    if [[ -f /etc/resolv.conf ]] && grep -qE '^nameserver[[:space:]]+127\.0\.0\.53' /etc/resolv.conf 2>/dev/null; then
        if ! command -v resolvectl &>/dev/null; then
            preflight_warn "resolv.conf uses 127.0.0.53 (systemd-resolved); ensure upstream DNS is configured"
        elif ! resolvectl status 2>/dev/null | grep -qE 'DNS Server'; then
            preflight_warn "systemd-resolved active but no upstream DNS visible (resolvectl status); CoreDNS may fail to reach upstreams"
        fi
    fi
}

# --- P0: kubeadm binary dependencies ------------------------------------------
preflight_check_kubeadm_deps() {
    preflight_skip "kubeadm-deps" && return 0
    if [[ "$(uname -s)" != "Linux" ]]; then
        preflight_ok "kubeadm tools (skip on non-Linux; use preflight on the install host)"
        return
    fi
    log_info "Checking kubeadm network utilities..."
    local miss=() t
    for t in conntrack socat ebtables ethtool ipset iptables; do
        if ! command -v "${t}" &>/dev/null; then
            miss+=("${t}")
        fi
    done
    if [[ ${#miss[@]} -gt 0 ]]; then
        preflight_warn "Optional kubeadm tools not found: ${miss[*]}. Not required by deploy.sh, but upstream Kubernetes docs recommend them (apt install conntrack socat ebtables ethtool ipset)"
    else
        preflight_ok "kubeadm dependency tools present (conntrack, socat, ebtables, ethtool, ipset, iptables)"
    fi
}

# --- P0: bridge-nf sysctls (when modules may be loaded) ------------------------
preflight_check_bridge_sysctl() {
    preflight_skip "bridge" && return 0
    # Only applicable after br_netfilter is loadable/loaded; check proc entries if they exist
    if [[ -f /proc/sys/net/bridge/bridge-nf-call-iptables ]]; then
        local b4 b6
        b4="$(cat /proc/sys/net/bridge/bridge-nf-call-iptables 2>/dev/null || echo 0)"
        b6="$(cat /proc/sys/net/bridge/bridge-nf-call-ip6tables 2>/dev/null || echo 0)"
        if [[ "${b4}" == "1" && "${b6}" == "1" ]]; then
            preflight_ok "bridge-nf-call-iptables=1, bridge-nf-call-ip6tables=1"
        else
            preflight_warn "bridge-nf: iptables=${b4} ip6tables=${b6} (expected 1/1; configure_system or manual sysctl)"
        fi
    else
        preflight_ok "bridge sysctl paths not present yet (br_netfilter not loaded — OK if fresh host)"
    fi
}

# --- P0: kernel / resource limits in sysctl -----------------------------------
preflight_check_kernel_limits() {
    preflight_skip "kernel-limits" && return 0
    if [[ "$(uname -s)" != "Linux" ]]; then
        preflight_ok "Kernel sysctl limits (skip on non-Linux)"
        return
    fi
    log_info "Checking kernel sysctl limits (inotify, vm.max_map_count, pid_max)..."
    local read_max inow inoinst pidm
    read_max="$(cat /proc/sys/vm/max_map_count 2>/dev/null || echo 0)"
    inow="$(cat /proc/sys/fs/inotify/max_user_watches 2>/dev/null || echo 0)"
    inoinst="$(cat /proc/sys/fs/inotify/max_user_instances 2>/dev/null || echo 0)"
    pidm="$(cat /proc/sys/kernel/pid_max 2>/dev/null || echo 0)"
    if [[ "${read_max}" -ge 262144 ]]; then
        preflight_ok "vm.max_map_count=${read_max} (>= 262144 for OpenSearch/ES style workloads)"
    else
        preflight_warn "vm.max_map_count=${read_max} (recommended >= 262144)"
    fi
    if [[ "${inow}" -ge 524288 ]]; then
        preflight_ok "fs.inotify.max_user_watches=${inow}"
    else
        preflight_warn "fs.inotify.max_user_watches=${inow} (recommended >= 524288 on busy nodes)"
    fi
    if [[ "${inoinst}" -ge 8192 ]]; then
        preflight_ok "fs.inotify.max_user_instances=${inoinst}"
    else
        preflight_warn "fs.inotify.max_user_instances=${inoinst} (recommended >= 8192 for K8s nodes; default 128 causes 'Too many open files' for systemd/journalctl/kubelet/containerd)"
    fi
    if [[ "${pidm}" -ge 4194304 ]]; then
        preflight_ok "kernel.pid_max=${pidm}"
    else
        preflight_warn "kernel.pid_max=${pidm} (recommended >= 4194304 on large clusters)"
    fi
}

# --- P0: ulimits ----------------------------------------------------------------
preflight_check_ulimits() {
    preflight_skip "ulimits" && return 0
    log_info "Checking ulimits (nofile)..."
    local soft hard
    soft="$(ulimit -Sn 2>/dev/null || echo 0)"
    hard="$(ulimit -Hn 2>/dev/null || echo 0)"
    if [[ "${soft}" =~ ^[0-9]+$ ]] && [[ "${soft}" -ge 65536 ]]; then
        preflight_ok "ulimit -n soft=${soft} (>= 65536)"
    else
        preflight_warn "ulimit -n soft=${soft} (recommended >= 65536; systemd may raise for kubelet/containerd)"
    fi
    if [[ "${hard}" =~ ^[0-9]+$ ]] && [[ "${hard}" -ge 65536 ]]; then
        preflight_ok "ulimit -n hard=${hard}"
    else
        preflight_warn "ulimit -n hard=${hard} (recommended >= 65536)"
    fi
}

# --- P0: Kubernetes API server version (existing cluster) ---------------------
preflight_check_k8s_version() {
    preflight_skip "k8s-version" && return 0
    log_info "Checking Kubernetes control plane version (if cluster reachable)..."
    if ! command -v kubectl &>/dev/null; then
        preflight_ok "kubectl not installed — skipping cluster version check"
        return
    fi
    if ! kubectl cluster-info &>/dev/null; then
        preflight_ok "No reachable cluster — skipping server version check"
        return
    fi
    local ver maj min
    ver="$(kubectl version -o json 2>/dev/null | python3 -c "import json,sys; d=json.load(sys.stdin); print(d.get('serverVersion',{}).get('gitVersion',''))" 2>/dev/null || true)"
    if [[ -z "${ver}" ]]; then
        ver="$(kubectl version --short 2>/dev/null | awk '/Server Version/{print $NF}' | tr -d 'v' || true)"
    fi
    if [[ -z "${ver}" ]]; then
        preflight_warn "Could not parse Kubernetes server version"
        return
    fi
    maj="${ver#v}"; maj="${maj%%.*}"; min="${ver#v}"; min="${min#*.}"; min="${min%%.*}"
    if ! [[ "${maj}" =~ ^[0-9]+$ && "${min}" =~ ^[0-9]+$ ]]; then
        preflight_warn "Unexpected version string: ${ver}"
        return
    fi
    if [[ "$((maj*100+min))" -lt 124 ]]; then
        preflight_fail "Kubernetes server too old: ${ver} (minimum supported 1.24 for this preflight policy)"
    elif [[ "$((maj*100+min))" -ge 132 ]]; then
        preflight_fail "Kubernetes server very new: ${ver} (verify KWeaver chart compatibility; not validated beyond 1.31 here)"
    elif [[ "$((maj*100+min))" -lt 126 || "$((maj*100+min))" -gt 130 ]]; then
        preflight_warn "Kubernetes server version ${ver} (recommended 1.26.x–1.30.x for this track)"
    else
        preflight_ok "Kubernetes server version ${ver} (in recommended range)"
    fi
}

# --- P0: pod / service CIDR vs host routes -----------------------------------
preflight_check_cidr_conflict() {
    preflight_skip "cidr" && return 0
    log_info "Checking pod/service CIDR vs local routes (if cluster or kubeadm config present)..."
    local pod_cidr="10.244.0.0/16" svc_cidr="10.96.0.0/12"
    if command -v kubectl &>/dev/null && kubectl cluster-info &>/dev/null; then
        local ccfg
        ccfg="$(kubectl -n kube-system get cm kubeadm-config -o jsonpath='{.data.ClusterConfiguration}' 2>/dev/null || true)"
        if [[ -n "${ccfg}" ]]; then
            local pc sc
            pc="$(echo "${ccfg}" | grep -E 'podSubnet:' | head -1 | awk '{print $2}' | tr -d '"')"
            sc="$(echo "${ccfg}" | grep -E 'serviceSubnet:' | head -1 | awk '{print $2}' | tr -d '"')"
            [[ -n "${pc}" ]] && pod_cidr="${pc}"
            [[ -n "${sc}" ]] && svc_cidr="${sc}"
        fi
    fi
    preflight_ok "Assumed/resolved pod CIDR: ${pod_cidr}, service CIDR: ${svc_cidr} (verify they do not overlap with VPC or docker0 route)"
    if command -v ip &>/dev/null; then
        if ip route 2>/dev/null | grep -qE '10\.(244|96)\.'; then
            preflight_warn "Host routes mention 10.244/10.96 ranges — double-check for overlap with CNI and kube-proxy"
        fi
    fi
    if ip -4 addr show docker0 2>/dev/null | grep -q inet; then
        preflight_warn "docker0 interface present; ensure it does not overlap with pod CIDR or disable Docker when using containerd"
    fi
}

# --- P0: disk for /var/lib/containerd -----------------------------------------
preflight_check_containerd_disk() {
    preflight_skip "containerd-disk" && return 0
    log_info "Checking free space for /var/lib/containerd..."
    local line avail_mib target="/var/lib/containerd"
    [[ -d "${target}" ]] || target="/var"
    if df -Pk "${target}" &>/dev/null; then
        line="$(df -Pk "${target}" 2>/dev/null | tail -1 || true)"
    else
        line="$(df -k "${target}" 2>/dev/null | tail -1 || true)"
    fi
    avail_mib="$(echo "${line}" | awk '{print int($4/1024)}')"
    if [[ -n "${avail_mib}" && "${avail_mib}" -ge 102400 ]]; then
        preflight_ok "Disk free for ${target}: ${avail_mib} MiB (>= 100 GiB for images)"
    elif [[ -n "${avail_mib}" ]]; then
        preflight_warn "Disk free for ${target}: ${avail_mib} MiB (recommended >= 100 GiB for container images)"
    else
        preflight_warn "Could not read free space for ${target}"
    fi
}

# --- P1: Docker still present / socket -----------------------------------------
preflight_check_docker_residue() {
    preflight_skip "docker" && return 0
    log_info "Checking Docker / dockershim residue vs containerd..."
    if systemctl is-active --quiet docker 2>/dev/null; then
        preflight_fail "Docker service is active; stop/disable Docker when using containerd for Kubernetes (or remove duplicate runtime)"
    elif [[ -S /var/run/docker.sock ]]; then
        preflight_fail "Docker socket /var/run/docker.sock exists; remove Docker or the socket to avoid CRI conflicts"
    else
        preflight_ok "No active Docker service / docker.sock"
    fi
}

# --- P1: containerd systemd cgroup driver --------------------------------------
preflight_check_containerd_cgroup_driver() {
    preflight_skip "containerd-cgroup" && return 0
    if ! command -v containerd &>/dev/null; then
        return 0
    fi
    local cf="/etc/containerd/config.toml"
    if [[ -f "${cf}" ]] && grep -qE 'SystemdCgroup[[:space:]]*=[[:space:]]*true' "${cf}" 2>/dev/null; then
        preflight_ok "containerd config: SystemdCgroup = true"
    else
        preflight_warn "containerd SystemdCgroup = true not found in ${cf} (required on systemd+cgroupv2; fix may write default config)"
    fi
}

# --- P1: iptables backend (Debian family) --------------------------------------
preflight_check_iptables_backend() {
    preflight_skip "iptables" && return 0
    if ! command -v update-alternatives &>/dev/null; then
        return 0
    fi
    local line
    line="$(update-alternatives --display iptables 2>/dev/null | head -3 || true)"
    if echo "${line}" | grep -qi nft; then
        preflight_warn "iptables may be using nft backend; kube-proxy iptables mode may need legacy (see: update-alternatives / preflight fix)"
    else
        preflight_ok "iptables alternative looks compatible (or not nft-only)"
    fi
}

# --- P1: NTP / chrony offset ----------------------------------------------------
preflight_check_ntp_drift() {
    preflight_skip "ntp-drift" && return 0
    if ! command -v chronyc &>/dev/null; then
        return 0
    fi
    if ! systemctl is-active --quiet chronyd 2>/dev/null; then
        return 0
    fi
    local off
    off="$(chronyc tracking 2>/dev/null | awk -F: '/^Last offset/ {gsub(/ +/,"",$2); gsub(/s$/,"",$2); print $2}' | head -1 || true)"
    if [[ -n "${off}" && "${off}" != "0" ]]; then
        # rough abs compare with awk
        if awk -v o="${off}" 'BEGIN{exit !(o+0>0.5 || o+0<-0.5)}'; then
            preflight_warn "chrony last offset is ${off}s (>|0.5s; check upstream NTP reachability)"
        else
            preflight_ok "chrony last offset: ${off}s (acceptable)"
        fi
    else
        preflight_ok "chronyc tracking: could not read offset (or zero)"
    fi
}

# --- P1: systemd + cgroup v2 ---------------------------------------------------
preflight_check_systemd_version() {
    preflight_skip "systemd-version" && return 0
    if ! command -v systemctl &>/dev/null; then
        return 0
    fi
    local ver
    ver="$(systemctl --version 2>/dev/null | head -1 | awk '{print $2}' | cut -d'~' -f1 || true)"
    if [[ -f /sys/fs/cgroup/cgroup.controllers ]] && [[ -n "${ver}" ]]; then
        local m="${ver%%.*}"
        if [[ "${m}" =~ ^[0-9]+$ ]] && [[ "${m}" -lt 244 ]]; then
            preflight_warn "systemd ${ver} on cgroup v2: versions < 244 can have issues with Kubernetes; prefer systemd 244+"
        else
            preflight_ok "systemd ${ver} with cgroup v2"
        fi
    fi
}

# --- P1: existing Helm / namespaces --------------------------------------------
preflight_check_existing_release() {
    preflight_skip "helm-releases" && return 0
    if ! command -v helm &>/dev/null; then
        return 0
    fi
    if ! command -v kubectl &>/dev/null || ! kubectl cluster-info &>/dev/null; then
        return 0
    fi
    log_info "Checking for existing kweaver/isf/dip-related Helm releases..."
    local r
    r="$(helm list -A 2>/dev/null | grep -iE 'kweaver|isf|dip' || true)"
    if [[ -n "${r}" ]]; then
        preflight_warn "Helm may already have related releases (reuse ok only if intended):\n${r}"
    else
        preflight_ok "No obvious kweaver/isf/dip Helm release names in helm list -A"
    fi
    local tns
    tns="$(kubectl get ns 2>/dev/null | awk '$2=="Terminating"{print $1}' || true)"
    if [[ -n "${tns}" ]]; then
        preflight_fail "Namespaces stuck in Terminating: ${tns} — resolve before install"
    fi
}

# --- P1: extra ports (etcd, scheduler, extra ingress) -------------------------
preflight_check_extended_ports() {
    preflight_skip "ports-ext" && return 0
    if ! command -v ss &>/dev/null; then
        return 0
    fi
    log_info "Checking additional Kubernetes / ingress related ports..."
    local p
    for p in 2379 2380 10257 10259 "${INGRESS_NGINX_NODEPORT_HTTP:-30080}" "${INGRESS_NGINX_NODEPORT_HTTPS:-30443}"; do
        if ss -H -lnt "sport = :${p}" 2>/dev/null | grep -q .; then
            preflight_ok "Port ${p} in use (expected if those components are running)"
        else
            preflight_ok "Port ${p} not listening (OK for fresh node)"
        fi
    done
}

# --- P1: node allocatable (single-node heuristic) ------------------------------
preflight_check_node_capacity() {
    preflight_skip "node-capacity" && return 0
    if ! command -v kubectl &>/dev/null || ! kubectl get nodes &>/dev/null; then
        return 0
    fi
    log_info "Checking node allocatable resources (heuristic)..."
    local cpu_s mem_s cpu_cores
    cpu_s="$(kubectl get nodes -o jsonpath='{.items[0].status.allocatable.cpu}' 2>/dev/null || true)"
    mem_s="$(kubectl get nodes -o jsonpath='{.items[0].status.allocatable.memory}' 2>/dev/null || true)"
    if echo "${cpu_s}" | grep -qE 'm$'; then
        cpu_cores="$(echo "${cpu_s}" | tr -d 'm')"
        cpu_cores=$((cpu_cores / 1000))
    else
        cpu_cores="${cpu_s%%.*}"
    fi
    if [[ -n "${cpu_cores}" && "${cpu_cores}" =~ ^[0-9]+$ ]] && [[ "${cpu_cores}" -ge 12 ]]; then
        preflight_ok "Node allocatable CPU ${cpu_s} (>= 12 cores equivalent; single-node heuristic)"
    elif [[ -n "${cpu_cores}" ]]; then
        preflight_warn "Node allocatable CPU ${cpu_s} (recommended >= 12 cores for a comfortable single-node install)"
    fi
    if [[ -n "${mem_s}" ]]; then
        preflight_ok "Node allocatable memory: ${mem_s} (verify vs chart requests)"
    fi
}

# --- P1: offline bundle assets -------------------------------------------------
preflight_check_offline_assets() {
    preflight_skip "offline" && return 0
    if [[ "${DEPLOY_OFFLINE:-false}" != "true" && "${OFFLINE:-false}" != "true" ]]; then
        return 0
    fi
    log_info "Checking offline deploy assets (DEPLOY_OFFLINE / OFFLINE)..."
    local root="${PREFLIGHT_ROOT:-.}"
    local has_img=0 has_chart=0
    compgen -G "${root}/images/*.tar" &>/dev/null && has_img=1
    compgen -G "${root}/images/*.tar.gz" &>/dev/null && has_img=1
    compgen -G "${root}/charts/"*.tgz &>/dev/null && has_chart=1
    if [[ ${has_img} -eq 0 ]]; then
        preflight_warn "Offline mode: no images/*.tar(.gz) found under ${root}/images/"
    else
        preflight_ok "Offline mode: found image tarballs under ${root}/images/"
    fi
    if [[ ${has_chart} -eq 0 ]]; then
        preflight_warn "Offline mode: no charts/*.tgz found under ${root}/charts/"
    else
        preflight_ok "Offline mode: found chart tgz under ${root}/charts/"
    fi
}

# --- P1: deploy/conf/config.yaml smoke ----------------------------------------
preflight_check_config_yaml() {
    preflight_skip "config-yaml" && return 0
    local cfg="${PREFLIGHT_CONFIG_YAML:-${PREFLIGHT_ROOT:-.}/conf/config.yaml}"
    log_info "Checking ${cfg} (required keys)..."
    if [[ ! -f "${cfg}" ]]; then
        preflight_warn "config.yaml not found at ${cfg} (set PREFLIGHT_CONFIG_YAML or run from deploy/)"
        return
    fi
    if ! grep -qE '^[[:space:]]*namespace:' "${cfg}" && ! grep -qE '^namespace:' "${cfg}"; then
        preflight_fail "config.yaml: missing 'namespace' key"
    else
        preflight_ok "config.yaml: namespace present"
    fi
    if ! grep -qE '^[[:space:]]*mode:' "${cfg}" && ! grep -qE '^mode:' "${cfg}"; then
        preflight_fail "config.yaml: missing 'mode' key"
    else
        preflight_ok "config.yaml: mode present"
    fi
    if ! grep -qE 'registry:' "${cfg}"; then
        preflight_fail "config.yaml: missing 'registry' (expected under top-level or image:)"
    else
        preflight_ok "config.yaml: registry field present"
    fi
    if grep -qE 'businessDomain' "${cfg}"; then
        preflight_ok "config.yaml: businessDomain present"
    else
        preflight_ok "config.yaml: no businessDomain (optional; only add or use --set when you need business-domain features; otherwise leave unset)"
    fi
}

# --- P2: locale / tz / apparmor / tmp / overlay / route / GPU ----------------
preflight_check_locale() {
    preflight_skip "locale" && return 0
    if ! command -v locale &>/dev/null; then
        return 0
    fi
    if locale 2>/dev/null | grep -q UTF-8; then
        preflight_ok "Locale is UTF-8"
    else
        preflight_warn "Locale may not be UTF-8; some Helm charts expect UTF-8"
    fi
}

preflight_check_timezone() {
    preflight_skip "timezone" && return 0
    if ! command -v timedatectl &>/dev/null; then
        return 0
    fi
    local tz
    tz="$(timedatectl show -p Timezone --value 2>/dev/null || true)"
    if [[ -n "${tz}" ]]; then
        preflight_ok "Timezone: ${tz}"
    else
        preflight_warn "Could not read timezone (timedatectl)"
    fi
}

preflight_check_apparmor() {
    preflight_skip "apparmor" && return 0
    if ! command -v aa-status &>/dev/null; then
        return 0
    fi
    if aa-status 2>/dev/null | grep -qi 'docker'; then
        preflight_ok "apparmor: docker profile present (verify not blocking your runtime)"
    else
        preflight_ok "apparmor: no obvious blockers from aa-status (quick scan)"
    fi
}

preflight_check_tmp() {
    preflight_skip "tmp" && return 0
    local line
    if df -Pk /tmp &>/dev/null; then
        line="$(df -Pk /tmp 2>/dev/null | tail -1)"
    else
        line="$(df -k /tmp 2>/dev/null | tail -1)"
    fi
    local avail
    avail="$(echo "${line}" | awk '{print int($4/1024)}')"
    if mount | grep -qE '[[:space:]]/tmp[[:space:]].*noexec'; then
        preflight_fail "/tmp is mounted noexec; Helm and many tools need exec on /tmp"
    elif [[ -n "${avail}" && "${avail}" -ge 2048 ]]; then
        preflight_ok "/tmp has ${avail} MiB free (>= 2 GiB)"
    else
        preflight_warn "/tmp has ${avail:-?} MiB free (low space may break Helm temp extraction)"
    fi
}

preflight_check_overlayfs() {
    preflight_skip "overlay" && return 0
    if [[ -f /proc/filesystems ]] && grep -q overlay /proc/filesystems; then
        preflight_ok "overlay fs available in /proc/filesystems"
    else
        preflight_warn "overlay not listed in /proc/filesystems (containerd needs overlay)"
    fi
}

preflight_check_default_route() {
    preflight_skip "defroute" && return 0
    if ! command -v ip &>/dev/null; then
        return 0
    fi
    local n
    n="$(ip -4 route show default 2>/dev/null | wc -l | tr -d ' ')"
    if [[ "${n}" == "1" ]]; then
        preflight_ok "Single IPv4 default route"
    else
        preflight_warn "IPv4 default routes: ${n} (multiple defaults can confuse NodePort/ingress; verify routing)"
    fi
}

preflight_check_gpu() {
    preflight_skip "gpu" && return 0
    if [[ "${PREFLIGHT_NEED_GPU:-false}" != "true" ]]; then
        return 0
    fi
    log_info "PREFLIGHT_NEED_GPU: checking nvidia-smi..."
    if command -v nvidia-smi &>/dev/null; then
        preflight_ok "nvidia-smi: $(nvidia-smi -L 2>/dev/null | head -1 || echo ok)"
    else
        preflight_fail "PREFLIGHT_NEED_GPU=true but nvidia-smi not in PATH"
    fi
}

# --- network reachability (optional domains) ---------------------------------
preflight_check_network() {
    preflight_skip "network" && return 0
    log_info "Checking outbound HTTPS to common registries (optional)..."

    if ! command -v curl &>/dev/null; then
        preflight_warn "curl not installed; skipping HTTP reachability checks"
        return
    fi

    local hosts=(
        "mirrors.aliyun.com"
        "mirrors.tuna.tsinghua.edu.cn"
        "registry.aliyuncs.com"
        "swr.cn-east-3.myhuaweicloud.com"
        "repo.huaweicloud.com"
        "kweaver-ai.github.io"
    )
    for h in "${hosts[@]}"; do
        local code
        code="$(
            curl -sS -o /dev/null --max-time 5 --connect-timeout 3 -w '%{http_code}' \
                "https://${h}/" 2>/dev/null || echo "000"
        )"
        if [[ "${code}" != "000" ]]; then
            preflight_ok "HTTPS reachability: ${h} (HTTP ${code})"
        else
            preflight_warn "HTTPS reachability: ${h} (connection/TLS failed; set proxy for air-gap)"
        fi
    done
}

# --- port usage ----------------------------------------------------------------
preflight_check_ports() {
    preflight_skip "ports" && return 0
    log_info "Checking listening ports (6443, 10250, ingress)..."

    if ! command -v ss &>/dev/null && ! command -v netstat &>/dev/null; then
        preflight_warn "Neither ss nor netstat available; skipping port checks"
        return
    fi

    preflight_check_port() {
        local port="$1" desc="$2"
        local busy="no" who=""
        if command -v ss &>/dev/null; then
            if ss -H -lntp "sport = :${port}" 2>/dev/null | grep -q .; then
                busy="yes"
                who="$(ss -H -lntp "sport = :${port}" 2>/dev/null | head -1 | tr -s ' ' | cut -c1-200)"
            fi
        elif netstat -lnt 2>/dev/null | awk -v p=":${port}" 'index($4, p) {f=1} END{exit !f}'; then
            busy="yes"
        fi
        if command -v lsof &>/dev/null && [[ "${busy}" == "yes" && -z "${who}" ]]; then
            who="$(lsof -nP -iTCP:"${port}" -sTCP:LISTEN 2>/dev/null | tail -1 || true)"
        fi
        if [[ "${busy}" == "yes" ]]; then
            if [[ "${port}" == "6443" || "${port}" == "10250" ]]; then
                preflight_ok "Port ${port} in use${who:+: ${who}}"
            else
                preflight_warn "Port ${port} (${desc}) in use${who:+ - ${who}}"
            fi
        else
            preflight_ok "Port ${port} (${desc}) not listening"
        fi
    }

    local hport="${INGRESS_NGINX_HTTP_PORT:-80}"
    local sport="${INGRESS_NGINX_HTTPS_PORT:-443}"
    preflight_check_port 6443 "apiserver"
    preflight_check_port 10250 "kubelet"
    preflight_check_port "${hport}" "ingress http"
    preflight_check_port "${sport}" "ingress https"
}

# --- old cluster / k3s residue -------------------------------------------------
preflight_check_residue() {
    preflight_skip "residue" && return 0
    log_info "Checking for K3s / prior Kubernetes / CNI residue..."

    if [[ -x /usr/local/bin/k3s ]] || command -v k3s &>/dev/null; then
        preflight_fail "K3s binary found; remove or use k3s-killall.sh before this installer"
    else
        preflight_ok "No K3s binary in PATH or /usr/local/bin/k3s"
    fi

    if [[ -f /etc/kubernetes/admin.conf ]]; then
        if command -v kubectl &>/dev/null \
            && KUBECONFIG=/etc/kubernetes/admin.conf kubectl get nodes &>/dev/null; then
            preflight_ok "Existing Kubernetes cluster is healthy at /etc/kubernetes/admin.conf; deploy will reuse it (no reset needed). To force a clean install run: ./deploy.sh k8s reset"
        else
            preflight_fail "Found /etc/kubernetes/admin.conf but cluster is not responding (kubectl get nodes failed). For clean install: ./deploy.sh k8s reset (or preflight fix: kubeadm-reset)"
        fi
    else
        preflight_ok "No /etc/kubernetes/admin.conf (fresh for kubeadm, if target)"
    fi

    if [[ -d /etc/cni/net.d ]] && ls /etc/cni/net.d/* &>/dev/null; then
        preflight_ok "CNI config present under /etc/cni/net.d/ (OK if reusing cluster)"
    fi
}

# --- client tools: target (install host) + admin (optional) -------------------
preflight_check_target_tools() {
    preflight_skip "tools" && return 0
    log_info "Checking target host tools (kubectl, helm)..."
    if command -v kubectl &>/dev/null; then
        preflight_ok "kubectl: $(command -v kubectl)"
    else
        preflight_warn "kubectl not found (k8s install will leave kubeconfig on host; optional on pure worker)"
    fi

    if command -v helm &>/dev/null; then
        local helm_ver
        helm_ver="$(helm version --short 2>/dev/null | awk '{print $1}' | cut -d'+' -f1 || true)"
        if [[ -z "${helm_ver}" ]]; then
            helm_ver="$(helm version --short --client 2>/dev/null | awk -F': ' 'NR==1{print $2}' | awk '{print $1}' | cut -d'+' -f1 || true)"
        fi
        case "${helm_ver}" in
            v3.*)
                preflight_ok "helm: $(command -v helm) (${helm_ver})"
                ;;
            v2.*)
                preflight_fail "helm ${helm_ver} at $(command -v helm) is unsupported; deploy requires Helm v3. Use preflight fix helm-v3 or deploy.sh k8s install path."
                ;;
            "")
                preflight_warn "helm at $(command -v helm) returned no version string; verify v3"
                ;;
            *)
                preflight_warn "helm version '${helm_ver}' at $(command -v helm) (expected v3.x); deploy can re-install ${HELM_VERSION:-v3.x}"
                ;;
        esac
    else
        preflight_warn "helm not found (install via deploy.sh k8s preinstall or preflight fix helm-v3)"
    fi
}

preflight_check_admin_tools() {
    preflight_skip "admin-tools" && return 0
    log_info "Checking admin / optional tools (kweaver, node)..."
    if command -v kweaver &>/dev/null; then
        preflight_ok "kweaver: $(kweaver --version 2>/dev/null | head -1 || echo ok)"
    else
        preflight_warn "kweaver CLI not in PATH (npm i -g @kweaver-ai/kweaver-sdk for onboard / API)"
    fi

    if command -v node &>/dev/null; then
        preflight_ok "node: $(node -v 2>/dev/null)"
    else
        preflight_warn "node not in PATH (use npx kweaver on admin machine as alternative)"
    fi
}

preflight_check_client_tools() {
    local role="${PREFLIGHT_ROLE:-both}"
    case "${role}" in
        target) preflight_check_target_tools ;;
        admin) preflight_check_admin_tools ;;
        both)
            preflight_check_target_tools
            preflight_check_admin_tools
            ;;
        *) preflight_check_target_tools; preflight_check_admin_tools ;;
    esac
}

# --- container runtime --------------------------------------------------------
preflight_check_container_runtime() {
    preflight_skip "containerd" && return 0
    log_info "Checking container runtime (containerd)..."

    if command -v containerd &>/dev/null; then
        local cv
        cv="$(containerd --version 2>/dev/null | awk '{print $3}' || true)"
        preflight_ok "containerd: $(command -v containerd) ${cv}"
    else
        preflight_warn "containerd not found (deploy.sh k8s install will install it; README recommends pre-installing 'containerd.io' on CentOS/openEuler)"
    fi

    if [[ -S /run/containerd/containerd.sock ]] || [[ -S /var/run/containerd/containerd.sock ]]; then
        preflight_ok "containerd socket present"
    else
        preflight_warn "containerd socket not present (start with: systemctl enable --now containerd)"
    fi
}

# --- package manager source health (apt / dnf / yum) --------------------------
# Detect dead repos before deploy.sh tries to install kubeadm/containerd/chrony.
# Most common failure on Ubuntu/Debian is the deprecated
# packages.cloud.google.com Kubernetes apt source returning 404 — that is a
# documented pitfall in deploy/README.zh.md and breaks `apt-get update` for
# every other repo line as well.
preflight_check_pkg_repos() {
    preflight_skip "pkg-repos" && return 0
    log_info "Checking package manager source health..."

    if command -v apt-get &>/dev/null; then
        # Look for the deprecated Google-hosted Kubernetes apt source first; even if
        # `apt-get update` happens to be cached, the 404 will bite at install time.
        local legacy_files=()
        local f
        for f in /etc/apt/sources.list /etc/apt/sources.list.d/*.list; do
            [[ -f "${f}" ]] || continue
            if grep -qE 'packages\.cloud\.google\.com/apt' "${f}" 2>/dev/null; then
                legacy_files+=("${f}")
            fi
        done
        if [[ ${#legacy_files[@]} -gt 0 ]]; then
            preflight_fail "Deprecated Kubernetes apt source detected (packages.cloud.google.com) in: ${legacy_files[*]}. This 404s and breaks 'apt-get update'. Migrate to pkgs.k8s.io (see deploy/README.zh.md 'Kubernetes apt 源 404'); --fix can do it for you."
        else
            preflight_ok "No deprecated packages.cloud.google.com apt source"
        fi

        local apt_log
        apt_log="$(mktemp 2>/dev/null || echo /tmp/preflight-apt.$$)"
        if apt-get update -o Acquire::ForceIPv4=true -o APT::Get::List-Cleanup="0" >"${apt_log}" 2>&1; then
            preflight_ok "apt-get update succeeded (includes sources.list.d)"
        else
            local err_excerpt err_file
            err_excerpt="$(grep -E '^(E:|W:|Err)' "${apt_log}" 2>/dev/null | head -5 | tr '\n' ' ' | sed 's/  */ /g')"
            err_file="$(grep -Eio 'The repository [^ ]+|(/etc/apt/sources\.list\.d/[^ )]+|/etc/apt/sources\.list[^ ]*)' "${apt_log}" 2>/dev/null | head -1 || true)"
            if grep -q 'packages\.cloud\.google\.com' "${apt_log}" 2>/dev/null; then
                preflight_fail "apt-get update failed due to deprecated packages.cloud.google.com source. ${err_excerpt} ${err_file:+(ref: $err_file)}"
            else
                preflight_warn "apt-get update reported errors: ${err_excerpt:-see ${apt_log}} ${err_file:+(suspect file: $err_file)}"
            fi
        fi
        rm -f "${apt_log}" 2>/dev/null || true
    elif command -v dnf &>/dev/null; then
        if dnf repolist --enabled >/dev/null 2>&1; then
            preflight_ok "dnf repolist OK"
        else
            preflight_warn "dnf repolist failed; check /etc/yum.repos.d/* (mirrors.aliyun.com / mirrors.tuna.tsinghua.edu.cn must be reachable)"
        fi
    elif command -v yum &>/dev/null; then
        if yum repolist >/dev/null 2>&1; then
            preflight_ok "yum repolist OK"
        else
            preflight_warn "yum repolist failed; check /etc/yum.repos.d/* (mirrors.aliyun.com / mirrors.tuna.tsinghua.edu.cn must be reachable)"
        fi
    else
        preflight_warn "No supported package manager (apt-get / dnf / yum) found"
    fi
}

# Apply the documented Kubernetes apt-source migration:
#   packages.cloud.google.com  ->  pkgs.k8s.io
# Mirrors deploy/README.zh.md 'Kubernetes apt 源 404'.
preflight_fix_kubernetes_apt_source() {
    if ! command -v apt-get &>/dev/null; then
        return 0
    fi

    local has_legacy=false
    local f
    for f in /etc/apt/sources.list /etc/apt/sources.list.d/*.list; do
        [[ -f "${f}" ]] || continue
        if grep -qE 'packages\.cloud\.google\.com/apt' "${f}" 2>/dev/null; then
            has_legacy=true
            preflight_backup_file "${f}"
        fi
    done
    [[ "${has_legacy}" == "true" ]] || return 0

    if ! command -v curl &>/dev/null || ! command -v gpg &>/dev/null; then
        preflight_warn "Cannot auto-migrate Kubernetes apt source: missing curl or gpg. Run the steps in deploy/README.zh.md manually."
        return 0
    fi

    local k8s_minor
    if [[ -n "${PREFLIGHT_K8S_APT_MINOR:-}" ]]; then
        k8s_minor="${PREFLIGHT_K8S_APT_MINOR}"
    else
        k8s_minor="$(preflight_resolve_k8s_apt_minor)"
    fi
    log_info "Migrating Kubernetes apt source to pkgs.k8s.io (${k8s_minor})..."

    preflight_backup_file /etc/apt/sources.list.d/kubernetes.list
    preflight_backup_file /etc/apt/keyrings/kubernetes-apt-keyring.gpg
    apt-mark unhold kubeadm kubelet kubectl 2>/dev/null || true
    rm -f /etc/apt/sources.list.d/kubernetes.list \
          /etc/apt/keyrings/kubernetes-apt-keyring.gpg 2>/dev/null || true

    mkdir -p /etc/apt/keyrings

    if curl -fsSL "https://pkgs.k8s.io/core:/stable:/${k8s_minor}/deb/Release.key" \
        | gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg 2>/dev/null; then
        echo "deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/${k8s_minor}/deb/ /" \
            > /etc/apt/sources.list.d/kubernetes.list
        if apt-get update -o Acquire::ForceIPv4=true >/dev/null 2>&1; then
            preflight_fixed "Migrated Kubernetes apt source to pkgs.k8s.io (${k8s_minor})"
        else
            preflight_warn "Migrated apt source written but 'apt-get update' still failed; see manual steps in deploy/README.zh.md"
        fi
    else
        preflight_warn "Failed to fetch pkgs.k8s.io Release.key; check network access to pkgs.k8s.io"
    fi
}

# Optional fixes (called from preflight_apply_safe_fixes) --------------------
preflight_fix_k3s_uninstall() {
    if [[ -x /usr/local/bin/k3s-killall.sh ]]; then
        /usr/local/bin/k3s-killall.sh 2>/dev/null || true
    fi
    if [[ -x /usr/local/bin/k3s-uninstall.sh ]]; then
        /usr/local/bin/k3s-uninstall.sh 2>/dev/null || true
    fi
    preflight_fixed "Attempted k3s-killall + k3s-uninstall (see logs if scripts missing)"
}

preflight_fix_kubeadm_reset() {
    local droot="${PREFLIGHT_ROOT:-.}"
    if [[ -f "${droot}/deploy.sh" ]]; then
        (cd "${droot}" && ASSUME_YES=true bash ./deploy.sh k8s reset) || true
    else
        log_warn "Could not find ${droot}/deploy.sh; set PREFLIGHT_ROOT to your deploy/ directory (contains deploy.sh)"
    fi
    preflight_fixed "Ran deploy.sh k8s reset (ASSUME_YES=true)"
}

preflight_fix_containerd_install() {
    if command -v install_containerd &>/dev/null; then
        install_containerd
        preflight_fixed "Ran install_containerd() from k8s.sh"
    elif command -v apt-get &>/dev/null; then
        apt-get update -y 2>/dev/null && apt-get install -y containerd.io 2>/dev/null || true
        mkdir -p /etc/containerd
        if command -v containerd &>/dev/null; then
            containerd config default 2>/dev/null | sed 's/SystemdCgroup = false/SystemdCgroup = true/' > /etc/containerd/config.toml
        fi
        systemctl enable --now containerd 2>/dev/null || true
        preflight_fixed "Installed containerd.io via apt and wrote config.toml (best-effort)"
    elif command -v dnf &>/dev/null; then
        dnf install -y containerd.io 2>/dev/null && systemctl enable --now containerd 2>/dev/null || true
        preflight_fixed "Installed containerd.io via dnf (best-effort)"
    else
        preflight_warn "Could not auto-install containerd (no known package path)"
    fi
}

preflight_fix_helm_v3() {
    if command -v install_helm &>/dev/null; then
        install_helm
        preflight_fixed "Ran install_helm() (Helm 3)"
    else
        preflight_warn "install_helm not available; source k8s.sh before preflight"
    fi
}

preflight_fix_iptables_legacy() {
    if ! command -v update-alternatives &>/dev/null; then
        return 0
    fi
    update-alternatives --set iptables /usr/sbin/iptables-legacy 2>/dev/null || true
    update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy 2>/dev/null || true
    preflight_fixed "Set iptables/ip6tables to legacy alternatives (best-effort)"
}

preflight_fix_kernel_limits_sysctl() {
    preflight_backup_file /etc/sysctl.d/99-kweaver-preflight.conf
    cat > /etc/sysctl.d/99-kweaver-preflight.conf <<'EOF' || true
# Added by KWeaver preflight
vm.max_map_count = 262144
fs.inotify.max_user_watches = 524288
fs.inotify.max_user_instances = 8192
kernel.pid_max = 4194304
EOF
    timeout 10 sysctl --system 2>/dev/null || true
    preflight_fixed "Wrote /etc/sysctl.d/99-kweaver-preflight.conf and ran sysctl --system"
}

preflight_fix_bridge_sysctl() {
    preflight_backup_file /etc/sysctl.d/99-kweaver-bridge.conf
    cat > /etc/sysctl.d/99-kweaver-bridge.conf <<'EOF' || true
# Added by KWeaver preflight
net.bridge.bridge-nf-call-iptables = 1
net.bridge.bridge-nf-call-ip6tables = 1
EOF
    timeout 10 sysctl --system 2>/dev/null || true
    preflight_fixed "Wrote /etc/sysctl.d/99-kweaver-bridge.conf and ran sysctl --system (may need br_netfilter loaded first)"
}

# Print suggested fixes from first-pass [FAIL] lines
preflight_print_fix_preview() {
    if [[ ${#PREFLIGHT_FAIL_SNAPSHOT[@]} -eq 0 ]]; then
        return 0
    fi
    log_info "---- Fix preview (from [FAIL] above; you will be asked per item) ----"
    local line
    for line in "${PREFLIGHT_FAIL_SNAPSHOT[@]}"; do
        log_info "  * ${line}"
    done
    log_info "  Suggested fix names: k3s-uninstall, kubeadm-reset, k8s-apt-source, containerd-install, helm-v3, chrony, firewalld, ufw, selinux, system-tuning, bridge-sysctl, kernel-limits, iptables-legacy, etc-hosts (only applicable steps run)"
    log_info "------------------------------------------------------------------"
}

# Re-run all checks after fixes (resets OK/WARN/FAIL counters; keeps FIXED/declined and fixed JSON)
preflight_recheck_after_fixes() {
    if [[ "${PREFLIGHT_NO_RECHECK:-false}" == "true" ]]; then
        return 0
    fi
    log_info "========== Re-check after fixes =========="
    PREFLIGHT_IN_RECHECK=true
    PREFLIGHT_OK_COUNT=0
    PREFLIGHT_WARN_COUNT=0
    PREFLIGHT_FAIL_COUNT=0
    PREFLIGHT_FAIL_SNAPSHOT=()
    if [[ "${PREFLIGHT_OUTPUT_JSON:-false}" == "true" ]]; then
        PREFLIGHT_JSON_OK=()
        PREFLIGHT_JSON_WARN=()
        PREFLIGHT_JSON_FAIL=()
    fi
    preflight_run_all_checks
    PREFLIGHT_IN_RECHECK=false
}

# --- safe auto-fixes (requires root) ------------------------------------------
preflight_apply_safe_fixes() {
    if [[ "${PREFLIGHT_CHECK_ONLY}" == "true" ]]; then
        log_info "Check-only mode: skipping automatic fixes."
        return 0
    fi
    if [[ "${EUID}" -ne 0 ]]; then
        if [[ "${PREFLIGHT_LIST_FIXES_ONLY:-false}" != "true" ]]; then
            preflight_warn "Not root: skipping automatic fixes (run with sudo for swap/selinux/sysctl/hosts/chrony)"
            return 0
        fi
    fi

    preflight_print_fix_preview

    log_info "Applying pre-install fixes (order: destructive cleanup first, then apt source, then packages; -y to auto-approve)..."

    # --- 1) k3s uninstall ----------------------------------------------------
    if [[ -x /usr/local/bin/k3s ]] || command -v k3s &>/dev/null; then
        if preflight_confirm_fix "k3s-uninstall" \
            "Run k3s-killall.sh and k3s-uninstall.sh" \
            "Removes k3s data; destructive. Only if you intend to use kubeadm/this installer."; then
            preflight_fix_k3s_uninstall
        fi
    fi

    # --- 2) kubeadm reset (existing cluster) ---------------------------------
    # Only offer reset when the cluster is actually broken; healthy clusters
    # are intended to be reused by deploy.sh (ensure_k8s skips re-install).
    if [[ -f /etc/kubernetes/admin.conf ]]; then
        if command -v kubectl &>/dev/null \
            && KUBECONFIG=/etc/kubernetes/admin.conf kubectl get nodes &>/dev/null; then
            log_info "  -> skipping kubeadm-reset: existing cluster is healthy and will be reused"
        elif preflight_confirm_fix "kubeadm-reset" \
            "cd PREFLIGHT_ROOT && ASSUME_YES=true ./deploy.sh k8s reset" \
            "Destroys the local Kubernetes control plane, certs, and kubeconfig. Irreversible."; then
            preflight_fix_kubeadm_reset
        fi
    fi

    # --- 3) Kubernetes apt source (BEFORE any apt install) -------------------
    local k8s_apt_resolved
    k8s_apt_resolved="${PREFLIGHT_K8S_APT_MINOR:-$(preflight_resolve_k8s_apt_minor)}"
    if command -v apt-get &>/dev/null; then
        local has_legacy=false f
        for f in /etc/apt/sources.list /etc/apt/sources.list.d/*.list; do
            [[ -f "${f}" ]] || continue
            if grep -qE 'packages\.cloud\.google\.com/apt' "${f}" 2>/dev/null; then
                has_legacy=true
                break
            fi
        done
        if [[ "${has_legacy}" == "true" ]]; then
            if preflight_confirm_fix "k8s-apt-source" \
                "Migrate k8s apt to pkgs.k8s.io/${k8s_apt_resolved}" \
                "Rewrites kubernetes.list and keyring; unholds kube packages. Set PREFLIGHT_K8S_APT_MINOR to override version."; then
                PREFLIGHT_K8S_APT_MINOR="${k8s_apt_resolved}" preflight_fix_kubernetes_apt_source
            fi
        fi
    fi

    # --- 4) containerd install -----------------------------------------------
    if ! command -v containerd &>/dev/null; then
        if preflight_confirm_fix "containerd-install" \
            "install_containerd (or apt/dnf install containerd.io) + SystemdCgroup" \
            "Installs a system package and overwrites /etc/containerd/config.toml; may change runtime behavior."; then
            preflight_fix_containerd_install
        fi
    fi

    # --- 5) Helm 3 -----------------------------------------------------------
    local hver=""
    if command -v helm &>/dev/null; then
        hver="$(helm version --short 2>/dev/null | awk '{print $1}' | cut -d'+' -f1 || true)"
        [[ -z "${hver}" ]] && hver="$(helm version --short --client 2>/dev/null | awk -F': ' 'NR==1{print $2}' | awk '{print $1}' | cut -d'+' -f1 || true)"
    fi
    if ! command -v helm &>/dev/null || [[ "${hver}" == v2.* ]]; then
        if preflight_confirm_fix "helm-v3" \
            "install_helm (Helm 3 from k8s.sh)" \
            "May replace /usr/local/bin/helm. Required for v3 --timeout duration syntax."; then
            preflight_fix_helm_v3
        fi
    fi

    # --- 6) chrony -----------------------------------------------------------
    if ! systemctl is-active --quiet chronyd 2>/dev/null \
        && ! systemctl is-active --quiet ntpd 2>/dev/null \
        && ! systemctl is-active --quiet systemd-timesyncd 2>/dev/null; then
        if preflight_confirm_fix "chrony" \
            "Install chrony via system package manager and enable chronyd" \
            "Adds a package + enables a system service."; then
            if command -v dnf &>/dev/null; then
                dnf install -y chrony 2>/dev/null && systemctl enable --now chronyd 2>/dev/null || true
                preflight_fixed "Installed/ensured chrony via dnf"
            elif command -v yum &>/dev/null; then
                yum install -y chrony 2>/dev/null && systemctl enable --now chronyd 2>/dev/null || true
                preflight_fixed "Installed/ensured chrony via yum"
            elif command -v apt-get &>/dev/null; then
                apt-get update -y 2>/dev/null && apt-get install -y chrony 2>/dev/null && systemctl enable --now chrony 2>/dev/null || true
                preflight_fixed "Installed/ensured chrony via apt"
            fi
        fi
    fi

    # --- 7) firewalld / ufw / selinux / system-tuning (existing order) -------
    if systemctl is-active --quiet firewalld 2>/dev/null; then
        if preflight_confirm_fix "firewalld" \
            "systemctl stop firewalld && systemctl disable firewalld" \
            "Disables host firewall for lab-style installs."; then
            systemctl stop firewalld 2>/dev/null || true
            systemctl disable firewalld 2>/dev/null || true
            preflight_fixed "Stopped and disabled firewalld"
        fi
    fi

    if command -v ufw &>/dev/null && ufw status 2>/dev/null | grep -qi "Status: active"; then
        if preflight_confirm_fix "ufw" \
            "ufw --force disable" \
            "Disables Ubuntu firewall; open required ports in production."; then
            ufw --force disable 2>/dev/null || true
            preflight_fixed "Disabled ufw"
        fi
    fi

    if command -v disable_selinux &>/dev/null && command -v getenforce &>/dev/null \
        && [[ "$(getenforce 2>/dev/null)" == "Enforcing" ]]; then
        if preflight_confirm_fix "selinux" \
            "disable_selinux (permissive + config)" \
            "Weakens MAC; typical for this kubeadm path."; then
            disable_selinux
            preflight_fixed "Applied disable_selinux()"
        fi
    fi

    if command -v configure_system &>/dev/null; then
        if preflight_confirm_fix "system-tuning" \
            "configure_system() (swap, sysctl, modules from k8s.sh)" \
            "Disables swap, sets ip_forward, loads modules. Required for Kubernetes."; then
            configure_system
            preflight_fixed "Applied configure_system() (swap, sysctl, modules)"
        fi
    fi

    if [[ -d /proc/sys/net/bridge ]] && { [[ -f /proc/sys/net/bridge/bridge-nf-call-iptables ]] && [[ "$(cat /proc/sys/net/bridge/bridge-nf-call-iptables 2>/dev/null)" != "1" ]]; }; then
        if preflight_confirm_fix "bridge-sysctl" \
            "Write /etc/sysctl.d/99-kweaver-bridge.conf" \
            "Sets bridge-nf-for-iptables; needs br_netfilter loaded to take effect."; then
            preflight_fix_bridge_sysctl
        fi
    fi

    if preflight_confirm_fix "kernel-limits" \
        "Write /etc/sysctl.d/99-kweaver-preflight.conf (vm, inotify, pid_max)" \
        "Persistent kernel tuning; remove file to revert."; then
        preflight_fix_kernel_limits_sysctl
    fi

    if command -v update-alternatives &>/dev/null && update-alternatives --display iptables 2>/dev/null | grep -qi 'current mode.*nf_tables'; then
        if preflight_confirm_fix "iptables-legacy" \
            "update-alternatives --set iptables ip6tables to *-legacy" \
            "Switches host iptables backend; can affect non-K8s firewall tooling."; then
            preflight_fix_iptables_legacy
        fi
    fi

    # --- /etc/hosts ----------------------------------------------------------
    local hn
    hn="$(hostname 2>/dev/null || true)"
    if [[ -n "${hn}" && -f /etc/hosts ]] \
        && ! grep -qE "127\.0\.0\.1[[:space:]]+${hn}" /etc/hosts; then
        if preflight_confirm_fix "etc-hosts" \
            "Append '127.0.0.1 ${hn}' to /etc/hosts" \
            "Single-line hostname mapping; backup recommended on manual edit."; then
            preflight_backup_file /etc/hosts
            echo "127.0.0.1 ${hn}" >> /etc/hosts
            preflight_fixed "Appended 127.0.0.1 ${hn} to /etc/hosts"
        fi
    fi
}

# --- run all checks in order ---------------------------------------------------
preflight_run_all_checks() {
    preflight_check_os
    preflight_check_arch
    preflight_check_hardware
    preflight_check_hostname_hosts
    preflight_check_swap_selinux
    preflight_check_firewall
    preflight_check_sysctl_modules
    preflight_check_bridge_sysctl
    preflight_check_kernel_limits
    preflight_check_ulimits
    preflight_check_cgroup
    preflight_check_time_sync
    preflight_check_proxy
    preflight_check_dns
    preflight_check_kubeadm_deps
    preflight_check_network
    preflight_check_k8s_version
    preflight_check_cidr_conflict
    preflight_check_ports
    preflight_check_extended_ports
    preflight_check_residue
    preflight_check_container_runtime
    preflight_check_containerd_disk
    preflight_check_pkg_repos
    preflight_check_docker_residue
    preflight_check_containerd_cgroup_driver
    preflight_check_iptables_backend
    preflight_check_ntp_drift
    preflight_check_systemd_version
    preflight_check_existing_release
    preflight_check_node_capacity
    preflight_check_offline_assets
    preflight_check_config_yaml
    preflight_check_locale
    preflight_check_timezone
    preflight_check_apparmor
    preflight_check_tmp
    preflight_check_overlayfs
    preflight_check_default_route
    preflight_check_gpu
    preflight_check_client_tools
}

# --- exit code: 0 ok, 1 fail, 2 warn only -------------------------------------
preflight_compute_exit_code() {
    if [[ "${PREFLIGHT_FAIL_COUNT}" -gt 0 ]]; then
        return 1
    fi
    if [[ "${PREFLIGHT_WARN_COUNT}" -gt 0 ]]; then
        return 2
    fi
    return 0
}
