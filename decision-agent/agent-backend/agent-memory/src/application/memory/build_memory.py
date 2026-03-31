from typing import List, Dict, Any, Optional
from src.domain.memory.mem0_adapter import Mem0MemoryAdapter
from src.utils.logger import logger


class BuildMemoryUseCase:
    """记忆构建用例"""

    def __init__(self):
        self.memory_adapter = None

    async def initialize(self):
        """初始化用例"""
        if self.memory_adapter is None:
            logger.info("Initializing memory adapter...")
            self.memory_adapter = await Mem0MemoryAdapter.create()
            logger.info("Memory adapter initialized.")
        return self

    async def execute(
        self,
        messages: List[Dict[str, Any]],
        user_id: Optional[str] = None,
        agent_id: Optional[str] = None,
        run_id: Optional[str] = None,
        metadata: Optional[Dict[str, Any]] = None,
        infer: bool = True,
        memory_type: Optional[str] = None,
        prompt: Optional[str] = None,
        context: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, Any]:
        """
        执行记忆构建

        Args:
            messages: 消息列表
            user_id: 用户ID
            agent_id: 代理ID
            run_id: 运行ID
            metadata: 元数据
            infer: 是否推理
            memory_type: 记忆类型
            prompt: 提示词
            context: 请求上下文，包含x-account-id和x-account-type信息

        Returns:
            Dict[str, Any]: 构建结果
        """
        if self.memory_adapter is None:
            await self.initialize()

        result = await self.memory_adapter.add(
            messages=messages,
            user_id=user_id,
            agent_id=agent_id,
            run_id=run_id,
            metadata=metadata,
            infer=infer,
            memory_type=memory_type,
            prompt=prompt,
            context=context,
        )
        return result
