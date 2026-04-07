"""
Executor Port Interface

Defines the contract for code execution operations.
This is an input port - called by application layer.
"""

from abc import ABC, abstractmethod
from typing import Optional

from executor.domain.entities import Execution
from executor.domain.value_objects import ExecutionResult


class IExecutorPort(ABC):
    """
    Port interface for code execution operations.

    Defines the contract for executing code in isolation.
    Implementations are provided by infrastructure layer.
    """

    @abstractmethod
    async def execute(
        self,
        execution: Execution,
    ) -> ExecutionResult:
        """
        Execute code in isolation.

        Args:
            execution: Execution entity with code and context

        Returns:
            ExecutionResult with stdout, stderr, exit code, and artifacts

        Raises:
            asyncio.TimeoutError: If execution exceeds timeout
            Exception: For execution errors
        """
        pass

    @abstractmethod
    def validate_execution(self, execution: Execution) -> bool:
        """
        Validate execution request before execution.

        Args:
            execution: Execution entity to validate

        Returns:
            True if valid, False otherwise
        """
        pass
