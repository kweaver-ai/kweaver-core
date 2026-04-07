"""
执行实体

表示一次代码执行，是会话聚合的一部分。
"""
from dataclasses import dataclass, field
from datetime import datetime

from src.domain.value_objects.execution_status import ExecutionStatus, ExecutionState
from src.domain.value_objects.artifact import Artifact


@dataclass
class Execution:
    """
    执行实体

    表示一次代码执行的完整生命周期。
    """
    id: str
    session_id: str
    code: str
    language: str
    state: ExecutionState
    timeout: int = 300  # 超时时间（秒）
    event_data: dict | None = None  # 事件数据（handler 函数输入）
    created_at: datetime = field(default_factory=datetime.now)
    completed_at: datetime | None = None
    execution_time: float | None = None  # 执行耗时（秒）
    stdout: str = ""
    stderr: str = ""
    artifacts: list[Artifact] = field(default_factory=list)
    retry_count: int = 0
    last_heartbeat_at: datetime | None = None
    # 新增字段：handler 返回值和性能指标
    return_value: dict | None = None  # handler 函数返回值（JSON 可序列化）
    metrics: dict | None = None  # 性能指标（JSON 对象）

    def __post_init__(self):
        """初始化后验证"""
        if not self.code:
            raise ValueError("code cannot be empty")
        if not self.language:
            raise ValueError("language cannot be empty")

    # ============== 领域行为 ==============

    def mark_running(self) -> None:
        """标记为运行中"""
        if self.state.status != ExecutionStatus.PENDING:
            raise ValueError(f"Cannot mark execution as running from status: {self.state.status}")

        self.state = ExecutionState(status=ExecutionStatus.RUNNING)
        self.last_heartbeat_at = datetime.now()

    def mark_completed(
        self,
        stdout: str,
        stderr: str,
        exit_code: int,
        execution_time: float,
        artifacts: list[Artifact] | None = None,
        return_value: dict | None = None,
        metrics: dict | None = None
    ) -> None:
        """标记为已完成"""
        if self.state.status != ExecutionStatus.RUNNING:
            raise ValueError(f"Cannot mark execution as completed from status: {self.state.status}")

        self.state = ExecutionState(
            status=ExecutionStatus.COMPLETED,
            exit_code=exit_code
        )
        self.stdout = stdout
        self.stderr = stderr
        self.execution_time = execution_time
        self.artifacts = artifacts or []
        self.return_value = return_value  # handler 返回值
        self.metrics = metrics  # 性能指标
        self.completed_at = datetime.now()
        self.last_heartbeat_at = datetime.now()

    def mark_failed(
        self,
        error_message: str,
        exit_code: int | None = None,
        stdout: str | None = None,
        stderr: str | None = None,
    ) -> None:
        """标记为失败"""
        self.state = ExecutionState(
            status=ExecutionStatus.FAILED,
            exit_code=exit_code,
            error_message=error_message
        )
        self.stdout = stdout or ""
        self.stderr = stderr or ""
        self.completed_at = datetime.now()

    def mark_timeout(self) -> None:
        """标记为超时"""
        self.state = ExecutionState(status=ExecutionStatus.TIMEOUT)
        self.completed_at = datetime.now()

    def mark_crashed(self) -> None:
        """标记为崩溃（可重试）"""
        self.state = ExecutionState(status=ExecutionStatus.CRASHED)

    def update_heartbeat(self) -> None:
        """更新心跳时间"""
        self.last_heartbeat_at = datetime.now()

    def increment_retry_count(self) -> None:
        """增加重试计数"""
        self.retry_count += 1

    # ============== 领域查询 ==============

    def is_running(self) -> bool:
        """是否正在运行"""
        return self.state.status == ExecutionStatus.RUNNING

    def is_terminal(self) -> bool:
        """是否为终态"""
        return self.state.is_terminal()

    def can_retry(self, max_retries: int = 3) -> bool:
        """是否可以重试"""
        return self.state.can_retry() and self.retry_count < max_retries

    def is_heartbeat_timeout(self, timeout_seconds: int = 15) -> bool:
        """心跳是否超时"""
        if not self.last_heartbeat_at or not self.is_running():
            return False
        elapsed = (datetime.now() - self.last_heartbeat_at).total_seconds()
        return elapsed > timeout_seconds
