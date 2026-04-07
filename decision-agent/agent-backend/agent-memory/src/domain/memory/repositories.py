from abc import ABC, abstractmethod
from typing import List, Optional
from .entities import Memory, MemoryChunk


class MemoryRepository(ABC):
    """记忆仓储接口"""

    @abstractmethod
    async def save(self, memory: Memory) -> Memory:
        """保存记忆"""
        pass

    @abstractmethod
    async def get_by_id(self, memory_id: str) -> Optional[Memory]:
        """根据ID获取记忆"""
        pass

    @abstractmethod
    async def delete(self, memory_id: str) -> bool:
        """删除记忆"""
        pass

    @abstractmethod
    async def update(self, memory: Memory) -> Memory:
        """更新记忆"""
        pass

    @abstractmethod
    async def list_by_tags(self, tags: List[str]) -> List[Memory]:
        """根据标签列表获取记忆"""
        pass


class MemoryChunkRepository(ABC):
    """记忆分块仓储接口"""

    @abstractmethod
    async def save(self, chunk: MemoryChunk) -> MemoryChunk:
        """保存记忆分块"""
        pass

    @abstractmethod
    async def get_by_memory_id(self, memory_id: str) -> List[MemoryChunk]:
        """获取记忆的所有分块"""
        pass

    @abstractmethod
    async def delete_by_memory_id(self, memory_id: str) -> bool:
        """删除记忆的所有分块"""
        pass
