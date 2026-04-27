.PHONY: build push test test-unit test-cov shell help

# Image configuration
IMAGE_NAME = sandbox-executor-base
REGISTRY ?= localhost:5000
VERSION ?= latest

# Project paths
PROJECT_ROOT := $(shell cd ../.. && pwd)
DOCKERFILE := $(PROJECT_ROOT)/runtime/executor/Dockerfile
BUILD_CONTEXT := $(PROJECT_ROOT)

# Docker build options
DOCKER_BUILDKIT ?= 1
export DOCKER_BUILDKIT

help: ## Show this help message
	@echo "Sandbox Executor Base Image - Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the executor base image
	@echo "Building $(IMAGE_NAME):$(VERSION)..."
	@echo "Dockerfile: $(DOCKERFILE)"
	@echo "Context: $(BUILD_CONTEXT)"
	docker build -f $(DOCKERFILE) -t $(IMAGE_NAME):$(VERSION) -t $(IMAGE_NAME):latest $(BUILD_CONTEXT)

build-no-cache: ## Build without cache
	@echo "Building $(IMAGE_NAME):$(VERSION) (no cache)..."
	docker build --no-cache -f $(DOCKERFILE) -t $(IMAGE_NAME):$(VERSION) -t $(IMAGE_NAME):latest $(BUILD_CONTEXT)

push: ## Push image to registry
	@echo "Tagging and pushing $(REGISTRY)/$(IMAGE_NAME):$(VERSION)..."
	docker tag $(IMAGE_NAME):$(VERSION) $(REGISTRY)/$(IMAGE_NAME):$(VERSION)
	docker tag $(IMAGE_NAME):latest $(REGISTRY)/$(IMAGE_NAME):latest
	docker push $(REGISTRY)/$(IMAGE_NAME):$(VERSION)
	docker push $(REGISTRY)/$(IMAGE_NAME):latest

test: ## Run basic tests
	@echo "Testing executor import..."
	docker run --rm $(IMAGE_NAME):$(VERSION) \
		python -c "from executor.interfaces.http.rest import create_app; print('OK')"
	@echo "Test passed!"

test-unit: ## Run unit tests using uv
	@echo "Running unit tests..."
	uv run pytest tests/unit/ -v --tb=short

test-cov: ## Run unit tests with coverage report
	@echo "Running unit tests with coverage..."
	rm -f .coverage
	uv run coverage run -m pytest tests/unit/ -m "not slow"
	uv run coverage report --fail-under=90
	uv run coverage html
	@echo "Coverage report generated in htmlcov/"

shell: ## Open a shell in the container
	docker run --rm -it $(IMAGE_NAME):$(VERSION) /bin/bash

run: ## Run the executor daemon
	docker run --rm -p 8080:8080 \
		-e DISABLE_BWRAP=true \
		$(IMAGE_NAME):$(VERSION)

clean: ## Remove dangling images
	@echo "Cleaning up dangling images..."
	docker image prune -f

inspect: ## Show image details
	docker inspect $(IMAGE_NAME):$(VERSION)

history: ## Show image layers
	docker history $(IMAGE_NAME):$(VERSION)
