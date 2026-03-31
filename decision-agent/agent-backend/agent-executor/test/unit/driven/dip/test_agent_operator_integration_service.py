"""单元测试 - driven/dip/agent_operator_integration_service 模块"""

import pytest
from unittest.mock import AsyncMock, MagicMock, patch
from http import HTTPStatus

from app.driven.dip.agent_operator_integration_service import (
    AgentOperatorIntegrationService,
)
from app.common.errors import CodeException


@pytest.fixture
def mock_config():
    """Mock 配置"""
    with patch("app.driven.dip.agent_operator_integration_service.Config") as mock:
        config = MagicMock()
        config.services.agent_operator_integration.host = "localhost"
        config.services.agent_operator_integration.port = 8081
        yield config


@pytest.fixture
def agent_operator_service():
    """创建 AgentOperatorIntegrationService 实例"""
    return AgentOperatorIntegrationService()


class TestAgentOperatorIntegrationServiceInit:
    """测试 AgentOperatorIntegrationService 初始化"""

    def test_init(self, mock_config):
        """测试初始化"""
        with patch(
            "app.driven.dip.agent_operator_integration_service.Config", mock_config
        ):
            service = AgentOperatorIntegrationService()
            assert service._host == "localhost"
            assert service._port == 8081
            assert service._basic_url == "http://localhost:8081"
            assert isinstance(service.headers, dict)

    def test_set_headers(self, agent_operator_service):
        """测试设置headers"""
        test_headers = {"Authorization": "Bearer token123"}
        agent_operator_service.set_headers(test_headers)
        assert agent_operator_service.headers == test_headers


class TestGetToolBoxList:
    """测试 get_tool_box_list 方法"""

    @pytest.mark.asyncio
    async def test_get_tool_box_list_success(self, agent_operator_service):
        """测试获取工具箱列表成功"""
        expected_response = {"data": [{"id": "1", "name": "toolbox1"}]}

        mock_response = AsyncMock()
        mock_response.status = HTTPStatus.OK
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

            result = await agent_operator_service.get_tool_box_list()

            assert result == expected_response

    @pytest.mark.asyncio
    async def test_get_tool_box_list_error_response(self, agent_operator_service):
        """测试获取工具箱列表错误响应"""
        mock_response = AsyncMock()
        mock_response.status = HTTPStatus.INTERNAL_SERVER_ERROR
        mock_response.text = AsyncMock(return_value="Server Error")

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
                await agent_operator_service.get_tool_box_list()


class TestGetToolList:
    """测试 get_tool_list 方法"""

    @pytest.mark.asyncio
    async def test_get_tool_list_success(self, agent_operator_service):
        """测试获取工具列表成功"""
        box_id = "box123"
        expected_response = {"data": [{"id": "tool1", "name": "tool1"}]}

        mock_response = AsyncMock()
        mock_response.status = HTTPStatus.OK
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

            result = await agent_operator_service.get_tool_list(box_id)

            assert result == expected_response

    @pytest.mark.asyncio
    async def test_get_tool_list_error_response(self, agent_operator_service):
        """测试获取工具列表错误响应"""
        box_id = "box123"

        mock_response = AsyncMock()
        mock_response.status = HTTPStatus.NOT_FOUND
        mock_response.text = AsyncMock(return_value="Not Found")

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
                await agent_operator_service.get_tool_list(box_id)


class TestGetToolInfo:
    """测试 get_tool_info 方法"""

    @pytest.mark.asyncio
    async def test_get_tool_info_success(self, agent_operator_service):
        """测试获取工具详情成功"""
        box_id = "box123"
        tool_id = "tool456"
        expected_response = {
            "id": tool_id,
            "name": "Test Tool",
            "description": "A test tool",
        }

        mock_response = AsyncMock()
        mock_response.status = HTTPStatus.OK
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

            result = await agent_operator_service.get_tool_info(box_id, tool_id)

            assert result == expected_response

    @pytest.mark.asyncio
    async def test_get_tool_info_error_response(self, agent_operator_service):
        """测试获取工具详情错误响应"""
        box_id = "box123"
        tool_id = "tool456"

        mock_response = AsyncMock()
        mock_response.status = HTTPStatus.FORBIDDEN
        mock_response.text = AsyncMock(return_value="Forbidden")

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

            result = await agent_operator_service.get_tool_info(box_id, tool_id)

            assert result is None


class TestGetMcpTools:
    """测试 get_mcp_tools 方法"""

    @pytest.mark.asyncio
    async def test_get_mcp_tools_success(self, agent_operator_service):
        """测试获取MCP工具成功"""
        mcp_server_id = "mcp_server_123"
        expected_response = {
            "tools": [
                {"name": "tool1", "description": "Tool 1"},
                {"name": "tool2", "description": "Tool 2"},
            ]
        }

        mock_response = AsyncMock()
        mock_response.status = HTTPStatus.OK
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

            result = await agent_operator_service.get_mcp_tools(mcp_server_id)

            assert result == expected_response

    @pytest.mark.asyncio
    async def test_get_mcp_tools_error_response(self, agent_operator_service):
        """测试获取MCP工具错误响应"""
        mcp_server_id = "mcp_server_123"

        mock_response = AsyncMock()
        mock_response.status = HTTPStatus.SERVICE_UNAVAILABLE
        mock_response.text = AsyncMock(return_value="Service Unavailable")

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
                await agent_operator_service.get_mcp_tools(mcp_server_id)


class TestGetMockToolInfo:
    """测试 get_mock_tool_info 方法"""

    def test_get_mock_tool_info(self, agent_operator_service):
        """测试获取mock工具信息"""

        # 预期返回的工具信息
        tool_info = {"name": "mock_tool", "version": "1.0"}

        # Mock open 来避免实际文件读取
        mock_file = MagicMock()

        # Mock json.load 来直接返回预期值
        with patch("builtins.open", return_value=mock_file):
            with patch("json.load", return_value=tool_info):
                result = agent_operator_service.get_mock_tool_info()
                assert result == tool_info
