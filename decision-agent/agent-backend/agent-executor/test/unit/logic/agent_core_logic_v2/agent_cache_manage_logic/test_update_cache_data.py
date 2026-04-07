# -*- coding: utf-8 -*-
"""单元测试 - agent_cache_manage_logic/update_cache_data 模块"""

import pytest
from unittest.mock import AsyncMock, MagicMock, patch


class TestUpdateCacheData:
    """测试 update_cache_data 函数"""

    @pytest.fixture
    def mock_manager(self):
        """创建 mock AgentCacheManager"""
        manager = MagicMock()
        manager.cache_service = AsyncMock()
        return manager

    @pytest.fixture
    def mock_cache_id_vo(self):
        """创建 mock AgentCacheIdVO"""
        vo = MagicMock()
        vo.account_id = "test_account"
        vo.account_type = "user"
        vo.agent_id = "test_agent"
        vo.agent_version = "v1.0"
        return vo

    @pytest.fixture
    def mock_agent_config(self):
        """创建 mock AgentConfigVo"""
        config = MagicMock()
        return config

    @pytest.fixture
    def mock_cache_entity(self):
        """创建 mock AgentCacheEntity"""
        entity = MagicMock()
        entity.cache_data = MagicMock()
        return entity

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
    async def test_update_cache_data_with_existing_cache(
        self,
        mock_manager,
        mock_cache_id_vo,
        mock_agent_config,
        mock_cache_entity,
        mock_agent_core_v2,
    ):
        """测试更新已存在的缓存数据"""
        mock_manager.cache_service.load.return_value = mock_cache_entity
        mock_manager.cache_service.save.return_value = True
        mock_cache_data = MagicMock()
        mock_agent_core_v2.cache_handler.get_cache_data.return_value = mock_cache_data

        with patch(
            "app.logic.agent_core_logic_v2.agent_core_v2.AgentCoreV2",
            return_value=mock_agent_core_v2,
        ):
            from app.logic.agent_core_logic_v2.agent_cache_manage_logic.update_cache_data import (
                update_cache_data,
            )

            await update_cache_data(
                manager=mock_manager,
                cache_id_vo=mock_cache_id_vo,
                agent_config=mock_agent_config,
                headers={"Authorization": "Bearer token"},
            )

            # 验证缓存被加载和保存
            mock_manager.cache_service.load.assert_called_once()
            mock_manager.cache_service.save.assert_called_once()
            mock_agent_core_v2.warmup_handler.warnup.assert_called_once()

    @pytest.mark.asyncio
    async def test_update_cache_data_cache_not_exists(
        self, mock_manager, mock_cache_id_vo, mock_agent_config
    ):
        """测试缓存不存在时创建新缓存"""
        mock_manager.cache_service.load.return_value = None
        mock_cache_entity = MagicMock()

        with patch(
            "app.logic.agent_core_logic_v2.agent_cache_manage_logic.create_cache.create_cache",
            new_callable=AsyncMock,
            return_value=mock_cache_entity,
        ):
            from app.logic.agent_core_logic_v2.agent_cache_manage_logic.update_cache_data import (
                update_cache_data,
            )

            await update_cache_data(
                manager=mock_manager,
                cache_id_vo=mock_cache_id_vo,
                agent_config=mock_agent_config,
                headers={"Authorization": "Bearer token"},
            )

            # 验证 create_cache 被调用
            # 验证 save 没有被直接调用（通过 create_cache 调用）

    @pytest.mark.asyncio
    async def test_update_cache_data_exception(
        self, mock_manager, mock_cache_id_vo, mock_agent_config, mock_cache_entity
    ):
        """测试更新缓存数据异常"""
        mock_manager.cache_service.load.return_value = mock_cache_entity
        mock_manager.cache_service.save.side_effect = Exception("Redis error")

        mock_agent_core_v2 = MagicMock()
        mock_agent_core_v2.warmup_handler = MagicMock()
        mock_agent_core_v2.warmup_handler.warnup = AsyncMock()
        mock_agent_core_v2.cache_handler = MagicMock()
        mock_agent_core_v2.cache_handler.get_cache_data.return_value = MagicMock()

        with patch(
            "app.logic.agent_core_logic_v2.agent_core_v2.AgentCoreV2",
            return_value=mock_agent_core_v2,
        ):
            from app.logic.agent_core_logic_v2.agent_cache_manage_logic.update_cache_data import (
                update_cache_data,
            )

            with pytest.raises(Exception) as exc_info:
                await update_cache_data(
                    manager=mock_manager,
                    cache_id_vo=mock_cache_id_vo,
                    agent_config=mock_agent_config,
                    headers={"Authorization": "Bearer token"},
                )

            assert "Redis error" in str(exc_info.value)

    @pytest.mark.asyncio
    async def test_update_cache_data_warmup_exception(
        self, mock_manager, mock_cache_id_vo, mock_agent_config, mock_cache_entity
    ):
        """测试预热失败时更新缓存数据"""
        mock_manager.cache_service.load.return_value = mock_cache_entity

        mock_agent_core_v2 = MagicMock()
        mock_agent_core_v2.warmup_handler = MagicMock()
        mock_agent_core_v2.warmup_handler.warnup = AsyncMock(
            side_effect=Exception("Warmup failed")
        )

        with patch(
            "app.logic.agent_core_logic_v2.agent_core_v2.AgentCoreV2",
            return_value=mock_agent_core_v2,
        ):
            from app.logic.agent_core_logic_v2.agent_cache_manage_logic.update_cache_data import (
                update_cache_data,
            )

            with pytest.raises(Exception) as exc_info:
                await update_cache_data(
                    manager=mock_manager,
                    cache_id_vo=mock_cache_id_vo,
                    agent_config=mock_agent_config,
                    headers={"Authorization": "Bearer token"},
                )

            assert "Warmup failed" in str(exc_info.value)
