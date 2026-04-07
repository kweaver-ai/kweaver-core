"""
Lifecycle Port Interface

Defines the contract for container lifecycle operations.
This is an output port - implemented by application layer.
"""

from abc import ABC, abstractmethod

from executor.domain.ports.callback_port import ContainerLifecycleEvent


class ILifecyclePort(ABC):
    """
    Port interface for container lifecycle operations.

    Defines the contract for managing container lifecycle events
    including startup (container_ready) and shutdown (container_exited).
    """

    @abstractmethod
    async def send_container_ready(self) -> bool:
        """
        Send container_ready event to Control Plane on startup.

        Should be called after HTTP server starts listening.

        Returns:
            True if successful, False otherwise
        """
        pass

    @abstractmethod
    async def send_container_exited(
        self,
        exit_code: int,
        exit_reason: str,
    ) -> bool:
        """
        Send container_exited event to Control Plane on shutdown.

        Args:
            exit_code: Process exit code
            exit_reason: Reason for exit (normal, sigterm, sigkill, oom_killed, error)

        Returns:
            True if successful, False otherwise
        """
        pass

    @abstractmethod
    async def shutdown(self, signum: int = None) -> None:
        """
        Handle graceful shutdown.

        Marks active executions as crashed, sends container_exited,
        and waits for shutdown to complete.

        Args:
            signum: Signal number (SIGTERM=15, SIGKILL=9, etc.)
        """
        pass

    @abstractmethod
    def get_container_id(self) -> str:
        """
        Get container ID from environment.

        Checks CONTAINER_ID and HOSTNAME environment variables.

        Returns:
            Container identifier or "unknown" if not detected
        """
        pass

    @abstractmethod
    def is_shutting_down(self) -> bool:
        """
        Check if shutdown is in progress.

        Returns:
            True if shutting down, False otherwise
        """
        pass
