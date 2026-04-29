
install_redis() {
    local ns="${REDIS_NAMESPACE}"
    
    # Create namespace if not exists
    kubectl create namespace "${ns}" 2>/dev/null || true

    install_redis_sentinel_local
    return $?
}

# Install Redis in sentinel mode using local chart (proton-redis)
install_redis_sentinel_local() {
    local ns="${REDIS_NAMESPACE}"
    local redis_release_name="redis"

    local chart_ref=""
    if [[ -f "${REDIS_CHART_TGZ}" ]]; then
        chart_ref="${REDIS_CHART_TGZ}"
    elif [[ -d "${REDIS_LOCAL_CHART_DIR}" ]]; then
        chart_ref="${REDIS_LOCAL_CHART_DIR}"
    else
        log_error "Redis chart not found (need tgz or dir)."
        log_error "  REDIS_CHART_TGZ=${REDIS_CHART_TGZ}"
        log_error "  REDIS_LOCAL_CHART_DIR=${REDIS_LOCAL_CHART_DIR}"
        return 1
    fi

    local fresh_install="true"
    if is_helm_installed "${redis_release_name}" "${ns}"; then
        fresh_install="false"
        log_info "Redis is already installed. Skipping installation."
        return 0
    fi
    log_info "Installing Redis in sentinel mode using proton-redis chart..."

    # Build image registry string (default from user's values)
    local image_registry="${REDIS_IMAGE_REGISTRY:-swr.cn-east-3.myhuaweicloud.com/kweaver-ai}"
    if [[ -z "${image_registry}" ]]; then
        # Try to get from config.yaml
        image_registry=$(grep -E "^[[:space:]]*registry:" "${SCRIPT_DIR}/conf/config.yaml" 2>/dev/null | head -1 | sed 's/.*registry:[[:space:]]*//' | tr -d "'\"")
    fi
    if [[ -z "${image_registry}" ]]; then
        image_registry="swr.cn-east-3.myhuaweicloud.com/kweaver-ai"
    fi

    local redis_password="${REDIS_PASSWORD}"
    if [[ -z "${redis_password}" ]]; then
        redis_password="$(generate_random_password 10)"
    fi
    if [[ -z "${redis_password}" ]]; then
        log_error "Failed to generate Redis password (install openssl or set REDIS_PASSWORD)."
        return 1
    fi
    REDIS_PASSWORD="${redis_password}"

    local redis_sc
    redis_sc="$(kweaver_resolve_redis_storage_class)"

    # Prepare Helm values according to user's specification
    local -a helm_args
    helm_args=(
        upgrade --install redis "${chart_ref}"
        --namespace "${ns}"
        --set enableSecurityContext=false
        --set env.language=en_US.UTF-8
        --set env.timezone=Asia/Shanghai
        --set image.registry="${image_registry}"
        --set namespace="${ns}"
        --set redis.masterGroupName="${REDIS_MASTER_GROUP_NAME:-mymaster}"
        --set redis.rootPassword="${redis_password}"
        --set redis.password="${redis_password}"
        --set sentinel.password="${redis_password}"
        --set redis.monitorUser="monitor-user"
        --set redis.monitorPassword="${redis_password}"
        --set redis.rootUsername=root
        --set replicaCount="${REDIS_REPLICA_COUNT:-1}"
        --set service.enableDualStack=false
        --set service.sentinel.port=26379
        --set storage.storageClassName="${redis_sc}"
        --wait --timeout=600s
    )

    # Set image repository and tag if provided
    if [[ -n "${REDIS_IMAGE_REPOSITORY}" ]]; then
        helm_args+=(--set image.redis.repository="${REDIS_IMAGE_REPOSITORY}")
    fi
    if [[ -n "${REDIS_IMAGE_TAG}" ]]; then
        helm_args+=(--set image.redis.tag="${REDIS_IMAGE_TAG}")
    fi

    # Set storage capacity if persistence is enabled
    if [[ "${REDIS_PERSISTENCE_ENABLED:-true}" == "true" ]]; then
        helm_args+=(--set storage.capacity="${REDIS_STORAGE_SIZE:-5Gi}")
    fi

    log_info "Installing Redis with values:"
    log_info "  Chart: ${chart_ref}"
    log_info "  Namespace: ${ns}"
    log_info "  Image Registry: ${image_registry}"
    log_info "  Replica Count: ${REDIS_REPLICA_COUNT:-1}"
    log_info "  Master Group: ${REDIS_MASTER_GROUP_NAME:-mymaster}"
    log_info "  Storage Class: ${redis_sc}"

    helm "${helm_args[@]}"
    
    # Wait for Pods to be ready
    log_info "Waiting for Redis Pods to be ready..."
    # Try multiple label selectors for different chart naming conventions
    kubectl wait --for=condition=ready pod -l app="${redis_release_name}-proton-redis" -n "${ns}" --timeout=300s 2>/dev/null || \
    kubectl wait --for=condition=ready pod -l "app.kubernetes.io/instance=${redis_release_name}" -n "${ns}" --timeout=300s 2>/dev/null || {
        log_warn "Redis Pod(s) may not be ready yet"
    }

    if [[ "${fresh_install}" == "true" && "${AUTO_GENERATE_CONFIG}" == "true" ]]; then
        generate_config_yaml
    fi
    
    log_info "Redis sentinel mode installed successfully"
    log_info "Redis sentinel connection info:"
    log_info "  Sentinel Host: redis-sentinel.${ns}.svc.cluster.local"
    log_info "  Sentinel Port: 26379"
    log_info "  Master Group: ${REDIS_MASTER_GROUP_NAME:-mymaster}"
    log_info "  Password: ${redis_password}"
    log_info "  Replicas: ${REDIS_REPLICA_COUNT:-1}"

    if [[ "${fresh_install}" == "true" && "${AUTO_GENERATE_CONFIG}" == "true" ]]; then
        log_info "Config.yaml updated after fresh Redis install"
    fi
}

uninstall_redis() {
    local ns="${REDIS_NAMESPACE}"
    log_info "Uninstalling Redis from namespace ${ns}..."

    helm uninstall redis -n "${ns}" 2>/dev/null || true
    # Best-effort cleanup for old Bitnami redis chart resources (may remain if release was upgraded/failed).
    kubectl delete -n "${ns}" sts,deploy,svc,pod,cm,secret,pdb -l app.kubernetes.io/instance=redis 2>/dev/null || true
    kubectl delete deploy/redis -n "${ns}" 2>/dev/null || true
    kubectl delete sts/redis -n "${ns}" 2>/dev/null || true
    kubectl delete svc/redis -n "${ns}" 2>/dev/null || true
    kubectl delete secret/redis-auth -n "${ns}" 2>/dev/null || true

    if [[ "${REDIS_PURGE_PVC}" == "true" ]]; then
        log_warn "REDIS_PURGE_PVC=true: deleting Redis PVCs (data loss!)"
        # Delete PVCs by label (Bitnami chart)
        kubectl delete pvc -n "${ns}" -l app.kubernetes.io/instance=redis 2>/dev/null || true
        kubectl delete pvc -n "${ns}" -l app.kubernetes.io/name=redis 2>/dev/null || true
        kubectl delete pvc -n "${ns}" -l app=redis 2>/dev/null || true
        # Delete PVCs by name pattern (local chart StatefulSet)
        # Local chart uses volumeClaimTemplates, so PVCs are named: redis-datadir-redis-0, redis-datadir-redis-1, etc.
        local redis_release_name="redis"
        local pvc_patterns=(
            "data-redis-0"
            "data-redis-1"
            "data-redis-2"
            "redis-datadir-${redis_release_name}-0"
            "redis-datadir-${redis_release_name}-1"
            "redis-datadir-${redis_release_name}-2"
        )
        for pvc_name in "${pvc_patterns[@]}"; do
            kubectl delete pvc -n "${ns}" "${pvc_name}" 2>/dev/null || true
        done
        # Also try to find and delete any PVCs that match the pattern
        local existing_pvcs
        existing_pvcs="$(kubectl -n "${ns}" get pvc -o jsonpath='{.items[*].metadata.name}' 2>/dev/null || echo "")"
        if [[ -n "${existing_pvcs}" ]]; then
            for pvc in ${existing_pvcs}; do
                if [[ "${pvc}" =~ ^redis-datadir-.*-redis-[0-9]+$ ]] || [[ "${pvc}" =~ ^data-redis-[0-9]+$ ]]; then
                    kubectl delete pvc -n "${ns}" "${pvc}" 2>/dev/null || true
                fi
            done
        fi
    else
        log_info "REDIS_PURGE_PVC=false: Redis PVCs were retained."
    fi

    log_info "Redis uninstall done"
}
