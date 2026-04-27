"""
执行仓储实现

使用 SQLAlchemy 实现执行仓储接口。
按照数据表命名规范使用 f_ 前缀字段名。
"""
import time
from typing import List, Optional
from datetime import datetime
from sqlalchemy import select, update, delete, func
from sqlalchemy.ext.asyncio import AsyncSession

from src.domain.repositories.execution_repository import IExecutionRepository
from src.domain.entities.execution import Execution
from src.infrastructure.persistence.models.execution_model import ExecutionModel


class SqlExecutionRepository(IExecutionRepository):
    """
    执行仓储实现

    这是基础设施层的 Adapter，实现领域层定义的 Port。
    """

    def __init__(self, session: AsyncSession):
        self._session = session

    async def save(self, execution: Execution) -> None:
        """保存执行记录"""
        import json
        model = await self._session.get(ExecutionModel, execution.id)
        now_ms = int(time.time() * 1000)

        if model:
            # 更新现有记录
            model.f_session_id = execution.session_id
            model.f_code = execution.code
            model.f_language = execution.language
            model.f_status = execution.state.status.value
            model.f_stdout = execution.stdout
            model.f_stderr = execution.stderr
            model.f_exit_code = execution.state.exit_code or 0
            model.f_return_value = json.dumps(execution.return_value, ensure_ascii=False) if execution.return_value else ""
            model.f_metrics = json.dumps(execution.metrics, ensure_ascii=False) if execution.metrics else ""
            model.f_error_message = execution.state.error_message or ""
            model.f_completed_at = int(execution.completed_at.timestamp() * 1000) if execution.completed_at else 0
            model.f_updated_at = now_ms
        else:
            # 创建新记录
            model = ExecutionModel.from_entity(execution)
            self._session.add(model)

        await self._session.flush()

    async def commit(self) -> None:
        """Explicitly commit the transaction"""
        await self._session.commit()

    async def find_by_id(self, execution_id: str) -> Optional[Execution]:
        """根据 ID 查找执行记录"""
        # Use a fresh query to avoid stale data from session cache
        # This is important for the sync execution polling loop
        stmt = select(ExecutionModel).where(ExecutionModel.f_id == execution_id)
        result = await self._session.execute(stmt)
        model = result.scalar_one_or_none()
        return model.to_entity() if model else None

    async def find_by_session_id(
        self,
        session_id: str,
        limit: int = 100
    ) -> List[Execution]:
        """根据会话 ID 查找执行记录"""
        stmt = (
            select(ExecutionModel)
            .where(ExecutionModel.f_session_id == session_id)
            .order_by(ExecutionModel.f_created_at.desc())
            .limit(limit)
        )
        result = await self._session.execute(stmt)
        return [model.to_entity() for model in result.scalars().all()]

    async def find_by_status(self, status: str, limit: int = 100) -> List[Execution]:
        """根据状态查找执行记录"""
        stmt = (
            select(ExecutionModel)
            .where(ExecutionModel.f_status == status)
            .limit(limit)
        )
        result = await self._session.execute(stmt)
        return [model.to_entity() for model in result.scalars().all()]

    async def find_crashed_executions(self, max_retry_count: int) -> List[Execution]:
        """查找可重试的崩溃执行"""
        # 注意：retry_count 字段不在数据库模型中，这里返回空列表
        # 如需支持，需要添加 f_retry_count 字段到 ExecutionModel
        return []

    async def find_heartbeat_timeouts(
        self,
        timeout_threshold: datetime
    ) -> List[Execution]:
        """查找心跳超时的执行"""
        # 注意：last_heartbeat_at 字段不在数据库模型中
        return []

    async def delete(self, execution_id: str) -> None:
        """删除执行记录"""
        stmt = delete(ExecutionModel).where(ExecutionModel.f_id == execution_id)
        await self._session.execute(stmt)
        await self._session.flush()

    async def delete_by_session_id(self, session_id: str) -> None:
        """删除会话的所有执行记录"""
        stmt = delete(ExecutionModel).where(ExecutionModel.f_session_id == session_id)
        await self._session.execute(stmt)
        await self._session.flush()

    async def count_by_status(self, status: str) -> int:
        """统计指定状态的执行数量"""
        stmt = (
            select(func.count())
            .select_from(ExecutionModel)
            .where(ExecutionModel.f_status == status)
        )
        result = await self._session.execute(stmt)
        return result.scalar() or 0
