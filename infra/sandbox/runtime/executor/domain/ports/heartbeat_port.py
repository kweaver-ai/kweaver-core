"""
Heartbeat Port Interface

Defines the contract for heartbeat operations.
This is an output port - implemented by application layer.
"""

from abc import ABC, abstractmethod

from executor.domain.ports.callback_port import HeartbeatSignal


class IHeartbeatPort(ABC):
    """
    Port interface for heartbeat operations.

    Defines the contract for sending periodic heartbeat signals
    during code execution to indicate liveness.
    """

    @abstractmethod
    async def start_heartbeat(self, execution_id: str) -> None:
        """
        Start heartbeat for an execution.

        Begins sending heartbeat signals every 5 seconds
        until stopped.

        Args:
            execution_id: Unique execution identifier
        """
        pass

    @abstractmethod
    async def stop_heartbeat(self, execution_id: str) -> None:
        """
        Stop heartbeat for an execution.

        Args:
            execution_id: Unique execution identifier
        """
        pass

    @abstractmethod
    async def send_heartbeat(
        self,
        execution_id: str,
        signal: HeartbeatSignal,
    ) -> bool:
        """
        Send a single heartbeat signal.

        Args:
            execution_id: Unique execution identifier
            signal: Heartbeat signal to send

        Returns:
            True if successful, False otherwise
        """
        pass

    @abstractmethod
    def is_running(self, execution_id: str) -> bool:
        """
        Check if heartbeat is running for an execution.

        Args:
            execution_id: Unique execution identifier

        Returns:
            True if heartbeat is running, False otherwise
        """
        pass
