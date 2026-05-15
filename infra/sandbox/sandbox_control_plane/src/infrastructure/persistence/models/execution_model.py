"""
执行 ORM 模型

SQLAlchemy 模型定义，用于数据库持久化。
按照数据表命名规范: t_{module}_{entity}, f_{field_name}
"""

from datetime import datetime

from sqlalchemy import Column, String, Integer, BigInteger, Text, Index
from sqlalchemy.orm import Mapped, mapped_column
from sqlalchemy import func

from src.infrastructure.persistence.database import Base


class ExecutionModel(Base):
    """
    执行 ORM 模型 - t_sandbox_execution

    这是基础设施层的实现细节，映射到数据库表。
    按照数据表命名规范实现。
    """

    __tablename__ = "t_sandbox_execution"

    # Primary fields
    f_id: Mapped[str] = mapped_column(String(255), primary_key=True)
    f_session_id: Mapped[str] = mapped_column(String(255), nullable=False)

    # Status
    f_status: Mapped[str] = mapped_column(String(20), nullable=False, default="pending")

    # Code and execution
    f_code = Column(Text, nullable=False, default="")
    f_language: Mapped[str] = mapped_column(String(32), nullable=False)
    f_entrypoint: Mapped[str] = mapped_column(String(255), nullable=False, default="")
    f_event_data = Column(Text, nullable=False, default="")
    f_timeout_sec = Column(Integer, nullable=False)

    # Results
    f_return_value = Column(Text, nullable=False, default="")
    f_stdout = Column(Text, nullable=False, default="")
    f_stderr = Column(Text, nullable=False, default="")
    f_exit_code = Column(Integer, nullable=False, default=0)
    f_metrics = Column(Text, nullable=False, default="")
    f_error_message = Column(Text, nullable=False, default="")

    # Timestamps (BIGINT - millisecond timestamps)
    f_started_at = Column(BigInteger, nullable=False, default=0)
    f_completed_at = Column(BigInteger, nullable=False, default=0)

    # Audit fields
    f_created_at = Column(BigInteger, nullable=False, default=0)
    f_created_by = Column(String(40), nullable=False, default="")
    f_updated_at = Column(BigInteger, nullable=False, default=0)
    f_updated_by = Column(String(40), nullable=False, default="")
    f_deleted_at = Column(BigInteger, nullable=False, default=0)
    f_deleted_by = Column(String(36), nullable=False, default="")

    # Indexes
    __table_args__ = (
        Index("t_sandbox_execution_idx_session_id", "f_session_id"),
        Index("t_sandbox_execution_idx_status", "f_status"),
        Index("t_sandbox_execution_idx_created_at", "f_created_at"),
        Index("t_sandbox_execution_idx_deleted_at", "f_deleted_at"),
        Index("t_sandbox_execution_idx_created_by", "f_created_by"),
    )

    def to_entity(self):
        """转换为领域实体"""
        from src.domain.entities.execution import Execution
        from src.domain.value_objects.execution_status import ExecutionStatus, ExecutionState

        return Execution(
            id=self.f_id,
            session_id=self.f_session_id,
            code=self.f_code or "",
            language=self.f_language,
            timeout=self.f_timeout_sec,
            event_data=self._parse_json(self.f_event_data),
            state=ExecutionState(
                status=ExecutionStatus(self.f_status),
                exit_code=self.f_exit_code,
                error_message=self.f_error_message or None,
            ),
            created_at=self._millis_to_datetime(self.f_created_at) or datetime.now(),
            completed_at=self._millis_to_datetime(self.f_completed_at),
            execution_time=None,  # Can be calculated from started_at/completed_at
            stdout=self.f_stdout or "",
            stderr=self.f_stderr or "",
            artifacts=[],  # Loaded separately if needed
            retry_count=0,  # Not in database schema
            last_heartbeat_at=None,  # Not in database schema
            return_value=self._parse_json(self.f_return_value),
            metrics=self._parse_json(self.f_metrics),
        )

    @classmethod
    def from_entity(cls, execution):
        """从领域实体创建 ORM 模型"""
        import json

        now_ms = int(datetime.now().timestamp() * 1000)

        return cls(
            f_id=execution.id,
            f_session_id=execution.session_id,
            f_status=execution.state.status.value,
            f_code=execution.code,
            f_language=execution.language,
            f_timeout_sec=execution.timeout,
            f_entrypoint="",
            f_event_data=(
                json.dumps(execution.event_data, ensure_ascii=False) if execution.event_data else ""
            ),
            f_return_value=(
                json.dumps(execution.return_value, ensure_ascii=False)
                if execution.return_value
                else ""
            ),
            f_stdout=execution.stdout,
            f_stderr=execution.stderr,
            f_exit_code=execution.state.exit_code or 0,
            f_metrics=(
                json.dumps(execution.metrics, ensure_ascii=False) if execution.metrics else ""
            ),
            f_error_message=execution.state.error_message or "",
            f_started_at=0,
            f_completed_at=(
                int(execution.completed_at.timestamp() * 1000) if execution.completed_at else 0
            ),
            # 审计字段
            f_created_at=(
                int(execution.created_at.timestamp() * 1000) if execution.created_at else now_ms
            ),
            f_created_by="",
            f_updated_at=now_ms,
            f_updated_by="",
            f_deleted_at=0,
            f_deleted_by="",
        )

    def _parse_json(self, value: str):
        """安全解析 JSON 字符串"""
        if not value or value.strip() == "":
            return None
        try:
            import json

            return json.loads(value)
        except (json.JSONDecodeError, ValueError):
            return None

    def _millis_to_datetime(self, millis: int):
        """将毫秒时间戳转换为 datetime"""
        if not millis or millis == 0:
            return None
        try:
            return datetime.fromtimestamp(millis / 1000)
        except (ValueError, OSError):
            return None
