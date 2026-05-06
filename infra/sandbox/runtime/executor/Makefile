.PHONY: build build-templates test test-unit test-cov help

# Project paths
PROJECT_ROOT := $(shell cd ../.. && pwd)

help: ## Show this help message
	@echo "Sandbox Executor - Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: build-templates ## Build default executor/template images via images/build.sh

build-templates: ## Build default executor/template images via images/build.sh
	cd $(PROJECT_ROOT)/images && ./build.sh

test: ## Run executor import test in the local Python environment
	@echo "Testing executor import..."
	PYTHONPATH=.. uv run python -c "from executor.interfaces.http.rest import create_app; print('OK')"
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
