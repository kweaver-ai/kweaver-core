"""
运行时节点仓储接口

定义运行时节点持久化的抽象接口（Port）。
"""
from abc import ABC, abstractmethod
from typing import List, Optional


class IRuntimeNodeRepository(ABC):
    """
    运行时节点仓储接口

    这是领域层定义的 Port，由基础设施层实现 Adapter。
    """

    @abstractmethod
    async def save(self, node) -> None:
        """保存节点（创建或更新）"""
        pass

    @abstractmethod
    async def find_by_id(self, node_id: str) -> Optional:
        """根据 ID 查找节点"""
        pass

    @abstractmethod
    async def find_by_hostname(self, hostname: str) -> Optional:
        """根据主机名查找节点"""
        pass

    @abstractmethod
    async def find_by_status(self, status: str) -> List:
        """根据状态查找节点"""
        pass

    @abstractmethod
    async def find_all(
        self,
        offset: int = 0,
        limit: int = 100
    ) -> List:
        """查找所有节点"""
        pass

    @abstractmethod
    async def update_status(
        self,
        node_id: str,
        status: str
    ) -> None:
        """更新节点状态"""
        pass

    @abstractmethod
    async def update_heartbeat(self, node_id: str) -> None:
        """更新节点心跳时间"""
        pass

    @abstractmethod
    async def allocate_resources(
        self,
        node_id: str,
        cpu_cores: float,
        memory_mb: int
    ) -> None:
        """分配资源"""
        pass

    @abstractmethod
    async def release_resources(
        self,
        node_id: str,
        cpu_cores: float,
        memory_mb: int
    ) -> None:
        """释放资源"""
        pass

    @abstractmethod
    async def increment_container_count(self, node_id: str) -> None:
        """增加容器计数"""
        pass

    @abstractmethod
    async def decrement_container_count(self, node_id: str) -> None:
        """减少容器计数"""
        pass
