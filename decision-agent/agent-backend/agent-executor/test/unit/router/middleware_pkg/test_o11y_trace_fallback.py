"""测试 o11y_trace 在无 TelemetrySDK 时的标准 OTel 回退逻辑。"""

from __future__ import annotations

import importlib.util
import sys
from pathlib import Path
from types import ModuleType, SimpleNamespace
from unittest.mock import AsyncMock, MagicMock, Mock, patch

import pytest
from fastapi import Request, Response


MODULE_PATH = (
    Path(__file__).resolve().parents[4] / "app/router/middleware_pkg/o11y_trace.py"
)


def _load_module(trace_enabled: bool):
    sdk_available = ModuleType("app.utils.observability.sdk_available")
    sdk_available.TELEMETRY_SDK_AVAILABLE = False

    o11y_log = ModuleType("app.utils.observability.observability_log")
    o11y_log.get_logger = lambda: None

    spec = importlib.util.spec_from_file_location(
        "test_o11y_trace_fallback_module", MODULE_PATH
    )
    module = importlib.util.module_from_spec(spec)
    assert spec.loader is not None

    with patch.dict(
        sys.modules,
        {
            "app.utils.observability.sdk_available": sdk_available,
            "app.utils.observability.observability_log": o11y_log,
        },
    ):
        spec.loader.exec_module(module)

    return module


def _config_module(trace_enabled: bool):
    config_module = ModuleType("app.common.config")
    config_module.Config = SimpleNamespace(
        o11y=SimpleNamespace(trace_enabled=trace_enabled),
        app=SimpleNamespace(host_prefix="/api/agent-executor/v1"),
    )
    return config_module


@pytest.mark.asyncio
async def test_standard_otel_fallback_used_when_sdk_unavailable():
    module = _load_module(trace_enabled=True)

    request = MagicMock(spec=Request)
    request.headers = {}
    request.method = "POST"
    request.url = MagicMock(__str__=Mock(return_value="http://test/api/test"))
    request.url.path = "/api/test"
    request.client = MagicMock(host="127.0.0.1")

    response = Mock(spec=Response)
    response.status_code = 200
    call_next = AsyncMock(return_value=response)

    mock_tracer = MagicMock()
    mock_span = MagicMock()
    mock_span.__enter__ = Mock(return_value=mock_span)
    mock_span.__exit__ = Mock(return_value=False)
    mock_tracer.start_as_current_span = Mock(return_value=mock_span)

    with patch.dict(sys.modules, {"app.common.config": _config_module(True)}):
        with patch.object(
            module.trace, "get_tracer", return_value=mock_tracer
        ) as mock_get:
            result = await module.o11y_trace(request, call_next)

    mock_get.assert_called_once_with("agent-executor.http")
    mock_tracer.start_as_current_span.assert_called_once()
    mock_span.set_attribute.assert_any_call("http.status_code", 200)
    call_next.assert_called_once_with(request)
    assert result is response


@pytest.mark.asyncio
async def test_trace_disabled_skips_tracer_when_sdk_unavailable():
    module = _load_module(trace_enabled=False)

    request = MagicMock(spec=Request)
    request.headers = {}
    request.method = "GET"
    request.url = MagicMock(__str__=Mock(return_value="http://test/api/test"))
    request.url.path = "/api/test"
    request.client = MagicMock(host="127.0.0.1")

    response = Mock(spec=Response)
    response.status_code = 204
    call_next = AsyncMock(return_value=response)

    with patch.dict(sys.modules, {"app.common.config": _config_module(False)}):
        with patch.object(module.trace, "get_tracer") as mock_get:
            result = await module.o11y_trace(request, call_next)

    mock_get.assert_not_called()
    call_next.assert_called_once_with(request)
    assert result is response


@pytest.mark.asyncio
async def test_health_ready_endpoint_skips_trace_reporting():
    module = _load_module(trace_enabled=True)

    request = MagicMock(spec=Request)
    request.headers = {}
    request.method = "GET"
    request.url = MagicMock(
        __str__=Mock(return_value="http://test/api/agent-executor/v1/health/ready")
    )
    request.url.path = "/api/agent-executor/v1/health/ready"
    request.client = MagicMock(host="127.0.0.1")

    response = Mock(spec=Response)
    response.status_code = 200
    call_next = AsyncMock(return_value=response)

    with patch.dict(sys.modules, {"app.common.config": _config_module(True)}):
        with patch.object(module.trace, "get_tracer") as mock_get:
            result = await module.o11y_trace(request, call_next)

    mock_get.assert_not_called()
    call_next.assert_called_once_with(request)
    assert result is response
