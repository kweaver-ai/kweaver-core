"""
REST API Interface

FastAPI application serving as the HTTP interface for the executor.
Runs inside the container and receives execution requests from Control Plane.
"""

import asyncio
import os
import platform
import time
from contextlib import asynccontextmanager
from pathlib import Path
from typing import Optional

import structlog
from fastapi import FastAPI, HTTPException, Request, status
from fastapi.responses import JSONResponse
from pydantic import BaseModel, ConfigDict, Field, ValidationError, field_validator

from executor.application.commands.execute_code import ExecuteCodeCommand
from executor.application.dto.execute_request import ExecuteRequestDTO
from executor.application.services.heartbeat_service import HeartbeatService
from executor.application.services.lifecycle_service import LifecycleService, register_lifecycle_service
from executor.application.services.session_config_sync_service import (
    InstalledDependency,
    SessionConfigSyncRequest,
    SessionConfigSyncResult,
    SessionConfigSyncService,
    SessionConfigSyncError,
    SessionConfigValidationError,
    SessionDependencyInstallError,
)
from executor.domain.value_objects import (
    ExecutionRequest as DomainExecutionRequest,
    ExecutionResult,
    ExecutionStatus,
)
from executor.infrastructure.config import settings
from executor.infrastructure.http.callback_client import CallbackClient
from executor.infrastructure.isolation.bwrap import BubblewrapRunner
from executor.infrastructure.logging import configure_logging, get_logger
from executor.infrastructure.monitoring.metrics import MetricsCollector
from executor.infrastructure.persistence.artifact_scanner import ArtifactScanner


# Configure structured logging
configure_logging(
    log_level=settings.log_level,
    log_format=settings.log_format,
)
logger = get_logger()


# Track executor startup time
startup_time = time.time()


# Request/Response Models
class ExecuteRequest(BaseModel):
    """Request model for code execution following AWS Lambda handler specification."""

    execution_id: str = Field(..., description="Unique execution identifier")
    session_id: str = Field(..., description="Session identifier")
    code: str = Field(..., description="AWS Lambda handler function code")
    language: str = Field(..., description="Programming language", pattern="^(python|javascript|shell)$")
    event: dict = Field(default_factory=dict, description="Business data passed to handler function")
    timeout: int = Field(default=300, description="Timeout in seconds", ge=1, le=3600)
    env_vars: dict = Field(default_factory=dict, description="Environment variables")
    working_directory: Optional[str] = Field(
        default=None,
        description="Optional execution directory relative to workspace root",
    )

    @field_validator("working_directory")
    @classmethod
    def validate_working_directory(cls, value: Optional[str]) -> Optional[str]:
        if value is None:
            return value

        from executor.domain.value_objects import _normalize_working_directory

        return _normalize_working_directory(value)

    model_config = ConfigDict(
        json_schema_extra={
            "examples": [
                {
                    "execution_id": "exec_20240115_test0001",
                    "session_id": "sess_test_001",
                    "code": 'def handler(event):\n    name = event.get("name", "World")\n    return {"message": f"Hello, {name}!"}',
                    "language": "python",
                    "event": {"name": "World"},
                    "timeout": 10,
                    "env_vars": {},
                    "working_directory": "src"
                },
                {
                    "execution_id": "exec_20240115_abc12345",
                    "session_id": "sess_abc123def4567890",
                    "code": 'def handler(event):\n    name = event.get("name", "World")\n    age = event.get("age", 0)\n    return {"message": f"Hello, {name}!", "age_doubled": age * 2}',
                    "language": "python",
                    "event": {"name": "Alice", "age": 25},
                    "timeout": 30,
                    "env_vars": {},
                    "working_directory": "scripts"
                }
            ]
        }
    )


class ErrorResponse(BaseModel):
    """Error response model."""

    error_code: str
    description: str
    error_detail: Optional[str] = None
    solution: Optional[str] = None


class HealthResponse(BaseModel):
    """Health check response."""

    status: str = "healthy"
    version: str = "1.0.0"
    uptime_seconds: Optional[float] = None
    active_executions: Optional[int] = None


class SessionConfigSyncRequestModel(BaseModel):
    """Internal request model for syncing session dependency configuration."""

    session_id: str = Field(..., description="Session identifier")
    language_runtime: str = Field(..., description="Language runtime, e.g. python3.11")
    python_package_index_url: str = Field(..., description="Python package index URL")
    dependencies: list[str] = Field(default_factory=list, description="Final pip spec list")
    sync_mode: str = Field(..., description="replace or merge", pattern="^(replace|merge)$")


class InstalledDependencyModel(BaseModel):
    """Installed dependency response model."""

    name: str
    version: str
    install_location: str
    install_time: str
    is_from_template: bool = False


class SessionConfigSyncResponseModel(BaseModel):
    """Internal response model for session dependency sync."""

    status: str
    installed_dependencies: list[InstalledDependencyModel] = Field(default_factory=list)
    error: str = ""
    started_at: str
    completed_at: str


# Global service instances
_execute_command: Optional[ExecuteCodeCommand] = None
_heartbeat_service: Optional[HeartbeatService] = None
_lifecycle_service: Optional[LifecycleService] = None
_callback_client: Optional[CallbackClient] = None
_metrics_collector: Optional[MetricsCollector] = None
_session_config_sync_service: Optional[SessionConfigSyncService] = None


def get_execute_command() -> ExecuteCodeCommand:
    """Get the execute command instance."""
    if _execute_command is None:
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="Executor service not initialized",
        )
    return _execute_command


def get_session_config_sync_service() -> SessionConfigSyncService:
    """Get the session config sync service instance."""
    if _session_config_sync_service is None:
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="Session config sync service not initialized",
        )
    return _session_config_sync_service


@asynccontextmanager
async def lifespan(app: FastAPI):
    """
    Lifespan context manager for startup and shutdown events.

    On startup:
    - Log startup
    - Verify bwrap availability
    - Verify workspace directory
    - Initialize services
    - Send container_ready

    On shutdown:
    - Log shutdown
    - Stop heartbeats
    - Send container_exited
    - Close connections

    Note: Uvicorn handles SIGINT/SIGTERM and triggers this shutdown automatically.
    """
    global _execute_command, _heartbeat_service, _lifecycle_service, _callback_client, _metrics_collector, _session_config_sync_service

    # Environment variables
    workspace_path = Path(os.environ.get("WORKSPACE_PATH", str(settings.workspace_path)))
    control_plane_url = os.environ.get("CONTROL_PLANE_URL", settings.control_plane_url)
    internal_api_token = os.environ.get("INTERNAL_API_TOKEN", settings.internal_api_token)
    container_id = os.environ.get("CONTAINER_ID", "unknown")
    pod_name = os.environ.get("POD_NAME", "unknown")
    executor_port = int(os.environ.get("EXECUTOR_PORT", str(settings.executor_port)))

    # Detect operating system
    is_macos = platform.system() == "Darwin"
    is_linux = platform.system() == "Linux"

    logger.info(
        "Executor starting",
        version="1.0.0",
        port=executor_port,
        workspace_path=str(workspace_path),
        control_plane_url=control_plane_url,
        container_id=container_id,
        pod_name=pod_name,
        platform=platform.system(),
        is_development_mode=is_macos,  # macOS is considered development mode
    )

    # T055 [US3]: Add bwrap availability check in startup
    # Only check Bubblewrap on Linux (not available on macOS)
    if is_linux:
        try:
            from executor.infrastructure.isolation.bwrap import check_bwrap_available, get_bwrap_version

            check_bwrap_available()
            bwrap_version = get_bwrap_version()
            logger.info("Bubblewrap verified", version=bwrap_version)

        except RuntimeError as e:
            logger.error("Bubblewrap check failed", error=str(e))
            # In production on Linux, this should exit with error
            logger.warning("Continuing without Bubblewrap (development mode)")
    else:
        logger.info("Skipping Bubblewrap check on macOS (Bubblewrap is Linux-only)")
        logger.warning("Code execution features will be limited on macOS")

    # T056 [US3]: Add workspace availability check
    # Create workspace if it doesn't exist
    if not workspace_path.exists():
        try:
            logger.info("Creating workspace directory", workspace_path=str(workspace_path))
            workspace_path.mkdir(parents=True, exist_ok=True)
        except OSError as e:
            logger.warning("Failed to create workspace directory", workspace_path=str(workspace_path), error=str(e))
            logger.warning("Continuing without workspace (development mode)")
    else:
        # Check workspace is writable
        if not os.access(workspace_path, os.W_OK):
            logger.warning("Workspace directory is not writable", workspace_path=str(workspace_path))
            logger.warning("Continuing with read-only workspace (development mode)")
        else:
            logger.info("Workspace directory verified", workspace_path=str(workspace_path))

    # Initialize infrastructure services
    # Linux uses Bubblewrap, macOS uses Seatbelt sandbox
    # Check if bwrap is disabled via environment variable
    disable_bwrap = os.environ.get("DISABLE_BWRAP", "false").lower() == "true"

    if is_linux and not disable_bwrap:
        try:
            bwrap_runner = BubblewrapRunner(workspace_path=workspace_path)
            logger.info("Using BubblewrapRunner for Linux isolation")
        except Exception as e:
            logger.warning("Failed to initialize BubblewrapRunner", error=str(e))
            bwrap_runner = None
    else:
        # macOS: Use Seatbelt sandbox (sandbox-exec)
        # or Linux with DISABLE_BWRAP=true
        if disable_bwrap:
            logger.info("Bubblewrap disabled via DISABLE_BWRAP environment variable")
            bwrap_runner = None
        elif is_macos:
            try:
                from executor.infrastructure.isolation.macseatbelt import MacSeatbeltRunner
                bwrap_runner = MacSeatbeltRunner(workspace_path=workspace_path)
                logger.info("Using MacSeatbeltRunner with sandbox-exec", sandbox_version=bwrap_runner.get_version())
            except Exception as e:
                logger.error("Failed to initialize MacSeatbeltRunner", error=str(e))
                bwrap_runner = None
        else:
            bwrap_runner = None

    # Fallback to SubprocessRunner if no isolation is available
    if bwrap_runner is None:
        from executor.infrastructure.isolation.subprocess import SubprocessRunner
        bwrap_runner = SubprocessRunner(workspace_path=workspace_path)
        logger.warning("Using SubprocessRunner - NO SECURITY ISOLATION (development mode only)")

    # ArtifactScanner doesn't need workspace_path in constructor
    artifact_scanner = ArtifactScanner()

    metrics_collector = MetricsCollector()

    # Initialize callback client
    callback_client = CallbackClient(
        control_plane_url=control_plane_url,
        api_token=internal_api_token,
    )
    _callback_client = callback_client

    # Initialize application services
    heartbeat_service = HeartbeatService(
        callback_port=callback_client,
        interval=5.0,
    )
    _heartbeat_service = heartbeat_service

    lifecycle_service = LifecycleService(
        callback_port=callback_client,
        executor_port=executor_port,
        heartbeat_port=heartbeat_service,
    )
    _lifecycle_service = lifecycle_service
    register_lifecycle_service(lifecycle_service)

    # Register signal handlers
    # Note: Uvicorn handles SIGINT/SIGTERM by default and will trigger lifespan shutdown
    # We don't need custom signal handlers - let Uvicorn handle it
    logger.info("Signal handlers will be managed by Uvicorn")

    # Initialize execute command
    execute_command = ExecuteCodeCommand(
        isolation_port=bwrap_runner,
        artifact_scanner_port=artifact_scanner,
        callback_port=callback_client,
        heartbeat_port=heartbeat_service,
        workspace_path=workspace_path,
        control_plane_url=control_plane_url,
    )
    _execute_command = execute_command
    _session_config_sync_service = SessionConfigSyncService(
        install_path=Path(settings.dependency_install_path),
        pip_cache_path=Path(settings.pip_cache_path),
    )

    logger.info("Executor startup complete")

    # Send container_ready signal in background after HTTP server is ready
    # We use asyncio.create_task() to run this in the background so it doesn't
    # block the HTTP server from starting. The signal will be sent after we verify
    # the HTTP server is ready by checking the health endpoint.
    async def send_ready_signal():
        """Send container_ready signal after HTTP server is ready.

        Uses a health check to verify the HTTP server is actually ready
        before signaling the Control Plane, avoiding race conditions where
        the Control Plane tries to connect before the server is listening.
        """
        import asyncio
        import httpx

        health_url = f"http://localhost:{executor_port}/health"
        logger.info("Waiting for HTTP server to be ready...", health_url=health_url)

        # Wait for HTTP server to be ready with exponential backoff
        # Start with 50ms, max 500ms, total timeout ~3 seconds
        base_delay = 0.05
        max_delay = 0.5
        max_attempts = 30

        for attempt in range(max_attempts):
            try:
                async with httpx.AsyncClient() as client:
                    response = await client.get(
                        health_url,
                        timeout=httpx.Timeout(1.0, connect=0.5)
                    )
                    if response.status_code == 200:
                        logger.info(
                            "HTTP server health check passed",
                            attempt=attempt + 1,
                            response_time=response.elapsed.total_seconds()
                        )
                        break
            except (httpx.ConnectError, httpx.ConnectTimeout, OSError) as e:
                # Server not ready yet, wait with exponential backoff
                delay = min(base_delay * (2 ** attempt), max_delay)
                logger.debug(
                    "HTTP server not ready, retrying...",
                    attempt=attempt + 1,
                    delay=delay,
                    error=str(e)
                )
                await asyncio.sleep(delay)
        else:
            # All attempts failed - log warning but still try to send signal
            logger.warning(
                "HTTP server health check failed after all attempts",
                max_attempts=max_attempts
            )

        # Send container_ready signal to Control Plane
        logger.info("HTTP server ready, sending container_ready signal")
        await lifecycle_service.send_container_ready()

    import asyncio
    asyncio.create_task(send_ready_signal())
    logger.info("Container ready signal scheduled to run in background")

    yield

    # Shutdown
    logger.info("Executor shutting down")

    # Stop all heartbeats
    await heartbeat_service.stop_all()

    # Send container_exited
    try:
        await lifecycle_service.shutdown()
        logger.info("Container exited signal sent")
    except Exception as e:
        logger.error("Failed to send container_exited", error=str(e))

    # Close callback client
    await callback_client.close()

    logger.info("Executor shutdown complete")


def create_app() -> FastAPI:
    """
    Create and configure the FastAPI application.

    Returns:
        Configured FastAPI application instance
    """
    app = FastAPI(
        title="Sandbox Executor API",
        description="HTTP API for executing code in secure sandbox environments",
        version="1.0.0",
        lifespan=lifespan,
    )

    # Exception handlers
    @app.exception_handler(ValidationError)
    async def validation_exception_handler(request: Request, exc: ValidationError):
        """Handle Pydantic validation errors."""
        logger.warning(
            "Request validation failed",
            errors=exc.errors(),
            path=request.url.path,
        )
        error_response = ErrorResponse(
            error_code="Executor.ValidationError",
            description="Request validation failed",
            error_detail=str(exc.errors()),
            solution="Check request format and required fields",
        )
        return JSONResponse(
            status_code=status.HTTP_400_BAD_REQUEST,
            content=error_response.model_dump(),
        )

    @app.exception_handler(ValueError)
    async def value_error_handler(request: Request, exc: ValueError):
        """Handle ValueError exceptions."""
        logger.warning("Value error", error=str(exc), path=request.url.path)
        error_response = ErrorResponse(
            error_code="Executor.ValidationError",
            description="Invalid value provided",
            error_detail=str(exc),
            solution="Check request parameters",
        )
        return JSONResponse(
            status_code=status.HTTP_400_BAD_REQUEST,
            content=error_response.model_dump(),
        )

    @app.exception_handler(Exception)
    async def general_exception_handler(request: Request, exc: Exception):
        """Handle unexpected exceptions."""
        logger.error(
            "Unexpected error",
            error=str(exc),
            error_type=type(exc).__name__,
            path=request.url.path,
        )
        error_response = ErrorResponse(
            error_code="Executor.InternalError",
            description="Executor encountered an unexpected error",
            error_detail=str(exc),
            solution="Check executor logs for details",
        )
        return JSONResponse(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            content=error_response.model_dump(),
        )

    @app.get(
        "/",
        include_in_schema=False,
    )
    async def root():
        """Root endpoint with API information."""
        return {
            "service": "sandbox-executor",
            "version": "1.0.0",
            "docs": "/docs",
            "health": "/health",
            "execute": "/execute",
        }

    @app.get(
        "/health",
        response_model=HealthResponse,
        responses={
            200: {"description": "Executor is healthy"},
            503: {"model": ErrorResponse, "description": "Executor is unhealthy"},
        },
        summary="Health check",
        description="Returns executor health status for readiness probes and load balancers",
        tags=["health"],
    )
    async def health_check():
        """
        Health check endpoint for container readiness probes.

        Returns:
            - status: "healthy" if all checks pass, "unhealthy" otherwise
            - version: Executor version
            - uptime_seconds: Time since executor started
            - active_executions: Number of currently active executions

        ## Health checks:
        - HTTP API is listening
        - Workspace directory is accessible
        - Isolation mechanism is available:
          - Linux: Bubblewrap binary
          - macOS: sandbox-exec binary (Seatbelt)
        """
        try:
            # Calculate uptime
            uptime = time.time() - startup_time

            # Get active execution count
            active_count = _execute_command.get_active_count() if _execute_command else 0

            # Check isolation availability based on platform
            is_macos = platform.system() == "Darwin"
            if is_macos:
                # macOS: Check sandbox-exec availability
                from executor.infrastructure.isolation.macseatbelt import check_sandbox_available
                try:
                    check_sandbox_available()
                except RuntimeError:
                    logger.warning("Health check failed: sandbox-exec not available")
                    raise HTTPException(
                        status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                        detail={
                            "status": "unhealthy",
                            "reason": "sandbox-exec binary not found",
                        },
                    )
            else:
                # Linux: Check bwrap availability
                from executor.infrastructure.isolation.bwrap import check_bwrap_available
                try:
                    check_bwrap_available()
                except RuntimeError:
                    logger.warning("Health check failed: bwrap not available")
                    raise HTTPException(
                        status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                        detail={
                            "status": "unhealthy",
                            "reason": "Bubblewrap binary not found",
                        },
                    )

            # Check workspace availability (only on Linux)
            is_macos = platform.system() == "Darwin"
            if not is_macos:
                # Get workspace path from environment or settings
                workspace_path_env = os.environ.get("WORKSPACE_PATH", str(settings.workspace_path))

                # If workspace is S3 path, check local /workspace directory instead
                # S3 paths are handled by the artifact storage layer, not filesystem checks
                if workspace_path_env.startswith("s3://"):
                    check_path = Path("/workspace")
                    logger.debug("Workspace is S3 path, checking local /workspace directory")
                else:
                    check_path = Path(workspace_path_env)

                workspace_accessible = check_path.exists() and os.access(check_path, os.W_OK)

                if not workspace_accessible:
                    logger.warning("Health check failed: workspace not accessible", workspace_path=str(check_path))
                    raise HTTPException(
                        status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                        detail={
                            "status": "unhealthy",
                            "reason": f"Workspace directory {check_path} is not accessible",
                        },
                    )
                logger.debug("Workspace health check passed", workspace_path=str(check_path))
            else:
                logger.debug("Skipping workspace health check on macOS")

            return HealthResponse(
                status="healthy",
                version="1.0.0",
                uptime_seconds=uptime,
                active_executions=active_count,
            )

        except HTTPException:
            raise
        except Exception as e:
            logger.error("Health check failed", error=str(e))
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail={
                    "status": "unhealthy",
                    "reason": f"Health check failed: {str(e)}",
                },
            )

    @app.post(
        "/execute",
        response_model=dict,
        responses={
            200: {"description": "Execution completed"},
            400: {"model": ErrorResponse, "description": "Invalid request"},
            500: {"model": ErrorResponse, "description": "Internal error"},
            503: {"model": ErrorResponse, "description": "Executor queue full"},
        },
        summary="Execute code in sandbox",
        description="Executes code in a sandboxed environment and returns the result",
        tags=["execution"],
    )
    async def execute_endpoint(request: ExecuteRequest, http_request: Request) -> dict:
        """
        Execute code in a sandboxed environment.

        - Accepts ExecutionRequest with code, language, timeout, and event
        - Executes code in isolation (with or without Bubblewrap)
        - Returns ExecutionResult with status, output, metrics, and artifacts
        - Supports Python Lambda handlers, JavaScript, and Shell scripts

        ## Validation
        - code size ≤ 1MB
        - timeout between 1-3600 seconds
        - language in {python, javascript, shell}
        - execution_id pattern: exec_[0-9]{8}_[a-z0-9]{8}
        """
        logger.info(
            "Execution request received",
            execution_id=request.execution_id,
            language=request.language,
            timeout=request.timeout,
            code_length=len(request.code),
            event_keys=list(request.event.keys()) if request.event else [],
            working_directory=request.working_directory,
        )

        command = get_execute_command()

        # Convert to domain request
        domain_request = DomainExecutionRequest(
            execution_id=request.execution_id,
            session_id=request.session_id,
            code=request.code,
            language=request.language,
            event=request.event,
            timeout=request.timeout,
            env_vars=request.env_vars,
            working_directory=request.working_directory,
        )

        # Execute in background - return immediately
        # The command will execute asynchronously and report results via callback
        asyncio.create_task(command.execute(domain_request))

        return {
            "execution_id": request.execution_id,
            "status": "PENDING",
            "message": "Execution submitted",
        }

    @app.post(
        "/internal/session-config/sync",
        response_model=SessionConfigSyncResponseModel,
        responses={
            200: {"description": "Session config synchronized"},
            400: {"model": ErrorResponse, "description": "Invalid sync request"},
            422: {"model": ErrorResponse, "description": "Dependency installation failed"},
            500: {"model": ErrorResponse, "description": "Internal error"},
        },
        summary="Sync session dependency configuration",
        description="Internal endpoint used by control plane to synchronize Python dependencies",
        tags=["internal"],
    )
    async def sync_session_config_endpoint(
        request: SessionConfigSyncRequestModel,
    ) -> SessionConfigSyncResponseModel:
        """Synchronize session-level Python dependency configuration."""
        service = get_session_config_sync_service()

        try:
            result = await service.sync(
                SessionConfigSyncRequest(
                    session_id=request.session_id,
                    language_runtime=request.language_runtime,
                    python_package_index_url=request.python_package_index_url,
                    dependencies=request.dependencies,
                    sync_mode=request.sync_mode,
                )
            )
        except SessionConfigValidationError as exc:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail={
                    "error_code": "Executor.ValidationError",
                    "description": "Invalid session config sync request",
                    "error_detail": str(exc),
                },
            ) from exc
        except SessionDependencyInstallError as exc:
            raise HTTPException(
                status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
                detail={
                    "error_code": "Executor.DependencyInstallError",
                    "description": "Dependency installation failed",
                    "error_detail": str(exc),
                },
            ) from exc
        except SessionConfigSyncError as exc:
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail={
                    "error_code": "Executor.SessionConfigSyncError",
                    "description": "Session config sync failed",
                    "error_detail": str(exc),
                },
            ) from exc

        return _map_session_config_sync_result(result)

    return app


# Create app instance for import
app = create_app()


# CLI entry point
def main():
    """Main entry point for running the executor."""
    import uvicorn

    port = int(os.environ.get("EXECUTOR_PORT", str(settings.executor_port)))
    host = os.environ.get("EXECUTOR_HOST", "0.0.0.0")

    uvicorn.run(
        "executor.interfaces.http.rest:app",
        host=host,
        port=port,
        log_level=settings.log_level.lower(),
        reload=False,
    )


if __name__ == "__main__":
    main()


def _map_session_config_sync_result(
    result: SessionConfigSyncResult,
) -> SessionConfigSyncResponseModel:
    return SessionConfigSyncResponseModel(
        status=result.status,
        installed_dependencies=[
            InstalledDependencyModel(
                name=dep.name,
                version=dep.version,
                install_location=dep.install_location,
                install_time=dep.install_time.isoformat(),
                is_from_template=dep.is_from_template,
            )
            for dep in result.installed_dependencies
        ],
        error=result.error,
        started_at=result.started_at.isoformat(),
        completed_at=result.completed_at.isoformat(),
    )
