# -*- coding:utf-8 -*-
"""Agent缓存创建逻辑

负责Agent缓存的创建
"""

import time
from typing import Dict, TYPE_CHECKING
from datetime import datetime

from app.domain.entity.agent_cache import AgentCacheEntity
from app.domain.vo.agent_cache import AgentCacheIdVO
from app.domain.vo.agentvo import AgentConfigVo

from app.utils.observability.observability_log import get_logger as o11y_logger


if TYPE_CHECKING:
    from .manager import AgentCacheManager


async def create_cache(
    manager: "AgentCacheManager",
    account_id: str,
    account_type: str,
    agent_id: str,
    agent_version: str,
    agent_config: AgentConfigVo,
    headers: Dict[str, str],
) -> AgentCacheEntity:
    """创建新的Agent缓存

    Args:
        manager: AgentCacheManager实例
        account_id: 账户ID
        account_type: 账户类型
        agent_id: Agent ID
        agent_version: Agent版本
        agent_config: Agent配置
        headers: HTTP请求头

    Returns:
        AgentCacheEntity: 创建的缓存实体
    """

    # 1. 生成CacheID
    agent_config_version_flag = str(agent_config.get_config_last_set_timestamp())
    cache_id_vo = AgentCacheIdVO(
        account_id=account_id,
        account_type=account_type,
        agent_id=agent_id,
        agent_version=agent_version,
        agent_config_version_flag=agent_config_version_flag,
    )

    try:
        # 2. 初始化Agent组件静态数据
        from ..agent_core_v2 import AgentCoreV2

        agent_core_v2 = AgentCoreV2(agent_config, True)
        await agent_core_v2.warmup_handler.warnup(
            headers=headers,
        )

        # 3. 创建Cache实体
        cache_entity = AgentCacheEntity(
            cache_id_vo=cache_id_vo,
            agent_id=agent_id,
            agent_version=agent_version,
            cache_data=agent_core_v2.cache_handler.get_cache_data(),
            cache_data_last_set_timestamp=int(time.time()),
            created_at=datetime.now(),
        )

        # 4. 保存到Redis
        await manager.cache_service.save(cache_entity)

    except Exception as e:
        o11y_logger().error(f"agent cache create failed: {e}")
        raise e

    return cache_entity
