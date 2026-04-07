#!/bin/bash
# Build script for sandbox images

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Configuration
BASE_IMAGE_NAME="${BASE_IMAGE_NAME:-sandbox-executor-base}"
BASE_IMAGE_TAG="${BASE_IMAGE_TAG:-latest}"
REGISTRY="${REGISTRY:-localhost:5000}"
PUSH="${PUSH:-false}"
USE_MIRROR="${USE_MIRROR:-false}"

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

# Build base image
build_base() {
    log_info "Building executor base image: ${BASE_IMAGE_NAME}:${BASE_IMAGE_TAG}"

    local build_args=""
    if [ "$USE_MIRROR" = "true" ]; then
        build_args="--build-arg USE_MIRROR=true"
        log_info "Using mirror sources for base image"
    fi

    docker build \
        -f "${PROJECT_ROOT}/runtime/executor/Dockerfile" \
        -t "${BASE_IMAGE_NAME}:${BASE_IMAGE_TAG}" \
        -t "${BASE_IMAGE_NAME}:latest" \
        $build_args \
        "${PROJECT_ROOT}"

    if [ "$PUSH" = "true" ]; then
        log_info "Pushing base image to registry..."
        docker tag "${BASE_IMAGE_NAME}:${BASE_IMAGE_TAG}" "${REGISTRY}/${BASE_IMAGE_NAME}:${BASE_IMAGE_TAG}"
        docker push "${REGISTRY}/${BASE_IMAGE_NAME}:${BASE_IMAGE_TAG}"
    fi
}

# Build template images
build_templates() {
    # local templates=("python-basic" "python-datascience" "nodejs-basic")
    local templates=("python-basic")

    local build_args=""
    if [ "$USE_MIRROR" = "true" ]; then
        build_args="--build-arg USE_MIRROR=true"
        log_info "Using mirror sources for template images"
    fi

    for template in "${templates[@]}"; do
        log_info "Building template: ${template}"

        local template_dir="${SCRIPT_DIR}/templates/${template}"
        local image_name="sandbox-template-${template}"
        local image_tag="v1.0.0"

        if [ ! -d "$template_dir" ]; then
            log_warn "Template directory not found: $template_dir"
            continue
        fi

        docker build \
            -f "${template_dir}/Dockerfile" \
            --build-arg "BASE_IMAGE=${BASE_IMAGE_NAME}:${BASE_IMAGE_TAG}" \
            $build_args \
            -t "${image_name}:${image_tag}" \
            -t "${image_name}:latest" \
            "${PROJECT_ROOT}"

        if [ "$PUSH" = "true" ]; then
            log_info "Pushing template image to registry..."
            docker tag "${image_name}:${image_tag}" "${REGISTRY}/${image_name}:${image_tag}"
            docker push "${REGISTRY}/${image_name}:${image_tag}"
        fi
    done
}

# Main build flow
main() {
    log_info "Starting sandbox image build..."
    log_info "Project root: ${PROJECT_ROOT}"

    # Build base image first
    build_base

    # Build template images
    build_templates

    log_info "Build complete!"
    log_info "Images:"
    log_info "  - ${BASE_IMAGE_NAME}:${BASE_IMAGE_TAG}"
    log_info "  - sandbox-template-python-basic:latest"
    log_info "  - sandbox-template-python-datascience:latest"
    log_info "  - sandbox-template-nodejs-basic:latest"
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --push)
            PUSH=true
            shift
            ;;
        --registry)
            REGISTRY="$2"
            shift 2
            ;;
        --base-tag)
            BASE_IMAGE_TAG="$2"
            shift 2
            ;;
        --use-mirror)
            USE_MIRROR=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --push        Push images to registry"
            echo "  --registry    Docker registry (default: localhost:5000)"
            echo "  --base-tag    Base image tag (default: latest)"
            echo "  --use-mirror  Use mirror sources for building images"
            echo "  -h, --help    Show this help message"
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Run main
main
