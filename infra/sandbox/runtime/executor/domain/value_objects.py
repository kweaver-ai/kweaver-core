"""
Execution Value Objects

Immutable value objects for execution-related concepts.
Merged from executor/src/models.py to provide complete domain model.
"""

from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from typing import List, Optional, Dict, Any, Literal
from pathlib import Path


class ExecutionStatus(str, Enum):
    """Status of a code execution."""

    PENDING = "pending"
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"
    TIMEOUT = "timeout"
    CRASHED = "crashed"
    # Additional status values for compatibility
    SUCCESS = "success"
    ERROR = "error"


class ArtifactType(str, Enum):
    """Type of artifact generated during execution."""

    OUTPUT = "output"
    LOG = "log"
    ARTIFACT = "artifact"
    TEMP = "temp"


class ExitReason(str, Enum):
    """Reason for container exit."""

    NORMAL = "normal"
    SIGTERM = "sigterm"
    SIGKILL = "sigkill"
    OOM_KILLED = "oom_killed"
    ERROR = "error"


@dataclass(frozen=True)
class Artifact:
    """
    Represents a file generated during execution.

    Merged from ArtifactMetadata to include all fields.
    Also adds frozen=True for immutability (hexagonal architecture).

    Attributes:
        path: Relative path from workspace root
        size: File size in bytes
        mime_type: MIME type of the file
        type: Category of artifact
        created_at: Timestamp when file was created
        checksum: Optional SHA256 checksum
        download_url: Optional pre-signed S3 URL
    """

    path: str
    size: int
    mime_type: str
    type: ArtifactType
    created_at: datetime
    checksum: Optional[str] = None
    download_url: Optional[str] = None

    def __post_init__(self):
        """Validate path for security (prevent traversal attacks)."""
        if ".." in self.path:
            raise ValueError("Path cannot contain '..'")
        if self.path.startswith("."):
            raise ValueError("Path cannot start with '.'")

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization."""
        return {
            "path": self.path,
            "size": self.size,
            "mime_type": self.mime_type,
            "type": self.type.value,
            "created_at": self.created_at.isoformat(),
            "checksum": self.checksum,
            "download_url": self.download_url,
        }


@dataclass(frozen=True)
class ResourceLimit:
    """
    Resource limits for execution.

    Attributes:
        timeout_seconds: Maximum execution time in seconds
        max_memory_mb: Maximum memory in megabytes
        max_processes: Maximum number of processes
        max_file_size_mb: Maximum file size in megabytes
    """

    timeout_seconds: int = 300
    max_memory_mb: int = 512
    max_processes: int = 128
    max_file_size_mb: int = 100

    def validate(self) -> None:
        """Validate resource limits."""
        if self.timeout_seconds <= 0:
            raise ValueError("timeout_seconds must be positive")
        if self.timeout_seconds > 3600:
            raise ValueError("timeout_seconds cannot exceed 3600 (1 hour)")
        if self.max_memory_mb <= 0:
            raise ValueError("max_memory_mb must be positive")


@dataclass
class ExecutionMetrics:
    """
    Performance metrics collected during execution.

    Moved from models.py as a domain value object.
    Not frozen because metrics are collected during execution.

    Attributes:
        duration_ms: Wall-clock time in milliseconds
        cpu_time_ms: CPU time in milliseconds
        peak_memory_mb: Peak memory usage in MB
        io_read_bytes: Bytes read from disk
        io_write_bytes: Bytes written to disk
    """

    duration_ms: float
    cpu_time_ms: float
    peak_memory_mb: Optional[float] = None
    io_read_bytes: Optional[int] = None
    io_write_bytes: Optional[int] = None

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization."""
        return {
            "duration_ms": self.duration_ms,
            "cpu_time_ms": self.cpu_time_ms,
            "peak_memory_mb": self.peak_memory_mb,
            "io_read_bytes": self.io_read_bytes,
            "io_write_bytes": self.io_write_bytes,
        }


@dataclass
class ExecutionResult:
    """
    Result of a code execution.

    Enhanced with additional fields from models.py.
    Not frozen because result is built during execution.

    Attributes:
        status: Final execution status
        stdout: Standard output from execution
        stderr: Standard error from execution
        exit_code: Process exit code
        execution_time_ms: Execution duration in milliseconds
        artifacts: List of generated files
        error: Optional error message if failed
        return_value: Handler function return value (JSON serializable)
        metrics: Performance metrics
    """

    status: ExecutionStatus
    stdout: str
    stderr: str
    exit_code: int
    execution_time_ms: float
    artifacts: List[Artifact] = field(default_factory=list)
    error: Optional[str] = None
    return_value: Optional[dict] = None
    metrics: Optional[ExecutionMetrics] = None

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization."""
        return {
            "status": self.status.value,
            "stdout": self.stdout,
            "stderr": self.stderr,
            "exit_code": self.exit_code,
            "execution_time_ms": self.execution_time_ms,
            # Send file paths (strings) instead of full artifact objects
            "artifacts": [str(a.path) for a in self.artifacts],
            "error": self.error,
            "return_value": self.return_value,
            "metrics": self.metrics.to_dict() if self.metrics else None,
        }


@dataclass(frozen=True)
class ExecutionContext:
    """
    Context information for an execution.

    Attributes:
        workspace_path: Path to workspace directory
        session_id: Session identifier
        execution_id: Execution identifier
        control_plane_url: URL of control plane for callbacks
        env_vars: Environment variables to inject
        event: Business data passed to handler function
    """

    workspace_path: Path
    session_id: str
    execution_id: str
    control_plane_url: str
    env_vars: Dict[str, str] = field(default_factory=dict)
    event: Dict[str, Any] = field(default_factory=dict)


@dataclass(frozen=True)
class HeartbeatSignal:
    """
    Liveness signal during execution.

    Moved from models.py to domain layer.

    Attributes:
        timestamp: Heartbeat time (ISO 8601 format)
        progress: Optional progress information
    """

    timestamp: datetime
    progress: Optional[Dict[str, Any]] = None

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization."""
        return {
            "timestamp": self.timestamp.isoformat(),
            "progress": self.progress,
        }


@dataclass(frozen=True)
class ContainerLifecycleEvent:
    """
    Container lifecycle event.

    Moved from models.py to domain layer.

    Attributes:
        event_type: Lifecycle event type ("ready" or "exited")
        container_id: Container ID
        pod_name: Optional pod name (Kubernetes only)
        executor_port: HTTP API port
        ready_at: When API started listening (for ready event)
        exit_code: Container exit code (for exited event)
        exit_reason: Exit reason (for exited event)
        exited_at: When container exited (for exited event)
    """

    event_type: Literal["ready", "exited"]
    container_id: str
    pod_name: Optional[str] = None
    executor_port: int = 8080
    ready_at: Optional[datetime] = None
    exit_code: Optional[int] = None
    exit_reason: Optional[ExitReason] = None
    exited_at: Optional[datetime] = None

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization."""
        return {
            "event_type": self.event_type,
            "container_id": self.container_id,
            "pod_name": self.pod_name,
            "executor_port": self.executor_port,
            "ready_at": self.ready_at.isoformat() if self.ready_at else None,
            "exit_code": self.exit_code,
            "exit_reason": self.exit_reason.value if self.exit_reason else None,
            "exited_at": self.exited_at.isoformat() if self.exited_at else None,
        }


@dataclass(frozen=True)
class ExecutionRequest:
    """
    Request to execute code in the sandbox.

    Moved from models.py to domain layer as value object.

    Attributes:
        code: User code to execute (AWS Lambda handler function, max 1MB)
        language: Programming language (python/javascript/shell)
        timeout: Maximum execution time in seconds (1-3600)
        event: Business data passed to handler function
        execution_id: Unique execution identifier (pattern: exec_[0-9]{8}_[a-z0-9]{8})
        session_id: Session identifier
        env_vars: Environment variables to inject
    """

    code: str
    language: Literal["python", "javascript", "shell"]
    timeout: int
    execution_id: str
    session_id: Optional[str] = None
    event: Dict[str, Any] = field(default_factory=dict)
    env_vars: Dict[str, str] = field(default_factory=dict)

    def __post_init__(self):
        """Validate execution request."""
        if len(self.code) > 1048576:
            raise ValueError("Code size exceeds 1MB limit")
        if self.timeout < 1 or self.timeout > 3600:
            raise ValueError("Timeout must be between 1 and 3600 seconds")

    def to_context(
        self,
        workspace_path: Path,
        control_plane_url: str,
    ) -> ExecutionContext:
        """
        Convert request to execution context.

        Args:
            workspace_path: Path to workspace directory
            control_plane_url: Control Plane URL for callbacks

        Returns:
            ExecutionContext value object
        """
        return ExecutionContext(
            workspace_path=workspace_path,
            session_id=self.session_id or self.execution_id,
            execution_id=self.execution_id,
            control_plane_url=control_plane_url,
            event=self.event,
            env_vars=self.env_vars,
        )
