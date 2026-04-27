"""单元测试 - app/infra/common/util/redis_cache/redis_cache.py"""

import pytest
from unittest.mock import AsyncMock, MagicMock, patch
import pickle

from app.infra.common.util.redis_cache.redis_cache import RedisCache, SerializationType


@pytest.fixture
def mock_redis_pool():
    """Mock Redis 连接池"""
    return MagicMock()


@pytest.fixture
def mock_connection():
    """Mock Redis 连接"""
    conn = AsyncMock()
    conn.get = AsyncMock(return_value=None)
    conn.set = AsyncMock(return_value=True)
    conn.delete = AsyncMock(return_value=1)
    conn.exists = AsyncMock(return_value=1)
    conn.expire = AsyncMock(return_value=True)
    return conn


@pytest.fixture
def redis_cache(mock_redis_pool, mock_connection):
    """创建 RedisCache 实例"""
    with patch(
        "app.infra.common.util.redis_cache.redis_cache.redis_pool", mock_redis_pool
    ):
        cache = RedisCache(db=3)
        return cache


class TestRedisCacheInit:
    """测试 RedisCache 初始化"""

    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    def test_init_with_default_db(self, mock_pool):
        """测试使用默认数据库编号初始化"""
        cache = RedisCache()
        assert cache.db == 3
        assert cache.json_serializer is not None
        assert cache.pickle_serializer is not None

    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    def test_init_with_custom_db(self, mock_pool):
        """测试使用自定义数据库编号初始化"""
        cache = RedisCache(db=5)
        assert cache.db == 5


class TestRedisCacheSet:
    """测试 RedisCache set 方法"""

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_set_with_json_serialization(self, mock_pool):
        """测试使用 JSON 序列化设置缓存"""
        # 设置 mock
        mock_conn = AsyncMock()
        mock_conn.set = AsyncMock(return_value=True)
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        result = await cache.set("test_key", {"name": "Alice", "age": 30})

        assert result is True
        mock_conn.set.assert_called_once()

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_set_with_ttl(self, mock_pool):
        """测试设置带 TTL 的缓存"""
        mock_conn = AsyncMock()
        mock_conn.set = AsyncMock(return_value=True)
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        result = await cache.set("test_key", "value", ttl=3600)

        assert result is True
        # 验证调用了带 ex 参数的 set
        call_args = mock_conn.set.call_args
        assert call_args[1]["ex"] == 3600

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_set_with_pickle_serialization(self, mock_pool):
        """测试使用 Pickle 序列化设置缓存"""
        mock_conn = AsyncMock()
        mock_conn.set = AsyncMock(return_value=True)
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        result = await cache.set(
            "test_key",
            {"complex": "object"},
            serialization_type=SerializationType.PICKLE,
        )

        assert result is True
        mock_conn.set.assert_called_once()

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_set_json_fallback_to_pickle(self, mock_pool):
        """测试 JSON 序列化失败时降级到 Pickle"""
        mock_conn = AsyncMock()
        mock_conn.set = AsyncMock(return_value=True)
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)

        # 使用 datetime 对象（JSON 序列化会特殊处理，但可以作为示例）
        from datetime import datetime

        test_data = {"timestamp": datetime.now()}

        result = await cache.set("test_key", test_data)

        assert result is True
        mock_conn.set.assert_called_once()

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_set_with_exception(self, mock_pool):
        """测试设置缓存时发生异常"""
        mock_conn = AsyncMock()
        mock_conn.set = AsyncMock(side_effect=Exception("Redis connection error"))
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        result = await cache.set("test_key", "value")

        assert result is False


class TestRedisCacheGet:
    """测试 RedisCache get 方法"""

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_get_existing_key(self, mock_pool):
        """测试获取存在的键"""
        # 准备测试数据
        test_data = {"name": "Bob", "age": 25}
        serialized = b'JSON:{"name": "Bob", "age": 25}'

        mock_conn = AsyncMock()
        mock_conn.get = AsyncMock(return_value=serialized)
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        result = await cache.get("test_key")

        assert result == test_data
        mock_conn.get.assert_called_once_with("test_key")

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_get_non_existing_key(self, mock_pool):
        """测试获取不存在的键"""
        mock_conn = AsyncMock()
        mock_conn.get = AsyncMock(return_value=None)
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        result = await cache.get("non_existing_key")

        assert result is None

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_get_pickle_serialized_data(self, mock_pool):
        """测试获取 Pickle 序列化的数据"""
        test_data = {"complex": "data"}
        serialized = b"PICKLE:" + pickle.dumps(test_data)

        mock_conn = AsyncMock()
        mock_conn.get = AsyncMock(return_value=serialized)
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        result = await cache.get("test_key")

        assert result == test_data

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_get_with_exception(self, mock_pool):
        """测试获取缓存时发生异常"""
        mock_conn = AsyncMock()
        mock_conn.get = AsyncMock(side_effect=Exception("Redis error"))
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        result = await cache.get("test_key")

        assert result is None

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_get_legacy_data_without_prefix(self, mock_pool):
        """测试获取没有类型标记的旧数据"""
        test_data = "legacy data"
        serialized = pickle.dumps(test_data)

        mock_conn = AsyncMock()
        mock_conn.get = AsyncMock(return_value=serialized)
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        result = await cache.get("legacy_key")

        assert result == test_data


class TestSerializationType:
    """测试 SerializationType 枚举"""

    def test_serialization_type_values(self):
        """测试序列化类型枚举值"""
        assert SerializationType.JSON.value == "json"
        assert SerializationType.PICKLE.value == "pickle"

    def test_serialization_type_comparison(self):
        """测试序列化类型比较"""
        assert SerializationType.JSON == SerializationType.JSON
        assert SerializationType.JSON != SerializationType.PICKLE


class TestRedisCacheIntegration:
    """测试 RedisCache 集成场景"""

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_set_and_get_cycle(self, mock_pool):
        """测试完整的设置和获取周期"""
        test_data = {"user_id": 123, "username": "test_user"}

        # Mock set 操作
        mock_conn_set = AsyncMock()
        mock_conn_set.set = AsyncMock(return_value=True)
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn_set

        cache = RedisCache(db=3)

        # 设置缓存
        set_result = await cache.set("user:123", test_data)
        assert set_result is True

        # Mock get 操作
        serialized = b'JSON:{"user_id": 123, "username": "test_user"}'
        mock_conn_get = AsyncMock()
        mock_conn_get.get = AsyncMock(return_value=serialized)
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn_get

        # 获取缓存
        get_result = await cache.get("user:123")
        assert get_result == test_data

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_set_multiple_types(self, mock_pool):
        """测试设置多种数据类型"""
        mock_conn = AsyncMock()
        mock_conn.set = AsyncMock(return_value=True)
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)

        # 测试字符串
        assert await cache.set("str_key", "string value")

        # 测试数字
        assert await cache.set("num_key", 42)

        # 测试布尔值
        assert await cache.set("bool_key", True)

        # 测试列表
        assert await cache.set("list_key", [1, 2, 3])

        # 测试字典
        assert await cache.set("dict_key", {"a": 1, "b": 2})

        # 验证所有 set 操作都被调用
        assert mock_conn.set.call_count == 5
