# -*- coding: utf-8 -*-
"""单元测试 - agent_cache_manage/manage_agent_cache 模块"""

import importlib

import pytest
from unittest.mock import MagicMock, AsyncMock, patch


class TestManageAgentCache:
    """测试 manage_agent_cache 函数"""

    @pytest.fixture
    def mock_request(self):
        """创建 mock Request"""
        request = MagicMock()
        request.headers = {}
        return request

    @pytest.fixture
    def mock_req_upsert(self):
        """创建 mock AgentCacheManageReq (upsert)"""
        from app.router.agent_controller_pkg.rdto.v1.req.agent_cache import (
            AgentCacheAction,
        )

        req = MagicMock()
        req.agent_id = "test_agent"
        req.agent_version = "v1.0"
        req.action = AgentCacheAction.UPSERT
        return req

    @pytest.fixture
    def mock_req_get_info(self):
        """创建 mock AgentCacheManageReq (get_info)"""
        from app.router.agent_controller_pkg.rdto.v1.req.agent_cache import (
            AgentCacheAction,
        )

        req = MagicMock()
        req.agent_id = "test_agent"
        req.agent_version = "v1.0"
        req.action = AgentCacheAction.GET_INFO
        return req

    @pytest.mark.asyncio
    async def test_manage_agent_cache_upsert(self, mock_request, mock_req_upsert):
        """测试 upsert action"""
        mock_agent_config = MagicMock()
        mock_response = MagicMock()
        mock_response.cache_id = "test_cache_id"
        module = importlib.import_module(
            "app.router.agent_controller_pkg.agent_cache_manage.manage_agent_cache"
        )

        with patch.object(module, "agent_factory_service") as mock_service:
            mock_service.set_headers = MagicMock()
            with (
                patch.object(
                    module,
                    "prepare_agent_config",
                    new_callable=AsyncMock,
                    return_value=mock_agent_config,
                ),
                patch.object(
                    module,
                    "handle_upsert",
                    new_callable=AsyncMock,
                    return_value=mock_response,
                ),
            ):
                result = await module.manage_agent_cache(
                    request=mock_request,
                    req=mock_req_upsert,
                    account_id="user123",
                    account_type="standard",
                    biz_domain_id="domain123",
                )

                assert result is not None
                assert result.cache_id == "test_cache_id"

    @pytest.mark.asyncio
    async def test_manage_agent_cache_get_info(self, mock_request, mock_req_get_info):
        """测试 get_info action"""
        mock_agent_config = MagicMock()
        mock_response = MagicMock()
        mock_response.cache_id = "test_cache_id"
        module = importlib.import_module(
            "app.router.agent_controller_pkg.agent_cache_manage.manage_agent_cache"
        )

        with patch.object(module, "agent_factory_service") as mock_service:
            mock_service.set_headers = MagicMock()
            with (
                patch.object(
                    module,
                    "prepare_agent_config",
                    new_callable=AsyncMock,
                    return_value=mock_agent_config,
                ),
                patch.object(
                    module,
                    "handle_get_info",
                    new_callable=AsyncMock,
                    return_value=mock_response,
                ),
            ):
                result = await module.manage_agent_cache(
                    request=mock_request,
                    req=mock_req_get_info,
                    account_id="user123",
                    account_type="standard",
                    biz_domain_id="domain123",
                )

                assert result is not None

    @pytest.mark.asyncio
    async def test_manage_agent_cache_get_info_not_found(
        self, mock_request, mock_req_get_info
    ):
        """测试 get_info action 缓存不存在"""
        mock_agent_config = MagicMock()
        module = importlib.import_module(
            "app.router.agent_controller_pkg.agent_cache_manage.manage_agent_cache"
        )

        with patch.object(module, "agent_factory_service") as mock_service:
            mock_service.set_headers = MagicMock()
            with (
                patch.object(
                    module,
                    "prepare_agent_config",
                    new_callable=AsyncMock,
                    return_value=mock_agent_config,
                ),
                patch.object(
                    module,
                    "handle_get_info",
                    new_callable=AsyncMock,
                    return_value=None,
                ),
            ):
                result = await module.manage_agent_cache(
                    request=mock_request,
                    req=mock_req_get_info,
                    account_id="user123",
                    account_type="standard",
                    biz_domain_id="domain123",
                )

                assert result is None

    @pytest.mark.asyncio
    async def test_manage_agent_cache_exception(self, mock_request, mock_req_upsert):
        """测试异常处理"""
        module = importlib.import_module(
            "app.router.agent_controller_pkg.agent_cache_manage.manage_agent_cache"
        )

        with patch.object(module, "agent_factory_service") as mock_service:
            mock_service.set_headers = MagicMock()

            with (
                patch.object(
                    module,
                    "prepare_agent_config",
                    new_callable=AsyncMock,
                    side_effect=Exception("Test error"),
                ),
                patch.object(module.StandLogger, "error") as mock_standard_error,
                patch.object(module, "o11y_logger") as mock_o11y_logger,
            ):
                with pytest.raises(Exception) as exc_info:
                    await module.manage_agent_cache(
                        request=mock_request,
                        req=mock_req_upsert,
                        account_id="user123",
                        account_type="standard",
                        biz_domain_id="domain123",
                    )

                assert "Test error" in str(exc_info.value)
                mock_standard_error.assert_called_once()
                mock_o11y_logger().error.assert_called_once()
