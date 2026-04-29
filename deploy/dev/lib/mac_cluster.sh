#!/usr/bin/env bash
# Kind cluster + ingress-nginx (kind provider) for Mac dev. Sourced by mac.sh.

: "${MAC_DEV_ROOT:?MAC_DEV_ROOT must be set}"
: "${KIND_CLUSTER_NAME:?KIND_CLUSTER_NAME must be set}"

INGRESS_NGINX_KIND_MANIFEST_URL="${INGRESS_NGINX_KIND_MANIFEST_URL:-https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml}"

mac_cluster_status() {
    mac_log_info "kind clusters:"
    kind get clusters 2>/dev/null || true
    echo ""
    if kubectl cluster-info >/dev/null 2>&1; then
        mac_log_info "kubectl: $(kubectl config current-context 2>/dev/null || echo '?')"
        kubectl get nodes 2>/dev/null || true
    else
        mac_log_warn "kubectl: no working cluster context"
    fi
}

mac_cluster_up() {
    if kind get clusters 2>/dev/null | grep -qx "${KIND_CLUSTER_NAME}"; then
        mac_log_info "Kind cluster '${KIND_CLUSTER_NAME}' already exists."
    else
        mac_log_info "Creating kind cluster '${KIND_CLUSTER_NAME}'..."
        kind create cluster --name "${KIND_CLUSTER_NAME}" --config "${MAC_DEV_ROOT}/kind-config.yaml"
    fi

    if ! kubectl get ns ingress-nginx >/dev/null 2>&1; then
        mac_log_info "Installing ingress-nginx (official kind provider manifest)..."
        kubectl apply -f "${INGRESS_NGINX_KIND_MANIFEST_URL}"
    else
        mac_log_info "Namespace 'ingress-nginx' already exists; skipping ingress install."
    fi

    mac_log_info "Waiting for ingress-nginx controller (ready)..."
    if ! kubectl wait --namespace ingress-nginx \
        --for=condition=ready pod \
        --selector=app.kubernetes.io/component=controller \
        --timeout=240s; then
        mac_log_warn "ingress-nginx did not become ready in time; check: kubectl -n ingress-nginx get pods"
        return 1
    fi

    mac_log_info "Cluster is up. Use kubectl context: kind-${KIND_CLUSTER_NAME}"
}

mac_cluster_down() {
    mac_log_info "Deleting kind cluster '${KIND_CLUSTER_NAME}'..."
    kind delete cluster --name "${KIND_CLUSTER_NAME}" || true
}

mac_cluster_dispatch() {
    local sub="${1:-status}"
    if [[ $# -gt 0 ]]; then
        shift
    fi
    case "${sub}" in
        up|start|create)
            mac_cluster_up
            ;;
        down|delete|destroy)
            mac_cluster_down
            ;;
        status|"")
            mac_cluster_status
            ;;
        *)
            mac_log_error "Unknown cluster action: ${sub} (expected up|down|status)"
            return 1
            ;;
    esac
}
