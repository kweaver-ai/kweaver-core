
# Detect package manager (prefer dnf, fallback to yum, then apt)
detect_package_manager() {
    if command -v dnf &> /dev/null; then
        PKG_MANAGER="dnf"
        PKG_MANAGER_UPDATE="dnf makecache"
        PKG_MANAGER_INSTALL="dnf install -y"
        PKG_MANAGER_HOLD="dnf mark install"
    elif command -v yum &> /dev/null; then
        PKG_MANAGER="yum"
        PKG_MANAGER_UPDATE="yum makecache"
        PKG_MANAGER_INSTALL="yum install -y"
        # yum doesn't have a direct hold command; use versionlock plugin if available, otherwise skip
        PKG_MANAGER_HOLD="yum versionlock add 2>/dev/null || true"
    elif command -v apt-get &> /dev/null; then
        PKG_MANAGER="apt"
        PKG_MANAGER_UPDATE="apt-get update -y"
        PKG_MANAGER_INSTALL="apt-get install -y"
        PKG_MANAGER_HOLD="apt-mark hold"
    else
        log_error "No supported package manager found (dnf, yum, or apt-get)"
        exit 1
    fi
    
    export PKG_MANAGER PKG_MANAGER_UPDATE PKG_MANAGER_INSTALL PKG_MANAGER_HOLD
    log_info "Using package manager: ${PKG_MANAGER}"
}

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "This script must be run as root"
        exit 1
    fi
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if kubeadm is installed
    if ! command -v kubeadm &> /dev/null; then
        log_error "kubeadm is not installed. Please install kubeadm first."
        exit 1
    fi
    
    # Check if kubelet is installed
    if ! command -v kubelet &> /dev/null; then
        log_error "kubelet is not installed. Please install kubelet first."
        exit 1
    fi
    
    # Check if kubectl is installed
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed. Please install kubectl first."
        exit 1
    fi
    
    # Check if container runtime is running, try to start if not
    if systemctl is-active --quiet containerd; then
        log_info "containerd is running"
    elif systemctl is-active --quiet docker; then
        log_info "docker is running"
    elif systemctl is-active --quiet crio; then
        log_info "cri-o is running"
    else
        # Try to start containerd if it's installed but not running
        if command -v containerd &> /dev/null; then
            log_info "containerd is installed but not running, attempting to start..."
            systemctl start containerd 2>/dev/null || true
            sleep 2
            if systemctl is-active --quiet containerd; then
                log_info "containerd started successfully"
            else
                log_error "Failed to start containerd"
                exit 1
            fi
        else
            log_error "No container runtime (containerd/docker/cri-o) is installed or running"
            exit 1
        fi
    fi
    
    log_info "Prerequisites check passed"
}

# Initialize Kubernetes master node
init_k8s_master() {
    log_info "Initializing Kubernetes master node..."
    log_info "Configuration: POD_CIDR=${POD_CIDR}, SERVICE_CIDR=${SERVICE_CIDR}"

    # Check if Kubernetes is already initialized before doing any system configuration
    if [[ -f /etc/kubernetes/admin.conf ]]; then
        if KUBECONFIG=/etc/kubernetes/admin.conf kubectl get nodes >/dev/null 2>&1; then
            log_info "Kubernetes already initialized (kubectl get nodes succeeded). Skipping system configuration and kubeadm init."
            mkdir -p /root/.kube
            cp -f /etc/kubernetes/admin.conf /root/.kube/config
            export KUBECONFIG=/root/.kube/config
            return 0
        fi
    fi

    if [[ -f /root/.kube/config ]]; then
        if KUBECONFIG=/root/.kube/config kubectl get nodes >/dev/null 2>&1; then
            log_info "Kubernetes already initialized (kubectl get nodes succeeded). Skipping system configuration and kubeadm init."
            export KUBECONFIG=/root/.kube/config
            return 0
        fi
    fi

    # Configure system for Kubernetes (only if not already initialized)
    log_info "Configuring system for Kubernetes..."
    disable_selinux
    configure_system
    
    # Get the default network interface IP if not specified
    if [[ -z "${API_SERVER_ADVERTISE_ADDRESS}" ]]; then
        API_SERVER_ADVERTISE_ADDRESS=$(hostname -I | awk '{print $1}')
    fi
    
    log_info "API Server advertise address: ${API_SERVER_ADVERTISE_ADDRESS}"
    
    # Pre-pull images before kubeadm init
    log_info "Pre-pulling Kubernetes images from ${IMAGE_REPOSITORY}..."
    kubeadm config images pull \
        --kubernetes-version=stable-1.28 \
        --image-repository="${IMAGE_REPOSITORY}" \
        2>&1 || log_warn "Some images may have failed to pull, continuing..."
    
    # Pre-pull pause image with all possible versions and tag them
    log_info "Pre-pulling pause images with all versions..."
    for pause_version in 3.6 3.9 3.10 3.10.0 3.10.1; do
        log_info "Pulling pause:${pause_version}..."
        crictl pull "${IMAGE_REPOSITORY}/pause:${pause_version}" 2>/dev/null || true
        # Tag the image as registry.k8s.io version for kubeadm
        ctr -n k8s.io image tag "${IMAGE_REPOSITORY}/pause:${pause_version}" "registry.k8s.io/pause:${pause_version}" 2>/dev/null || true
    done
    
    # Create kubeadm config file to specify image repository
    log_info "Creating kubeadm configuration file..."
    mkdir -p /tmp/kubeadm
    cat > /tmp/kubeadm/config.yaml <<EOF
apiVersion: kubeadm.k8s.io/v1beta3
kind: InitConfiguration
nodeRegistration:
  criSocket: unix:///var/run/containerd/containerd.sock
  kubeletExtraArgs:
    pod-infra-container-image: ${IMAGE_REPOSITORY}/pause:3.9
---
apiVersion: kubeadm.k8s.io/v1beta3
kind: ClusterConfiguration
kubernetesVersion: stable-1.28
controlPlaneEndpoint: ${API_SERVER_ADVERTISE_ADDRESS}:6443
networking:
  podSubnet: ${POD_CIDR}
  serviceSubnet: ${SERVICE_CIDR}
imageRepository: ${IMAGE_REPOSITORY}
EOF
    
    # Final CRI check before init
    if ! crictl info &>/dev/null; then
        log_warn "CRI is not responding. Attempting to fix and restart containerd..."
        install_containerd
    fi

    log_info "Initializing the cluster..."
    # Initialize the cluster with config file
    local init_rc=0
    set +e
    kubeadm init \
        --config=/tmp/kubeadm/config.yaml \
        --ignore-preflight-errors=NumCPU,Mem 2>&1 | tee /tmp/kubeadm-init.log
    init_rc=$?
    set -e

    if [[ ${init_rc} -ne 0 ]]; then
        log_error "kubeadm init failed (exit code: ${init_rc}). Check /tmp/kubeadm-init.log for details."
        # Check for common CRI error and provide automated fix hint
        if grep -q "unknown service runtime.v1.RuntimeService" /tmp/kubeadm-init.log; then
            log_warn "Detected CRI v1 API mismatch. This usually means containerd config is missing SystemdCgroup=true or not restarted."
            log_info "Attempting automated fix for containerd..."
            install_containerd
            log_info "Retrying kubeadm init..."
            kubeadm init --config=/tmp/kubeadm/config.yaml --ignore-preflight-errors=NumCPU,Mem 2>&1 | tee /tmp/kubeadm-init.log
        else
            return 1
        fi
    fi
    
    # Fix pause image version in kubelet configuration
    log_info "Fixing pause image version in kubelet configuration..."
    systemctl stop kubelet
    
    # Replace pause image versions with 3.9 in all kubelet config files
    # Replace registry.k8s.io with aliyun registry for all pause versions (including 3.6)
    sed -i 's|registry\.k8s\.io/pause:[0-9.]*|registry.aliyuncs.com/google_containers/pause:3.9|g' /var/lib/kubelet/kubeadm-flags.env 2>/dev/null || true
    sed -i 's|registry\.k8s\.io/pause:[0-9.]*|registry.aliyuncs.com/google_containers/pause:3.9|g' /var/lib/kubelet/config.yaml 2>/dev/null || true
    
    # Ensure pause image is set correctly in kubelet extra args
    if ! grep -q 'pod-infra-container-image' /var/lib/kubelet/kubeadm-flags.env; then
        sed -i 's|--container-runtime-endpoint|--pod-infra-container-image=registry.aliyuncs.com/google_containers/pause:3.9 --container-runtime-endpoint|g' /var/lib/kubelet/kubeadm-flags.env
    fi
    
    systemctl start kubelet
    
    # Wait for control plane to stabilize
    log_info "Waiting for control plane to stabilize..."
    sleep 30
    
    # Setup kubeconfig for root user
    log_info "Setting up kubeconfig..."
    mkdir -p /root/.kube
    cp -f /etc/kubernetes/admin.conf /root/.kube/config
    chown root:root /root/.kube/config
    
    # Fix API server address to use IPv4 (in case IPv6 is disabled)
    log_info "Ensuring kubeconfig uses IPv4 API server address..."
    sed -i "s|https://\[::1\]:6443|https://${API_SERVER_ADVERTISE_ADDRESS}:6443|g" /root/.kube/config
    sed -i "s|https://localhost:6443|https://${API_SERVER_ADVERTISE_ADDRESS}:6443|g" /root/.kube/config
    sed -i "s|https://127.0.0.1:6443|https://${API_SERVER_ADVERTISE_ADDRESS}:6443|g" /root/.kube/config
    
    # Setup kubeconfig for current user if not root
    if [[ -n "${SUDO_USER}" ]]; then
        USER_HOME=$(getent passwd "${SUDO_USER}" | cut -d: -f6)
        mkdir -p "${USER_HOME}/.kube"
        cp -f /root/.kube/config "${USER_HOME}/.kube/config"
        chown -R "${SUDO_USER}:${SUDO_USER}" "${USER_HOME}/.kube"
    fi
    
    export KUBECONFIG=/root/.kube/config
    
    log_info "Kubernetes master node initialized successfully"
}

# Remove taint to allow scheduling on master node
allow_master_scheduling() {
    log_info "Allowing scheduling on master node..."
    
    # Remove the NoSchedule taint from master/control-plane node
    kubectl taint nodes --all node-role.kubernetes.io/control-plane- 2>/dev/null || true
    kubectl taint nodes --all node-role.kubernetes.io/master- 2>/dev/null || true
    
    log_info "Master node is now schedulable"
}

# Install Weave CNI plugin (simpler alternative to Calico)
install_cni() {
    log_info "Installing Flannel CNI plugin..."

    if kubectl get daemonset kube-flannel-ds -n kube-flannel >/dev/null 2>&1; then
        local desired
        local ready
        desired=$(kubectl get daemonset kube-flannel-ds -n kube-flannel -o jsonpath='{.status.desiredNumberScheduled}' 2>/dev/null || echo "")
        ready=$(kubectl get daemonset kube-flannel-ds -n kube-flannel -o jsonpath='{.status.numberReady}' 2>/dev/null || echo "")
        if [[ -n "${desired}" && -n "${ready}" && "${desired}" == "${ready}" && "${ready}" != "0" ]]; then
            log_info "Flannel is already installed and ready (daemonset Ready ${ready}/${desired}), skipping"
            return 0
        fi
    fi
    
    # Install Flannel CNI (ensure network CIDR matches POD_CIDR)
    # Note: Image addresses are already configured in the YAML file
    cat "${FLANNEL_MANIFEST_PATH}" | \
        sed "s|10.244.0.0/16|${POD_CIDR}|g" | \
        kubectl apply -f -
    
    log_info "Waiting for Flannel pods to be ready..."
    sleep 10
    kubectl wait --for=condition=Ready pods --all -n kube-flannel --timeout=300s 2>/dev/null || true
    
    # Restart containerd to ensure CNI plugin is properly initialized
    log_info "Restarting containerd to ensure CNI plugin initialization..."
    systemctl restart containerd
    sleep 5
    
    # Wait for node network to be ready (CNI plugin initialized)
    log_info "Waiting for CNI plugin to initialize network..."
    local max_attempts=30
    local attempt=0
    while [[ ${attempt} -lt ${max_attempts} ]]; do
        if kubectl get nodes -o jsonpath='{.items[0].status.conditions[?(@.type=="Ready")].status}' 2>/dev/null | grep -q "True"; then
            log_info "Node network is ready"
            break
        fi
        attempt=$((attempt + 1))
        log_info "Waiting for node network to be ready... (${attempt}/${max_attempts})"
        sleep 5
    done
    
    # If node is still not ready, try to remove not-ready taint to allow pods to schedule and trigger CNI init
    if ! kubectl get nodes -o jsonpath='{.items[0].status.conditions[?(@.type=="Ready")].status}' 2>/dev/null | grep -q "True"; then
        log_info "Node still not ready, removing not-ready taint to allow pod scheduling..."
        local node_name
        node_name=$(kubectl get nodes -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
        if [[ -n "${node_name}" ]]; then
            kubectl taint nodes "${node_name}" node.kubernetes.io/not-ready:NoSchedule- 2>/dev/null || true
            log_info "Waiting for CNI to initialize after taint removal..."
            sleep 15
        fi
    fi
    
    # Wait for subnet.env file to be created (required by pods for networking)
    log_info "Waiting for Flannel subnet configuration..."
    local subnet_attempts=0
    local subnet_max_attempts=30
    while [[ ${subnet_attempts} -lt ${subnet_max_attempts} ]]; do
        if [[ -f /run/flannel/subnet.env ]]; then
            log_info "Flannel subnet.env created successfully"
            cat /run/flannel/subnet.env
            break
        fi
        subnet_attempts=$((subnet_attempts + 1))
        log_info "Waiting for /run/flannel/subnet.env... (${subnet_attempts}/${subnet_max_attempts})"
        sleep 2
    done
    
    if [[ ! -f /run/flannel/subnet.env ]]; then
        log_warn "Flannel subnet.env not found after waiting, CoreDNS may have issues"
    fi
    
    # Delete any existing CoreDNS pods that might be stuck
    log_info "Deleting existing CoreDNS pods to restart with CNI ready..."
    kubectl -n kube-system delete pod -l k8s-app=kube-dns --force --grace-period=0 2>/dev/null || true
    sleep 5
    
    # Wait for new CoreDNS pods to be ready using kubectl wait
    log_info "Waiting for CoreDNS pods to be ready..."
    local dns_attempts=0
    local dns_max_attempts=60
    while [[ ${dns_attempts} -lt ${dns_max_attempts} ]]; do
        # Count ready pods using simple parsing
        local ready_count
        ready_count=$(kubectl get pods -n kube-system -l k8s-app=kube-dns --no-headers 2>/dev/null | grep -c "1/1.*Running" || echo "0")
        
        if [[ ${ready_count} -ge 2 ]]; then
            log_info "CoreDNS is ready (${ready_count} pods running)"
            break
        fi
        dns_attempts=$((dns_attempts + 1))
        log_info "Waiting for CoreDNS pods to be ready... (${dns_attempts}/${dns_max_attempts}, ${ready_count}/2 ready)"
        sleep 5
    done
    
    if [[ ${dns_attempts} -ge ${dns_max_attempts} ]]; then
        log_warn "CoreDNS may not be fully ready, but continuing..."
    fi
    
    log_info "Flannel CNI plugin installed successfully"
}

# Wait for CoreDNS to be ready (it's installed by kubeadm by default)
wait_for_dns() {
    log_info "Waiting for CoreDNS to be ready..."
    
    kubectl wait --for=condition=Ready pods -l k8s-app=kube-dns -n kube-system --timeout=300s
    
    log_info "CoreDNS is ready"
}

# Install Helm 3
install_helm() {
    log_info "Installing Helm 3..."

    local desired="${HELM_VERSION}"
    local existing=""
    if command -v helm &> /dev/null; then
        existing="$(helm version --short 2>/dev/null | awk '{print $1}' | cut -d'+' -f1 || true)"
        if [[ -n "${existing}" && "${existing}" == "${desired}" ]]; then
            log_info "Helm ${desired} is already installed"
            return 0
        fi
        if [[ -n "${existing}" ]]; then
            log_warn "Helm version ${existing} detected; installing desired ${desired}"
        fi
    fi

    # Prefer HuaweiCloud tarball (pinned by version + arch); fallback to get-helm-3 script if needed.
    local arch=""
    case "$(uname -m)" in
        x86_64|amd64)
            arch="amd64"
            ;;
        aarch64|arm64)
            arch="arm64"
            ;;
        *)
            log_error "Unsupported architecture for Helm: $(uname -m)"
            return 1
            ;;
    esac

    local base="${HELM_TARBALL_BASEURL%/}/"
    local tarball="helm-${desired}-linux-${arch}.tar.gz"
    local url="${base}${tarball}"

    log_info "Downloading Helm ${desired} from ${url}..."
    local tmpdir
    tmpdir="$(mktemp -d /tmp/helm.XXXXXX)"
    if curl -fsSLo "${tmpdir}/${tarball}" "${url}"; then
        tar -xzf "${tmpdir}/${tarball}" -C "${tmpdir}"
        install -m 0755 "${tmpdir}/linux-${arch}/helm" /usr/local/bin/helm
        rm -rf "${tmpdir}" 2>/dev/null || true
        log_info "Helm ${desired} installed successfully"
        return 0
    fi
    rm -rf "${tmpdir}" 2>/dev/null || true

    log_warn "Failed to download Helm tarball from HuaweiCloud; falling back to get-helm-3 script..."
    if [[ -f "${HELM_INSTALL_SCRIPT_PATH}" ]]; then
        bash "${HELM_INSTALL_SCRIPT_PATH}"
    else
        curl -fsSL "${HELM_INSTALL_SCRIPT_URL}" | bash
    fi

    # Do not auto-add Helm repos here: modules add repos only when a local chart is not available.
    log_info "Helm 3 installed successfully"
}

# Install containerd container runtime
install_containerd() {
    log_info "Checking containerd installation..."

    detect_package_manager
    
    # Step 1: Check if containerd binary exists (indicates it's already installed)
    if command -v containerd &> /dev/null; then
        log_info "containerd binary found, skipping package installation"
        # Check and configure containerd
        configure_containerd_runtime
        
        # Try to start if not running
        if ! systemctl is-active --quiet containerd 2>/dev/null; then
            log_info "Starting containerd service..."
            systemctl daemon-reload
            systemctl enable containerd
            systemctl start containerd
            sleep 2
        fi
        
        if systemctl is-active --quiet containerd 2>/dev/null; then
            log_info "containerd is running and configured"
            return 0
        else
            log_warn "Failed to start containerd, but binary exists. Continuing..."
            return 0
        fi
    fi
    
    # Step 2: containerd is not installed, proceed with installation
    log_info "containerd is not installed, proceeding with installation..."
    
    # Function to configure Docker repo (Tsinghua mirror)
    configure_docker_repo() {
        local url="$1"
        curl -fsSLo /etc/yum.repos.d/docker-ce.repo "${url}"
        
        # Replace official Docker download URLs with Tsinghua mirror
        log_info "Replacing Docker official URLs with Tsinghua mirror..."
        sed -i 's+https://download.docker.com+https://mirrors.tuna.tsinghua.edu.cn/docker-ce+g' /etc/yum.repos.d/docker-ce.repo
        
        # Fix for openEuler: replace $releasever with 9 in repo file
        if [[ -f /etc/os-release ]]; then
            source /etc/os-release
            if [[ "${ID}" == "openEuler" ]] || [[ "${ID}" == "openeuler" ]]; then
                log_info "Detected openEuler system, fixing Docker CE repo paths..."
                sed -i 's|\$releasever|9|g' /etc/yum.repos.d/docker-ce.repo
            fi
        fi
        
        # Clean and makecache for the new repo
        ${PKG_MANAGER} clean all
        rm -rf /var/cache/dnf /var/cache/yum
    }
    
    if [[ "${PKG_MANAGER}" == "dnf" ]] || [[ "${PKG_MANAGER}" == "yum" ]]; then
        # For RHEL/CentOS/Fedora systems
        
        # Detect OS version for CentOS 7 special handling
        local os_id=""
        local os_version_id=""
        if [[ -f /etc/os-release ]]; then
            source /etc/os-release
            os_id="${ID}"
            os_version_id="${VERSION_ID%%.*}"  # Get major version only
        fi
        
        # For CentOS 7: configure Aliyun base repo (official CentOS repos are EOL)
        if [[ "${os_id}" == "centos" ]] && [[ "${os_version_id}" == "7" ]]; then
            if [[ ! -f /etc/yum.repos.d/CentOS-Base.repo ]] || ! grep -q "mirrors.aliyun.com" /etc/yum.repos.d/CentOS-Base.repo 2>/dev/null; then
                log_info "Detected CentOS 7 (EOL), configuring Aliyun base repo..."
                curl -fsSLo /etc/yum.repos.d/CentOS-Base.repo https://mirrors.aliyun.com/repo/Centos-7.repo
                ${PKG_MANAGER} clean all
                rm -rf /var/cache/yum
            fi
            
            # Install prerequisite dependencies from base repo
            log_info "Installing prerequisite dependencies (tar, libseccomp, container-selinux)..."
            set +e
            ${PKG_MANAGER_INSTALL} tar libseccomp container-selinux
            local dep_rc=$?
            set -e
            if [[ ${dep_rc} -ne 0 ]]; then
                log_warn "Some dependencies may not have installed correctly, continuing anyway..."
            fi
        fi
        
        # Configure Docker CE repo if not exists (Aliyun mirror only)
        if [[ ! -f /etc/yum.repos.d/docker-ce.repo ]]; then
            log_info "Configuring Docker CE yum repo: ${DOCKER_CE_REPO_URL}"
            configure_docker_repo "${DOCKER_CE_REPO_URL}"
            
            set +e
            ${PKG_MANAGER_UPDATE}
            local update_rc=$?
            set -e

            if [[ ${update_rc} -ne 0 ]]; then
                log_error "Failed to update package metadata with Docker CE repo."
                log_error "Please ensure network connectivity and try again, or manually install containerd.io package."
                log_info "You can manually install containerd using one of these methods:"
                log_info "  1. dnf install -y containerd.io (after configuring Docker repo)"
                log_info "  2. Download and install RPM: https://mirrors.tuna.tsinghua.edu.cn/docker-ce/linux/centos/"
                return 1
            fi
        fi

        # Attempt installation with both base and docker-ce repos for CentOS 7
        log_info "Attempting to install containerd.io..."
        local docker_repo_name="docker-ce-stable"
        
        set +e
        if [[ "${os_id}" == "centos" ]] && [[ "${os_version_id}" == "7" ]]; then
            # CentOS 7: use both base and docker-ce repos for dependency resolution
            ${PKG_MANAGER_INSTALL} --enablerepo="base,extras,${docker_repo_name}" --nogpgcheck containerd.io
        else
            # Other distros: only use docker-ce repo
            ${PKG_MANAGER_INSTALL} --disablerepo="*" --enablerepo="${docker_repo_name}" --nogpgcheck containerd.io
        fi
        local install_rc=$?
        set -e
        
        if [[ ${install_rc} -ne 0 ]]; then
            log_warn "Failed to install containerd.io from package repo, attempting to download and install RPM directly..."
            
            # Detect OS type and version
            local os_id=""
            local rhel_version=""
            if [[ -f /etc/os-release ]]; then
                source /etc/os-release
                os_id="${ID}"
                rhel_version="${VERSION_ID%%.*}"
            fi
            
            # Default to 8 if version detection fails
            if [[ -z "${rhel_version}" ]]; then
                rhel_version="8"
            fi
            
            local arch=$(uname -m)
            local rpm_url
            
            # Construct correct URL based on OS type
            if [[ "${os_id}" == "rhel" ]] || [[ "${os_id}" == "rocky" ]] || [[ "${os_id}" == "almalinux" ]]; then
                # RHEL-based systems use /rhel/ path
                rpm_url="https://mirrors.tuna.tsinghua.edu.cn/docker-ce/linux/rhel/${rhel_version}/${arch}/stable/Packages/containerd.io-1.6.33-3.1.el${rhel_version}.${arch}.rpm"
            else
                # CentOS uses /centos/ path
                rpm_url="https://mirrors.tuna.tsinghua.edu.cn/docker-ce/linux/centos/${rhel_version}/${arch}/stable/Packages/containerd.io-1.6.33-3.1.el${rhel_version}.${arch}.rpm"
            fi
            
            local rpm_file="/tmp/containerd.io.rpm"
            
            log_info "Downloading containerd.io v1.6.33 RPM from Tsinghua mirror..."
            log_info "URL: ${rpm_url}"
            
            set +e
            if curl -fsSLo "${rpm_file}" "${rpm_url}"; then
                log_info "Downloaded RPM successfully, installing..."
                ${PKG_MANAGER_INSTALL} "${rpm_file}"
                install_rc=$?
                rm -f "${rpm_file}"
                
                if [[ ${install_rc} -eq 0 ]]; then
                    log_info "✓ containerd.io installed successfully from RPM"
                else
                    log_error "Failed to install downloaded RPM"
                fi
            else
                log_error "Failed to download containerd.io RPM"
                install_rc=1
            fi
            set -e
        fi
        
        if [[ ${install_rc} -ne 0 ]]; then
            log_error "=========================================="
            log_error "Failed to install containerd.io package."
            log_error "=========================================="
            log_error ""
            log_error "Please install containerd manually using one of these methods:"
            log_error ""
            log_info "  Option 1: Download and install RPM directly"
            log_info "    curl -fsSLo /tmp/containerd.io.rpm https://mirrors.tuna.tsinghua.edu.cn/docker-ce/linux/centos/\$(rpm -E %rhel)/\$(uname -m)/stable/Packages/containerd.io-1.6.33-3.1.el\$(rpm -E %rhel).\$(uname -m).rpm"
            log_info "    dnf install -y /tmp/containerd.io.rpm"
            log_error ""
            log_info "  Option 2: Install from Aliyun mirror"
            log_info "    dnf config-manager --add-repo http://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo"
            log_info "    dnf install -y containerd.io"
            log_error ""
            log_error "After installing containerd, run this script again."
            log_error "=========================================="
            return 1
        fi
    else
        # For Ubuntu/Debian systems
        if [[ ! -f /etc/apt/sources.list.d/docker.list ]]; then
            curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg 2>/dev/null || true
            echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | \
                tee /etc/apt/sources.list.d/docker.list > /dev/null
            ${PKG_MANAGER_UPDATE}
        fi

        set +e
        ${PKG_MANAGER_INSTALL} containerd.io
        local install_rc=$?
        set -e
        
        if [[ ${install_rc} -ne 0 ]]; then
            log_error "=========================================="
            log_error "Failed to install containerd.io package."
            log_error "=========================================="
            log_error ""
            log_error "Please install containerd manually:"
            log_info "  apt-get update && apt-get install -y containerd.io"
            log_error ""
            log_error "After installing containerd, run this script again."
            log_error "=========================================="
            return 1
        fi
    fi
    
    # Configure containerd after installation
    configure_containerd_runtime
    
    log_info "containerd installed and configured successfully"
}

# Configure containerd runtime (extracted for reuse)
configure_containerd_runtime() {
    
    # Configure containerd
    mkdir -p /etc/containerd
    
    # Check if config file exists and is initialized
    if [[ ! -f /etc/containerd/config.toml ]]; then
        log_info "Creating containerd configuration file..."
        containerd config default > /etc/containerd/config.toml
        if [[ ! -s /etc/containerd/config.toml ]]; then
            log_error "Failed to generate containerd config file"
            return 1
        fi
        log_info "✓ Configuration file created and initialized"
    else
        log_info "Configuration file exists, checking initialization..."
        if [[ ! -s /etc/containerd/config.toml ]]; then
            log_warn "Configuration file is empty, reinitializing..."
            containerd config default > /etc/containerd/config.toml
        elif ! grep -q "SystemdCgroup" /etc/containerd/config.toml; then
            log_warn "Configuration file appears incomplete (missing runtime settings), reinitializing..."
            containerd config default > /etc/containerd/config.toml
        else
            log_info "✓ Configuration file is initialized"
        fi
    fi
    
    # Ensure sandbox_image uses the configured image repository (avoid pulling from registry.k8s.io)
    local desired_sandbox_image="${IMAGE_REPOSITORY}/pause:3.9"
    if grep -q 'sandbox_image' /etc/containerd/config.toml; then
        local current_sandbox_image
        current_sandbox_image=$(grep 'sandbox_image' /etc/containerd/config.toml | head -1 | sed 's/.*= *"//;s/".*//')
        if [[ "${current_sandbox_image}" != "${desired_sandbox_image}" ]]; then
            log_info "Updating sandbox_image from ${current_sandbox_image} to ${desired_sandbox_image}..."
            sed -i "s|sandbox_image = \".*\"|sandbox_image = \"${desired_sandbox_image}\"|g" /etc/containerd/config.toml
        else
            log_info "✓ sandbox_image is correctly configured"
        fi
    else
        log_warn "sandbox_image not found in config, this may cause issues with default containerd sandbox image"
    fi
    
    # Ensure CRI plugin is enabled
    if grep -q "disabled_plugins.*cri" /etc/containerd/config.toml; then
        log_info "Enabling CRI plugin in containerd..."
        sed -i 's/disabled_plugins.*=.*\[.*"cri".*\]/disabled_plugins = []/g' /etc/containerd/config.toml
    else
        log_info "✓ CRI plugin is enabled"
    fi
    
    # Ensure systemd cgroup driver is enabled
    if grep -q "SystemdCgroup = false" /etc/containerd/config.toml; then
        log_info "Enabling systemd cgroup driver..."
        sed -i 's/SystemdCgroup = false/SystemdCgroup = true/g' /etc/containerd/config.toml
    else
        log_info "✓ Systemd cgroup driver is enabled"
    fi
    
    # Restart containerd to apply configuration
    log_info "Restarting containerd to apply configuration..."
    systemctl daemon-reload
    systemctl enable containerd
    systemctl restart containerd
    
    # Wait for containerd to be ready
    sleep 2
    
    # Verify CRI connection
    if command -v crictl &> /dev/null; then
        log_info "Verifying CRI connection..."
        if ! crictl info &> /dev/null; then
            log_warn "crictl failed to connect to containerd. Ensuring /etc/crictl.yaml is correct..."
            cat > /etc/crictl.yaml <<EOF
runtime-endpoint: unix:///run/containerd/containerd.sock
image-endpoint: unix:///run/containerd/containerd.sock
timeout: 10
debug: false
EOF
            # Final check
            if ! crictl info &> /dev/null; then
                log_error "CRI runtime is still not responsive."
                return 1
            fi
        fi
        log_info "✓ CRI connection verified"
    fi
    
    log_info "✓ containerd runtime configured successfully"
}

# Install crictl (container runtime interface CLI)
install_crictl() {
    log_info "Installing crictl..."
    
    if command -v crictl &> /dev/null; then
        log_info "crictl is already installed"
        # Still ensure config file exists
        if [[ ! -f /etc/crictl.yaml ]]; then
            log_info "Creating crictl configuration file..."
            cat > /etc/crictl.yaml <<EOF
runtime-endpoint: unix:///run/containerd/containerd.sock
image-endpoint: unix:///run/containerd/containerd.sock
timeout: 10
debug: false
EOF
        fi
        return 0
    fi

    detect_package_manager

    if [[ "${PKG_MANAGER}" == "dnf" ]] || [[ "${PKG_MANAGER}" == "yum" ]]; then
        log_info "Attempting to install cri-tools (crictl) from ${PKG_MANAGER} repo..."
        if ${PKG_MANAGER_INSTALL} cri-tools; then
            log_info "cri-tools installed successfully"
        else
            log_warn "Failed to install cri-tools from ${PKG_MANAGER} repo; falling back to GitHub release tarball"
        fi
    elif [[ "${PKG_MANAGER}" == "apt" ]]; then
        log_info "Attempting to install cri-tools (crictl) from apt repo..."
        if ${PKG_MANAGER_INSTALL} cri-tools; then
            log_info "cri-tools installed successfully"
        else
            log_warn "Failed to install cri-tools from apt repo; falling back to GitHub release tarball"
        fi
    fi

    if command -v crictl &> /dev/null; then
        # Create crictl configuration
        log_info "Creating crictl configuration file..."
        cat > /etc/crictl.yaml <<EOF
runtime-endpoint: unix:///run/containerd/containerd.sock
image-endpoint: unix:///run/containerd/containerd.sock
timeout: 10
debug: false
EOF
        return 0
    fi
    
    # Download and install crictl
    CRICTL_VERSION="v1.28.0"
    ARCH="amd64"
    
    log_info "Downloading crictl ${CRICTL_VERSION}..."
    curl -L https://github.com/kubernetes-sigs/cri-tools/releases/download/${CRICTL_VERSION}/crictl-${CRICTL_VERSION}-linux-${ARCH}.tar.gz | tar -C /usr/local/bin -xz
    
    # Create crictl configuration
    log_info "Creating crictl configuration file..."
    cat > /etc/crictl.yaml <<EOF
runtime-endpoint: unix:///run/containerd/containerd.sock
image-endpoint: unix:///run/containerd/containerd.sock
timeout: 10
debug: false
EOF
    
    log_info "crictl installed successfully"
}

# Install Kubernetes components (kubeadm, kubelet, kubectl)
install_kubernetes() {
    log_info "Installing Kubernetes components..."
    
    detect_package_manager

    if [[ "${PKG_MANAGER}" == "dnf" ]] || [[ "${PKG_MANAGER}" == "yum" ]]; then
        # Check if Kubernetes repo already exists
        if [[ ! -f /etc/yum.repos.d/kubernetes.repo ]]; then
            log_info "Configuring Kubernetes yum repo..."
            cat > /etc/yum.repos.d/kubernetes.repo <<EOF
[kubernetes]
name=Kubernetes (Aliyun mirror)
baseurl=${K8S_RPM_REPO_BASEURL}
enabled=1
gpgcheck=1
gpgkey=${K8S_RPM_REPO_GPGKEY}
EOF
            ${PKG_MANAGER_UPDATE}
        fi
    fi
    
    if ! command -v kubeadm &> /dev/null || ! command -v kubelet &> /dev/null || ! command -v kubectl &> /dev/null; then
        if [[ "${PKG_MANAGER}" == "dnf" ]] || [[ "${PKG_MANAGER}" == "yum" ]]; then
            # For RHEL/CentOS/Fedora systems
            ${PKG_MANAGER_INSTALL} kubeadm kubelet kubectl kubernetes-cni
            # Only use hold command if it's available (dnf mark install works, yum versionlock may need plugin)
            if [[ "${PKG_MANAGER}" == "dnf" ]]; then
                ${PKG_MANAGER_HOLD} kubeadm kubelet kubectl kubernetes-cni 2>/dev/null || true
            fi
        else
            # For Ubuntu/Debian systems
            curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg https://packages.cloud.google.com/apt/doc/apt-key.gpg
            echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | \
                tee /etc/apt/sources.list.d/kubernetes.list
            
            ${PKG_MANAGER_UPDATE}
            ${PKG_MANAGER_INSTALL} kubeadm kubelet kubectl
            ${PKG_MANAGER_HOLD} kubeadm kubelet kubectl
        fi
        
        log_info "Kubernetes components installed"
    else
        log_info "Kubernetes components are already installed"
    fi
    
    # Install crictl (always install, even if K8s components already exist)
    install_crictl
    
    # Enable kubelet service
    systemctl daemon-reload
    systemctl enable kubelet
}

# Disable SELinux
disable_selinux() {
    log_info "Disabling SELinux..."
    
    if command -v getenforce &> /dev/null; then
        SELINUX_STATUS=$(getenforce)
        if [[ "${SELINUX_STATUS}" != "Disabled" ]]; then
            log_info "Current SELinux status: ${SELINUX_STATUS}"
            
            # Disable SELinux immediately
            setenforce 0 2>/dev/null || log_warn "Failed to disable SELinux immediately"
            
            # Disable SELinux permanently
            sed -i 's/^SELINUX=.*/SELINUX=disabled/' /etc/selinux/config 2>/dev/null || true
            
            log_info "SELinux has been disabled (reboot required for permanent effect)"
        else
            log_info "SELinux is already disabled"
        fi
    else
        log_info "SELinux is not installed on this system"
    fi
}

# Configure system for Kubernetes
configure_system() {
    log_info "Configuring system for Kubernetes..."
    
    # Disable swap
    log_info "Disabling swap..."
    swapoff -a 2>/dev/null || true
    sed -i '/ swap / s/^/#/' /etc/fstab 2>/dev/null || true
    
    # Load required kernel modules
    log_info "Loading kernel modules..."
    modprobe overlay 2>/dev/null || true
    modprobe br_netfilter 2>/dev/null || true

    # Ensure required kernel modules load on boot
    log_info "Configuring kernel modules to load on boot..."
    mkdir -p /etc/modules-load.d 2>/dev/null || true
    cat > /etc/modules-load.d/kubernetes.conf <<EOF
br_netfilter
EOF
    
    # Configure kernel parameters
    log_info "Configuring kernel parameters..."
    cat > /etc/sysctl.d/99-kubernetes.conf <<EOF
net.bridge.bridge-nf-call-iptables = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward = 1
EOF
    
    # Apply sysctl settings with timeout to avoid hanging
    timeout 10 sysctl --system 2>/dev/null || log_warn "sysctl configuration may have timed out"
    
    # Ensure ip_forward is enabled immediately (in case sysctl didn't apply it)
    echo 1 > /proc/sys/net/ipv4/ip_forward 2>/dev/null || true
    
    log_info "System configured for Kubernetes"
}

# Pre-install all dependencies
preinstall_all() {
    log_info "Starting pre-installation of all dependencies..."
    
    check_root
    detect_package_manager
    install_containerd
    install_kubernetes
    install_helm
    
    log_info "Pre-installation completed successfully"
}


reset_k8s() {
    log_info "Resetting Kubernetes cluster state..."
    
    check_root
    
    # Confirmation prompt
    echo ""
    echo "WARNING: This will reset Kubernetes and clean up CNI/kubeconfig files."
    echo "This action cannot be undone."
    read -p "Type 'Y' or 'y' to confirm: " -r confirm
    
    if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
        log_info "Reset cancelled by user"
        return 0
    fi
    
    systemctl stop kubelet 2>/dev/null || true
    kubeadm reset -f 2>/dev/null || true
    
    rm -rf /etc/cni/net.d 2>/dev/null || true
    rm -rf /var/lib/cni 2>/dev/null || true
    rm -rf /root/.kube 2>/dev/null || true
    rm -f /etc/kubernetes/admin.conf 2>/dev/null || true
    
    log_warn "Reset completed. iptables/IPVS rules are not automatically cleaned by this script."
    log_info "Kubernetes reset done"
}
