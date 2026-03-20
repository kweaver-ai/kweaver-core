#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONF_DIR="${CONF_DIR:-${SCRIPT_DIR}/conf}"
CONFIG_YAML_PATH="${CONFIG_YAML_PATH:-${CONF_DIR}/config.yaml}"

# Source all service libraries
source "${SCRIPT_DIR}/scripts/lib/common.sh"
source "${SCRIPT_DIR}/scripts/services/config.sh"
source "${SCRIPT_DIR}/scripts/services/k8s.sh"
source "${SCRIPT_DIR}/scripts/services/storage.sh"
source "${SCRIPT_DIR}/scripts/services/mariadb.sh"
source "${SCRIPT_DIR}/scripts/services/redis.sh"
source "${SCRIPT_DIR}/scripts/services/kafka.sh"
source "${SCRIPT_DIR}/scripts/services/zookeeper.sh"
source "${SCRIPT_DIR}/scripts/services/mongodb.sh"
source "${SCRIPT_DIR}/scripts/services/ingress_nginx.sh"
source "${SCRIPT_DIR}/scripts/services/opensearch.sh"
source "${SCRIPT_DIR}/scripts/services/studio.sh"
source "${SCRIPT_DIR}/scripts/services/ontology.sh"
source "${SCRIPT_DIR}/scripts/services/agentoperator.sh"
source "${SCRIPT_DIR}/scripts/services/decisionagent.sh"
source "${SCRIPT_DIR}/scripts/services/flowautomation.sh"
source "${SCRIPT_DIR}/scripts/services/sandboxruntime.sh"
source "${SCRIPT_DIR}/scripts/services/isf.sh"

usage() {
    echo "Kubernetes Infrastructure Initialization Script"
    echo ""
    echo "Usage: $0 <module> [action]"
    echo ""
    echo "Modules and Actions:"
    echo "  k8s init                      Initialize K8s master node with CNI and DNS"
    echo "  k8s reset                     Reset Kubernetes cluster state (kubeadm reset -f + cleanup)"
    echo "  k8s status                    Show cluster status"
    echo "  mariadb init                  Install single-node MariaDB 11"
    echo "  mariadb uninstall             Uninstall MariaDB (optionally purge PVC)"
    echo "  redis init                    Install single-node Redis 7"
    echo "  redis uninstall               Uninstall Redis (PVCs will be deleted by default)"
    echo "  kafka init                    Install single-node Kafka"
    echo "  kafka uninstall               Uninstall Kafka (PVCs will be deleted by default)"
    echo "  opensearch init               Install single-node OpenSearch"
    echo "  opensearch uninstall          Uninstall OpenSearch (optionally purge PVC)"
    echo "  mongodb init                  Install MongoDB"
    echo "  mongodb uninstall             Uninstall MongoDB (PVCs will be deleted)"
    echo "  zookeeper init                Install single-node Zookeeper"
    echo "  zookeeper uninstall           Uninstall Zookeeper (PVCs will be deleted by default)"
    echo "  ingress-nginx init            Install ingress-nginx-controller"
    echo "  ingress-nginx uninstall       Uninstall ingress-nginx-controller"
    echo "  studio init                   Install Studio services (deploy-web, studio-web, etc.)"
    echo "  studio uninstall              Uninstall Studio services"
    echo "  studio status                 Show Studio services status"
    echo "  ontology init                 Install Ontology services (ontology-manager, vega-web, etc.)"
    echo "  ontology uninstall            Uninstall Ontology services"
    echo "  ontology status               Show Ontology services status"
    echo "  agent_operator init           Install Agent Operator services (agent-operator-app, operator-web, etc.)"
    echo "  agent_operator uninstall      Uninstall Agent Operator services"
    echo "  agent_operator status         Show Agent Operator services status"
    echo "  decisionagent init            Install DecisionAgent services (agent-backend, agent-web, etc.)"
    echo "  decisionagent uninstall       Uninstall DecisionAgent services"
    echo "  decisionagent status          Show DecisionAgent services status"
    echo "  flowautomation init           Install FlowAutomation services (flow-web, flow-automation, etc.)"
    echo "  flowautomation uninstall      Uninstall FlowAutomation services"
    echo "  flowautomation status         Show FlowAutomation services status"
    echo "  sandboxruntime init           Install SandboxRuntime services (sandbox-runtime, etc.)"
    echo "  sandboxruntime uninstall      Uninstall SandboxRuntime services"
    echo "  sandboxruntime status         Show SandboxRuntime services status"
    echo "  isf init                      Install ISF services (informationsecurityfabric, hydra, sharemgnt, etc.)"
    echo "  isf uninstall                 Uninstall ISF services"
    echo "  isf status                    Show ISF services status"
    echo ""
    echo "Examples:"
    echo "  $0 k8s init                   # Initialize K8s master node with default settings"
    echo "  $0 k8s reset                  # Reset cluster state before re-init"
    echo "  $0 k8s status                 # Show cluster status"
    echo "  POD_CIDR=10.0.0.0/16 $0 k8s init  # Initialize with custom POD_CIDR"
    echo "  $0 mariadb init               # Install MariaDB"
    echo "  $0 mariadb uninstall          # Uninstall MariaDB"
    echo "  $0 mariadb uninstall --delete-data  # Uninstall MariaDB and delete PVC (data loss!)"
    echo "  MARIADB_PURGE_PVC=true $0 mariadb uninstall  # Same as --delete-data (data loss!)"
    echo "  $0 redis init                 # Install Redis"
    echo "  $0 redis uninstall            # Uninstall Redis"
    echo "  $0 redis uninstall                         # Uninstall Redis (PVCs deleted by default)"
    echo "  REDIS_PURGE_PVC=false $0 redis uninstall   # Uninstall Redis but keep PVCs"
    echo "  $0 kafka init                 # Install Kafka"
    echo "  $0 kafka uninstall                         # Uninstall Kafka (PVCs deleted by default)"
    echo "  KAFKA_PURGE_PVC=false $0 kafka uninstall   # Uninstall Kafka but keep PVCs"
    echo "  $0 opensearch init            # Install OpenSearch"
    echo "  $0 opensearch uninstall       # Uninstall OpenSearch"
    echo "  OPENSEARCH_PURGE_PVC=true $0 opensearch uninstall  # Uninstall OpenSearch and delete PVC (data loss!)"
    echo "  $0 mongodb init               # Install MongoDB"
    echo "  $0 mongodb uninstall          # Uninstall MongoDB (PVCs will be deleted)"
    echo "  $0 zookeeper init             # Install Zookeeper"
    echo "  $0 zookeeper uninstall        # Uninstall Zookeeper (PVCs deleted by default)"
    echo "  ZOOKEEPER_PURGE_PVC=false $0 zookeeper uninstall  # Uninstall Zookeeper but keep PVCs"
    echo "  # Install from remote repo with version and devel:"
    echo "  ZOOKEEPER_CHART_REF=dip/zookeeper ZOOKEEPER_CHART_VERSION=0.0.0-feature-800792 ZOOKEEPER_CHART_DEVEL=true $0 zookeeper init"
    echo "  # Install with additional values file and --set:"
    echo "  ZOOKEEPER_VALUES_FILE=conf/config.yaml ZOOKEEPER_EXTRA_SET_VALUES='image.registry=swr.cn-east-3.myhuaweicloud.com/kweaver-ai' $0 zookeeper init"
    echo "  $0 ingress-nginx init         # Install ingress-nginx-controller"
    echo "  $0 ingress-nginx uninstall    # Uninstall ingress-nginx-controller"
    echo "  $0 config generate            # Generate/update conf/config.yaml"
    echo "  $0 infra init                 # Initialize infrastructure components"
    echo "  $0 full init                  # Full initialization with all components"
}


# Main function
main() {
    local module="${1:-}"
    local action="${2:-}"
    
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
            init)
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
        case "${action}" in
            init)
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
                        log_error "Helm is required to install ingress-nginx. Please run: $0 k8s init"
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
    
    # Handle mariadb module
    if [[ "${module}" == "mariadb" ]]; then
        case "${action}" in
            init)
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
            init)
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
            init)
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

    # Handle mongodb module
    if [[ "${module}" == "mongodb" ]]; then
        case "${action}" in
            init)
                check_root
                install_mongodb
                ;;
            uninstall)
                check_root
                uninstall_mongodb
                ;;
            *)
                log_error "Unknown mongodb action: ${action}"
                usage
                exit 1
                ;;
        esac
        return 0
    fi

    # Handle zookeeper module
    if [[ "${module}" == "zookeeper" ]]; then
        case "${action}" in
            init)
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
            init)
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
            init)
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
    
    # Handle studio module
    if [[ "${module}" == "studio" ]]; then
        case "${action}" in
            init)
                shift 2
                parse_studio_args "init" "$@"
                install_studio
                ;;
            uninstall)
                shift 2
                parse_studio_args "uninstall" "$@"
                uninstall_studio
                ;;
            status)
                show_studio_status
                ;;
            *)
                log_error "Unknown studio action: ${action}"
                usage
                exit 1
                ;;
        esac
        return 0
    fi
    
    # Handle ontology module
    if [[ "${module}" == "ontology" ]]; then
        case "${action}" in
            init)
                shift 2
                parse_ontology_args "init" "$@"
                install_ontology
                ;;
            uninstall)
                shift 2
                parse_ontology_args "uninstall" "$@"
                uninstall_ontology
                ;;
            status)
                show_ontology_status
                ;;
            *)
                log_error "Unknown ontology action: ${action}"
                usage
                exit 1
                ;;
        esac
        return 0
    fi
    
    # Handle agent_operator module (supports both agent_operator and agentoperator)
    if [[ "${module}" == "agent_operator" ]] || [[ "${module}" == "agentoperator" ]]; then
        case "${action}" in
            init)
                shift 2
                parse_agentoperator_args "init" "$@"
                install_agentoperator
                ;;
            uninstall)
                shift 2
                parse_agentoperator_args "uninstall" "$@"
                uninstall_agentoperator
                ;;
            status)
                show_agentoperator_status
                ;;
            *)
                log_error "Unknown agentoperator action: ${action}"
                usage
                exit 1
                ;;
        esac
        return 0
    fi
    
    # Handle decisionagent module
    if [[ "${module}" == "decisionagent" ]]; then
        case "${action}" in
            init)
                shift 2
                parse_decisionagent_args "init" "$@"
                install_decisionagent
                ;;
            uninstall)
                shift 2
                parse_decisionagent_args "uninstall" "$@"
                uninstall_decisionagent
                ;;
            status)
                show_decisionagent_status
                ;;
            *)
                log_error "Unknown decisionagent action: ${action}"
                usage
                exit 1
                ;;
        esac
        return 0
    fi
    
    # Handle flowautomation module
    if [[ "${module}" == "flowautomation" ]]; then
        case "${action}" in
            init)
                shift 2
                parse_flowautomation_args "init" "$@"
                install_flowautomation
                ;;
            uninstall)
                shift 2
                parse_flowautomation_args "uninstall" "$@"
                uninstall_flowautomation
                ;;
            status)
                show_flowautomation_status
                ;;
            *)
                log_error "Unknown flowautomation action: ${action}"
                usage
                exit 1
                ;;
        esac
        return 0
    fi
    
    # Handle sandboxruntime module
    if [[ "${module}" == "sandboxruntime" ]]; then
        case "${action}" in
            init)
                shift 2
                parse_sandboxruntime_args "init" "$@"
                install_sandboxruntime
                ;;
            uninstall)
                shift 2
                parse_sandboxruntime_args "uninstall" "$@"
                uninstall_sandboxruntime
                ;;
            status)
                show_sandboxruntime_status
                ;;
            *)
                log_error "Unknown sandboxruntime action: ${action}"
                usage
                exit 1
                ;;
        esac
        return 0
    fi
    
    # Handle isf module
    if [[ "${module}" == "isf" ]]; then
        case "${action}" in
            init)
                shift 2
                parse_isf_args "init" "$@"
                install_isf
                ;;
            uninstall)
                shift 2
                parse_isf_args "uninstall" "$@"
                uninstall_isf
                ;;
            status)
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
    
    # Handle infra module (infrastructure: k8s + data services)
    if [[ "${module}" == "infra" ]]; then
        case "${action}" in
            init)
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
                install_mongodb
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
                uninstall_mongodb || true
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
                
                # Validate and auto-repair missing credentials before installing services
                validate_config_credentials

                # Install all KWeaver services in order
                install_isf
                install_studio
                install_ontology
                install_agentoperator
                install_decisionagent
                install_flowautomation
                install_sandboxruntime

                log_info "KWeaver application services deployment completed!"
                print_web_access_info
                ;;
            uninstall)
                check_root
                log_info "Uninstalling KWeaver application services..."
                uninstall_sandboxruntime || true
                uninstall_flowautomation || true
                uninstall_decisionagent || true
                uninstall_agentoperator || true
                uninstall_ontology || true
                uninstall_studio || true
                uninstall_isf || true
                log_info "KWeaver application services uninstalled!"
                ;;
            status)
                log_info "KWeaver application services status:"
                show_isf_status
                show_studio_status
                show_ontology_status
                show_agentoperator_status
                show_decisionagent_status
                show_flowautomation_status
                show_sandboxruntime_status
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
                install_mongodb
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

                # Validate and auto-repair missing credentials before installing services
                validate_config_credentials
                
                install_isf
                install_studio
                install_ontology
                install_agentoperator
                install_decisionagent
                install_flowautomation
                install_sandboxruntime

                show_status
                log_info ""
                log_info "╔════════════════════════════════════════════════════════════════╗"
                log_info "║                   Full Deployment Completed!                   ║"
                log_info "╚════════════════════════════════════════════════════════════════╝"
                print_web_access_info
                ;;
            reset)
                check_root
                log_info "Full reset: Uninstalling all components..."
                
                # Uninstall KWeaver services first
                uninstall_sandboxruntime || true
                uninstall_flowautomation || true
                uninstall_decisionagent || true
                uninstall_agentoperator || true
                uninstall_ontology || true
                uninstall_studio || true
                uninstall_isf || true
                
                # Then uninstall infrastructure
                uninstall_opensearch || true
                uninstall_ingress_nginx || true
                uninstall_mongodb || true
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
