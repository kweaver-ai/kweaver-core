# -*- coding:utf-8 -*-
"""
Agent缓存管理 - upsert action处理

逻辑：
- 如果agent_id对应的缓存存在(redis中)，则更新并返回缓存信息
  - 获取新的缓存数据，并更新redis中的缓存（恢复有效期ttl）
  - 返回更新后的缓存信息
- 如果agent_id对应的缓存不存在，则创建缓存并返回缓存信息
  - 获取新的缓存数据，并创建redis中的缓存（设置有效期ttl）
  - 返回创建后的缓存信息
"""

from fastapi import Request

from app.domain.vo.agentvo import AgentConfigVo
from app.utils.observability.observability_log import get_logger as o11y_logger

from ..rdto.v1.res.agent_cache import AgentCacheManageRes
from .common import (
    cache_manager,
    build_cache_id_vo,
    get_cache_data_for_debug_mode,
    create_cache_and_build_response,
)


async def handle_upsert(
    request: Request,
    account_id: str,
    account_type: str,
    agent_id: str,
    agent_version: str,
    agent_config: AgentConfigVo,
) -> AgentCacheManageRes:
    """处理 upsert action

    Args:
        request: FastAPI请求对象
        account_id: 账户ID
        account_type: 账户类型
        agent_id: Agent ID
        agent_version: Agent版本
        agent_config: Agent配置

    Returns:
        AgentCacheManageRes: 响应对象
    """
    # 1. 构造cache_id
    cache_id_vo = build_cache_id_vo(
        account_id, account_type, agent_id, agent_version, agent_config
    )

    # 2. 检查Cache是否已存在
    cache_entity = await cache_manager.cache_service.load(cache_id_vo)

    if cache_entity:
        # Cache存在，更新缓存数据并恢复TTL
        o11y_logger().info(f"Agent缓存已存在，更新缓存数据: cache_id={cache_id_vo}")

        # 更新缓存数据
        await cache_manager.update_cache_data(
            cache_id_vo=cache_id_vo,
            agent_config=agent_config,
            headers=dict(request.headers),
        )

        # 获取更新后的缓存
        cache_entity = await cache_manager.cache_service.load(cache_id_vo)
        if cache_entity is None:
            # 缓存在更新过程中过期，创建新缓存
            o11y_logger().warn(f"Agent缓存在更新过程中过期: cache_id={cache_id_vo}")
            return await create_cache_and_build_response(
                request=request,
                account_id=account_id,
                account_type=account_type,
                agent_id=agent_id,
                agent_version=agent_version,
                agent_config=agent_config,
            )

        # 获取实时TTL
        ttl = await cache_manager.cache_service.get_ttl(cache_id_vo)

        return AgentCacheManageRes(
            cache_id=cache_id_vo.get_cache_id(),
            ttl=ttl,
            created_at=cache_entity.created_at,
            cache_data=get_cache_data_for_debug_mode(cache_entity),
        )

    # 3. Cache不存在，创建新Cache并返回响应
    o11y_logger().info(f"Agent缓存不存在，创建新Cache: agent_id={agent_id}")

    return await create_cache_and_build_response(
        request=request,
        account_id=account_id,
        account_type=account_type,
        agent_id=agent_id,
        agent_version=agent_version,
        agent_config=agent_config,
    )
