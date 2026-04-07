"""
日志中间件单元测试

测试 RequestLoggingMiddleware 的功能。
"""
import pytest
from unittest.mock import Mock, AsyncMock, patch, MagicMock
import uuid

from fastapi import Request, Response
from starlette.types import ASGIApp, Receive, Scope, Send, Message

from src.interfaces.rest.middleware.logging_middleware import RequestLoggingMiddleware


class TestRequestLoggingMiddleware:
    """日志中间件测试"""

    @pytest.fixture
    def mock_app(self):
        """模拟 ASGI 应用"""
        return Mock()

    @pytest.fixture
    def middleware(self, mock_app):
        """创建中间件"""
        return RequestLoggingMiddleware(mock_app)

    @pytest.fixture
    def mock_request(self):
        """创建模拟请求"""
        request = MagicMock(spec=Request)
        request.method = "GET"
        request.url = MagicMock()
        request.url.path = "/api/v1/sessions"
        request.client = MagicMock()
        request.client.host = "127.0.0.1"
        request.state = MagicMock()
        return request

    @pytest.fixture
    def mock_response(self):
        """创建模拟响应"""
        response = MagicMock(spec=Response)
        response.status_code = 200
        response.headers = {}
        return response

    @pytest.mark.asyncio
    async def test_dispatch_generates_request_id(self, middleware, mock_request, mock_response):
        """测试生成请求 ID"""
        async def call_next(request):
            return mock_response

        with patch('src.interfaces.rest.middleware.logging_middleware.bind_context'):
            with patch('src.interfaces.rest.middleware.logging_middleware.clear_context'):
                response = await middleware.dispatch(mock_request, call_next)

                # Check that request_id was set
                assert hasattr(mock_request.state, 'request_id')
                # Verify it's a valid UUID
                uuid.UUID(mock_request.state.request_id)

    @pytest.mark.asyncio
    async def test_dispatch_binds_context(self, middleware, mock_request, mock_response):
        """测试绑定上下文"""
        async def call_next(request):
            return mock_response

        with patch('src.interfaces.rest.middleware.logging_middleware.bind_context') as mock_bind:
            with patch('src.interfaces.rest.middleware.logging_middleware.clear_context'):
                await middleware.dispatch(mock_request, call_next)

                # Check bind_context was called with correct parameters
                mock_bind.assert_called_once()
                call_kwargs = mock_bind.call_args[1]
                assert 'request_id' in call_kwargs
                assert call_kwargs['method'] == 'GET'
                assert call_kwargs['path'] == '/api/v1/sessions'

    @pytest.mark.asyncio
    async def test_dispatch_adds_request_id_header(self, middleware, mock_request, mock_response):
        """测试添加请求 ID 响应头"""
        async def call_next(request):
            return mock_response

        with patch('src.interfaces.rest.middleware.logging_middleware.bind_context'):
            with patch('src.interfaces.rest.middleware.logging_middleware.clear_context'):
                response = await middleware.dispatch(mock_request, call_next)

                # Check X-Request-ID header was added
                assert 'X-Request-ID' in response.headers

    @pytest.mark.asyncio
    async def test_dispatch_adds_process_time_header(self, middleware, mock_request, mock_response):
        """测试添加处理时间响应头"""
        async def call_next(request):
            return mock_response

        with patch('src.interfaces.rest.middleware.logging_middleware.bind_context'):
            with patch('src.interfaces.rest.middleware.logging_middleware.clear_context'):
                response = await middleware.dispatch(mock_request, call_next)

                # Check X-Process-Time header was added
                assert 'X-Process-Time' in response.headers

    @pytest.mark.asyncio
    async def test_dispatch_clears_context(self, middleware, mock_request, mock_response):
        """测试清理上下文"""
        async def call_next(request):
            return mock_response

        with patch('src.interfaces.rest.middleware.logging_middleware.bind_context'):
            with patch('src.interfaces.rest.middleware.logging_middleware.clear_context') as mock_clear:
                await middleware.dispatch(mock_request, call_next)

                # Check clear_context was called
                mock_clear.assert_called_once()

    @pytest.mark.asyncio
    async def test_dispatch_clears_context_on_exception(self, middleware, mock_request):
        """测试异常时清理上下文"""
        async def call_next(request):
            raise RuntimeError("Test error")

        with patch('src.interfaces.rest.middleware.logging_middleware.bind_context'):
            with patch('src.interfaces.rest.middleware.logging_middleware.clear_context') as mock_clear:
                with pytest.raises(RuntimeError):
                    await middleware.dispatch(mock_request, call_next)

                # Check clear_context was still called
                mock_clear.assert_called_once()

    @pytest.mark.asyncio
    async def test_dispatch_handles_no_client(self, middleware, mock_request, mock_response):
        """测试处理无客户端信息"""
        mock_request.client = None

        async def call_next(request):
            return mock_response

        with patch('src.interfaces.rest.middleware.logging_middleware.bind_context'):
            with patch('src.interfaces.rest.middleware.logging_middleware.clear_context'):
                # Should not raise error
                response = await middleware.dispatch(mock_request, call_next)
                assert response is mock_response

    @pytest.mark.asyncio
    async def test_dispatch_with_different_status_codes(self, middleware, mock_request):
        """测试不同状态码"""
        for status_code in [200, 201, 400, 404, 500]:
            mock_response = MagicMock(spec=Response)
            mock_response.status_code = status_code
            mock_response.headers = {}

            async def call_next(request):
                return mock_response

            with patch('src.interfaces.rest.middleware.logging_middleware.bind_context'):
                with patch('src.interfaces.rest.middleware.logging_middleware.clear_context'):
                    response = await middleware.dispatch(mock_request, call_next)
                    assert response.status_code == status_code

    @pytest.mark.asyncio
    async def test_dispatch_with_post_request(self, middleware, mock_response):
        """测试 POST 请求"""
        mock_request = MagicMock(spec=Request)
        mock_request.method = "POST"
        mock_request.url = MagicMock()
        mock_request.url.path = "/api/v1/sessions"
        mock_request.client = MagicMock()
        mock_request.client.host = "127.0.0.1"
        mock_request.state = MagicMock()

        async def call_next(request):
            return mock_response

        with patch('src.interfaces.rest.middleware.logging_middleware.bind_context') as mock_bind:
            with patch('src.interfaces.rest.middleware.logging_middleware.clear_context'):
                await middleware.dispatch(mock_request, call_next)

                call_kwargs = mock_bind.call_args[1]
                assert call_kwargs['method'] == 'POST'
