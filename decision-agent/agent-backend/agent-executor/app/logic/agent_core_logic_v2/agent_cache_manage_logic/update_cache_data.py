# -*- coding:utf-8 -*-
"""Agent缓存数据更新逻辑

负责Agent缓存数据的更新
"""

import time
from typing import Dict, TYPE_CHECKING

from app.domain.vo.agent_cache import AgentCacheIdVO
from app.domain.vo.agentvo import AgentConfigVo

from app.utils.observability.observability_log import get_logger as o11y_logger


if TYPE_CHECKING:
    from .manager import AgentCacheManager


async def update_cache_data(
    manager: "AgentCacheManager",
    cache_id_vo: AgentCacheIdVO,
    agent_config: AgentConfigVo,
    headers: Dict[str, str],
) -> None:
    """更新Agent缓存数据
    保持缓存的ttl不变

    Args:
        manager: AgentCacheManager实例
        cache_id_vo: 缓存ID
        agent_config: Agent配置
        headers: HTTP请求头
    """

    try:
        # 1. load cache
        cache_entity = await manager.cache_service.load(cache_id_vo)
        if not cache_entity:
            # 缓存不存在，创建新缓存
            from .create_cache import create_cache

            await create_cache(
                manager=manager,
                account_id=cache_id_vo.account_id,
                account_type=cache_id_vo.account_type,
                agent_id=cache_id_vo.agent_id,
                agent_version=cache_id_vo.agent_version,
                agent_config=agent_config,
                headers=headers,
            )
            return

        # 2. get cache data
        from ..agent_core_v2 import AgentCoreV2

        agent_core_v2 = AgentCoreV2(agent_config, True)
        await agent_core_v2.warmup_handler.warnup(
            headers=headers,
        )

        # 3. update cache data
        cache_entity.cache_data = agent_core_v2.cache_handler.get_cache_data()
        cache_entity.cache_data_last_set_timestamp = int(time.time())

        # 4. 保存到Redis(ttl恢复为完整TTL)
        await manager.cache_service.save(cache_entity)

    except Exception as e:
        o11y_logger().error(f"agent cache update data failed: {e}")
        raise e
