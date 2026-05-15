"""
Unit tests for Application Services.

Tests application services that orchestrate domain objects:
- HeartbeatService
- LifecycleService
"""

import pytest
import asyncio
import signal
import os
from datetime import datetime
from pathlib import Path
from unittest.mock import AsyncMock, Mock, patch

from executor.application.services.heartbeat_service import (
    HeartbeatService,
    get_heartbeat_service,
    register_heartbeat_service,
)
from executor.application.services.lifecycle_service import (
    LifecycleService,
    register_signal_handlers,
    map_exit_code_to_reason,
    get_lifecycle_service,
    register_lifecycle_service,
)
from executor.domain.value_objects import (
    HeartbeatSignal,
    ContainerLifecycleEvent,
    ExitReason,
    ExecutionStatus,
    ExecutionResult,
)


class TestHeartbeatService:
    """Tests for HeartbeatService."""

    @pytest.fixture
    def mock_callback_port(self):
        """Create a mock callback port."""
        mock = AsyncMock()
        mock.report_heartbeat.return_value = True
        return mock

    @pytest.fixture
    def heartbeat_service(self, mock_callback_port):
        """Create a HeartbeatService instance."""
        return HeartbeatService(
            callback_port=mock_callback_port,
            interval=1.0,  # 1 second for testing
        )

    @pytest.mark.asyncio
    async def test_start_heartbeat_creates_task(self, heartbeat_service):
        """Test that starting heartbeat creates a background task."""
        execution_id = "exec_001"

        await heartbeat_service.start_heartbeat(execution_id)

        assert execution_id in heartbeat_service._tasks
        assert execution_id in heartbeat_service._stop_events

        # Cleanup
        await heartbeat_service.stop_heartbeat(execution_id)

    @pytest.mark.asyncio
    async def test_send_heartbeat(self, heartbeat_service, mock_callback_port):
        """Test sending a heartbeat signal."""
        execution_id = "exec_001"
        signal_data = HeartbeatSignal(
            timestamp=datetime.now(),
            progress={"status": "running"},
        )

        result = await heartbeat_service.send_heartbeat(execution_id, signal_data)

        assert result is True
        mock_callback_port.report_heartbeat.assert_called_once_with(execution_id, signal_data)

    @pytest.mark.asyncio
    async def test_stop_heartbeat_stops_task(self, heartbeat_service):
        """Test that stopping heartbeat cancels the task."""
        execution_id = "exec_001"

        await heartbeat_service.start_heartbeat(execution_id)

        # Wait a bit to ensure heartbeat is running
        await asyncio.sleep(0.1)

        # Stop the heartbeat
        await heartbeat_service.stop_heartbeat(execution_id)

        # Task should be removed
        assert execution_id not in heartbeat_service._tasks
        assert execution_id not in heartbeat_service._stop_events

    @pytest.mark.asyncio
    @pytest.mark.slow
    async def test_heartbeat_sends_signals_periodically(self, heartbeat_service, mock_callback_port):
        """Test that heartbeat sends signals periodically."""
        execution_id = "exec_001"

        await heartbeat_service.start_heartbeat(execution_id)

        # Wait for at least 2 heartbeats
        await asyncio.sleep(2.5)

        # Should have been called at least twice
        assert mock_callback_port.report_heartbeat.call_count >= 2

        # Cleanup
        await heartbeat_service.stop_heartbeat(execution_id)

    @pytest.mark.asyncio
    async def test_stop_all_stops_all_heartbeats(self, heartbeat_service):
        """Test that stop_all stops all active heartbeats."""
        exec_ids = ["exec_001", "exec_002", "exec_003"]

        # Start multiple heartbeats
        for exec_id in exec_ids:
            await heartbeat_service.start_heartbeat(exec_id)

        # Wait a bit
        await asyncio.sleep(0.1)

        # Stop all
        await heartbeat_service.stop_all()

        # All should be stopped
        assert len(heartbeat_service._tasks) == 0
        assert len(heartbeat_service._stop_events) == 0

    @pytest.mark.asyncio
    async def test_start_heartbeat_twice_warning(self, heartbeat_service):
        """Test that starting heartbeat twice for same execution logs warning."""
        execution_id = "exec_001"

        await heartbeat_service.start_heartbeat(execution_id)
        await heartbeat_service.start_heartbeat(execution_id)  # Should log warning but not fail

        # Should still only have one task
        assert len(heartbeat_service._tasks) == 1

        # Cleanup
        await heartbeat_service.stop_heartbeat(execution_id)

    @pytest.mark.asyncio
    async def test_stop_heartbeat_not_running(self, heartbeat_service):
        """Test stopping heartbeat that is not running."""
        # Should not raise error
        await heartbeat_service.stop_heartbeat("nonexistent")

    @pytest.mark.asyncio
    async def test_send_heartbeat_on_exception(self, heartbeat_service, mock_callback_port):
        """Test send_heartbeat handles exceptions."""
        mock_callback_port.report_heartbeat.side_effect = Exception("Network error")

        signal_data = HeartbeatSignal(timestamp=datetime.now(), progress={})
        result = await heartbeat_service.send_heartbeat("exec_001", signal_data)

        assert result is False

    def test_is_running(self, heartbeat_service):
        """Test is_running method."""
        assert heartbeat_service.is_running("exec_001") is False


class TestLifecycleService:
    """Tests for LifecycleService."""

    @pytest.fixture
    def mock_callback_port(self):
        """Create a mock callback port."""
        mock = AsyncMock()
        mock.report_lifecycle.return_value = True
        mock.report_heartbeat.return_value = True
        return mock

    @pytest.fixture
    def mock_heartbeat_port(self):
        """Create a mock heartbeat port."""
        mock = AsyncMock()
        mock.stop_all.return_value = None
        mock._tasks = {}
        return mock

    @pytest.fixture
    def lifecycle_service(self, mock_callback_port, mock_heartbeat_port):
        """Create a LifecycleService instance."""
        with patch.dict(os.environ, {"CONTAINER_ID": "container-123", "POD_NAME": "pod-456"}):
            return LifecycleService(
                callback_port=mock_callback_port,
                heartbeat_port=mock_heartbeat_port,
                executor_port=8080,
            )

    @pytest.mark.asyncio
    async def test_send_container_ready(self, lifecycle_service, mock_callback_port):
        """Test sending container_ready event."""
        result = await lifecycle_service.send_container_ready()

        assert result is True
        mock_callback_port.report_lifecycle.assert_called_once()

        # Verify the event has correct type
        event = mock_callback_port.report_lifecycle.call_args[0][0]
        assert event.event_type == "ready"
        assert event.container_id == "container-123"
        assert event.pod_name == "pod-456"

    @pytest.mark.asyncio
    async def test_send_container_ready_failure(self, lifecycle_service, mock_callback_port):
        """Test send_container_ready handles failure."""
        mock_callback_port.report_lifecycle.return_value = False

        result = await lifecycle_service.send_container_ready()
        assert result is False

    @pytest.mark.asyncio
    async def test_send_container_ready_exception(self, lifecycle_service, mock_callback_port):
        """Test send_container_ready handles exception."""
        mock_callback_port.report_lifecycle.side_effect = Exception("Network error")

        result = await lifecycle_service.send_container_ready()
        assert result is False

    @pytest.mark.asyncio
    async def test_send_container_exited(self, lifecycle_service, mock_callback_port):
        """Test sending container_exited event."""
        result = await lifecycle_service.send_container_exited(0, "normal")

        assert result is True
        mock_callback_port.report_lifecycle.assert_called_once()

        event = mock_callback_port.report_lifecycle.call_args[0][0]
        assert event.event_type == "exited"
        assert event.exit_code == 0
        assert event.exit_reason == ExitReason.NORMAL

    @pytest.mark.asyncio
    async def test_send_container_exited_with_sigterm(self, lifecycle_service, mock_callback_port):
        """Test sending container_exited with sigterm reason."""
        result = await lifecycle_service.send_container_exited(143, "sigterm")

        event = mock_callback_port.report_lifecycle.call_args[0][0]
        assert event.exit_reason == ExitReason.SIGTERM
        assert result is True

    @pytest.mark.asyncio
    async def test_send_container_exited_failure_and_invalid_reason(
        self, lifecycle_service, mock_callback_port
    ):
        """Test send_container_exited handles callback failure and invalid reason."""
        mock_callback_port.report_lifecycle.return_value = False

        assert await lifecycle_service.send_container_exited(1, "error") is False

        mock_callback_port.report_lifecycle.reset_mock()
        assert await lifecycle_service.send_container_exited(1, "not-a-reason") is False
        mock_callback_port.report_lifecycle.assert_not_called()

    @pytest.mark.asyncio
    async def test_shutdown_without_signal(self, lifecycle_service, mock_callback_port, mock_heartbeat_port):
        """Test shutdown without signal (normal exit)."""
        await lifecycle_service.shutdown()

        # Should send container_exited event
        mock_callback_port.report_lifecycle.assert_called()

        event = mock_callback_port.report_lifecycle.call_args[0][0]
        assert event.event_type == "exited"
        assert event.exit_reason == ExitReason.ERROR  # No signal = error

    @pytest.mark.asyncio
    async def test_shutdown_with_sigterm(self, lifecycle_service, mock_callback_port, mock_heartbeat_port):
        """Test shutdown with SIGTERM signal."""
        await lifecycle_service.shutdown(signal.SIGTERM)

        # Should send container_exited with SIGTERM reason
        event = mock_callback_port.report_lifecycle.call_args[0][0]
        assert event.event_type == "exited"
        assert event.exit_reason == ExitReason.SIGTERM
        assert event.exit_code == 143

    @pytest.mark.asyncio
    async def test_shutdown_with_sigkill(self, lifecycle_service, mock_callback_port, mock_heartbeat_port):
        """Test shutdown with SIGKILL signal."""
        await lifecycle_service.shutdown(signal.SIGKILL)

        event = mock_callback_port.report_lifecycle.call_args[0][0]
        assert event.exit_reason == ExitReason.SIGKILL
        assert event.exit_code == 137

    @pytest.mark.asyncio
    async def test_shutdown_stops_heartbeat_when_port_is_lifecycle_service(self, mock_callback_port):
        """Test shutdown invokes stop_all when heartbeat port is also a lifecycle service."""
        with patch.dict(os.environ, {"CONTAINER_ID": "container-123"}):
            service = LifecycleService(
                callback_port=mock_callback_port,
                heartbeat_port=mock_callback_port,
                executor_port=8080,
            )
        service._heartbeat_port = service
        service._tasks = {"exec-1": object()}
        service.stop_all = AsyncMock()
        service._report_execution_crash = AsyncMock()

        await service.shutdown(signal.SIGTERM)

        service._report_execution_crash.assert_awaited_once_with("exec-1")
        service.stop_all.assert_awaited_once()

    @pytest.mark.asyncio
    async def test_shutdown_twice(self, lifecycle_service, mock_callback_port, mock_heartbeat_port):
        """Test that calling shutdown twice only triggers once."""
        await lifecycle_service.shutdown(signal.SIGTERM)
        call_count = mock_callback_port.report_lifecycle.call_count

        await lifecycle_service.shutdown(signal.SIGKILL)
        # Should not increase call count
        assert mock_callback_port.report_lifecycle.call_count == call_count

    def test_map_exit_code_to_reason(self):
        """Test mapping exit codes to exit reason strings."""
        assert map_exit_code_to_reason(0) == "normal"
        assert map_exit_code_to_reason(143) == "sigterm"
        assert map_exit_code_to_reason(137) == "sigkill"
        assert map_exit_code_to_reason(134) == "oom_killed"
        assert map_exit_code_to_reason(1) == "error"
        assert map_exit_code_to_reason(255) == "error"

    def test_get_container_id(self, lifecycle_service):
        """Test getting container ID."""
        assert lifecycle_service.get_container_id() == "container-123"

    def test_is_shutting_down(self, lifecycle_service):
        """Test is_shutting_down method."""
        assert lifecycle_service.is_shutting_down() is False

    @pytest.mark.asyncio
    async def test_wait_for_shutdown(self, lifecycle_service, mock_callback_port, mock_heartbeat_port):
        """Test wait_for_shutdown method."""
        # Start shutdown in background
        shutdown_task = asyncio.create_task(lifecycle_service.shutdown())

        # Wait for shutdown to complete
        await lifecycle_service.wait_for_shutdown()

        assert lifecycle_service.is_shutting_down() is True

        # Clean up the task
        await shutdown_task

    def test_container_id_from_hostname(self, mock_callback_port, mock_heartbeat_port):
        """Test container ID detection from HOSTNAME."""
        with patch.dict(os.environ, {"HOSTNAME": "hostname-123"}, clear=True):
            if "CONTAINER_ID" in os.environ:
                del os.environ["CONTAINER_ID"]

            service = LifecycleService(
                callback_port=mock_callback_port,
                heartbeat_port=mock_heartbeat_port,
                executor_port=8080,
            )
            assert service.get_container_id() == "hostname-123"

    def test_container_id_fallback(self, mock_callback_port, mock_heartbeat_port):
        """Test container ID fallback to 'unknown'."""
        with patch.dict(os.environ, {}, clear=True):
            service = LifecycleService(
                callback_port=mock_callback_port,
                heartbeat_port=mock_heartbeat_port,
                executor_port=8080,
            )
            assert service.get_container_id() == "unknown"

    @pytest.mark.asyncio
    async def test_mark_active_executions_crashed_handles_report_errors(
        self, lifecycle_service
    ):
        """Test active execution crash reporting continues after individual failures."""
        lifecycle_service._heartbeat_port = lifecycle_service
        lifecycle_service._tasks = {"exec-1": object(), "exec-2": object()}
        lifecycle_service._report_execution_crash = AsyncMock(side_effect=[RuntimeError("boom"), None])

        await lifecycle_service._mark_active_executions_crashed()

        assert lifecycle_service._report_execution_crash.await_count == 2

    @pytest.mark.asyncio
    async def test_report_execution_crash_success_and_callback_failure(
        self, lifecycle_service, mock_callback_port
    ):
        """Test crash heartbeat payload and internal exception handling."""
        await lifecycle_service._report_execution_crash("exec-1")

        execution_id, heartbeat = mock_callback_port.report_heartbeat.call_args.args
        assert execution_id == "exec-1"
        assert heartbeat.progress["status"] == "crashed"
        assert heartbeat.progress["reason"] == "executor_shutdown"

        mock_callback_port.report_heartbeat.side_effect = RuntimeError("callback failed")
        await lifecycle_service._report_execution_crash("exec-2")

    def test_register_signal_handlers_registers_sigterm_and_sigint(self):
        """Test signal handler registration without installing real handlers."""
        registered = {}

        def fake_signal(signum, handler):
            registered[signum] = handler

        with patch("executor.application.services.lifecycle_service.signal.signal", fake_signal):
            register_signal_handlers()

        assert signal.SIGTERM in registered
        assert signal.SIGINT in registered

    def test_registered_signal_handler_schedules_shutdown(self, lifecycle_service):
        """Test the generated signal handler schedules shutdown when service exists."""
        registered = {}
        created_tasks = []

        def fake_signal(signum, handler):
            registered[signum] = handler

        with (
            patch("executor.application.services.lifecycle_service.signal.signal", fake_signal),
            patch(
                "executor.application.services.lifecycle_service.get_lifecycle_service",
                return_value=lifecycle_service,
            ),
            patch(
                "executor.application.services.lifecycle_service.asyncio.create_task",
                side_effect=lambda coro: created_tasks.append(coro),
            ),
        ):
            register_signal_handlers()
            registered[signal.SIGTERM](signal.SIGTERM, None)

        assert len(created_tasks) == 1
        created_tasks[0].close()

    def test_registered_signal_handler_no_service(self):
        """Test the generated signal handler is a no-op without a lifecycle service."""
        registered = {}

        def fake_signal(signum, handler):
            registered[signum] = handler

        with (
            patch("executor.application.services.lifecycle_service.signal.signal", fake_signal),
            patch(
                "executor.application.services.lifecycle_service.get_lifecycle_service",
                return_value=None,
            ),
        ):
            register_signal_handlers()
            registered[signal.SIGTERM](signal.SIGTERM, None)


class TestGlobalHeartbeatService:
    """Tests for global heartbeat service functions."""

    def test_get_heartbeat_service_none(self):
        """Test get_heartbeat_service returns None initially."""
        # Reset global instance
        import executor.application.services.heartbeat_service as hs
        hs._heartbeat_service = None

        assert get_heartbeat_service() is None

    def test_register_and_get_heartbeat_service(self, mock_callback_port):
        """Test registering and getting heartbeat service."""
        service = HeartbeatService(callback_port=mock_callback_port)
        register_heartbeat_service(service)

        assert get_heartbeat_service() is service

        # Reset
        import executor.application.services.heartbeat_service as hs
        hs._heartbeat_service = None


class TestGlobalLifecycleService:
    """Tests for global lifecycle service functions."""

    def test_get_lifecycle_service_none(self):
        """Test get_lifecycle_service returns None initially."""
        # Reset global instance
        import executor.application.services.lifecycle_service as ls
        ls._lifecycle_service = None

        assert get_lifecycle_service() is None

    def test_register_and_get_lifecycle_service(self, mock_callback_port, mock_heartbeat_port):
        """Test registering and getting lifecycle service."""
        with patch.dict(os.environ, {"CONTAINER_ID": "test-container"}):
            service = LifecycleService(
                callback_port=mock_callback_port,
                heartbeat_port=mock_heartbeat_port,
            )
            register_lifecycle_service(service)

            assert get_lifecycle_service() is service

        # Reset
        import executor.application.services.lifecycle_service as ls
        ls._lifecycle_service = None


# Fixtures for global tests
@pytest.fixture
def mock_callback_port():
    """Create a mock callback port."""
    mock = AsyncMock()
    mock.report_lifecycle.return_value = True
    mock.report_heartbeat.return_value = True
    return mock


@pytest.fixture
def mock_heartbeat_port():
    """Create a mock heartbeat port."""
    mock = AsyncMock()
    mock.stop_all.return_value = None
    mock._tasks = {}
    return mock
