# Fetch Log Tool

A Kubernetes log collection tool for data-agent services.

## Features

- Automatically collects logs from specified Kubernetes services
- Supports both console logs and container file logs
- Special handling for agent-executor service (collects both console and file logs)
- Outputs logs in structured JSON format
- Cross-platform support (Linux AMD64 and ARM64)

## Prerequisites

- Go 1.21 or higher
- kubectl configured with cluster access
- Permissions to read pods and logs from the target namespace

## Installation

### Build from source

```bash
# Build for current platform
make build

# Build for all supported platforms
make build-all

# Build for specific platform
make build-linux-amd64
make build-linux-arm64
```

## Usage

### Basic usage (default services: agent-app, agent-executor)

```bash
./build/fetch_log
```

### Specify custom services

```bash
./build/fetch_log --svc_list "agent-factory,agent-memory,mf-model-manager"
```

### Using make commands

```bash
# Test with default services
make test

# Test with custom services
make test-custom
```

## Output

The tool generates a JSON file named `log_<timestamp>.json` with the following structure:

```json
[
  {
    "svc_name": "agent-app",
    "pod": "agent-app-7d9f8b6c-x5k2p",
    "fetch_time": "2025-01-08 10:30:45",
    "fecth_log_lines": 300,
    "log_detail": "console log content here..."
  },
  {
    "svc_name": "agent-executor",
    "pod": "agent-executor-5c8d9f7a-n3k4m",
    "fetch_time": "2025-01-08 10:30:46",
    "fecth_log_lines": 300,
    "log_detail": "console logs\n\n=== Container File Logs ===\n=== agent-executor.log ===\n...\n=== dolphin.log ===\n...\n=== request.log ===\n..."
  }
]
```

## Special Handling for agent-executor

For the `agent-executor` service, the tool collects:
1. Console logs (via `kubectl logs`)
2. Container file logs from the `log/` directory:
   - `agent-executor.log`
   - `dolphin.log`
   - `request.log`

## Log Collection Process

1. **Pod Discovery**: Uses `kubectl get pods -A | grep [service-name]` to find pods
2. **Console Logs**: Uses `kubectl logs -n [namespace] [pod-name] --tail=300`
3. **File Logs** (agent-executor only): Uses `kubectl exec` to read files from container

## Supported Platforms

- Linux x86_64 (AMD64)
- Linux ARM64

## Troubleshooting

### No pods found
- Verify kubectl is configured correctly: `kubectl get nodes`
- Check if the service name is correct
- Ensure you have permissions to list pods in all namespaces

### Permission denied
- Ensure your kubeconfig has proper RBAC permissions
- Required permissions: `pods/log`, `pods/get` on target namespaces

### kubectl command not found
- Install kubectl or add it to your PATH
- Verify installation: `kubectl version --client`

## Development

```bash
# Run without building
make run

# Clean build artifacts
make clean

# Show all available commands
make help
```

## License

Internal tool for data-agent log collection.
