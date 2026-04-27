"""单元测试 - driven/dip/agent_factory_service 模块"""

import pytest
from unittest.mock import AsyncMock, MagicMock, patch

from app.driven.dip.agent_factory_service import AgentFactoryService
from app.common.errors import CodeException


@pytest.fixture
def mock_config():
    """Mock 配置"""
    with patch("app.driven.dip.agent_factory_service.Config") as mock:
        config = MagicMock()
        config.services.agent_factory.host = "localhost"
        config.services.agent_factory.port = 8081
        yield config


@pytest.fixture
def agent_factory_service():
    """创建 AgentFactoryService 实例"""
    return AgentFactoryService()


class TestAgentFactoryServiceInit:
    """测试 AgentFactoryService 初始化"""

    def test_init(self, mock_config):
        """测试初始化"""
        with patch("app.driven.dip.agent_factory_service.Config", mock_config):
            service = AgentFactoryService()
            assert service._host == "localhost"
            assert service._port == 8081
            assert service._basic_url == "http://localhost:8081"


class TestGetAgentConfig:
    """测试 get_agent_config 方法"""

    @pytest.mark.asyncio
    async def test_get_agent_config_success(self, agent_factory_service):
        """测试成功获取 agent 配置"""
        agent_key = "test_agent_key"

        expected_response = {"key": "test_key_value"}

        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.json = AsyncMock(return_value=expected_response)

        mock_get_context = AsyncMock()
        mock_get_context.__aenter__.return_value = mock_response
        mock_get_context.__aexit__.return_value = None

        with patch("aiohttp.ClientSession") as mock_session_class:
            mock_session_instance = MagicMock()
            mock_session_instance.get = MagicMock(return_value=mock_get_context)
            mock_session_instance.__aenter__ = AsyncMock(
                return_value=mock_session_instance
            )
            mock_session_instance.__aexit__ = AsyncMock(return_value=None)
            mock_session_class.return_value = mock_session_instance

            result = await agent_factory_service.get_agent_config_by_key(agent_key)

            assert result == expected_response

    @pytest.mark.asyncio
    async def test_get_agent_config_404_error(self, agent_factory_service):
        """测试 404 错误响应"""
        agent_key = "test_agent_key"

        mock_response = AsyncMock()
        mock_response.status = 404
        mock_response.text = AsyncMock(return_value="Agent not found")

        mock_get_context = AsyncMock()
        mock_get_context.__aenter__.return_value = mock_response
        mock_get_context.__aexit__.return_value = None

        with patch("aiohttp.ClientSession") as mock_session_class:
            mock_session_instance = MagicMock()
            mock_session_instance.get = MagicMock(return_value=mock_get_context)
            mock_session_instance.__aenter__ = AsyncMock(
                return_value=mock_session_instance
            )
            mock_session_instance.__aexit__ = AsyncMock(return_value=None)
            mock_session_class.return_value = mock_session_instance

            with pytest.raises(CodeException):
                await agent_factory_service.get_agent_config_by_key(agent_key)


class TestGetAgentConfigByKey:
    """测试 get_agent_config_by_key 方法"""

    @pytest.mark.asyncio
    async def test_get_agent_config_by_key_success(self, agent_factory_service):
        """测试成功通过 key 获取 agent 配置"""
        agent_key = "test_key_123"

        expected_response = {"key": "test_key_value"}

        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.json = AsyncMock(return_value=expected_response)

        mock_get_context = AsyncMock()
        mock_get_context.__aenter__.return_value = mock_response
        mock_get_context.__aexit__.return_value = None

        with patch("aiohttp.ClientSession") as mock_session_class:
            mock_session_instance = MagicMock()
            mock_session_instance.get = MagicMock(return_value=mock_get_context)
            mock_session_instance.__aenter__ = AsyncMock(
                return_value=mock_session_instance
            )
            mock_session_instance.__aexit__ = AsyncMock(return_value=None)
            mock_session_class.return_value = mock_session_instance

            result = await agent_factory_service.get_agent_config_by_key(agent_key)

            assert result == expected_response

    @pytest.mark.asyncio
    async def test_get_agent_config_by_key_404_error(self, agent_factory_service):
        """测试 404 错误响应"""
        agent_key = "test_key_123"

        mock_response = AsyncMock()
        mock_response.status = 404
        mock_response.text = AsyncMock(return_value="Agent not found")

        mock_get_context = AsyncMock()
        mock_get_context.__aenter__.return_value = mock_response
        mock_get_context.__aexit__.return_value = None

        with patch("aiohttp.ClientSession") as mock_session_class:
            mock_session_instance = MagicMock()
            mock_session_instance.get = MagicMock(return_value=mock_get_context)
            mock_session_instance.__aenter__ = AsyncMock(
                return_value=mock_session_instance
            )
            mock_session_instance.__aexit__ = AsyncMock(return_value=None)
            mock_session_class.return_value = mock_session_instance

            with pytest.raises(CodeException):
                await agent_factory_service.get_agent_config_by_key(agent_key)


class TestCheckAgentPermission:
    """测试 check_agent_permission 方法"""

    @pytest.mark.asyncio
    async def test_check_agent_permission_user_allowed(self, agent_factory_service):
        """测试用户权限检查-允许访问"""
        agent_id = "agent123"
        user_id = "user456"
        visitor_type = "user"

        expected_response = {"is_allowed": True}

        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.json = AsyncMock(return_value=expected_response)

        mock_post_context = AsyncMock()
        mock_post_context.__aenter__.return_value = mock_response
        mock_post_context.__aexit__.return_value = None

        with patch("aiohttp.ClientSession") as mock_session_class:
            mock_session_instance = MagicMock()
            mock_session_instance.post = MagicMock(return_value=mock_post_context)
            mock_session_instance.__aenter__ = AsyncMock(
                return_value=mock_session_instance
            )
            mock_session_instance.__aexit__ = AsyncMock(return_value=None)
            mock_session_class.return_value = mock_session_instance

            with patch("app.driven.dip.agent_factory_service.Config") as mock_config:
                mock_config.local_dev.is_skip_pms_check = False

                result = await agent_factory_service.check_agent_permission(
                    agent_id, user_id, visitor_type
                )

            assert result is True
            mock_session_instance.post.assert_called_once()

    @pytest.mark.asyncio
    async def test_check_agent_permission_app_allowed(self, agent_factory_service):
        """测试 app 类型的访客检查"""
        agent_id = "agent123"
        user_id = "user456"
        visitor_type = "app"

        expected_response = {"is_allowed": True}

        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.json = AsyncMock(return_value=expected_response)

        mock_post_context = AsyncMock()
        mock_post_context.__aenter__.return_value = mock_response
        mock_post_context.__aexit__.return_value = None

        with patch("aiohttp.ClientSession") as mock_session_class:
            mock_session_instance = MagicMock()
            mock_session_instance.post = MagicMock(return_value=mock_post_context)
            mock_session_instance.__aenter__ = AsyncMock(
                return_value=mock_session_instance
            )
            mock_session_instance.__aexit__ = AsyncMock(return_value=None)
            mock_session_class.return_value = mock_session_instance

            with patch("app.driven.dip.agent_factory_service.Config") as mock_config:
                mock_config.local_dev.is_skip_pms_check = False

                result = await agent_factory_service.check_agent_permission(
                    agent_id, user_id, visitor_type
                )

            assert result is True
            mock_session_instance.post.assert_called_once()

    @pytest.mark.asyncio
    async def test_check_agent_permission_denied(self, agent_factory_service):
        """测试权限检查-拒绝访问"""
        agent_id = "agent123"
        user_id = "user456"
        visitor_type = "user"

        expected_response = {"is_allowed": False}

        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.json = AsyncMock(return_value=expected_response)

        mock_post_context = AsyncMock()
        mock_post_context.__aenter__.return_value = mock_response
        mock_post_context.__aexit__.return_value = None

        with patch("aiohttp.ClientSession") as mock_session_class:
            mock_session_instance = MagicMock()
            mock_session_instance.post = MagicMock(return_value=mock_post_context)
            mock_session_instance.__aenter__ = AsyncMock(
                return_value=mock_session_instance
            )
            mock_session_instance.__aexit__ = AsyncMock(return_value=None)
            mock_session_class.return_value = mock_session_instance

            with patch("app.driven.dip.agent_factory_service.Config") as mock_config:
                mock_config.local_dev.is_skip_pms_check = False

                result = await agent_factory_service.check_agent_permission(
                    agent_id, user_id, visitor_type
                )

            assert result is False
            mock_session_instance.post.assert_called_once()

    @pytest.mark.asyncio
    async def test_check_agent_permission_skip_check(self, agent_factory_service):
        """测试跳过权限检查"""
        agent_id = "agent123"
        user_id = "user456"
        visitor_type = "user"

        expected_response = {"is_allowed": True}

        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.json = AsyncMock(return_value=expected_response)

        mock_post_context = AsyncMock()
        mock_post_context.__aenter__.return_value = mock_response
        mock_post_context.__aexit__.return_value = None

        with patch("aiohttp.ClientSession") as mock_session_class:
            mock_session_instance = MagicMock()
            mock_session_instance.post = MagicMock(return_value=mock_post_context)
            mock_session_instance.__aenter__ = AsyncMock(
                return_value=mock_session_instance
            )
            mock_session_instance.__aexit__ = AsyncMock(return_value=None)
            mock_session_class.return_value = mock_session_instance

            # Mock Config to return is_skip_pms_check = True
            with patch("app.driven.dip.agent_factory_service.Config") as mock_config:
                mock_config.local_dev.is_skip_pms_check = True

                result = await agent_factory_service.check_agent_permission(
                    agent_id, user_id, visitor_type
                )

            assert result is True
            mock_session_instance.post.assert_not_called()

    @pytest.mark.asyncio
    async def test_check_agent_permission_error_response(self, agent_factory_service):
        """测试权限检查错误响应"""
        agent_id = "agent123"
        user_id = "user456"
        visitor_type = "user"

        mock_response = AsyncMock()
        mock_response.status = 500
        mock_response.text = AsyncMock(return_value="Internal Server Error")

        mock_post_context = AsyncMock()
        mock_post_context.__aenter__.return_value = mock_response
        mock_post_context.__aexit__.return_value = None

        with patch("aiohttp.ClientSession") as mock_session_class:
            mock_session_instance = MagicMock()
            mock_session_instance.post = MagicMock(return_value=mock_post_context)
            mock_session_instance.__aenter__ = AsyncMock(
                return_value=mock_session_instance
            )
            mock_session_instance.__aexit__ = AsyncMock(return_value=None)
            mock_session_class.return_value = mock_session_instance

            with patch("app.driven.dip.agent_factory_service.Config") as mock_config:
                mock_config.local_dev.is_skip_pms_check = False

                with pytest.raises(CodeException):
                    await agent_factory_service.check_agent_permission(
                        agent_id, user_id, visitor_type
                    )

            mock_session_instance.post.assert_called_once()


class TestGetAgentConfigByAgentIdAndVersion:
    """测试 get_agent_config_by_agent_id_and_version 方法"""

    @pytest.mark.asyncio
    async def test_get_agent_config_by_agent_id_and_version_success(
        self, agent_factory_service
    ):
        """测试成功获取 agent 配置"""
        agent_id = "agent123"
        version = "v1.0"

        expected_response = {"config": "test_config"}

        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.json = AsyncMock(return_value=expected_response)

        mock_get_context = AsyncMock()
        mock_get_context.__aenter__.return_value = mock_response
        mock_get_context.__aexit__.return_value = None

        with patch("aiohttp.ClientSession") as mock_session_class:
            mock_session_instance = MagicMock()
            mock_session_instance.get = MagicMock(return_value=mock_get_context)
            mock_session_instance.__aenter__ = AsyncMock(
                return_value=mock_session_instance
            )
            mock_session_instance.__aexit__ = AsyncMock(return_value=None)
            mock_session_class.return_value = mock_session_instance

            result = (
                await agent_factory_service.get_agent_config_by_agent_id_and_version(
                    agent_id, version
                )
            )

            assert result == expected_response

    @pytest.mark.asyncio
    async def test_get_agent_config_by_agent_id_and_version_404_error(
        self, agent_factory_service
    ):
        """测试 404 错误响应"""
        agent_id = "agent123"
        version = "v1.0"

        mock_response = AsyncMock()
        mock_response.status = 404
        mock_response.text = AsyncMock(return_value="Agent not found")

        mock_get_context = AsyncMock()
        mock_get_context.__aenter__.return_value = mock_response
        mock_get_context.__aexit__.return_value = None

        with patch("aiohttp.ClientSession") as mock_session_class:
            mock_session_instance = MagicMock()
            mock_session_instance.get = MagicMock(return_value=mock_get_context)
            mock_session_instance.__aenter__ = AsyncMock(
                return_value=mock_session_instance
            )
            mock_session_instance.__aexit__ = AsyncMock(return_value=None)
            mock_session_class.return_value = mock_session_instance

            with pytest.raises(CodeException):
                await agent_factory_service.get_agent_config_by_agent_id_and_version(
                    agent_id, version
                )
