#!/bin/bash
# Build script for sandbox images

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Configuration
VERSION_FILE="${VERSION_FILE:-${PROJECT_ROOT}/VERSION}"
PROJECT_VERSION="${PROJECT_VERSION:-$(tr -d '[:space:]' < "$VERSION_FILE")}"
PYTHON_BASE_IMAGE_NAME="${PYTHON_BASE_IMAGE_NAME:-sandbox-python-executor-base}"
PYTHON_BASE_IMAGE_TAG="${PYTHON_BASE_IMAGE_TAG:-python3.11-v1}"
MULTI_BASE_IMAGE_NAME="${MULTI_BASE_IMAGE_NAME:-sandbox-multi-executor-base}"
MULTI_BASE_IMAGE_TAG="${MULTI_BASE_IMAGE_TAG:-go1.25-python3.11-v1}"
REGISTRY="${REGISTRY:-localhost:5000}"
PUSH="${PUSH:-false}"
USE_MIRROR="${USE_MIRROR:-false}"
DOCKER_BUILD_PLATFORM="${DOCKER_BUILD_PLATFORM:-}"
PYTHON_VERSION="${PYTHON_VERSION:-3.11}"
PYTHON_IMAGE="${PYTHON_IMAGE:-}"
PYTHON_IMAGE_MIRROR="${PYTHON_IMAGE_MIRROR:-docker.m.daocloud.io/library/python:${PYTHON_VERSION}-slim}"
GO_DOWNLOAD_BASE="${GO_DOWNLOAD_BASE:-}"
GO_DOWNLOAD_MIRROR="${GO_DOWNLOAD_MIRROR:-https://mirrors.ustc.edu.cn/golang}"
TEMPLATES="${TEMPLATES:-python-basic multi-language}"
TEMPLATE_IMAGE_TAG="${TEMPLATE_IMAGE_TAG:-${PROJECT_VERSION}}"
BUILD_PYTHON_BASE="${BUILD_PYTHON_BASE:-false}"
BUILD_MULTI_BASE="${BUILD_MULTI_BASE:-false}"
PUSH_SWR_BASES="${PUSH_SWR_BASES:-false}"
SWR_REGISTRY="${SWR_REGISTRY:-}"
SWR_NAMESPACE="${SWR_NAMESPACE:-}"
SWR_CREDS="${SWR_CREDS:-}"
SWR_DEST_TLS_VERIFY="${SWR_DEST_TLS_VERIFY:-false}"
SWR_PYTHON_BASE_REPOSITORY="${SWR_PYTHON_BASE_REPOSITORY:-${PYTHON_BASE_IMAGE_NAME}}"
SWR_MULTI_BASE_REPOSITORY="${SWR_MULTI_BASE_REPOSITORY:-${MULTI_BASE_IMAGE_NAME}}"
SWR_PLATFORMS="${SWR_PLATFORMS:-linux/amd64,linux/arm64}"
SWR_OCI_OUTPUT_DIR="${SWR_OCI_OUTPUT_DIR:-/tmp/sandbox-base-oci-images}"
SWR_BUILDX_BUILDER="${SWR_BUILDX_BUILDER:-sandbox-swr-builder}"

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
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

require_swr_config() {
    if ! docker buildx version >/dev/null 2>&1; then
        log_error "docker buildx is required for multi-arch SWR push but was not found"
        exit 1
    fi
    if ! command -v skopeo >/dev/null 2>&1; then
        log_error "skopeo is required for SWR push but was not found in PATH"
        exit 1
    fi
    if [ -z "$SWR_REGISTRY" ]; then
        log_error "SWR registry is required. Set SWR_REGISTRY or pass --swr-registry"
        exit 1
    fi
    if [ -z "$SWR_NAMESPACE" ]; then
        log_error "SWR namespace is required. Set SWR_NAMESPACE or pass --swr-namespace"
        exit 1
    fi
    if [ -z "$SWR_CREDS" ]; then
        log_error "SWR credentials are required. Set SWR_CREDS or pass --swr-creds"
        exit 1
    fi
}

ensure_swr_buildx_builder() {
    require_swr_config

    if docker buildx inspect "$SWR_BUILDX_BUILDER" >/dev/null 2>&1; then
        log_info "Using buildx builder for SWR: ${SWR_BUILDX_BUILDER}"
    else
        log_info "Creating buildx builder for SWR with docker-container driver: ${SWR_BUILDX_BUILDER}"
        docker buildx create \
            --name "$SWR_BUILDX_BUILDER" \
            --driver docker-container >/dev/null
    fi

    docker buildx inspect --bootstrap "$SWR_BUILDX_BUILDER" >/dev/null
}

push_base_to_swr() {
    local dockerfile="$1"
    local base_build_args="$2"
    local image_name="$3"
    local image_tag="$4"
    local swr_repository="$5"
    local archive_name="${image_name//\//_}-${image_tag}.tar"
    local oci_archive="${SWR_OCI_OUTPUT_DIR}/${archive_name}"
    local source_ref="oci-archive:${oci_archive}"
    local dest_ref="docker://${SWR_REGISTRY}/${SWR_NAMESPACE}/${swr_repository}:${image_tag}"

    ensure_swr_buildx_builder
    mkdir -p "$SWR_OCI_OUTPUT_DIR"

    log_info "Building multi-arch OCI archive for SWR: ${image_name}:${image_tag}"
    log_info "Platforms: ${SWR_PLATFORMS}"
    docker buildx build \
        --builder "$SWR_BUILDX_BUILDER" \
        --platform "$SWR_PLATFORMS" \
        -f "$dockerfile" \
        -t "${image_name}:${image_tag}" \
        $base_build_args \
        --output "type=oci,dest=${oci_archive}" \
        "${PROJECT_ROOT}"

    log_info "Pushing base image to SWR: ${SWR_REGISTRY}/${SWR_NAMESPACE}/${swr_repository}:${image_tag}"
    skopeo copy \
        --all \
        --dest-tls-verify="${SWR_DEST_TLS_VERIFY}" \
        --dest-creds="${SWR_CREDS}" \
        "$source_ref" \
        "$dest_ref"
}

# Build stable Python runtime base image
build_python_base() {
    log_info "Building Python runtime base image: ${PYTHON_BASE_IMAGE_NAME}:${PYTHON_BASE_IMAGE_TAG}"
    log_info "This stable base contains system/runtime dependencies only; it does not include executor code."
    local build_args=""
    local python_image="$PYTHON_IMAGE"
    local platform_args=""
    if [ -n "$DOCKER_BUILD_PLATFORM" ]; then
        platform_args="--platform ${DOCKER_BUILD_PLATFORM}"
        log_info "Docker build platform: ${DOCKER_BUILD_PLATFORM}"
    fi
    if [ "$USE_MIRROR" = "true" ]; then
        build_args="--build-arg USE_MIRROR=true"
        python_image="${python_image:-$PYTHON_IMAGE_MIRROR}"
        log_info "Using mirror sources for Python runtime base"
    fi
    if [ -n "$python_image" ]; then
        build_args="${build_args} --build-arg PYTHON_IMAGE=${python_image}"
        log_info "Python base image: ${python_image}"
    fi

    docker build \
        $platform_args \
        -f "${SCRIPT_DIR}/bases/python/Dockerfile" \
        -t "${PYTHON_BASE_IMAGE_NAME}:${PYTHON_BASE_IMAGE_TAG}" \
        -t "${PYTHON_BASE_IMAGE_NAME}:latest" \
        $build_args \
        "${PROJECT_ROOT}"

    if [ "$PUSH" = "true" ]; then
        log_info "Pushing Python runtime base image to registry..."
        docker tag "${PYTHON_BASE_IMAGE_NAME}:${PYTHON_BASE_IMAGE_TAG}" "${REGISTRY}/${PYTHON_BASE_IMAGE_NAME}:${PYTHON_BASE_IMAGE_TAG}"
        docker push "${REGISTRY}/${PYTHON_BASE_IMAGE_NAME}:${PYTHON_BASE_IMAGE_TAG}"
    fi

    if [ "$PUSH_SWR_BASES" = "true" ]; then
        local swr_build_args="$build_args"
        push_base_to_swr \
            "${SCRIPT_DIR}/bases/python/Dockerfile" \
            "$swr_build_args" \
            "$PYTHON_BASE_IMAGE_NAME" \
            "$PYTHON_BASE_IMAGE_TAG" \
            "$SWR_PYTHON_BASE_REPOSITORY"
    fi
}

# Build stable multi-language runtime base image
build_multi_base() {
    log_info "Building multi-language runtime base image: ${MULTI_BASE_IMAGE_NAME}:${MULTI_BASE_IMAGE_TAG}"
    log_info "This stable base contains Python, Go, Bash, and toolchain dependencies only; it does not include executor code."
    local build_args="--build-arg BASE_IMAGE=${PYTHON_BASE_IMAGE_NAME}:${PYTHON_BASE_IMAGE_TAG}"
    local platform_args=""
    if [ -n "$DOCKER_BUILD_PLATFORM" ]; then
        platform_args="--platform ${DOCKER_BUILD_PLATFORM}"
        log_info "Docker build platform: ${DOCKER_BUILD_PLATFORM}"
    fi
    build_args="${build_args} --build-arg GO_DOWNLOAD_MIRROR=${GO_DOWNLOAD_MIRROR}"
    if [ -n "$GO_DOWNLOAD_BASE" ]; then
        build_args="${build_args} --build-arg GO_DOWNLOAD_BASE=${GO_DOWNLOAD_BASE}"
    fi
    if [ "$USE_MIRROR" = "true" ]; then
        build_args="${build_args} --build-arg USE_MIRROR=true"
        log_info "Using mirror sources for multi-language runtime base"
    fi

    docker build \
        $platform_args \
        -f "${SCRIPT_DIR}/bases/multi-language/Dockerfile" \
        $build_args \
        -t "${MULTI_BASE_IMAGE_NAME}:${MULTI_BASE_IMAGE_TAG}" \
        -t "${MULTI_BASE_IMAGE_NAME}:latest" \
        "${PROJECT_ROOT}"

    if [ "$PUSH" = "true" ]; then
        log_info "Pushing multi-language runtime base image to registry..."
        docker tag "${MULTI_BASE_IMAGE_NAME}:${MULTI_BASE_IMAGE_TAG}" "${REGISTRY}/${MULTI_BASE_IMAGE_NAME}:${MULTI_BASE_IMAGE_TAG}"
        docker push "${REGISTRY}/${MULTI_BASE_IMAGE_NAME}:${MULTI_BASE_IMAGE_TAG}"
    fi

    if [ "$PUSH_SWR_BASES" = "true" ]; then
        local swr_python_base_ref="${SWR_REGISTRY}/${SWR_NAMESPACE}/${SWR_PYTHON_BASE_REPOSITORY}:${PYTHON_BASE_IMAGE_TAG}"
        local swr_build_args="--build-arg BASE_IMAGE=${swr_python_base_ref}"
        swr_build_args="${swr_build_args} --build-arg GO_DOWNLOAD_MIRROR=${GO_DOWNLOAD_MIRROR}"
        if [ -n "$GO_DOWNLOAD_BASE" ]; then
            swr_build_args="${swr_build_args} --build-arg GO_DOWNLOAD_BASE=${GO_DOWNLOAD_BASE}"
        fi
        if [ "$USE_MIRROR" = "true" ]; then
            swr_build_args="${swr_build_args} --build-arg USE_MIRROR=true"
        fi
        push_base_to_swr \
            "${SCRIPT_DIR}/bases/multi-language/Dockerfile" \
            "$swr_build_args" \
            "$MULTI_BASE_IMAGE_NAME" \
            "$MULTI_BASE_IMAGE_TAG" \
            "$SWR_MULTI_BASE_REPOSITORY"
    fi
}

get_template_base_image() {
    local template="$1"
    case "$template" in
        python-basic)
            echo "${PYTHON_BASE_IMAGE_NAME}:${PYTHON_BASE_IMAGE_TAG}"
            ;;
        multi-language)
            echo "${MULTI_BASE_IMAGE_NAME}:${MULTI_BASE_IMAGE_TAG}"
            ;;
        *)
            log_error "Unsupported template: ${template}. Supported templates: python-basic, multi-language"
            exit 1
            ;;
    esac
}

get_template_description() {
    local template="$1"
    case "$template" in
        python-basic)
            echo "Minimal Python executor environment"
            ;;
        multi-language)
            echo "Python, Go, and Bash executor environment"
            ;;
        *)
            log_error "Unsupported template: ${template}. Supported templates: python-basic, multi-language"
            exit 1
            ;;
    esac
}

is_builtin_template() {
    local template="$1"
    case "$template" in
        python-basic|multi-language)
            return 0
            ;;
        *)
            return 1
            ;;
    esac
}

# Build template images
build_templates() {
    read -r -a templates <<< "$TEMPLATES"

    local build_args=""
    if [ "$USE_MIRROR" = "true" ]; then
        build_args="--build-arg USE_MIRROR=true"
        log_info "Using mirror sources for template images"
    fi

    for template in "${templates[@]}"; do
        local dockerfile="${SCRIPT_DIR}/templates/executor/Dockerfile"
        local image_name="sandbox-template-${template}"
        local image_tag="${TEMPLATE_IMAGE_TAG}"
        local base_image
        local template_description

        if ! is_builtin_template "$template"; then
            log_error "Unsupported template: ${template}. Supported templates: python-basic, multi-language"
            exit 1
        fi

        log_info "Building template: ${template}"

        base_image="$(get_template_base_image "$template")"
        template_description="$(get_template_description "$template")"

        docker build \
            -f "$dockerfile" \
            --build-arg "BASE_IMAGE=${base_image}" \
            --build-arg "TEMPLATE_NAME=${template}" \
            --build-arg "TEMPLATE_DESCRIPTION=${template_description}" \
            --build-arg "TEMPLATE_VERSION=${image_tag}" \
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
    log_info "Project version: ${PROJECT_VERSION}"
    log_info "Template image tag: ${TEMPLATE_IMAGE_TAG}"

    if [ "$PUSH_SWR_BASES" = "true" ] && [ "$BUILD_PYTHON_BASE" != "true" ] && [ "$BUILD_MULTI_BASE" != "true" ]; then
        log_warn "SWR base image push is enabled, but no base image build is enabled. Use --build-bases, --build-python-base, or --build-multi-base."
    fi

    if [ "$BUILD_PYTHON_BASE" = "true" ]; then
        build_python_base
    else
        log_info "Skipping Python runtime base build. Using existing image: ${PYTHON_BASE_IMAGE_NAME}:${PYTHON_BASE_IMAGE_TAG}"
    fi

    if [ "$BUILD_MULTI_BASE" = "true" ]; then
        build_multi_base
    else
        log_info "Skipping multi-language runtime base build. Using existing image: ${MULTI_BASE_IMAGE_NAME}:${MULTI_BASE_IMAGE_TAG}"
    fi

    # Build final versioned executor/template images
    build_templates

    log_info "Build complete!"
    log_info "Images:"
    if [ "$BUILD_PYTHON_BASE" = "true" ]; then
        log_info "  - ${PYTHON_BASE_IMAGE_NAME}:${PYTHON_BASE_IMAGE_TAG}"
    fi
    if [ "$BUILD_MULTI_BASE" = "true" ]; then
        log_info "  - ${MULTI_BASE_IMAGE_NAME}:${MULTI_BASE_IMAGE_TAG}"
    fi
    read -r -a templates <<< "$TEMPLATES"
    for template in "${templates[@]}"; do
        log_info "  - sandbox-template-${template}:latest"
    done
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
        --python-base-image)
            PYTHON_BASE_IMAGE_NAME="$2"
            shift 2
            ;;
        --python-base-tag)
            PYTHON_BASE_IMAGE_TAG="$2"
            shift 2
            ;;
        --multi-base-image)
            MULTI_BASE_IMAGE_NAME="$2"
            shift 2
            ;;
        --multi-base-tag)
            MULTI_BASE_IMAGE_TAG="$2"
            shift 2
            ;;
        --build-bases)
            BUILD_PYTHON_BASE=true
            BUILD_MULTI_BASE=true
            shift
            ;;
        --build-python-base)
            BUILD_PYTHON_BASE=true
            shift
            ;;
        --build-multi-base)
            BUILD_MULTI_BASE=true
            shift
            ;;
        --push-swr-bases)
            PUSH_SWR_BASES=true
            shift
            ;;
        --swr-registry)
            SWR_REGISTRY="$2"
            shift 2
            ;;
        --swr-namespace)
            SWR_NAMESPACE="$2"
            shift 2
            ;;
        --swr-creds)
            SWR_CREDS="$2"
            shift 2
            ;;
        --swr-dest-tls-verify)
            SWR_DEST_TLS_VERIFY="$2"
            shift 2
            ;;
        --swr-python-base-repository)
            SWR_PYTHON_BASE_REPOSITORY="$2"
            shift 2
            ;;
        --swr-multi-base-repository)
            SWR_MULTI_BASE_REPOSITORY="$2"
            shift 2
            ;;
        --swr-platforms)
            SWR_PLATFORMS="$2"
            shift 2
            ;;
        --swr-oci-output-dir)
            SWR_OCI_OUTPUT_DIR="$2"
            shift 2
            ;;
        --swr-buildx-builder)
            SWR_BUILDX_BUILDER="$2"
            shift 2
            ;;
        --template-tag)
            TEMPLATE_IMAGE_TAG="$2"
            shift 2
            ;;
        --templates)
            TEMPLATES="$2"
            shift 2
            ;;
        --use-mirror)
            USE_MIRROR=true
            shift
            ;;
        --docker-build-platform)
            DOCKER_BUILD_PLATFORM="$2"
            shift 2
            ;;
        --go-download-base)
            GO_DOWNLOAD_BASE="$2"
            shift 2
            ;;
        --go-download-mirror)
            GO_DOWNLOAD_MIRROR="$2"
            shift 2
            ;;
        --python-image)
            PYTHON_IMAGE="$2"
            shift 2
            ;;
        --python-image-mirror)
            PYTHON_IMAGE_MIRROR="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --push        Push images to registry"
            echo "  --registry    Docker registry (default: localhost:5000)"
            echo "  --build-bases Build both stable runtime base images"
            echo "  --build-python-base Build stable Python runtime base image"
            echo "  --build-multi-base Build stable multi-language runtime base image"
            echo "  --push-swr-bases Push built base images to SWR with skopeo"
            echo "  --swr-registry SWR registry host, e.g. swr.cn-east-3.myhuaweicloud.com"
            echo "  --swr-namespace SWR namespace path, e.g. kweaver-ai/sandbox"
            echo "  --swr-creds SWR credentials in username:password format"
            echo "  --swr-dest-tls-verify SWR destination TLS verify flag (default: false)"
            echo "  --swr-python-base-repository SWR repository for Python base (default: sandbox-python-executor-base)"
            echo "  --swr-multi-base-repository SWR repository for multi-language base (default: sandbox-multi-executor-base)"
            echo "  --swr-platforms Multi-arch platforms for SWR base builds (default: linux/amd64,linux/arm64)"
            echo "  --swr-oci-output-dir Directory for temporary OCI archives (default: /tmp/sandbox-base-oci-images)"
            echo "  --swr-buildx-builder Buildx builder for SWR multi-arch OCI exports (default: sandbox-swr-builder)"
            echo "  --python-base-image Python runtime base image name (default: sandbox-python-executor-base)"
            echo "  --python-base-tag Python runtime base image tag (default: python3.11-v1)"
            echo "  --multi-base-image Multi-language runtime base image name (default: sandbox-multi-executor-base)"
            echo "  --multi-base-tag Multi-language runtime base image tag (default: go1.25-python3.11-v1)"
            echo "  --template-tag Template image tag (default: contents of VERSION)"
            echo "  --templates    Space-separated templates to build (default: \"python-basic multi-language\")"
            echo "  --use-mirror  Use mirror sources for building images"
            echo "  --docker-build-platform Platform for local docker build commands, e.g. linux/amd64 or linux/arm64"
            echo "  --go-download-base Go tarball download base URL (default: https://go.dev/dl)"
            echo "  --go-download-mirror Go tarball mirror URL used with --use-mirror (default: https://mirrors.ustc.edu.cn/golang)"
            echo "  --python-image Python base image for images/bases/python (default: python:3.11-slim)"
            echo "  --python-image-mirror Python base image used with --use-mirror (default: docker.m.daocloud.io/library/python:3.11-slim)"
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
