
# Install Zookeeper via Helm
install_zookeeper() {
    log_info "Installing Zookeeper via Helm..."

    kubectl create namespace "${ZOOKEEPER_NAMESPACE}" 2>/dev/null || true

    local fresh_install="true"
    if is_helm_installed "${ZOOKEEPER_RELEASE_NAME}" "${ZOOKEEPER_NAMESPACE}"; then
        fresh_install="false"
        log_info "Zookeeper is already installed. Skipping installation."
        return 0
    fi

    # Check for StorageClass if persistence is enabled
    local storage_class="${ZOOKEEPER_STORAGE_CLASS}"
    if [[ -z "${storage_class}" ]]; then
        if [[ -z "$(kubectl get storageclass --no-headers 2>/dev/null)" ]]; then
            log_warn "No StorageClass found. Zookeeper PVC will stay Pending."
            if [[ "${AUTO_INSTALL_LOCALPV}" == "true" ]]; then
                install_localpv
            fi
            # Try to use local-path if available
            if kubectl get storageclass local-path >/dev/null 2>&1; then
                storage_class="local-path"
                log_info "Using local-path StorageClass for Zookeeper"
            fi
        else
            # Use first available StorageClass
            storage_class="$(kubectl get storageclass -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)"
            log_info "Using StorageClass: ${storage_class}"
        fi
    fi

    # Determine chart reference (prefer local tgz in production)
    local chart_ref="${ZOOKEEPER_CHART_TGZ}"
    if [[ -z "${chart_ref}" ]]; then
        chart_ref="${SCRIPT_DIR}/charts/proton-zookeeper-5.6.0.tgz"
    fi
    if [[ ! -f "${chart_ref}" ]]; then
        log_error "Zookeeper chart tgz not found: ${chart_ref}"
        return 1
    fi

    # Fresh install password generation (10 chars) for SASL
    if [[ "${ZOOKEEPER_SASL_ENABLED}" == "true" && -z "${ZOOKEEPER_SASL_PASSWORD}" ]]; then
        ZOOKEEPER_SASL_PASSWORD="$(generate_random_password 10)"
    fi

    # Temporary values file (template must end with XXXXXX for macOS mktemp)
    local tmp_values
    tmp_values="$(mktemp "${TMPDIR:-/tmp}/zookeeper-values.XXXXXX")"
    cat > "${tmp_values}" <<EOF
namespace: ${ZOOKEEPER_NAMESPACE}
replicaCount: ${ZOOKEEPER_REPLICAS}
antiAffinity:
  enabled: false
image:
  registry: ${ZOOKEEPER_IMAGE_REGISTRY}
  zookeeper:
    repository: ${ZOOKEEPER_IMAGE_REPOSITORY}
    tag: ${ZOOKEEPER_IMAGE_TAG}
  exporter:
    repository: ${ZOOKEEPER_EXPORTER_IMAGE_REPOSITORY}
    tag: ${ZOOKEEPER_EXPORTER_IMAGE_TAG}
service:
  zookeeper:
    port: ${ZOOKEEPER_SERVICE_PORT}
  exporter:
    port: ${ZOOKEEPER_EXPORTER_PORT}
  jmxExporter:
    port: ${ZOOKEEPER_JMX_EXPORTER_PORT}
config:
  zookeeperENV:
    JVMFLAGS: ${ZOOKEEPER_JVMFLAGS}
  sasl:
    enabled: ${ZOOKEEPER_SASL_ENABLED}
EOF

    # Add SASL users if enabled
    if [[ "${ZOOKEEPER_SASL_ENABLED}" == "true" ]]; then
        cat >> "${tmp_values}" <<EOF
    user:
      - username: ${ZOOKEEPER_SASL_USER}
        password: ${ZOOKEEPER_SASL_PASSWORD}
EOF
    fi

    # Add storage and resources
    # Note: Chart logic:
    # - If storageClassName is set (non-empty), uses that StorageClass
    # - If storageClassName is empty/not set, chart checks storage.local for local PVs
    # - If neither, PVC will use cluster's default StorageClass (if exists)
    cat >> "${tmp_values}" <<EOF
storage:
  capacity: ${ZOOKEEPER_STORAGE_SIZE}
EOF

    # Set storageClassName if we have one
    if [[ -n "${storage_class}" ]]; then
        cat >> "${tmp_values}" <<EOF
  storageClassName: ${storage_class}
EOF
        log_info "Using StorageClass: ${storage_class}"
    else
        # If no StorageClass found, set to empty string
        # Chart will check storage.local for local storage, or use default StorageClass
        cat >> "${tmp_values}" <<EOF
  storageClassName: ""
EOF
        log_info "No StorageClass specified, using empty string (chart will handle)"
    fi

    cat >> "${tmp_values}" <<EOF
resources:
  requests:
    cpu: ${ZOOKEEPER_RESOURCES_REQUESTS_CPU}
    memory: ${ZOOKEEPER_RESOURCES_REQUESTS_MEMORY}
  limits:
    cpu: ${ZOOKEEPER_RESOURCES_LIMITS_CPU}
    memory: ${ZOOKEEPER_RESOURCES_LIMITS_MEMORY}
EOF

    # Install via Helm with values file
    log_info "Installing Zookeeper Helm chart..."
    log_info "Chart tgz: ${chart_ref}"
    log_info "Release name: ${ZOOKEEPER_RELEASE_NAME}"
    log_info "Namespace: ${ZOOKEEPER_NAMESPACE}"
    
    local helm_cmd=(
        helm upgrade --install "${ZOOKEEPER_RELEASE_NAME}" "${chart_ref}"
        --namespace "${ZOOKEEPER_NAMESPACE}"
        -f "${tmp_values}"
    )
    
    # Add additional values file if specified
    if [[ -n "${ZOOKEEPER_VALUES_FILE}" ]]; then
        helm_cmd+=(-f "${ZOOKEEPER_VALUES_FILE}")
    fi
    
    # Add extra --set values if specified
    if [[ -n "${ZOOKEEPER_EXTRA_SET_VALUES}" ]]; then
        helm_cmd+=(--set ${ZOOKEEPER_EXTRA_SET_VALUES})
    fi
    
    # Add chart version if specified
    if [[ -n "${ZOOKEEPER_CHART_VERSION}" ]]; then
        helm_cmd+=(--version "${ZOOKEEPER_CHART_VERSION}")
    fi
    
    # Add --devel flag if specified
    if [[ "${ZOOKEEPER_CHART_DEVEL}" == "true" ]]; then
        helm_cmd+=(--devel)
    fi
    
    helm_cmd+=(--wait --timeout=600s)
    
    log_info "Running: ${helm_cmd[*]}"
    
    local helm_output
    helm_output=$("${helm_cmd[@]}" 2>&1)
    local helm_exit_code=$?
    
    rm -f "${tmp_values}" 2>/dev/null || true
    
    if [[ ${helm_exit_code} -ne 0 ]]; then
        log_error "Failed to install Zookeeper"
        log_error "Helm command exit code: ${helm_exit_code}"
        log_error "Helm output:"
        echo "${helm_output}" | while IFS= read -r line; do
            log_error "  ${line}"
        done
        return 1
    fi

    log_info "Zookeeper installed successfully"
    
    # Wait for Zookeeper to be ready
    log_info "Waiting for Zookeeper Pod to be ready..."
    kubectl wait --for=condition=ready pod -l "app=${ZOOKEEPER_RELEASE_NAME}" -n "${ZOOKEEPER_NAMESPACE}" --timeout=300s 2>/dev/null || {
        log_warn "Zookeeper Pod may not be ready yet"
    }
    
    log_info "Zookeeper connection info:"
    log_info "  Service: ${ZOOKEEPER_RELEASE_NAME}-headless.${ZOOKEEPER_NAMESPACE}.svc.cluster.local"
    log_info "  Port: ${ZOOKEEPER_SERVICE_PORT}"
    if [[ "${ZOOKEEPER_SASL_ENABLED}" == "true" ]]; then
        log_info "  SASL enabled: true"
        log_info "  SASL user: ${ZOOKEEPER_SASL_USER}"
    fi

    if [[ "${fresh_install}" == "true" && "${AUTO_GENERATE_CONFIG}" == "true" ]]; then
        generate_config_yaml
    fi
}

# Uninstall Zookeeper
uninstall_zookeeper() {
    log_info "Uninstalling Zookeeper from namespace ${ZOOKEEPER_NAMESPACE}..."

    helm uninstall "${ZOOKEEPER_RELEASE_NAME}" -n "${ZOOKEEPER_NAMESPACE}" 2>/dev/null || true

    # Delete PVCs by default (Zookeeper PVCs are deleted on uninstall)
    if [[ "${ZOOKEEPER_PURGE_PVC}" == "true" ]]; then
        log_info "Deleting Zookeeper PVCs..."
        
        # Get all PVCs related to Zookeeper before deletion
        local pvc_names
        pvc_names=$(kubectl get pvc -n "${ZOOKEEPER_NAMESPACE}" -l "app=${ZOOKEEPER_RELEASE_NAME}" -o jsonpath='{.items[*].metadata.name}' 2>/dev/null || true)
        
        # Delete PVCs by label
        kubectl delete pvc -n "${ZOOKEEPER_NAMESPACE}" -l "app=${ZOOKEEPER_RELEASE_NAME}" 2>/dev/null || true
        
        # Also try to delete by name pattern
        local i
        for ((i=0; i<ZOOKEEPER_REPLICAS; i++)); do
            kubectl delete pvc -n "${ZOOKEEPER_NAMESPACE}" "data-${ZOOKEEPER_RELEASE_NAME}-${i}" 2>/dev/null || true
        done
        
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
        log_warn "ZOOKEEPER_PURGE_PVC=false: Zookeeper PVCs were retained."
    fi

    log_info "Zookeeper uninstall done"
}
