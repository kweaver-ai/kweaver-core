from typing import List, Dict, Any, Optional
from src.domain.memory.mem0_adapter import Mem0MemoryAdapter


class RetrievalMemoryUseCase:
    """记忆召回用例"""

    def __init__(self):
        self.memory_adapter = None

    async def initialize(self):
        """初始化用例"""
        if self.memory_adapter is None:
            self.memory_adapter = await Mem0MemoryAdapter.create()
        return self

    async def execute(
        self,
        query: str,
        user_id: Optional[str] = None,
        agent_id: Optional[str] = None,
        run_id: Optional[str] = None,
        limit: int = 5,
        filters: Optional[Dict[str, Any]] = None,
        threshold: Optional[float] = None,
        rerank_threshold: Optional[float] = None,
        context: Optional[Dict[str, Any]] = None,
    ) -> List[Dict[str, Any]]:
        """
        执行记忆召回

        Args:
            query: 查询文本
            user_id: 用户ID
            agent_id: 代理ID
            run_id: 运行ID
            limit: 返回结果数量限制
            filters: 过滤条件
            threshold: 相似度阈值
            rerank_threshold: Rerank 阈值
            context: 请求上下文，包含x-account-id和x-account-type信息
        Returns:
            List[Dict[str, Any]]: 召回的记忆列表
        """
        if self.memory_adapter is None:
            await self.initialize()

        results = await self.memory_adapter.search(
            query=query,
            user_id=user_id,
            agent_id=agent_id,
            run_id=run_id,
            limit=limit,
            filters=filters,
            threshold=threshold,
            rerank_threshold=rerank_threshold,
            context=context,
        )
        return results
