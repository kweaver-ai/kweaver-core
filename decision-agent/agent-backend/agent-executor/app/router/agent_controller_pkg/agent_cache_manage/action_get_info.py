# -*- coding:utf-8 -*-
"""
Agent缓存管理 - get_info action处理

逻辑：
- 如果agent_id对应的缓存存在(redis中)，则返回缓存信息
- 如果agent_id对应的缓存不存在，则返回null
"""

from typing import Optional
from fastapi import Request

from app.domain.vo.agentvo import AgentConfigVo
from app.utils.observability.observability_log import get_logger as o11y_logger

from ..rdto.v1.res.agent_cache import AgentCacheManageRes
from .common import (
    cache_manager,
    build_cache_id_vo,
    get_cache_data_for_debug_mode,
)


async def handle_get_info(
    request: Request,
    account_id: str,
    account_type: str,
    agent_id: str,
    agent_version: str,
    agent_config: AgentConfigVo,
) -> Optional[AgentCacheManageRes]:
    """处理 get_info action

    Args:
        request: FastAPI请求对象
        account_id: 账户ID
        account_type: 账户类型
        agent_id: Agent ID
        agent_version: Agent版本
        agent_config: Agent配置

    Returns:
        Optional[AgentCacheManageRes]: 响应对象，缓存不存在时返回None
    """
    # 1. 构造cache_id
    cache_id_vo = build_cache_id_vo(
        account_id, account_type, agent_id, agent_version, agent_config
    )

    # 2. 检查Cache是否存在
    cache_entity = await cache_manager.cache_service.load(cache_id_vo)

    if cache_entity is None:
        # Cache不存在，返回None
        o11y_logger().info(f"Agent缓存不存在: cache_id={cache_id_vo}")
        return None

    # 3. Cache存在，返回缓存信息
    o11y_logger().info(f"Agent缓存存在，返回缓存信息: cache_id={cache_id_vo}")

    # 获取实时TTL
    ttl = await cache_manager.cache_service.get_ttl(cache_id_vo)

    return AgentCacheManageRes(
        cache_id=cache_id_vo.get_cache_id(),
        ttl=ttl,
        created_at=cache_entity.created_at,
        cache_data=get_cache_data_for_debug_mode(cache_entity),
    )
