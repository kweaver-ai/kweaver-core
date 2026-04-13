
# Install ingress-nginx-controller
install_ingress_nginx() {
    log_info "Installing ingress-nginx-controller..."

    # Create namespace if not exists
    kubectl create namespace ingress-nginx 2>/dev/null || true

    # Use images from common.sh (already set with kweaver-ai registry)
    # INGRESS_NGINX_CONTROLLER_IMAGE and INGRESS_NGINX_WEBHOOK_CERTGEN_IMAGE are pre-configured

    if helm status ingress-nginx -n ingress-nginx >/dev/null 2>&1; then
        if kubectl -n ingress-nginx rollout status deployment/ingress-nginx-controller --timeout=3s >/dev/null 2>&1; then
            local current_image=""
            current_image="$(kubectl -n ingress-nginx get deploy ingress-nginx-controller -o jsonpath='{.spec.template.spec.containers[0].image}' 2>/dev/null || true)"
            if [[ -n "${current_image}" && "${current_image}" == "${INGRESS_NGINX_CONTROLLER_IMAGE}" ]]; then
                if [[ "${INGRESS_NGINX_ADMISSION_WEBHOOKS_ENABLED}" != "true" ]]; then
                    if kubectl get validatingwebhookconfiguration ingress-nginx-admission >/dev/null 2>&1; then
                        log_warn "ingress-nginx admission webhook still exists but is disabled in config; upgrading/cleaning..."
                    else
                        log_info "ingress-nginx is already installed with desired image (${current_image}), skipping"
                        return 0
                    fi
                else
                    log_info "ingress-nginx is already installed with desired image (${current_image}), skipping"
                    return 0
                fi
            fi
            log_warn "ingress-nginx installed but image differs (${current_image} != ${INGRESS_NGINX_CONTROLLER_IMAGE}); upgrading..."
        fi
    fi

    # Parse image repository/tag from INGRESS_NGINX_CONTROLLER_IMAGE
    local controller_repo="${INGRESS_NGINX_CONTROLLER_IMAGE%:*}"
    local controller_tag="${INGRESS_NGINX_CONTROLLER_IMAGE##*:}"
    if [[ "${controller_repo}" == "${INGRESS_NGINX_CONTROLLER_IMAGE}" || -z "${controller_tag}" ]]; then
        log_error "Invalid INGRESS_NGINX_CONTROLLER_IMAGE (expected repo:tag): ${INGRESS_NGINX_CONTROLLER_IMAGE}"
        return 1
    fi

    # Parse webhook certgen image repository/tag (admission webhook jobs)
    local certgen_repo="${INGRESS_NGINX_WEBHOOK_CERTGEN_IMAGE%:*}"
    local certgen_tag="${INGRESS_NGINX_WEBHOOK_CERTGEN_IMAGE##*:}"
    if [[ "${certgen_repo}" == "${INGRESS_NGINX_WEBHOOK_CERTGEN_IMAGE}" || -z "${certgen_tag}" ]]; then
        log_error "Invalid INGRESS_NGINX_WEBHOOK_CERTGEN_IMAGE (expected repo:tag): ${INGRESS_NGINX_WEBHOOK_CERTGEN_IMAGE}"
        return 1
    fi

    local chart_ref="ingress-nginx/ingress-nginx"
    local use_local_chart="false"
    if [[ -f "${INGRESS_NGINX_CHART_TGZ}" ]]; then
        chart_ref="${INGRESS_NGINX_CHART_TGZ}"
        use_local_chart="true"
        log_info "Using local ingress-nginx chart: ${chart_ref}"
    else
        log_info "Using remote ingress-nginx chart: ${chart_ref} (version ${INGRESS_NGINX_CHART_VERSION})"
    fi

    if [[ "${use_local_chart}" != "true" ]]; then
        helm repo add --force-update ingress-nginx "${HELM_REPO_INGRESS_NGINX}"
        helm repo update
    fi

    local -a helm_args
    helm_args=(
        upgrade --install ingress-nginx "${chart_ref}"
        --namespace ingress-nginx
        --set controller.image.repository="${controller_repo}"
        --set controller.image.tag="${controller_tag}"
        --set-string controller.image.digest=
        --set controller.admissionWebhooks.patch.image.repository="${certgen_repo}"
        --set controller.admissionWebhooks.patch.image.tag="${certgen_tag}"
        --set-string controller.admissionWebhooks.patch.image.digest=
        --set controller.ingressClassResource.enabled=true
        --set controller.ingressClassResource.name="${INGRESS_NGINX_CLASS}"
        --set controller.ingressClassResource.default=true
        --set controller.ingressClass="${INGRESS_NGINX_CLASS}"
        --set controller.config.proxy-body-size="1024m"
        --set defaultBackend.enabled=false
        --wait --timeout=600s
    )

    if [[ "${INGRESS_NGINX_ADMISSION_WEBHOOKS_ENABLED}" != "true" ]]; then
        helm_args+=(
            --set controller.admissionWebhooks.enabled=false
        )
    fi

    if [[ "${INGRESS_NGINX_HOSTNETWORK}" == "true" ]]; then
        helm_args+=(
            --set controller.hostNetwork=true
            --set controller.dnsPolicy=ClusterFirstWithHostNet
            --set controller.service.type=ClusterIP
            --set controller.containerPort.http="${INGRESS_NGINX_HTTP_PORT:-80}"
            --set controller.containerPort.https="${INGRESS_NGINX_HTTPS_PORT:-443}"
            --set controller.hostPort.enabled=true
            --set controller.hostPort.ports.http="${INGRESS_NGINX_HTTP_PORT:-80}"
            --set controller.hostPort.ports.https="${INGRESS_NGINX_HTTPS_PORT:-443}"
            --set controller.extraArgs."http-port"="${INGRESS_NGINX_HTTP_PORT:-80}"
            --set controller.extraArgs."https-port"="${INGRESS_NGINX_HTTPS_PORT:-443}"
        )
    else
        helm_args+=(
            --set controller.service.type=NodePort
            --set controller.service.nodePorts.http="${INGRESS_NGINX_HTTP_PORT}"
            --set controller.service.nodePorts.https="${INGRESS_NGINX_HTTPS_PORT}"
        )
    fi

    if [[ "${use_local_chart}" != "true" && -n "${INGRESS_NGINX_CHART_VERSION}" ]]; then
        helm_args+=(--version "${INGRESS_NGINX_CHART_VERSION}")
    fi

    local helm_out=""
    local helm_rc=0
    set +e
    helm_out="$(helm "${helm_args[@]}" 2>&1)"
    helm_rc=$?
    set -e
    if [[ ${helm_rc} -ne 0 ]]; then
        echo "${helm_out}" >&2
        if [[ "${INGRESS_NGINX_HOSTNETWORK}" != "true" ]] && echo "${helm_out}" | grep -q "nodePort: Invalid value: .*provided port is already allocated"; then
            log_warn "NodePort ${INGRESS_NGINX_HTTP_PORT}/${INGRESS_NGINX_HTTPS_PORT} already allocated; uninstalling existing ingress-nginx release and retrying..."
            helm uninstall ingress-nginx -n ingress-nginx 2>/dev/null || true
            kubectl -n ingress-nginx delete svc -l app.kubernetes.io/instance=ingress-nginx 2>/dev/null || true
            helm "${helm_args[@]}"
        else
            return "${helm_rc}"
        fi
    else
        echo "${helm_out}"
    fi

    if [[ "${INGRESS_NGINX_ADMISSION_WEBHOOKS_ENABLED}" != "true" ]]; then
        # Best-effort cleanup of webhook resources that may linger from previous installs.
        kubectl delete validatingwebhookconfiguration ingress-nginx-admission 2>/dev/null || true
        kubectl delete mutatingwebhookconfiguration ingress-nginx-admission 2>/dev/null || true
        kubectl -n ingress-nginx delete job ingress-nginx-admission-create ingress-nginx-admission-patch 2>/dev/null || true
        kubectl -n ingress-nginx delete secret ingress-nginx-admission 2>/dev/null || true
    fi
    
    log_info "ingress-nginx-controller installed successfully"
    log_info "Ingress-nginx access info:"
    log_info "  Namespace: ingress-nginx"
    log_info "  IngressClass: ${INGRESS_NGINX_CLASS}"
    if [[ "${INGRESS_NGINX_HOSTNETWORK}" == "true" ]]; then
        log_info "  Mode: hostNetwork (ports ${INGRESS_NGINX_HTTP_PORT:-80}/${INGRESS_NGINX_HTTPS_PORT:-443} on node)"
        log_info ""
        log_info "To access the ingress controller:"
        log_info "  http://<node-ip>:${INGRESS_NGINX_HTTP_PORT:-80}"
        log_info "  https://<node-ip>:${INGRESS_NGINX_HTTPS_PORT:-443}"
    else
        log_info "  Service type: NodePort"
        log_info "  HTTP NodePort: ${INGRESS_NGINX_HTTP_PORT}"
        log_info "  HTTPS NodePort: ${INGRESS_NGINX_HTTPS_PORT}"
        log_info ""
        log_info "To access the ingress controller:"
        log_info "  http://<node-ip>:${INGRESS_NGINX_HTTP_PORT}"
        log_info "  https://<node-ip>:${INGRESS_NGINX_HTTPS_PORT}"
    fi

    if [[ "${AUTO_GENERATE_CONFIG}" == "true" ]]; then
        log_info "Calling generate_config_yaml to update config.yaml..."
        generate_config_yaml
        log_info "Config.yaml generation completed"
    else
        log_info "AUTO_GENERATE_CONFIG is false, skipping config generation"
    fi
}

# Uninstall ingress-nginx-controller
uninstall_ingress_nginx() {
    log_info "Uninstalling ingress-nginx-controller from namespace ingress-nginx..."

    # 1. Uninstall Helm release
    if helm status ingress-nginx -n ingress-nginx >/dev/null 2>&1; then
        log_info "Removing Helm release: ingress-nginx"
        helm uninstall ingress-nginx -n ingress-nginx --timeout=120s || true
    else
        log_info "Helm release ingress-nginx not found, skipping Helm uninstall"
    fi

    # 2. Clean up cluster-scoped resources (not removed by namespace deletion)
    kubectl delete validatingwebhookconfiguration ingress-nginx-admission 2>/dev/null || true
    kubectl delete mutatingwebhookconfiguration ingress-nginx-admission 2>/dev/null || true
    kubectl delete ingressclass "${INGRESS_NGINX_CLASS:-class-443}" 2>/dev/null || true

    # 3. If namespace still exists, clean up remaining resources
    if kubectl get namespace ingress-nginx >/dev/null 2>&1; then
        log_info "Cleaning up remaining resources in ingress-nginx namespace..."
        kubectl -n ingress-nginx delete deploy,daemonset,replicaset,statefulset --all --timeout=30s 2>/dev/null || true
        kubectl -n ingress-nginx delete svc,endpoints,configmap,secret,sa,role,rolebinding --all --timeout=30s 2>/dev/null || true
        kubectl -n ingress-nginx delete job --all --timeout=30s 2>/dev/null || true

        # Force-delete any pods stuck in Terminating
        local stuck_pods
        stuck_pods=$(kubectl -n ingress-nginx get pods -o jsonpath='{.items[*].metadata.name}' 2>/dev/null || true)
        if [[ -n "${stuck_pods}" ]]; then
            log_warn "Force-deleting stuck pods: ${stuck_pods}"
            kubectl -n ingress-nginx delete pods --all --force --grace-period=0 2>/dev/null || true
        fi

        # Delete namespace
        kubectl delete namespace ingress-nginx --timeout=60s 2>/dev/null || true
    fi

    log_info "ingress-nginx-controller uninstall done"
}
