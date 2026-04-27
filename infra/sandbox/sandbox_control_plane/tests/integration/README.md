# Integration Tests

Integration tests for the Sandbox Control Plane. These tests run against a live docker-compose stack to validate end-to-end functionality.

## Test Results (Latest)

- **65 passed** ✅
- **1 failed** (environment variables issue - known)
- **5 skipped** (Docker not available, executor connection issues)

## Prerequisites

- Docker (20.10+)
- docker-compose (v2.0+) or Docker Compose plugin
- Python 3.11+
- uv (Python package manager) - Install via: `curl -LsSf https://astral.sh/uv/install.sh | sh`

## Test Structure

```
tests/integration/
├── conftest.py                          # Shared fixtures and configuration
├── api/                                 # API endpoint tests
│   ├── test_internal_api.py            # Internal callback API
│   └── test_sessions_api.py            # Session management endpoints
├── test_templates_api.py                # Template CRUD operations
├── test_executions_api.py               # Code execution endpoints
├── test_e2e_workflow.py                 # End-to-end workflow tests
├── test_e2e_s3_workspace.py            # S3 workspace mounting tests
├── test_state_sync.py                   # State synchronization tests
├── test_background_tasks.py             # Background task tests
└── test_docker_scheduler.py             # Docker scheduler integration tests
```

## Setup

### 1. Build Template Images

Before running tests, build the required Docker images:

```bash
cd /path/to/sandbox

# Build all executor and template images
bash images/build.sh

# Or build manually
docker build -t sandbox-executor-base:latest -f runtime/executor/Dockerfile .
```

This builds:
- `sandbox-executor-base:latest` - Base executor image
- `sandbox-template-python-basic:latest` - Python template

### 2. Start Services

Start the docker-compose stack:

```bash
cd sandbox_control_plane

# Start all services
docker-compose up -d

# Verify services are running
docker-compose ps

# Check control plane is healthy
curl http://localhost:8000/api/v1/health
```

Expected output:
```json
{
  "status": "healthy",
  "version": "2.1.0",
  "uptime": ...
}
```

### 3. Install Test Dependencies

```bash
cd sandbox_control_plane

# Install dependencies with uv
uv sync

# Install dev dependencies (includes pytest, httpx, etc.)
uv sync --extra dev
```

## Running Tests

### Run All Integration Tests

```bash
# From sandbox_control_plane directory
pytest tests/integration/ -v
```

### Run Specific Test File

```bash
pytest tests/integration/test_templates_api.py -v
```

### Run Specific Test

```bash
pytest tests/integration/test_executions_api.py::TestExecutionsAPI::test_execute_python_code -v
```

### Run with Coverage

```bash
pytest tests/integration/ --cov=src --cov-report=html --cov-report=term-missing
```

### Run with Verbose Output

```bash
pytest tests/integration/ -vv -s
```

## Test Configuration

Tests use environment variables for configuration:

```bash
# Control plane URL (default: http://localhost:8000)
export CONTROL_PLANE_URL="http://localhost:8000"
```

**Important**: Tests run sequentially (not in parallel) to avoid system overload:
- Added 0.5s delay between tests
- Blocks pytest-xdist parallel execution
- Automatic session cleanup after each test

## Code Execution Format

**The executor expects AWS Lambda-style handler syntax:**

```python
def handler(event):
    # Your code here
    print("Hello, World!")
    return {"status": "success"}
```

Parameters:
- `event` - Optional dictionary containing event data passed to the handler
- Return value - Must be JSON-serializable (dict, list, str, int, float, bool, None)

Example with event data:
```python
def handler(event):
    name = event.get("name", "World")
    return {"message": f"Hello, {name}!"}
```

## Test Fixtures

Key fixtures in `conftest.py`:

| Fixture | Scope | Description |
|---------|-------|-------------|
| `http_client` | function | AsyncClient for API calls (auto-trust_env=False) |
| `test_template_id` | function | Creates/gets test template |
| `test_session_id` | function | Creates ephemeral test session |
| `persistent_session_id` | function | Creates persistent test session |
| `wait_for_execution_completion` | function | Waits for execution to finish |
| `auto_cleanup_sessions` | function | Auto-cleanup sessions after test |

**Session Tracking**: All sessions created via `_create_session_and_track()` are automatically tracked and cleaned up after each test.

## API Endpoints Tested

### Templates (`/api/v1/templates`)
- `POST /templates` - Create template
- `GET /templates` - List templates
- `GET /templates/{id}` - Get template details
- `PUT /templates/{id}` - Update template
- `DELETE /templates/{id}` - Delete template

### Sessions (`/api/v1/sessions`)
- `POST /sessions` - Create session
- `GET /sessions/{id}` - Get session details
- `DELETE /sessions/{id}` - Terminate session

### Executions (`/api/v1/executions`)
- `POST /executions/sessions/{session_id}/execute` - Execute code
- `GET /executions/{execution_id}/status` - Get execution status
- `GET /executions/{execution_id}/result` - Get execution result
- `GET /executions/sessions/{session_id}/executions` - List session executions

### Files (`/api/v1/sessions/{id}/files`)
- `POST /files/upload?path={path}` - Upload file to workspace
- `GET /files/{path}` - Download file from workspace

### Health (`/api/v1/health`)
- `GET /health` - Health check

## Troubleshooting

### Control Plane Not Starting

```bash
# Check logs
docker-compose logs control-plane

# Restart service
docker-compose restart control-plane
```

### Executor Image Outdated

```bash
# Rebuild executor images (from project root)
cd /path/to/sandbox
bash images/build.sh

# Restart control plane
cd sandbox_control_plane
docker-compose restart control-plane
```

### Tests Failing with 422 Unprocessable Entity

This usually indicates a schema mismatch between executor and control plane:

1. Check executor callback code matches API schema
2. Rebuild executor images
3. Restart control plane

```bash
# Fix and rebuild
bash images/build.sh
docker-compose restart control-plane
```

### Tests Timing Out

```bash
# Check if control plane is responsive
curl http://localhost:8000/api/v1/health

# Check executor connection
docker logs sandbox-control-plane | grep executor
```

### Permission Denied (Docker Socket)

```bash
# Ensure Docker socket is mounted correctly
docker-compose exec control-plane ls -la /var/run/docker.sock

# Check control plane container has Docker access
docker-compose exec control-plane docker ps
```

### Session Cleanup Issues

Sessions are tracked at module level and cleaned up automatically:

```python
# Manual cleanup if needed
curl http://localhost:8000/api/v1/sessions | jq -r '.[].id' | xargs -I {} curl -X DELETE http://localhost:8000/api/v1/sessions/{}
```

## Known Issues

1. **Environment Variables Not Passed**: The `test_execute_python_with_env_vars` test fails because environment variables passed via `env_vars` are not being injected into the execution environment.

2. **S3 Workspace Mounting**: Tests that require S3 workspace mounting will fail if S3/MinIO is not properly configured.

3. **Bubblewrap Permissions**: Some tests may be skipped if bubblewrap namespace permissions are not available (on macOS).

## Stopping Services

```bash
# Stop services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

## Manual API Testing

### Create Template

```bash
curl -X POST http://localhost:8000/api/v1/templates \
  -H "Content-Type: application/json" \
  -d '{
    "id": "test_template_python",
    "name": "Python Basic Test",
    "image_url": "sandbox-template-python-basic:latest",
    "runtime_type": "python3.11",
    "default_cpu_cores": 1.0,
    "default_memory_mb": 512,
    "default_disk_mb": 1024,
    "default_timeout_sec": 300
  }'
```

### Create Session

```bash
curl -X POST http://localhost:8000/api/v1/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": "test_template_python",
    "timeout": 300,
    "cpu": "1",
    "memory": "512Mi",
    "disk": "1Gi"
  }'
```

### Execute Code

```bash
SESSION_ID="your_session_id"
curl -X POST http://localhost:8000/api/v1/executions/sessions/$SESSION_ID/execute \
  -H "Content-Type: application/json" \
  -d '{
    "code": "def handler(event):\n    return {\"message\": \"Hello, World!\"}",
    "language": "python",
    "timeout": 10,
    "event": {},
    "env_vars": {}
  }'
```

### Get Execution Result

```bash
EXECUTION_ID="your_execution_id"
curl http://localhost:8000/api/v1/executions/$EXECUTION_ID/result
```

## CI/CD Integration

For CI/CD pipelines, use the following commands:

```bash
# Build and start services
cd sandbox_control_plane
docker-compose up -d --build

# Wait for services to be healthy
sleep 30

# Run tests
pytest tests/integration/ -v --junitxml=test-results.xml

# Stop services
docker-compose down -v
```

## Test Categories

### Unit Integration Tests (Fast)
- API contract tests
- Template CRUD operations
- Health checks

### State Sync Tests (Medium)
- Session state consistency
- Container lifecycle
- State recovery scenarios

### End-to-End Tests (Slow)
- Full workflow tests
- Multiple executions in session
- File operations

### Background Task Tests (Medium)
- Session cleanup service
- Health check execution
- Graceful shutdown
