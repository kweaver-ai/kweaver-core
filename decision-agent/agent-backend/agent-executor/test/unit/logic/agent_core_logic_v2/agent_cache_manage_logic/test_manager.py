# -*- coding: utf-8 -*-
"""单元测试 - agent_cache_manage_logic/manager 模块"""

import pytest
from unittest.mock import AsyncMock, MagicMock, patch


class TestAgentCacheManager:
    """测试 AgentCacheManager 类"""

    @pytest.fixture
    def mock_cache_service(self):
        """创建 mock AgentCacheService"""
        mock = AsyncMock()
        return mock

    @pytest.fixture
    def mock_agent_config(self):
        """创建 mock AgentConfigVo"""
        config = MagicMock()
        config.get_config_last_set_timestamp.return_value = 1234567890
        return config

    @pytest.fixture
    def mock_cache_id_vo(self):
        """创建 mock AgentCacheIdVO"""
        vo = MagicMock()
        vo.account_id = "test_account"
        vo.account_type = "user"
        vo.agent_id = "test_agent"
        vo.agent_version = "v1.0"
        return vo

    @pytest.mark.asyncio
    async def test_init(self, mock_cache_service):
        """测试初始化"""
        with patch(
            "app.logic.agent_core_logic_v2.agent_cache_manage_logic.manager.AgentCacheService",
            return_value=mock_cache_service,
        ):
            from app.logic.agent_core_logic_v2.agent_cache_manage_logic.manager import (
                AgentCacheManager,
            )

            manager = AgentCacheManager()
            assert manager.cache_service == mock_cache_service

    @pytest.mark.asyncio
    async def test_create_cache_success(self, mock_cache_service, mock_agent_config):
        """测试创建缓存成功"""
        mock_cache_entity = MagicMock()
        mock_cache_service.save.return_value = True

        with patch(
            "app.logic.agent_core_logic_v2.agent_cache_manage_logic.manager.AgentCacheService",
            return_value=mock_cache_service,
        ):
            with patch(
                "app.logic.agent_core_logic_v2.agent_cache_manage_logic.create_cache.create_cache",
                new_callable=AsyncMock,
                return_value=mock_cache_entity,
            ):
                from app.logic.agent_core_logic_v2.agent_cache_manage_logic.manager import (
                    AgentCacheManager,
                )

                manager = AgentCacheManager()
                result = await manager.create_cache(
                    account_id="test_account",
                    account_type="user",
                    agent_id="test_agent",
                    agent_version="v1.0",
                    agent_config=mock_agent_config,
                    headers={"Authorization": "Bearer token"},
                )

                assert result == mock_cache_entity

    @pytest.mark.asyncio
    async def test_update_cache_data_success(
        self, mock_cache_service, mock_agent_config, mock_cache_id_vo
    ):
        """测试更新缓存数据成功"""
        with patch(
            "app.logic.agent_core_logic_v2.agent_cache_manage_logic.manager.AgentCacheService",
            return_value=mock_cache_service,
        ):
            with patch(
                "app.logic.agent_core_logic_v2.agent_cache_manage_logic.update_cache_data.update_cache_data",
                new_callable=AsyncMock,
            ):
                from app.logic.agent_core_logic_v2.agent_cache_manage_logic.manager import (
                    AgentCacheManager,
                )

                manager = AgentCacheManager()
                await manager.update_cache_data(
                    cache_id_vo=mock_cache_id_vo,
                    agent_config=mock_agent_config,
                    headers={"Authorization": "Bearer token"},
                )

                # 验证 update_cache_data 被调用
