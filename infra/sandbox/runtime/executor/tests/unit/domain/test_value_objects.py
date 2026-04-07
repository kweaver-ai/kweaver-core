"""
Unit tests for Domain Value Objects.

Tests value objects that follow hexagonal architecture principles:
- Immutability (frozen dataclasses)
- Validation in __post_init__
- Type safety with proper enums
"""

import pytest
from datetime import datetime
from dataclasses import FrozenInstanceError
from pathlib import Path

from executor.domain.value_objects import (
    ExecutionRequest,
    ExecutionResult,
    ExecutionStatus,
    Artifact,
    ArtifactType,
    ResourceLimit,
    ExecutionMetrics,
    HeartbeatSignal,
    ContainerLifecycleEvent,
    ExitReason,
    ExecutionContext,
)


class TestExecutionRequest:
    """Tests for ExecutionRequest value object."""

    def test_create_valid_request(self):
        """Test creating a valid execution request."""
        request = ExecutionRequest(
            execution_id="exec_12345678_abc12345",
            session_id="session_001",
            code="print('hello')",
            language="python",
            timeout=10,
            event={"key": "value"},
            env_vars={"ENV": "test"},
        )

        assert request.execution_id == "exec_12345678_abc12345"
        assert request.code == "print('hello')"
        assert request.language == "python"
        assert request.timeout == 10
        assert request.event == {"key": "value"}
        assert request.env_vars == {"ENV": "test"}

    def test_create_request_with_defaults(self):
        """Test creating a request with default values."""
        request = ExecutionRequest(
            execution_id="exec_001",
            code="test",
            language="python",
            timeout=10,
        )

        assert request.session_id is None
        assert request.event == {}
        assert request.env_vars == {}

    def test_request_is_immutable(self):
        """Test that ExecutionRequest is immutable (frozen)."""
        request = ExecutionRequest(
            execution_id="exec_001",
            code="test",
            language="python",
            timeout=10,
        )

        # Should raise FrozenInstanceError when trying to modify
        with pytest.raises(FrozenInstanceError):
            request.code = "modified"

    def test_request_rejects_large_code(self):
        """Test that ExecutionRequest rejects code larger than 1MB."""
        large_code = "x" * (1048576 + 1)  # 1MB + 1 byte

        with pytest.raises(ValueError, match="Code size exceeds 1MB limit"):
            ExecutionRequest(
                execution_id="exec_001",
                code=large_code,
                language="python",
                timeout=10,
            )

    def test_request_rejects_invalid_timeout(self):
        """Test that ExecutionRequest rejects invalid timeout values."""
        # Timeout too small
        with pytest.raises(ValueError, match="Timeout must be between 1 and 3600"):
            ExecutionRequest(
                execution_id="exec_001",
                code="test",
                language="python",
                timeout=0,
            )

        # Timeout too large
        with pytest.raises(ValueError, match="Timeout must be between 1 and 3600"):
            ExecutionRequest(
                execution_id="exec_001",
                code="test",
                language="python",
                timeout=3601,
            )

    def test_to_context_creates_valid_context(self):
        """Test that to_context() creates a valid ExecutionContext."""
        request = ExecutionRequest(
            execution_id="exec_001",
            session_id="session_001",
            code="test",
            language="python",
            timeout=10,
            event={"input": "data"},
            env_vars={"KEY": "VALUE"},
        )

        context = request.to_context(
            workspace_path=Path("/workspace"),
            control_plane_url="http://localhost:8000",
        )

        assert context.execution_id == "exec_001"
        assert context.session_id == "session_001"
        assert context.event == {"input": "data"}
        assert context.env_vars == {"KEY": "VALUE"}

    def test_to_context_uses_execution_id_as_session_id(self):
        """Test that to_context uses execution_id as session_id when not provided."""
        request = ExecutionRequest(
            execution_id="exec_001",
            code="test",
            language="python",
            timeout=10,
        )

        context = request.to_context(
            workspace_path=Path("/workspace"),
            control_plane_url="http://localhost:8000",
        )

        assert context.session_id == "exec_001"


class TestExecutionContext:
    """Tests for ExecutionContext value object."""

    def test_create_context(self):
        """Test creating an execution context."""
        context = ExecutionContext(
            workspace_path=Path("/workspace"),
            session_id="session_001",
            execution_id="exec_001",
            control_plane_url="http://localhost:8000",
            env_vars={"KEY": "VALUE"},
            event={"input": "data"},
        )

        assert context.session_id == "session_001"
        assert context.execution_id == "exec_001"
        assert context.env_vars == {"KEY": "VALUE"}
        assert context.event == {"input": "data"}

    def test_context_with_defaults(self):
        """Test context with default values."""
        context = ExecutionContext(
            workspace_path=Path("/workspace"),
            session_id="session_001",
            execution_id="exec_001",
            control_plane_url="http://localhost:8000",
        )

        assert context.session_id == "session_001"
        assert context.execution_id == "exec_001"
        assert context.env_vars == {}
        assert context.event == {}

    def test_context_is_immutable(self):
        """Test that ExecutionContext is immutable (frozen)."""
        context = ExecutionContext(
            workspace_path=Path("/workspace"),
            session_id="session_001",
            execution_id="exec_001",
            control_plane_url="http://localhost:8000",
        )

        with pytest.raises(FrozenInstanceError):
            context.session_id = "modified"


class TestExecutionResult:
    """Tests for ExecutionResult value object."""

    def test_create_success_result(self):
        """Test creating a successful execution result."""
        result = ExecutionResult(
            status=ExecutionStatus.SUCCESS,
            stdout="output",
            stderr="",
            exit_code=0,
            execution_time_ms=100,
        )

        assert result.status == ExecutionStatus.SUCCESS
        assert result.exit_code == 0

    def test_result_with_metrics(self):
        """Test result with execution metrics."""
        metrics = ExecutionMetrics(
            duration_ms=100,
            cpu_time_ms=80,
            peak_memory_mb=50,
        )

        result = ExecutionResult(
            status=ExecutionStatus.SUCCESS,
            stdout="output",
            stderr="",
            exit_code=0,
            execution_time_ms=100,
            metrics=metrics,
        )

        assert result.metrics.duration_ms == 100
        assert result.metrics.peak_memory_mb == 50

    def test_result_with_artifacts(self):
        """Test result with artifacts."""
        artifact = Artifact(
            path="output.txt",
            size=100,
            mime_type="text/plain",
            type=ArtifactType.OUTPUT,
            created_at=datetime.now(),
        )

        result = ExecutionResult(
            status=ExecutionStatus.SUCCESS,
            stdout="output",
            stderr="",
            exit_code=0,
            execution_time_ms=100,
            artifacts=[artifact],
        )

        assert len(result.artifacts) == 1
        assert result.artifacts[0].path == "output.txt"

    def test_result_is_mutable(self):
        """Test that ExecutionResult is mutable (not frozen)."""
        result = ExecutionResult(
            status=ExecutionStatus.SUCCESS,
            stdout="output",
            stderr="",
            exit_code=0,
            execution_time_ms=100,
        )

        # ExecutionResult is NOT frozen, so we can modify it
        result.status = ExecutionStatus.FAILED
        assert result.status == ExecutionStatus.FAILED

    def test_result_to_dict(self):
        """Test that ExecutionResult can be converted to dict."""
        result = ExecutionResult(
            status=ExecutionStatus.SUCCESS,
            stdout="output",
            stderr="",
            exit_code=0,
            execution_time_ms=100,
        )

        result_dict = result.to_dict()
        assert result_dict["status"] == "success"
        assert result_dict["stdout"] == "output"
        assert result_dict["exit_code"] == 0


class TestArtifact:
    """Tests for Artifact value object."""

    def test_create_valid_artifact(self):
        """Test creating a valid artifact."""
        artifact = Artifact(
            path="output/data.json",
            size=1024,
            mime_type="application/json",
            type=ArtifactType.ARTIFACT,
            created_at=datetime.now(),
        )

        assert artifact.path == "output/data.json"
        assert artifact.size == 1024
        assert artifact.type == ArtifactType.ARTIFACT

    def test_artifact_rejects_path_traversal(self):
        """Test that artifact rejects path traversal attempts."""
        with pytest.raises(ValueError, match=".*cannot contain.*"):
            Artifact(
                path="../etc/passwd",
                size=100,
                mime_type="text/plain",
                type=ArtifactType.ARTIFACT,
                created_at=datetime.now(),
            )

    def test_artifact_rejects_hidden_files(self):
        """Test that artifact rejects hidden files."""
        with pytest.raises(ValueError, match=".*cannot start with.*"):
            Artifact(
                path=".secret/key.txt",
                size=100,
                mime_type="text/plain",
                type=ArtifactType.ARTIFACT,
                created_at=datetime.now(),
            )

    def test_artifact_with_optional_fields(self):
        """Test artifact with optional checksum and download_url."""
        artifact = Artifact(
            path="output/result.pdf",
            size=2048,
            mime_type="application/pdf",
            type=ArtifactType.OUTPUT,
            created_at=datetime.now(),
            checksum="abc123",
            download_url="http://storage/result.pdf",
        )

        assert artifact.checksum == "abc123"
        assert artifact.download_url == "http://storage/result.pdf"

    def test_artifact_to_dict(self):
        """Test that Artifact can be converted to dict."""
        artifact = Artifact(
            path="output.txt",
            size=100,
            mime_type="text/plain",
            type=ArtifactType.OUTPUT,
            created_at=datetime.now(),
        )

        result = artifact.to_dict()
        assert result["path"] == "output.txt"
        assert result["type"] == "output"


class TestExecutionMetrics:
    """Tests for ExecutionMetrics value object."""

    def test_create_metrics(self):
        """Test creating execution metrics."""
        metrics = ExecutionMetrics(
            duration_ms=1000,
            cpu_time_ms=800,
            peak_memory_mb=128,
            io_read_bytes=4096,
            io_write_bytes=2048,
        )

        assert metrics.duration_ms == 1000
        assert metrics.cpu_time_ms == 800
        assert metrics.peak_memory_mb == 128
        assert metrics.io_read_bytes == 4096
        assert metrics.io_write_bytes == 2048

    def test_metrics_with_optional_fields(self):
        """Test metrics with optional fields omitted."""
        metrics = ExecutionMetrics(
            duration_ms=100,
            cpu_time_ms=80,
        )

        assert metrics.duration_ms == 100
        assert metrics.peak_memory_mb is None

    def test_metrics_to_dict(self):
        """Test that ExecutionMetrics can be converted to dict."""
        metrics = ExecutionMetrics(
            duration_ms=100,
            cpu_time_ms=80,
        )

        result = metrics.to_dict()
        assert result["duration_ms"] == 100
        assert result["cpu_time_ms"] == 80


class TestHeartbeatSignal:
    """Tests for HeartbeatSignal value object."""

    def test_create_heartbeat(self):
        """Test creating a heartbeat signal."""
        signal = HeartbeatSignal(
            timestamp=datetime.now(),
            progress={"status": "running", "percent": 50},
        )

        assert signal.progress["status"] == "running"
        assert signal.progress["percent"] == 50

    def test_heartbeat_is_immutable(self):
        """Test that heartbeat signal is immutable."""
        signal = HeartbeatSignal(
            timestamp=datetime.now(),
            progress={},
        )

        with pytest.raises(FrozenInstanceError):
            signal.progress = {"modified": True}

    def test_heartbeat_to_dict(self):
        """Test that HeartbeatSignal can be converted to dict."""
        signal = HeartbeatSignal(
            timestamp=datetime.now(),
            progress={"status": "running"},
        )

        result = signal.to_dict()
        assert "timestamp" in result
        assert result["progress"]["status"] == "running"


class TestContainerLifecycleEvent:
    """Tests for ContainerLifecycleEvent value object."""

    def test_create_ready_event(self):
        """Test creating a container ready event."""
        event = ContainerLifecycleEvent(
            event_type="ready",
            container_id="container-123",
            pod_name="pod-456",
            executor_port=8080,
            ready_at=datetime.now(),
        )

        assert event.event_type == "ready"
        assert event.container_id == "container-123"

    def test_create_exited_event(self):
        """Test creating a container exited event."""
        event = ContainerLifecycleEvent(
            event_type="exited",
            container_id="container-123",
            pod_name="pod-456",
            executor_port=8080,
            exited_at=datetime.now(),
            exit_reason=ExitReason.SIGTERM,
            exit_code=143,
        )

        assert event.event_type == "exited"
        assert event.exit_reason == ExitReason.SIGTERM
        assert event.exit_code == 143

    def test_event_to_dict(self):
        """Test that ContainerLifecycleEvent can be converted to dict."""
        event = ContainerLifecycleEvent(
            event_type="ready",
            container_id="container-123",
            pod_name="pod-456",
            executor_port=8080,
            ready_at=datetime.now(),
        )

        result = event.to_dict()
        assert result["event_type"] == "ready"
        assert result["container_id"] == "container-123"


class TestResourceLimit:
    """Tests for ResourceLimit value object."""

    def test_create_resource_limits(self):
        """Test creating resource limits."""
        limits = ResourceLimit(
            timeout_seconds=300,
            max_memory_mb=512,
            max_processes=128,
            max_file_size_mb=100,
        )

        assert limits.timeout_seconds == 300
        assert limits.max_memory_mb == 512
        assert limits.max_processes == 128
        assert limits.max_file_size_mb == 100

    def test_default_resource_limits(self):
        """Test default resource limits."""
        limits = ResourceLimit()

        assert limits.timeout_seconds == 300
        assert limits.max_memory_mb == 512

    def test_validate_positive_timeout(self):
        """Test that timeout must be positive."""
        with pytest.raises(ValueError, match="timeout_seconds must be positive"):
            limits = ResourceLimit(timeout_seconds=0)
            limits.validate()

    def test_validate_max_timeout(self):
        """Test that timeout cannot exceed 3600."""
        with pytest.raises(ValueError, match="timeout_seconds cannot exceed 3600"):
            limits = ResourceLimit(timeout_seconds=3601)
            limits.validate()

    def test_validate_positive_memory(self):
        """Test that memory must be positive."""
        with pytest.raises(ValueError, match="max_memory_mb must be positive"):
            limits = ResourceLimit(max_memory_mb=0)
            limits.validate()


class TestExitReason:
    """Tests for ExitReason enum."""

    def test_exit_reason_values(self):
        """Test ExitReason enum values."""
        assert ExitReason.NORMAL.value == "normal"
        assert ExitReason.SIGTERM.value == "sigterm"
        assert ExitReason.SIGKILL.value == "sigkill"
        assert ExitReason.OOM_KILLED.value == "oom_killed"
        assert ExitReason.ERROR.value == "error"


class TestExecutionStatus:
    """Tests for ExecutionStatus enum."""

    def test_status_values(self):
        """Test ExecutionStatus enum values."""
        assert ExecutionStatus.PENDING.value == "pending"
        assert ExecutionStatus.RUNNING.value == "running"
        assert ExecutionStatus.COMPLETED.value == "completed"
        assert ExecutionStatus.FAILED.value == "failed"
        assert ExecutionStatus.TIMEOUT.value == "timeout"
        assert ExecutionStatus.CRASHED.value == "crashed"
        assert ExecutionStatus.SUCCESS.value == "success"
        assert ExecutionStatus.ERROR.value == "error"


class TestArtifactType:
    """Tests for ArtifactType enum."""

    def test_artifact_type_values(self):
        """Test ArtifactType enum values."""
        assert ArtifactType.OUTPUT.value == "output"
        assert ArtifactType.LOG.value == "log"
        assert ArtifactType.ARTIFACT.value == "artifact"
        assert ArtifactType.TEMP.value == "temp"
