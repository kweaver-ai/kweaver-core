# Post-install: import Context Loader toolset (agent-retrieval ADP) via operator-integration impex.
# Sourced from deploy.sh after common.sh. Requires deploy.sh's _read_access_address_field for ingress import.

_context_loader_log_manual_import_instructions() {
    log_info "  Please import manually after install:"
    log_info "    kweaver auth login"
    log_info "    kweaver toolbox import --file <repo>/adp/context-loader/agent-retrieval/docs/release/toolset/context_loader_toolset.adp --mode upsert"
}

# v1 impex 在开启 OAuth / Hydra 或 auth.enabled=true 时需要管理员 Bearer；不要走无 Token 的 port-forward（必然 401）。
_context_loader_impex_requires_bearer() {
    local namespace="$1"
    local auth_env
    if kubectl get deploy agent-operator-integration -n "${namespace}" &>/dev/null; then
        auth_env="$(kubectl get deploy agent-operator-integration -n "${namespace}" -o jsonpath='{.spec.template.spec.containers[0].env[?(@.name=="AUTH_ENABLED")].value}' 2>/dev/null || true)"
        if [[ "${auth_env}" == "true" ]]; then
            return 0
        fi
    fi
    if kubectl get deploy hydra -n "${namespace}" &>/dev/null; then
        return 0
    fi
    if kubectl get deploy -n "${namespace}" --no-headers 2>/dev/null | awk '{print tolower($1)}' | grep -qE '^hydra'; then
        return 0
    fi
    if helm status hydra -n "${namespace}" &>/dev/null; then
        return 0
    fi
    return 1
}

_import_context_loader_toolset_via_port_forward() {
    local namespace="$1"
    local adp="$2"
    local local_port="${CONTEXT_LOADER_PF_LOCAL_PORT:-}"
    if [[ -z "${local_port}" ]]; then
        local_port=$((37100 + (RANDOM % 900)))
    fi

    # Pod 监听端口：与 chart 中 container name public-port 一致，默认 9000（见 agent-operator-integration/values.yaml service.port）
    local remote_port="${CONTEXT_LOADER_REMOTE_PORT:-}"
    if [[ -z "${remote_port}" ]]; then
        remote_port="$(kubectl -n "${namespace}" get deploy agent-operator-integration -o jsonpath='{.spec.template.spec.containers[0].ports[?(@.name=="public-port")].containerPort}' 2>/dev/null || true)"
    fi
    remote_port="${remote_port:-9000}"

    log_info "Context Loader toolset: importing via kubectl port-forward (127.0.0.1:${local_port} -> pod:${remote_port})..."

    kubectl -n "${namespace}" port-forward "deploy/agent-operator-integration" "${local_port}:${remote_port}" >/dev/null 2>&1 &
    local pf_pid=$!

    _stop_context_loader_port_forward() {
        if kill -0 "${pf_pid}" 2>/dev/null; then
            kill "${pf_pid}" 2>/dev/null || true
            wait "${pf_pid}" 2>/dev/null || true
            log_info "Context Loader toolset: kubectl port-forward stopped (127.0.0.1:${local_port})."
        fi
    }

    local ready=false
    local i
    for i in $(seq 1 60); do
        if curl -sf --connect-timeout 1 "http://127.0.0.1:${local_port}/health/ready" >/dev/null 2>&1; then
            ready=true
            break
        fi
        if ! kill -0 "${pf_pid}" 2>/dev/null; then
            log_warn "Context Loader toolset: port-forward exited early."
            return 1
        fi
        sleep 0.5
    done

    if [[ "${ready}" != "true" ]]; then
        _stop_context_loader_port_forward
        log_warn "Context Loader toolset: port-forward did not become ready in time."
        return 1
    fi

    local bd="${DEPLOY_BUSINESS_DOMAIN:-bd_public}"

    local result http_code body
    # 与公开 API 一致：impex 只挂在 v1（internal-v1 无 impex 路由）。最小安装 / 未配置 Bearer 时依赖网关或 Hydra 校验关闭。
    result="$(curl -s -w "\n%{http_code}" --connect-timeout 120 -X POST \
        "http://127.0.0.1:${local_port}/api/agent-operator-integration/v1/impex/import/toolbox" \
        -H "x-business-domain: ${bd}" \
        -F "data=@${adp}" \
        -F "mode=upsert" 2>/dev/null || true)"
    # 无论导入成功与否，立即结束 port-forward，避免本地端口长期占用
    _stop_context_loader_port_forward

    http_code="$(echo "${result}" | tail -n 1)"
    body="$(echo "${result}" | sed '$d')"

    if [[ "${http_code}" =~ ^2[0-9][0-9]$ ]]; then
        log_info "Context Loader toolset import succeeded via port-forward (HTTP ${http_code})."
        return 0
    fi

    log_warn "Context Loader toolset import via port-forward failed (HTTP ${http_code}). If OAuth / Hydra verification is enabled, set DEPLOY_PLATFORM_ACCESS_TOKEN and use access-address import, or ensure v1 impex is allowed without Bearer. Response: ${body:0:500}"
    return 1
}

maybe_import_context_loader_toolset_post_core() {
    local namespace="${1:-kweaver}"

    if [[ "${IMPORT_CONTEXT_LOADER_TOOLSET:-true}" == "false" ]]; then
        log_info "Skipping Context Loader toolset import (IMPORT_CONTEXT_LOADER_TOOLSET=false)"
        return 0
    fi

    if ! command -v _read_access_address_field &>/dev/null; then
        log_warn "Context Loader toolset import skipped (_read_access_address_field unavailable)."
        return 0
    fi

    if ! kubectl get deploy agent-operator-integration -n "${namespace}" &>/dev/null; then
        log_warn "Context Loader toolset import skipped (no deploy/agent-operator-integration in namespace ${namespace})."
        return 0
    fi

    # ISF（Hydra）或 Core auth.enabled=true 时，impex 会强校验 Bearer + admin 权限。
    # 仅匹配 deploy/hydra 会漏掉 chart 生成的 hydra-* 等名称；以集群内 AUTH_ENABLED + Hydra 迹象为准。
    if _context_loader_impex_requires_bearer "${namespace}"; then
        if [[ -n "${DEPLOY_PLATFORM_ACCESS_TOKEN:-}" ]]; then
            log_info "Context Loader toolset: authenticated cluster detected; will use DEPLOY_PLATFORM_ACCESS_TOKEN for impex."
        else
            log_info "Context Loader toolset auto-import skipped: authenticated API (Hydra/ISF or auth.enabled=true) in namespace ${namespace}."
            log_info "  Set DEPLOY_PLATFORM_ACCESS_TOKEN for automated import, or import manually:"
            _context_loader_log_manual_import_instructions
            return 0
        fi
    fi

    local _lib_dir
    _lib_dir="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" && pwd)"
    local repo_root
    repo_root="$(cd "${_lib_dir}/../../.." && pwd)"
    local adp="${CONTEXT_LOADER_TOOLSET_ADP_PATH:-${repo_root}/adp/context-loader/agent-retrieval/docs/release/toolset/context_loader_toolset.adp}"

    if [[ ! -f "${adp}" ]]; then
        log_warn "Context Loader toolset ADP not found: ${adp}"
        return 0
    fi

    log_info "Waiting for agent-operator-integration to be ready (toolset import)..."
    if ! kubectl rollout status deploy/agent-operator-integration -n "${namespace}" --timeout=300s &>/dev/null; then
        log_warn "agent-operator-integration not ready within 300s; skipping toolset import."
        return 0
    fi

    local token="${DEPLOY_PLATFORM_ACCESS_TOKEN:-}"
    if [[ -n "${token}" ]]; then
        local host port scheme
        host="$(_read_access_address_field "host" 2>/dev/null || true)"
        port="$(_read_access_address_field "port" 2>/dev/null || true)"
        scheme="$(_read_access_address_field "scheme" 2>/dev/null || true)"
        host="${host:-localhost}"
        port="${port:-443}"
        scheme="${scheme:-https}"

        local base_url
        if [[ "${port}" == "443" || "${port}" == "80" ]]; then
            base_url="${scheme}://${host}"
        else
            base_url="${scheme}://${host}:${port}"
        fi

        local bd="${DEPLOY_BUSINESS_DOMAIN:-bd_public}"

        log_info "Importing Context Loader toolset (impex upsert via access address)..."
        local result http_code body
        result="$(curl -s -k -w "\n%{http_code}" -X POST \
            "${base_url}/api/agent-operator-integration/v1/impex/import/toolbox" \
            -H "x-business-domain: ${bd}" \
            -H "Authorization: Bearer ${token}" \
            -F "data=@${adp}" \
            -F "mode=upsert" 2>/dev/null || true)"
        http_code="$(echo "${result}" | tail -n 1)"
        body="$(echo "${result}" | sed '$d')"

        if [[ "${http_code}" =~ ^2[0-9][0-9]$ ]]; then
            log_info "Context Loader toolset import succeeded (HTTP ${http_code})."
            return 0
        fi

        log_warn "Context Loader toolset import via access address failed (HTTP ${http_code}). Response: ${body:0:500}"
        if _context_loader_impex_requires_bearer "${namespace}"; then
            log_info "Context Loader toolset: not retrying port-forward (impex requires Bearer in this cluster)."
            _context_loader_log_manual_import_instructions
            return 0
        fi
        log_info "Context Loader toolset: retrying via kubectl port-forward..."
    else
        log_info "Context Loader toolset: no DEPLOY_PLATFORM_ACCESS_TOKEN; trying kubectl port-forward (typical for auth.enabled=false / --minimum)."
    fi

    _import_context_loader_toolset_via_port_forward "${namespace}" "${adp}" || true
    return 0
}
