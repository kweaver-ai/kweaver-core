# Changelog

All new features and capabilities added in this branch (`feature/803264`) are documented below.

## [0.3.3]

### 🚀 New Features

- **Control Plane Takeover for K8s Sessions**
  - Added startup takeover flow so a restarted or upgraded control plane can detect active K8s session pods created by a previous control-plane pod
  - Added control-plane pod identity propagation (`POD_NAME`, `POD_UID`) and executor pod owner references so recreated session pods are rebound to the current control plane
  - Added startup takeover handling for interrupted in-flight executions, marking them failed when the executor pod must be recreated during upgrade

### 🐛 Bug Fixes

- **Session Recovery Stability After Control Plane Restart**
  - Fixed K8s session recovery to resolve templates from the direct template repository during startup state sync, preventing recovery failures caused by missing template resolution
  - Fixed same-name pod recreation conflicts by waiting for stale terminating pods to be fully deleted before retrying executor pod creation
  - Refreshed recovered sessions' activity timestamps after takeover recovery so the idle cleanup task does not immediately terminate newly recovered sessions

### 🔧 Improvements

- Added direct session and execution repository support for startup state sync so takeover logic can paginate all active sessions and persist interrupted execution state changes without request-scoped dependencies
- Expanded unit coverage for K8s takeover, owner reference handling, stale pod recreation conflicts, dependency wiring, and recovery activity refresh

### 📚 Documentation

- Added PRD and design docs for control-plane and executor lifecycle binding during restart and upgrade scenarios
- Added a `sandbox_local` Helm chart with local deployment templates, component metadata, RBAC, and operational documentation for packaging and environment setup

---

*Released on 2026-04-20*

## [0.3.2]

### 🚀 New Features

- **Archive Upload and Extraction**
  - Added ZIP archive upload support for session workspaces
  - Added optional automatic archive extraction to a target workspace path
  - Added overwrite control, conflict skipping, and extraction metadata in upload responses
  - Added path-safety validation to reject invalid archive entries and traversal attempts

- **Shell Execution and Working Directory Support**
  - Added `language=shell` support across control plane, executor, and web UI
  - Added optional `working_directory` for execute and execute-sync requests
  - Added shell command normalization for accidental redundant `bash/sh` prefixes
  - Added end-to-end coverage for shell execution with relative paths and chained commands

### 📚 Documentation

- Added PRD and design docs for session archive upload and shell execution
- Updated sandbox OpenAPI documentation for archive upload and shell execution request changes

### 🔧 Improvements

- **Helm Chart Image Extraction Compatibility**
  - Standardized sandbox Helm chart image settings under the top-level `image` values tree
  - Updated control plane, web, MariaDB, MinIO, default template, and BusyBox image references to use `image.<name>.repository` and `image.<name>.tag`
  - Removed the hardcoded web initContainer BusyBox image so offline packaging tools can discover it from chart values
  - Updated Helm README, local install overrides, and component image metadata to match the new image structure

---

*Released on 2026-04-07*

## [0.3.1]

### 🚀 New Features

- **Database Name Upgrade Handling**
  - Changed the control plane default database name from `adp` to `kweaver`
  - Added startup upgrade handling to detect the legacy `adp` schema and migrate tables into `kweaver`
  - Normalized runtime database connections to `kweaver` to prevent recreating the legacy database during startup

### 🔧 Improvements

- Added unit test coverage for legacy database rename flow and runtime database URL normalization

---

*Released on 2026-04-02*

## [0.3.0]

### 🚀 New Features

- **Session-Level Python Dependency Management**
  - Added session-scoped dependency configuration and installation status tracking
  - Added background initial dependency sync during session creation
  - Added synchronous `POST /api/v1/sessions/{session_id}/dependencies/install` API
  - Added installed dependency details and error reporting in session responses

- **Runtime Executor Dependency Sync**
  - Added executor-side session config sync service for full dependency reconciliation
  - Added isolated dependency directory reset before reinstalling packages
  - Added Python executable detection for uv/virtualenv-based pip installs
  - Improved compatibility across bwrap, subprocess, and macOS seatbelt isolation backends

- **Session Management UI**
  - Added dependency install actions and status display in the session list page
  - Added frontend API types and hooks for manual dependency installation
  - Added dependency install progress and failure visibility in session details

- **Database & Upgrade Support**
  - Added `0.3.0` MariaDB and DM8 initialization SQL
  - Added startup schema migration support for upgrading existing deployments

### 🔧 Improvements

- Unified REST OpenAPI documentation into `docs/api/rest/sandbox-openapi.json`
- Extended session DTOs, persistence models, and APIs with dependency metadata
- Added integration and unit tests for initial sync, manual install, and dependency execution flows

### 📚 Documentation

- Added detailed design and PRD documents for session Python dependency management
- Reorganized repository documentation under architecture, development, operations, and product sections
- Added standalone OpenAPI description for synchronous execution endpoints

---

*Released on 2026-03-11*

## [0.2.1]

### 🐛 Bug Fixes

- **K8s Scheduler Container Resilience**
  - Changed Pod `restartPolicy` from `Never` to `Always`
  - Ensures containers automatically restart after exit (including exit code 0)
  - Fixes issue where runtime becomes unavailable when s3fs mount disconnects

### 🚀 New Features

- **Heartbeat Service Reliability**
  - Improved heartbeat service with better error handling
  - Added comprehensive test coverage for heartbeat functionality

- **State Sync Service**
  - Made control plane URL configurable via environment variable
  - Added settings initialization in state sync service

- **Callback Client**
  - Added JSON sanitization for non-compliant float values (NaN, Infinity)
  - Ensures proper JSON serialization for callback responses

- **Runtime Executor**
  - Made command execution asynchronous for better performance
  - Switched to uv for faster dependency installation in Dockerfile

- **Helm Chart Improvements**
  - Added fallback image registry support for template images
  - Added CONTROL_PLANE_URL to ConfigMap and deployment
  - Switched to Aliyun PyPI mirror for faster dependency installation

- **Session Management**
  - Added hard delete functionality with cascade removal
  - Increased string field length for ID columns
  - Set idle sessions to never be cleaned up (configurable)

- **Template Management**
  - Added template ID validation
  - Added default timeout configuration
  - Added template name update functionality

- **MCP Server**
  - Added MCP server implementation for synchronous code execution

### 🔧 Improvements

- Updated MariaDB schema definitions for sandbox tables
- Updated API documentation for OpenAPI 3.1.0 spec
- Added uv.lock for reproducible dependency management

---

*Released on 2025-03-05*

## [0.2.0]

### 🚀 New Features

#### Storage & Workspace
- **S3 Storage Integration with MinIO**
  - S3-compatible object storage backend
  - Direct file upload/download API
  - Workspace path management (`s3://bucket/sessions/{id}/`)
  - Multi-format file support

- **s3fs Workspace Mounting (Kubernetes)**
  - Container-level S3 bucket mounting via s3fs FUSE
  - Bind mount session directory to `/workspace`
  - No additional metadata database required
  - Production-ready for multi-node K8s clusters

- **Docker Volume Mounting**
  - Local development volume mounting
  - Workspace file persistence
  - Seamless S3 integration

#### Session Management
- **List Sessions API**
  - `GET /api/v1/sessions` with filtering support
  - Filter by: `status`, `template_id`, `created_after`, `created_before`
  - Pagination with `limit` and `offset` parameters
  - Optimized with database indexing

- **Session Cleanup Service**
  - Automatic cleanup of idle sessions
  - Configurable idle threshold (`IDLE_THRESHOLD_MINUTES`, default: 30)
  - Maximum lifetime enforcement (`MAX_LIFETIME_HOURS`, default: 6)
  - Background task with configurable interval (`CLEANUP_INTERVAL_SECONDS`, default: 300)
  - Set to `-1` to disable cleanup

- **State Sync Service**
  - Startup synchronization with runtime containers
  - Orphaned session recovery
  - Automatic status correction based on container health
  - Health check integration

#### Kubernetes Support
- **Helm Chart Deployment**
  - Complete Helm chart for production deployment
  - Configurable services: Control Plane, Web Console, MariaDB, MinIO
  - RBAC, ServiceAccount, and network policies
  - Values-based configuration for different environments

- **Kubernetes Scheduler**
  - Full K8s runtime support
  - Pod creation and lifecycle management
  - S3 workspace mounting via s3fs
  - Support for ephemeral and persistent session modes

- **Native K8s Manifests**
  - Standalone YAML manifests for K8s deployment
  - Namespace, ConfigMap, Secret, ServiceAccount, Role definitions
  - MariaDB and MinIO deployment configurations
  - s3fs password secret management

#### Runtime Executor
- **Python Dependency Installation**
  - Automatic `requirements.txt` installation from workspace
  - Dependencies installed to local filesystem (isolated from `/workspace`)
  - Pre-execution dependency setup
  - Support for custom package indexes (mirrors)

- **Hexagonal Architecture**
  - Clean separation: Domain, Application, Infrastructure, Interfaces layers
  - Port-Adapter pattern for external dependencies
  - Improved testability with dependency injection
  - Executor ports: IExecutorPort, ICallbackPort, IIsolationPort, IArtifactScannerPort, IHeartbeatPort, ILifecyclePort

- **Enhanced Execution Model**
  - Return value storage and retrieval
  - Metrics collection (CPU, memory, execution time)
  - Error message capturing
  - Dependency installation status tracking

#### Development Tools
- **Docker Compose Setup**
  - Complete development environment
  - One-command deployment: `docker-compose up -d`
  - Runtime node registration
  - Health check integration

- **Build System**
  - UV package manager integration
  - Configurable base image arguments
  - Mirror source support for faster Chinese downloads
  - Multi-stage Docker builds

### 📚 Documentation

- **Architecture Documentation**
  - Complete system architecture overview
  - Control Plane design and components
  - Storage architecture (MinIO + s3fs)
  - Kubernetes deployment guides

- **API Documentation**
  - RESTful API endpoint reference
  - Request/response schemas
  - Authentication and security

- **Technical Specifications**
  - Python dependency installation spec
  - S3 workspace mounting architecture
  - Kubernetes runtime design
  - Container scheduler architecture

- **Project Structure**
  - PROJECT_STRUCTURE.md with hexagonal architecture details
  - Updated architecture diagrams with Mermaid
  - Service access documentation

### 🎯 Key Capabilities

| Capability | Description |
|------------|-------------|
| **Multi-Runtime** | Docker and Kubernetes runtime support |
| **S3 Storage** | MinIO with s3fs mounting for workspace persistence |
| **Session Lifecycle** | Creation, execution, monitoring, cleanup |
| **Dependency Management** | Automatic Python package installation |
| **Health Monitoring** | Container health checks and state synchronization |
| **Production Ready** | Helm chart for K8s, Docker Compose for local |

### 📦 Configuration

#### New Environment Variables

```bash
# S3 Storage
S3_BUCKET=sandbox-workspace
S3_REGION=us-east-1
S3_ACCESS_KEY_ID=minioadmin
S3_SECRET_ACCESS_KEY=minioadmin
S3_ENDPOINT_URL=http://minio:9000

# Session Cleanup
IDLE_THRESHOLD_MINUTES=30      # -1 to disable idle cleanup
MAX_LIFETIME_HOURS=6           # -1 to disable lifetime limit
CLEANUP_INTERVAL_SECONDS=300

# Kubernetes
KUBERNETES_NAMESPACE=sandbox-runtime
KUBECONFIG=/path/to/kubeconfig

# Executor
CONTROL_PLANE_URL=http://control-plane:8000
EXECUTOR_PORT=8080
DISABLE_BWRAP=true
```

### 🔜 Service Access

| Service | URL | Credentials |
|---------|-----|-------------|
| **API Documentation** | http://localhost:8000/docs | - |
| **Control Plane API** | http://localhost:8000/api/v1 | - |
| **Web Console** | http://localhost:1101 | - |
| **MinIO Console** | http://localhost:9001 | minioadmin/minioadmin |

---

*Released on 2025-02-05*
