
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
    # ACL drift guard: the proton-redis chart bakes sha256(password) into users.acl on
    # first init and writes it onto the PVC. Subsequent installs that retain the PVC
    # but generate a NEW password produce on-disk ACL <-> Secret mismatch and the
    # liveness probe fails with WRONGPASS in a CrashLoop. Prefer reusing the existing
    # Secret password; refuse to randomly regenerate when a stale PVC is present.
    local redis_secret_name="${redis_release_name}-proton-redis-secret"
    local redis_pvc_name="redis-datadir-${redis_release_name}-proton-redis-0"
    if [[ -z "${redis_password}" ]]; then
        local existing_password
        existing_password="$(kubectl get secret "${redis_secret_name}" -n "${ns}" \
            -o jsonpath='{.data.nonEncrpt-password}' 2>/dev/null \
            | base64 -d 2>/dev/null || true)"
        if [[ -n "${existing_password}" ]]; then
            redis_password="${existing_password}"
            log_info "Reusing existing Redis password from secret ${redis_secret_name} (avoids ACL drift on retained PVC)."
        fi
    fi
    if [[ -z "${redis_password}" ]] && kubectl get pvc "${redis_pvc_name}" -n "${ns}" >/dev/null 2>&1; then
        log_error "PVC ${redis_pvc_name} exists from a previous Redis install but no Secret was found,"
        log_error "and no REDIS_PASSWORD was supplied. A new random password would not match the ACL"
        log_error "hash baked into ${redis_pvc_name}/users.acl, and Redis would CrashLoop on the"
        log_error "liveness probe with WRONGPASS."
        log_error "Either:"
        log_error "  - export REDIS_PASSWORD=<previous-password> and re-run, or"
        log_error "  - REDIS_PURGE_PVC=true bash ${SCRIPT_DIR}/deploy.sh redis uninstall && bash ${SCRIPT_DIR}/deploy.sh redis install"
        return 1
    fi
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

    # Redis process maxmemory (chart key uses Redis-style suffix: mb/gb)
    if [[ -n "${REDIS_MAXMEMORY}" ]]; then
        helm_args+=(--set redis.maxmemory="${REDIS_MAXMEMORY}")
    fi

    # K8s container resources (chart only wires `resources` into the redis container;
    # sentinel/exporter have no resources block in the template).
    if [[ -n "${REDIS_MEMORY_REQUEST}" ]]; then
        helm_args+=(--set resources.requests.memory="${REDIS_MEMORY_REQUEST}")
    fi
    if [[ -n "${REDIS_CPU_REQUEST}" ]]; then
        helm_args+=(--set resources.requests.cpu="${REDIS_CPU_REQUEST}")
    fi
    if [[ -n "${REDIS_MEMORY_LIMIT}" ]]; then
        helm_args+=(--set resources.limits.memory="${REDIS_MEMORY_LIMIT}")
    fi
    if [[ -n "${REDIS_CPU_LIMIT}" ]]; then
        helm_args+=(--set resources.limits.cpu="${REDIS_CPU_LIMIT}")
    fi

    log_info "Installing Redis with values:"
    log_info "  Chart: ${chart_ref}"
    log_info "  Namespace: ${ns}"
    log_info "  Image Registry: ${image_registry}"
    log_info "  Replica Count: ${REDIS_REPLICA_COUNT:-1}"
    log_info "  Master Group: ${REDIS_MASTER_GROUP_NAME:-mymaster}"
    log_info "  Storage Class: ${redis_sc}"
    log_info "  maxmemory: ${REDIS_MAXMEMORY:-<chart default>}"
    log_info "  resources.requests: cpu=${REDIS_CPU_REQUEST:-<unset>} memory=${REDIS_MEMORY_REQUEST:-<unset>}"
    log_info "  resources.limits:   cpu=${REDIS_CPU_LIMIT:-<unset>} memory=${REDIS_MEMORY_LIMIT:-<unset>}"

    helm "${helm_args[@]}"

    # ACL self-heal patch (opt-out: REDIS_AUTO_PATCH_ACL=false).
    [[ "${REDIS_AUTO_PATCH_ACL}" == "true" ]] && \
        _redis_inject_wipe_acl_init "${ns}" "${redis_release_name}-proton-redis"

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

# Prepend a `wipe-stale-acl` initContainer (using the same image as the redis
# container) so every Pod start clears /data/conf/{users,sentinel-users}.acl.
# The chart's own `config-init` then re-seeds them from the ConfigMap with
# hashes matching the Secret — the only way to defeat the runtime ACL drift
# that sentinel/exporter sidecars cause via `ACL SAVE`. Idempotent.
_redis_inject_wipe_acl_init() {
    local ns="$1" sts="$2"
    local image
    image="$(kubectl get sts "${sts}" -n "${ns}" \
        -o jsonpath='{.spec.template.spec.containers[?(@.name=="redis")].image}' 2>/dev/null)"
    [[ -n "${image}" ]] || { log_warn "ACL self-heal: ${ns}/${sts} not found; skip."; return 1; }
    if kubectl get sts "${sts}" -n "${ns}" -o jsonpath='{.spec.template.spec.initContainers[*].name}' 2>/dev/null \
        | tr ' ' '\n' | grep -qx 'wipe-stale-acl'; then
        return 0
    fi
    log_info "Injecting wipe-stale-acl initContainer on ${ns}/${sts} (permanent ACL drift fix)..."
    kubectl patch sts "${sts}" -n "${ns}" --type=strategic -p "$(cat <<EOF
{"spec":{"template":{"spec":{"initContainers":[{"name":"wipe-stale-acl","image":"${image}","imagePullPolicy":"IfNotPresent","command":["sh","-c","rm -f /data/conf/users.acl /data/conf/sentinel-users.acl"],"volumeMounts":[{"name":"redis-datadir","mountPath":"/data"}]}]}}}}
EOF
)" >/dev/null
}

# Recover from a `WRONGPASS` CrashLoop caused by drifted on-disk ACL files.
#
# Background: the proton-redis chart and image hard-code parts of the ACL flow:
#   - templates/_configs.tpl bakes sha256(redis.password) into a ConfigMap users.acl
#   - the image's /config-init.sh only seeds /data/conf/users.acl when it does not
#     already exist on the PVC, otherwise it sed-replaces existing lines (and does
#     NOT re-add lines that got dropped or scrambled at runtime)
#   - sentinel/exporter sidecars run `ACL SETUSER` + `ACL SAVE` during normal
#     operation and helm upgrades, which can rewrite users.acl with a hash that
#     no longer matches the Secret. After a VM/pod restart, the liveness probe
#     AUTHs with the Secret password and Redis answers WRONGPASS in a CrashLoop.
#
# Workaround (preserves data): delete the ACL files and the Pod. The init
# container then re-enters its "if file does not exist" branch and copies fresh
# ACL files from the ConfigMap, with hashes that match the Secret.
fix_redis_acl() {
    local ns="${REDIS_NAMESPACE}"
    local pod="redis-proton-redis-0"

    if ! kubectl get pod "${pod}" -n "${ns}" >/dev/null 2>&1; then
        log_error "Pod ${pod} not found in namespace ${ns}; nothing to fix."
        return 1
    fi

    log_info "Fixing Redis ACL drift on ${ns}/${pod} ..."
    log_info "  Step 1/2: removing /data/conf/{users,sentinel-users}.acl from PVC"
    # Try the redis container first; fall back to sentinel/exporter (they share /data).
    local removed=0
    for c in redis sentinel exporter; do
        if kubectl exec -n "${ns}" "${pod}" -c "${c}" -- \
            sh -c 'rm -f /data/conf/users.acl /data/conf/sentinel-users.acl' 2>/dev/null; then
            removed=1
            break
        fi
    done
    if [[ "${removed}" != "1" ]]; then
        log_warn "Could not exec into any container to remove ACL files (all may be CrashLooping)."
        log_warn "Falling back to deleting the Pod only; if WRONGPASS persists, re-run after the next backoff."
    fi

    log_info "  Step 2/2: deleting Pod to trigger init-container re-seed from ConfigMap"
    kubectl delete pod "${pod}" -n "${ns}" --wait=false >/dev/null 2>&1 || true

    log_info "Waiting for ${pod} to become Ready (up to 180s)..."
    if kubectl wait --for=condition=ready pod "${pod}" -n "${ns}" --timeout=180s 2>/dev/null; then
        log_info "✓ Redis ACL recovered. ${pod} is Ready (3/3)."
        return 0
    fi
    log_error "Pod did not become Ready in 180s. Inspect with: kubectl describe pod ${pod} -n ${ns}"
    return 1
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
