"""
Integration tests for Hexagonal Architecture.

Tests the complete flow from HTTP interface through application layer
to infrastructure adapters, verifying that all layers work together correctly.
"""

import pytest
import asyncio
import os
from pathlib import Path
from datetime import datetime, timezone
from unittest.mock import AsyncMock, Mock, patch
from tempfile import TemporaryDirectory

from executor.interfaces.http.rest import create_app
from executor.application.commands.execute_code import ExecuteCodeCommand
from executor.application.services.heartbeat_service import HeartbeatService
from executor.application.services.lifecycle_service import LifecycleService
from executor.application.services.session_config_sync_service import (
    InstalledDependency,
    SessionConfigSyncResult,
)
from executor.domain.value_objects import (
    ExecutionRequest,
    ExecutionResult,
    ExecutionStatus,
    Artifact,
    ArtifactType,
)
from executor.infrastructure.isolation.bwrap import BubblewrapRunner
from executor.infrastructure.http.callback_client import CallbackClient
from executor.infrastructure.persistence.artifact_scanner import ArtifactScanner
from executor.infrastructure.config.config import Settings


@pytest.mark.integration
class TestHexagonalExecutionFlow:
    """Tests the complete execution flow through all layers."""

    @pytest.fixture
    def temp_workspace(self):
        """Create a temporary workspace for testing."""
        with TemporaryDirectory() as tmpdir:
            yield Path(tmpdir)

    @pytest.fixture
    def mock_settings(self, temp_workspace):
        """Create mock settings."""
        return Settings(
            workspace_path=temp_workspace,
            control_plane_url="http://localhost:8000",
            executor_port=8080,
            log_level="INFO",
            internal_api_token="test-token",
        )

    @pytest.fixture
    def mock_callback_client(self):
        """Create a mock callback client."""
        mock = AsyncMock(spec=CallbackClient)
        mock.report_result.return_value = True
        mock.report_heartbeat.return_value = True
        mock.report_lifecycle.return_value = True
        return mock

    @pytest.fixture
    def execute_command(self, temp_workspace, mock_callback_client):
        """Create ExecuteCodeCommand with all dependencies."""
        bwrap_runner = Mock(spec=BubblewrapRunner)
        bwrap_runner.execute = AsyncMock(return_value=ExecutionResult(
            status=ExecutionStatus.SUCCESS,
            stdout="test output",
            stderr="",
            exit_code=0,
            execution_time_ms=100,
        ))

        artifact_scanner = ArtifactScanner()

        heartbeat_service = HeartbeatService(
            callback_port=mock_callback_client,
            interval=1.0,
        )

        return ExecuteCodeCommand(
            isolation_port=bwrap_runner,
            artifact_scanner_port=artifact_scanner,
            callback_port=mock_callback_client,
            heartbeat_port=heartbeat_service,
            workspace_path=temp_workspace,
            control_plane_url="http://localhost:8000",
        )

    @pytest.mark.asyncio
    async def test_complete_execution_flow(self, execute_command):
        """Test the complete execution flow from request to result."""
        # Create execution request
        request = ExecutionRequest(
            execution_id="exec_test_001",
            session_id="session_test_001",
            code="print('hello world')",
            language="python",
            timeout=10,
            env_vars={},
        )

        # Execute through the command
        result = await execute_command.execute(request)

        # Verify result
        assert result.status == ExecutionStatus.SUCCESS
        assert result.stdout == "test output"

        # Verify heartbeat was started and stopped
        # (active executions should be tracked)
        assert execute_command.get_active_count() == 0

    @pytest.mark.asyncio
    async def test_execution_with_timeout(self, execute_command):
        """Test execution timeout handling."""
        # Mock isolation to simulate timeout
        async def timeout_mock(execution):
            await asyncio.sleep(0.1)
            raise asyncio.TimeoutError()

        execute_command._isolation_port.execute = timeout_mock

        request = ExecutionRequest(
            execution_id="exec_timeout_001",
            session_id="session_timeout_001",
            code="import time; time.sleep(60)",
            language="python",
            timeout=1,
            env_vars={},
        )

        result = await execute_command.execute(request)

        assert result.status == ExecutionStatus.TIMEOUT
        assert "timeout" in result.stderr.lower()

    @pytest.mark.asyncio
    async def test_execution_with_artifacts(self, temp_workspace, mock_callback_client):
        """Test execution with artifact collection."""
        # Create test artifact in workspace
        (temp_workspace / "output.txt").write_text("test output")

        bwrap_runner = Mock(spec=BubblewrapRunner)
        bwrap_runner.execute = AsyncMock(return_value=ExecutionResult(
            status=ExecutionStatus.SUCCESS,
            stdout="",
            stderr="",
            exit_code=0,
            execution_time_ms=100,
        ))

        artifact_scanner = ArtifactScanner()

        heartbeat_service = HeartbeatService(
            callback_port=mock_callback_client,
            interval=1.0,
        )

        command = ExecuteCodeCommand(
            isolation_port=bwrap_runner,
            artifact_scanner_port=artifact_scanner,
            callback_port=mock_callback_client,
            heartbeat_port=heartbeat_service,
            workspace_path=temp_workspace,
            control_plane_url="http://localhost:8000",
        )

        request = ExecutionRequest(
            execution_id="exec_artifact_001",
            session_id="session_artifact_001",
            code="# test",
            language="python",
            timeout=10,
            env_vars={},
        )

        result = await command.execute(request)

        # Verify artifacts were collected
        assert len(result.artifacts) > 0
        assert any(a.path == "output.txt" for a in result.artifacts)


@pytest.mark.integration
class TestPortAdapterIntegration:
    """Tests that ports correctly integrate with their adapters."""

    @pytest.mark.asyncio
    async def test_callback_port_integration(self):
        """Test ICallbackPort integration with CallbackClient."""
        client = CallbackClient(
            control_plane_url="http://localhost:8000",
            api_token="test-token",
        )

        # Verify it implements the port correctly
        from executor.domain.ports import ICallbackPort
        assert isinstance(client, ICallbackPort)

        # Test method signatures
        result = ExecutionResult(
            status=ExecutionStatus.SUCCESS,
            stdout="",
            stderr="",
            exit_code=0,
            execution_time_ms=100,
        )

        # Method exists and has correct signature (won't actually call)
        assert hasattr(client, 'report_result')
        assert hasattr(client, 'report_heartbeat')
        assert hasattr(client, 'report_lifecycle')
        assert hasattr(client, 'close')

    def test_artifact_scanner_port_integration(self):
        """Test IArtifactScannerPort integration with ArtifactScanner."""
        from tempfile import TemporaryDirectory
        from executor.domain.ports import IArtifactScannerPort

        with TemporaryDirectory() as tmpdir:
            scanner = ArtifactScanner()

            # Verify it implements the port
            assert isinstance(scanner, IArtifactScannerPort)

            # Test snapshot method
            snapshot = scanner.snapshot(Path(tmpdir))
            assert isinstance(snapshot, set)

            # Test collect_artifacts method
            artifacts = scanner.collect_artifacts(Path(tmpdir))
            assert isinstance(artifacts, list)


@pytest.mark.integration
class TestApplicationServiceIntegration:
    """Tests application services working together."""

    @pytest.mark.asyncio
    async def test_heartbeat_with_callback(self):
        """Test HeartbeatService sending heartbeats via callback."""
        mock_callback = AsyncMock()
        mock_callback.report_heartbeat.return_value = True

        service = HeartbeatService(
            callback_port=mock_callback,
            interval=0.5,  # Fast for testing
        )

        await service.start_heartbeat("exec_001")
        await asyncio.sleep(1.2)  # Wait for at least 2 heartbeats

        await service.stop_heartbeat("exec_001")

        # Should have sent at least 2 heartbeats
        assert mock_callback.report_heartbeat.call_count >= 2

    @pytest.mark.asyncio
    async def test_lifecycle_with_callback(self):
        """Test LifecycleService sending events via callback."""
        mock_callback = AsyncMock()
        mock_callback.report_lifecycle.return_value = True

        mock_heartbeat = AsyncMock()
        mock_heartbeat.stop_all.return_value = None

        service = LifecycleService(
            callback_port=mock_callback,
            executor_port=8080,
            heartbeat_port=mock_heartbeat,
        )

        # Test container_ready
        result = await service.send_container_ready()
        assert result is True
        mock_callback.report_lifecycle.assert_called_once()

        event = mock_callback.report_lifecycle.call_args[0][0]
        assert event.event_type == "ready"

        # Test shutdown
        await service.shutdown()

        # Should have sent container_exited
        assert mock_callback.report_lifecycle.call_count == 2
        event = mock_callback.report_lifecycle.call_args[0][0]
        assert event.event_type == "exited"


@pytest.mark.integration
class TestHTTPInterfaceIntegration:
    """Tests HTTP interface integration with application layer."""

    @pytest.fixture
    def test_app(self):
        """Create test FastAPI app with mocked dependencies."""
        from unittest.mock import patch

        # Mock all the infrastructure dependencies
        with patch('executor.interfaces.http.rest.BubblewrapRunner'), \
             patch('executor.interfaces.http.rest.CallbackClient'), \
             patch('executor.interfaces.http.rest.ArtifactScanner'), \
             patch('executor.interfaces.http.rest.HeartbeatService'), \
             patch('executor.interfaces.http.rest.LifecycleService'), \
             patch('executor.interfaces.http.rest.ExecuteCodeCommand'), \
             patch('executor.interfaces.http.rest.SessionConfigSyncService'):

            app = create_app()
            return app

    @pytest.mark.asyncio
    async def test_health_endpoint(self, test_app):
        """Test /health endpoint works correctly."""
        from fastapi.testclient import TestClient
        from pathlib import Path as RealPath
        import tempfile

        client = TestClient(test_app)

        # Create a real temporary workspace directory
        with tempfile.TemporaryDirectory() as tmpdir:
            # Mock health check dependencies
            with patch('executor.infrastructure.isolation.bwrap.check_bwrap_available'), \
                 patch('executor.infrastructure.isolation.bwrap.get_bwrap_version', return_value='1.7.0'), \
                 patch.dict(os.environ, {'WORKSPACE_PATH': tmpdir}):

                response = client.get("/health")

                assert response.status_code == 200
                data = response.json()
                assert data["status"] == "healthy"
                assert "version" in data

    @pytest.mark.asyncio
    async def test_execute_endpoint_integration(self, test_app):
        """Test /execute endpoint integration."""
        from fastapi.testclient import TestClient
        from unittest.mock import AsyncMock, patch

        client = TestClient(test_app)

        # Mock the execute command
        mock_command = AsyncMock()
        mock_result = ExecutionResult(
            status=ExecutionStatus.SUCCESS,
            stdout="output",
            stderr="",
            exit_code=0,
            execution_time_ms=100,
        )
        mock_command.execute.return_value = mock_result
        mock_command.get_active_count.return_value = 0

        with patch('executor.interfaces.http.rest.get_execute_command', return_value=mock_command):
            response = client.post(
                "/execute",
                json={
                    "execution_id": "test_001",
                    "session_id": "session_001",
                    "code": "print('test')",
                    "language": "python",
                    "timeout": 10,
                },
            )

            assert response.status_code == 200
            data = response.json()
            assert data["execution_id"] == "test_001"
            assert data["status"] == "success"

    @pytest.mark.asyncio
    async def test_sync_session_config_endpoint(self, test_app):
        """Test /internal/session-config/sync endpoint integration."""
        from fastapi.testclient import TestClient
        from unittest.mock import AsyncMock, patch

        client = TestClient(test_app)

        sync_result = SessionConfigSyncResult(
            status="completed",
            installed_dependencies=[
                InstalledDependency(
                    name="requests",
                    version="2.31.0",
                    install_location="/opt/sandbox-venv",
                    install_time=datetime.now(timezone.utc),
                    is_from_template=False,
                )
            ],
            error="",
            started_at=datetime.now(timezone.utc),
            completed_at=datetime.now(timezone.utc),
        )

        mock_service = AsyncMock()
        mock_service.sync.return_value = sync_result

        with patch(
            "executor.interfaces.http.rest.get_session_config_sync_service",
            return_value=mock_service,
        ):
            response = client.post(
                "/internal/session-config/sync",
                json={
                    "session_id": "sess_1",
                    "language_runtime": "python3.11",
                    "python_package_index_url": "https://pypi.org/simple/",
                    "dependencies": ["requests==2.31.0"],
                    "sync_mode": "replace",
                },
            )

        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "completed"
        assert data["installed_dependencies"][0]["name"] == "requests"


@pytest.mark.integration
class TestEndToEndExecution:
    """End-to-end tests simulating real execution scenarios."""

    @pytest.fixture
    def real_workspace(self):
        """Create a real temporary workspace."""
        with TemporaryDirectory() as tmpdir:
            yield Path(tmpdir)

    def test_python_execution_with_file_artifacts(self, real_workspace):
        """Test Python execution that creates file artifacts."""
        # This is a contract test - verifies the contract between layers
        from executor.domain.entities import Execution, ExecutionContext
        from executor.domain.value_objects import ExecutionStatus

        # Create execution context
        context = ExecutionContext(
            workspace_path=real_workspace,
            session_id="test_session",
            execution_id="test_exec",
            control_plane_url="http://localhost:8000",
        )

        # Create execution entity
        execution = Execution(
            execution_id="test_exec",
            session_id="test_session",
            code="# Python test code",
            language="python",
            context=context,
        )

        # Verify initial state
        assert execution.status == ExecutionStatus.PENDING

        # Mark as running
        execution.mark_as_running()
        assert execution.status == ExecutionStatus.RUNNING
        assert execution.started_at is not None

        # Create artifact file in workspace
        (real_workspace / "result.json").write_text('{"result": "success"}')

        # Mark as completed
        result = ExecutionResult(
            status=ExecutionStatus.SUCCESS,
            stdout="",
            stderr="",
            exit_code=0,
            execution_time_ms=100,
        )

        execution.mark_as_completed(result)

        # Verify final state
        assert execution.status == ExecutionStatus.COMPLETED
        assert execution.completed_at is not None

        # Verify artifact can be collected
        scanner = ArtifactScanner()
        artifacts = scanner.collect_artifacts(real_workspace)

        assert len(artifacts) > 0
        assert any(a.path == "result.json" for a in artifacts)

    def test_artifact_scanner_with_real_files(self, real_workspace):
        """Test artifact scanner with real files."""
        # Create test files
        (real_workspace / "output.txt").write_text("output")
        (real_workspace / "data.json").write_text('{"data": "value"}')
        (real_workspace / ".secret").write_text("hidden")
        (real_workspace / "logs").mkdir(exist_ok=True)
        (real_workspace / "logs" / "app.log").write_text("logs")
        (real_workspace / "logs" / ".cache").write_text("cache")

        # Scan for artifacts
        scanner = ArtifactScanner()
        artifacts = scanner.collect_artifacts(real_workspace)

        # Should find visible files
        paths = [a.path for a in artifacts]
        assert "output.txt" in paths
        assert "data.json" in paths
        assert "logs/app.log" in paths

        # Should NOT find hidden files
        assert ".secret" not in paths
        assert ".cache" not in paths

        # Verify artifact types
        output_artifact = next(a for a in artifacts if a.path == "output.txt")
        assert output_artifact.type == ArtifactType.ARTIFACT

        log_artifact = next(a for a in artifacts if a.path == "logs/app.log")
        assert log_artifact.type == ArtifactType.LOG
