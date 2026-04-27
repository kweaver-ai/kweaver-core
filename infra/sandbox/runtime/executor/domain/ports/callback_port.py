"""
Callback Port Interface

Defines the contract for Control Plane callback operations.
This is an output port - implemented by infrastructure layer.
"""

from abc import ABC, abstractmethod

from executor.domain.value_objects import ExecutionResult, HeartbeatSignal, ContainerLifecycleEvent


class ICallbackPort(ABC):
    """
    Port interface for Control Plane callback operations.

    Defines the contract for reporting execution results,
    heartbeat signals, and lifecycle events to the Control Plane.
    """

    @abstractmethod
    async def report_result(
        self,
        execution_id: str,
        result: ExecutionResult,
    ) -> bool:
        """
        Report execution result to Control Plane.

        Args:
            execution_id: Unique execution identifier
            result: Execution result to report

        Returns:
            True if successful, False otherwise
        """
        pass

    @abstractmethod
    async def report_heartbeat(
        self,
        execution_id: str,
        signal: HeartbeatSignal,
    ) -> bool:
        """
        Report heartbeat signal to Control Plane.

        Args:
            execution_id: Unique execution identifier
            signal: Heartbeat signal

        Returns:
            True if successful, False otherwise
        """
        pass

    @abstractmethod
    async def report_lifecycle(
        self,
        event: ContainerLifecycleEvent,
    ) -> bool:
        """
        Report container lifecycle event to Control Plane.

        Args:
            event: Lifecycle event (ready or exited)

        Returns:
            True if successful, False otherwise
        """
        pass

    @abstractmethod
    async def close(self) -> None:
        """
        Close the callback client and cleanup resources.

        Should be called during shutdown.
        """
        pass
