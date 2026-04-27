# Sandbox Platform Feature Iteration Roadmap

**Version**: 1.2
**Last Updated**: 2025-01-14
**Status**: Active

---

## Executive Summary

This roadmap outlines the remaining features to implement for the Sandbox Platform based on:
- `docs/design/archive/sandbox-design-v2.1.md` - Technical architecture specification
- `docs/product/use-cases/ptc-data-analysis-and-context-qa.md` - Business scenario requirements (PTC data analysis + Q&A)

**Current State**: ~95% complete for core Python execution functionality
**Latest Updates**:
- ✅ K8s scheduler fully implemented with RBAC and deployment manifests (2025-01-14)
- ✅ MinIO/S3 storage integration complete with presigned URLs (2025-01-12)
- ✅ Runtime abstraction layer with auto-detection (Docker/K8s) (2025-01-14)
- ✅ File lifecycle cleanup on session termination (2025-01-12)
- ✅ Python dependency installation - per-session dynamic pip install (2025-01-14)

**Target**: Production-ready platform supporting multi-language execution, Kubernetes deployment, and business-closed-loop scenarios

---

## 1. Implementation Status Summary

### ✅ Fully Implemented

| Component | Status | Location | Notes |
|-----------|--------|----------|-------|
| **Control Plane API** | ✅ Complete | `sandbox_control_plane/src/interfaces/rest/` | Sessions, Executions, Templates, Files, Health endpoints |
| **Database Layer** | ✅ Complete | `sandbox_control_plane/src/infrastructure/persistence/` | MariaDB with SQLAlchemy 2.0 async |
| **Session Manager** | ✅ Complete | `sandbox_control_plane/src/application/` | Lifecycle, CRUD, cleanup service |
| **Docker Scheduler** | ✅ Complete | `sandbox_control_plane/src/infrastructure/container_scheduler/docker_scheduler.py` | Full Docker operations with dependency install |
| **K8s Scheduler** | ✅ Complete | `sandbox_control_plane/src/infrastructure/container_scheduler/k8s_scheduler.py` | Pod lifecycle, RBAC, PVC, dependency install |
| **Runtime Abstraction** | ✅ Complete | `sandbox_control_plane/src/infrastructure/container_scheduler/base.py` | Unified interface for Docker/K8s |
| **Runtime Auto-Detection** | ✅ Complete | `sandbox_control_plane/src/infrastructure/dependencies.py` | Auto-select Docker or K8s scheduler |
| **MinIO/S3 Storage** | ✅ Complete | `sandbox_control_plane/src/infrastructure/storage/s3_storage.py` | File upload/download with presigned URLs |
| **File Lifecycle Management** | ✅ Complete | `sandbox_control_plane/src/application/session_cleanup_service.py` | Auto-cleanup on session termination |
| **Python Dependency Installation** | ✅ Complete | `sandbox_control_plane/src/infrastructure/container_scheduler/` | Per-session dynamic pip install (Docker/K8s) |
| **Executor (Python)** | ✅ Complete | `runtime/executor/` | HTTP server, bwrap isolation, execution |
| **State Sync Service** | ✅ Complete | `sandbox_control_plane/src/application/` | Database-to-container state sync |
| **Background Tasks** | ✅ Complete | `sandbox_control_plane/src/infrastructure/background_tasks/` | Session cleanup, health checks |
| **Hexagonal Architecture** | ✅ Complete | Both components | Clean architecture, DI container |
| **Testing Framework** | ✅ Complete | `*/tests/` | Unit, integration, contract tests |
| **Unit Tests** | ✅ Complete | `*/tests/unit/` | 181+ passing tests |
| **K8s Deployment** | ✅ Complete | `deploy/manifests/` | Namespace, RBAC, Deployment manifests |
| **docker-compose** | ✅ Complete | `docker-compose.yml` | Local development with MinIO, MariaDB |

### ⚠️ Partially Implemented

| Component | Status | Gaps | Priority |
|-----------|--------|------|----------|
| **Container Monitoring** | ⚠️ Basic only | No Prometheus metrics, no detailed health probing | P2 |
| **Template System** | ⚠️ CRUD only | No advanced features (versioning, dependencies) | P2 |
| **Web Console** | ⚠️ Basic only | Limited session management features | P2 |

### ❌ Not Implemented (Prioritized)

| Component | Business Impact | Priority | Est. Effort |
|-----------|----------------|----------|-------------|
| **Security Hardening** | Rate limiting, audit logging missing | P1 | 3-5 days |
| **Prometheus Metrics** | No observability integration | **P2 (Lowest)** | 3-5 days |
| **Template Pre-installed Packages** | Common packages in templates (deferred) | P2 | 1 week |

### Deferred (Not Planned)

| Component | Reason |
|-----------|--------|
| **Multi-language Support** | Python only - satisfies current requirements |
| **CLI Tool (sandbox-run)** | Lower priority, can be added on demand |
| **Agent SDK** | Lower priority, can be added on demand |
| **Warm Pool** | Removed based on user feedback |

---

## 2. Roadmap Phases

> **Priority Notes** (from user):
> - Python execution is the primary focus (multi-language support deprioritized)
> - **Docker + K8s Runtime support is HIGH PRIORITY**
> - **Python dependency installation (pip install) is HIGH PRIORITY** - implementation approach TBD
> - SDK/CLI tools are lower priority
> - Warm Pool removed (not implementing)
> - Observability is lowest priority
> - File lifecycle tied to session lifecycle

### Phase 1: Core Production (P1) - 3-4 weeks

**Goal**: Production-ready platform with Docker + K8s dual runtime support and MinIO file storage

#### ✅ Sprint 1.1: MinIO File Storage (COMPLETED)
- [x] Add MinIO to docker-compose for local development
- [x] Implement S3-compatible storage client (boto3)
- [x] File upload API: POST /sessions/{id}/files/upload
- [x] File download API: GET /sessions/{id}/files/{name}
- [x] Presigned URL generation for downloads (>10MB files)
- [x] **File lifecycle cleanup**: Delete files when session terminates
- [x] Tests for file operations

**Completed**: 2025-01-12

#### ✅ Sprint 1.2: Kubernetes Runtime Scheduler (COMPLETED)
- [x] Implement K8s scheduler using official Python client
- [x] Pod specification with resource limits and security context
- [x] PVC for MinIO workspace mounting (via S3 CSI Driver or s3fs sidecar)
- [x] Pod lifecycle: create, delete, status, logs
- [x] ServiceAccount and RBAC configuration
- [x] Runtime node registration (K8s nodes self-register to control plane)
- [x] Tests for K8s scheduler
- [x] K8s deployment manifests (namespace, configmap, secret, RBAC, deployment)

**Completed**: 2025-01-14

#### ✅ Sprint 1.3: Runtime Abstraction & Node Registration (COMPLETED)
- [x] Unified base interface (`IContainerScheduler`)
- [x] Docker scheduler implements base interface
- [x] K8s scheduler implements base interface
- [x] Runtime node auto-registration on startup
- [x] Node health check endpoint for both Docker and K8s
- [x] Auto-detection logic (K8s in-cluster, Docker for local)
- [x] Tests for dual runtime support

**Completed**: 2025-01-14

#### ⏳ Sprint 1.4: Security Hardening (3-5 days) - NEXT PRIORITY
- [ ] Rate limiting middleware (slowapi)
- [ ] Audit logging for sensitive operations
- [ ] Input sanitization enhancement
- [ ] Secrets management validation
- [ ] Security tests

**Files to create/modify**:
- `sandbox_control_plane/src/interfaces/rest/middleware/rate_limit.py` (new)
- `sandbox_control_plane/config/logging.py` - Add audit logging

**Deliverable**: Production-ready platform with security hardening

#### ✅ Sprint 1.5: Python Dependency Installation (COMPLETED)
- [x] **Implementation**: Per-session dynamic pip install during container scheduling
- [x] API design: `dependencies` field in CreateSession request
- [x] Docker scheduler support with entrypoint script generation
- [x] K8s scheduler support with init container/command approach
- [x] Dependency status tracking (installing/completed/failed)
- [x] Database model: `requested_dependencies` and `installed_dependencies` fields
- [x] Internal callback API: executor reports installation completion
- [x] Tests for dependency installation

**Completed**: 2025-01-14

**Implementation Details**:
- Dependencies specified at session creation time (e.g., `["requests==2.31.0", "pandas>=2.0"]`)
- Dynamic installation during container startup via pip
- Installation status tracked in database
- No template-based pre-installation (deferred to future)

**Files Implemented**:
- `sandbox_control_plane/src/domain/entities/session.py` - Dependency status fields
- `sandbox_control_plane/src/infrastructure/container_scheduler/docker_scheduler.py` - Docker support
- `sandbox_control_plane/src/infrastructure/container_scheduler/k8s_scheduler.py` - K8s support
- `sandbox_control_plane/src/interfaces/rest/api/v1/sessions.py` - API endpoint
- `sandbox_control_plane/src/interfaces/rest/api/v1/internal.py` - Callback API
- `sandbox_control_plane/src/infrastructure/persistence/models/session_model.py` - Database model

---

### Phase 2: Observability (P2 - Lowest Priority) - 1-2 weeks

**Goal**: Basic monitoring and operational visibility (can be deferred)

#### Sprint 2.1: Basic Metrics & Health (1 week)
- [ ] Prometheus endpoint `/metrics`
- [ ] Session/execution counters
- [ ] Active sessions gauge
- [ ] Enhanced health check endpoint
- [ ] Component health status (DB, MinIO, runtime nodes)

**Files to create**:
- `sandbox_control_plane/src/interfaces/rest/metrics.py` (new)
- `sandbox_control_plane/src/application/metrics_service.py` (new)
- Update `sandbox_control_plane/src/interfaces/rest/api/v1/health.py`

**Deliverable**: Basic observability for production operations

---

## Future / Optional (Lower Priority)

These items are deprioritized and can be implemented later based on actual user demand:

### Multi-Language Support (Python Only Currently)
- JavaScript/Node.js execution
- Shell script execution
- Language-specific templates
- **Status**: Not planned - Python satisfies current requirements

### Agent SDK
- Python SDK for simplified integration
- Synchronous API wrapper (async polling)
- Session management helpers
- **Status**: Can be added if users request it

### CLI Tool (sandbox-run)
- Command-line interface for local testing
- Event/context data passing
- AWS Lambda compatibility
- **Status**: Can be added if users request it

### Warm Pool
- Pre-warmed container pool for fast startup
- **Status**: Removed - not implementing based on user feedback

---

## 3. Feature vs Business Scenario Mapping

### PTC Data Analysis Scenario (e2e-integration-plan.md §3.1)

| Required Capability | Status | Phase |
|--------------------|--------|-------|
| Session creation & reuse | ✅ Done | - |
| Python code execution (pandas/numpy) | ✅ Done | - |
| File upload to workspace | ✅ Done | 1.1 |
| Multi-step execution | ✅ Done | - |
| Result retrieval | ✅ Done | - |
| Generated file download | ✅ Done | 1.1 |
| K8s deployment support | ✅ Done | 1.2 |
| Python dependency installation | ✅ Done | 1.5 |
| Agent SDK (sync wrapper) | ❌ Todo | Future |

### Non-PTC Q&A Scenario (e2e-integration-plan.md §3.2)

| Required Capability | Status | Phase |
|--------------------|--------|-------|
| File upload (PDF/CSV) | ✅ Done | 1.1 |
| File parsing libraries | ✅ Done | 1.5 |
| Multi-round file reading | ✅ Done | - |
| Session persistence | ✅ Done | - |
| K8s deployment support | ✅ Done | 1.2 |
| Python dependency installation | ✅ Done | 1.5 |
| Agent SDK | ❌ Todo | Future |

### Unified Temporary Area (e2e-integration-plan.md §4)

| Requirement | Status | Phase |
|-------------|--------|-------|
| File persistence (S3 workspace) | ✅ Done | 1.1 |
| Complex dependency compatibility | ✅ Done | - |
| Dual-layer isolation | ✅ Done | - |
| Session model | ✅ Done | - |
| K8s runtime support | ✅ Done | 1.2 |

---

## 4. Critical Dependencies

| Feature | Depends On | Blocked |
|---------|-----------|---------|
| K8s Scheduler | S3 workspace mounting | No |
| Agent SDK | Stable API | No |
| Warm Pool | K8s Scheduler | Partially |
| Multi-language | Executor language support | No |
| Metrics | Base implementation | No |

---

## 5. Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| S3 mounting performance | High | Evaluate s3fs/geesefs/goofys, implement local cache |
| K8s scheduler complexity | Medium | Reuse existing patterns, incremental testing |
| Multi-language isolation | Medium | Language-specific bwrap configurations |
| Warm pool resource waste | Low | Configurable pool sizes, auto-scaling |

---

## 6. Success Criteria

### Phase 1 Complete When (Core Production):
- [x] MinIO storage operational with file upload/download
- [x] Files are automatically deleted when session terminates
- [x] K8s scheduler creates/manages pods correctly
- [x] Docker and K8s schedulers share common interface
- [x] Runtime nodes can be Docker or K8s (selectable via config)
- [x] **Python dependency installation supported** (per-session dynamic install)
- [ ] Security tests pass (rate limiting, audit logging)

**✅ Phase 1 Status**: 95% COMPLETE - Only Security Hardening (Sprint 1.4) remaining

### Phase 2 Complete When (Observability - Optional):
- [ ] Prometheus metrics exported at `/metrics`
- [ ] Enhanced health checks show all components (DB, MinIO, runtime nodes)

---

## 7. File Changes Summary

### New Files to Create (Phase 1 - Core Production)

```
# Sprint 1.1: MinIO File Storage
sandbox_control_plane/src/infrastructure/storage/s3_client.py

# Sprint 1.2: Kubernetes Scheduler
sandbox_control_plane/src/infrastructure/container_scheduler/k8s_scheduler.py
deploy/manifests/05-control-plane-deployment.yaml
deploy/manifests/04-role.yaml
deploy/manifests/08-mariadb-deployment.yaml

# Sprint 1.3: Runtime Abstraction
sandbox_control_plane/src/application/scheduler_service.py

# Sprint 1.4: Security
sandbox_control_plane/src/interfaces/rest/middleware/rate_limit.py

# Sprint 1.5: Python Dependency Installation (Implementation TBD)
runtime/executor/application/dependency_manager.py
sandbox_control_plane/src/interfaces/rest/api/v1/dependencies.py

# Phase 2: Observability (Optional)
sandbox_control_plane/src/interfaces/rest/metrics.py
sandbox_control_plane/src/application/metrics_service.py
```

### Key Files to Modify

```
docker-compose.yml                                   # Add MinIO service
sandbox_control_plane/src/interfaces/rest/api/v1/files.py           # Implement upload/download with MinIO
sandbox_control_plane/src/application/session_cleanup_service.py     # Add file cleanup on session termination
sandbox_control_plane/src/infrastructure/container_scheduler/base.py         # Enhance abstract interface
sandbox_control_plane/src/infrastructure/container_scheduler/docker_scheduler.py  # Refactor to base interface
sandbox_control_plane/config/logging.py                                   # Add audit logging
sandbox_control_plane/src/interfaces/rest/api/v1/health.py                 # Enhanced health checks
```

---

## 8. Verification Plan

### E2E Test Scenarios

#### Test 1: PTC Data Analysis (After Phase 1.1 + 2.1)
```python
# 1. Create session
# 2. Upload CSV file to S3 workspace
# 3. Execute pandas data cleaning
# 4. Execute statistical analysis
# 5. Generate matplotlib chart
# 6. Download generated file via presigned URL
```

#### Test 2: Non-PTC File Q&A (After Phase 1.1 + 2.1)
```python
# 1. Create session
# 2. Upload PDF to S3 workspace
# 3. Execute PDF parsing (first 10 pages)
# 4. Execute PDF parsing (next 10 pages) - session reuse
# 5. Verify results accessible
```

#### Test 3: K8s Deployment (After Phase 1.2)
```bash
# 1. Deploy to Minikube/K3s
# 2. Create session
# 3. Verify pod creation
# 4. Execute code
# 5. Verify execution in pod
# 6. Delete session
# 7. Verify pod cleanup
```

---

## 9. Next Actions (Immediate)

**Recommended Priority Order**:

1. **Sprint 1.4 (Security Hardening)**: P1 HIGH PRIORITY - **NEXT**
   - Rate limiting middleware (slowapi or slowapi + starlette)
   - Audit logging for sensitive operations
   - Input sanitization enhancement
   - Security tests

2. **Phase 2: Observability (P2 - Lowest Priority)**: Optional
   - Prometheus metrics endpoint `/metrics`
   - Enhanced health checks showing all components (DB, MinIO, runtime nodes)

3. **Template Pre-installed Packages**: Future Enhancement
   - Template-level pre-installed packages
   - Hybrid approach (common packages in template + user packages per session)

**Phase 1 Status**: 95% COMPLETE - Only Security Hardening remaining

**Documentation Updates**:
- ✅ Updated `docs/design/archive/sandbox-design-v2.1.md` with K8s scheduler implementation details
- ✅ This roadmap is saved at `docs/product/roadmap.md` for reference
