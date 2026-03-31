# -*- coding:utf-8 -*-
"""Agent缓存管理器主类

负责协调各个子模块，提供统一的Agent缓存管理接口
"""

from typing import Dict

from app.domain.vo.agent_cache import AgentCacheIdVO
from app.domain.entity.agent_cache import AgentCacheEntity
from app.domain.vo.agentvo import AgentConfigVo
from .agent_cache_service import AgentCacheService


class AgentCacheManager:
    """Agent缓存管理器 - 主类

    负责协调各个子模块，提供统一的Agent缓存管理接口
    """

    def __init__(self):
        """初始化AgentCacheManager"""
        # 初始化基础服务
        self.cache_service = AgentCacheService()

    async def create_cache(
        self,
        account_id: str,
        account_type: str,
        agent_id: str,
        agent_version: str,
        agent_config: AgentConfigVo,
        headers: Dict[str, str],
    ) -> AgentCacheEntity:
        """创建新的Agent缓存

        Args:
            account_id: 账户ID
            account_type: 账户类型
            agent_id: Agent ID
            agent_version: Agent版本
            agent_config: Agent配置
            headers: HTTP请求头

        Returns:
            AgentCacheEntity: 创建的缓存实体
        """
        from .create_cache import create_cache

        return await create_cache(
            manager=self,
            account_id=account_id,
            account_type=account_type,
            agent_id=agent_id,
            agent_version=agent_version,
            agent_config=agent_config,
            headers=headers,
        )

    async def update_cache_data(
        self,
        cache_id_vo: AgentCacheIdVO,
        agent_config: AgentConfigVo,
        headers: Dict[str, str],
    ) -> None:
        """更新Agent缓存数据
        保持缓存的ttl不变

        Args:
            cache_id_vo: 缓存ID
            agent_config: Agent配置
            headers: HTTP请求头
        """
        from .update_cache_data import update_cache_data

        await update_cache_data(
            manager=self,
            cache_id_vo=cache_id_vo,
            agent_config=agent_config,
            headers=headers,
        )
