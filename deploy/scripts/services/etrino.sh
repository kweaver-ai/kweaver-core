#!/bin/bash

# Etrino Deployment Script
# Supports install/status/uninstall for vega-hdfs, vega-calculate, and vega-metadata.

NAMESPACE="${NAMESPACE:-kweaver}"
CONFIG_FILE="${CONFIG_FILE:-$HOME/.kweaver-ai/config.yaml}"

VEGA_METADATA_VERSION="3.3.1-release"
VEGA_CALCULATE_VERSION="3.3.4-release"
VEGA_HDFS_VERSION="3.1.0-release"

usage() {
    echo "Usage: $0 [install|status|uninstall] [--config <path>] [--namespace <name>]"
    echo ""
    echo "Defaults to 'install' when no action is provided."
}

parse_args() {
    local action="${1:-install}"
    if [[ "${action}" == --* ]]; then
        action="install"
    else
        shift || true
    fi

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --config=*)
                CONFIG_FILE="${1#*=}"
                shift
                ;;
            --config)
                CONFIG_FILE="$2"
                shift 2
                ;;
            --namespace=*)
                NAMESPACE="${1#*=}"
                shift
                ;;
            --namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                echo "Unknown argument: $1"
                usage
                exit 1
                ;;
        esac
    done

    echo "${action}"
}

load_namespace_from_config() {
    if [[ -f "${CONFIG_FILE}" ]]; then
        local config_namespace
        config_namespace="$(awk '$1=="namespace:"{print $2; exit}' "${CONFIG_FILE}" 2>/dev/null | sed -e 's/^["'\'']//; s/["'\'']$//' | tr -d '\r')"
        NAMESPACE="${config_namespace:-${NAMESPACE}}"
    fi
}

check_helm_release() {
    local release_name="$1"
    helm list -n "${NAMESPACE}" 2>/dev/null | grep -q "^${release_name}$"
}

check_install_state() {
    echo "Checking Etrino services installation status in namespace: ${NAMESPACE}..."
    VEGA_HDFS_INSTALLED=false
    VEGA_CALCULATE_INSTALLED=false
    VEGA_METADATA_INSTALLED=false

    if check_helm_release "vega-hdfs"; then
        echo "✓ vega-hdfs is already installed"
        VEGA_HDFS_INSTALLED=true
    else
        echo "✗ vega-hdfs is not installed"
    fi

    if check_helm_release "vega-calculate"; then
        echo "✓ vega-calculate is already installed"
        VEGA_CALCULATE_INSTALLED=true
    else
        echo "✗ vega-calculate is not installed"
    fi

    if check_helm_release "vega-metadata"; then
        echo "✓ vega-metadata is already installed"
        VEGA_METADATA_INSTALLED=true
    else
        echo "✗ vega-metadata is not installed"
    fi
}

ensure_node_labels_and_dirs() {
    echo "Getting node names from Kubernetes cluster..."
    local nodes
    nodes=($(kubectl get nodes -o jsonpath='{.items[*].metadata.name}'))

    echo "Adding labels to nodes..."
    local i node label_name
    for i in "${!nodes[@]}"; do
        node="${nodes[$i]}"
        label_name="node${i}"
        if kubectl get node "${node}" -o jsonpath='{.metadata.labels}' | grep -q "aishu.io/hostname"; then
            echo "  Node ${node} already has aishu.io/hostname label, skipping"
        else
            echo "  Labeling node ${node} with: aishu.io/hostname=${label_name}"
            kubectl label nodes "${node}" aishu.io/hostname="${label_name}" --overwrite
        fi
    done
    echo "Node labeling completed!"

    echo "Creating required directories on nodes..."
    local required_dirs
    required_dirs="/sysvol/journalnode/mycluster /sysvol/namenode /sysvol/datanode /sysvol/namenode-slaves"

    for node in "${nodes[@]}"; do
        echo "Checking directories on node ${node}..."
        if [[ "${#nodes[@]}" -eq 1 ]]; then
            echo "  Single-node cluster detected, creating directories locally..."
            mkdir -p ${required_dirs}
            echo "  Directories created locally"
        else
            echo "  Attempting SSH to ${node}..."
            if ssh -o ConnectTimeout=5 -o StrictHostKeyChecking=no "${node}" "mkdir -p ${required_dirs}" 2>/dev/null; then
                echo "  Directories created via SSH on ${node}"
            else
                echo "  WARNING: Cannot SSH to ${node}. Please manually create directories: ${required_dirs}"
            fi
        fi
    done
    echo "Directory creation completed!"
}

ensure_helm_repo() {
    echo "Adding Helm repository..."
    if helm repo list 2>/dev/null | grep -q "^myrepo"; then
        echo "Helm repo 'myrepo' already exists, updating..."
        helm repo update myrepo
    else
        echo "Adding new Helm repo 'myrepo'..."
        helm repo add myrepo https://kweaver-ai.github.io/helm-repo/
        helm repo update
    fi
}

install_etrino() {
    check_install_state
    if [[ "${VEGA_HDFS_INSTALLED}" == true && "${VEGA_CALCULATE_INSTALLED}" == true && "${VEGA_METADATA_INSTALLED}" == true ]]; then
        echo "All Etrino services are already installed. Skipping installation."
        return 0
    fi

    echo "Starting Etrino installation..."
    ensure_node_labels_and_dirs
    ensure_helm_repo

    local values_args=()
    if [[ -f "${CONFIG_FILE}" ]]; then
        echo "Using config file: ${CONFIG_FILE}"
        values_args=(-f "${CONFIG_FILE}")
    else
        echo "WARNING: Config file not found at ${CONFIG_FILE}"
        echo "Installing without custom values file..."
    fi

    echo "Installing Etrino services in ${NAMESPACE} namespace..."

    if [[ "${VEGA_HDFS_INSTALLED}" == false ]]; then
        echo "Installing vega-hdfs..."
        helm install -n "${NAMESPACE}" vega-hdfs myrepo/vega-hdfs --version $VEGA_HDFS_VERSION "${values_args[@]}"
    else
        echo "Skipping vega-hdfs (already installed)"
    fi

    if [[ "${VEGA_CALCULATE_INSTALLED}" == false ]]; then
        echo "Installing vega-calculate..."
        helm install -n "${NAMESPACE}" vega-calculate myrepo/vega-calculate --version $VEGA_CALCULATE_VERSION "${values_args[@]}"
    else
        echo "Skipping vega-calculate (already installed)"
    fi

    if [[ "${VEGA_METADATA_INSTALLED}" == false ]]; then
        echo "Installing vega-metadata..."
        helm install -n "${NAMESPACE}" vega-metadata myrepo/vega-metadata --version $VEGA_METADATA_VERSION "${values_args[@]}"
    else
        echo "Skipping vega-metadata (already installed)"
    fi

    echo "Etrino services installation completed!"
}

show_etrino_status() {
    echo "Etrino services status:"
    echo "Namespace: ${NAMESPACE}"
    echo ""

    local release_name
    for release_name in vega-hdfs vega-calculate vega-metadata; do
        if helm status "${release_name}" -n "${NAMESPACE}" >/dev/null 2>&1; then
            local status
            status=$(helm status "${release_name}" -n "${NAMESPACE}" -o json 2>/dev/null | grep -o '"status":"[^"]*"' | head -1 | cut -d'"' -f4)
            echo "  ✓ ${release_name}: ${status}"
        else
            echo "  ✗ ${release_name}: not installed"
        fi
    done
}

uninstall_etrino() {
    echo "Uninstalling Etrino services from namespace: ${NAMESPACE}..."
    local release_name
    for release_name in vega-metadata vega-calculate vega-hdfs; do
        echo "Uninstalling ${release_name}..."
        if helm uninstall "${release_name}" -n "${NAMESPACE}" >/dev/null 2>&1; then
            echo "✓ ${release_name} uninstalled successfully"
        else
            echo "⚠ ${release_name} not found or already uninstalled"
        fi
    done
    echo "Etrino services uninstallation completed!"
}

main() {
    local action
    action=$(parse_args "$@")
    load_namespace_from_config

    case "${action}" in
        install|init)
            install_etrino
            ;;
        status)
            show_etrino_status
            ;;
        uninstall)
            uninstall_etrino
            ;;
        *)
            echo "Unknown action: ${action}"
            usage
            exit 1
            ;;
    esac
}

main "$@"
