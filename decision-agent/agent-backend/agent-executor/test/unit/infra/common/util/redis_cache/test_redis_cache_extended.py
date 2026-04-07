# -*- coding: utf-8 -*-
"""单元测试 - app/infra/common/util/redis_cache/redis_cache.py 补充测试"""

import pytest
from unittest.mock import AsyncMock, patch

from app.infra.common.util.redis_cache.redis_cache import RedisCache


class TestRedisCacheSetFallback:
    """测试 RedisCache set 方法降级场景"""

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_set_json_type_error_fallback(self, mock_pool):
        """测试 JSON 序列化 TypeError 时降级到 Pickle"""
        mock_conn = AsyncMock()
        mock_conn.set = AsyncMock(return_value=True)
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)

        # Mock JSON serializer to raise TypeError
        with patch.object(
            cache.json_serializer,
            "serialize",
            side_effect=TypeError("Not JSON serializable"),
        ):
            # Mock pickle serializer to succeed
            with patch.object(
                cache.pickle_serializer,
                "serialize",
                return_value=b"\x80\x04\x95\x05\x00\x00\x00\x00\x00\x00\x00\x8c\x05test\x94.",
            ):
                result = await cache.set("test_key", {"data": "value"})

                assert result is True
                mock_conn.set.assert_called_once()

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_set_json_value_error_fallback(self, mock_pool):
        """测试 JSON 序列化 ValueError 时降级到 Pickle"""
        mock_conn = AsyncMock()
        mock_conn.set = AsyncMock(return_value=True)
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)

        # Mock JSON serializer to raise ValueError
        with patch.object(
            cache.json_serializer, "serialize", side_effect=ValueError("Invalid value")
        ):
            result = await cache.set("test_key", {"data": "value"})

            assert result is True
            mock_conn.set.assert_called_once()


class TestRedisCacheDelete:
    """测试 RedisCache delete 方法"""

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_delete_success(self, mock_pool):
        """测试成功删除键"""
        mock_conn = AsyncMock()
        mock_conn.delete = AsyncMock(return_value=1)
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        result = await cache.delete("test_key")

        assert result is True
        mock_conn.delete.assert_called_once_with("test_key")

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_delete_non_existing_key(self, mock_pool):
        """测试删除不存在的键"""
        mock_conn = AsyncMock()
        mock_conn.delete = AsyncMock(return_value=0)
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        result = await cache.delete("non_existing_key")

        assert result is False

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_delete_with_exception(self, mock_pool):
        """测试删除缓存时发生异常"""
        mock_conn = AsyncMock()
        mock_conn.delete = AsyncMock(side_effect=Exception("Redis connection error"))
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        result = await cache.delete("test_key")

        assert result is False


class TestRedisCacheExists:
    """测试 RedisCache exists 方法"""

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_exists_true(self, mock_pool):
        """测试键存在"""
        mock_conn = AsyncMock()
        mock_conn.exists = AsyncMock(return_value=1)
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        result = await cache.exists("test_key")

        assert result is True
        mock_conn.exists.assert_called_once_with("test_key")

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_exists_false(self, mock_pool):
        """测试键不存在"""
        mock_conn = AsyncMock()
        mock_conn.exists = AsyncMock(return_value=0)
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        result = await cache.exists("non_existing_key")

        assert result is False

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_exists_with_exception(self, mock_pool):
        """测试检查键存在时发生异常"""
        mock_conn = AsyncMock()
        mock_conn.exists = AsyncMock(side_effect=Exception("Redis error"))
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        result = await cache.exists("test_key")

        assert result is False


class TestRedisCacheExpire:
    """测试 RedisCache expire 方法"""

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_expire_success(self, mock_pool):
        """测试成功设置过期时间"""
        mock_conn = AsyncMock()
        mock_conn.expire = AsyncMock(return_value=True)
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        result = await cache.expire("test_key", 3600)

        assert result is True
        mock_conn.expire.assert_called_once_with("test_key", 3600)

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_expire_non_existing_key(self, mock_pool):
        """测试对不存在的键设置过期时间"""
        mock_conn = AsyncMock()
        mock_conn.expire = AsyncMock(return_value=False)
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        result = await cache.expire("non_existing_key", 3600)

        assert result is False

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_expire_with_exception(self, mock_pool):
        """测试设置过期时间时发生异常"""
        mock_conn = AsyncMock()
        mock_conn.expire = AsyncMock(side_effect=Exception("Redis error"))
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        result = await cache.expire("test_key", 3600)

        assert result is False


class TestRedisCacheTTL:
    """测试 RedisCache ttl 方法"""

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_ttl_existing_key(self, mock_pool):
        """测试获取存在键的TTL"""
        mock_conn = AsyncMock()
        mock_conn.ttl = AsyncMock(return_value=3600)
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        result = await cache.ttl("test_key")

        assert result == 3600
        mock_conn.ttl.assert_called_once_with("test_key")

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_ttl_no_expiry(self, mock_pool):
        """测试获取永不过期的键的TTL"""
        mock_conn = AsyncMock()
        mock_conn.ttl = AsyncMock(return_value=-1)
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        result = await cache.ttl("test_key")

        assert result == -1

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_ttl_non_existing_key(self, mock_pool):
        """测试获取不存在键的TTL"""
        mock_conn = AsyncMock()
        mock_conn.ttl = AsyncMock(return_value=-2)
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        result = await cache.ttl("non_existing_key")

        assert result == -2

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_ttl_with_exception(self, mock_pool):
        """测试获取TTL时发生异常"""
        mock_conn = AsyncMock()
        mock_conn.ttl = AsyncMock(side_effect=Exception("Redis error"))
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        result = await cache.ttl("test_key")

        # 异常时应返回 -2
        assert result == -2


class TestRedisCacheEval:
    """测试 RedisCache eval 方法"""

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_eval_success(self, mock_pool):
        """测试成功执行Lua脚本"""
        mock_conn = AsyncMock()
        mock_conn.eval = AsyncMock(return_value=1)
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        script = "return redis.call('GET', KEYS[1])"
        result = await cache.eval(script, ["test_key"], [])

        assert result == 1
        mock_conn.eval.assert_called_once()

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_eval_with_args(self, mock_pool):
        """测试带参数执行Lua脚本"""
        mock_conn = AsyncMock()
        mock_conn.eval = AsyncMock(return_value="OK")
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        script = """
            if redis.call("get", KEYS[1]) == ARGV[1] then
                return redis.call("del", KEYS[1])
            else
                return 0
            end
        """
        result = await cache.eval(script, ["lock:resource"], ["owner_id"])

        assert result == "OK"

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_eval_with_exception(self, mock_pool):
        """测试执行Lua脚本时发生异常"""
        mock_conn = AsyncMock()
        mock_conn.eval = AsyncMock(side_effect=Exception("Lua script error"))
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        script = "invalid script"

        with pytest.raises(Exception) as exc_info:
            await cache.eval(script, ["test_key"], [])

        assert "Lua script error" in str(exc_info.value)


class TestRedisCacheEdgeCases:
    """测试 RedisCache 边界情况"""

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_set_empty_dict(self, mock_pool):
        """测试设置空字典"""
        mock_conn = AsyncMock()
        mock_conn.set = AsyncMock(return_value=True)
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        result = await cache.set("empty_dict", {})

        assert result is True

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_set_empty_list(self, mock_pool):
        """测试设置空列表"""
        mock_conn = AsyncMock()
        mock_conn.set = AsyncMock(return_value=True)
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        result = await cache.set("empty_list", [])

        assert result is True

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_set_none_value(self, mock_pool):
        """测试设置 None 值"""
        mock_conn = AsyncMock()
        mock_conn.set = AsyncMock(return_value=True)
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        result = await cache.set("none_key", None)

        assert result is True

    @pytest.mark.asyncio
    @patch("app.infra.common.util.redis_cache.redis_cache.redis_pool")
    async def test_set_with_zero_ttl(self, mock_pool):
        """测试设置 TTL 为 0"""
        mock_conn = AsyncMock()
        mock_conn.set = AsyncMock(return_value=True)
        mock_pool.acquire.return_value.__aenter__.return_value = mock_conn

        cache = RedisCache(db=3)
        result = await cache.set("test_key", "value", ttl=0)

        assert result is True
        call_args = mock_conn.set.call_args
        assert call_args[1]["ex"] == 0
