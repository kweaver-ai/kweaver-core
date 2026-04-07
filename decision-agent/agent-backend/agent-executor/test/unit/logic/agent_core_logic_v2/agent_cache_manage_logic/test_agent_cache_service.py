# -*- coding: utf-8 -*-
"""单元测试 - agent_cache_manage_logic/agent_cache_service 模块"""

import pytest
from unittest.mock import AsyncMock, MagicMock, patch


class TestAgentCacheService:
    """测试 AgentCacheService 类"""

    @pytest.fixture
    def mock_redis_cache(self):
        """创建 mock RedisCache"""
        mock = AsyncMock()
        return mock

    @pytest.fixture
    def mock_cache_entity(self):
        """创建 mock AgentCacheEntity"""
        entity = MagicMock()
        entity.cache_id_vo = MagicMock()
        entity.cache_id_vo.to_redis_key.return_value = "test_cache_key"
        return entity

    @pytest.fixture
    def mock_cache_id_vo(self):
        """创建 mock AgentCacheIdVO"""
        vo = MagicMock()
        vo.to_redis_key.return_value = "test_cache_key"
        return vo

    @pytest.mark.asyncio
    async def test_init(self, mock_redis_cache):
        """测试初始化"""
        with patch(
            "app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service.RedisCache",
            return_value=mock_redis_cache,
        ):
            from app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service import (
                AgentCacheService,
            )

            service = AgentCacheService(redis_db=3)
            assert service.redis_cache == mock_redis_cache
            assert service.redis_db == 3

    @pytest.mark.asyncio
    async def test_save_success(self, mock_redis_cache, mock_cache_entity):
        """测试保存缓存成功"""
        mock_redis_cache.set.return_value = True

        with patch(
            "app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service.RedisCache",
            return_value=mock_redis_cache,
        ):
            from app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service import (
                AgentCacheService,
            )

            service = AgentCacheService()
            result = await service.save(mock_cache_entity, ttl=3600)

            assert result is True
            mock_redis_cache.set.assert_called_once()

    @pytest.mark.asyncio
    async def test_save_failure(self, mock_redis_cache, mock_cache_entity):
        """测试保存缓存失败"""
        mock_redis_cache.set.return_value = False

        with patch(
            "app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service.RedisCache",
            return_value=mock_redis_cache,
        ):
            from app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service import (
                AgentCacheService,
            )

            service = AgentCacheService()
            result = await service.save(mock_cache_entity, ttl=3600)

            assert result is False

    @pytest.mark.asyncio
    async def test_save_exception(self, mock_redis_cache, mock_cache_entity):
        """测试保存缓存异常"""
        mock_redis_cache.set.side_effect = Exception("Redis error")

        with patch(
            "app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service.RedisCache",
            return_value=mock_redis_cache,
        ):
            from app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service import (
                AgentCacheService,
            )

            service = AgentCacheService()
            result = await service.save(mock_cache_entity, ttl=3600)

            assert result is False

    @pytest.mark.asyncio
    async def test_load_success(self, mock_redis_cache, mock_cache_id_vo):
        """测试加载缓存成功"""
        expected_data = MagicMock()
        mock_redis_cache.get.return_value = expected_data

        with patch(
            "app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service.RedisCache",
            return_value=mock_redis_cache,
        ):
            from app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service import (
                AgentCacheService,
            )

            service = AgentCacheService()
            result = await service.load(mock_cache_id_vo)

            assert result == expected_data

    @pytest.mark.asyncio
    async def test_load_not_found(self, mock_redis_cache, mock_cache_id_vo):
        """测试加载缓存不存在"""
        mock_redis_cache.get.return_value = None

        with patch(
            "app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service.RedisCache",
            return_value=mock_redis_cache,
        ):
            from app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service import (
                AgentCacheService,
            )

            service = AgentCacheService()
            result = await service.load(mock_cache_id_vo)

            assert result is None

    @pytest.mark.asyncio
    async def test_load_exception(self, mock_redis_cache, mock_cache_id_vo):
        """测试加载缓存异常"""
        mock_redis_cache.get.side_effect = Exception("Redis error")

        with patch(
            "app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service.RedisCache",
            return_value=mock_redis_cache,
        ):
            from app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service import (
                AgentCacheService,
            )

            service = AgentCacheService()
            result = await service.load(mock_cache_id_vo)

            assert result is None

    @pytest.mark.asyncio
    async def test_update_ttl_success(self, mock_redis_cache, mock_cache_id_vo):
        """测试更新TTL成功"""
        mock_redis_cache.expire.return_value = True

        with patch(
            "app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service.RedisCache",
            return_value=mock_redis_cache,
        ):
            from app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service import (
                AgentCacheService,
            )

            service = AgentCacheService()
            result = await service.update_ttl(mock_cache_id_vo, ttl=7200)

            assert result is True

    @pytest.mark.asyncio
    async def test_update_ttl_failure(self, mock_redis_cache, mock_cache_id_vo):
        """测试更新TTL失败"""
        mock_redis_cache.expire.return_value = False

        with patch(
            "app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service.RedisCache",
            return_value=mock_redis_cache,
        ):
            from app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service import (
                AgentCacheService,
            )

            service = AgentCacheService()
            result = await service.update_ttl(mock_cache_id_vo, ttl=7200)

            assert result is False

    @pytest.mark.asyncio
    async def test_update_ttl_exception(self, mock_redis_cache, mock_cache_id_vo):
        """测试更新TTL异常"""
        mock_redis_cache.expire.side_effect = Exception("Redis error")

        with patch(
            "app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service.RedisCache",
            return_value=mock_redis_cache,
        ):
            from app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service import (
                AgentCacheService,
            )

            service = AgentCacheService()
            result = await service.update_ttl(mock_cache_id_vo, ttl=7200)

            assert result is False

    @pytest.mark.asyncio
    async def test_exists_true(self, mock_redis_cache, mock_cache_id_vo):
        """测试检查缓存存在"""
        mock_redis_cache.exists.return_value = True

        with patch(
            "app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service.RedisCache",
            return_value=mock_redis_cache,
        ):
            from app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service import (
                AgentCacheService,
            )

            service = AgentCacheService()
            result = await service.exists(mock_cache_id_vo)

            assert result is True

    @pytest.mark.asyncio
    async def test_exists_false(self, mock_redis_cache, mock_cache_id_vo):
        """测试检查缓存不存在"""
        mock_redis_cache.exists.return_value = False

        with patch(
            "app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service.RedisCache",
            return_value=mock_redis_cache,
        ):
            from app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service import (
                AgentCacheService,
            )

            service = AgentCacheService()
            result = await service.exists(mock_cache_id_vo)

            assert result is False

    @pytest.mark.asyncio
    async def test_exists_exception(self, mock_redis_cache, mock_cache_id_vo):
        """测试检查缓存异常"""
        mock_redis_cache.exists.side_effect = Exception("Redis error")

        with patch(
            "app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service.RedisCache",
            return_value=mock_redis_cache,
        ):
            from app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service import (
                AgentCacheService,
            )

            service = AgentCacheService()
            result = await service.exists(mock_cache_id_vo)

            assert result is False

    @pytest.mark.asyncio
    async def test_get_ttl_success(self, mock_redis_cache, mock_cache_id_vo):
        """测试获取TTL成功"""
        mock_redis_cache.ttl.return_value = 3600

        with patch(
            "app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service.RedisCache",
            return_value=mock_redis_cache,
        ):
            from app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service import (
                AgentCacheService,
            )

            service = AgentCacheService()
            result = await service.get_ttl(mock_cache_id_vo)

            assert result == 3600

    @pytest.mark.asyncio
    async def test_get_ttl_exception(self, mock_redis_cache, mock_cache_id_vo):
        """测试获取TTL异常"""
        mock_redis_cache.ttl.side_effect = Exception("Redis error")

        with patch(
            "app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service.RedisCache",
            return_value=mock_redis_cache,
        ):
            from app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service import (
                AgentCacheService,
            )

            service = AgentCacheService()
            result = await service.get_ttl(mock_cache_id_vo)

            assert result == -2

    @pytest.mark.asyncio
    async def test_delete_success(self, mock_redis_cache, mock_cache_id_vo):
        """测试删除缓存成功"""
        mock_redis_cache.delete.return_value = True

        with patch(
            "app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service.RedisCache",
            return_value=mock_redis_cache,
        ):
            from app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service import (
                AgentCacheService,
            )

            service = AgentCacheService()
            result = await service.delete(mock_cache_id_vo)

            assert result is True

    @pytest.mark.asyncio
    async def test_delete_failure(self, mock_redis_cache, mock_cache_id_vo):
        """测试删除缓存失败"""
        mock_redis_cache.delete.return_value = False

        with patch(
            "app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service.RedisCache",
            return_value=mock_redis_cache,
        ):
            from app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service import (
                AgentCacheService,
            )

            service = AgentCacheService()
            result = await service.delete(mock_cache_id_vo)

            assert result is False

    @pytest.mark.asyncio
    async def test_delete_exception(self, mock_redis_cache, mock_cache_id_vo):
        """测试删除缓存异常"""
        mock_redis_cache.delete.side_effect = Exception("Redis error")

        with patch(
            "app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service.RedisCache",
            return_value=mock_redis_cache,
        ):
            from app.logic.agent_core_logic_v2.agent_cache_manage_logic.agent_cache_service import (
                AgentCacheService,
            )

            service = AgentCacheService()
            result = await service.delete(mock_cache_id_vo)

            assert result is False
