"""
运行时节点仓储实现

使用 SQLAlchemy 实现运行时节点仓储接口。
按照数据表命名规范使用 f_ 前缀字段名。
"""
import time
from typing import List, Optional
from decimal import Decimal
from sqlalchemy import select, update
from sqlalchemy.ext.asyncio import AsyncSession

from src.domain.repositories.runtime_node_repository import IRuntimeNodeRepository
from src.infrastructure.persistence.models.runtime_node_model import RuntimeNodeModel


class SqlRuntimeNodeRepository(IRuntimeNodeRepository):
    """
    运行时节点仓储实现

    这是基础设施层的 Adapter，实现领域层定义的 Port。
    """

    def __init__(self, session: AsyncSession):
        self._session = session

    async def save(self, node) -> None:
        """保存节点（创建或更新）"""
        import json
        model = await self._session.get(RuntimeNodeModel, node.node_id)
        now_ms = int(time.time() * 1000)

        if model:
            # 更新现有记录
            model.f_hostname = node.hostname
            model.f_runtime_type = node.type
            model.f_ip_address = node.ip_address
            model.f_api_endpoint = node.url
            model.f_status = "online"
            model.f_total_cpu_cores = Decimal(str(node.total_cpu_cores))
            model.f_total_memory_mb = node.total_memory_mb
            model.f_max_containers = node.max_sessions
            model.f_cached_images = json.dumps(node.cached_templates, ensure_ascii=False) if node.cached_templates else "[]"
            model.f_updated_at = now_ms
        else:
            # 创建新记录
            model = RuntimeNodeModel(
                f_node_id=node.node_id,
                f_hostname=node.hostname,
                f_runtime_type=node.type,
                f_ip_address=node.ip_address,
                f_api_endpoint=node.url,
                f_status="online",
                f_total_cpu_cores=Decimal(str(node.total_cpu_cores)),
                f_total_memory_mb=node.total_memory_mb,
                f_max_containers=node.max_sessions,
                f_cached_images=json.dumps(node.cached_templates, ensure_ascii=False) if node.cached_templates else "[]",
                f_labels="{}",
                f_running_containers=0,
                f_allocated_cpu_cores=Decimal("0"),
                f_allocated_memory_mb=0,
                f_last_heartbeat_at=now_ms,
                f_created_at=now_ms,
                f_created_by="system",
                f_updated_at=now_ms,
                f_updated_by="system",
                f_deleted_at=0,
                f_deleted_by="",
            )
            self._session.add(model)

        await self._session.flush()

    async def find_by_id(self, node_id: str) -> Optional:
        """根据 ID 查找节点"""
        model = await self._session.get(RuntimeNodeModel, node_id)
        return model if model else None

    async def find_by_hostname(self, hostname: str) -> Optional:
        """根据主机名查找节点"""
        stmt = select(RuntimeNodeModel).where(RuntimeNodeModel.f_hostname == hostname)
        result = await self._session.execute(stmt)
        return result.scalar_one_or_none()

    async def find_by_status(self, status: str) -> List:
        """根据状态查找节点"""
        stmt = select(RuntimeNodeModel).where(RuntimeNodeModel.f_status == status)
        result = await self._session.execute(stmt)
        return list(result.scalars().all())

    async def find_all(
        self,
        offset: int = 0,
        limit: int = 100
    ) -> List:
        """查找所有节点"""
        stmt = (
            select(RuntimeNodeModel)
            .offset(offset)
            .limit(limit)
            .order_by(RuntimeNodeModel.f_hostname)
        )
        result = await self._session.execute(stmt)
        return list(result.scalars().all())

    async def update_status(
        self,
        node_id: str,
        status: str
    ) -> None:
        """更新节点状态"""
        stmt = (
            update(RuntimeNodeModel)
            .where(RuntimeNodeModel.f_node_id == node_id)
            .values(f_status=status, f_updated_at=int(time.time() * 1000))
        )
        await self._session.execute(stmt)
        await self._session.flush()

    async def update_heartbeat(self, node_id: str) -> None:
        """更新节点心跳时间"""
        stmt = (
            update(RuntimeNodeModel)
            .where(RuntimeNodeModel.f_node_id == node_id)
            .values(f_last_heartbeat_at=int(time.time() * 1000))
        )
        await self._session.execute(stmt)
        await self._session.flush()

    async def allocate_resources(
        self,
        node_id: str,
        cpu_cores: float,
        memory_mb: int
    ) -> None:
        """分配资源"""
        stmt = (
            update(RuntimeNodeModel)
            .where(RuntimeNodeModel.f_node_id == node_id)
            .values(
                f_allocated_cpu_cores=RuntimeNodeModel.f_allocated_cpu_cores + Decimal(str(cpu_cores)),
                f_allocated_memory_mb=RuntimeNodeModel.f_allocated_memory_mb + memory_mb,
                f_updated_at=int(time.time() * 1000),
            )
        )
        await self._session.execute(stmt)
        await self._session.flush()

    async def release_resources(
        self,
        node_id: str,
        cpu_cores: float,
        memory_mb: int
    ) -> None:
        """释放资源"""
        stmt = (
            update(RuntimeNodeModel)
            .where(RuntimeNodeModel.f_node_id == node_id)
            .values(
                f_allocated_cpu_cores=RuntimeNodeModel.f_allocated_cpu_cores - Decimal(str(cpu_cores)),
                f_allocated_memory_mb=RuntimeNodeModel.f_allocated_memory_mb - memory_mb,
                f_updated_at=int(time.time() * 1000),
            )
        )
        await self._session.execute(stmt)
        await self._session.flush()

    async def increment_container_count(self, node_id: str) -> None:
        """增加容器计数"""
        stmt = (
            update(RuntimeNodeModel)
            .where(RuntimeNodeModel.f_node_id == node_id)
            .values(
                f_running_containers=RuntimeNodeModel.f_running_containers + 1,
                f_updated_at=int(time.time() * 1000),
            )
        )
        await self._session.execute(stmt)
        await self._session.flush()

    async def decrement_container_count(self, node_id: str) -> None:
        """减少容器计数"""
        stmt = (
            update(RuntimeNodeModel)
            .where(RuntimeNodeModel.f_node_id == node_id)
            .values(
                f_running_containers=RuntimeNodeModel.f_running_containers - 1,
                f_updated_at=int(time.time() * 1000),
            )
        )
        await self._session.execute(stmt)
        await self._session.flush()
