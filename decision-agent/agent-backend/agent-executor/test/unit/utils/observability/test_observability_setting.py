"""Unit tests for app/utils/observability/observability_setting.py."""

from unittest.mock import MagicMock, patch

from app.utils.observability.observability_setting import (
    LogSetting,
    ObservabilitySetting,
    ServerInfo,
    TraceSetting,
    extract_trace_context,
    inject_trace_context,
)


class TestLogSetting:
    def test_log_setting_only_keeps_latest_otel_fields(self):
        setting = LogSetting(log_enabled=True, log_level="warning")

        assert setting.log_enabled is True
        assert setting.log_level == "warning"
        assert not hasattr(setting, "log_exporter")
        assert not hasattr(setting, "log_load_interval")
        assert not hasattr(setting, "log_load_max_log")
        assert not hasattr(setting, "http_log_feed_ingester_url")


class TestTraceSetting:
    def test_trace_setting_only_keeps_latest_otel_fields(self):
        setting = TraceSetting(
            trace_enabled=True,
            otlp_endpoint="http://otelcol-contrib:4318",
            environment="staging",
            sampling_rate=0.25,
            trace_max_queue_size=1024,
        )

        assert setting.trace_enabled is True
        assert setting.otlp_endpoint == "http://otelcol-contrib:4318"
        assert setting.environment == "staging"
        assert setting.sampling_rate == 0.25
        assert setting.trace_max_queue_size == 1024
        assert not hasattr(setting, "trace_provider")
        assert not hasattr(setting, "max_export_batch_size")
        assert not hasattr(setting, "http_trace_feed_ingester_url")
        assert not hasattr(setting, "grpc_trace_feed_ingester_url")
        assert not hasattr(setting, "grpc_trace_job_id")


class TestObservabilitySetting:
    def test_observability_setting_only_contains_log_and_trace(self):
        setting = ObservabilitySetting(
            log=LogSetting(log_enabled=True, log_level="info"),
            trace=TraceSetting(trace_enabled=True, otlp_endpoint="http://otel:4318"),
        )

        assert isinstance(setting.log, LogSetting)
        assert isinstance(setting.trace, TraceSetting)
        assert not hasattr(setting, "metric")


class TestServerInfo:
    def test_server_info_initialization(self):
        info = ServerInfo(
            server_name="agent-executor",
            server_version="1.0.0",
            language="python",
            python_version="3.12",
        )

        assert info.server_name == "agent-executor"
        assert info.server_version == "1.0.0"
        assert info.language == "python"
        assert info.python_version == "3.12"


class TestInjectTraceContext:
    @patch("app.utils.observability.observability_setting.trace")
    def test_inject_trace_context_with_recording_span(self, mock_trace):
        mock_span = MagicMock()
        mock_span.is_recording.return_value = True
        mock_trace.get_current_span.return_value = mock_span

        mock_propagator = MagicMock()
        mock_propagator.inject = MagicMock()

        with patch(
            "app.utils.observability.observability_setting.get_global_textmap",
            return_value=mock_propagator,
        ):
            headers = {}
            result = inject_trace_context(headers)

        mock_propagator.inject.assert_called_once_with(headers)
        assert result == headers


class TestExtractTraceContext:
    @patch("app.utils.observability.observability_setting.trace")
    def test_extract_trace_context_with_valid_headers(self, mock_trace):
        mock_propagator = MagicMock()
        mock_context = MagicMock()
        mock_span_context = MagicMock()

        mock_propagator.extract.return_value = mock_context
        mock_context.get.return_value.get_span_context.return_value = mock_span_context

        with patch(
            "app.utils.observability.observability_setting.get_global_textmap",
            return_value=mock_propagator,
        ):
            headers = {"traceparent": "test-trace"}
            extract_trace_context(headers)

        mock_propagator.extract.assert_called_once_with(headers)
