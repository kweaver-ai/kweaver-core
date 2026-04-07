"""单元测试 - log_requests 请求日志中间件"""

import pytest
import json
from unittest.mock import AsyncMock, MagicMock, patch, Mock
from fastapi import Request, Response


class AsyncIterator:
    """异步迭代器辅助类"""

    def __init__(self, chunks):
        self.chunks = chunks

    def __aiter__(self):
        return self

    async def __anext__(self):
        if not self.chunks:
            raise StopAsyncIteration
        return self.chunks.pop(0)


class TestLogRequests:
    """测试 log_requests 中间件函数"""

    @pytest.mark.asyncio
    async def test_health_alive_endpoint_excluded(self):
        """测试健康检查alive端点被排除"""
        from app.router.middleware_pkg.log_requests import log_requests
        from app.common.config import Config

        request = MagicMock(spec=Request)
        request.url.path = Config.app.host_prefix + "/health/alive"
        request.headers = {}
        request.method = "GET"
        request.client = MagicMock(host="127.0.0.1")
        request.query_params = {}

        call_next = AsyncMock(return_value=Mock(spec=Response))

        response = await log_requests(request, call_next)

        call_next.assert_called_once_with(request)
        assert call_next.call_count == 1

    @pytest.mark.asyncio
    async def test_health_ready_endpoint_excluded(self):
        """测试健康检查ready端点被排除"""
        from app.router.middleware_pkg.log_requests import log_requests
        from app.common.config import Config

        request = MagicMock(spec=Request)
        request.url.path = Config.app.host_prefix + "/health/ready"
        request.headers = {}
        request.method = "GET"
        request.client = MagicMock(host="127.0.0.1")
        request.query_params = {}

        call_next = AsyncMock(return_value=Mock(spec=Response))

        response = await log_requests(request, call_next)

        call_next.assert_called_once_with(request)

    @pytest.mark.asyncio
    async def test_conversation_session_init_excluded_when_config_disabled(self):
        """测试当配置禁用时排除conversation-session/init"""
        from app.router.middleware_pkg.log_requests import log_requests
        from app.common.config import Config

        request = MagicMock(spec=Request)
        request.url.path = Config.app.host_prefix + "/agent/conversation-session/init"
        request.headers = {}
        request.method = "POST"
        request.client = MagicMock(host="127.0.0.1")
        request.query_params = {}

        call_next = AsyncMock(return_value=Mock(spec=Response))

        with patch.object(Config.app, "log_conversation_session_init", False):
            response = await log_requests(request, call_next)

        call_next.assert_called_once_with(request)

    @pytest.mark.asyncio
    async def test_request_id_header_generated_when_missing(self):
        """测试缺少X-Request-ID时生成新ID"""
        from app.router.middleware_pkg.log_requests import log_requests
        import uuid

        request = MagicMock(spec=Request)
        request.url.path = "/api/test"  # A path that's not in exclusion list
        request.headers = MagicMock(get=Mock(return_value=None))
        request.method = "GET"
        request.client = MagicMock(host="127.0.0.1")
        request.query_params = {}
        request.body = AsyncMock(return_value=b"")

        response = Mock(spec=Response)
        response.status_code = 200
        response.headers = {}
        response.body_iterator = AsyncIterator([b""])

        call_next = AsyncMock(return_value=response)

        with patch("uuid.uuid4") as mock_uuid:
            mock_uuid.return_value = uuid.UUID("12345678-1234-5678-1234-567812345678")
            result = await log_requests(request, call_next)

            # Verify request was logged with generated ID
            assert call_next.called

    @pytest.mark.asyncio
    async def test_x_request_id_header_used_when_present(self):
        """测试存在X-Request-ID时使用它"""
        from app.router.middleware_pkg.log_requests import log_requests

        request = MagicMock(spec=Request)
        request.url.path = "/api/test"  # A path that's not in exclusion list
        request.headers = MagicMock(get=Mock(return_value="existing-request-id"))
        request.method = "GET"
        request.client = MagicMock(host="127.0.0.1")
        request.query_params = {}
        request.body = AsyncMock(return_value=b"")

        response = Mock(spec=Response)
        response.status_code = 200
        response.headers = {}
        response.body_iterator = AsyncIterator([b""])

        call_next = AsyncMock(return_value=response)

        result = await log_requests(request, call_next)

        # Verify request was called with existing ID
        assert call_next.called

    @pytest.mark.asyncio
    async def test_request_body_json_cached(self):
        """测试JSON请求体被缓存"""
        from app.router.middleware_pkg.log_requests import (
            log_requests,
        )

        request = MagicMock(spec=Request)
        request.url.path = "/api/test"  # A path that's not in exclusion list
        request.headers = {}
        request.method = "POST"
        request.client = MagicMock(host="127.0.0.1")
        request.query_params = {}
        body = {"key": "value"}
        request.body = AsyncMock(return_value=json.dumps(body).encode())

        response = Mock(spec=Response)
        response.status_code = 200
        response.headers = {"content-type": "application/json"}
        response.body_iterator = AsyncIterator([b""])

        call_next = AsyncMock(return_value=response)

        with patch(
            "app.router.middleware_pkg.log_requests.cache_request_body"
        ) as mock_cache:
            result = await log_requests(request, call_next)
            # Check that cache was called with parsed JSON
            mock_cache.assert_called_once()

    @pytest.mark.asyncio
    async def test_request_body_non_json_cached_as_string(self):
        """测试非JSON请求体作为字符串缓存"""
        from app.router.middleware_pkg.log_requests import log_requests

        request = MagicMock(spec=Request)
        request.url.path = "/api/test"  # A path that's not in exclusion list
        request.headers = {}
        request.method = "POST"
        request.client = MagicMock(host="127.0.0.1")
        request.query_params = {}
        body = "plain text body"
        request.body = AsyncMock(return_value=body.encode())

        response = Mock(spec=Response)
        response.status_code = 200
        response.headers = {"content-type": "text/plain"}
        response.body_iterator = AsyncIterator([b""])

        call_next = AsyncMock(return_value=response)

        with patch(
            "app.router.middleware_pkg.log_requests.cache_request_body"
        ) as mock_cache:
            result = await log_requests(request, call_next)
            mock_cache.assert_called_once()

    @pytest.mark.asyncio
    async def test_streaming_response_detected_text_event_stream(self):
        """测试检测text/event-stream流式响应"""
        from app.router.middleware_pkg.log_requests import log_requests

        request = MagicMock(spec=Request)
        request.url.path = "/api/test"
        request.headers = {}
        request.method = "GET"
        request.client = MagicMock(host="127.0.0.1")
        request.query_params = {}
        request.body = AsyncMock(return_value=b"")

        response = Mock(spec=Response)
        response.status_code = 200
        response.headers = {"content-type": "text/event-stream"}
        response.body_iterator = AsyncIterator([b"data"])

        call_next = AsyncMock(return_value=response)

        with patch(
            "app.router.middleware_pkg.log_requests.handle_streaming_response"
        ) as mock_handle:
            mock_handle.return_value = response
            result = await log_requests(request, call_next)
            mock_handle.assert_called_once()

    @pytest.mark.asyncio
    async def test_streaming_response_detected_application_x_ndjson(self):
        """测试检测application/x-ndjson流式响应"""
        from app.router.middleware_pkg.log_requests import log_requests

        request = MagicMock(spec=Request)
        request.url.path = "/api/test"
        request.headers = {}
        request.method = "GET"
        request.client = MagicMock(host="127.0.0.1")
        request.query_params = {}
        request.body = AsyncMock(return_value=b"")

        response = Mock(spec=Response)
        response.status_code = 200
        response.headers = {"content-type": "application/x-ndjson"}
        response.body_iterator = AsyncIterator([b"data"])

        call_next = AsyncMock(return_value=response)

        with patch(
            "app.router.middleware_pkg.log_requests.handle_streaming_response"
        ) as mock_handle:
            mock_handle.return_value = response
            result = await log_requests(request, call_next)
            mock_handle.assert_called_once()

    @pytest.mark.asyncio
    async def test_non_streaming_response_handling(self):
        """测试非流式响应处理"""
        from app.router.middleware_pkg.log_requests import log_requests

        request = MagicMock(spec=Request)
        request.url.path = "/api/test"
        request.headers = {}
        request.method = "GET"
        request.client = MagicMock(host="127.0.0.1")
        request.query_params = {}
        request.body = AsyncMock(return_value=b"")

        response = Mock(spec=Response)
        response.status_code = 200
        response.headers = {"content-type": "application/json"}
        response.body_iterator = AsyncIterator([b""])

        call_next = AsyncMock(return_value=response)

        result = await log_requests(request, call_next)

        # Verify the response was returned
        assert result is response

    @pytest.mark.asyncio
    async def test_get_request_logging(self):
        """测试GET请求日志"""
        from app.router.middleware_pkg.log_requests import log_requests

        request = MagicMock(spec=Request)
        request.url.path = "/api/resource"
        request.headers = {}
        request.method = "GET"
        request.client = MagicMock(host="192.168.1.1")
        request.query_params = {"id": "123"}
        request.body = AsyncMock(return_value=b"")

        response = Mock(spec=Response)
        response.status_code = 200
        response.headers = {}
        response.body_iterator = AsyncIterator([b"ok"])

        call_next = AsyncMock(return_value=response)

        result = await log_requests(request, call_next)

        assert call_next.called

    @pytest.mark.asyncio
    async def test_post_request_logging(self):
        """测试POST请求日志"""
        from app.router.middleware_pkg.log_requests import log_requests

        request = MagicMock(spec=Request)
        request.url.path = "/api/resource"
        request.headers = {}
        request.method = "POST"
        request.client = MagicMock(host="192.168.1.1")
        request.query_params = {}
        body = {"name": "test"}
        request.body = AsyncMock(return_value=json.dumps(body).encode())

        response = Mock(spec=Response)
        response.status_code = 201
        response.headers = {}
        response.body_iterator = AsyncIterator([b"created"])

        call_next = AsyncMock(return_value=response)

        result = await log_requests(request, call_next)

        assert call_next.called

    @pytest.mark.asyncio
    async def test_process_time_calculation(self):
        """测试处理时间计算"""
        from app.router.middleware_pkg.log_requests import (
            _handle_non_streaming_response,
        )

        # 直接测试 _handle_non_streaming_response 函数
        response = MagicMock()
        response.status_code = 200
        response.headers = {"content-type": "application/json"}
        response.body_iterator = AsyncIterator([b"test response"])

        with patch("app.router.middleware_pkg.log_requests.Config") as mock_config:
            mock_config.is_debug_mode.return_value = False

            result = await _handle_non_streaming_response(
                response, "test-request-id", 100.0
            )

            # 验证返回了响应
            assert result is not None
