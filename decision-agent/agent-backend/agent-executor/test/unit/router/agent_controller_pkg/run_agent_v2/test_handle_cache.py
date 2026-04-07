# -*- coding: utf-8 -*-
"""单元测试 - run_agent_v2/handle_cache 模块"""

import pytest
from unittest.mock import MagicMock, AsyncMock, patch


class TestHandleCache:
    """测试 handle_cache 函数"""

    @pytest.fixture
    def mock_agent_core_v2(self):
        """创建 mock AgentCoreV2"""
        core = MagicMock()
        core.agent_config = MagicMock()
        core.agent_config.agent_version = "v1.0"
        core.agent_config.get_config_last_set_timestamp.return_value = 1234567890
        core.cache_handler = MagicMock()
        return core

    @pytest.mark.asyncio
    async def test_handle_cache_debug_mode(self, mock_agent_core_v2):
        """测试调试模式跳过缓存"""
        from app.router.agent_controller_pkg.run_agent_v2.handle_cache import (
            handle_cache,
        )

        result = await handle_cache(
            agent_id="test_agent",
            agent_core_v2=mock_agent_core_v2,
            is_debug_run=True,
            headers={},
            account_id="user123",
            account_type="standard",
        )

        assert result is None

    @pytest.mark.asyncio
    async def test_handle_cache_existing_cache(self, mock_agent_core_v2):
        """测试缓存存在的情况"""
        mock_cache_entity = MagicMock()
        mock_cache_entity.cache_data = MagicMock()

        with patch(
            "app.router.agent_controller_pkg.run_agent_v2.handle_cache.cache_manager"
        ) as mock_manager:
            mock_manager.cache_service.load = AsyncMock(return_value=mock_cache_entity)
            mock_manager.cache_service.get_ttl = AsyncMock(return_value=3600)

            from app.router.agent_controller_pkg.run_agent_v2.handle_cache import (
                handle_cache,
            )

            result = await handle_cache(
                agent_id="test_agent",
                agent_core_v2=mock_agent_core_v2,
                is_debug_run=False,
                headers={},
                account_id="user123",
                account_type="standard",
            )

            assert result is not None
            mock_agent_core_v2.cache_handler.set_cache_data.assert_called_once()

    @pytest.mark.asyncio
    async def test_handle_cache_create_new_cache(self, mock_agent_core_v2):
        """测试缓存不存在时创建新缓存"""
        mock_cache_entity = MagicMock()
        mock_cache_entity.cache_data = MagicMock()

        with patch(
            "app.router.agent_controller_pkg.run_agent_v2.handle_cache.cache_manager"
        ) as mock_manager:
            mock_manager.cache_service.load = AsyncMock(return_value=None)
            mock_manager.cache_service.get_ttl = AsyncMock(return_value=7200)
            mock_manager.create_cache = AsyncMock(return_value=mock_cache_entity)

            from app.router.agent_controller_pkg.run_agent_v2.handle_cache import (
                handle_cache,
            )

            result = await handle_cache(
                agent_id="test_agent",
                agent_core_v2=mock_agent_core_v2,
                is_debug_run=False,
                headers={"Authorization": "Bearer token"},
                account_id="user123",
                account_type="standard",
            )

            assert result is not None
            mock_manager.create_cache.assert_called_once()

    @pytest.mark.asyncio
    async def test_handle_cache_create_cache_exception(self, mock_agent_core_v2):
        """测试创建缓存异常"""
        with patch(
            "app.router.agent_controller_pkg.run_agent_v2.handle_cache.cache_manager"
        ) as mock_manager:
            mock_manager.cache_service.load = AsyncMock(return_value=None)
            mock_manager.create_cache = AsyncMock(
                side_effect=Exception("Create failed")
            )

            from app.router.agent_controller_pkg.run_agent_v2.handle_cache import (
                handle_cache,
            )

            result = await handle_cache(
                agent_id="test_agent",
                agent_core_v2=mock_agent_core_v2,
                is_debug_run=False,
                headers={},
                account_id="user123",
                account_type="standard",
            )

            assert result is None
