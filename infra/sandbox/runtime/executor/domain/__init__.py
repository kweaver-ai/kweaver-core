"""
Executor Domain Layer

This module contains the core domain logic for the sandbox executor,
including execution entities, value objects, and domain services.
"""

from .entities import Execution, ExecutionContext
from .value_objects import ExecutionResult, ExecutionStatus, Artifact

__all__ = [
    "Execution",
    "ExecutionContext",
    "ExecutionResult",
    "ExecutionStatus",
    "Artifact",
]
