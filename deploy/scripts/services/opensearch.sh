
install_opensearch() {
    log_info "Installing OpenSearch via Helm..."

    local fresh_install="true"

    # OpenSearch password handling
    local existing_pass=$(get_existing_password "opensearch.password")
    if [[ -n "${existing_pass}" ]]; then
        OPENSEARCH_INITIAL_ADMIN_PASSWORD="${existing_pass}"
        log_info "Using existing OpenSearch password from config.yaml"
    else
        OPENSEARCH_INITIAL_ADMIN_PASSWORD=$(generate_random_password 10)
        log_info "Generated random 10-character OpenSearch password"
    fi

    if is_helm_installed "${OPENSEARCH_RELEASE_NAME}" "${OPENSEARCH_NAMESPACE}"; then
        fresh_install="false"
        log_info "OpenSearch is already installed. Skipping installation."
        return 0
    fi

    if [[ -z "${OPENSEARCH_INITIAL_ADMIN_PASSWORD}" ]]; then
        log_error "OPENSEARCH_INITIAL_ADMIN_PASSWORD is empty; OpenSearch will not start."
        return 1
    fi

    kubectl create namespace "${OPENSEARCH_NAMESPACE}" 2>/dev/null || true

    if [[ -z "${OPENSEARCH_IMAGE}" ]]; then
        OPENSEARCH_IMAGE="$(image_from_registry "${OPENSEARCH_IMAGE_REPOSITORY}" "${OPENSEARCH_IMAGE_TAG}" "${OPENSEARCH_IMAGE_FALLBACK}")"
    fi

    local persistence_enabled="${OPENSEARCH_PERSISTENCE_ENABLED}"
    if [[ "${persistence_enabled}" == "true" ]]; then
        if [[ -z "$(kubectl get storageclass --no-headers 2>/dev/null)" ]]; then
            log_warn "No StorageClass found. OpenSearch PVC will stay Pending."
            if [[ "${AUTO_INSTALL_LOCALPV}" == "true" ]]; then
                install_localpv
            fi

            if [[ -z "$(kubectl get storageclass --no-headers 2>/dev/null)" ]]; then
                log_warn "Still no StorageClass found. Disabling OpenSearch persistence to avoid Pending PVC."
                persistence_enabled="false"
            fi
        fi
    fi

    # Parse image repository/tag from OPENSEARCH_IMAGE (chart expects image.repository + image.tag)
    local os_image_repo="${OPENSEARCH_IMAGE%:*}"
    local os_image_tag="${OPENSEARCH_IMAGE##*:}"
    if [[ "${os_image_repo}" == "${OPENSEARCH_IMAGE}" || -z "${os_image_tag}" ]]; then
        log_error "Invalid OPENSEARCH_IMAGE (expected repo:tag): ${OPENSEARCH_IMAGE}"
        return 1
    fi

    # Parse init image repo/tag for busybox-based initContainers (chart expects persistence.image/imageTag + sysctlInit.image/imageTag)
    local init_image_repo="${OPENSEARCH_INIT_IMAGE%:*}"
    local init_image_tag="${OPENSEARCH_INIT_IMAGE##*:}"
    if [[ "${init_image_repo}" == "${OPENSEARCH_INIT_IMAGE}" || -z "${init_image_tag}" ]]; then
        log_error "Invalid OPENSEARCH_INIT_IMAGE (expected repo:tag): ${OPENSEARCH_INIT_IMAGE}"
        return 1
    fi

    local chart_ref="opensearch/opensearch"
    local use_local_chart="false"
    if [[ -f "${OPENSEARCH_CHART_TGZ}" ]]; then
        chart_ref="${OPENSEARCH_CHART_TGZ}"
        use_local_chart="true"
        log_info "Using local OpenSearch chart: ${chart_ref}"
    else
        log_info "Using remote OpenSearch chart: ${chart_ref} (version ${OPENSEARCH_CHART_VERSION})"
    fi

    if [[ "${use_local_chart}" != "true" ]]; then
        helm repo add --force-update opensearch "${HELM_REPO_OPENSEARCH}"
        helm repo update
    fi

    local disable_security="${OPENSEARCH_DISABLE_SECURITY}"
    if [[ -z "${disable_security}" ]]; then
        if [[ "${OPENSEARCH_PROTOCOL}" == "http" ]]; then
            disable_security="true"
        else
            disable_security="false"
        fi
    fi

    local tmp_os_yml
    tmp_os_yml="$(mktemp)"
    cat >"${tmp_os_yml}" <<EOF
cluster.name: ${OPENSEARCH_CLUSTER_NAME}
network.host: 0.0.0.0
EOF
    if [[ "${disable_security}" == "true" ]]; then
        cat >>"${tmp_os_yml}" <<EOF
plugins.security.disabled: true
EOF
    fi

    local -a helm_args
    helm_args=(
        upgrade --install "${OPENSEARCH_RELEASE_NAME}" "${chart_ref}"
        --namespace "${OPENSEARCH_NAMESPACE}"
        --set image.repository="${os_image_repo}"
        --set image.tag="${os_image_tag}"
        --set-file config.opensearch\\.yml="${tmp_os_yml}"
        --set persistence.image="${init_image_repo}"
        --set persistence.imageTag="${init_image_tag}"
        --set opensearchJavaOpts="${OPENSEARCH_JAVA_OPTS}"
        --set clusterName="${OPENSEARCH_CLUSTER_NAME}"
        --set nodeGroup="${OPENSEARCH_NODE_GROUP}"
        --set singleNode="${OPENSEARCH_SINGLE_NODE}"
        --set replicas=1
        --set sysctlInit.enabled="${OPENSEARCH_SYSCTL_INIT_ENABLED}"
        --set sysctlInit.image="${init_image_repo}"
        --set sysctlInit.imageTag="${init_image_tag}"
        --set sysctlVmMaxMapCount="${OPENSEARCH_SYSCTL_VM_MAX_MAP_COUNT}"
        --set-string extraEnvs[0].name=OPENSEARCH_INITIAL_ADMIN_PASSWORD
        --set-string extraEnvs[0].value="${OPENSEARCH_INITIAL_ADMIN_PASSWORD}"
        --wait --timeout=900s
    )

    if [[ "${OPENSEARCH_HELM_ATOMIC}" == "true" ]]; then
        helm_args+=(--atomic)
    fi

    if [[ "${use_local_chart}" != "true" ]]; then
        helm_args+=(--version "${OPENSEARCH_CHART_VERSION}")
    fi

    if [[ "${persistence_enabled}" == "true" ]]; then
        helm_args+=(
            --set persistence.enabled=true
            --set persistence.size="${OPENSEARCH_STORAGE_SIZE}"
        )
        if [[ -n "${OPENSEARCH_STORAGE_CLASS}" ]]; then
            helm_args+=(--set persistence.storageClass="${OPENSEARCH_STORAGE_CLASS}")
        fi
    else
        helm_args+=(--set persistence.enabled=false)
    fi

    if [[ -n "${OPENSEARCH_MEMORY_REQUEST}" ]]; then
        helm_args+=(--set resources.requests.memory="${OPENSEARCH_MEMORY_REQUEST}")
    fi
    if [[ -n "${OPENSEARCH_MEMORY_LIMIT}" ]]; then
        helm_args+=(--set resources.limits.memory="${OPENSEARCH_MEMORY_LIMIT}")
    fi

    helm "${helm_args[@]}"
    rm -f "${tmp_os_yml}" 2>/dev/null || true

    local service_name="${OPENSEARCH_CLUSTER_NAME}-${OPENSEARCH_NODE_GROUP}"
    log_info "OpenSearch installed successfully"
    log_info "OpenSearch connection info:"
    log_info "  Service: ${service_name}.${OPENSEARCH_NAMESPACE}.svc.cluster.local"
    log_info "  Port: 9200"
    log_info "  Username: admin"
    log_info "  Password: ${OPENSEARCH_INITIAL_ADMIN_PASSWORD}"
    if [[ "${disable_security}" == "true" ]]; then
        log_warn "  Security: disabled (HTTP, no auth)"
    else
        log_info "  Security: enabled (HTTPS + basic auth)"
    fi

    if [[ "${fresh_install}" == "true" && "${AUTO_GENERATE_CONFIG}" == "true" ]]; then
        log_info "Updating conf/config.yaml after OpenSearch fresh install..."
        generate_config_yaml
    fi
}

uninstall_opensearch() {
    log_info "Uninstalling OpenSearch from namespace ${OPENSEARCH_NAMESPACE}..."

    helm uninstall "${OPENSEARCH_RELEASE_NAME}" -n "${OPENSEARCH_NAMESPACE}" 2>/dev/null || true

    # Delete PVCs (this will also trigger PV deletion if reclaim policy is Delete)
    if [[ "${OPENSEARCH_PURGE_PVC}" == "true" ]]; then
        log_warn "OPENSEARCH_PURGE_PVC=true: deleting OpenSearch PVCs (data loss!)"
        
        # Get all PVCs related to OpenSearch before deletion
        local pvc_names
        pvc_names=$(kubectl get pvc -n "${OPENSEARCH_NAMESPACE}" -l "app.kubernetes.io/instance=${OPENSEARCH_RELEASE_NAME}" -o jsonpath='{.items[*].metadata.name}' 2>/dev/null || true)
        
        # Delete PVCs by label
        kubectl delete pvc -n "${OPENSEARCH_NAMESPACE}" -l "app.kubernetes.io/instance=${OPENSEARCH_RELEASE_NAME}" 2>/dev/null || true
        # Also try to delete by name pattern (OpenSearch chart naming)
        kubectl delete pvc -n "${OPENSEARCH_NAMESPACE}" -l "app.kubernetes.io/name=opensearch" 2>/dev/null || true
        # Try to delete PVCs by name pattern
        local pvc_name="${OPENSEARCH_CLUSTER_NAME}-${OPENSEARCH_NODE_GROUP}-${OPENSEARCH_CLUSTER_NAME}-${OPENSEARCH_NODE_GROUP}-0"
        kubectl delete pvc -n "${OPENSEARCH_NAMESPACE}" "${pvc_name}" 2>/dev/null || true
        
        # Wait a bit for PVs to be released
        sleep 2
        
        # Try to delete PVs that are in Released state
        log_info "Cleaning up Released PVs..."
        local released_pvs
        released_pvs=$(kubectl get pv -o jsonpath='{.items[?(@.status.phase=="Released")].metadata.name}' 2>/dev/null || true)
        if [[ -n "${released_pvs}" ]]; then
            for pv in ${released_pvs}; do
                # Check if this PV was bound to one of the deleted PVCs
                local pv_claim
                pv_claim=$(kubectl get pv "${pv}" -o jsonpath='{.spec.claimRef.name}' 2>/dev/null || true)
                if [[ -n "${pv_claim}" ]] && echo "${pvc_names}" | grep -q "${pv_claim}"; then
                    log_info "Deleting Released PV: ${pv}"
                    kubectl delete pv "${pv}" 2>/dev/null || true
                fi
            done
        fi
    else
        # Even if PURGE_PVC is false, try to find and delete orphaned PVCs
        log_info "Checking for orphaned PVCs..."
        local orphaned_pvcs
        orphaned_pvcs=$(kubectl get pvc -n "${OPENSEARCH_NAMESPACE}" -l "app.kubernetes.io/instance=${OPENSEARCH_RELEASE_NAME}" -o name 2>/dev/null || true)
        if [[ -n "${orphaned_pvcs}" ]]; then
            log_warn "Found orphaned PVCs. Use OPENSEARCH_PURGE_PVC=true to delete them."
        fi
    fi

    log_info "OpenSearch uninstall done"
}
