"""
Isolation Port Interface

Defines the contract for process isolation operations.
This is an output port - implemented by infrastructure layer (Bubblewrap).
"""

from abc import ABC, abstractmethod

from executor.domain.entities import Execution
from executor.domain.value_objects import ExecutionResult


class IIsolationPort(ABC):
    """
    Port interface for process isolation operations.

    Defines the contract for executing code in isolation using
    Bubblewrap or similar isolation mechanisms.
    """

    @abstractmethod
    async def execute(self, execution: Execution) -> ExecutionResult:
        """
        Execute code in isolated environment.

        Args:
            execution: Execution entity with code and context

        Returns:
            ExecutionResult with stdout, stderr, exit code, timing

        Raises:
            subprocess.TimeoutExpired: If execution exceeds timeout
            Exception: For isolation errors
        """
        pass

    @abstractmethod
    def is_available(self) -> bool:
        """
        Check if isolation mechanism is available.

        Returns:
            True if bwrap or similar is available, False otherwise
        """
        pass

    @abstractmethod
    def get_version(self) -> str:
        """
        Get isolation mechanism version.

        Returns:
            Version string (e.g., "bwrap 1.7.0")
        """
        pass
