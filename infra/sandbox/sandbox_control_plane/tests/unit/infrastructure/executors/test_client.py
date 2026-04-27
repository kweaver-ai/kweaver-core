"""
执行器 HTTP 客户端单元测试

测试 ExecutorClient 的功能。
"""
import pytest
from unittest.mock import Mock, AsyncMock, patch, MagicMock
import httpx

from src.infrastructure.executors.client import ExecutorClient
from src.infrastructure.executors.errors import (
    ExecutorConnectionError,
    ExecutorTimeoutError,
    ExecutorUnavailableError,
    ExecutorResponseError,
    ExecutorValidationError,
)


class TestExecutorClient:
    """执行器客户端测试"""

    @pytest.fixture
    def client(self):
        """创建执行器客户端"""
        return ExecutorClient(
            timeout=30.0,
            max_retries=3,
            retry_delay=0.1
        )

    @pytest.fixture
    def mock_httpx_client(self):
        """创建模拟 httpx 客户端"""
        mock = Mock()
        mock.post = AsyncMock()
        mock.get = AsyncMock()
        mock.aclose = AsyncMock()
        return mock

    def test_init_default_params(self):
        """测试默认参数初始化"""
        client = ExecutorClient()

        assert client._timeout == 30.0
        assert client._max_retries == 3
        assert client._retry_delay == 0.5

    def test_init_custom_params(self):
        """测试自定义参数初始化"""
        client = ExecutorClient(
            timeout=60.0,
            max_retries=5,
            retry_delay=1.0
        )

        assert client._timeout == 60.0
        assert client._max_retries == 5
        assert client._retry_delay == 1.0

    @pytest.mark.asyncio
    async def test_context_manager(self):
        """测试上下文管理器"""
        async with ExecutorClient() as client:
            assert client._client is not None

        # Client should be closed after exiting context
        # The _client attribute still exists but connection is closed

    @pytest.mark.asyncio
    async def test_context_manager_creates_client(self):
        """测试上下文管理器创建客户端"""
        client = ExecutorClient()
        assert client._client is None

        async with client:
            assert client._client is not None

    @pytest.mark.asyncio
    async def test_close(self, client):
        """测试关闭客户端"""
        # First, get a client instance
        client._get_client()
        assert client._client is not None

        await client.close()
        assert client._client is None

    @pytest.mark.asyncio
    async def test_close_when_no_client(self, client):
        """测试关闭不存在的客户端"""
        # Should not raise error
        await client.close()

    @pytest.mark.asyncio
    async def test_submit_execution_success(self, client, mock_httpx_client):
        """测试成功提交执行请求"""
        client._client = mock_httpx_client

        mock_response = Mock()
        mock_response.status_code = 200
        mock_response.json.return_value = {
            "execution_id": "exec-123",
            "status": "submitted",
            "message": ""
        }
        mock_httpx_client.post.return_value = mock_response

        result = await client.submit_execution(
            executor_url="http://localhost:8080",
            execution_id="exec-123",
            session_id="sess-456",
            code="print('hello')",
            language="python",
            event={},
            timeout=60,
            env_vars={}
        )

        assert result == "exec-123"
        mock_httpx_client.post.assert_called_once()

    @pytest.mark.asyncio
    async def test_submit_execution_validation_error(self, client, mock_httpx_client):
        """测试验证错误（不重试）"""
        client._client = mock_httpx_client

        mock_response = Mock()
        mock_response.status_code = 400
        mock_response.json.return_value = {"errors": ["Invalid code"]}
        mock_httpx_client.post.return_value = mock_response

        with pytest.raises(ExecutorValidationError) as exc_info:
            await client.submit_execution(
                executor_url="http://localhost:8080",
                execution_id="exec-123",
                session_id="sess-456",
                code="print('hello')",
                language="python",
                event={},
                timeout=60,
                env_vars={}
            )

        assert "Invalid code" in str(exc_info.value.validation_errors)
        # Should only call once (no retry for 400)
        mock_httpx_client.post.assert_called_once()

    @pytest.mark.asyncio
    async def test_submit_execution_5xx_retry_then_success(self, client, mock_httpx_client):
        """测试 5xx 错误后重试成功"""
        client._client = mock_httpx_client

        # First call: 500 error
        mock_response_500 = Mock()
        mock_response_500.status_code = 500
        mock_response_500.text = "Internal Server Error"

        # Second call: success
        mock_response_200 = Mock()
        mock_response_200.status_code = 200
        mock_response_200.json.return_value = {
            "execution_id": "exec-123",
            "status": "submitted",
            "message": ""
        }

        mock_httpx_client.post.side_effect = [mock_response_500, mock_response_200]

        result = await client.submit_execution(
            executor_url="http://localhost:8080",
            execution_id="exec-123",
            session_id="sess-456",
            code="print('hello')",
            language="python",
            event={},
            timeout=60,
            env_vars={}
        )

        assert result == "exec-123"
        assert mock_httpx_client.post.call_count == 2

    @pytest.mark.asyncio
    async def test_submit_execution_5xx_max_retries(self, client, mock_httpx_client):
        """测试 5xx 错误达到最大重试次数"""
        client._client = mock_httpx_client

        mock_response = Mock()
        mock_response.status_code = 500
        mock_response.text = "Internal Server Error"
        mock_httpx_client.post.return_value = mock_response

        with pytest.raises(ExecutorResponseError):
            await client.submit_execution(
                executor_url="http://localhost:8080",
                execution_id="exec-123",
                session_id="sess-456",
                code="print('hello')",
                language="python",
                event={},
                timeout=60,
                env_vars={}
            )

        # Should have tried max_retries times
        assert mock_httpx_client.post.call_count == 3

    @pytest.mark.asyncio
    async def test_submit_execution_connection_error_retry(self, client, mock_httpx_client):
        """测试连接错误重试"""
        client._client = mock_httpx_client

        # First two calls: connection error
        # Third call: success
        mock_response = Mock()
        mock_response.status_code = 200
        mock_response.json.return_value = {
            "execution_id": "exec-123",
            "status": "submitted",
            "message": ""
        }

        mock_httpx_client.post.side_effect = [
            httpx.ConnectError("Connection refused"),
            httpx.ConnectError("Connection refused"),
            mock_response
        ]

        result = await client.submit_execution(
            executor_url="http://localhost:8080",
            execution_id="exec-123",
            session_id="sess-456",
            code="print('hello')",
            language="python",
            event={},
            timeout=60,
            env_vars={}
        )

        assert result == "exec-123"
        assert mock_httpx_client.post.call_count == 3

    @pytest.mark.asyncio
    async def test_submit_execution_connection_error_max_retries(self, client, mock_httpx_client):
        """测试连接错误达到最大重试次数"""
        client._client = mock_httpx_client

        mock_httpx_client.post.side_effect = httpx.ConnectError("Connection refused")

        with pytest.raises(ExecutorConnectionError):
            await client.submit_execution(
                executor_url="http://localhost:8080",
                execution_id="exec-123",
                session_id="sess-456",
                code="print('hello')",
                language="python",
                event={},
                timeout=60,
                env_vars={}
            )

        assert mock_httpx_client.post.call_count == 3

    @pytest.mark.asyncio
    async def test_submit_execution_timeout_error(self, client, mock_httpx_client):
        """测试超时错误"""
        client._client = mock_httpx_client

        mock_httpx_client.post.side_effect = httpx.TimeoutException("Timeout")

        with pytest.raises(ExecutorTimeoutError):
            await client.submit_execution(
                executor_url="http://localhost:8080",
                execution_id="exec-123",
                session_id="sess-456",
                code="print('hello')",
                language="python",
                event={},
                timeout=60,
                env_vars={}
            )

    @pytest.mark.asyncio
    async def test_submit_execution_other_error_response(self, client, mock_httpx_client):
        """测试其他错误响应（如 404）"""
        client._client = mock_httpx_client

        mock_response = Mock()
        mock_response.status_code = 404
        mock_response.text = "Not Found"
        mock_httpx_client.post.return_value = mock_response

        with pytest.raises(ExecutorResponseError) as exc_info:
            await client.submit_execution(
                executor_url="http://localhost:8080",
                execution_id="exec-123",
                session_id="sess-456",
                code="print('hello')",
                language="python",
                event={},
                timeout=60,
                env_vars={}
            )

        assert exc_info.value.status_code == 404

    @pytest.mark.asyncio
    async def test_health_check_success(self, client, mock_httpx_client):
        """测试健康检查成功"""
        client._client = mock_httpx_client

        mock_response = Mock()
        mock_response.status_code = 200
        mock_response.json.return_value = {
            "status": "healthy",
            "version": "1.0.0"
        }
        mock_httpx_client.get.return_value = mock_response

        result = await client.health_check("http://localhost:8080")

        assert result.status == "healthy"
        assert result.version == "1.0.0"

    @pytest.mark.asyncio
    async def test_health_check_unhealthy(self, client, mock_httpx_client):
        """测试健康检查不健康"""
        client._client = mock_httpx_client

        mock_response = Mock()
        mock_response.status_code = 503
        mock_httpx_client.get.return_value = mock_response

        with pytest.raises(ExecutorUnavailableError):
            await client.health_check("http://localhost:8080")

    @pytest.mark.asyncio
    async def test_health_check_connection_error(self, client, mock_httpx_client):
        """测试健康检查连接错误"""
        client._client = mock_httpx_client

        mock_httpx_client.get.side_effect = httpx.ConnectError("Connection refused")

        with pytest.raises(ExecutorConnectionError):
            await client.health_check("http://localhost:8080")

    @pytest.mark.asyncio
    async def test_health_check_timeout(self, client, mock_httpx_client):
        """测试健康检查超时"""
        client._client = mock_httpx_client

        mock_httpx_client.get.side_effect = httpx.TimeoutException("Timeout")

        with pytest.raises(ExecutorTimeoutError):
            await client.health_check("http://localhost:8080")

    @pytest.mark.asyncio
    async def test_get_client_creates_if_none(self, client):
        """测试获取客户端时自动创建"""
        assert client._client is None

        c = client._get_client()

        assert c is not None
        assert client._client is not None

    @pytest.mark.asyncio
    async def test_get_client_returns_existing(self, client, mock_httpx_client):
        """测试获取客户端时返回现有实例"""
        client._client = mock_httpx_client

        c = client._get_client()

        assert c is mock_httpx_client
