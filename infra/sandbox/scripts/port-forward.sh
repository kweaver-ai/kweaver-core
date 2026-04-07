#!/bin/bash

################################################################################
# K8s Port-Forwarding Script for Control Plane, Web Console and MinIO
################################################################################
# This script manages port-forwarding for sandbox services in Kubernetes
#
# Services:
#   - control-plane: FastAPI REST API (8000)
#   - sandbox-web: Web Console (1101)
#   - minio-api: MinIO S3 API (9000)
#   - minio-console: MinIO Web Console (9001)
################################################################################

set -eo pipefail

# Default configuration
NAMESPACE="sandbox-system"
PID_DIR="${TMPDIR:-/tmp}/sandbox-port-forward"
LOG_DIR="${PID_DIR}/logs"

# Service definitions (using arrays instead of associative arrays for bash 3.x compatibility)
# Format: "service_name:local_port:remote_port:resource_type:resource_name:display_name"
CONTROL_PLANE_DEF="control-plane:8000:8000:service/sandbox-control-plane:Control Plane API"
SANDBOX_WEB_DEF="sandbox-web:1101:80:service/sandbox-web:Web Console"
MINIO_API_DEF="minio-api:9000:9000:service/minio:MinIO S3 API"
MINIO_CONSOLE_DEF="minio-console:9001:9001:service/minio:MinIO Web Console"

ALL_SERVICES=("$CONTROL_PLANE_DEF" "$SANDBOX_WEB_DEF" "$MINIO_API_DEF" "$MINIO_CONSOLE_DEF")

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

################################################################################
# Helper Functions
################################################################################

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Create necessary directories
init_dirs() {
    mkdir -p "$PID_DIR"
    mkdir -p "$LOG_DIR"
}

# Check if kubectl is installed
check_kubectl() {
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed. Please install kubectl first."
        exit 1
    fi
}

# Check if cluster is accessible
check_cluster() {
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster. Please check your kubeconfig."
        exit 1
    fi
}

# Get service definition by name
get_service_def() {
    local service_name=$1
    for def in "${ALL_SERVICES[@]}"; do
        local name=$(echo "$def" | cut -d':' -f1)
        if [[ "$name" == "$service_name" ]]; then
            echo "$def"
            return 0
        fi
    done
    return 1
}

# Get specific field from service definition
get_service_field() {
    local service_name=$1
    local field_num=$2
    local def=$(get_service_def "$service_name")
    if [[ -n "$def" ]]; then
        echo "$def" | cut -d':' -f"$field_num"
    fi
}

# Check if service exists in namespace
check_service() {
    local service=$1
    local resource=$(get_service_field "$service" 4)

    if ! kubectl get -n "$NAMESPACE" "$resource" &> /dev/null; then
        log_error "Service '$resource' not found in namespace '$NAMESPACE'"
        return 1
    fi
    return 0
}

# Check if port is already in use
check_port() {
    local port=$1

    # Check using lsof or netstat depending on availability
    if command -v lsof &> /dev/null; then
        if lsof -Pi ":$port" -sTCP:LISTEN -t >/dev/null 2>&1; then
            return 0
        fi
    elif command -v netstat &> /dev/null; then
        if netstat -an 2>/dev/null | grep -q "\.$port.*LISTEN"; then
            return 0
        fi
    fi

    return 1
}

# Get the PID of a process using a specific port
get_pid_on_port() {
    local port=$1

    if command -v lsof &> /dev/null; then
        lsof -ti ":$port" 2>/dev/null || echo ""
    else
        # Fallback using ps and grep
        ps aux | grep "port-forward.*:${port}" | grep -v grep | awk '{print $2}' | head -1
    fi
}

################################################################################
# Port Forwarding Functions
################################################################################

# Start port forwarding for a service
start_service() {
    local service=$1
    local custom_local_port=${2:-}
    local background=${3:-false}

    local display_name=$(get_service_field "$service" 5)
    local default_local=$(get_service_field "$service" 2)
    local remote_port=$(get_service_field "$service" 3)
    local resource=$(get_service_field "$service" 4)
    local local_port=${custom_local_port:-$default_local}

    # Check if service exists
    if ! check_service "$service"; then
        return 1
    fi

    # Check if already running
    local pid_file="$PID_DIR/${service}.pid"
    if [[ -f "$pid_file" ]]; then
        local pid=$(cat "$pid_file")
        if kill -0 "$pid" 2>/dev/null; then
            log_warning "$service is already running (PID: $pid, Port: $local_port)"
            return 0
        else
            # Clean up stale PID file
            rm -f "$pid_file"
        fi
    fi

    # Check if port is in use
    if check_port "$local_port"; then
        local existing_pid=$(get_pid_on_port "$local_port")
        log_error "Port $local_port is already in use"
        if [[ -n "$existing_pid" ]]; then
            log_error "  Process $existing_pid is using this port"
            log_error "  Run: kill $existing_pid"
        fi
        return 1
    fi

    # Start port forwarding
    local log_file="$LOG_DIR/${service}.log"

    if [[ "$background" == "true" ]]; then
        log_info "Starting $display_name in background..."
        kubectl port-forward -n "$NAMESPACE" "$resource" --address 0.0.0.0 "$local_port:$remote_port" > "$log_file" 2>&1 &
        local pid=$!
        echo $pid > "$pid_file"

        # Wait a moment to ensure it starts successfully
        sleep 1
        if kill -0 "$pid" 2>/dev/null; then
            log_success "Started $display_name (PID: $pid, Port: $local_port)"
            log_info "  Logs: $log_file"
        else
            log_error "Failed to start $display_name"
            rm -f "$pid_file"
            return 1
        fi
    else
        log_info "Starting $display_name on port $local_port..."
        log_info "Press Ctrl+C to stop"
        kubectl port-forward -n "$NAMESPACE" "$resource" --address 0.0.0.0 "$local_port:$remote_port"
    fi

    return 0
}

# Stop port forwarding for a service
stop_service() {
    local service=$1
    local display_name=$(get_service_field "$service" 5)
    local default_local=$(get_service_field "$service" 2)
    local pid_file="$PID_DIR/${service}.pid"

    if [[ ! -f "$pid_file" ]]; then
        # Check if port is still in use and try to kill by port
        local pid=$(get_pid_on_port "$default_local")
        if [[ -n "$pid" ]]; then
            log_info "Found orphaned process on port $default_local (PID: $pid)"
            kill "$pid" 2>/dev/null || true
            log_success "Stopped orphaned $display_name process"
        else
            log_warning "$display_name is not running"
        fi
        return 0
    fi

    local pid=$(cat "$pid_file")

    if kill -0 "$pid" 2>/dev/null; then
        kill "$pid"
        # Wait for process to terminate
        local count=0
        while kill -0 "$pid" 2>/dev/null && [[ $count -lt 5 ]]; do
            sleep 1
            ((count++))
        done

        # Force kill if still running
        if kill -0 "$pid" 2>/dev/null; then
            kill -9 "$pid" 2>/dev/null || true
        fi

        log_success "Stopped $display_name (PID: $pid)"
    else
        log_warning "$display_name was not running (stale PID file)"
    fi

    rm -f "$pid_file"
    return 0
}

# Show status of all services
show_status() {
    echo ""
    echo "=== Port Forwarding Status ==="
    echo "Namespace: $NAMESPACE"
    echo ""

    for def in "${ALL_SERVICES[@]}"; do
        local service=$(echo "$def" | cut -d':' -f1)
        local display_name=$(echo "$def" | cut -d':' -f5)
        local local_port=$(echo "$def" | cut -d':' -f2)
        local remote_port=$(echo "$def" | cut -d':' -f3)
        local resource=$(echo "$def" | cut -d':' -f4)

        local pid_file="$PID_DIR/${service}.pid"
        local status=""
        local pid_info=""

        if [[ -f "$pid_file" ]]; then
            local pid=$(cat "$pid_file")
            if kill -0 "$pid" 2>/dev/null; then
                status="✓"
                pid_info="PID: $pid"
            else
                status="✗"
                pid_info="Stopped (stale PID file)"
            fi
        else
            # Check if port is in use by another process
            if check_port "$local_port"; then
                local orphan_pid=$(get_pid_on_port "$local_port")
                status="?"
                if [[ -n "$orphan_pid" ]]; then
                    pid_info="Port in use by PID: $orphan_pid (not managed by this script)"
                else
                    pid_info="Port in use (unknown process)"
                fi
            else
                status="-"
                pid_info="Not running"
            fi
        fi

        printf "%-25s %-3s localhost:%-5s → %-35s %s\n" \
            "$display_name" \
            "$status" \
            "$local_port" \
            "$resource:$remote_port" \
            "$pid_info"
    done

    echo ""
}

# Show logs for a service
show_logs() {
    local service=$1
    local log_file="$LOG_DIR/${service}.log"
    local display_name=$(get_service_field "$service" 5)

    if [[ ! -f "$log_file" ]]; then
        log_error "No log file found for $display_name"
        return 1
    fi

    echo "=== Logs for $display_name ==="
    tail -f "$log_file"
}

# Clean up orphaned processes
cleanup_orphans() {
    log_info "Checking for orphaned port-forward processes..."

    for def in "${ALL_SERVICES[@]}"; do
        local service=$(echo "$def" | cut -d':' -f1)
        local display_name=$(echo "$def" | cut -d':' -f5)
        local local_port=$(echo "$def" | cut -d':' -f2)
        local pid_file="$PID_DIR/${service}.pid"
        local orphan_pid=$(get_pid_on_port "$local_port")

        if [[ -n "$orphan_pid" ]]; then
            if [[ -f "$pid_file" ]]; then
                local registered_pid=$(cat "$pid_file")
                if [[ "$orphan_pid" != "$registered_pid" ]]; then
                    log_warning "Found orphaned process for $display_name on port $local_port (PID: $orphan_pid)"
                    if [[ "$AUTO_CONFIRM" == "true" ]]; then
                        kill "$orphan_pid" 2>/dev/null || true
                        log_success "Killed orphaned process"
                    else
                        read -p "Kill this process? [y/N] " -n 1 -r
                        echo
                        if [[ $REPLY =~ ^[Yy]$ ]]; then
                            kill "$orphan_pid" 2>/dev/null || true
                            log_success "Killed orphaned process"
                        fi
                    fi
                fi
            else
                log_warning "Found unmanaged process for $display_name on port $local_port (PID: $orphan_pid)"
            fi
        fi
    done
}

################################################################################
# Usage and Help
################################################################################

usage() {
    cat << EOF
Usage: $0 [OPTIONS] COMMAND

Commands:
    start           Start port forwarding for specified services
    stop            Stop port forwarding for specified services
    restart         Restart specified services
    status          Show status of all services
    logs            Show logs for a service (follow mode)
    cleanup         Clean up orphaned processes

Services:
    --control-plane      Control Plane API (port 8000)
    --sandbox-web        Web Console (port 1101)
    --minio-api         MinIO S3 API (port 9000)
    --minio-console     MinIO Web Console (port 9001)
    --all               All services (default)

Options:
    -b, --background    Run in background mode
    -p, --port PORT     Custom local port (format: local_port)
    -n, --namespace NS  Specify namespace (default: sandbox-system)
    -y, --yes           Auto-confirm prompts (for cleanup)
    -h, --help          Show this help message

Examples:
    # Start all services in background
    $0 start --all --background

    # Start only Control Plane and Web Console
    $0 start --control-plane --sandbox-web -b

    # Start with custom local port
    $0 start --sandbox-web --port 11101 -b

    # Check status
    $0 status

    # Stop all services
    $0 stop --all

    # View logs for a service
    $0 logs sandbox-web

    # Clean up orphaned processes
    $0 cleanup -y
EOF
}

################################################################################
# Main Script Logic
################################################################################

# Initialize
init_dirs
check_kubectl
check_cluster

# Parse arguments
COMMAND=""
SERVICES=()
BACKGROUND=false
CUSTOM_PORT=""
AUTO_CONFIRM=false

while [[ $# -gt 0 ]]; do
    case $1 in
        start|stop|restart|status|logs|cleanup)
            COMMAND="$1"
            shift
            ;;
        --control-plane|--sandbox-web|--minio-api|--minio-console)
            SERVICES+=("${1#--}")
            shift
            ;;
        --all)
            SERVICES=("control-plane" "sandbox-web" "minio-api" "minio-console")
            shift
            ;;
        -b|--background)
            BACKGROUND=true
            shift
            ;;
        -p|--port)
            CUSTOM_PORT="$2"
            shift 2
            ;;
        -n|--namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        -y|--yes)
            AUTO_CONFIRM=true
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Default to all services if none specified
if [[ ${#SERVICES[@]} -eq 0 && "$COMMAND" != "status" && "$COMMAND" != "cleanup" ]]; then
    SERVICES=("control-plane" "sandbox-web" "minio-api" "minio-console")
fi

# Execute command
case "$COMMAND" in
    start)
        for service in "${SERVICES[@]}"; do
            start_service "$service" "$CUSTOM_PORT" "$BACKGROUND"
        done
        ;;
    stop)
        for service in "${SERVICES[@]}"; do
            stop_service "$service"
        done
        ;;
    restart)
        for service in "${SERVICES[@]}"; do
            stop_service "$service"
            sleep 1
            start_service "$service" "$CUSTOM_PORT" "$BACKGROUND"
        done
        ;;
    status)
        show_status
        ;;
    logs)
        if [[ ${#SERVICES[@]} -eq 0 ]]; then
            log_error "Please specify a service to view logs"
            exit 1
        fi
        show_logs "${SERVICES[0]}"
        ;;
    cleanup)
        cleanup_orphans
        ;;
    *)
        log_error "No command specified"
        usage
        exit 1
        ;;
esac
