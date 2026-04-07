from typing import Dict, Optional

from app.logic.agent_core_logic_v2.agent_core_v2 import AgentCoreV2
from app.logic.agent_core_logic_v2.agent_cache_manage_logic import AgentCacheManager
from app.domain.vo.agentvo import AgentConfigVo
from app.domain.vo.agent_cache import AgentCacheIdVO
from app.domain.constant import AGENT_CACHE_TTL
from app.common.stand_log import StandLogger

# 全局AgentCacheManager实例
cache_manager = AgentCacheManager()


async def handle_cache(
    agent_id: str,
    agent_core_v2: AgentCoreV2,
    is_debug_run: bool,
    headers: Dict[str, str] = None,
    account_id: str = "",
    account_type: str = "",
) -> Optional[AgentCacheIdVO]:
    """
    处理缓存初始化和状态管理

    Args:
        agent_id: Agent ID
        agent_core_v2: Agent核心实例
        is_debug_run: 是否为调试模式
        headers: HTTP请求头（创建缓存时需要）
        account_id: 账户ID
        account_type: 账户类型

    Returns:
        AgentCacheIdVO: 缓存ID值对象，用于后续更新缓存时使用
    """
    agent_config: AgentConfigVo = agent_core_v2.agent_config

    # 1. 如果是调试模式，不处理缓存
    if is_debug_run:
        StandLogger.debug("[handle_cache] 调试模式，跳过缓存处理")
        return None

    # 2. 构造cache_id
    cache_id_vo = AgentCacheIdVO(
        account_id=account_id,
        account_type=account_type,
        agent_id=agent_id,
        agent_version=agent_config.agent_version or "",
        agent_config_version_flag=str(agent_config.get_config_last_set_timestamp()),
    )

    StandLogger.debug(f"[handle_cache] 开始处理缓存, cache_id={cache_id_vo}")

    # 3. 尝试从缓存加载
    cache_entity = await cache_manager.cache_service.load(cache_id_vo)

    # 4. 如果缓存存在，加载缓存数据到cache_handler
    if cache_entity:
        # 获取当前缓存的TTL
        current_ttl = await cache_manager.cache_service.get_ttl(cache_id_vo)
        ttl_passed = AGENT_CACHE_TTL - current_ttl
        StandLogger.debug(
            f"[handle_cache] ✅ 使用已有缓存: cache_id={cache_id_vo}, 剩余TTL={current_ttl}秒, 已过={ttl_passed}秒"
        )
        agent_core_v2.cache_handler.set_cache_data(cache_entity.cache_data)
    else:
        # 5. 如果缓存不存在，创建新缓存
        StandLogger.debug(
            f"[handle_cache] 缓存不存在，开始创建新缓存: cache_id={cache_id_vo}"
        )
        try:
            cache_entity = await cache_manager.create_cache(
                account_id=account_id,
                account_type=account_type,
                agent_id=agent_id,
                agent_version=agent_config.agent_version or "",
                agent_config=agent_config,
                headers=headers or {},
            )
            # 加载刚创建的缓存数据到cache_handler
            agent_core_v2.cache_handler.set_cache_data(cache_entity.cache_data)
            # 获取新创建缓存的TTL
            new_ttl = await cache_manager.cache_service.get_ttl(cache_id_vo)
            StandLogger.debug(
                f"[handle_cache] ✅ 成功创建新缓存: cache_id={cache_id_vo}, TTL={new_ttl}秒"
            )
        except Exception as e:
            StandLogger.error(
                f"[handle_cache] ❌ 创建缓存失败: cache_id={cache_id_vo}, error={e}"
            )
            return None

    return cache_id_vo
