"""测试 - 中间件整合"""

import pytest
from unittest.mock import AsyncMock, MagicMock, Mock, patch


@pytest.mark.asyncio
class TestLogRequestsMiddleware:
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

        # Should not log health endpoints
        call_next.assert_called_once_with(request)

    async def test_generates_request_id(self):
        """测试生成请求ID"""
        from app.router.middleware_pkg.log_requests import log_requests

        request = MagicMock()
        request.url.path = "/api/test"
        request.headers = {}
        request.method = "POST"
        request.client = MagicMock(host="192.168.1.1")
        request.query_params = {}
        request.body = AsyncMock(return_value=b'{"test": "data"}')

        call_next = AsyncMock(return_value=Mock(status_code=200))

        with patch("app.router.middleware_pkg.log_requests.uuid") as mock_uuid:
            mock_uuid.uuid4 = Mock(return_value="test-uuid-123")

            response = await log_requests(request, call_next)

        # Verify a request ID was generated
        mock_uuid.uuid4.assert_called_once()


@pytest.mark.asyncio
class TestStreamingRateLimiter:
    """测试流式速率限制器"""

    def test_init_with_default_rate(self):
        """测试默认速率初始化"""
        from app.router.middleware_pkg.streaming_rate_limiter import (
            StreamingRateLimiter,
            DEFAULT_RATE_LIMIT,
        )

        limiter = StreamingRateLimiter()

        assert limiter.rate_limit == DEFAULT_RATE_LIMIT

    async def test_rate_limits_first_10_chunks(self):
        """测试前10个块不限制"""
        from app.router.middleware_pkg.streaming_rate_limiter import (
            RateLimitedStreamingIterator,
        )

        async def mock_iterator():
            for i in range(15):
                yield f"chunk{i}".encode()

        rate_limited = RateLimitedStreamingIterator(mock_iterator())

        chunks = []
        async for item in rate_limited:
            chunks.append(item)

        assert len(chunks) == 15


@pytest.mark.asyncio
class TestO11yTrace:
    """测试链路追踪中间件"""

    async def test_trace_disabled(self):
        """测试禁用追踪时通过"""
        from app.router.middleware_pkg.o11y_trace import o11y_trace

        request = MagicMock()
        request.url.path = "/api/test"
        request.headers = {}
        request.method = "GET"
        request.url = MagicMock(__str__=Mock(return_value="http://test/api"))
        request.client = MagicMock(host="127.0.0.1")

        response = Mock(status_code=200)
        call_next = AsyncMock(return_value=response)

        with patch("app.common.config.Config") as mock_config:
            mock_config.o11y.trace_enabled = False
            mock_config.app.host_prefix = "/api/agent-executor/v1"
            result = await o11y_trace(request, call_next)

        call_next.assert_called_once_with(request)
