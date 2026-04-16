#!/bin/bash

# Etrino Deployment Script
# Automatically add node labels, create directories and install Etrino services

# Read namespace from config file or use default
NAMESPACE="${NAMESPACE:-kweaver}"
CONFIG_FILE="${CONFIG_FILE:-$HOME/.kweaver-ai/config.yaml}"

VEGA_METADATA_VERSION="3.3.1-release"
VEGA_CALCULATE_VERSION="3.3.4-release"
VEGA_HDFS_VERSION="3.1.0-release"

# Try to read namespace from config file if it exists
if [[ -f "${CONFIG_FILE}" ]]; then
    CONFIG_NAMESPACE=$(grep "^namespace:" "${CONFIG_FILE}" 2>/dev/null | head -1 | awk '{print $2}' | tr -d "'\"")
    NAMESPACE="${CONFIG_NAMESPACE:-${NAMESPACE}}"
fi

# Check if helm release exists
check_helm_release() {
    local release_name=$1
    helm list -n "$NAMESPACE" 2>/dev/null | grep -q "^${release_name}"
}

# Check existing installations
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

# If all installed, skip
if [ "$VEGA_HDFS_INSTALLED" = true ] && [ "$VEGA_CALCULATE_INSTALLED" = true ] && [ "$VEGA_METADATA_INSTALLED" = true ]; then
    echo "All Etrino services are already installed. Skipping installation."
    exit 0
fi

echo "Starting Etrino installation..."

# 1. Dynamically get node names from Kubernetes cluster
echo "Getting node names from Kubernetes cluster..."
NODES=($(kubectl get nodes -o jsonpath='{.items[*].metadata.name}'))

# 2. Add labels to nodes (skip if already labeled)
echo "Adding labels to nodes..."
for i in "${!NODES[@]}"; do
    node="${NODES[$i]}"
    label_name="node${i}"
    if kubectl get node "${node}" -o jsonpath='{.metadata.labels}' | grep -q "aishu.io/hostname"; then
        echo "  Node ${node} already has aishu.io/hostname label, skipping"
    else
        echo "  Labeling node ${node} with: aishu.io/hostname=${label_name}"
        kubectl label nodes "${node}" aishu.io/hostname="${label_name}" --overwrite
    fi
done

echo "Node labeling completed!"

# 3. Create directories on nodes
echo "Creating required directories on nodes..."
REQUIRED_DIRS="/sysvol/journalnode/mycluster /sysvol/namenode /sysvol/datanode /sysvol/namenode-slaves"

for node in "${NODES[@]}"; do
    echo "Checking directories on node ${node}..."
    if [ "${#NODES[@]}" -eq 1 ]; then
        echo "  Single-node cluster detected, creating directories locally..."
        mkdir -p $REQUIRED_DIRS
        echo "  Directories created locally"
    else
        echo "  Attempting SSH to ${node}..."
        if ssh -o ConnectTimeout=5 -o StrictHostKeyChecking=no "${node}" "mkdir -p $REQUIRED_DIRS" 2>/dev/null; then
            echo "  Directories created via SSH on ${node}"
        else
            echo "  WARNING: Cannot SSH to ${node}. Please manually create directories: $REQUIRED_DIRS"
        fi
    fi
done

echo "Directory creation completed!"

# 4. Add Helm repository
echo "Adding Helm repository..."
if helm repo list 2>/dev/null | grep -q "^myrepo"; then
    echo "Helm repo 'myrepo' already exists, updating..."
    helm repo update myrepo
else
    echo "Adding new Helm repo 'myrepo'..."
    helm repo add myrepo https://kweaver-ai.github.io/helm-repo/
    helm repo update
fi

# 5. Check config file
if [ ! -f "$CONFIG_FILE" ]; then
    echo "WARNING: Config file not found at $CONFIG_FILE"
    echo "Installing without custom values file..."
    VALUES_FLAG=""
else
    echo "Using config file: $CONFIG_FILE"
    VALUES_FLAG="-f $CONFIG_FILE"
fi

# 6. Install services in kweaver namespace
echo "Installing Etrino services in ${NAMESPACE} namespace..."

if [ "$VEGA_HDFS_INSTALLED" = false ]; then
    echo "Installing vega-hdfs..."
    helm install -n "$NAMESPACE" vega-hdfs myrepo/vega-hdfs --version $VEGA_HDFS_VERSION $VALUES_FLAG
else
    echo "Skipping vega-hdfs (already installed)"
fi

if [ "$VEGA_CALCULATE_INSTALLED" = false ]; then
    echo "Installing vega-calculate..."
    helm install -n "$NAMESPACE" vega-calculate myrepo/vega-calculate --version $VEGA_CALCULATE_VERSION $VALUES_FLAG
else
    echo "Skipping vega-calculate (already installed)"
fi

if [ "$VEGA_METADATA_INSTALLED" = false ]; then
    echo "Installing vega-metadata..."
    helm install -n "$NAMESPACE" vega-metadata myrepo/vega-metadata --version $VEGA_METADATA_VERSION $VALUES_FLAG
else
    echo "Skipping vega-metadata (already installed)"
fi

echo "Etrino services installation completed!"
