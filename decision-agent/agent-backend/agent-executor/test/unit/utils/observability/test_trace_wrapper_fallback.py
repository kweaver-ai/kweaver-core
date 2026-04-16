"""测试 trace_wrapper 在无 TelemetrySDK 时回退到标准 OTel tracer。"""

from __future__ import annotations

import importlib.util
import sys
from pathlib import Path
from types import ModuleType, SimpleNamespace
from unittest.mock import MagicMock, patch


MODULE_PATH = (
    Path(__file__).resolve().parents[4]
    / "app/utils/observability/trace_wrapper.py"
)


def _load_module():
    sdk_available = ModuleType("app.utils.observability.sdk_available")
    sdk_available.TELEMETRY_SDK_AVAILABLE = False

    utils_common = ModuleType("app.utils.common")
    utils_common.func_judgment = lambda func: (False, False)

    spec = importlib.util.spec_from_file_location(
        "test_trace_wrapper_fallback_module", MODULE_PATH
    )
    module = importlib.util.module_from_spec(spec)
    assert spec.loader is not None

    with patch.dict(
        sys.modules,
        {
            "app.utils.observability.sdk_available": sdk_available,
            "app.utils.common": utils_common,
        },
    ):
        spec.loader.exec_module(module)

    return module


def _config_module(trace_enabled: bool):
    config_module = ModuleType("app.common.config")
    config_module.Config = SimpleNamespace(
        is_o11y_trace_enabled=lambda: trace_enabled
    )
    return config_module


def _stand_log_module():
    module = ModuleType("app.common.stand_log")
    module.StandLogger = SimpleNamespace(info_log=lambda *args, **kwargs: None)
    return module


def _build_tracer():
    mock_tracer = MagicMock()
    mock_span = MagicMock()
    mock_span.is_recording.return_value = True
    mock_span.get_span_context.return_value = SimpleNamespace(trace_id=1, span_id=2)
    mock_tracer.start_span.return_value = mock_span
    return mock_tracer, mock_span


def test_internal_span_uses_standard_otel_tracer_when_sdk_unavailable():
    module = _load_module()
    mock_tracer, mock_span = _build_tracer()

    with patch.dict(
        sys.modules,
        {
            "app.common.config": _config_module(True),
            "app.common.stand_log": _stand_log_module(),
        },
    ):
        with patch("opentelemetry.trace.get_tracer", return_value=mock_tracer) as mock_get:

            @module.internal_span(name="fallback_span")
            def test_func(span=None):
                return "result"

            result = test_func()

    assert result == "result"
    mock_get.assert_called_once_with("agent-executor.internal")
    mock_tracer.start_span.assert_called_once()
    mock_span.end.assert_called_once()


def test_internal_span_returns_original_function_when_trace_disabled():
    module = _load_module()

    with patch.dict(
        sys.modules,
        {
            "app.common.config": _config_module(False),
            "app.common.stand_log": _stand_log_module(),
        },
    ):

        @module.internal_span(name="disabled_span")
        def test_func():
            return "result"

        result = test_func()

    assert result == "result"
