BINARY_NAME=fetch_log
BUILD_DIR=build

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run

# Cross-compilation targets
LINUX_AMD64=linux/amd64
LINUX_ARM64=linux/arm64

.PHONY: all build clean run test

all: build

# Build for current platform
build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./main.go

# Build for Linux AMD64
build-linux-amd64:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./main.go

# Build for Linux ARM64
build-linux-arm64:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./main.go

# Build for all platforms
build-all: build-linux-amd64 build-linux-arm64

# Run the application
run:
	$(GORUN) main.go

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f log_*.json

# Test with default services
test: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Test with specific services
test-custom: build
	./$(BUILD_DIR)/$(BINARY_NAME) --svc_list "agent-factory,agent-memory"

# Help target
help:
	@echo "Available targets:"
	@echo "  make              - Build for current platform"
	@echo "  make build        - Build for current platform"
	@echo "  make build-linux-amd64  - Build for Linux AMD64"
	@echo "  make build-linux-arm64  - Build for Linux ARM64"
	@echo "  make build-all    - Build for all supported platforms"
	@echo "  make run          - Run the application"
	@echo "  make test         - Build and test with default services"
	@echo "  make test-custom  - Build and test with specific services"
	@echo "  make clean        - Remove build artifacts and log files"
	@echo "  make help         - Show this help message"
