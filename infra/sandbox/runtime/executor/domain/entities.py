"""
Execution Entities

Core domain entities for code execution within the sandbox.
"""

from dataclasses import dataclass, field
from datetime import datetime
from typing import Optional
from executor.domain.value_objects import ExecutionStatus, ExecutionContext


@dataclass
class Execution:
    """
    Represents a single code execution request within the sandbox.

    This entity tracks the lifecycle of a code execution, from request
    to completion, including all metadata and results.
    """

    execution_id: str
    session_id: str
    code: str
    language: str
    context: ExecutionContext
    status: ExecutionStatus = ExecutionStatus.PENDING
    result: Optional["ExecutionResult"] = None
    created_at: datetime = field(default_factory=datetime.utcnow)
    started_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None
    retry_count: int = 0
    error_message: Optional[str] = None

    def mark_as_running(self) -> None:
        """Mark the execution as running."""
        self.status = ExecutionStatus.RUNNING
        self.started_at = datetime.utcnow()

    def mark_as_completed(self, result: "ExecutionResult") -> None:
        """Mark the execution as completed with results."""
        self.status = ExecutionStatus.COMPLETED
        self.result = result
        self.completed_at = datetime.utcnow()

    def mark_as_failed(self, error: str) -> None:
        """Mark the execution as failed."""
        self.status = ExecutionStatus.FAILED
        self.error_message = error
        self.completed_at = datetime.utcnow()

    def mark_as_timeout(self) -> None:
        """Mark the execution as timed out."""
        self.status = ExecutionStatus.TIMEOUT
        self.completed_at = datetime.utcnow()

    def increment_retry(self) -> None:
        """Increment the retry counter."""
        self.retry_count += 1

    @property
    def duration_ms(self) -> Optional[int]:
        """Calculate execution duration in milliseconds."""
        if self.started_at and self.completed_at:
            delta = self.completed_at - self.started_at
            return int(delta.total_seconds() * 1000)
        return None

    def can_retry(self, max_retries: int = 3) -> bool:
        """Check if this execution can be retried."""
        return self.retry_count < max_retries
