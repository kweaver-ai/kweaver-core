# -*- coding: utf-8 -*-
"""单元测试 - agent_cache_manage/common 模块"""

import pytest
from unittest.mock import MagicMock, AsyncMock, patch
import json


class TestPrepareAgentConfig:
    """测试 prepare_agent_config 函数"""

    @pytest.fixture
    def mock_req(self):
        """创建 mock AgentCacheManageReq"""
        req = MagicMock()
        req.agent_id = "test_agent"
        return req

    @pytest.mark.asyncio
    async def test_prepare_agent_config_success(self, mock_req):
        """测试成功准备Agent配置"""
        config_dict = {
            "agent_id": "test_agent",
            "agent_run_id": None,
        }
        config_data = {"config": json.dumps(config_dict)}

        with patch(
            "app.router.agent_controller_pkg.agent_cache_manage.common.agent_factory_service"
        ) as mock_service:
            mock_service.get_agent_config = AsyncMock(return_value=config_data)
            mock_service.check_agent_permission = AsyncMock(return_value=True)

            with patch(
                "app.router.agent_controller_pkg.agent_cache_manage.common.AgentConfigVo"
            ) as mock_config_vo:
                mock_config = MagicMock()
                mock_config.agent_id = "test_agent"
                mock_config_vo.return_value = mock_config

                from app.router.agent_controller_pkg.agent_cache_manage.common import (
                    prepare_agent_config,
                )

                result = await prepare_agent_config(mock_req, "user123", "standard")

                assert result is not None
                mock_service.get_agent_config.assert_called_once_with("test_agent")
                mock_service.check_agent_permission.assert_called_once()

    @pytest.mark.asyncio
    async def test_prepare_agent_config_permission_denied(self, mock_req):
        """测试权限检查失败"""
        config_dict = {"agent_id": "test_agent"}
        config_data = {"config": json.dumps(config_dict)}

        with patch(
            "app.router.agent_controller_pkg.agent_cache_manage.common.agent_factory_service"
        ) as mock_service:
            mock_service.get_agent_config = AsyncMock(return_value=config_data)
            mock_service.check_agent_permission = AsyncMock(return_value=False)

            with patch(
                "app.router.agent_controller_pkg.agent_cache_manage.common.AgentConfigVo"
            ) as mock_config_vo:
                mock_config = MagicMock()
                mock_config.agent_id = "test_agent"
                mock_config_vo.return_value = mock_config

                from app.router.agent_controller_pkg.agent_cache_manage.common import (
                    prepare_agent_config,
                )
                from app.common.errors import AgentPermissionException

                with pytest.raises(AgentPermissionException):
                    await prepare_agent_config(mock_req, "user123", "standard")


class TestBuildCacheIdVo:
    """测试 build_cache_id_vo 函数"""

    def test_build_cache_id_vo(self):
        """测试构建缓存ID VO"""
        mock_agent_config = MagicMock()
        mock_agent_config.get_config_last_set_timestamp.return_value = 1234567890

        with patch(
            "app.router.agent_controller_pkg.agent_cache_manage.common.AgentCacheIdVO"
        ) as mock_vo_class:
            mock_vo = MagicMock()
            mock_vo_class.return_value = mock_vo

            from app.router.agent_controller_pkg.agent_cache_manage.common import (
                build_cache_id_vo,
            )

            result = build_cache_id_vo(
                account_id="user123",
                account_type="standard",
                agent_id="test_agent",
                agent_version="v1.0",
                agent_config=mock_agent_config,
            )

            assert result is not None
            mock_vo_class.assert_called_once()


class TestGetCacheDataForDebugMode:
    """测试 get_cache_data_for_debug_mode 函数"""

    def test_get_cache_data_debug_mode(self):
        """测试debug模式下获取缓存数据"""
        mock_cache_entity = MagicMock()
        mock_cache_entity.cache_data = MagicMock()

        with patch(
            "app.router.agent_controller_pkg.agent_cache_manage.common.Config"
        ) as mock_config:
            mock_config.is_debug_mode.return_value = True

            with patch(
                "app.router.agent_controller_pkg.agent_cache_manage.common.asdict",
                return_value={"key": "value"},
            ):
                from app.router.agent_controller_pkg.agent_cache_manage.common import (
                    get_cache_data_for_debug_mode,
                )

                result = get_cache_data_for_debug_mode(mock_cache_entity)

                assert result == {"key": "value"}

    def test_get_cache_data_non_debug_mode(self):
        """测试非debug模式"""
        mock_cache_entity = MagicMock()

        with patch(
            "app.router.agent_controller_pkg.agent_cache_manage.common.Config"
        ) as mock_config:
            mock_config.is_debug_mode.return_value = False

            from app.router.agent_controller_pkg.agent_cache_manage.common import (
                get_cache_data_for_debug_mode,
            )

            result = get_cache_data_for_debug_mode(mock_cache_entity)

            assert result == {}

    def test_get_cache_data_no_cache_data(self):
        """测试缓存数据为空"""
        mock_cache_entity = MagicMock()
        mock_cache_entity.cache_data = None

        with patch(
            "app.router.agent_controller_pkg.agent_cache_manage.common.Config"
        ) as mock_config:
            mock_config.is_debug_mode.return_value = True

            from app.router.agent_controller_pkg.agent_cache_manage.common import (
                get_cache_data_for_debug_mode,
            )

            result = get_cache_data_for_debug_mode(mock_cache_entity)

            assert result == {}


class TestCreateCacheAndBuildResponse:
    """测试 create_cache_and_build_response 函数"""

    @pytest.fixture
    def mock_request(self):
        """创建 mock Request"""
        request = MagicMock()
        request.headers = {"Authorization": "Bearer token"}
        return request

    @pytest.mark.asyncio
    async def test_create_cache_success(self, mock_request):
        """测试成功创建缓存"""
        mock_cache_id_vo = MagicMock()
        mock_cache_id_vo.get_cache_id.return_value = "test_cache_id"

        mock_cache_entity = MagicMock()
        mock_cache_entity.cache_id_vo = mock_cache_id_vo
        mock_cache_entity.created_at = "2024-01-01T00:00:00"
        mock_cache_entity.cache_data = MagicMock()

        mock_agent_config = MagicMock()

        with patch(
            "app.router.agent_controller_pkg.agent_cache_manage.common.cache_manager"
        ) as mock_manager:
            mock_manager.create_cache = AsyncMock(return_value=mock_cache_entity)
            mock_manager.cache_service.get_ttl = AsyncMock(return_value=3600)

            with patch(
                "app.router.agent_controller_pkg.agent_cache_manage.common.get_cache_data_for_debug_mode",
                return_value={},
            ):
                with patch(
                    "app.router.agent_controller_pkg.agent_cache_manage.common.AgentCacheManageRes"
                ) as mock_res:
                    mock_response = MagicMock()
                    mock_res.return_value = mock_response

                    from app.router.agent_controller_pkg.agent_cache_manage.common import (
                        create_cache_and_build_response,
                    )

                    result = await create_cache_and_build_response(
                        request=mock_request,
                        account_id="user123",
                        account_type="standard",
                        agent_id="test_agent",
                        agent_version="v1.0",
                        agent_config=mock_agent_config,
                    )

                    assert result is not None
                    mock_manager.create_cache.assert_called_once()

    @pytest.mark.asyncio
    async def test_create_cache_ttl_invalid(self, mock_request):
        """测试TTL无效时抛出异常"""
        mock_cache_id_vo = MagicMock()
        mock_cache_id_vo.get_cache_id.return_value = "test_cache_id"

        mock_cache_entity = MagicMock()
        mock_cache_entity.cache_id_vo = mock_cache_id_vo

        mock_agent_config = MagicMock()

        with patch(
            "app.router.agent_controller_pkg.agent_cache_manage.common.cache_manager"
        ) as mock_manager:
            mock_manager.create_cache = AsyncMock(return_value=mock_cache_entity)
            mock_manager.cache_service.get_ttl = AsyncMock(return_value=-1)

            from app.router.agent_controller_pkg.agent_cache_manage.common import (
                create_cache_and_build_response,
            )
            from app.common.errors import CodeException

            with pytest.raises(CodeException):
                await create_cache_and_build_response(
                    request=mock_request,
                    account_id="user123",
                    account_type="standard",
                    agent_id="test_agent",
                    agent_version="v1.0",
                    agent_config=mock_agent_config,
                )
