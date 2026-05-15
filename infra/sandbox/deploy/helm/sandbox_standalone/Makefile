.PHONY: help install upgrade uninstall lint template test clean

# Default target
help:
	@echo "Sandbox Standalone Helm Chart"
	@echo ""
	@echo "Usage:"
	@echo "  make [TARGET]"
	@echo ""
	@echo "Targets:"
	@echo "  help       Show this help message"
	@echo "  lint       Lint the Helm chart"
	@echo "  template   Generate the template YAML"
	@echo "  install    Install the Helm chart"
	@echo "  upgrade    Upgrade the Helm chart"
	@echo "  uninstall  Uninstall the Helm chart"
	@echo "  test       Test the Helm chart with dry-run"
	@echo "  clean      Remove generated files"

# Lint the Helm chart
lint:
	helm lint .

# Generate the template YAML
template:
	helm template sandbox . > helm-template-gen.yaml
	@echo "Template generated: helm-template-gen.yaml"

# Install the Helm chart
install:
	helm install sandbox . --namespace sandbox-system --create-namespace

# Install with custom values
install-dev:
	helm install sandbox . --namespace sandbox-system --create-namespace --set controlPlane.env.ENVIRONMENT=development

# Install with local images
install-local:
	helm install sandbox . --namespace sandbox-system --create-namespace \
		--set image.registry= \
		--set image.controlPlane.repository=sandbox-control-plane \
		--set image.controlPlane.tag=latest \
		--set image.web.repository=sandbox-web \
		--set image.web.tag=latest

# Upgrade the Helm chart
upgrade:
	helm upgrade sandbox . --namespace sandbox-system

# Upgrade with custom values
upgrade-dev:
	helm upgrade sandbox . --namespace sandbox-system --set controlPlane.env.ENVIRONMENT=development

# Uninstall the Helm chart
uninstall:
	helm uninstall sandbox --namespace sandbox-system

# Test the Helm chart with dry-run
test:
	helm install sandbox . --namespace sandbox-system --dry-run --debug

# Show values
show-values:
	helm show values .

# Clean generated files
clean:
	rm -f helm-template-gen.yaml
	@echo "Cleaned up generated files"
