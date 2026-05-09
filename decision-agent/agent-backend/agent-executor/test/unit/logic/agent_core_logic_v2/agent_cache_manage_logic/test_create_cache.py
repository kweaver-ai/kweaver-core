# -*- coding: utf-8 -*-
"""单元测试 - agent_cache_manage_logic/create_cache 模块"""

import pytest
from unittest.mock import AsyncMock, MagicMock, patch


class TestCreateCache:
    """测试 create_cache 函数"""

    @pytest.fixture
    def mock_manager(self):
        """创建 mock AgentCacheManager"""
        manager = MagicMock()
        manager.cache_service = AsyncMock()
        manager.cache_service.save.return_value = True
        return manager

    @pytest.fixture
    def mock_agent_config(self):
        """创建 mock AgentConfigVo"""
        config = MagicMock()
        config.get_config_last_set_timestamp.return_value = 1234567890
        return config

    @pytest.fixture
    def mock_agent_core_v2(self):
        """创建 mock AgentCoreV2"""
        core = MagicMock()
        core.warmup_handler = MagicMock()
        core.warmup_handler.warnup = AsyncMock()
        core.cache_handler = MagicMock()
        core.cache_handler.get_cache_data.return_value = MagicMock()
        return core

    @pytest.mark.asyncio
    async def test_create_cache_success(
        self, mock_manager, mock_agent_config, mock_agent_core_v2
    ):
        """测试创建缓存成功"""
        mock_cache_data = MagicMock()

        with patch(
            "app.logic.agent_core_logic_v2.agent_cache_manage_logic.create_cache.AgentCacheIdVO"
        ) as mock_cache_id_vo_class:
            with patch(
                "app.logic.agent_core_logic_v2.agent_cache_manage_logic.create_cache.AgentCacheEntity"
            ) as mock_entity_class:
                with patch(
                    "app.logic.agent_core_logic_v2.agent_core_v2.AgentCoreV2",
                    return_value=mock_agent_core_v2,
                ):
                    mock_cache_id_vo = MagicMock()
                    mock_cache_id_vo_class.return_value = mock_cache_id_vo

                    mock_entity = MagicMock()
                    mock_entity_class.return_value = mock_entity

                    mock_agent_core_v2.cache_handler.get_cache_data.return_value = (
                        mock_cache_data
                    )

                    from app.logic.agent_core_logic_v2.agent_cache_manage_logic.create_cache import (
                        create_cache,
                    )

                    result = await create_cache(
                        manager=mock_manager,
                        account_id="test_account",
                        account_type="user",
                        agent_id="test_agent",
                        agent_version="v1.0",
                        agent_config=mock_agent_config,
                        headers={"Authorization": "Bearer token"},
                    )

                    assert result == mock_entity
                    mock_agent_core_v2.warmup_handler.warnup.assert_called_once()
                    mock_manager.cache_service.save.assert_called_once()

    @pytest.mark.asyncio
    async def test_create_cache_exception(
        self, mock_manager, mock_agent_config, mock_agent_core_v2
    ):
        """测试创建缓存异常"""
        with patch(
            "app.logic.agent_core_logic_v2.agent_cache_manage_logic.create_cache.AgentCacheIdVO"
        ) as mock_cache_id_vo_class:
            with (
                patch(
                    "app.logic.agent_core_logic_v2.agent_core_v2.AgentCoreV2",
                    side_effect=Exception("Test error"),
                ),
                patch(
                    "app.logic.agent_core_logic_v2.agent_cache_manage_logic.create_cache.StandLogger.error"
                ) as mock_standard_error,
                patch(
                    "app.logic.agent_core_logic_v2.agent_cache_manage_logic.create_cache.o11y_logger"
                ) as mock_o11y_logger,
            ):
                mock_cache_id_vo = MagicMock()
                mock_cache_id_vo_class.return_value = mock_cache_id_vo

                from app.logic.agent_core_logic_v2.agent_cache_manage_logic.create_cache import (
                    create_cache,
                )

                with pytest.raises(Exception) as exc_info:
                    await create_cache(
                        manager=mock_manager,
                        account_id="test_account",
                        account_type="user",
                        agent_id="test_agent",
                        agent_version="v1.0",
                        agent_config=mock_agent_config,
                        headers={"Authorization": "Bearer token"},
                    )

                assert "Test error" in str(exc_info.value)
                mock_standard_error.assert_called_once()
                mock_o11y_logger().error.assert_called_once()

    @pytest.mark.asyncio
    async def test_create_cache_with_warmup_failure(
        self, mock_manager, mock_agent_config
    ):
        """测试预热失败时创建缓存"""
        mock_agent_core_v2 = MagicMock()
        mock_agent_core_v2.warmup_handler = MagicMock()
        mock_agent_core_v2.warmup_handler.warnup = AsyncMock(
            side_effect=Exception("Warmup failed")
        )

        with patch(
            "app.logic.agent_core_logic_v2.agent_cache_manage_logic.create_cache.AgentCacheIdVO"
        ) as mock_cache_id_vo_class:
            with patch(
                "app.logic.agent_core_logic_v2.agent_core_v2.AgentCoreV2",
                return_value=mock_agent_core_v2,
            ):
                mock_cache_id_vo = MagicMock()
                mock_cache_id_vo_class.return_value = mock_cache_id_vo

                from app.logic.agent_core_logic_v2.agent_cache_manage_logic.create_cache import (
                    create_cache,
                )

                with pytest.raises(Exception) as exc_info:
                    await create_cache(
                        manager=mock_manager,
                        account_id="test_account",
                        account_type="user",
                        agent_id="test_agent",
                        agent_version="v1.0",
                        agent_config=mock_agent_config,
                        headers={"Authorization": "Bearer token"},
                    )

                assert "Warmup failed" in str(exc_info.value)
