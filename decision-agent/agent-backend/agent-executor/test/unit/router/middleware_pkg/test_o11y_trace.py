"""单元测试 - o11y_trace 链路追踪中间件"""

import sys
import pytest
from unittest.mock import AsyncMock, MagicMock, patch, Mock
from fastapi import Request, Response


class TestO11yTrace:
    """测试 o11y_trace 中间件函数"""

    @pytest.mark.asyncio
    async def test_sdk_unavailable_passes_through(self):
        """测试SDK不可用时直接通过"""
        from app.router.middleware_pkg.o11y_trace import o11y_trace

        request = MagicMock(spec=Request)
        request.url.path = "/api/test"
        request.headers = {}
        request.method = "GET"
        request.url = MagicMock(__str__=Mock(return_value="http://test/api/test"))
        request.client = MagicMock(host="127.0.0.1")

        call_next = AsyncMock(return_value=Mock(spec=Response))

        with patch(
            "app.router.middleware_pkg.o11y_trace.TELEMETRY_SDK_AVAILABLE", False
        ):
            response = await o11y_trace(request, call_next)

        call_next.assert_called_once_with(request)

    @pytest.mark.asyncio
    async def test_trace_disabled_in_config(self):
        """测试配置中禁用追踪时直接通过"""
        from app.router.middleware_pkg.o11y_trace import o11y_trace

        request = MagicMock(spec=Request)
        request.url.path = "/api/test"
        request.headers = {}
        request.method = "GET"
        request.url = MagicMock(__str__=Mock(return_value="http://test/api/test"))
        request.client = MagicMock(host="127.0.0.1")

        response = Mock(spec=Response)
        response.status_code = 200

        call_next = AsyncMock(return_value=response)

        with patch("app.common.config.Config") as MockConfig:
            MockConfig.o11y.trace_enabled = False
            with patch(
                "app.router.middleware_pkg.o11y_trace.TELEMETRY_SDK_AVAILABLE", True
            ):
                result = await o11y_trace(request, call_next)

        call_next.assert_called_once_with(request)

    @pytest.mark.asyncio
    async def test_normal_request_flow_with_tracing(self):
        """测试正常请求流程的追踪"""
        from app.router.middleware_pkg.o11y_trace import o11y_trace

        request = MagicMock(spec=Request)
        request.url.path = "/api/agents/123"
        request.headers = {}
        request.method = "POST"
        request.url = MagicMock(__str__=Mock(return_value="http://test/api/agents/123"))
        request.client = MagicMock(host="192.168.1.100")

        response = Mock(spec=Response)
        response.status_code = 201

        call_next = AsyncMock(return_value=response)

        mock_tracer = MagicMock()
        mock_span = MagicMock()
        mock_span.__enter__ = Mock(return_value=mock_span)
        mock_span.__exit__ = Mock(return_value=False)
        mock_span.is_recording = Mock(return_value=True)

        mock_tracer.start_as_current_span = Mock(return_value=mock_span)

        # Create a mock trace_exporter module
        mock_trace_exporter = MagicMock()
        mock_trace_exporter.tracer = mock_tracer

        with patch(
            "app.router.middleware_pkg.o11y_trace.TELEMETRY_SDK_AVAILABLE", True
        ):
            with patch("app.common.config.Config") as MockConfig:
                MockConfig.o11y.trace_enabled = True
                # Inject the mock module into sys.modules
                with patch.dict(
                    sys.modules,
                    {"exporter.ar_trace.trace_exporter": mock_trace_exporter},
                ):
                    result = await o11y_trace(request, call_next)

        call_next.assert_called_once_with(request)

    @pytest.mark.asyncio
    async def test_request_with_trace_context_from_headers(self):
        """测试从请求头提取追踪上下文"""
        from app.router.middleware_pkg.o11y_trace import o11y_trace

        request = MagicMock(spec=Request)
        request.url.path = "/api/test"
        request.headers = {
            "traceparent": "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01",
            "tracestate": "congo=ttl%3D28msrojo%3Dgreen",
        }
        request.method = "GET"
        request.url = MagicMock(__str__=Mock(return_value="http://test/api/test"))
        request.client = MagicMock(host="127.0.0.1")

        response = Mock(spec=Response)
        response.status_code = 200

        call_next = AsyncMock(return_value=response)

        mock_tracer = MagicMock()
        mock_span = MagicMock()
        mock_span.__enter__ = Mock(return_value=mock_span)
        mock_span.__exit__ = Mock(return_value=False)
        mock_span.is_recording = Mock(return_value=True)

        mock_tracer.start_as_current_span = Mock(return_value=mock_span)

        mock_trace_exporter = MagicMock()
        mock_trace_exporter.tracer = mock_tracer

        with patch(
            "app.router.middleware_pkg.o11y_trace.TELEMETRY_SDK_AVAILABLE", True
        ):
            with patch("app.common.config.Config") as MockConfig:
                MockConfig.o11y.trace_enabled = True
                with patch.dict(
                    sys.modules,
                    {"exporter.ar_trace.trace_exporter": mock_trace_exporter},
                ):
                    result = await o11y_trace(request, call_next)

        call_next.assert_called_once_with(request)

    @pytest.mark.asyncio
    async def test_failed_request_records_exception(self):
        """测试失败的请求记录异常"""
        from app.router.middleware_pkg.o11y_trace import o11y_trace

        request = MagicMock(spec=Request)
        request.url.path = "/api/test"
        request.headers = {}
        request.method = "GET"
        request.url = MagicMock(__str__=Mock(return_value="http://test/api/test"))
        request.client = MagicMock(host="127.0.0.1")

        test_error = ValueError("Test error")

        call_next = AsyncMock(side_effect=test_error)

        mock_tracer = MagicMock()
        mock_span = MagicMock()
        mock_span.__enter__ = Mock(return_value=mock_span)
        mock_span.__exit__ = Mock(return_value=False)
        mock_span.is_recording = Mock(return_value=True)

        mock_tracer.start_as_current_span = Mock(return_value=mock_span)

        mock_trace_exporter = MagicMock()
        mock_trace_exporter.tracer = mock_tracer

        with patch(
            "app.router.middleware_pkg.o11y_trace.TELEMETRY_SDK_AVAILABLE", True
        ):
            with patch("app.common.config.Config") as MockConfig:
                MockConfig.o11y.trace_enabled = True
                with patch.dict(
                    sys.modules,
                    {"exporter.ar_trace.trace_exporter": mock_trace_exporter},
                ):
                    with pytest.raises(ValueError):
                        result = await o11y_trace(request, call_next)

        # Verify exception was recorded
        mock_span.record_exception.assert_called_once_with(test_error)

    @pytest.mark.asyncio
    async def test_response_status_code_set_on_span(self):
        """测试响应状态码设置到span"""
        from app.router.middleware_pkg.o11y_trace import o11y_trace

        request = MagicMock(spec=Request)
        request.url.path = "/api/test"
        request.headers = {}
        request.method = "GET"
        request.url = MagicMock(__str__=Mock(return_value="http://test/api/test"))
        request.client = MagicMock(host="127.0.0.1")

        response = Mock(spec=Response)
        response.status_code = 201

        call_next = AsyncMock(return_value=response)

        mock_tracer = MagicMock()
        mock_span = MagicMock()
        mock_span.__enter__ = Mock(return_value=mock_span)
        mock_span.__exit__ = Mock(return_value=False)
        mock_span.is_recording = Mock(return_value=True)

        mock_tracer.start_as_current_span = Mock(return_value=mock_span)

        mock_trace_exporter = MagicMock()
        mock_trace_exporter.tracer = mock_tracer

        with patch(
            "app.router.middleware_pkg.o11y_trace.TELEMETRY_SDK_AVAILABLE", True
        ):
            with patch("app.common.config.Config") as MockConfig:
                MockConfig.o11y.trace_enabled = True
                with patch.dict(
                    sys.modules,
                    {"exporter.ar_trace.trace_exporter": mock_trace_exporter},
                ):
                    result = await o11y_trace(request, call_next)

        # Verify status code was set
        mock_span.set_attribute.assert_any_call("http.status_code", 201)

    @pytest.mark.asyncio
    async def test_error_status_code_set_on_exception(self):
        """测试异常时设置500状态码"""
        from app.router.middleware_pkg.o11y_trace import o11y_trace

        request = MagicMock(spec=Request)
        request.url.path = "/api/test"
        request.headers = {}
        request.method = "GET"
        request.url = MagicMock(__str__=Mock(return_value="http://test/api/test"))
        request.client = MagicMock(host="127.0.0.1")

        test_error = ValueError("Test error")

        call_next = AsyncMock(side_effect=test_error)

        mock_tracer = MagicMock()
        mock_span = MagicMock()
        mock_span.__enter__ = Mock(return_value=mock_span)
        mock_span.__exit__ = Mock(return_value=False)
        mock_span.is_recording = Mock(return_value=True)

        mock_tracer.start_as_current_span = Mock(return_value=mock_span)

        mock_trace_exporter = MagicMock()
        mock_trace_exporter.tracer = mock_tracer

        with patch(
            "app.router.middleware_pkg.o11y_trace.TELEMETRY_SDK_AVAILABLE", True
        ):
            with patch("app.common.config.Config") as MockConfig:
                MockConfig.o11y.trace_enabled = True
                with patch.dict(
                    sys.modules,
                    {"exporter.ar_trace.trace_exporter": mock_trace_exporter},
                ):
                    with pytest.raises(ValueError):
                        result = await o11y_trace(request, call_next)

        # Verify error status code was set
        mock_span.set_attribute.assert_any_call("http.status_code", 500)

    @pytest.mark.asyncio
    async def test_span_attributes_set(self):
        """测试span属性设置"""
        from app.router.middleware_pkg.o11y_trace import o11y_trace

        request = MagicMock(spec=Request)
        request.url.path = "/api/agents/run"
        request.headers = {}
        request.method = "POST"
        request.url = MagicMock(__str__=Mock(return_value="http://test/api/agents/run"))
        request.client = MagicMock(host="192.168.1.1")

        response = Mock(spec=Response)
        response.status_code = 200

        call_next = AsyncMock(return_value=response)

        mock_tracer = MagicMock()
        mock_span = MagicMock()
        mock_span.__enter__ = Mock(return_value=mock_span)
        mock_span.__exit__ = Mock(return_value=False)
        mock_span.is_recording = Mock(return_value=True)

        mock_tracer.start_as_current_span = Mock(return_value=mock_span)

        mock_trace_exporter = MagicMock()
        mock_trace_exporter.tracer = mock_tracer

        with patch(
            "app.router.middleware_pkg.o11y_trace.TELEMETRY_SDK_AVAILABLE", True
        ):
            with patch("app.common.config.Config") as MockConfig:
                MockConfig.o11y.trace_enabled = True
                with patch.dict(
                    sys.modules,
                    {"exporter.ar_trace.trace_exporter": mock_trace_exporter},
                ):
                    result = await o11y_trace(request, call_next)

        # Verify tracer was called
        mock_tracer.start_as_current_span.assert_called_once()
