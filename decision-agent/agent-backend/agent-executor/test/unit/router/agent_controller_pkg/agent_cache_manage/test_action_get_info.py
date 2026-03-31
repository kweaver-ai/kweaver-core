# -*- coding: utf-8 -*-
"""单元测试 - agent_cache_manage/action_get_info 模块"""

import pytest
from unittest.mock import MagicMock, AsyncMock, patch


class TestHandleGetInfo:
    """测试 handle_get_info 函数"""

    @pytest.fixture
    def mock_request(self):
        """创建 mock Request"""
        request = MagicMock()
        request.headers = {}
        return request

    @pytest.fixture
    def mock_agent_config(self):
        """创建 mock AgentConfigVo"""
        config = MagicMock()
        config.get_config_last_set_timestamp.return_value = 1234567890
        return config

    @pytest.mark.asyncio
    async def test_handle_get_info_cache_exists(self, mock_request, mock_agent_config):
        """测试缓存存在的情况"""
        mock_cache_entity = MagicMock()
        mock_cache_entity.created_at = "2024-01-01T00:00:00"
        mock_cache_entity.cache_data = MagicMock()

        mock_cache_id_vo = MagicMock()
        mock_cache_id_vo.get_cache_id.return_value = "test_cache_id"

        with patch(
            "app.router.agent_controller_pkg.agent_cache_manage.action_get_info.build_cache_id_vo",
            return_value=mock_cache_id_vo,
        ):
            with patch(
                "app.router.agent_controller_pkg.agent_cache_manage.action_get_info.cache_manager"
            ) as mock_manager:
                mock_manager.cache_service.load = AsyncMock(
                    return_value=mock_cache_entity
                )
                mock_manager.cache_service.get_ttl = AsyncMock(return_value=3600)

                with patch(
                    "app.router.agent_controller_pkg.agent_cache_manage.action_get_info.get_cache_data_for_debug_mode",
                    return_value={},
                ):
                    from app.router.agent_controller_pkg.agent_cache_manage.action_get_info import (
                        handle_get_info,
                    )

                    result = await handle_get_info(
                        request=mock_request,
                        account_id="test_account",
                        account_type="user",
                        agent_id="test_agent",
                        agent_version="v1.0",
                        agent_config=mock_agent_config,
                    )

                    assert result is not None
                    assert result.cache_id == "test_cache_id"
                    assert result.ttl == 3600

    @pytest.mark.asyncio
    async def test_handle_get_info_cache_not_exists(
        self, mock_request, mock_agent_config
    ):
        """测试缓存不存在的情况"""
        mock_cache_id_vo = MagicMock()

        with patch(
            "app.router.agent_controller_pkg.agent_cache_manage.action_get_info.build_cache_id_vo",
            return_value=mock_cache_id_vo,
        ):
            with patch(
                "app.router.agent_controller_pkg.agent_cache_manage.action_get_info.cache_manager"
            ) as mock_manager:
                mock_manager.cache_service.load = AsyncMock(return_value=None)

                from app.router.agent_controller_pkg.agent_cache_manage.action_get_info import (
                    handle_get_info,
                )

                result = await handle_get_info(
                    request=mock_request,
                    account_id="test_account",
                    account_type="user",
                    agent_id="test_agent",
                    agent_version="v1.0",
                    agent_config=mock_agent_config,
                )

                assert result is None
