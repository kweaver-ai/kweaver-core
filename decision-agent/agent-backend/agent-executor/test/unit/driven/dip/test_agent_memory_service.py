"""单元测试 - driven/dip/agent_memory_service 模块"""

import pytest
from unittest.mock import AsyncMock, MagicMock, patch
import asyncio
from http import HTTPStatus

from app.driven.dip.agent_memory_service import AgentMemoryService
from app.common.errors import CodeException


@pytest.fixture
def mock_config():
    """Mock 配置"""
    with patch("app.driven.dip.agent_memory_service.Config") as mock:
        config = MagicMock()
        config.services.agent_memory.host = "localhost"
        config.services.agent_memory.port = 8082
        yield config


@pytest.fixture
def agent_memory_service():
    """创建 AgentMemoryService 实例"""
    return AgentMemoryService()


class TestAgentMemoryServiceInit:
    """测试 AgentMemoryService 初始化"""

    def test_init(self, mock_config):
        """测试初始化"""
        with patch("app.driven.dip.agent_memory_service.Config", mock_config):
            service = AgentMemoryService()
            assert service._host == "localhost"
            assert service._port == 8082
            assert service._basic_url == "http://localhost:8082"


class TestBuildMemory:
    """测试 build_memory 方法"""

    @pytest.mark.asyncio
    async def test_build_memory_success(self, agent_memory_service):
        """测试构建记忆成功"""
        messages = [{"role": "user", "content": "test message"}]

        mock_response = AsyncMock()
        mock_response.status = HTTPStatus.NO_CONTENT
        mock_response.text = AsyncMock(return_value="")

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

            # 不应该抛出异常
            await agent_memory_service.build_memory(messages)

    @pytest.mark.asyncio
    async def test_build_memory_with_user_id(self, agent_memory_service):
        """测试带user_id的构建记忆"""
        messages = [{"role": "user", "content": "test"}]
        user_id = "user123"

        mock_response = AsyncMock()
        mock_response.status = HTTPStatus.NO_CONTENT

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

            await agent_memory_service.build_memory(messages, user_id=user_id)

    @pytest.mark.asyncio
    async def test_build_memory_with_all_parameters(self, agent_memory_service):
        """测试带所有参数的构建记忆"""
        messages = [{"role": "user", "content": "test"}]
        user_id = "user123"
        agent_id = "agent456"
        run_id = "run789"
        metadata = {"key": "value"}
        headers = {"Authorization": "Bearer token"}

        mock_response = AsyncMock()
        mock_response.status = HTTPStatus.NO_CONTENT

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

            await agent_memory_service.build_memory(
                messages=messages,
                user_id=user_id,
                agent_id=agent_id,
                run_id=run_id,
                metadata=metadata,
                headers=headers,
            )

    @pytest.mark.asyncio
    async def test_build_memory_error_response(self, agent_memory_service):
        """测试错误响应"""
        messages = [{"role": "user", "content": "test"}]

        mock_response = AsyncMock()
        mock_response.status = HTTPStatus.INTERNAL_SERVER_ERROR
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

            with pytest.raises(CodeException):
                await agent_memory_service.build_memory(messages)

    @pytest.mark.asyncio
    async def test_build_memory_timeout_error(self, agent_memory_service):
        """测试超时错误"""
        messages = [{"role": "user", "content": "test"}]

        mock_post_context = AsyncMock()
        mock_post_context.__aenter__.side_effect = asyncio.TimeoutError(
            "Request timeout"
        )
        mock_post_context.__aexit__.return_value = None

        with patch("aiohttp.ClientSession") as mock_session_class:
            mock_session_instance = MagicMock()
            mock_session_instance.post = MagicMock(return_value=mock_post_context)
            mock_session_instance.__aenter__ = AsyncMock(
                return_value=mock_session_instance
            )
            mock_session_instance.__aexit__ = AsyncMock(return_value=None)
            mock_session_class.return_value = mock_session_instance

            with pytest.raises(CodeException):
                await agent_memory_service.build_memory(messages)

    @pytest.mark.asyncio
    async def test_build_memory_client_error(self, agent_memory_service):
        """测试客户端错误"""
        messages = [{"role": "user", "content": "test"}]

        mock_post_context = AsyncMock()
        mock_post_context.__aenter__.side_effect = Exception("Connection error")
        mock_post_context.__aexit__.return_value = None

        with patch("aiohttp.ClientSession") as mock_session_class:
            mock_session_instance = MagicMock()
            mock_session_instance.post = MagicMock(return_value=mock_post_context)
            mock_session_instance.__aenter__ = AsyncMock(
                return_value=mock_session_instance
            )
            mock_session_instance.__aexit__ = AsyncMock(return_value=None)
            mock_session_class.return_value = mock_session_instance

            with pytest.raises(CodeException) as exc_info:
                await agent_memory_service.build_memory(messages)
                # 验证异常消息包含 "client error"
                assert "client error" in str(exc_info.value)

    @pytest.mark.asyncio
    async def test_build_memory_unexpected_exception(self, agent_memory_service):
        """测试未预期的异常"""
        messages = [{"role": "user", "content": "test"}]

        mock_post_context = AsyncMock()
        mock_post_context.__aenter__.side_effect = Exception("Unexpected error")
        mock_post_context.__aexit__.return_value = None

        with patch("aiohttp.ClientSession") as mock_session_class:
            mock_session_instance = MagicMock()
            mock_session_instance.post = MagicMock(return_value=mock_post_context)
            mock_session_instance.__aenter__ = AsyncMock(
                return_value=mock_session_instance
            )
            mock_session_instance.__aexit__ = AsyncMock(return_value=None)
            mock_session_class.return_value = mock_session_instance

            with pytest.raises(CodeException) as exc_info:
                await agent_memory_service.build_memory(messages)
                # 验证异常消息包含 "unexpected error"
                assert "unexpected error" in str(exc_info.value)
