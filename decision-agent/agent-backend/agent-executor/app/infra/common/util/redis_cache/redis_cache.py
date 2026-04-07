"""Redis缓存操作类

优先使用JSON序列化，失败时降级使用Pickle
"""

import logging
from enum import Enum
from typing import Any, List, Optional


from app.driven.infrastructure.redis import redis_pool

from .json_serializer import JSONSerializer
from .pickle_serializer import PickleSerializer

logger = logging.getLogger(__name__)


class SerializationType(Enum):
    """序列化类型枚举"""

    JSON = "json"
    PICKLE = "pickle"


class RedisCache:
    """Redis缓存操作类

    优先使用JSON序列化，失败时降级使用Pickle。
    所有操作都是异步的，使用async/await。

    特性：
    - 自动序列化/反序列化
    - JSON优先，Pickle备选
    - 支持TTL管理
    - 支持分布式锁操作
    - 支持Lua脚本执行
    """

    # 序列化类型标记
    _SERIALIZATION_TYPE_JSON = b"JSON:"
    _SERIALIZATION_TYPE_PICKLE = b"PICKLE:"

    def __init__(self, db: int = 3):
        """初始化Redis缓存

        Args:
            db: Redis数据库编号，默认为3
        """
        self.db = db
        self.json_serializer = JSONSerializer()
        self.pickle_serializer = PickleSerializer()

    async def set(
        self,
        key: str,
        value: Any,
        ttl: Optional[int] = None,
        serialization_type: SerializationType = SerializationType.JSON,
    ) -> bool:
        """设置缓存

        根据指定的序列化类型进行序列化。如果指定JSON但序列化失败，则自动降级使用Pickle。

        Args:
            key: 缓存键
            value: 缓存值
            ttl: 过期时间（秒），None表示永不过期
            serialization_type: 序列化类型，默认为JSON

        Returns:
            是否成功

        Example:
            >>> cache = RedisCache()
            >>> await cache.set("user:123", {"name": "Alice"}, ttl=3600)
            True
            >>> await cache.set("obj:456", custom_obj, serialization_type=SerializationType.PICKLE)
            True
        """
        try:
            # 根据指定的序列化类型进行序列化
            if serialization_type == SerializationType.JSON:
                try:
                    serialized = self.json_serializer.serialize(value)
                    data = self._SERIALIZATION_TYPE_JSON + serialized.encode("utf-8")
                    logger.debug(f"使用JSON序列化存储键: {key}")
                except (TypeError, ValueError) as e:
                    # JSON序列化失败，降级使用Pickle
                    logger.debug(f"JSON序列化失败，降级使用Pickle: {e}")
                    serialized = self.pickle_serializer.serialize(value)
                    data = self._SERIALIZATION_TYPE_PICKLE + serialized
            else:  # SerializationType.PICKLE
                serialized = self.pickle_serializer.serialize(value)
                data = self._SERIALIZATION_TYPE_PICKLE + serialized
                logger.debug(f"使用Pickle序列化存储键: {key}")

            # 存储到Redis
            async with redis_pool.acquire(self.db, "write") as conn:
                if ttl is not None:
                    await conn.set(key, data, ex=ttl)
                else:
                    await conn.set(key, data)
                return True

        except Exception as e:
            logger.error(f"设置缓存失败 key={key}: {e}")
            return False

    async def get(self, key: str) -> Optional[Any]:
        """获取缓存

        自动检测是JSON还是Pickle并反序列化。

        Args:
            key: 缓存键

        Returns:
            缓存值，不存在则返回None

        Example:
            >>> cache = RedisCache()
            >>> user = await cache.get("user:123")
            >>> print(user)
            {"name": "Alice"}
        """
        try:
            async with redis_pool.acquire(self.db, "read") as conn:
                data = await conn.get(key)

                if data is None:
                    return None

                # 检测序列化类型并反序列化
                if data.startswith(self._SERIALIZATION_TYPE_JSON):
                    json_str = data[len(self._SERIALIZATION_TYPE_JSON) :].decode(
                        "utf-8"
                    )
                    return self.json_serializer.deserialize(json_str)
                elif data.startswith(self._SERIALIZATION_TYPE_PICKLE):
                    pickle_bytes = data[len(self._SERIALIZATION_TYPE_PICKLE) :]
                    return self.pickle_serializer.deserialize(pickle_bytes)
                else:
                    # 兼容旧数据（无类型标记，默认为Pickle）
                    logger.warning(f"键 {key} 缺少序列化类型标记，尝试Pickle反序列化")
                    return self.pickle_serializer.deserialize(data)

        except Exception as e:
            logger.error(f"获取缓存失败 key={key}: {e}")
            return None

    async def delete(self, key: str) -> bool:
        """删除缓存

        Args:
            key: 缓存键

        Returns:
            是否成功

        Example:
            >>> cache = RedisCache()
            >>> await cache.delete("user:123")
            True
        """
        try:
            async with redis_pool.acquire(self.db, "write") as conn:
                result = await conn.delete(key)
                return result > 0
        except Exception as e:
            logger.error(f"删除缓存失败 key={key}: {e}")
            return False

    async def exists(self, key: str) -> bool:
        """检查键是否存在

        Args:
            key: 缓存键

        Returns:
            是否存在

        Example:
            >>> cache = RedisCache()
            >>> exists = await cache.exists("user:123")
            >>> print(exists)
            True
        """
        try:
            async with redis_pool.acquire(self.db, "read") as conn:
                result = await conn.exists(key)
                return result > 0
        except Exception as e:
            logger.error(f"检查键存在失败 key={key}: {e}")
            return False

    async def expire(self, key: str, ttl: int) -> bool:
        """设置过期时间

        Args:
            key: 缓存键
            ttl: 过期时间（秒）

        Returns:
            是否成功

        Example:
            >>> cache = RedisCache()
            >>> await cache.expire("user:123", 3600)
            True
        """
        try:
            async with redis_pool.acquire(self.db, "write") as conn:
                result = await conn.expire(key, ttl)
                return bool(result)
        except Exception as e:
            logger.error(f"设置过期时间失败 key={key}, ttl={ttl}: {e}")
            return False

    async def ttl(self, key: str) -> int:
        """获取剩余过期时间

        Args:
            key: 缓存键

        Returns:
            剩余秒数，-1表示永不过期，-2表示不存在

        Example:
            >>> cache = RedisCache()
            >>> remaining = await cache.ttl("user:123")
            >>> print(f"剩余 {remaining} 秒")
        """
        try:
            async with redis_pool.acquire(self.db, "read") as conn:
                return await conn.ttl(key)
        except Exception as e:
            logger.error(f"获取TTL失败 key={key}: {e}")
            return -2

    async def eval(self, script: str, keys: List[str], args: List[Any]) -> Any:
        """执行Lua脚本

        Args:
            script: Lua脚本内容
            keys: 脚本中使用的键列表
            args: 脚本参数列表

        Returns:
            脚本执行结果

        Example:
            >>> cache = RedisCache()
            >>> script = '''
            ...     if redis.call("get", KEYS[1]) == ARGV[1] then
            ...         return redis.call("del", KEYS[1])
            ...     else
            ...         return 0
            ...     end
            ... '''
            >>> result = await cache.eval(script, ["lock:resource"], ["owner_id"])
        """
        try:
            async with redis_pool.acquire(self.db, "write") as conn:
                return await conn.eval(script, len(keys), *keys, *args)
        except Exception as e:
            logger.error(f"执行Lua脚本失败: {e}")
            raise
