"""
执行状态值对象

定义执行和会话的所有可能状态。
"""
from enum import Enum
from dataclasses import dataclass


class SessionStatus(str, Enum):
    """会话状态枚举"""
    CREATING = "creating"
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"
    TIMEOUT = "timeout"
    TERMINATED = "terminated"


class ExecutionStatus(str, Enum):
    """执行状态枚举"""
    PENDING = "pending"
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"
    TIMEOUT = "timeout"
    CRASHED = "crashed"


@dataclass(frozen=True)
class ExecutionState:
    """执行状态值对象（不可变）"""
    status: ExecutionStatus
    exit_code: int | None = None
    error_message: str | None = None

    def is_terminal(self) -> bool:
        """是否为终态（不可再变更）"""
        return self.status in {
            ExecutionStatus.COMPLETED,
            ExecutionStatus.FAILED,
            ExecutionStatus.TIMEOUT
        }

    def can_retry(self) -> bool:
        """是否可以重试"""
        return self.status == ExecutionStatus.CRASHED
