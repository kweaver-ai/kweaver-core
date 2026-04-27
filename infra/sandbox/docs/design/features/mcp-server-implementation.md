# MCP Server Implementation Plan for Synchronous Code Execution

## Overview

Package the existing synchronous code execution endpoint (`/api/v1/sessions/{session_id}/execute-sync`) as an MCP (Model Context Protocol) server, embedded as an optional module within the `sandbox_control_plane` service.

**User Requirements:**
- Deployment: Embedded module within `sandbox_control_plane`
- Transport: STDIO + SSE (Server-Sent Events)
- Session Management: Caller provides `session_id` explicitly, no automatic session creation/management

---

## Architecture

```
sandbox_control_plane/
├── src/
│   ├── interfaces/
│   │   └── mcp/                 # NEW: MCP server interface
│   │       ├── __init__.py
│   │       ├── main.py          # MCP server entry point
│   │       ├── tools.py         # MCP tool definitions
│   │       └── transport.py     # STDIO + SSE transport handlers
│   └── shared/
│       └── mcp_settings.py      # MCP configuration
├── pyproject.toml               # Add MCP dependencies
└── README.md                    # Add MCP documentation
```

---

## Implementation Steps

### Step 1: Add MCP Dependencies

**File:** `sandbox_control_plane/pyproject.toml`

Add MCP dependencies as optional extras:

```toml
[project.optional-dependencies]
mcp = [
    "mcp[cli]>=1.2.0",
    "fastmcp>=0.1.0",
]
```

---

### Step 2: Create MCP Settings Module

**File:** `sandbox_control_plane/src/shared/mcp_settings.py`

Configuration for MCP server using Pydantic Settings:

```python
from pydantic_settings import BaseSettings

class MCPSettings(BaseSettings):
    """MCP server configuration"""
    ENABLED: bool = True
    TRANSPORT: str = "stdio"  # "stdio" or "sse"
    HOST: str = "0.0.0.0"
    PORT: int = 8001
    CONTROL_PLANE_URL: str = "http://localhost:8000"

    class Config:
        env_prefix = "MCP_"
```

---

### Step 3: Create MCP Tools Module

**File:** `sandbox_control_plane/src/interfaces/mcp/tools.py`

Define MCP tools that wrap existing synchronous execution:

```python
from mcp.server.fastmcp import FastMCP
from src.application.services.session_service import SessionService
from src.interfaces.rest.schemas.request import ExecuteCodeRequest
from src.interfaces.rest.schemas.response import ExecutionResponse

mcp = FastMCP("sandbox-executor")

@mcp.tool()
async def execute_code(
    session_id: str,
    code: str,
    language: str = "python",
    timeout: int = 30,
    event: dict | None = None
) -> dict:
    """Execute code in a sandbox session.

    Args:
        session_id: The sandbox session ID (must exist and be active)
        code: Code to execute (AWS Lambda handler format)
        language: Programming language (python, javascript, shell)
        timeout: Execution timeout in seconds (1-3600)
        event: Optional event data for the handler

    Returns:
        Execution result including stdout, stderr, exit_code, return_value
    """
    # Call existing synchronous execution endpoint logic
    pass

# Note: Session management tools are NOT included per user requirement
# Caller must provide valid session_id
```

---

### Step 4: Create Transport Handlers

**File:** `sandbox_control_plane/src/interfaces/mcp/transport.py`

Implement both STDIO and SSE transports:

```python
from mcp.server.fastmcp import FastMCP
from .tools import mcp
from .mcp_settings import MCPSettings

async def run_stdio():
    """Run MCP server with STDIO transport (for Claude Desktop)"""
    await mcp.run(transport="stdio")

async def run_sse(host: str, port: int):
    """Run MCP server with SSE transport"""
    await mcp.run(transport="sse", host=host, port=port)
```

---

### Step 5: Create MCP Entry Point

**File:** `sandbox_control_plane/src/interfaces/mcp/main.py`

```python
import asyncio
from src.interfaces.mcp.transport import run_stdio, run_sse
from src.shared.mcp_settings import MCPSettings

async def main():
    settings = MCPSettings()
    if settings.TRANSPORT == "stdio":
        await run_stdio()
    else:
        await run_sse(settings.HOST, settings.PORT)

if __name__ == "__main__":
    asyncio.run(main())
```

---

### Step 6: Update Project Configuration

**File:** `sandbox_control_plane/pyproject.toml`

Add console script entry point:

```toml
[project.scripts]
sandbox-mcp-server = "src.interfaces.mcp.main:main"
```

---

### Step 7: Add Docker Configuration (Optional)

**File:** `deploy/k8s/XX-mcp-server-deployment.yaml` (NEW)

K8s deployment for MCP server with SSE transport:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: sandbox-mcp-server
spec:
  selector:
    app: sandbox-mcp-server
  ports:
  - port: 8001
    targetPort: 8001
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sandbox-mcp-server
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sandbox-mcp-server
  template:
    metadata:
      labels:
        app: sandbox-mcp-server
    spec:
      containers:
      - name: mcp-server
        image: sandbox-control-plane:latest
        command: ["python", "-m", "src.interfaces.mcp.main"]
        env:
        - name: MCP_TRANSPORT
          value: "sse"
        ports:
        - containerPort: 8001
```

---

## Critical Files to Modify

| File Path | Action |
|-----------|--------|
| `sandbox_control_plane/pyproject.toml` | Add MCP dependencies and entry point |
| `sandbox_control_plane/src/shared/mcp_settings.py` | CREATE - MCP configuration |
| `sandbox_control_plane/src/interfaces/mcp/__init__.py` | CREATE - Package init |
| `sandbox_control_plane/src/interfaces/mcp/main.py` | CREATE - Entry point |
| `sandbox_control_plane/src/interfaces/mcp/tools.py` | CREATE - MCP tool definitions |
| `sandbox_control_plane/src/interfaces/mcp/transport.py` | CREATE - Transport handlers |
| `sandbox_control_plane/README.md` | UPDATE - Add MCP documentation |
| `deploy/k8s/XX-mcp-server-deployment.yaml` | CREATE - K8s deployment (optional) |

---

## Verification Plan

### 1. Local Testing (STDIO Transport)

```bash
# Install with MCP dependencies
cd sandbox_control_plane
pip install -e ".[mcp]"

# Run MCP server
python -m src.interfaces.mcp.main
```

**Test with Claude Desktop configuration:**
```json
{
  "mcpServers": {
    "sandbox-executor": {
      "command": "python",
      "args": ["-m", "src.interfaces.mcp.main"],
      "cwd": "/path/to/sandbox-v2/sandbox/sandbox_control_plane",
      "env": {
        "CONTROL_PLANE_URL": "http://localhost:8000"
      }
    }
  }
}
```

### 2. SSE Transport Testing

```bash
# Start MCP server with SSE
MCP_TRANSPORT=sse MCP_PORT=8001 python -m src.interfaces.mcp.main

# Test connection
curl http://localhost:8001/health
```

### 3. End-to-End Test Flow

1. Create a session via Control Plane API
2. Use MCP tool to execute code in that session
3. Verify execution results are returned

```python
# Via MCP
result = await execute_code(
    session_id="test-session-123",
    code="def handler(event, context): return {'result': 'hello world'}",
    language="python"
)
```

---

## Dependencies

- `mcp[cli]>=1.2.0` - Official MCP SDK
- `fastmcp>=0.1.0` - FastMCP framework for decorator-based tools
- Existing control plane dependencies (httpx, pydantic, etc.)

---

## Notes

- **Session Management**: Per user requirement, the MCP server will NOT create/destroy sessions. Callers must provide a valid `session_id` that exists and is active.
- **Error Handling**: The MCP tools should propagate Control Plane API errors clearly to MCP clients.
- **Authentication**: Future enhancement - add API key authentication for MCP connections.
- **Performance**: The MCP layer adds minimal overhead as it directly calls existing service methods.
