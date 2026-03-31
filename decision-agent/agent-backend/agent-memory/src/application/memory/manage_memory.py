from typing import List, Dict, Any, Optional
from src.domain.memory.mem0_adapter import Mem0MemoryAdapter


class ManageMemoryUseCase:
    """记忆管理用例"""

    def __init__(self):
        self.memory_adapter = None

    async def initialize(self):
        """初始化用例"""
        if self.memory_adapter is None:
            self.memory_adapter = await Mem0MemoryAdapter.create()
        return self

    async def get_memory(self, memory_id: str) -> Dict[str, Any]:
        """获取单个记忆"""
        if self.memory_adapter is None:
            await self.initialize()
        return await self.memory_adapter.get(memory_id)

    async def get_all_memories(
        self,
        user_id: Optional[str] = None,
        agent_id: Optional[str] = None,
        run_id: Optional[str] = None,
        filters: Optional[Dict[str, Any]] = None,
        limit: int = 100,
    ) -> List[Dict[str, Any]]:
        """获取所有记忆"""
        if self.memory_adapter is None:
            await self.initialize()
        return await self.memory_adapter.get_all(
            user_id=user_id,
            agent_id=agent_id,
            run_id=run_id,
            filters=filters,
            limit=limit,
        )

    async def update_memory(self, memory_id: str, data: str) -> Dict[str, Any]:
        """更新记忆"""
        if self.memory_adapter is None:
            await self.initialize()
        return await self.memory_adapter.update(memory_id, data)

    async def delete_memory(self, memory_id: str) -> bool:
        """删除记忆"""
        if self.memory_adapter is None:
            await self.initialize()
        return await self.memory_adapter.delete(memory_id)

    async def get_memory_history(self, memory_id: str) -> List[Dict[str, Any]]:
        """获取记忆历史"""
        if self.memory_adapter is None:
            await self.initialize()
        return await self.memory_adapter.history(memory_id)
