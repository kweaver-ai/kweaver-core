"""Unit tests for app/router/middleware_pkg/o11y_trace.py."""

from __future__ import annotations

import importlib.util
import sys
from pathlib import Path
from types import ModuleType
from unittest.mock import AsyncMock, MagicMock, Mock, patch

import pytest
from fastapi import Request, Response
from app.common.config import Config


MODULE_PATH = (
    Path(__file__).resolve().parents[4] / "app/router/middleware_pkg/o11y_trace.py"
)


def _load_module():
    o11y_log = ModuleType("app.utils.observability.observability_log")
    o11y_log.get_logger = lambda: None
    stand_log = ModuleType("app.common.stand_log")
    stand_log.StandLogger = MagicMock()

    spec = importlib.util.spec_from_file_location("test_o11y_trace_module", MODULE_PATH)
    module = importlib.util.module_from_spec(spec)
    assert spec.loader is not None

    with patch.dict(
        sys.modules,
        {
            "app.utils.observability.observability_log": o11y_log,
            "app.common.stand_log": stand_log,
        },
    ):
        spec.loader.exec_module(module)

    return module


class TestO11yTrace:
    @pytest.mark.asyncio
    async def test_trace_disabled_in_config_passes_through(self):
        module = _load_module()

        request = MagicMock(spec=Request)
        request.headers = {}
        request.method = "GET"
        request.url = MagicMock(__str__=Mock(return_value="http://test/api/test"))
        request.url.path = "/api/test"
        request.client = MagicMock(host="127.0.0.1")

        response = Mock(spec=Response)
        response.status_code = 200
        call_next = AsyncMock(return_value=response)

        with patch.object(Config.o11y, "trace_enabled", False):
            with patch.object(Config.app, "host_prefix", "/api/agent-executor/v1"):
                result = await module.o11y_trace(request, call_next)

        call_next.assert_called_once_with(request)
        assert result is response

    @pytest.mark.asyncio
    async def test_uses_standard_otel_tracer_when_enabled(self):
        module = _load_module()

        request = MagicMock(spec=Request)
        request.headers = {}
        request.method = "POST"
        request.url = MagicMock(__str__=Mock(return_value="http://test/api/test"))
        request.url.path = "/api/test"
        request.client = MagicMock(host="127.0.0.1")

        response = Mock(spec=Response)
        response.status_code = 201
        call_next = AsyncMock(return_value=response)

        mock_tracer = MagicMock()
        mock_span = MagicMock()
        mock_span.__enter__ = Mock(return_value=mock_span)
        mock_span.__exit__ = Mock(return_value=False)
        mock_tracer.start_as_current_span.return_value = mock_span
        with patch.object(Config.o11y, "trace_enabled", True):
            with patch.object(Config.app, "host_prefix", "/api/agent-executor/v1"):
                with patch.object(
                    module, "_get_request_tracer", return_value=mock_tracer
                ) as mock_get_tracer:
                    result = await module.o11y_trace(request, call_next)

        mock_get_tracer.assert_called_once_with()
        mock_tracer.start_as_current_span.assert_called_once()
        mock_span.set_attribute.assert_any_call("http.status_code", 201)
        module.StandLogger.info.assert_not_called()
        assert result is response

    @pytest.mark.asyncio
    async def test_health_ready_endpoint_skips_trace_reporting(self):
        module = _load_module()

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

        with patch.object(Config.o11y, "trace_enabled", True):
            with patch.object(Config.app, "host_prefix", "/api/agent-executor/v1"):
                with patch.object(module, "_get_request_tracer") as mock_get_tracer:
                    result = await module.o11y_trace(request, call_next)

        mock_get_tracer.assert_not_called()
        call_next.assert_called_once_with(request)
        assert result is response

    @pytest.mark.asyncio
    async def test_failed_request_records_exception(self):
        module = _load_module()

        request = MagicMock(spec=Request)
        request.headers = {}
        request.method = "GET"
        request.url = MagicMock(__str__=Mock(return_value="http://test/api/test"))
        request.url.path = "/api/test"
        request.client = MagicMock(host="127.0.0.1")

        test_error = ValueError("Test error")
        call_next = AsyncMock(side_effect=test_error)

        mock_tracer = MagicMock()
        mock_span = MagicMock()
        mock_span.__enter__ = Mock(return_value=mock_span)
        mock_span.__exit__ = Mock(return_value=False)
        mock_tracer.start_as_current_span.return_value = mock_span
        mock_logger = MagicMock()

        with patch.object(Config.o11y, "trace_enabled", True):
            with patch.object(Config.app, "host_prefix", "/api/agent-executor/v1"):
                with patch.object(
                    module, "_get_request_tracer", return_value=mock_tracer
                ):
                    with patch.object(module, "o11y_logger", return_value=mock_logger):
                        with pytest.raises(ValueError, match="Test error"):
                            await module.o11y_trace(request, call_next)

        mock_span.set_attribute.assert_any_call("http.status_code", 500)
        mock_span.record_exception.assert_called_once_with(test_error)
        module.StandLogger.error.assert_called_once_with("Error: Test error")
        mock_logger.error.assert_called_once()
