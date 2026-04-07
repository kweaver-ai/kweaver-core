# -*- coding:utf-8 -*-
"""Agent缓存服务

负责Agent缓存数据的缓存操作，使用PICKLE序列化存储
"""

import logging
from typing import Optional

from app.domain.constant import AGENT_CACHE_TTL
from app.domain.entity.agent_cache import AgentCacheEntity
from app.domain.vo.agent_cache import AgentCacheIdVO
from app.infra.common.util.redis_cache import RedisCache
from app.infra.common.util.redis_cache.redis_cache import SerializationType

logger = logging.getLogger(__name__)


class AgentCacheService:
    """Agent缓存服务

    负责：
    1. Agent缓存数据的序列化和反序列化
    2. Redis缓存操作（保存、加载、更新TTL、删除）

    存储策略：
    - 存储在 {cache_id} key中，包含创建DolphinAgent所需的所有静态数据
    """

    def __init__(self, redis_db: int = 3):
        """初始化缓存服务

        Args:
            redis_db: Redis数据库编号，默认为3
        """
        self.redis_cache = RedisCache(db=redis_db)
        self.redis_db = redis_db

    async def save(
        self, cache_entity: AgentCacheEntity, ttl: int = AGENT_CACHE_TTL
    ) -> bool:
        """保存Agent缓存

        Args:
            cache_entity: Agent缓存实体
            ttl: 缓存有效期（秒）

        Returns:
            是否成功保存
        """
        try:
            key = cache_entity.cache_id_vo.to_redis_key()

            success = await self.redis_cache.set(
                key,
                cache_entity,
                ttl=ttl,
                serialization_type=SerializationType.PICKLE,
            )

            if not success:
                logger.error(
                    f"保存Agent缓存数据失败: cache_id={cache_entity.cache_id_vo}"
                )
                return False

            logger.info(
                f"成功保存Agent缓存: cache_id={cache_entity.cache_id_vo}, ttl={ttl}"
            )
            return True

        except Exception as e:
            logger.error(
                f"保存Agent缓存失败: cache_id={cache_entity.cache_id_vo}, error={e}",
                exc_info=True,
            )
            return False

    async def load(self, cache_id_vo: AgentCacheIdVO) -> Optional[AgentCacheEntity]:
        """从缓存加载Agent缓存

        Args:
            cache_id_vo: Agent缓存ID

        Returns:
            AgentCacheEntity或None（不存在）
        """
        try:
            key = cache_id_vo.to_redis_key()

            data = await self.redis_cache.get(key)

            if not data:
                logger.debug(f"Agent缓存不存在: cache_id={cache_id_vo}")
                return None

            logger.info(f"成功加载Agent缓存: cache_id={cache_id_vo}")

            return data

        except Exception as e:
            logger.error(
                f"加载Agent缓存失败: cache_id={cache_id_vo}, error={e}", exc_info=True
            )
            return None

    async def update_ttl(self, cache_id_vo: AgentCacheIdVO, ttl: int) -> bool:
        """更新TTL

        Args:
            cache_id_vo: Agent缓存ID
            ttl: 新的TTL（秒）

        Returns:
            是否成功更新
        """
        try:
            key = cache_id_vo.to_redis_key()
            success = await self.redis_cache.expire(key, ttl)

            if not success:
                logger.warning(
                    f"更新TTL失败(key可能已过期): cache_id={cache_id_vo}, ttl={ttl}"
                )
                return False

            logger.debug(f"成功更新TTL: cache_id={cache_id_vo}, ttl={ttl}")
            return True

        except Exception as e:
            logger.error(
                f"更新TTL失败: cache_id={cache_id_vo}, error={e}", exc_info=True
            )
            return False

    async def exists(self, cache_id_vo: AgentCacheIdVO) -> bool:
        """检查Agent缓存是否存在

        Args:
            cache_id_vo: Agent缓存ID

        Returns:
            是否存在
        """
        try:
            key = cache_id_vo.to_redis_key()
            return await self.redis_cache.exists(key)
        except Exception as e:
            logger.error(
                f"检查Agent缓存存在失败: cache_id={cache_id_vo}, error={e}",
                exc_info=True,
            )
            return False

    async def get_ttl(self, cache_id_vo: AgentCacheIdVO) -> int:
        """获取剩余TTL

        Args:
            cache_id_vo: Agent缓存ID

        Returns:
            剩余秒数，-1表示永不过期，-2表示不存在
        """
        try:
            key = cache_id_vo.to_redis_key()
            ttl = await self.redis_cache.ttl(key)
            logger.debug(f"获取TTL: cache_id={cache_id_vo}, ttl={ttl}")
            return ttl
        except Exception as e:
            logger.error(
                f"获取TTL失败: cache_id={cache_id_vo}, error={e}", exc_info=True
            )
            return -2

    async def delete(self, cache_id_vo: AgentCacheIdVO) -> bool:
        """删除Agent缓存

        Args:
            cache_id_vo: Agent缓存ID

        Returns:
            是否成功删除
        """
        try:
            key = cache_id_vo.to_redis_key()
            success = await self.redis_cache.delete(key)

            if success:
                logger.info(f"成功删除Agent缓存: cache_id={cache_id_vo}")
            else:
                logger.warning(
                    f"删除Agent缓存失败(key可能不存在): cache_id={cache_id_vo}"
                )

            return success

        except Exception as e:
            logger.error(
                f"删除Agent缓存失败: cache_id={cache_id_vo}, error={e}", exc_info=True
            )
            return False
