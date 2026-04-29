#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONF_DIR="${CONF_DIR:-${HOME}/.kweaver-ai}"
CONFIG_YAML_PATH="${CONFIG_YAML_PATH:-${CONF_DIR}/config.yaml}"

# Global flag: skip all interactive prompts and use defaults
ASSUME_YES="${ASSUME_YES:-false}"

# Global flag: bypass "skip when chart version unchanged" optimization so that
# helm upgrade re-renders templates with the latest values.yaml. Use this after
# editing config.yaml on a previously-installed cluster.
FORCE_UPGRADE="${FORCE_UPGRADE:-false}"

# Fix paths to use script's conf directory, not user home
FLANNEL_MANIFEST_PATH="${SCRIPT_DIR}/conf/kube-flannel.yml"
LOCALPV_MANIFEST_PATH="${SCRIPT_DIR}/conf/local-path-storage.yaml"
HELM_INSTALL_SCRIPT_PATH="${SCRIPT_DIR}/conf/get-helm-3"

# Source all service libraries
source "${SCRIPT_DIR}/scripts/lib/common.sh"
source "${SCRIPT_DIR}/scripts/services/config.sh"
source "${SCRIPT_DIR}/scripts/services/k8s.sh"
source "${SCRIPT_DIR}/scripts/services/k3s.sh"
source "${SCRIPT_DIR}/scripts/services/storage.sh"
source "${SCRIPT_DIR}/scripts/services/mariadb.sh"
source "${SCRIPT_DIR}/scripts/services/redis.sh"
source "${SCRIPT_DIR}/scripts/services/kafka.sh"
source "${SCRIPT_DIR}/scripts/services/zookeeper.sh"
# source "${SCRIPT_DIR}/scripts/services/mongodb.sh"  # MongoDB disabled
source "${SCRIPT_DIR}/scripts/services/ingress_nginx.sh"
source "${SCRIPT_DIR}/scripts/services/opensearch.sh"
source "${SCRIPT_DIR}/scripts/services/core.sh"
source "${SCRIPT_DIR}/scripts/services/isf.sh"
source "${SCRIPT_DIR}/scripts/services/dip.sh"

usage() {
    echo "Kubernetes Infrastructure Initialization Script"
    echo ""
    echo "Usage: $0 <module> [action]"
    echo ""
    echo "Modules and Actions:"
    echo "  k8s install                   Initialize K8s master node with CNI and DNS"
    echo "  k8s reset                     Reset Kubernetes cluster state (kubeadm reset -f + cleanup)"
    echo "  k8s status                    Show cluster status"
    echo "  k3s install                   Install single-node k3s (Linux), Traefik disabled; uses ingress-nginx"
    echo "  k3s uninstall                 Run k3s-uninstall.sh (removes k3s)"
    echo "  k3s status                    Show cluster status (nodes and pods)"
    echo "  mariadb install               Install single-node MariaDB 11"
    echo "  mariadb uninstall             Uninstall MariaDB (optionally purge PVC)"
    echo "  redis install                 Install single-node Redis 7"
    echo "  redis uninstall               Uninstall Redis (PVCs will be deleted by default)"
    echo "  kafka install                 Install single-node Kafka"
    echo "  kafka uninstall               Uninstall Kafka (PVCs will be deleted by default)"
    echo "  opensearch install            Install single-node OpenSearch"
    echo "  opensearch uninstall          Uninstall OpenSearch (optionally purge PVC)"
    echo "  zookeeper install             Install single-node Zookeeper"
    echo "  zookeeper uninstall           Uninstall Zookeeper (PVCs will be deleted by default)"
    echo "  ingress-nginx install         Install ingress-nginx-controller"
    echo "  ingress-nginx uninstall       Uninstall ingress-nginx-controller"
    echo "  kweaver-core install          Install KWeaver Core services; auto-installs K8s/data services if missing"
    echo "  kweaver-core install --minimum  Minimum install (skip auth & business-domain modules)"
    echo "  kweaver-core download         Download/update KWeaver Core charts into deploy/.tmp/charts"
    echo "  kweaver-core uninstall        Uninstall KWeaver Core services"
    echo "  kweaver-core status           Show KWeaver Core services status"
    echo "                                Use --set to pass custom values to all charts"
    echo "  isf install                   Install ISF services; auto-installs K8s/data services if missing"
    echo "  isf download                  Download/update ISF charts into deploy/.tmp/charts"
    echo "  isf uninstall                 Uninstall ISF services"
    echo "  isf status                    Show ISF services status"
    echo "  kweaver-dip install           Install KWeaver DIP services (17 charts); auto-installs K8s/data services if missing"
    echo "  kweaver-dip download          Download/update DIP + Core + ISF charts into deploy/.tmp/charts"
    echo "  kweaver-dip uninstall         Uninstall KWeaver DIP services"
    echo "  kweaver-dip status            Show KWeaver DIP services status"
    echo "  etrino install                Install Etrino services (vega-hdfs, vega-calculate, vega-metadata)"
    echo "  etrino uninstall              Uninstall Etrino services"
    echo "  etrino status                 Show Etrino services status"
    echo "  all install                   Run full initialization (k8s + mariadb + redis + ingress-nginx)"
    echo ""
    echo "Examples:"
    echo "  $0 k8s install                # Initialize K8s master node with default settings"
    echo "  $0 k8s reset                  # Reset cluster state before re-install"
    echo "  $0 k8s status                 # Show cluster status"
    echo "  $0 k3s install                # Install single-node k3s + ingress-nginx (Linux)"
    echo "  $0 --distro=k3s kweaver-core install --minimum  # Same as KUBE_DISTRO=k3s env (k3s not kubeadm)"
    echo "  POD_CIDR=10.0.0.0/16 $0 k8s install  # Initialize with custom POD_CIDR"
    echo "  $0 mariadb install            # Install MariaDB"
    echo "  $0 mariadb uninstall          # Uninstall MariaDB"
    echo "  $0 mariadb uninstall --delete-data  # Uninstall MariaDB and delete PVC (data loss!)"
    echo "  MARIADB_PURGE_PVC=true $0 mariadb uninstall  # Same as --delete-data (data loss!)"
    echo "  $0 redis install              # Install Redis"
    echo "  $0 redis uninstall            # Uninstall Redis"
    echo "  $0 redis uninstall                         # Uninstall Redis (PVCs deleted by default)"
    echo "  REDIS_PURGE_PVC=false $0 redis uninstall   # Uninstall Redis but keep PVCs"
    echo "  $0 kafka install              # Install Kafka"
    echo "  $0 kafka uninstall                         # Uninstall Kafka (PVCs deleted by default)"
    echo "  KAFKA_PURGE_PVC=false $0 kafka uninstall   # Uninstall Kafka but keep PVCs"
    echo "  $0 opensearch install         # Install OpenSearch"
    echo "  $0 opensearch uninstall       # Uninstall OpenSearch"
    echo "  OPENSEARCH_PURGE_PVC=true $0 opensearch uninstall  # Uninstall OpenSearch and delete PVC (data loss!)"
    echo "  $0 zookeeper install          # Install Zookeeper"
    echo "  $0 zookeeper uninstall        # Uninstall Zookeeper (PVCs deleted by default)"
    echo "  ZOOKEEPER_PURGE_PVC=false $0 zookeeper uninstall  # Uninstall Zookeeper but keep PVC's"
    echo "  # Install from remote repo with version and devel:"
    echo "  ZOOKEEPER_CHART_REF=dip/zookeeper ZOOKEEPER_CHART_VERSION=0.0.0-feature-800792 ZOOKEEPER_CHART_DEVEL=true $0 zookeeper install"
    echo "  # Install with additional values file and --set:"
    echo "  ZOOKEEPER_VALUES_FILE=~/.kweaver-ai/config.yaml ZOOKEEPER_EXTRA_SET_VALUES='image.registry=swr.cn-east-3.myhuaweicloud.com/kweaver-ai' $0 zookeeper install"
    echo "  $0 ingress-nginx install      # Install ingress-nginx-controller"
    echo "  $0 ingress-nginx uninstall    # Uninstall ingress-nginx-controller"
    echo "  $0 kweaver-dip install        # Install KWeaver DIP (auto-installs K8s/data services if absent)"
    echo "  $0 kweaver-dip download       # Download DIP + dependency charts into deploy/.tmp/charts"
    echo "  $0 kweaver-dip download --charts_dir=/path/to/charts # Download DIP charts into a specific local directory"
    echo "  $0 kweaver-dip install --charts_dir=/path/to/charts  # Install DIP from a local charts directory"
    echo "  $0 kweaver-dip uninstall      # Uninstall KWeaver DIP services"
    echo "  $0 kweaver-dip status         # Show KWeaver DIP services status"
    echo "  $0 kweaver-dip install --confirm-missing-openclaw-paths  # Continue even if configured dipStudio OpenClaw paths do not exist"
    echo "  $0 etrino install             # Install Etrino services only"
    echo "  $0 etrino status              # Show Etrino services status"
    echo "  $0 etrino uninstall           # Uninstall Etrino services"
    echo "  $0 config generate            # Generate/update ~/.kweaver-ai/config.yaml"
    echo "  $0 all install                # Full initialization with all components"
    echo ""
    echo "Global Options:"
    echo "  -y, --yes                     Skip all interactive prompts and use defaults"
    echo "  --force-upgrade               Always run helm upgrade even if installed chart version equals target."
    echo "                                Use this after editing config.yaml on a previously-installed cluster."
    echo "  --distro=k3s|kubeadm          Cluster bootstrap when modules auto-ensure K8s (default: kubeadm)."
    echo "                                Exported as KUBE_DISTRO=...; k3s uses single-node k3s instead of kubeadm."
    echo "  --config=<path>               Specify config.yaml path (values file for helm installs)"
    echo "                                (default: ~/.kweaver-ai/config.yaml or \$CONFIG_YAML_PATH env var)"
    echo "  --charts_dir=<path>           Use a specific local chart directory for download/install"
    echo "                                install only uses local charts when this option is explicitly set"
    echo "  --version_file=<path>         Use an aggregate release manifest to resolve exact chart versions"
    echo "                                (default auto path: deploy/release-manifests/<version>/<product>.yaml)"
    echo "  --confirm-missing-openclaw-paths"
    echo "                                Continue DIP install when dipStudio OpenClaw host paths are configured"
    echo "                                but missing on disk. Only applies to the dip-studio chart."
    echo "  --access_address=<addr>       KWeaver access address: host, host:port, or scheme://host:port/path"
    echo "                                Example: --access_address=10.0.0.5 or --access_address=https://kweaver.example.com:443"
    echo "  --api_server_address=<ip>     Kubernetes API server advertise address (must be a local interface IP)"
    echo "                                Defaults to auto-detect from hostname -I"
    echo "  --minimum, --min              Minimum install: skip auth & business-domain modules"
    echo "                                Equivalent to: --set auth.enabled=false --set businessDomain.enabled=false"
    echo "  --set <key>=<value>           Pass custom values to helm charts (can be used multiple times)"
    echo "                                Example: --set auth.enabled=false --set image.tag=latest"
    echo ""
    echo "Environment (optional, kweaver-core install):"
    echo "  (Context Loader ADP import moved to deploy/onboard.sh after kweaver auth — kweaver call impex; see onboard -h.)"
    echo "  DEPLOY_BUSINESS_DOMAIN        x-business-domain for kweaver/onboard (default: bd_public)."
    echo ""
    echo "  $0 kweaver-core install --minimum                 # Minimum install (skip auth & business-domain)"
    echo "  $0 kweaver-core install --set auth.enabled=false  # Install KWeaver Core without ISF"
    echo "  $0 kweaver-core install --set auth.enabled=false --set businessDomain.enabled=false  # Same as --minimum"
    echo "  $0 kweaver-core install --set image.registry=my-registry.com --set image.tag=v1.0.0  # Custom image settings"
    echo "  $0 kweaver-core download --charts_dir=/path/to/charts # Download Core charts into a specific local directory"
    echo "  $0 kweaver-core install --charts_dir=/path/to/charts  # Install Core from a local charts directory"
    echo "  $0 kweaver-core download --version=0.4.0  # Auto-uses ./release-manifests/0.4.0/kweaver-core.yaml when present"
    echo "  $0 kweaver-core download --version=0.4.0 --version_file=./release-manifests/0.4.0/kweaver-core.yaml"
    echo "  $0 kweaver-core install --config=/root/.kweaver-ai/config.yaml --helm_repo_name=kweaver"
    echo "  $0 isf download --charts_dir=/path/to/charts         # Download ISF charts into a specific local directory"
    echo "  $0 isf install --charts_dir=/path/to/charts          # Install ISF from a local charts directory; auto-installs K8s/data services if absent"
    echo "  $0 isf install --config=/root/.kweaver-ai/config.yaml --helm_repo_name=kweaver"
    echo "  $0 isf download --force-refresh              # Force re-download ISF charts to deploy/.tmp/charts"
    echo "  $0 kweaver-dip install --config=/root/.kweaver-ai/config.yaml"
    echo "  $0 etrino install --config=/root/.kweaver-ai/config.yaml"
}

_detect_node_ip() {
    local node_ip
    node_ip="$(hostname -I 2>/dev/null | tr ' ' '\n' | grep -v '^127\.' | head -1 | tr -d '\n' || true)"
    if [[ -z "${node_ip}" ]] || [[ "${node_ip}" == "127.0.0.1" ]]; then
        node_ip="$(ip addr show 2>/dev/null | grep -oE 'inet [0-9]+(\.[0-9]+){3}' | awk '{print $2}' | grep -v '^127\.' | head -1 || true)"
    fi
    if [[ -z "${node_ip}" ]]; then
        node_ip="10.x.x.x"
    fi
    echo "${node_ip}"
}

_read_access_address_field() {
    local field="$1"
    if [[ ! -f "${CONFIG_YAML_PATH}" ]]; then
        return 0
    fi
    awk -v key="${field}:" '
        $1=="accessAddress:" {in_block=1; next}
        in_block && $1==key {print $2; exit}
        in_block && $0 ~ /^[^ ]/ {in_block=0}
    ' "${CONFIG_YAML_PATH}" 2>/dev/null | sed -e 's/^"//; s/"$//' -e "s/^'//; s/'$//"
}

_upsert_access_address() {
    local host="$1"
    local port="$2"
    local path="$3"
    local scheme="$4"
    local tmp
    local src

    mkdir -p "$(dirname "${CONFIG_YAML_PATH}")"
    tmp="$(mktemp)"
    src="${CONFIG_YAML_PATH}"
    if [[ ! -f "${src}" ]]; then
        src="/dev/null"
    fi

    awk -v host="${host}" -v port="${port}" -v path="${path}" -v scheme="${scheme}" '
        BEGIN {in_block=0; replaced=0}
        {
            if ($1=="accessAddress:") {
                print "accessAddress:"
                print "  host: " host
                print "  port: " port
                print "  scheme: " scheme
                print "  path: " path
                in_block=1
                replaced=1
                next
            }

            if (in_block==1) {
                if ($0 ~ /^[^ ]/) {
                    in_block=0
                    print $0
                }
                next
            }

            print $0
        }
        END {
            if (replaced==0) {
                print "accessAddress:"
                print "  host: " host
                print "  port: " port
                print "  scheme: " scheme
                print "  path: " path
            }
        }
    ' "${src}" > "${tmp}"

    mv "${tmp}" "${CONFIG_YAML_PATH}"
}

confirm_access_address_before_install() {
    local confirm_switch="${CONFIRM_ACCESS_ADDRESS:-true}"
    local config_missing_before="false"
    if [[ ! -f "${CONFIG_YAML_PATH}" ]]; then
        config_missing_before="true"
    fi
    if [[ "${confirm_switch}" == "false" ]]; then
        return 0
    fi

    local raw_host raw_port raw_path raw_scheme
    raw_host="$(_read_access_address_field "host")"
    raw_port="$(_read_access_address_field "port")"
    raw_path="$(_read_access_address_field "path")"
    raw_scheme="$(_read_access_address_field "scheme")"

    local host port path scheme

    # --access_address supports: "host", "host:port", or "scheme://host:port/path"
    if [[ -n "${KWEAVER_ACCESS_ADDRESS:-}" ]]; then
        local addr="${KWEAVER_ACCESS_ADDRESS}"
        if [[ "${addr}" == *"://"* ]]; then
            scheme="${addr%%://*}"
            local remainder="${addr#*://}"
            if [[ "${remainder}" == *":"* ]]; then
                host="${remainder%%:*}"
                remainder="${remainder#*:}"
                port="${remainder%%/*}"
                local after_port="${remainder#*/}"
                if [[ "${after_port}" != "${remainder}" ]]; then
                    path="/${after_port}"
                fi
            else
                host="${remainder%%/*}"
            fi
        elif [[ "${addr}" == *":"* ]]; then
            host="${addr%%:*}"
            port="${addr#*:}"
        else
            host="${addr}"
        fi
        port="${port:-${raw_port:-${INGRESS_NGINX_HTTPS_PORT:-443}}}"
        path="${path:-${raw_path:-/}}"
        scheme="${scheme:-${raw_scheme:-https}}"
    else
        host="${raw_host:-$(_detect_node_ip)}"
        port="${raw_port:-${INGRESS_NGINX_HTTPS_PORT:-443}}"
        path="${raw_path:-/}"
        scheme="${raw_scheme:-https}"
    fi

    local url="${scheme}://${host}:${port}${path}"

    # If provided via CLI arg, skip interactive confirmation
    if [[ -n "${KWEAVER_ACCESS_ADDRESS:-}" ]]; then
        log_info "Using accessAddress from --access_address: ${url}"
        # For first-time initialization, generate full config first.
        if [[ "${config_missing_before}" == "true" ]]; then
            log_info "Config not found, generating: ${CONFIG_YAML_PATH}"
            generate_config_yaml
        fi
        # Then upsert the confirmed accessAddress into full config.
        _upsert_access_address "${host}" "${port}" "${path}" "${scheme}"
        return 0
    fi

    echo ""
    echo "============================================"
    echo "  KWeaver Access Address Configuration"
    echo "============================================"
    echo "  Current detected values:"
    echo "    Host     : ${host}"
    echo "    Port     : ${port}"
    echo "    URL Root : ${path}"
    echo "    Protocol : ${scheme}  (http or https)"
    echo "    URL      : ${url}"
    echo "============================================"

    if [[ "${ASSUME_YES}" == "true" ]]; then
        log_info "Using defaults (-y)."
    elif [[ -t 0 ]]; then
        echo ""
        echo "Press Enter to keep the default, or type a new value."
        echo ""
        local input_host input_port input_path input_scheme
        read -r -p "  Host     [${host}]: " input_host
        read -r -p "  Port     [${port}]: " input_port
        read -r -p "  URL Root [${path}]: " input_path
        read -r -p "  Protocol [${scheme}]: " input_scheme

        host="${input_host:-${host}}"
        port="${input_port:-${port}}"
        path="${input_path:-${path}}"
        scheme="${input_scheme:-${scheme}}"
    else
        log_info "Non-interactive mode detected, using defaults."
    fi

    # For first-time initialization, generate full config first.
    if [[ "${config_missing_before}" == "true" ]]; then
        log_info "Config not found, generating: ${CONFIG_YAML_PATH}"
        generate_config_yaml
    fi

    # Then upsert the confirmed accessAddress into full config.
    _upsert_access_address "${host}" "${port}" "${path}" "${scheme}"
    log_info "accessAddress written to ${CONFIG_YAML_PATH}: ${scheme}://${host}:${port}${path}"
}


# Main function
main() {
    # Parse global flags before module/action
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -y|--yes) ASSUME_YES="true"; shift ;;
            --force-upgrade) FORCE_UPGRADE="true"; shift ;;
            --distro=k3s|--distro=kubeadm)
                export KUBE_DISTRO="${1#*=}"
                shift
                ;;
            --distro)
                export KUBE_DISTRO="$2"
                shift 2
                ;;
            *) break ;;
        esac
    done

    local module="${1:-}"
    local action="${2:-}"
    shift 2 2>/dev/null || true

    # If no arguments, show usage
    if [[ -z "${module}" ]]; then
        usage
        exit 0
    fi

    if [[ "${module}" == "config" ]]; then
        case "${action}" in
            generate)
                check_root
                generate_config_yaml
                ;;
            *)
                log_error "Unknown config action: ${action}"
                usage
                exit 1
                ;;
        esac
        return 0
    fi

    # Handle storage module
    if [[ "${module}" == "storage" ]]; then
        case "${action}" in
            install|init)
                check_root
                install_localpv
                ;;
            *)
                log_error "Unknown storage action: ${action}"
                usage
                exit 1
                ;;
        esac
        return 0
    fi
    
    # Handle k8s module
    if [[ "${module}" == "k8s" ]]; then
        while [[ $# -gt 0 ]]; do
            case "$1" in
                --api_server_address=*) API_SERVER_ADVERTISE_ADDRESS="${1#*=}"; shift ;;
                --api_server_address)   API_SERVER_ADVERTISE_ADDRESS="$2"; shift 2 ;;
                --force-upgrade)        FORCE_UPGRADE="true"; shift ;;
                --access_address=*)     KWEAVER_ACCESS_ADDRESS="${1#*=}"; shift ;;
                --access_address)       KWEAVER_ACCESS_ADDRESS="$2"; shift 2 ;;
                -y|--yes)               ASSUME_YES="true"; shift ;;
                *) shift ;;
            esac
        done
        case "${action}" in
            install|init)
                check_root
                # Pre-install dependencies (containerd, k8s, helm) before k8s init
                log_info "Pre-installing dependencies..."
                detect_package_manager
                install_containerd
                install_kubernetes
                install_helm
                
                check_prerequisites
                init_k8s_master
                allow_master_scheduling
                install_cni
                wait_for_dns

                if [[ "${AUTO_INSTALL_LOCALPV}" == "true" ]]; then
                    if [[ -z "$(kubectl get storageclass --no-headers 2>/dev/null)" ]]; then
                        install_localpv
                    fi
                fi

                if [[ "${AUTO_INSTALL_INGRESS_NGINX}" == "true" ]]; then
                    if ! command -v helm >/dev/null 2>&1; then
                        log_error "Helm is required to install ingress-nginx. Please run: $0 k8s install"
                        exit 1
                    fi
                    install_ingress_nginx
                fi

                if [[ "${AUTO_GENERATE_CONFIG}" == "true" ]]; then
                    generate_config_yaml
                fi
                show_status
                ;;
            reset)
                reset_k8s
                ;;
            status)
                show_status
                ;;
            *)
                log_error "Unknown k8s action: ${action}"
                usage
                exit 1
                ;;
        esac
        return 0
    fi

    # Handle k3s module (single-node k3s on Linux; kubeadm path unchanged in k8s module)
    if [[ "${module}" == "k3s" ]]; then
        case "${action}" in
            install|init)
                check_root
                install_helm || exit 1
                install_k3s || exit 1
                if [[ "${AUTO_INSTALL_INGRESS_NGINX}" == "true" ]]; then
                    install_ingress_nginx || exit 1
                fi
                if [[ "${AUTO_GENERATE_CONFIG}" == "true" ]]; then
                    generate_config_yaml || exit 1
                fi
                show_k3s_status
                ;;
            uninstall)
                check_root
                uninstall_k3s
                ;;
            status)
                show_k3s_status
                ;;
            *)
                log_error "Unknown k3s action: ${action}"
                usage
                exit 1
                ;;
        esac
        return 0
    fi
    
    # Handle mariadb module
    if [[ "${module}" == "mariadb" ]]; then
        case "${action}" in
            install|init)
                check_root
                install_mariadb
                ;;
            uninstall)
                check_root
                shift 2
                uninstall_mariadb "$@"
                ;;
            *)
                log_error "Unknown mariadb action: ${action}"
                usage
                exit 1
                ;;
        esac
        return 0
    fi
    
    # Handle redis module
    if [[ "${module}" == "redis" ]]; then
        case "${action}" in
            install|init)
                check_root
                install_redis
                ;;
            uninstall)
                check_root
                uninstall_redis
                ;;
            *)
                log_error "Unknown redis action: ${action}"
                usage
                exit 1
                ;;
        esac
        return 0
    fi

    # Handle opensearch module
    if [[ "${module}" == "opensearch" ]]; then
        case "${action}" in
            install|init)
                check_root
                install_opensearch
                ;;
            uninstall)
                check_root
                uninstall_opensearch
                ;;
            *)
                log_error "Unknown opensearch action: ${action}"
                usage
                exit 1
                ;;
        esac
        return 0
    fi

    # Handle mongodb module (disabled)
    # if [[ "${module}" == "mongodb" ]]; then
    #     case "${action}" in
    #         install|init)
    #             check_root
    #             # install_mongodb  # MongoDB disabled
    #             ;;
    #         uninstall)
    #             check_root
    #             # uninstall_mongodb  # MongoDB disabled
    #             ;;
    #         *)
    #             log_error "Unknown mongodb action: ${action}"
    #             usage
    #             exit 1
    #             ;;
    #     esac
    #     return 0
    # fi

    # Handle zookeeper module
    if [[ "${module}" == "zookeeper" ]]; then
        case "${action}" in
            install|init)
                check_root
                install_zookeeper
                ;;
            uninstall)
                check_root
                uninstall_zookeeper
                ;;
            *)
                log_error "Unknown zookeeper action: ${action}"
                usage
                exit 1
                ;;
        esac
        return 0
    fi

    # Handle kafka module
    if [[ "${module}" == "kafka" ]]; then
        case "${action}" in
            install|init)
                check_root
                install_kafka
                ;;
            uninstall)
                check_root
                uninstall_kafka
                ;;
            *)
                log_error "Unknown kafka action: ${action}"
                usage
                exit 1
                ;;
        esac
        return 0
    fi
    
    # Handle ingress-nginx module
    if [[ "${module}" == "ingress-nginx" ]]; then
        case "${action}" in
            install|init)
                check_root
                install_ingress_nginx
                ;;
            uninstall)
                check_root
                uninstall_ingress_nginx
                ;;
            *)
                log_error "Unknown ingress-nginx action: ${action}"
                usage
                exit 1
                ;;
        esac
        return 0
    fi
    
    # Handle kweaver-core module
    if [[ "${module}" == "kweaver-core" ]] || [[ "${module}" == "core" ]]; then
        case "${action}" in
            install|init)
                parse_core_args "install" "$@"
                confirm_access_address_before_install
                install_core
                ;;
            download)
                parse_core_args "download" "$@"
                download_core
                ;;
            uninstall)
                parse_core_args "uninstall" "$@"
                uninstall_core
                ;;
            status)
                parse_core_args "status" "$@"
                show_core_status
                ;;
            *)
                log_error "Unknown kweaver-core action: ${action}"
                usage
                exit 1
                ;;
        esac
        return 0
    fi
    
    # Handle kweaver-dip module
    if [[ "${module}" == "kweaver-dip" ]] || [[ "${module}" == "dip" ]]; then
        case "${action}" in
            install|init)
                check_root
                parse_dip_args "install" "$@"
                confirm_access_address_before_install
                install_dip
                ;;
            download)
                parse_dip_args "download" "$@"
                download_dip
                ;;
            uninstall)
                check_root
                parse_dip_args "uninstall" "$@"
                uninstall_dip
                ;;
            status)
                parse_dip_args "status" "$@"
                show_dip_status
                ;;
            *)
                log_error "Unknown kweaver-dip action: ${action}"
                usage
                exit 1
                ;;
        esac
        return 0
    fi

    # Handle etrino module
    if [[ "${module}" == "etrino" ]]; then
        local etrino_script="${SCRIPT_DIR}/scripts/services/etrino.sh"
        if [[ ! -f "${etrino_script}" ]]; then
            log_error "Etrino script not found at ${etrino_script}"
            exit 1
        fi

        case "${action}" in
            install|init)
                check_root
                CONFIG_FILE="${CONFIG_YAML_PATH}" bash "${etrino_script}" install "$@"
                ;;
            status)
                CONFIG_FILE="${CONFIG_YAML_PATH}" bash "${etrino_script}" status "$@"
                ;;
            uninstall)
                check_root
                CONFIG_FILE="${CONFIG_YAML_PATH}" bash "${etrino_script}" uninstall "$@"
                ;;
            *)
                log_error "Unknown etrino action: ${action}"
                usage
                exit 1
                ;;
        esac
        return 0
    fi

    # Handle isf module
    if [[ "${module}" == "isf" ]]; then
        case "${action}" in
            install|init)
                parse_isf_args "install" "$@"
                install_isf
                ;;
            download)
                parse_isf_args "download" "$@"
                download_isf
                ;;
            uninstall)
                parse_isf_args "uninstall" "$@"
                uninstall_isf
                ;;
            status)
                parse_isf_args "status" "$@"
                show_isf_status
                ;;
            *)
                log_error "Unknown isf action: ${action}"
                usage
                exit 1
                ;;
        esac
        return 0
    fi
    
    # Handle all/infra module (infrastructure: k8s + data services)
    # 'all' is an alias for 'infra' for backward compatibility
    if [[ "${module}" == "all" ]] || [[ "${module}" == "infra" ]]; then
        case "${action}" in
            install|init)
                check_root
                log_info "=========================================="
                log_info "  Deploying Infrastructure (K8s + Data Services)"
                log_info "=========================================="
                
                # Pre-install dependencies (containerd, k8s, helm) before k8s init
                log_info "Pre-installing dependencies..."
                detect_package_manager
                install_containerd
                install_kubernetes
                install_helm
                
                check_prerequisites
                init_k8s_master
                allow_master_scheduling
                install_cni
                wait_for_dns

                if [[ "${AUTO_INSTALL_LOCALPV}" == "true" ]]; then
                    if [[ -z "$(kubectl get storageclass --no-headers 2>/dev/null)" ]]; then
                        install_localpv
                    fi
                fi
                install_mariadb
                install_redis
                install_kafka
                install_zookeeper
                # install_mongodb  # MongoDB disabled
                if [[ "${AUTO_INSTALL_INGRESS_NGINX}" == "true" ]]; then
                    install_ingress_nginx
                fi
                install_opensearch
                if [[ "${AUTO_GENERATE_CONFIG}" == "true" ]]; then
                    generate_config_yaml
                fi
                show_status
                log_info "Infrastructure deployment completed!"
                ;;
            reset)
                check_root
                log_info "Resetting infrastructure..."
                uninstall_opensearch || true
                uninstall_ingress_nginx || true
                # uninstall_mongodb || true  # MongoDB disabled
                uninstall_zookeeper || true
                uninstall_kafka || true
                uninstall_redis || true
                uninstall_mariadb || true
                reset_k8s
                log_info "Infrastructure reset completed!"
                ;;
            *)
                log_error "Unknown infra action: ${action}"
                usage
                exit 1
                ;;
        esac
        return 0
    fi
    
    # Handle kweaver module (application services)
    if [[ "${module}" == "kweaver" ]]; then
        case "${action}" in
            init)
                check_root
                shift 2
                log_info "=========================================="
                log_info "  Deploying KWeaver Application Services"
                log_info "=========================================="
                
                # Parse common args for all kweaver services
                while [[ $# -gt 0 ]]; do
                    case "$1" in
                        --version=*)
                            HELM_CHART_VERSION="${1#*=}"
                            shift
                            ;;
                        --version)
                            HELM_CHART_VERSION="$2"
                            shift 2
                            ;;
                        --helm_repo=*)
                            HELM_CHART_REPO_URL="${1#*=}"
                            shift
                            ;;
                        --helm_repo)
                            HELM_CHART_REPO_URL="$2"
                            shift 2
                            ;;
                        *)
                            shift
                            ;;
                    esac
                done
                
                # Install all KWeaver services in order
                install_isf
                install_core

                log_info "KWeaver application services deployment completed!"
                ;;
            uninstall)
                check_root
                log_info "Uninstalling KWeaver application services..."
                uninstall_core || true
                uninstall_isf || true
                log_info "KWeaver application services uninstalled!"
                ;;
            status)
                log_info "KWeaver application services status:"
                show_isf_status
                show_core_status
                show_dip_status
                ;;
            *)
                log_error "Unknown kweaver action: ${action}"
                usage
                exit 1
                ;;
        esac
        return 0
    fi
    
    # Handle full module (complete deployment: infra + kweaver)
    if [[ "${module}" == "full" ]]; then
        case "${action}" in
            init)
                check_root
                shift 2
                log_info "╔════════════════════════════════════════════════════════════════╗"
                log_info "║       Full Deployment: Infrastructure + KWeaver Services       ║"
                log_info "╚════════════════════════════════════════════════════════════════╝"
                
                # Save args for kweaver
                local kweaver_args=("$@")
                
                # Step 1: Deploy infrastructure
                log_info ""
                log_info "Step 1/2: Deploying Infrastructure..."
                log_info ""
                
                detect_package_manager
                install_containerd
                install_kubernetes
                install_helm
                
                check_prerequisites
                init_k8s_master
                allow_master_scheduling
                install_cni
                wait_for_dns

                if [[ "${AUTO_INSTALL_LOCALPV}" == "true" ]]; then
                    if [[ -z "$(kubectl get storageclass --no-headers 2>/dev/null)" ]]; then
                        install_localpv
                    fi
                fi
                install_mariadb
                install_redis
                install_kafka
                install_zookeeper
                # install_mongodb  # MongoDB disabled
                if [[ "${AUTO_INSTALL_INGRESS_NGINX}" == "true" ]]; then
                    install_ingress_nginx
                fi
                install_opensearch
                if [[ "${AUTO_GENERATE_CONFIG}" == "true" ]]; then
                    generate_config_yaml
                fi
                
                # Step 2: Deploy KWeaver services
                log_info ""
                log_info "Step 2/2: Deploying KWeaver Application Services..."
                log_info ""
                
                # Parse kweaver args
                for arg in "${kweaver_args[@]}"; do
                    case "$arg" in
                        --version=*)
                            HELM_CHART_VERSION="${arg#*=}"
                            ;;
                        --helm_repo=*)
                            HELM_CHART_REPO_URL="${arg#*=}"
                            ;;
                    esac
                done
                
                install_isf
                install_studio
                install_bkn
                install_vega
                install_agentoperator
                install_decisionagent
                install_flowautomation
                install_sandboxruntime

                show_status
                log_info ""
                log_info "╔════════════════════════════════════════════════════════════════╗"
                log_info "║                   Full Deployment Completed!                   ║"
                log_info "╚════════════════════════════════════════════════════════════════╝"
                ;;
            reset)
                check_root
                log_info "Full reset: Uninstalling all components..."
                
                # Uninstall KWeaver services first
                uninstall_sandboxruntime || true
                uninstall_flowautomation || true
                uninstall_decisionagent || true
                uninstall_agentoperator || true
                uninstall_bkn || true
                uninstall_vega || true
                uninstall_studio || true
                uninstall_isf || true
                
                # Then uninstall infrastructure
                uninstall_opensearch || true
                uninstall_ingress_nginx || true
                # uninstall_mongodb || true  # MongoDB disabled
                uninstall_zookeeper || true
                uninstall_kafka || true
                uninstall_redis || true
                uninstall_mariadb || true
                reset_k8s
                
                log_info "Full reset completed!"
                ;;
            *)
                log_error "Unknown full action: ${action}"
                usage
                exit 1
                ;;
        esac
        return 0
    fi
    
    # Unknown module
    usage
    exit 1
}

main "$@"
