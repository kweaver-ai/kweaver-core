
install_kafka() {
    log_info "Installing Kafka (1 controller + 1 broker) via Helm..."

    local fresh_install="true"
    if is_helm_installed "${KAFKA_RELEASE_NAME}" "${KAFKA_NAMESPACE}"; then
        fresh_install="false"
        log_info "Kafka is already installed. Skipping installation."
        return 0
    fi

    # Kafka password handling
    local existing_pass
    existing_pass="$(get_existing_password "kafka.password")"
    if [[ -n "${existing_pass}" ]]; then
        KAFKA_CLIENT_PASSWORD="${existing_pass}"
        log_info "Using existing Kafka password from config.yaml"
    else
        KAFKA_CLIENT_PASSWORD="$(generate_random_password 10)"
        if [[ -z "${KAFKA_CLIENT_PASSWORD}" ]]; then
            log_error "Failed to generate Kafka password (install openssl or set kafka.password in config)."
            return 1
        fi
        log_info "Generated random 10-character Kafka password"
    fi

    if ! command -v helm >/dev/null 2>&1; then
        log_error "Helm is required to install Kafka. Please run: $0 k8s init"
        return 1
    fi

    if ! kweaver_require_helm_min_for_bitnami; then
        return 1
    fi

    if [[ -z "${KAFKA_IMAGE}" ]]; then
        KAFKA_IMAGE="$(image_from_registry "${KAFKA_IMAGE_REPOSITORY}" "${KAFKA_IMAGE_TAG}" "${KAFKA_IMAGE_FALLBACK}")"
    fi

    if [[ "${KAFKA_AUTH_ENABLED}" == "true" ]]; then
        # Ensure passwords exist (and persist via a Secret for idempotency)
        if kubectl -n "${KAFKA_NAMESPACE}" get secret "${KAFKA_SASL_SECRET_NAME}" >/dev/null 2>&1; then
            if [[ -z "${KAFKA_CLIENT_PASSWORD}" ]]; then
                KAFKA_CLIENT_PASSWORD="$(get_secret_b64_key "${KAFKA_NAMESPACE}" "${KAFKA_SASL_SECRET_NAME}" client-passwords)"
                KAFKA_CLIENT_PASSWORD="${KAFKA_CLIENT_PASSWORD%%,*}"
            fi
            if [[ -z "${KAFKA_INTERBROKER_PASSWORD}" ]]; then
                KAFKA_INTERBROKER_PASSWORD="$(get_secret_b64_key "${KAFKA_NAMESPACE}" "${KAFKA_SASL_SECRET_NAME}" inter-broker-password)"
            fi
            if [[ -z "${KAFKA_CONTROLLER_PASSWORD}" ]]; then
                KAFKA_CONTROLLER_PASSWORD="$(get_secret_b64_key "${KAFKA_NAMESPACE}" "${KAFKA_SASL_SECRET_NAME}" controller-password)"
            fi
        fi

        if [[ -z "${KAFKA_CLIENT_PASSWORD}" ]]; then
            KAFKA_CLIENT_PASSWORD="$(random_password)"
        fi
        if [[ -z "${KAFKA_INTERBROKER_PASSWORD}" ]]; then
            KAFKA_INTERBROKER_PASSWORD="$(random_password)"
        fi
        if [[ -z "${KAFKA_CONTROLLER_PASSWORD}" ]]; then
            KAFKA_CONTROLLER_PASSWORD="$(random_password)"
        fi

        cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: ${KAFKA_SASL_SECRET_NAME}
  namespace: ${KAFKA_NAMESPACE}
type: Opaque
stringData:
  client-passwords: "${KAFKA_CLIENT_PASSWORD}"
  inter-broker-password: "${KAFKA_INTERBROKER_PASSWORD}"
  controller-password: "${KAFKA_CONTROLLER_PASSWORD}"
EOF

        # Force SASL listeners when auth is enabled
        KAFKA_PROTOCOL="SASL_PLAINTEXT"
    fi

    local persistence_enabled="${KAFKA_PERSISTENCE_ENABLED}"
    if [[ "${persistence_enabled}" == "true" ]]; then
        if [[ -z "$(kubectl get storageclass --no-headers 2>/dev/null)" ]]; then
            log_warn "No StorageClass found. Kafka PVC will stay Pending."
            if [[ "${AUTO_INSTALL_LOCALPV}" == "true" ]]; then
                install_localpv
            fi

            if [[ -z "$(kubectl get storageclass --no-headers 2>/dev/null)" ]]; then
                log_warn "Still no StorageClass found. Disabling Kafka persistence to avoid Pending PVC."
                persistence_enabled="false"
            fi
        fi
    fi

    # Parse image registry/repository/tag from KAFKA_IMAGE
    local image_without_tag="${KAFKA_IMAGE%:*}"
    local image_tag="${KAFKA_IMAGE##*:}"
    local image_registry="${image_without_tag%%/*}"
    local image_repo="${image_without_tag#*/}"

    local chart_ref="bitnami/kafka"
    local use_local_chart="false"
    if [[ -f "${KAFKA_CHART_TGZ}" ]]; then
        chart_ref="${KAFKA_CHART_TGZ}"
        use_local_chart="true"
        log_info "Using local Kafka chart: ${chart_ref}"
    else
        log_info "Using remote Kafka chart: ${chart_ref} (version ${KAFKA_CHART_VERSION})"
    fi

    if [[ "${use_local_chart}" != "true" ]]; then
        helm repo add --force-update bitnami "${HELM_REPO_BITNAMI}"
        helm repo update
    fi

    # Temporary values (XXXXXX must be at end of template for macOS mktemp)
    local tmp_values
    tmp_values="$(mktemp "${TMPDIR:-/tmp}/kafka-values.XXXXXX")"
    
    # KRaft requires controller.quorum.voters (or --initial-controllers during storage format).
    # Without it, controller pods will CrashLoop with:
    #   Because controller.quorum.voters is not set on this controller, you must specify ... --initial-controllers ...
    local quorum_voters=""
    local i
    for ((i = 0; i < KAFKA_REPLICAS; i++)); do
        local voter="${i}@${KAFKA_RELEASE_NAME}-controller-${i}.${KAFKA_RELEASE_NAME}-controller-headless.${KAFKA_NAMESPACE}.svc.cluster.local:9093"
        if [[ -z "${quorum_voters}" ]]; then
            quorum_voters="${voter}"
        else
            quorum_voters="${quorum_voters},${voter}"
        fi
    done
    
    # Convert boolean to string for Kafka server.properties (true/false -> "true"/"false")
    local auto_create_topics_value="true"
    if [[ "${KAFKA_AUTO_CREATE_TOPICS_ENABLE}" != "true" ]]; then
        auto_create_topics_value="false"
    fi
    
    cat > "${tmp_values}" <<EOF
# Global overrideConfiguration (applies to both controller and broker)
# Both controller and broker need controller.quorum.voters to know the controller quorum addresses
overrideConfiguration:
  "controller.quorum.voters": "${quorum_voters}"
  "auto.create.topics.enable": "${auto_create_topics_value}"
EOF

    if [[ "${KAFKA_REPLICAS}" == "1" ]]; then
        cat >> "${tmp_values}" <<EOF
  "default.replication.factor": 1
  "min.insync.replicas": 1
  "offsets.topic.replication.factor": 1
  "transaction.state.log.replication.factor": 1
  "transaction.state.log.min.isr": 1
EOF
    fi

    cat >> "${tmp_values}" <<EOF
listeners:
  # Avoid clients dialing per-pod headless DNS from advertised.listeners; advertise the stable service DNS instead.
  advertisedListeners: "CLIENT://${KAFKA_RELEASE_NAME}.${KAFKA_NAMESPACE}.svc.cluster.local:9092,INTERNAL://${KAFKA_RELEASE_NAME}.${KAFKA_NAMESPACE}.svc.cluster.local:9094"
EOF

    local -a helm_args
    helm_args=(
        upgrade --install "${KAFKA_RELEASE_NAME}" "${chart_ref}"
        --namespace "${KAFKA_NAMESPACE}"
        -f "${tmp_values}"
        --set global.security.allowInsecureImages=true
        --set image.registry="${image_registry}"
        --set image.repository="${image_repo}"
        --set image.tag="${image_tag}"
        # KRaft: run controller-eligible nodes as controllers only, and run dedicated broker nodes for client traffic.
        # This prevents clients from ever landing on controller pods (which would break consumer groups / JoinGroup).
        --set controller.replicaCount="${KAFKA_REPLICAS}"
        --set controller.controllerOnly=true
        --set broker.replicaCount="${KAFKA_REPLICAS}"
        --set broker.heapOpts="${KAFKA_HEAP_OPTS}"
        # Client/inter-broker can be SASL, but controller listener should stay PLAINTEXT for compatibility
        --set listeners.client.protocol="${KAFKA_PROTOCOL}"
        --set listeners.controller.protocol=PLAINTEXT
        --set listeners.interbroker.protocol="${KAFKA_PROTOCOL}"
        --wait --timeout="${KAFKA_HELM_TIMEOUT}"
    )

    if [[ "${KAFKA_HELM_ATOMIC}" == "true" ]]; then
        helm_args+=(--atomic)
    fi

    if [[ "${KAFKA_AUTH_ENABLED}" == "true" ]]; then
        helm_args+=(
            --set sasl.enabledMechanisms="${KAFKA_SASL_MECHANISM}"
            --set sasl.interBrokerMechanism="${KAFKA_SASL_MECHANISM}"
            --set sasl.interbroker.user="${KAFKA_INTERBROKER_USER}"
            --set sasl.client.users[0]="${KAFKA_CLIENT_USER}"
            --set sasl.existingSecret="${KAFKA_SASL_SECRET_NAME}"
        )
    fi

    if [[ -n "${KAFKA_MEMORY_REQUEST}" || -n "${KAFKA_MEMORY_LIMIT}" ]]; then
        helm_args+=(--set controller.resourcesPreset=none --set broker.resourcesPreset=none)
        if [[ -n "${KAFKA_MEMORY_REQUEST}" ]]; then
            helm_args+=(--set controller.resources.requests.memory="${KAFKA_MEMORY_REQUEST}" --set broker.resources.requests.memory="${KAFKA_MEMORY_REQUEST}")
        fi
        if [[ -n "${KAFKA_MEMORY_LIMIT}" ]]; then
            helm_args+=(--set controller.resources.limits.memory="${KAFKA_MEMORY_LIMIT}" --set broker.resources.limits.memory="${KAFKA_MEMORY_LIMIT}")
        fi
    fi

    if [[ "${use_local_chart}" != "true" ]]; then
        helm_args+=(--version "${KAFKA_CHART_VERSION}")
    fi

    if [[ "${persistence_enabled}" == "true" ]]; then
        helm_args+=(
            --set controller.persistence.enabled=true
            --set controller.persistence.size="${KAFKA_STORAGE_SIZE}"
            --set broker.persistence.enabled=true
            --set broker.persistence.size="${KAFKA_STORAGE_SIZE}"
        )
        if [[ -n "${KAFKA_STORAGE_CLASS}" ]]; then
            helm_args+=(
                --set controller.persistence.storageClass="${KAFKA_STORAGE_CLASS}"
                --set broker.persistence.storageClass="${KAFKA_STORAGE_CLASS}"
            )
        fi
    else
        helm_args+=(--set controller.persistence.enabled=false --set broker.persistence.enabled=false)
    fi

    kweaver_helm_uninstall_if_not_deployed "${KAFKA_RELEASE_NAME}" "${KAFKA_NAMESPACE}"

    local helm_out=""
    set +e
    helm_out="$(helm "${helm_args[@]}" 2>&1)"
    local helm_rc=$?
    set -e
    rm -f "${tmp_values}" 2>/dev/null || true

    if [[ ${helm_rc} -ne 0 ]]; then
        log_error "Kafka Helm install/upgrade failed (exit code: ${helm_rc})."
        echo "${helm_out}" >&2
        log_info "Collecting Kafka diagnostics (pods/events/pvc)..."
        kubectl -n "${KAFKA_NAMESPACE}" get pods -o wide 2>/dev/null || true
        kubectl -n "${KAFKA_NAMESPACE}" get pvc 2>/dev/null || true
        kubectl -n "${KAFKA_NAMESPACE}" get svc,endpoints 2>/dev/null || true
        kubectl -n "${KAFKA_NAMESPACE}" get events --sort-by=.lastTimestamp 2>/dev/null | tail -n 60 || true
        return ${helm_rc}
    fi

    log_info "Kafka installed successfully"
    log_info "Kafka connection info:"
    log_info "  Namespace: ${KAFKA_NAMESPACE}"
    log_info "  Release: ${KAFKA_RELEASE_NAME}"
    log_info "  Port: 9092"
    if [[ "${KAFKA_AUTH_ENABLED}" == "true" ]]; then
        log_info "  SASL enabled: true"
        log_info "  Client Password: ${KAFKA_CLIENT_PASSWORD}"
    fi

    # Post-install health check (pods might exist but be CrashLooping; Helm can still "succeed").
    if ! kubectl -n "${KAFKA_NAMESPACE}" wait --for=condition=Ready pod -l "app.kubernetes.io/instance=${KAFKA_RELEASE_NAME}" --timeout="${KAFKA_READY_TIMEOUT}" >/dev/null 2>&1; then
        log_error "Kafka pods are not Ready within ${KAFKA_READY_TIMEOUT}."
        log_info "Kafka pods:"
        kubectl -n "${KAFKA_NAMESPACE}" get pod -o wide 2>/dev/null || true
        log_info "Recent events:"
        kubectl -n "${KAFKA_NAMESPACE}" get events --sort-by=.lastTimestamp 2>/dev/null | tail -n 80 || true
        log_info "Describe controller/broker (best-effort):"
        kubectl -n "${KAFKA_NAMESPACE}" describe pod "${KAFKA_RELEASE_NAME}-controller-0" 2>/dev/null | tail -n 200 || true
        kubectl -n "${KAFKA_NAMESPACE}" describe pod "${KAFKA_RELEASE_NAME}-broker-0" 2>/dev/null | tail -n 200 || true
        log_info "Logs (best-effort):"
        kubectl -n "${KAFKA_NAMESPACE}" logs "${KAFKA_RELEASE_NAME}-controller-0" --tail=200 2>/dev/null || true
        kubectl -n "${KAFKA_NAMESPACE}" logs "${KAFKA_RELEASE_NAME}-controller-0" --previous --tail=200 2>/dev/null || true
        kubectl -n "${KAFKA_NAMESPACE}" logs "${KAFKA_RELEASE_NAME}-broker-0" --tail=200 2>/dev/null || true
        kubectl -n "${KAFKA_NAMESPACE}" logs "${KAFKA_RELEASE_NAME}-broker-0" --previous --tail=200 2>/dev/null || true
        return 1
    fi

    if [[ "${fresh_install}" == "true" && "${AUTO_GENERATE_CONFIG}" == "true" ]]; then
        log_info "Updating conf/config.yaml after Kafka fresh install..."
        generate_config_yaml
    fi
}

uninstall_kafka() {
    log_info "Uninstalling Kafka from namespace ${KAFKA_NAMESPACE}..."

    helm uninstall "${KAFKA_RELEASE_NAME}" -n "${KAFKA_NAMESPACE}" 2>/dev/null || true

    if [[ "${KAFKA_PURGE_PVC}" == "true" ]]; then
        log_info "Deleting Kafka PVCs..."

        # Capture PV names bound to Kafka PVCs (so we can clean up PVs after PVC deletion)
        local pv_names=()
        local pvc_list
        pvc_list="$(kubectl -n "${KAFKA_NAMESPACE}" get pvc -l "app.kubernetes.io/instance=${KAFKA_RELEASE_NAME}" -o jsonpath='{range .items[*]}{.metadata.name}{"|"}{.spec.volumeName}{"\n"}{end}' 2>/dev/null || true)"
        if [[ -n "${pvc_list}" ]]; then
            while IFS="|" read -r pvc_name pv_name; do
                if [[ -n "${pv_name}" ]]; then
                    pv_names+=("${pv_name}")
                fi
            done <<< "${pvc_list}"
        fi

        # Delete PVCs (common label sets + fallback patterns)
        kubectl delete pvc -n "${KAFKA_NAMESPACE}" -l "app.kubernetes.io/instance=${KAFKA_RELEASE_NAME}" 2>/dev/null || true
        kubectl delete pvc -n "${KAFKA_NAMESPACE}" -l "app.kubernetes.io/name=kafka" 2>/dev/null || true
        # Common Bitnami PVC names (best-effort)
        kubectl delete pvc -n "${KAFKA_NAMESPACE}" -l "app=${KAFKA_RELEASE_NAME}" 2>/dev/null || true

        # Best-effort PV cleanup: delete PVs that become Released
        if [[ ${#pv_names[@]} -gt 0 ]]; then
            local pv
            for pv in "${pv_names[@]}"; do
                # Wait briefly for PV to transition after PVC deletion
                sleep 1
                local phase
                phase="$(kubectl get pv "${pv}" -o jsonpath='{.status.phase}' 2>/dev/null || true)"
                if [[ "${phase}" == "Released" ]]; then
                    kubectl delete pv "${pv}" 2>/dev/null || true
                fi
            done
        fi
    else
        log_warn "KAFKA_PURGE_PVC=false: Kafka PVCs were retained."
    fi

    log_info "Kafka uninstall done"
}
