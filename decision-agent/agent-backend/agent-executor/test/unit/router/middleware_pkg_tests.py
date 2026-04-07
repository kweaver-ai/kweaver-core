# -*- coding:utf-8 -*-
"""单元测试 - 中间件模块"""

import pytest
from unittest.mock import AsyncMock, MagicMock, patch


@pytest.mark.asyncio
class TestLogRequests:
    """测试 log_requests 中间件"""

    async def test_excludes_health_endpoints(self):
        """测试排除健康检查端点"""
        from app.router.middleware_pkg.log_requests import log_requests
        from app.common.config import Config

        request = MagicMock()
        request.url.path = Config.app.host_prefix + "/health/alive"
        request.headers = {}
        request.method = "GET"
        request.client = MagicMock(host="127.0.0.1")
        request.query_params = {}
        request.body = AsyncMock(return_value=b"")

        call_next = AsyncMock(return_value=Mock(status_code=200))

        response = await log_requests(request, call_next)

        call_next.assert_called_once_with(request)

    async def test_logs_request_info(self):
        """测试记录请求信息"""
        from app.router.middleware_pkg.log_requests import log_requests

        request = MagicMock()
        request.url.path = "/api/test"
        request.url.path.__eq__ = Mock(return_value=False)
        request.headers = {"x-request-id": "test-123"}
        request.method = "POST"
        request.client = MagicMock(host="192.168.1.1")
        request.url.query = "param=value"
        request.url.__str__ = Mock(return_value="http://test/api/test?param=value")
        request.body = AsyncMock(return_value=b'{"test": "data"}')

        call_next = AsyncMock(return_value=Mock(status_code=200))

        with patch("app.router.middleware_pkg.log_requests.StandLogger") as mock_logger:
            response = await log_requests(request, call_next)

        mock_logger.console_request_log.assert_called_once()


@pytest.mark.asyncio
class TestStreamingRateLimiter:
    """测试 streaming_rate_limiter 中间件"""

    def test_init_with_default_rate(self):
        """测试默认速率初始化"""
        from app.router.middleware_pkg.streaming_rate_limiter import (
            StreamingRateLimiter,
            DEFAULT_RATE_LIMIT,
        )

        limiter = StreamingRateLimiter()

        assert limiter.rate_limit == DEFAULT_RATE_LIMIT
        assert limiter.min_interval == 1.0 / DEFAULT_RATE_LIMIT

    def test_init_with_custom_rate(self):
        """测试自定义速率初始化"""
        from app.router.middleware_pkg.streaming_rate_limiter import (
            StreamingRateLimiter,
        )

        limiter = StreamingRateLimiter(rate_limit=5)

        assert limiter.rate_limit == 5
        assert limiter.min_interval == 0.2

    def test_init_with_rate_zero_clamps(self):
        """测试速率为0时限制为1"""
        from app.router.middleware_pkg.streaming_rate_limiter import (
            StreamingRateLimiter,
        )

        limiter = StreamingRateLimiter(rate_limit=0)

        assert limiter.rate_limit == 1
        assert limiter.min_interval == 1.0


@pytest.mark.asyncio
class TestStreamingResponseHandler:
    """测试 streaming_response_handler 中间件"""

    async def test_get_streaming_log_file_path(self):
        """测试获取日志文件路径"""
        from app.router.middleware_pkg.streaming_response_handler import (
            _get_streaming_log_file_path,
            STREAMING_RESPONSE_LOG_DIR,
        )

        request_id = "test-request-123"

        file_path = _get_streaming_log_file_path(request_id)

        assert STREAMING_RESPONSE_LOG_DIR in file_path
        assert request_id in file_path
        assert file_path.endswith(".log")

    async def test_write_chunk_to_file(self):
        """测试写入块到文件"""
        from app.router.middleware_pkg.streaming_response_handler import (
            _write_chunk_to_file,
        )

        import tempfile
        import os

        with tempfile.NamedTemporaryFile(mode="w", delete=False, suffix=".log") as f:
            file_path = f.name

        _write_chunk_to_file(file_path, "test content", 1, 12)

        with open(file_path, "r", encoding="utf-8") as f:
            content = f.read()

        assert "test content" in content
        assert "Chunk 1" in content
        assert "12 bytes" in content

        os.unlink(file_path)


@pytest.mark.asyncio
class TestO11yTrace:
    """测试 o11y_trace 中间件"""

    async def test_sdk_unavailable_passes_through(self):
        """测试SDK不可用时直接通过"""
        from app.router.middleware_pkg.o11y_trace import o11y_trace

        request = MagicMock()
        request.url.path = "/api/test"
        request.headers = {}
        request.method = "GET"
        request.url = MagicMock(__str__=Mock(return_value="http://test/api/test"))
        request.client = MagicMock(host="127.0.0.1")

        call_next = AsyncMock(return_value=Mock(status_code=200))

        with patch(
            "app.router.middleware_pkg.o11y_trace.TELEMETRY_SDK_AVAILABLE", False
        ):
            response = await o11y_trace(request, call_next)

        call_next.assert_called_once_with(request)

    async def test_with_tracing_enabled(self):
        """测试启用追踪时创建span"""
        from app.router.middleware_pkg.o11y_trace import o11y_trace

        request = MagicMock()
        request.url.path = "/api/test"
        request.headers = {}
        request.method = "GET"
        request.url = MagicMock(__str__=Mock(return_value="http://test/api/test"))
        request.client = MagicMock(host="127.0.0.1")

        response = Mock(status_code=200)
        call_next = AsyncMock(return_value=response)

        mock_tracer = MagicMock()
        mock_span = MagicMock()
        mock_span.__enter__ = Mock(return_value=mock_span)
        mock_span.__exit__ = Mock(return_value=False)

        mock_tracer.start_as_current_span = Mock(return_value=mock_span)

        with patch(
            "app.router.middleware_pkg.o11y_trace.TELEMETRY_SDK_AVAILABLE", True
        ):
            with patch("app.router.middleware_pkg.o11y_trace.Config") as MockConfig:
                with patch.object(MockConfig.o11y, "trace_enabled", True):
                    with patch(
                        "app.router.middleware_pkg.o11y_trace.tracer", mock_tracer
                    ):
                        result = await o11y_trace(request, call_next)

        mock_tracer.start_as_current_span.assert_called_once()


@pytest.mark.asyncio
class TestExceptionHandlers:
    """测试异常处理器"""

    async def test_handle_param_exception(self):
        """测试参数异常处理"""
        from app.router.exception_handler.param_handler import handle_param_exception

        request = MagicMock(spec=Request)
        request.headers = MagicMock()

        exc = ParamException("Test param error")

        with patch("app.router.exception_handler.param_handler.struct_logger.error"):
            response = handle_param_exception(request, exc)

        assert response.status_code == 400

    async def test_handle_permission_exception(self):
        """测试权限异常处理"""
        from app.router.exception_handler.permission_handler import (
            handle_permission_exception,
        )
        from app.common.errors import AgentPermissionException

        request = MagicMock(spec=Request)
        request.headers = MagicMock()

        exc = AgentPermissionException("agent123", "user456")

        with patch(
            "app.router.exception_handler.permission_handler.struct_logger.error"
        ):
            response = handle_permission_exception(request, exc)

        assert response.status_code == 403
