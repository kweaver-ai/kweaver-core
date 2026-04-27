# -*- coding: utf-8 -*-
"""单元测试 - agent_cache_manage/action_upsert 模块"""

import pytest
from unittest.mock import MagicMock, AsyncMock, patch


class TestHandleUpsert:
    """测试 handle_upsert 函数"""

    @pytest.fixture
    def mock_request(self):
        """创建 mock Request"""
        request = MagicMock()
        request.headers = {"Authorization": "Bearer token"}
        return request

    @pytest.fixture
    def mock_agent_config(self):
        """创建 mock AgentConfigVo"""
        config = MagicMock()
        config.get_config_last_set_timestamp.return_value = 1234567890
        return config

    @pytest.mark.asyncio
    async def test_handle_upsert_cache_exists(self, mock_request, mock_agent_config):
        """测试缓存存在时更新"""
        mock_cache_entity = MagicMock()
        mock_cache_entity.created_at = "2024-01-01T00:00:00"
        mock_cache_entity.cache_data = MagicMock()

        mock_cache_id_vo = MagicMock()
        mock_cache_id_vo.get_cache_id.return_value = "test_cache_id"

        with patch(
            "app.router.agent_controller_pkg.agent_cache_manage.action_upsert.build_cache_id_vo",
            return_value=mock_cache_id_vo,
        ):
            with patch(
                "app.router.agent_controller_pkg.agent_cache_manage.action_upsert.cache_manager"
            ) as mock_manager:
                mock_manager.cache_service.load = AsyncMock(
                    return_value=mock_cache_entity
                )
                mock_manager.cache_service.get_ttl = AsyncMock(return_value=3600)
                mock_manager.update_cache_data = AsyncMock()

                with (
                    patch(
                        "app.router.agent_controller_pkg.agent_cache_manage.action_upsert.get_cache_data_for_debug_mode",
                        return_value={},
                    ),
                    patch(
                        "app.router.agent_controller_pkg.agent_cache_manage.action_upsert.StandLogger.info"
                    ) as mock_standard_info,
                ):
                    from app.router.agent_controller_pkg.agent_cache_manage.action_upsert import (
                        handle_upsert,
                    )

                    result = await handle_upsert(
                        request=mock_request,
                        account_id="test_account",
                        account_type="user",
                        agent_id="test_agent",
                        agent_version="v1.0",
                        agent_config=mock_agent_config,
                    )

                    assert result is not None
                    assert result.cache_id == "test_cache_id"
                    mock_manager.update_cache_data.assert_called_once()
                    mock_standard_info.assert_called_once()

    @pytest.mark.asyncio
    async def test_handle_upsert_cache_not_exists(
        self, mock_request, mock_agent_config
    ):
        """测试缓存不存在时创建"""
        mock_cache_id_vo = MagicMock()
        mock_cache_id_vo.get_cache_id.return_value = "test_cache_id"

        mock_cache_entity = MagicMock()
        mock_cache_entity.cache_id_vo = mock_cache_id_vo
        mock_cache_entity.created_at = "2024-01-01T00:00:00"
        mock_cache_entity.cache_data = MagicMock()

        with patch(
            "app.router.agent_controller_pkg.agent_cache_manage.action_upsert.build_cache_id_vo",
            return_value=mock_cache_id_vo,
        ):
            with patch(
                "app.router.agent_controller_pkg.agent_cache_manage.action_upsert.cache_manager"
            ) as mock_manager:
                # 第一次 load 返回 None（缓存不存在）
                mock_manager.cache_service.load = AsyncMock(return_value=None)
                mock_manager.cache_service.get_ttl = AsyncMock(return_value=3600)

                with (
                    patch(
                        "app.router.agent_controller_pkg.agent_cache_manage.action_upsert.create_cache_and_build_response",
                        new_callable=AsyncMock,
                    ) as mock_create,
                    patch(
                        "app.router.agent_controller_pkg.agent_cache_manage.action_upsert.StandLogger.info"
                    ) as mock_standard_info,
                ):
                    mock_response = MagicMock()
                    mock_response.cache_id = "test_cache_id"
                    mock_create.return_value = mock_response

                    from app.router.agent_controller_pkg.agent_cache_manage.action_upsert import (
                        handle_upsert,
                    )

                    result = await handle_upsert(
                        request=mock_request,
                        account_id="test_account",
                        account_type="user",
                        agent_id="test_agent",
                        agent_version="v1.0",
                        agent_config=mock_agent_config,
                    )

                    assert result is not None
                    mock_create.assert_called_once()
                    mock_standard_info.assert_called_once()

    @pytest.mark.asyncio
    async def test_handle_upsert_cache_expired_during_update(
        self, mock_request, mock_agent_config
    ):
        """测试缓存更新过程中过期"""
        mock_cache_id_vo = MagicMock()
        mock_cache_id_vo.get_cache_id.return_value = "test_cache_id"

        mock_cache_entity = MagicMock()
        mock_cache_entity.created_at = "2024-01-01T00:00:00"

        with patch(
            "app.router.agent_controller_pkg.agent_cache_manage.action_upsert.build_cache_id_vo",
            return_value=mock_cache_id_vo,
        ):
            with patch(
                "app.router.agent_controller_pkg.agent_cache_manage.action_upsert.cache_manager"
            ) as mock_manager:
                # 第一次 load 返回缓存存在
                # 第二次 load 返回 None（更新后过期）
                load_side_effects = [mock_cache_entity, None]
                mock_manager.cache_service.load = AsyncMock(
                    side_effect=load_side_effects
                )
                mock_manager.update_cache_data = AsyncMock()

                with (
                    patch(
                        "app.router.agent_controller_pkg.agent_cache_manage.action_upsert.create_cache_and_build_response",
                        new_callable=AsyncMock,
                    ) as mock_create,
                    patch(
                        "app.router.agent_controller_pkg.agent_cache_manage.action_upsert.StandLogger.info"
                    ) as mock_standard_info,
                    patch(
                        "app.router.agent_controller_pkg.agent_cache_manage.action_upsert.StandLogger.warn"
                    ) as mock_standard_warn,
                ):
                    mock_response = MagicMock()
                    mock_response.cache_id = "new_cache_id"
                    mock_create.return_value = mock_response

                    from app.router.agent_controller_pkg.agent_cache_manage.action_upsert import (
                        handle_upsert,
                    )

                    result = await handle_upsert(
                        request=mock_request,
                        account_id="test_account",
                        account_type="user",
                        agent_id="test_agent",
                        agent_version="v1.0",
                        agent_config=mock_agent_config,
                    )

                    assert result is not None
                    mock_create.assert_called_once()
                    mock_standard_info.assert_called_once()
                    mock_standard_warn.assert_called_once()
