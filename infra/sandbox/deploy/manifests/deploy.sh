#!/bin/bash
# Kubernetes 本地部署脚本
# 用于在本地 Kind/Minikube/K3s/Docker Desktop 环境中快速部署和测试
#
# 使用 s3fs + bind mount 方式挂载 S3 workspace（适合单节点开发环境）

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 函数：打印信息
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# 检查 kubectl 是否安装
check_kubectl() {
    if ! command -v kubectl &> /dev/null; then
        error "kubectl not found. Please install kubectl first."
        exit 1
    fi
    info "kubectl found: $(kubectl version --client --short 2>/dev/null || echo 'unknown')"
}

# 检查 Docker 是否安装
check_docker() {
    if ! command -v docker &> /dev/null; then
        error "Docker not found. Please install Docker first."
        exit 1
    fi
    info "Docker found: $(docker --version)"
}

# 检测 K8s 环境
detect_k8s_env() {
    if kubectl cluster-info &> /dev/null; then
        info "Kubernetes cluster is accessible"
        kubectl cluster-info
    else
        error "Cannot access Kubernetes cluster. Please start your K8s environment first."
        info ""
        info "For Docker Desktop:"
        info "  Start Docker Desktop, enable Kubernetes in settings"
        info ""
        info "For Kind:"
        info "  kind create cluster --name sandbox"
        info ""
        info "For Minikube:"
        info "  minikube start --cpus=4 --memory=8192"
        info ""
        info "For K3s:"
        info "  curl -sfL https://get.k3s.io | sh"
        info "  export KUBECONFIG=/etc/rancher/k3s/k3s.yaml"
        exit 1
    fi
}

# 构建 Docker 镜像
build_image() {
    step "Building Docker image..."
    docker build -f ../../sandbox_control_plane/Dockerfile -t sandbox-control-plane:latest ../../.
    info "Docker image built successfully"
}

# 加载镜像到 K8s
load_image() {
    local cluster_type=$1

    step "Loading image to Kubernetes..."

    case $cluster_type in
        kind)
            if kind load docker-image sandbox-control-plane:latest 2>/dev/null; then
                info "Image loaded to Kind"
            else
                error "Failed to load image to Kind"
                exit 1
            fi
            ;;
        minikube)
            if minikube image load sandbox-control-plane:latest 2>/dev/null; then
                info "Image loaded to Minikube"
            else
                error "Failed to load image to Minikube"
                exit 1
            fi
            ;;
        docker-desktop|orbstack)
            info "$cluster_type shares images automatically, no need to load"
            ;;
        k3s)
            info "K3s uses Docker images directly, no need to load"
            ;;
        *)
            warn "Unknown cluster type: $cluster_type, skipping image load"
            ;;
    esac
}

# 部署 K8s 资源（按依赖顺序）
deploy_resources() {
    step "Deploying Kubernetes resources in dependency order..."

    info ""
    info "=== Step 1: 基础资源 ==="
    kubectl apply -f 00-namespace.yaml
    kubectl apply -f 01-configmap.yaml
    kubectl apply -f 02-secret.yaml
    kubectl apply -f 03-serviceaccount.yaml
    kubectl apply -f 04-role.yaml
    info "基础资源部署完成"

    info ""
    info "=== Step 2: 存储层 ==="
    kubectl apply -f 08-mariadb-deployment.yaml
    kubectl apply -f 07-minio-deployment.yaml

    info "Waiting for MariaDB to be ready..."
    kubectl wait --for=condition=ready pod -l app=mariadb -n sandbox-system --timeout=120s || {
        warn "MariaDB wait timeout, continuing anyway..."
    }

    info "Waiting for MinIO to be ready..."
    kubectl wait --for=condition=ready pod -l app=minio -n sandbox-system --timeout=120s || {
        warn "MinIO wait timeout, continuing anyway..."
    }

    info "存储层部署完成"

    info ""
    info "=== Step 3: Control Plane ==="
    kubectl apply -f 05-control-plane-deployment.yaml

    info "Control Plane 部署完成"
}

# 等待 Control Plane 就绪
wait_for_control_plane() {
    step "Waiting for Control Plane to be ready..."
    kubectl wait --for=condition=available deployment/sandbox-control-plane -n sandbox-system --timeout=300s

    info "Control Plane is ready!"
}

# 显示部署状态
show_status() {
    info ""
    info "=== Deployment Status ==="
    info ""
    info "Pods in sandbox-system:"
    kubectl get pods -n sandbox-system
    info ""
    info "Services:"
    kubectl get svc -n sandbox-system
    info ""
    info "Deployments:"
    kubectl get deploy -n sandbox-system
    info ""
    info "DaemonSets:"
    kubectl get ds -n sandbox-system
    info ""
}

# 验证部署
verify_deployment() {
    step "Verifying deployment..."

    info ""
    info "=== 验证数据库连接 ==="
    kubectl exec -n sandbox-system deployment/mariadb -- mariadb -u root -p"password" -e "SHOW DATABASES;"

    info ""
    info "=== 验证 Control Plane ==="
    kubectl exec -n sandbox-system deployment/sandbox-control-plane -- curl -s http://localhost:8000/api/v1/health || echo "Health check not accessible"
}

# 设置端口转发
setup_port_forward() {
    info ""
    info "=== Port Forwarding ==="
    info "Run the following command to forward Control Plane port:"
    info ""
    echo -e "${YELLOW}kubectl port-forward svc/sandbox-control-plane 8000:8000 -n sandbox-system${NC}"
    info ""
    info "Then access the API at:"
    echo -e "${GREEN}  - API: http://localhost:8000${NC}"
    echo -e "${GREEN}  - Docs: http://localhost:8000/docs${NC}"
    echo -e "${GREEN}  - Health: http://localhost:8000/api/v1/health${NC}"
    info ""
}

# 显示资源说明
show_resource_info() {
    cat << 'EOF'

=== Kubernetes 资源说明 ===

📦 基础资源 (00-04)
  00-namespace.yaml         - 创建 sandbox-system 和 sandbox-runtime 命名空间
  01-configmap.yaml        - 应用配置（环境变量、超时设置等）
  02-secret.yaml           - 敏感信息（S3 凭证、数据库密码）
  03-serviceaccount.yaml   - ServiceAccount（Pod 访问 K8s API 的身份）
  04-role.yaml             - RBAC 权限（Pod 操作权限）

💾 存储层 (07-08)
  07-minio-deployment.yaml  - MinIO 对象存储（S3 兼容，存储 workspace 文件）
  08-mariadb-deployment.yaml - MariaDB 数据库（存储会话、执行记录）

🎮 Control Plane (05)
  05-control-plane-deployment.yaml - Control Plane 服务
                                   • REST API（会话管理、执行调度）
                                   • 使用 s3fs 挂载 S3 workspace

=== 架构说明 ===

┌─────────────────────────────────────────────────────────────┐
│                     Kubernetes Cluster                      │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │       Executor Pod (emptyDir + s3fs)                  │  │
│  │  • s3fs 挂载 S3 bucket 到 /mnt/s3-root                 │  │
│  │  • mount --bind overlay session 目录到 /workspace      │  │
│  └──────────────────────┬───────────────────────────────┘  │
│                         │                                    │
│  ┌──────────────────────▼───────────────────────────────┐  │
│  │            Storage Layer                                 │  │
│  │  ┌──────────────┐  ┌──────────────┐                    │  │
│  │  │ MariaDB      │  │ MinIO        │                    │  │
│  │  │ (会话数据)   │  │ (文件存储)   │                    │  │
│  │  └──────────────┘  └──────────────┘                    │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘

EOF
}

# 主函数
main() {
    local cluster_type=${1:-auto}
    local skip_build=${2:-false}
    local show_info=${3:-false}

    info "=== Sandbox Control Plane K8s Deployment ==="
    info "Deployment Mode: s3fs + bind mount (适合 OrbStack/Docker Desktop/单节点环境)"
    info ""

    # 显示资源说明
    if [ "$show_info" = "true" ]; then
        show_resource_info
        return 0
    fi

    # 检测 K8s 类型
    if [ "$cluster_type" = "auto" ]; then
        # OrbStack detection (check before Docker Desktop)
        if pgrep -f "OrbStack" > /dev/null || docker context inspect 2>/dev/null | grep -q "orbstack"; then
            cluster_type="orbstack"
        elif pgrep -f "Docker Desktop" > /dev/null || docker context inspect 2>/dev/null | grep -q "docker-desktop"; then
            cluster_type="docker-desktop"
        elif kind get clusters 2>/dev/null | grep -q "kind"; then
            cluster_type="kind"
        elif pgrep -f "minikube" > /dev/null; then
            cluster_type="minikube"
        elif [ -f /etc/rancher/k3s/k3s.yaml ]; then
            cluster_type="k3s"
        else
            cluster_type="generic"
        fi
    fi

    info "Detected cluster type: $cluster_type"
    info ""

    # 执行部署步骤
    check_kubectl
    check_docker
    detect_k8s_env

    if [ "$skip_build" != "true" ]; then
        build_image
        load_image "$cluster_type"
    else
        info "Skipping image build"
    fi

    deploy_resources
    wait_for_control_plane
    show_status
    verify_deployment
    setup_port_forward

    info ""
    info "=== Deployment Complete! ==="
    info ""
}

# 帮助信息
show_help() {
    echo "Usage: $0 [cluster_type] [skip_build] [show_info]"
    echo ""
    echo "Arguments:"
    echo "  cluster_type  Type of Kubernetes cluster"
    echo "                Options: orbstack, docker-desktop, kind, minikube, k3s, auto"
    echo "                Default: auto (automatically detect)"
    echo "  skip_build    Skip Docker image build"
    echo "                Options: true, false"
    echo "                Default: false"
    echo "  show_info     Show resource information only"
    echo "                Options: true, false"
    echo "                Default: false"
    echo ""
    echo "Examples:"
    echo "  $0                      # Auto-detect cluster and build image"
    echo "  $0 docker-desktop        # Deploy to Docker Desktop"
    echo "  $0 kind true            # Deploy to Kind without rebuilding"
    echo "  $0 auto false true      # Show resource information only"
    echo ""
    echo "Resource Information:"
    echo "  $0 --info               # Show detailed resource information"
    echo ""
}

# 处理帮助参数
if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    show_help
    exit 0
fi

if [ "$1" = "--info" ]; then
    show_resource_info
    exit 0
fi

# 执行主函数
main "$@"
