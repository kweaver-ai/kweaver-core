# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **Sandbox Platform** designed for secure code execution in AI agent applications. The project provides isolated execution environments for running untrusted code, with multi-layer security isolation (container + Bubblewrap).

**Current State**: This is an **active implementation** project. The codebase contains implemented services for session management, code execution, template management, and web-based administration interface.

## Architecture

### High-Level Design

The system uses a **Control Plane + Container Scheduler** architecture:

- **Control Plane (管理中心)**: FastAPI-based management service handling API requests, scheduling, session management, and monitoring
- **Web Console (Web管理界面)**: React-based web application for visual management of sessions, templates, and executions
- **Container Scheduler (容器调度器)**: Internal module that directly calls Docker/K8s APIs
  - **Docker Scheduler**: Accesses Docker socket directly via aiodocker SDK
  - **K8s Scheduler**: Accesses Kubernetes API via ServiceAccount and python client
- **Executor (执行器)**: HTTP daemon running inside containers (sandbox-executor) that receives execution requests and spawns isolated processes
- **Dual-Layer Isolation**: Container isolation (first layer) + Bubblewrap process isolation (second layer, configurable)

### Key Components

1. **Control Plane Components** (`sandbox_control_plane/`):
   - API Gateway (FastAPI + Uvicorn)
   - Scheduler with intelligent task distribution (supports ephemeral and persistent session modes)
   - Session Manager with database-backed state (MariaDB)
   - Template Manager for sandbox environment definitions
   - Health Probe for container monitoring
   - Session Cleanup Service for automatic resource reclamation
   - State Sync Service for startup health checks

2. **Web Console** (`sandbox_web/`):
   - React 18 + TypeScript frontend
   - Rsbuild build system
   - Ant Design UI components
   - Monaco Editor for code editing
   - Session management and monitoring interface
   - Template CRUD operations
   - Real-time execution status tracking

3. **Executor** (`runtime/executor/`):
   - HTTP daemon running inside sandbox containers
   - Python code execution with isolated subprocess
   - Optional Bubblewrap integration for enhanced security
   - Workspace file management
   - Execution result streaming

### Session Modes

The system supports two session modes with different scheduling strategies:

- **Ephemeral Mode**: One-time execution, container destroyed after completion. Optimal for isolated, stateless tasks.
- **Persistent Mode**: Long-running session that can accept multiple execution requests. Optimal for interactive workflows and stateful operations.

### Protocol Design

The Control Plane exposes RESTful API to external clients, while Container Scheduler directly calls Docker/K8s SDKs:

```
# Session Management
POST   /api/v1/sessions                 # Create session
GET    /api/v1/sessions                 # List sessions (with filtering)
GET    /api/v1/sessions/{id}            # Get session details
DELETE /api/v1/sessions/{id}            # Terminate session

# Execution
POST   /api/v1/sessions/{id}/execute    # Submit execution task
GET    /api/v1/sessions/{id}/status     # Query execution status
GET    /api/v1/sessions/{id}/result     # Get execution results

# Template Management
GET    /api/v1/templates                # List templates
GET    /api/v1/templates/{id}           # Get template details

# File Operations
POST   /api/v1/sessions/{id}/files      # Upload workspace files
GET    /api/v1/sessions/{id}/files/{filename}  # Download workspace files

# Health Check
GET    /api/v1/health                   # Service health check
GET    /api/v1/internal/executor-health # Executor callback endpoint
```

### Security Model

Multi-layer isolation strategy:

1. **Container Layer**: `NetworkMode=none`, `CAP_DROP=ALL`, non-privileged user (UID:GID=1000:1000)
2. **Bubblewrap Layer**: Namespace isolation (PID/NET/MNT/IPC/UTS), read-only filesystem, seccomp filtering (configurable via `DISABLE_BWRAP`)
3. **Resource Limits**: CPU/Memory quotas, process limits, ulimit constraints

### Tech Stack

**Control Plane**:
- Python 3.11+ with FastAPI (0.104+), Uvicorn (0.24+), Pydantic (2.5+)
- SQLAlchemy (2.0+) with aiomysql for async database operations
- aiodocker (0.21+) for Docker SDK integration
- kubernetes (28.0+) for K8s integration
- boto3 (1.29+) for S3 operations
- httpx (0.25+) for async HTTP client
- structlog (23.2+) for structured logging

**Storage & Database**:
- MariaDB 11.2+ for session, execution, and template storage
- S3-compatible object storage (MinIO/AWS S3) for workspace files

**Web Console**:
- React 18.3+ with TypeScript
- Rsbuild build system
- Ant Design 5.26+ UI framework
- Monaco Editor for code editing
- React Router 6.26+ for navigation

**Executor**:
- Python 3.11+
- FastAPI for HTTP API
- subprocess with optional Bubblewrap isolation

## Project Structure

```
sandbox/
├── deploy/                   # Deployment configurations
│   ├── k8s/                  # Kubernetes resource manifests
│   │   ├── 00-namespace.yaml
│   │   ├── 01-configmap.yaml
│   │   ├── 05-control-plane-deployment.yaml
│   │   ├── 11-sandbox-web-deployment.yaml
│   │   └── ...
│   └── docker-compose/       # Docker Compose deployment
│       └── docker-compose.yml
│
├── sandbox_control_plane/    # FastAPI control plane service
│   ├── src/
│   │   ├── application/      # Application services (business logic)
│   │   ├── domain/           # Domain models and interfaces
│   │   ├── infrastructure/   # External dependencies (DB, Docker, S3)
│   │   ├── interfaces/       # REST API endpoints
│   │   └── shared/           # Shared utilities
│   └── tests/                # Unit, integration, and contract tests
│
├── sandbox_web/              # React web management console
│   ├── src/                  # React components and pages
│   │   ├── pages/            # Page components
│   │   ├── components/       # Reusable components
│   │   ├── services/         # API client services
│   │   └── utils/            # Utilities
│   └── package.json          # NPM dependencies
│
├── runtime/executor/          # Sandbox executor daemon
│   ├── application/          # Execution logic
│   ├── domain/               # Domain models
│   ├── infrastructure/       # External dependencies
│   ├── interfaces/           # HTTP API endpoints
│   └── Dockerfile            # Executor container image
│
├── scripts/                  # Utility scripts
├── specs/                    # Implementation specifications
└── docs/                     # Documentation
```

## Important Design Decisions

### Scheduler Behavior

The scheduler implements intelligent node selection for sessions:
- Prioritizes template affinity (nodes with cached images)
- Falls back to load-balanced cold start
- Container lifecycle follows session lifecycle (delete session = destroy container)

### Session Cleanup Strategy

Automatic cleanup of idle and expired sessions:
- Idle timeout: configurable via `IDLE_THRESHOLD_MINUTES` env var (default: 30 minutes, -1 disables)
- Max lifetime: configurable via `MAX_LIFETIME_HOURS` env var (default: 6 hours, -1 disables)
- Background task runs every 5 minutes (configurable via `CLEANUP_INTERVAL_SECONDS`)
- Settings are loaded from `.env` file via Pydantic Settings

### Timeout Control

Timeout is enforced at multiple levels:
- API level: `asyncio.wait_for()` with configurable timeout
- Event parameter: `__timeout` key in event object (default: 300s, max: 3600s)
- Daemon level: Threading with timeout

The `__timeout` parameter uses double-underscore prefix to avoid conflicts with user business parameters (e.g., a user may have their own `timeout` parameter for business logic).

### State Sync Service

On startup, the Control Plane performs a comprehensive state synchronization:
- Queries all containers with sandbox labels
- Correlates container state with database session records
- Updates session status based on actual container health
- Recovers orphaned sessions (containers running but not in DB)
- Marks dead sessions (in DB but container not running)

### Service Access URLs

After starting the system with docker-compose, the following services are available:

| Service | URL | Description |
|---------|-----|-------------|
| **API Documentation** | http://localhost:8000/docs | Swagger UI - Interactive API documentation |
| **Control Plane Console** | http://localhost:1101 | Web management interface (React app) |
| **MinIO Console** | http://localhost:9001 | S3-compatible storage management |

## Development Commands

### Control Plane

```bash
cd sandbox_control_plane

# 使用 uv 同步依赖
# uv.lock is committed for reproducible builds across all environments
uv sync

# Local development with docker-compose
docker-compose up -d

# Run tests (使用 uv)
uv run pytest tests/
uv run pytest tests/contract/
uv run pytest tests/integration/
uv run pytest tests/unit/

# Run with coverage
uv run pytest --cov=sandbox_control_plane --cov-report=html

# Code quality
uv run black sandbox_control_plane/ tests/
uv run ruff check sandbox_control_plane/ tests/
uv run mypy sandbox_control_plane/
```

### Web Console

```bash
cd sandbox_web

# Development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Lint and format
npm run lint
npm run format
```

### Executor

```bash
cd runtime/executor

# Run locally
python -m sandbox_executor.interfaces.http.main

# Build container image
docker build -t sandbox-executor .

# Run in container
docker run -p 8080:8080 sandbox-executor
```

## Configuration

Key environment variables (see `.env` file):

```bash
# Database
DATABASE_URL=mysql+aiomysql://root:password@localhost:3308/sandbox

# Docker/K8s
DOCKER_HOST=unix:///var/run/docker.sock
KUBERNETES_NAMESPACE=sandbox-runtime

# S3/MinIO
S3_ENDPOINT_URL=http://localhost:9000
S3_ACCESS_KEY_ID=minioadmin
S3_SECRET_ACCESS_KEY=minioadmin
S3_BUCKET=sandbox-workspace

# Session Cleanup (NEW)
IDLE_THRESHOLD_MINUTES=30      # -1 to disable idle cleanup
MAX_LIFETIME_HOURS=6           # -1 to disable lifetime limit
CLEANUP_INTERVAL_SECONDS=300

# Execution
DEFAULT_TIMEOUT=300
MAX_TIMEOUT=3600
DISABLE_BWRAP=true             # Disable Bubblewrap in local development

# Control Plane URL (for executor callback)
CONTROL_PLANE_URL=http://control-plane:8000
```

## Design Philosophy

- **Protocol-Driven**: All communication via standardized RESTful API
- **Cloud-Native**: Designed for Kubernetes deployment with HPA
- **Security-First**: Multiple isolation layers, defense-in-depth
- **Performance**: Async processing, connection pooling, template affinity scheduling
- **Simplicity**: Direct container creation without warm pool complexity
- **Compatibility**: AWS Lambda handler specification for easy migration

## Recent Changes

- **Session Cleanup**: Added environment variable support for `idle_timeout_minutes` and `max_lifetime_hours`
- **State Sync**: Implemented startup state synchronization service for session recovery
- **Web Console**: Added React-based web management interface
- **Executor**: Implemented HTTP-based executor with workspace file support
