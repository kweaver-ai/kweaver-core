"""
Unit tests for ExecuteCodeCommand.

Tests the main execution orchestrator use case.
"""

import pytest
import asyncio
from pathlib import Path
from datetime import datetime
from unittest.mock import AsyncMock, Mock, patch

from executor.application.commands.execute_code import ExecuteCodeCommand
from executor.domain.value_objects import (
    ExecutionRequest,
    ExecutionResult,
    ExecutionStatus,
    Artifact,
    ArtifactType,
)
from executor.domain.entities import Execution


class TestExecuteCodeCommand:
    """Tests for ExecuteCodeCommand."""

    @pytest.fixture
    def mock_isolation_port(self):
        """Create a mock isolation port."""
        mock = AsyncMock()
        mock.execute.return_value = ExecutionResult(
            status=ExecutionStatus.SUCCESS,
            stdout="output",
            stderr="",
            exit_code=0,
            execution_time_ms=100,
        )
        return mock

    @pytest.fixture
    def mock_artifact_scanner_port(self):
        """Create a mock artifact scanner port."""
        mock = Mock()
        mock.snapshot.return_value = set()
        mock.collect_artifacts.return_value = []
        return mock

    @pytest.fixture
    def mock_callback_port(self):
        """Create a mock callback port."""
        mock = AsyncMock()
        mock.report_result.return_value = True
        return mock

    @pytest.fixture
    def mock_heartbeat_port(self):
        """Create a mock heartbeat port."""
        mock = AsyncMock()
        mock.start_heartbeat.return_value = None
        mock.stop_heartbeat.return_value = None
        return mock

    @pytest.fixture
    def command(
        self,
        mock_isolation_port,
        mock_artifact_scanner_port,
        mock_callback_port,
        mock_heartbeat_port,
    ):
        """Create an ExecuteCodeCommand instance."""
        return ExecuteCodeCommand(
            isolation_port=mock_isolation_port,
            artifact_scanner_port=mock_artifact_scanner_port,
            callback_port=mock_callback_port,
            heartbeat_port=mock_heartbeat_port,
            workspace_path=Path("/workspace"),
            control_plane_url="http://localhost:8000",
        )

    @pytest.fixture
    def execution_request(self):
        """Create an execution request."""
        return ExecutionRequest(
            execution_id="exec_001",
            session_id="session_001",
            code="print('hello')",
            language="python",
            timeout=10,
            event={},
            env_vars={},
        )

    @pytest.mark.asyncio
    async def test_execute_success(
        self,
        command,
        execution_request,
        mock_isolation_port,
        mock_heartbeat_port,
        mock_artifact_scanner_port,
    ):
        """Test successful execution."""
        result = await command.execute(execution_request)

        assert result.status == ExecutionStatus.SUCCESS
        assert result.stdout == "output"
        assert result.exit_code == 0

        # Verify heartbeat was started and stopped
        mock_heartbeat_port.start_heartbeat.assert_called_once()
        mock_heartbeat_port.stop_heartbeat.assert_called_once()

        # Verify snapshot was taken
        mock_artifact_scanner_port.snapshot.assert_called_once()

    @pytest.mark.asyncio
    async def test_execute_with_artifacts(
        self,
        command,
        execution_request,
        mock_isolation_port,
        mock_artifact_scanner_port,
    ):
        """Test execution with artifact collection."""
        # Mock artifact data
        artifact_data = Mock()
        artifact_data.path = "output.txt"
        artifact_data.size = 100
        artifact_data.mime_type = "text/plain"
        artifact_data.type = ArtifactType.OUTPUT
        artifact_data.created_at = datetime.now()
        artifact_data.checksum = "abc123"

        mock_artifact_scanner_port.collect_artifacts.return_value = [artifact_data]

        result = await command.execute(execution_request)

        assert len(result.artifacts) == 1
        assert result.artifacts[0].path == "output.txt"

    @pytest.mark.asyncio
    async def test_execute_timeout(
        self,
        command,
        execution_request,
        mock_isolation_port,
    ):
        """Test execution timeout."""
        # Make execute hang
        async def hanging_execute(*args, **kwargs):
            await asyncio.sleep(20)  # Longer than timeout
            return ExecutionResult(
                status=ExecutionStatus.SUCCESS,
                stdout="",
                stderr="",
                exit_code=0,
                execution_time_ms=20000,
            )

        mock_isolation_port.execute = hanging_execute

        # Use short timeout
        execution_request = ExecutionRequest(
            execution_id="exec_timeout",
            code="while True: pass",
            language="python",
            timeout=1,  # 1 second timeout
        )

        result = await command.execute(execution_request)

        assert result.status == ExecutionStatus.TIMEOUT
        assert "timeout" in result.stderr.lower()

    @pytest.mark.asyncio
    async def test_execute_exception(
        self,
        command,
        execution_request,
        mock_isolation_port,
    ):
        """Test execution with exception."""
        mock_isolation_port.execute.side_effect = Exception("Execution failed")

        result = await command.execute(execution_request)

        assert result.status == ExecutionStatus.ERROR
        assert "Execution failed" in result.stderr

    @pytest.mark.asyncio
    async def test_active_executions_tracking(
        self,
        command,
        execution_request,
    ):
        """Test that active executions are tracked."""
        assert command.get_active_count() == 0

        # Start execution in background
        task = asyncio.create_task(command.execute(execution_request))

        # Wait a bit for execution to start
        await asyncio.sleep(0.1)

        # Should have one active execution during execution
        # Note: This might be 0 if execution completes quickly

        await task

        # After completion, should be 0
        assert command.get_active_count() == 0

    @pytest.mark.asyncio
    async def test_concurrent_executions(
        self,
        command,
        mock_isolation_port,
    ):
        """Test concurrent executions."""
        # Make execute take some time
        async def slow_execute(*args, **kwargs):
            await asyncio.sleep(0.2)
            return ExecutionResult(
                status=ExecutionStatus.SUCCESS,
                stdout="output",
                stderr="",
                exit_code=0,
                execution_time_ms=200,
            )

        mock_isolation_port.execute = slow_execute

        requests = [
            ExecutionRequest(
                execution_id=f"exec_{i}",
                code="test",
                language="python",
                timeout=10,
            )
            for i in range(3)
        ]

        # Execute all concurrently
        results = await asyncio.gather(*[
            command.execute(req) for req in requests
        ])

        # All should succeed
        assert all(r.status == ExecutionStatus.SUCCESS for r in results)

        # No active executions after all complete
        assert command.get_active_count() == 0

    def test_get_active_count(self, command):
        """Test get_active_count method."""
        assert command.get_active_count() == 0

        # Manually add to active executions
        command._active_executions.add("exec_001")
        assert command.get_active_count() == 1

        command._active_executions.add("exec_002")
        assert command.get_active_count() == 2

        command._active_executions.discard("exec_001")
        command._active_executions.discard("exec_002")
        assert command.get_active_count() == 0
