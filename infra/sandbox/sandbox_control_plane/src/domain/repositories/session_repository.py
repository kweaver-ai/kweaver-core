"""
会话仓储接口

定义会话持久化的抽象接口（Port）。
"""
from abc import ABC, abstractmethod
from typing import List, Optional
from datetime import datetime

from src.domain.entities.session import Session


class ISessionRepository(ABC):
    """
    会话仓储接口

    这是领域层定义的 Port，由基础设施层实现 Adapter。
    """

    @abstractmethod
    async def save(self, session: Session) -> None:
        """保存会话（创建或更新）"""
        pass

    @abstractmethod
    async def find_by_id(self, session_id: str) -> Optional[Session]:
        """根据 ID 查找会话"""
        pass

    @abstractmethod
    async def find_by_container_id(self, container_id: str) -> Optional[Session]:
        """根据容器 ID 查找会话"""
        pass

    @abstractmethod
    async def find_by_status(
        self,
        status: str,
        limit: int = 100
    ) -> List[Session]:
        """根据状态查找会话"""
        pass

    @abstractmethod
    async def find_by_template(self, template_id: str) -> List[Session]:
        """根据模板 ID 查找会话"""
        pass

    @abstractmethod
    async def find_idle_sessions(
        self,
        idle_threshold: datetime
    ) -> List[Session]:
        """查找空闲会话（用于自动清理）"""
        pass

    @abstractmethod
    async def find_expired_sessions(
        self,
        created_before: datetime
    ) -> List[Session]:
        """查找过期会话（用于自动清理）"""
        pass

    @abstractmethod
    async def delete(self, session_id: str) -> None:
        """删除会话"""
        pass

    @abstractmethod
    async def exists(self, session_id: str) -> bool:
        """检查会话是否存在"""
        pass

    @abstractmethod
    async def count_by_status(self, status: str) -> int:
        """统计指定状态的会话数量"""
        pass

    @abstractmethod
    async def count_by_node(self, runtime_node: str) -> int:
        """统计指定节点的会话数量"""
        pass

    @abstractmethod
    async def find_sessions(
        self,
        status: Optional[str] = None,
        template_id: Optional[str] = None,
        limit: int = 50,
        offset: int = 0
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
        pass

    @abstractmethod
    async def count_sessions(
        self,
        status: Optional[str] = None,
        template_id: Optional[str] = None
    ) -> int:
        """
        统计会话数量（支持筛选）

        Args:
            status: 会话状态筛选（可选）
            template_id: 模板 ID 筛选（可选）

        Returns:
            会话总数
        """
        pass
