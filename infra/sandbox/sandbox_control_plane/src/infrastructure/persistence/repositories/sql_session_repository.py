"""
会话仓储实现

使用 SQLAlchemy 实现会话仓储接口。
按照数据表命名规范使用 f_ 前缀字段名。
"""

import time
from typing import List, Optional
from datetime import datetime
from sqlalchemy import select, update, delete, func
from sqlalchemy.ext.asyncio import AsyncSession

from src.domain.repositories.session_repository import ISessionRepository
from src.domain.entities.session import Session
from src.infrastructure.persistence.models.session_model import SessionModel


class SqlSessionRepository(ISessionRepository):
    """
    会话仓储实现

    这是基础设施层的 Adapter，实现领域层定义的 Port。
    """

    def __init__(self, session: AsyncSession, execution_repo=None):
        self._session = session
        self._execution_repo = execution_repo

    async def save(self, session: Session) -> None:
        """保存会话"""
        import json

        model = await self._session.get(SessionModel, session.id)
        now_ms = int(time.time() * 1000)

        if model:
            # 更新现有记录
            model.f_template_id = session.template_id
            model.f_status = (
                session.status.value if hasattr(session.status, "value") else session.status
            )
            model.f_runtime_type = session.runtime_type
            model.f_runtime_node = session.runtime_node or ""
            model.f_container_id = session.container_id or ""
            model.f_pod_name = session.pod_name or ""
            model.f_workspace_path = session.workspace_path
            model.f_resources_cpu = session.resource_limit.cpu
            model.f_resources_memory = session.resource_limit.memory
            model.f_resources_disk = session.resource_limit.disk
            model.f_env_vars = (
                json.dumps(session.env_vars, ensure_ascii=False) if session.env_vars else ""
            )
            model.f_timeout = session.timeout
            model.f_python_package_index_url = session.python_package_index_url
            model.f_last_activity_at = (
                int(session.last_activity_at.timestamp() * 1000)
                if session.last_activity_at
                else now_ms
            )
            model.f_updated_at = now_ms
            model.f_completed_at = (
                int(session.completed_at.timestamp() * 1000) if session.completed_at else 0
            )

            # 依赖安装字段
            model.f_requested_dependencies = (
                json.dumps(session.requested_dependencies, ensure_ascii=False)
                if session.requested_dependencies
                else ""
            )
            if session.installed_dependencies:
                deps_list = [
                    {
                        "name": dep.name,
                        "version": dep.version,
                        "install_location": dep.install_location,
                        "install_time": dep.install_time.isoformat(),
                        "is_from_template": dep.is_from_template,
                    }
                    for dep in session.installed_dependencies
                ]
                model.f_installed_dependencies = json.dumps(deps_list, ensure_ascii=False)
            model.f_dependency_install_status = session.dependency_install_status
            model.f_dependency_install_error = session.dependency_install_error or ""
            model.f_dependency_install_started_at = (
                int(session.dependency_install_started_at.timestamp() * 1000)
                if session.dependency_install_started_at
                else 0
            )
            model.f_dependency_install_completed_at = (
                int(session.dependency_install_completed_at.timestamp() * 1000)
                if session.dependency_install_completed_at
                else 0
            )
        else:
            # 创建新记录
            model = SessionModel.from_entity(session)
            self._session.add(model)

        await self._session.flush()

    async def find_by_id(self, session_id: str) -> Optional[Session]:
        """根据 ID 查找会话"""
        model = await self._session.get(SessionModel, session_id)
        return model.to_entity() if model else None

    async def find_by_container_id(self, container_id: str) -> Optional[Session]:
        """根据容器 ID 查找会话"""
        stmt = select(SessionModel).where(SessionModel.f_container_id == container_id)
        result = await self._session.execute(stmt)
        model = result.scalar_one_or_none()
        return model.to_entity() if model else None

    async def find_by_status(self, status: str, limit: int = 100) -> List[Session]:
        """根据状态查找会话"""
        stmt = select(SessionModel).where(SessionModel.f_status == status).limit(limit)
        result = await self._session.execute(stmt)
        return [model.to_entity() for model in result.scalars().all()]

    async def find_by_template(self, template_id: str) -> List[Session]:
        """根据模板 ID 查找会话"""
        stmt = select(SessionModel).where(SessionModel.f_template_id == template_id)
        result = await self._session.execute(stmt)
        return [model.to_entity() for model in result.scalars().all()]

    async def find_idle_sessions(self, idle_threshold: datetime) -> List[Session]:
        """查找空闲会话"""
        threshold_ms = int(idle_threshold.timestamp() * 1000)
        stmt = select(SessionModel).where(
            SessionModel.f_status.in_(["creating", "running"]),
            SessionModel.f_last_activity_at < threshold_ms,
        )
        result = await self._session.execute(stmt)
        return [model.to_entity() for model in result.scalars().all()]

    async def find_expired_sessions(self, created_before: datetime) -> List[Session]:
        """查找过期会话"""
        before_ms = int(created_before.timestamp() * 1000)
        stmt = (
            select(SessionModel)
            .where(SessionModel.f_created_at < before_ms)
            .where(SessionModel.f_status.in_(["creating", "running"]))
        )
        result = await self._session.execute(stmt)
        return [model.to_entity() for model in result.scalars().all()]

    async def delete(self, session_id: str) -> None:
        """
        删除会话及其所有执行记录（级联删除）

        先删除关联的 execution 记录，再删除 session 记录。
        """
        # 1. 先删除关联的 execution 记录（级联删除）
        if self._execution_repo:
            await self._execution_repo.delete_by_session_id(session_id)
        # 2. 再删除 session 记录
        stmt = delete(SessionModel).where(SessionModel.f_id == session_id)
        await self._session.execute(stmt)
        await self._session.flush()

    async def exists(self, session_id: str) -> bool:
        """检查会话是否存在"""
        model = await self._session.get(SessionModel, session_id)
        return model is not None

    async def count_by_status(self, status: str) -> int:
        """统计指定状态的会话数量"""
        stmt = select(func.count()).select_from(SessionModel).where(SessionModel.f_status == status)
        result = await self._session.execute(stmt)
        return result.scalar() or 0

    async def count_by_node(self, runtime_node: str) -> int:
        """统计指定节点的会话数量"""
        stmt = (
            select(func.count())
            .select_from(SessionModel)
            .where(SessionModel.f_runtime_node == runtime_node)
            .where(SessionModel.f_status.in_(["creating", "running"]))
        )
        result = await self._session.execute(stmt)
        return result.scalar() or 0

    async def find_sessions(
        self,
        status: Optional[str] = None,
        template_id: Optional[str] = None,
        limit: int = 50,
        offset: int = 0,
    ) -> List[Session]:
        """
        查找会话列表（支持筛选和分页）

        Args:
            status: 会话状态筛选（可选）
            template_id: 模板 ID 筛选（可选）
            limit: 返回数量限制（1-200，默认 50）
            offset: 偏移量（用于分页）

        Returns:
            会话列表
        """
        # 验证 limit 范围
        limit = max(1, min(limit, 200))
        offset = max(0, offset)

        # 构建查询
        stmt = select(SessionModel)

        # 添加筛选条件
        if status:
            stmt = stmt.where(SessionModel.f_status == status)
        if template_id:
            stmt = stmt.where(SessionModel.f_template_id == template_id)

        # 排序和分页
        stmt = stmt.order_by(SessionModel.f_created_at.desc()).limit(limit).offset(offset)

        result = await self._session.execute(stmt)
        return [model.to_entity() for model in result.scalars().all()]

    async def count_sessions(
        self, status: Optional[str] = None, template_id: Optional[str] = None
    ) -> int:
        """
        统计会话数量（支持筛选）

        Args:
            status: 会话状态筛选（可选）
            template_id: 模板 ID 筛选（可选）

        Returns:
            会话总数
        """
        stmt = select(func.count()).select_from(SessionModel)

        # 添加筛选条件
        if status:
            stmt = stmt.where(SessionModel.f_status == status)
        if template_id:
            stmt = stmt.where(SessionModel.f_template_id == template_id)

        result = await self._session.execute(stmt)
        return result.scalar() or 0
