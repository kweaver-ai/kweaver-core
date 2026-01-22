
# =============================================================================
# Kubernetes Infrastructure Initialization Script
# =============================================================================
# Features:
#   1. Initialize K8s master node with scheduling enabled
#   2. Auto-install CNI (Calico) and DNS (CoreDNS)
#   3. Install Helm 3
#   4. Install single-node MariaDB 11 via Helm
#   5. Install single-node Redis 7 via Helm
# =============================================================================

# =============================================================================
# Global Configuration Variables
# =============================================================================
# Script directory (used for local chart paths)
SCRIPT_DIR="${SCRIPT_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"

# Local config/manifest directory (vendored files to avoid runtime fetching)
CONF_DIR="${CONF_DIR:-${SCRIPT_DIR}/conf}"

CONFIG_YAML_PATH="${CONFIG_YAML_PATH:-${CONF_DIR}/config.yaml}"

AUTO_GENERATE_CONFIG="${AUTO_GENERATE_CONFIG:-true}"

# Local Helm charts directory
LOCAL_CHARTS_DIR="${LOCAL_CHARTS_DIR:-${SCRIPT_DIR}/charts}"

# Default namespace for infrastructure components (MariaDB/Redis/Kafka/OpenSearch, etc.)
RESOURCE_NAMESPACE="${RESOURCE_NAMESPACE:-resource}"

# Generate a random password
generate_random_password() {
    local length="${1:-16}"
    # Use tr to get only alphanumeric characters and some safe special characters
    cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w "${length}" | head -n 1
}

# Check if component is already installed in Helm
is_helm_installed() {
    local release="$1"
    local ns="$2"
    helm list -n "${ns}" --short | grep -q "^${release}$"
}

# Get existing password from config.yaml if it exists
get_existing_password() {
    local key="$1"
    if [[ -f "${CONFIG_YAML_PATH}" ]]; then
        grep "${key}:" "${CONFIG_YAML_PATH}" | awk '{print $2}' | tr -d '"'\'' '
    fi
}

# Check if RDS is internal (MariaDB installed in cluster)
is_rds_internal() {
    if [[ ! -f "${CONFIG_YAML_PATH}" ]]; then
        return 1
    fi
    # Check if rds section has source_type: internal
    grep -A 10 "^  rds:" "${CONFIG_YAML_PATH}" | grep -q "source_type: internal"
}

# Show prominent warning when RDS is external and manual SQL import is required
warn_external_rds_sql_required() {
    local module_name="$1"
    local sql_dir="$2"
    
    echo ""
    echo "╔════════════════════════════════════════════════════════════════════════════╗"
    echo "║                                                                            ║"
    echo "║  ⚠️  WARNING: EXTERNAL DATABASE - MANUAL SQL INITIALIZATION REQUIRED  ⚠️   ║"
    echo "║                                                                            ║"
    echo "╠════════════════════════════════════════════════════════════════════════════╣"
    echo "║                                                                            ║"
    echo "║  RDS source_type is set to 'external' in config.yaml.                      ║"
    echo "║  You MUST manually execute SQL scripts to initialize the database.         ║"
    echo "║                                                                            ║"
    echo "║  Module: ${module_name}"
    echo "║  SQL Directory: ${sql_dir}"
    echo "║                                                                            ║"
    echo "║  Steps:                                                                    ║"
    echo "║    1. Connect to your external database server                             ║"
    echo "║    2. Execute all .sql files in the directory above                        ║"
    echo "║    3. Ensure all required databases and tables are created                 ║"
    echo "║                                                                            ║"
    echo "╚════════════════════════════════════════════════════════════════════════════╝"
    echo ""
}

# Image registry prefix loaded from conf/config.yaml (image.registry) or env
IMAGE_REGISTRY="${IMAGE_REGISTRY:-}"

# Kubernetes Network Configuration
POD_CIDR="${POD_CIDR:-192.169.0.0/16}"
SERVICE_CIDR="${SERVICE_CIDR:-10.96.0.0/12}"

# Kubernetes API Server Configuration
API_SERVER_ADVERTISE_ADDRESS="${API_SERVER_ADVERTISE_ADDRESS:-}"

# Kubernetes Image Repository Configuration
IMAGE_REPOSITORY="${IMAGE_REPOSITORY:-registry.aliyuncs.com/google_containers}"

# Kubernetes yum repo (Aliyun mirror) for kubeadm/kubelet/kubectl/cri-tools
K8S_RPM_REPO_BASEURL="${K8S_RPM_REPO_BASEURL:-https://mirrors.aliyun.com/kubernetes-new/core/stable/v1.28/rpm/}"
K8S_RPM_REPO_GPGKEY="${K8S_RPM_REPO_GPGKEY:-https://mirrors.aliyun.com/kubernetes-new/core/stable/v1.28/rpm/repodata/repomd.xml.key}"

# Flannel CNI Image Repository Configuration
FLANNEL_IMAGE_REPO="${FLANNEL_IMAGE_REPO:-swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/}"
FLANNEL_MANIFEST_PATH="${FLANNEL_MANIFEST_PATH:-${CONF_DIR}/kube-flannel.yml}"
FLANNEL_MANIFEST_URL="${FLANNEL_MANIFEST_URL:-https://raw.githubusercontent.com/flannel-io/flannel/v0.25.5/Documentation/kube-flannel.yml}"


# Helm Configuration
HELM_REPO_BITNAMI="${HELM_REPO_BITNAMI:-https://charts.bitnami.com/bitnami}"
HELM_REPO_INGRESS_NGINX="${HELM_REPO_INGRESS_NGINX:-https://kubernetes.github.io/ingress-nginx}"
HELM_REPO_OPENSEARCH="${HELM_REPO_OPENSEARCH:-https://opensearch-project.github.io/helm-charts}"
HELM_INSTALL_SCRIPT_PATH="${HELM_INSTALL_SCRIPT_PATH:-${CONF_DIR}/get-helm-3}"
HELM_INSTALL_SCRIPT_URL="${HELM_INSTALL_SCRIPT_URL:-https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3}"
HELM_VERSION="${HELM_VERSION:-v3.19.0}"
HELM_TARBALL_BASEURL="${HELM_TARBALL_BASEURL:-https://repo.huaweicloud.com/helm/${HELM_VERSION}/}"

# Global Helm Chart Configuration (for Studio, Ontology, and other modules)
HELM_CHART_VERSION="${HELM_CHART_VERSION:-}"
HELM_CHART_REPO_URL="${HELM_CHART_REPO_URL:-https://kweaver-ai.github.io/helm-repo/}"
HELM_CHART_REPO_NAME="${HELM_CHART_REPO_NAME:-kweaver}"

DOCKER_IO_MIRROR_PREFIX="${DOCKER_IO_MIRROR_PREFIX:-swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/}"
DOCKER_CE_REPO_URL="${DOCKER_CE_REPO_URL:-http://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo}"
LOCALPV_PROVISIONER_IMAGE="${LOCALPV_PROVISIONER_IMAGE:-swr.cn-east-3.myhuaweicloud.com/kweaver-ai/rancher/local-path-provisioner:v0.0.32}"
LOCALPV_HELPER_IMAGE="${LOCALPV_HELPER_IMAGE:-swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/busybox:1.36.1}"
LOCALPV_MANIFEST_PATH="${LOCALPV_MANIFEST_PATH:-${CONF_DIR}/local-path-storage.yaml}"
LOCALPV_MANIFEST_URL="${LOCALPV_MANIFEST_URL:-https://raw.githubusercontent.com/rancher/local-path-provisioner/v0.0.32/deploy/local-path-storage.yaml}"
LOCALPV_BASE_PATH="${LOCALPV_BASE_PATH:-/opt/local-path-provisioner}"
LOCALPV_SET_DEFAULT="${LOCALPV_SET_DEFAULT:-true}"
AUTO_INSTALL_LOCALPV="${AUTO_INSTALL_LOCALPV:-true}"
STORAGE_STORAGE_CLASS_NAME="${STORAGE_STORAGE_CLASS_NAME:-}"

# MariaDB Configuration
MARIADB_NAMESPACE="${MARIADB_NAMESPACE:-${RESOURCE_NAMESPACE}}"
MARIADB_IMAGE="${MARIADB_IMAGE:-}"
MARIADB_IMAGE_REPOSITORY="${MARIADB_IMAGE_REPOSITORY:-mariadb}"
MARIADB_IMAGE_TAG="${MARIADB_IMAGE_TAG:-11.4.7}"
MARIADB_IMAGE_FALLBACK="${MARIADB_IMAGE_FALLBACK:-mariadb:11.4.7}"
MARIADB_VERSION="${MARIADB_VERSION:-11.4}"
MARIADB_CHART_VERSION="${MARIADB_CHART_VERSION:-1.0.0}"
MARIADB_CHART_TGZ="${MARIADB_CHART_TGZ:-${SCRIPT_DIR}/charts/proton-mariadb-${MARIADB_CHART_VERSION}.tgz}"
MARIADB_PERSISTENCE_ENABLED="${MARIADB_PERSISTENCE_ENABLED:-true}"
MARIADB_STORAGE_CLASS="${MARIADB_STORAGE_CLASS:-}"
MARIADB_PURGE_PVC="${MARIADB_PURGE_PVC:-false}"
MARIADB_ROOT_PASSWORD="${MARIADB_ROOT_PASSWORD:-}"
MARIADB_DATABASE="${MARIADB_DATABASE:-adp}"
MARIADB_USER="${MARIADB_USER:-adp}"
MARIADB_PASSWORD="${MARIADB_PASSWORD:-}"
MARIADB_STORAGE_SIZE="${MARIADB_STORAGE_SIZE:-10Gi}"
MARIADB_MAX_CONNECTIONS="${MARIADB_MAX_CONNECTIONS:-5000}"

# Redis Configuration
REDIS_NAMESPACE="${REDIS_NAMESPACE:-${RESOURCE_NAMESPACE}}"
REDIS_VERSION="${REDIS_VERSION:-7.4}"
REDIS_CHART_VERSION="${REDIS_CHART_VERSION:-1.11.2}"
REDIS_CHART_TGZ="${REDIS_CHART_TGZ:-${SCRIPT_DIR}/charts/proton-redis-${REDIS_CHART_VERSION}.tgz}"
REDIS_LOCAL_CHART_DIR="${REDIS_LOCAL_CHART_DIR:-${SCRIPT_DIR}/charts/proton-redis}"
REDIS_ARCHITECTURE="${REDIS_ARCHITECTURE:-sentinel}"  # standalone or sentinel
REDIS_IMAGE="${REDIS_IMAGE:-}"
REDIS_IMAGE_REGISTRY="${REDIS_IMAGE_REGISTRY:-}"
REDIS_IMAGE_REPOSITORY="${REDIS_IMAGE_REPOSITORY:-proton/proton-redis}"
REDIS_IMAGE_TAG="${REDIS_IMAGE_TAG:-1.11.2-20251029.2.169ac3c0}"
REDIS_PERSISTENCE_ENABLED="${REDIS_PERSISTENCE_ENABLED:-true}"
REDIS_STORAGE_CLASS="${REDIS_STORAGE_CLASS:-}"
REDIS_PURGE_PVC="${REDIS_PURGE_PVC:-true}"
REDIS_PASSWORD="${REDIS_PASSWORD:-}"
REDIS_STORAGE_SIZE="${REDIS_STORAGE_SIZE:-5Gi}"
REDIS_MASTER_GROUP_NAME="${REDIS_MASTER_GROUP_NAME:-mymaster}"
REDIS_REPLICA_COUNT="${REDIS_REPLICA_COUNT:-1}"
REDIS_SENTINEL_QUORUM="${REDIS_SENTINEL_QUORUM:-1}"

# Kafka Configuration
KAFKA_NAMESPACE="${KAFKA_NAMESPACE:-${RESOURCE_NAMESPACE}}"
KAFKA_RELEASE_NAME="${KAFKA_RELEASE_NAME:-kafka}"
KAFKA_CHART_VERSION="${KAFKA_CHART_VERSION:-32.4.3}"
KAFKA_CHART_TGZ="${KAFKA_CHART_TGZ:-${SCRIPT_DIR}/charts/kafka-${KAFKA_CHART_VERSION}.tgz}"
# NOTE: Bitnami Kafka chart expects Bitnami Kafka images (/opt/bitnami/kafka/*).
# NOTE: Kafka 4.0 drops support for some older client protocol versions. Some apps (e.g. older Go clients)
# may still send JoinGroup v1 and will fail with:
#   UnsupportedVersionException: Received request for api with key 11 (JoinGroup) and unsupported version 1
# Default to a Kafka 3.x image for broader client compatibility; you can override via KAFKA_IMAGE/KAFKA_IMAGE_TAG.
# Use an SWR mirror by default to improve pull reliability in restricted networks.
KAFKA_IMAGE="${KAFKA_IMAGE:-swr.cn-east-3.myhuaweicloud.com/kweaver-ai/bitnami/kafka:3.9.0-debian-12-r10}"
KAFKA_IMAGE_REPOSITORY="${KAFKA_IMAGE_REPOSITORY:-bitnami/kafka}"
KAFKA_IMAGE_TAG="${KAFKA_IMAGE_TAG:-3.9.0-debian-12-r10}"
KAFKA_IMAGE_FALLBACK="${KAFKA_IMAGE_FALLBACK:-swr.cn-east-3.myhuaweicloud.com/kweaver-ai/bitnami/kafka:3.9.0-debian-12-r10}"
KAFKA_HELM_TIMEOUT="${KAFKA_HELM_TIMEOUT:-1800s}"
# NOTE: --atomic will auto-uninstall on failure, which makes debugging hard. Default to false.
KAFKA_HELM_ATOMIC="${KAFKA_HELM_ATOMIC:-false}"
KAFKA_READY_TIMEOUT="${KAFKA_READY_TIMEOUT:-600s}"
KAFKA_HEAP_OPTS="${KAFKA_HEAP_OPTS:--Xms256m -Xmx256m}"
KAFKA_MEMORY_REQUEST="${KAFKA_MEMORY_REQUEST:-256Mi}"
KAFKA_MEMORY_LIMIT="${KAFKA_MEMORY_LIMIT:-512Mi}"
KAFKA_PERSISTENCE_ENABLED="${KAFKA_PERSISTENCE_ENABLED:-true}"
KAFKA_STORAGE_CLASS="${KAFKA_STORAGE_CLASS:-}"
KAFKA_STORAGE_SIZE="${KAFKA_STORAGE_SIZE:-8Gi}"
# Delete Kafka PVCs by default on uninstall (set false to retain data)
KAFKA_PURGE_PVC="${KAFKA_PURGE_PVC:-true}"
KAFKA_AUTH_ENABLED="${KAFKA_AUTH_ENABLED:-true}"
KAFKA_PROTOCOL="${KAFKA_PROTOCOL:-SASL_PLAINTEXT}"
KAFKA_SASL_MECHANISM="${KAFKA_SASL_MECHANISM:-PLAIN}"
KAFKA_CLIENT_USER="${KAFKA_CLIENT_USER:-kafkauser}"
KAFKA_CLIENT_PASSWORD="${KAFKA_CLIENT_PASSWORD:-}"
KAFKA_INTERBROKER_USER="${KAFKA_INTERBROKER_USER:-inter_broker_user}"
KAFKA_INTERBROKER_PASSWORD="${KAFKA_INTERBROKER_PASSWORD:-}"
KAFKA_CONTROLLER_USER="${KAFKA_CONTROLLER_USER:-controller_user}"
KAFKA_CONTROLLER_PASSWORD="${KAFKA_CONTROLLER_PASSWORD:-}"
KAFKA_SASL_SECRET_NAME="${KAFKA_SASL_SECRET_NAME:-${KAFKA_RELEASE_NAME}-sasl}"
KAFKA_REPLICAS="${KAFKA_REPLICAS:-1}"
KAFKA_AUTO_CREATE_TOPICS_ENABLE="${KAFKA_AUTO_CREATE_TOPICS_ENABLE:-true}"

# OpenSearch Configuration
LOCAL_OPENSEARCH_CHARTS_DIR="${LOCAL_OPENSEARCH_CHARTS_DIR:-${SCRIPT_DIR}/charts/opensearch}"
OPENSEARCH_NAMESPACE="${OPENSEARCH_NAMESPACE:-${RESOURCE_NAMESPACE}}"
OPENSEARCH_RELEASE_NAME="${OPENSEARCH_RELEASE_NAME:-opensearch}"
OPENSEARCH_CLUSTER_NAME="${OPENSEARCH_CLUSTER_NAME:-opensearch-cluster}"
OPENSEARCH_NODE_GROUP="${OPENSEARCH_NODE_GROUP:-master}"
OPENSEARCH_CHART_VERSION="${OPENSEARCH_CHART_VERSION:-2.36.0}"
OPENSEARCH_CHART_TGZ="${OPENSEARCH_CHART_TGZ:-${SCRIPT_DIR}/charts/opensearch-${OPENSEARCH_CHART_VERSION}.tgz}"
OPENSEARCH_IMAGE="${OPENSEARCH_IMAGE:-swr.cn-east-3.myhuaweicloud.com/kweaver-ai/opensearchproject/opensearch:2.19.4}"
OPENSEARCH_IMAGE_REPOSITORY="${OPENSEARCH_IMAGE_REPOSITORY:-swr.cn-east-3.myhuaweicloud.com/kweaver-ai/opensearchproject/opensearch}"
OPENSEARCH_IMAGE_TAG="${OPENSEARCH_IMAGE_TAG:-2.19.4}"
OPENSEARCH_IMAGE_FALLBACK="${OPENSEARCH_IMAGE_FALLBACK:-swr.cn-east-3.myhuaweicloud.com/kweaver-ai/opensearchproject/opensearch:2.19.4}"
# OpenSearch chart uses busybox initContainers (fsgroup-volume/sysctl) by default; set a mirror to avoid Docker Hub pulls.
OPENSEARCH_INIT_IMAGE="${OPENSEARCH_INIT_IMAGE:-${LOCALPV_HELPER_IMAGE}}"
OPENSEARCH_JAVA_OPTS="${OPENSEARCH_JAVA_OPTS:--Xms512m -Xmx512m -XX:MaxDirectMemorySize=128m}"
OPENSEARCH_MEMORY_REQUEST="${OPENSEARCH_MEMORY_REQUEST:-512Mi}"
# NOTE: OpenSearch uses heap + direct memory + native overhead. 768Mi is too tight for -Xmx512m.
# Increased to 2Gi to support plugin installation (IK analyzer, etc.)
OPENSEARCH_MEMORY_LIMIT="${OPENSEARCH_MEMORY_LIMIT:-2048Mi}"
OPENSEARCH_PROTOCOL="${OPENSEARCH_PROTOCOL:-http}" # http (default) or https (requires enabling security)
OPENSEARCH_DISABLE_SECURITY="${OPENSEARCH_DISABLE_SECURITY:-}"
OPENSEARCH_SINGLE_NODE="${OPENSEARCH_SINGLE_NODE:-true}"
OPENSEARCH_PERSISTENCE_ENABLED="${OPENSEARCH_PERSISTENCE_ENABLED:-true}"
OPENSEARCH_STORAGE_CLASS="${OPENSEARCH_STORAGE_CLASS:-}"
OPENSEARCH_STORAGE_SIZE="${OPENSEARCH_STORAGE_SIZE:-8Gi}"
OPENSEARCH_PURGE_PVC="${OPENSEARCH_PURGE_PVC:-false}"
OPENSEARCH_INITIAL_ADMIN_PASSWORD="${OPENSEARCH_INITIAL_ADMIN_PASSWORD:-OpenSearch@123456}"
OPENSEARCH_SYSCTL_INIT_ENABLED="${OPENSEARCH_SYSCTL_INIT_ENABLED:-true}"
OPENSEARCH_SYSCTL_VM_MAX_MAP_COUNT="${OPENSEARCH_SYSCTL_VM_MAX_MAP_COUNT:-262144}"

# MongoDB Configuration
LOCAL_MONGODB_CHARTS_DIR="${LOCAL_MONGODB_CHARTS_DIR:-${SCRIPT_DIR}/charts/mongodb}"
MONGODB_CHART_TGZ="${MONGODB_CHART_TGZ:-${SCRIPT_DIR}/charts/mongodb-1.0.0.tgz}"
MONGODB_NAMESPACE="${MONGODB_NAMESPACE:-${RESOURCE_NAMESPACE}}"
MONGODB_RELEASE_NAME="${MONGODB_RELEASE_NAME:-mongodb}"
MONGODB_IMAGE="${MONGODB_IMAGE:-}"
MONGODB_IMAGE_REPOSITORY="${MONGODB_IMAGE_REPOSITORY:-swr.cn-east-3.myhuaweicloud.com/kweaver-ai/proton/proton-mongo}"
MONGODB_IMAGE_TAG="${MONGODB_IMAGE_TAG:-2.1.0-feature-mongo-4.4.30}"
MONGODB_REPLICAS="${MONGODB_REPLICAS:-1}"
MONGODB_REPLSET_ENABLED="${MONGODB_REPLSET_ENABLED:-true}"  # Default: single-node replica set mode (requires keyfile)
MONGODB_REPLSET_NAME="${MONGODB_REPLSET_NAME:-rs0}"
MONGODB_SERVICE_TYPE="${MONGODB_SERVICE_TYPE:-ClusterIP}"
MONGODB_SERVICE_PORT="${MONGODB_SERVICE_PORT:-30280}"
MONGODB_WIRED_TIGER_CACHE_SIZE_GB="${MONGODB_WIRED_TIGER_CACHE_SIZE_GB:-4}"
MONGODB_STORAGE_CLASS="${MONGODB_STORAGE_CLASS:-}"
MONGODB_STORAGE_SIZE="${MONGODB_STORAGE_SIZE:-10Gi}"
MONGODB_SECRET_NAME="${MONGODB_SECRET_NAME:-mongodb-secret}"
MONGODB_SECRET_USERNAME="${MONGODB_SECRET_USERNAME:-admin}"
MONGODB_SECRET_PASSWORD="${MONGODB_SECRET_PASSWORD:-}"
MONGODB_RESOURCES_REQUESTS_CPU="${MONGODB_RESOURCES_REQUESTS_CPU:-100m}"
MONGODB_RESOURCES_REQUESTS_MEMORY="${MONGODB_RESOURCES_REQUESTS_MEMORY:-128Mi}"
MONGODB_RESOURCES_LIMITS_CPU="${MONGODB_RESOURCES_LIMITS_CPU:-1}"
MONGODB_RESOURCES_LIMITS_MEMORY="${MONGODB_RESOURCES_LIMITS_MEMORY:-1Gi}"

# Zookeeper Configuration
LOCAL_ZOOKEEPER_CHARTS_DIR="${LOCAL_ZOOKEEPER_CHARTS_DIR:-${SCRIPT_DIR}/charts/zookeeper}"
ZOOKEEPER_CHART_TGZ="${ZOOKEEPER_CHART_TGZ:-${SCRIPT_DIR}/charts/proton-zookeeper-5.6.0.tgz}"
ZOOKEEPER_NAMESPACE="${ZOOKEEPER_NAMESPACE:-${RESOURCE_NAMESPACE}}"
ZOOKEEPER_RELEASE_NAME="${ZOOKEEPER_RELEASE_NAME:-zookeeper}"
ZOOKEEPER_CHART_REF="${ZOOKEEPER_CHART_REF:-}"  # e.g., "dip/zookeeper" for remote repo, or local path
ZOOKEEPER_CHART_VERSION="${ZOOKEEPER_CHART_VERSION:-}"  # Chart version (--version)
ZOOKEEPER_CHART_DEVEL="${ZOOKEEPER_CHART_DEVEL:-false}"  # Use --devel flag
ZOOKEEPER_VALUES_FILE="${ZOOKEEPER_VALUES_FILE:-}"  # Additional values file (e.g., conf/config.yaml)
ZOOKEEPER_REPLICAS="${ZOOKEEPER_REPLICAS:-1}"
ZOOKEEPER_IMAGE_REGISTRY="${ZOOKEEPER_IMAGE_REGISTRY:-swr.cn-east-3.myhuaweicloud.com/kweaver-ai}"
ZOOKEEPER_IMAGE_REPOSITORY="${ZOOKEEPER_IMAGE_REPOSITORY:-proton/proton-zookeeper}"
ZOOKEEPER_IMAGE_TAG="${ZOOKEEPER_IMAGE_TAG:-5.6.0-20250625.2.138fb9}"
ZOOKEEPER_EXPORTER_IMAGE_REPOSITORY="${ZOOKEEPER_EXPORTER_IMAGE_REPOSITORY:-proton/proton-zookeeper-exporter}"
ZOOKEEPER_EXPORTER_IMAGE_TAG="${ZOOKEEPER_EXPORTER_IMAGE_TAG:-5.6.0-20250625.2.138fb9}"
ZOOKEEPER_SERVICE_PORT="${ZOOKEEPER_SERVICE_PORT:-2181}"
ZOOKEEPER_EXPORTER_PORT="${ZOOKEEPER_EXPORTER_PORT:-9101}"
ZOOKEEPER_JMX_EXPORTER_PORT="${ZOOKEEPER_JMX_EXPORTER_PORT:-9995}"
ZOOKEEPER_STORAGE_CLASS="${ZOOKEEPER_STORAGE_CLASS:-}"
ZOOKEEPER_STORAGE_SIZE="${ZOOKEEPER_STORAGE_SIZE:-1Gi}"
ZOOKEEPER_PURGE_PVC="${ZOOKEEPER_PURGE_PVC:-true}"
ZOOKEEPER_RESOURCES_REQUESTS_CPU="${ZOOKEEPER_RESOURCES_REQUESTS_CPU:-500m}"
ZOOKEEPER_RESOURCES_REQUESTS_MEMORY="${ZOOKEEPER_RESOURCES_REQUESTS_MEMORY:-1Gi}"
ZOOKEEPER_RESOURCES_LIMITS_CPU="${ZOOKEEPER_RESOURCES_LIMITS_CPU:-1000m}"
ZOOKEEPER_RESOURCES_LIMITS_MEMORY="${ZOOKEEPER_RESOURCES_LIMITS_MEMORY:-2Gi}"
ZOOKEEPER_JVMFLAGS="${ZOOKEEPER_JVMFLAGS:--Xms500m -Xmx500m}"
ZOOKEEPER_SASL_ENABLED="${ZOOKEEPER_SASL_ENABLED:-true}"
ZOOKEEPER_SASL_USER="${ZOOKEEPER_SASL_USER:-kafka}"
ZOOKEEPER_SASL_PASSWORD="${ZOOKEEPER_SASL_PASSWORD:-}"
ZOOKEEPER_EXTRA_SET_VALUES="${ZOOKEEPER_EXTRA_SET_VALUES:-}"  # Additional --set values (space-separated, e.g., "image.registry=xxx key2=value2")

# Ingress-Nginx Configuration
INGRESS_NGINX_HTTP_PORT="${INGRESS_NGINX_HTTP_PORT:-30080}"
INGRESS_NGINX_HTTPS_PORT="${INGRESS_NGINX_HTTPS_PORT:-30443}"
INGRESS_NGINX_CLASS="${INGRESS_NGINX_CLASS:-class-443}"
INGRESS_NGINX_CONTROLLER_IMAGE="${INGRESS_NGINX_CONTROLLER_IMAGE:-swr.cn-east-3.myhuaweicloud.com/kweaver-ai/ingress-nginx/controller:v1.14.1}"
INGRESS_NGINX_CONTROLLER_IMAGE_REPOSITORY="${INGRESS_NGINX_CONTROLLER_IMAGE_REPOSITORY:-swr.cn-east-3.myhuaweicloud.com/kweaver-ai/ingress-nginx/controller}"
INGRESS_NGINX_CONTROLLER_IMAGE_TAG="${INGRESS_NGINX_CONTROLLER_IMAGE_TAG:-v1.14.1}"
INGRESS_NGINX_CHART_VERSION="${INGRESS_NGINX_CHART_VERSION:-4.13.1}"
INGRESS_NGINX_CHART_TGZ="${INGRESS_NGINX_CHART_TGZ:-${SCRIPT_DIR}/charts/ingress-nginx-${INGRESS_NGINX_CHART_VERSION}.tgz}"
INGRESS_NGINX_WEBHOOK_CERTGEN_IMAGE="${INGRESS_NGINX_WEBHOOK_CERTGEN_IMAGE:-swr.cn-east-3.myhuaweicloud.com/kweaver-ai/ingress-nginx/kube-webhook-certgen:v1.6.1}"
INGRESS_NGINX_WEBHOOK_CERTGEN_IMAGE_REPOSITORY="${INGRESS_NGINX_WEBHOOK_CERTGEN_IMAGE_REPOSITORY:-swr.cn-east-3.myhuaweicloud.com/kweaver-ai/ingress-nginx/kube-webhook-certgen}"
INGRESS_NGINX_WEBHOOK_CERTGEN_IMAGE_TAG="${INGRESS_NGINX_WEBHOOK_CERTGEN_IMAGE_TAG:-v1.6.1}"
INGRESS_NGINX_HOSTNETWORK="${INGRESS_NGINX_HOSTNETWORK:-true}"
INGRESS_NGINX_ADMISSION_WEBHOOKS_ENABLED="${INGRESS_NGINX_ADMISSION_WEBHOOKS_ENABLED:-false}"
AUTO_INSTALL_INGRESS_NGINX="${AUTO_INSTALL_INGRESS_NGINX:-true}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

random_password() {
    if command -v openssl >/dev/null 2>&1; then
        openssl rand -base64 18 | tr -d '\n'
        return 0
    fi
    head -c 32 /dev/urandom | base64 | tr -d '\n' | head -c 24
}

# Quote a string for YAML single-quoted scalars.
yaml_quote() {
    local s="$1"
    s="${s//\'/\'\'}"
    printf "'%s'" "${s}"
}

get_config_image_registry() {
    local cfg="${CONFIG_YAML_PATH}"
    if [[ ! -f "${cfg}" ]]; then
        return 0
    fi

    awk '
      $1 == "image:" { in_image=1; next }
      in_image && $1 == "registry:" { print $2; exit }
      in_image && $0 ~ /^[^ ]/ { in_image=0 }
    ' "${cfg}" 2>/dev/null | sed -e 's/^["'\'']//; s/["'\'']$//' | tr -d '\r' || true
}

load_image_registry_from_config() {
    if [[ -n "${IMAGE_REGISTRY}" ]]; then
        return 0
    fi
    IMAGE_REGISTRY="$(get_config_image_registry)"
    IMAGE_REGISTRY="${IMAGE_REGISTRY%/}"
    if [[ -z "${IMAGE_REGISTRY}" ]]; then
        IMAGE_REGISTRY="swr.cn-east-3.myhuaweicloud.com/kweaver-ai"
    fi
}

image_from_registry() {
    local repository="$1"
    local tag="$2"
    local fallback="$3"

    load_image_registry_from_config
    if [[ -n "${IMAGE_REGISTRY}" ]]; then
        echo "${IMAGE_REGISTRY}/${repository}:${tag}"
    else
        echo "${fallback}"
    fi
}

get_secret_b64_key() {
    local namespace="$1"
    local name="$2"
    local key="$3"
    local safe_key="${key//\'/\\\'}"
    kubectl -n "${namespace}" get secret "${name}" -o "jsonpath={.data['${safe_key}']}" 2>/dev/null | base64 -d 2>/dev/null || true
}

first_service_with_port() {
    local namespace="$1"
    local selector="$2"
    local port="$3"
    kubectl -n "${namespace}" get svc -l "${selector}" -o jsonpath='{range .items[*]}{.metadata.name}{" "}{range .spec.ports[*]}{.port}{" "}{end}{"\n"}{end}' 2>/dev/null | \
        awk -v want="${port}" '$0 ~ (" " want " ") {print $1; exit}'
}

# Read vendored file if exists; otherwise fetch from URL.
read_or_fetch() {
    local path="$1"
    local url="$2"

    if [[ -n "${path}" && -f "${path}" ]]; then
        cat "${path}"
        return 0
    fi

    if [[ -z "${url}" ]]; then
        log_error "No local file found and no URL provided"
        return 1
    fi

    curl -fsSL "${url}"
}

# Initialize database by connecting to MariaDB pod and executing SQL files
# Usage: init_module_database "module_name" "sql_directory"
# Example: init_module_database "dataagent" "${SCRIPT_DIR}/scripts/sql/dataagent"
init_module_database() {
    local module_name="$1"
    local sql_dir="$2"
    local mariadb_namespace="${MARIADB_NAMESPACE:-resource}"
    
    if [[ -z "${module_name}" || -z "${sql_dir}" ]]; then
        log_error "Usage: init_module_database <module_name> <sql_directory>"
        return 1
    fi
    
    log_info "Initializing ${module_name} database..."
    
    # Check if MariaDB pod is running
    local mariadb_pod=$(kubectl get pods -n "${mariadb_namespace}" -l "app.kubernetes.io/name=proton-mariadb" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    if [[ -z "${mariadb_pod}" ]]; then
        log_error "MariaDB pod not found in namespace ${mariadb_namespace}"
        return 1
    fi
    
    log_info "Found MariaDB pod: ${mariadb_pod}"
    
    # Get MariaDB credentials from config.yaml (under depServices.rds section)
    local mariadb_user=$(grep -A 20 "^  rds:" "${CONFIG_YAML_PATH}" | grep "user:" | head -1 | awk '{print $2}' | tr -d "'\"")
    local mariadb_password=$(grep -A 20 "^  rds:" "${CONFIG_YAML_PATH}" | grep "password:" | head -1 | awk '{print $2}' | tr -d "'\"")
    
    # Set defaults if not found
    mariadb_user="${mariadb_user:-adp}"
    mariadb_password="${mariadb_password:-adp}"
    
    log_info "Using MariaDB user: ${mariadb_user}"
    
    # Check if SQL directory exists
    if [[ ! -d "${sql_dir}" ]]; then
        log_error "SQL directory not found: ${sql_dir}"
        return 1
    fi
    
    # Execute all SQL files in the directory in order
    local sql_files=($(find "${sql_dir}" -name "*.sql" -type f | sort))
    if [[ ${#sql_files[@]} -eq 0 ]]; then
        log_error "No SQL files found in ${sql_dir}"
        return 1
    fi
    
    for sql_file in "${sql_files[@]}"; do
        local sql_filename=$(basename "${sql_file}")
        log_info "Executing SQL file: ${sql_filename}"
        
        # Execute SQL in MariaDB pod using cat pipe with mariadb command
        local exec_output
        exec_output=$(cat "${sql_file}" | kubectl exec -i -n "${mariadb_namespace}" "${mariadb_pod}" -- \
            mariadb -u "${mariadb_user}" -p"${mariadb_password}" 2>&1)
        
        if [[ $? -ne 0 ]]; then
            log_error "Failed to execute SQL file ${sql_filename} in MariaDB pod"
            log_error "Error output: ${exec_output}"
            return 1
        fi
        
        log_info "✓ ${sql_filename} executed successfully"
    done
    
    log_info "✓ ${module_name} database initialized successfully"
}

# Create databases without initializing SQL
# Usage: create_databases "database_name1" "database_name2" ...
# Example: create_databases "user_management" "anyshare" "policy_mgnt"
create_databases() {
    local mariadb_namespace="${MARIADB_NAMESPACE:-resource}"
    local db_user=$(grep -A 20 "^  rds:" "${CONFIG_YAML_PATH}" | grep "user:" | head -1 | awk '{print $2}' | tr -d "'\"")
    local root_password=$(grep -A 20 "^  rds:" "${CONFIG_YAML_PATH}" | grep "root_password:" | head -1 | awk '{print $2}' | tr -d "'\"")
    
    # Set defaults if not found
    db_user="${db_user:-adp}"
    root_password="${root_password:-}"
    
    log_info "Creating databases..."
    
    # Check if MariaDB pod is running
    local mariadb_pod=$(kubectl get pods -n "${mariadb_namespace}" -l "app.kubernetes.io/name=proton-mariadb" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    if [[ -z "${mariadb_pod}" ]]; then
        log_error "MariaDB pod not found in namespace ${mariadb_namespace}"
        return 1
    fi
    
    log_info "Found MariaDB pod: ${mariadb_pod}"
    
    # Create each database using root account
    for db_name in "$@"; do
        log_info "Creating database: ${db_name}"
        
        # Create database and grant privileges using root account
        if [[ -n "${root_password}" ]]; then
            kubectl exec -n "${mariadb_namespace}" "${mariadb_pod}" -- mariadb -u root -p"${root_password}" -e "
                CREATE DATABASE IF NOT EXISTS \`${db_name}\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
                GRANT ALL PRIVILEGES ON \`${db_name}\`.* TO '${db_user}'@'%';
                FLUSH PRIVILEGES;
            " 2>/dev/null || log_warn "Failed to create database ${db_name} (may already exist)"
        else
            log_error "root_password not found in config.yaml, cannot create database ${db_name}"
            return 1
        fi
    done
    
    log_info "✓ Databases created successfully"
}

# Show cluster status
show_status() {
    log_info "Cluster Status:"
    echo ""
    kubectl get nodes -o wide
    echo ""
    kubectl get pods -A
}
