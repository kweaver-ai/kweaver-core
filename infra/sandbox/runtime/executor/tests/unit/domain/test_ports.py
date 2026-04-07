"""
Unit tests for Domain Ports.

Tests that port interfaces are properly defined and can be implemented.
"""

import pytest
from abc import ABC

from executor.domain.ports import (
    IExecutorPort,
    ICallbackPort,
    IIsolationPort,
    IArtifactScannerPort,
    IHeartbeatPort,
    ILifecyclePort,
)
from executor.domain.entities import Execution
from executor.domain.value_objects import (
    ExecutionResult,
    ExecutionStatus,
    Artifact,
    ArtifactType,
    HeartbeatSignal,
    ContainerLifecycleEvent,
    ExitReason,
)
from pathlib import Path
from datetime import datetime


class TestIExecutorPort:
    """Tests for IExecutorPort interface."""

    def test_port_is_abstract(self):
        """Test that IExecutorPort is an abstract base class."""
        assert issubclass(IExecutorPort, ABC)

    def test_port_requires_execute_method(self):
        """Test that IExecutorPort requires execute method."""
        # Should not be able to instantiate without implementation
        with pytest.raises(TypeError):
            IExecutorPort()

    def test_port_can_be_implemented(self):
        """Test that IExecutorPort can be implemented."""

        class MockExecutor(IExecutorPort):
            async def execute(self, execution: Execution) -> ExecutionResult:
                return ExecutionResult(
                    execution_id=execution.execution_id,
                    status=ExecutionStatus.SUCCESS,
                    stdout="",
                    stderr="",
                    exit_code=0,
                    execution_time_ms=0,
                )

            def validate_execution(self, execution: Execution) -> bool:
                return True

        mock = MockExecutor()
        assert isinstance(mock, IExecutorPort)


class TestICallbackPort:
    """Tests for ICallbackPort interface."""

    def test_port_is_abstract(self):
        """Test that ICallbackPort is an abstract base class."""
        assert issubclass(ICallbackPort, ABC)

    def test_port_requires_callback_methods(self):
        """Test that ICallbackPort requires callback methods."""
        # Should not be able to instantiate without implementation
        with pytest.raises(TypeError):
            ICallbackPort()

    def test_port_can_be_implemented(self):
        """Test that ICallbackPort can be implemented."""

        class MockCallback(ICallbackPort):
            async def report_result(self, execution_id: str, result: ExecutionResult) -> bool:
                return True

            async def report_heartbeat(self, execution_id: str, signal: HeartbeatSignal) -> bool:
                return True

            async def report_lifecycle(self, event: ContainerLifecycleEvent) -> bool:
                return True

            async def close(self) -> None:
                pass

        mock = MockCallback()
        assert isinstance(mock, ICallbackPort)


class TestIIsolationPort:
    """Tests for IIsolationPort interface."""

    def test_port_is_abstract(self):
        """Test that IIsolationPort is an abstract base class."""
        assert issubclass(IIsolationPort, ABC)

    def test_port_can_be_implemented(self):
        """Test that IIsolationPort can be implemented."""

        class MockIsolation(IIsolationPort):
            async def execute(self, execution: Execution) -> ExecutionResult:
                return ExecutionResult(
                    status=ExecutionStatus.SUCCESS,
                    stdout="",
                    stderr="",
                    exit_code=0,
                    execution_time_ms=0,
                )

            def is_available(self) -> bool:
                return True

            def get_version(self) -> str:
                return "bwrap 1.7.0"

        mock = MockIsolation()
        assert isinstance(mock, IIsolationPort)


class TestIArtifactScannerPort:
    """Tests for IArtifactScannerPort interface."""

    def test_port_is_abstract(self):
        """Test that IArtifactScannerPort is an abstract base class."""
        assert issubclass(IArtifactScannerPort, ABC)

    def test_port_can_be_implemented(self):
        """Test that IArtifactScannerPort can be implemented."""

        class MockScanner(IArtifactScannerPort):
            def collect_artifacts(
                self,
                workspace_path: Path,
                include_hidden: bool = False,
                include_temp: bool = False,
            ) -> list:
                return []

            def snapshot(self, workspace_path: Path) -> set:
                return set()

        mock = MockScanner()
        assert isinstance(mock, IArtifactScannerPort)


class TestIHeartbeatPort:
    """Tests for IHeartbeatPort interface."""

    def test_port_is_abstract(self):
        """Test that IHeartbeatPort is an abstract base class."""
        assert issubclass(IHeartbeatPort, ABC)

    def test_port_can_be_implemented(self):
        """Test that IHeartbeatPort can be implemented."""

        class MockHeartbeat(IHeartbeatPort):
            async def start_heartbeat(self, execution_id: str) -> None:
                pass

            async def stop_heartbeat(self, execution_id: str) -> None:
                pass

            async def send_heartbeat(self, execution_id: str, signal: HeartbeatSignal) -> bool:
                return True

            def is_running(self, execution_id: str) -> bool:
                return False

            async def stop_all(self) -> None:
                pass

        mock = MockHeartbeat()
        assert isinstance(mock, IHeartbeatPort)


class TestILifecyclePort:
    """Tests for ILifecyclePort interface."""

    def test_port_is_abstract(self):
        """Test that ILifecyclePort is an abstract base class."""
        assert issubclass(ILifecyclePort, ABC)

    def test_port_can_be_implemented(self):
        """Test that ILifecyclePort can be implemented."""

        class MockLifecycle(ILifecyclePort):
            async def send_container_ready(self) -> bool:
                return True

            async def send_container_exited(self, exit_code: int, exit_reason: str) -> bool:
                return True

            async def shutdown(self, signum: int = None) -> None:
                pass

            def get_container_id(self) -> str:
                return "container-123"

            def is_shutting_down(self) -> bool:
                return False

        mock = MockLifecycle()
        assert isinstance(mock, ILifecyclePort)


class TestPortCompliance:
    """Tests that ports follow hexagonal architecture principles."""

    def test_ports_are_abstract(self):
        """Test that all ports are abstract base classes."""
        ports = [
            IExecutorPort,
            ICallbackPort,
            IIsolationPort,
            IArtifactScannerPort,
            IHeartbeatPort,
            ILifecyclePort,
        ]

        for port in ports:
            assert issubclass(port, ABC), f"{port.__name__} should be abstract"

    def test_ports_cannot_be_instantiated(self):
        """Test that ports cannot be directly instantiated."""
        ports = [
            IExecutorPort,
            ICallbackPort,
            IIsolationPort,
            IArtifactScannerPort,
            IHeartbeatPort,
            ILifecyclePort,
        ]

        for port in ports:
            with pytest.raises(TypeError):
                port()

    def test_ports_define_clear_contracts(self):
        """Test that ports define clear method contracts."""
        # IExecutorPort should have execute and validate_execution
        assert hasattr(IExecutorPort, 'execute')
        assert hasattr(IExecutorPort, 'validate_execution')

        # ICallbackPort should have reporting methods
        assert hasattr(ICallbackPort, 'report_result')
        assert hasattr(ICallbackPort, 'report_heartbeat')
        assert hasattr(ICallbackPort, 'report_lifecycle')

        # IIsolationPort should have execute
        assert hasattr(IIsolationPort, 'execute')

        # IArtifactScannerPort should have collect and snapshot
        assert hasattr(IArtifactScannerPort, 'collect_artifacts')
        assert hasattr(IArtifactScannerPort, 'snapshot')

        # IHeartbeatPort should have heartbeat management
        assert hasattr(IHeartbeatPort, 'start_heartbeat')
        assert hasattr(IHeartbeatPort, 'stop_heartbeat')

        # ILifecyclePort should have lifecycle methods
        assert hasattr(ILifecyclePort, 'send_container_ready')
        assert hasattr(ILifecyclePort, 'shutdown')
