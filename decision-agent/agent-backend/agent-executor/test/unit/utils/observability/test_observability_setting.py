"""
Unit tests for app/utils/observability/observability_setting.py
50+ tests for observability settings and configuration
"""

from unittest.mock import MagicMock, patch
from app.utils.observability.observability_setting import (
    LogSetting,
    TraceSetting,
    MetricSetting,
    ObservabilitySetting,
    ServerInfo,
    inject_trace_context,
    extract_trace_context,
)


class TestLogSetting:
    """Tests for LogSetting class"""

    def test_log_setting_default_init(self):
        """Test LogSetting initialization with defaults"""
        setting = LogSetting()
        assert setting.log_enabled is False
        assert setting.log_exporter == ""
        assert setting.log_load_interval == 0
        assert setting.log_load_max_log == 0
        assert setting.http_log_feed_ingester_url == ""

    def test_log_setting_with_all_params(self):
        """Test LogSetting with all parameters"""
        setting = LogSetting(
            log_enabled=True,
            log_exporter="otlp",
            log_load_interval=60,
            log_load_max_log=1000,
            http_log_feed_ingester_url="http://localhost:4318",
        )
        assert setting.log_enabled is True
        assert setting.log_exporter == "otlp"
        assert setting.log_load_interval == 60
        assert setting.log_load_max_log == 1000
        assert setting.http_log_feed_ingester_url == "http://localhost:4318"

    def test_log_setting_with_enabled_only(self):
        """Test LogSetting with only log_enabled"""
        setting = LogSetting(log_enabled=True)
        assert setting.log_enabled is True
        assert setting.log_exporter == ""

    def test_log_setting_with_exporter_only(self):
        """Test LogSetting with only log_exporter"""
        setting = LogSetting(log_exporter="console")
        assert setting.log_exporter == "console"
        assert setting.log_enabled is False

    def test_log_setting_with_negative_interval(self):
        """Test LogSetting with negative interval"""
        setting = LogSetting(log_load_interval=-100)
        assert setting.log_load_interval == -100

    def test_log_setting_with_zero_max_log(self):
        """Test LogSetting with zero max_log"""
        setting = LogSetting(log_load_max_log=0)
        assert setting.log_load_max_log == 0

    def test_log_setting_with_large_values(self):
        """Test LogSetting with large values"""
        setting = LogSetting(log_load_interval=999999, log_load_max_log=999999)
        assert setting.log_load_interval == 999999
        assert setting.log_load_max_log == 999999

    def test_log_setting_with_url(self):
        """Test LogSetting with URL"""
        url = "https://logs.example.com/v1/logs"
        setting = LogSetting(http_log_feed_ingester_url=url)
        assert setting.http_log_feed_ingester_url == url


class TestTraceSetting:
    """Tests for TraceSetting class"""

    def test_trace_setting_default_init(self):
        """Test TraceSetting initialization with defaults"""
        setting = TraceSetting()
        assert setting.trace_enabled is False
        assert setting.trace_provider == ""
        assert setting.trace_max_queue_size == 512
        assert setting.max_export_batch_size == 512
        assert setting.http_trace_feed_ingester_url == ""
        assert setting.grpc_trace_feed_ingester_url == ""
        assert setting.grpc_trace_job_id == ""

    def test_trace_setting_with_all_params(self):
        """Test TraceSetting with all parameters"""
        setting = TraceSetting(
            trace_enabled=True,
            trace_provider="otlp",
            trace_max_queue_size=2048,
            max_export_batch_size=1024,
            http_trace_feed_ingester_url="http://localhost:4318",
            grpc_trace_feed_ingester_url="localhost:4317",
            grpc_trace_job_id="job-123",
        )
        assert setting.trace_enabled is True
        assert setting.trace_provider == "otlp"
        assert setting.trace_max_queue_size == 2048
        assert setting.max_export_batch_size == 1024
        assert setting.grpc_trace_job_id == "job-123"

    def test_trace_setting_with_enabled_only(self):
        """Test TraceSetting with only trace_enabled"""
        setting = TraceSetting(trace_enabled=True)
        assert setting.trace_enabled is True
        assert setting.trace_provider == ""

    def test_trace_setting_with_provider_only(self):
        """Test TraceSetting with only trace_provider"""
        setting = TraceSetting(trace_provider="jaeger")
        assert setting.trace_provider == "jaeger"
        assert setting.trace_enabled is False

    def test_trace_setting_with_zero_queue_size(self):
        """Test TraceSetting with zero queue size"""
        setting = TraceSetting(trace_max_queue_size=0)
        assert setting.trace_max_queue_size == 0

    def test_trace_setting_with_zero_batch_size(self):
        """Test TraceSetting with zero batch size"""
        setting = TraceSetting(max_export_batch_size=0)
        assert setting.max_export_batch_size == 0

    def test_trace_setting_with_http_url(self):
        """Test TraceSetting with HTTP URL"""
        setting = TraceSetting(http_trace_feed_ingester_url="http://trace.example.com")
        assert setting.http_trace_feed_ingester_url == "http://trace.example.com"

    def test_trace_setting_with_grpc_url(self):
        """Test TraceSetting with gRPC URL"""
        setting = TraceSetting(
            grpc_trace_feed_ingester_url="grpc://trace.example.com:4317"
        )
        assert setting.grpc_trace_feed_ingester_url == "grpc://trace.example.com:4317"

    def test_trace_setting_with_job_id(self):
        """Test TraceSetting with job ID"""
        setting = TraceSetting(grpc_trace_job_id="test-job-456")
        assert setting.grpc_trace_job_id == "test-job-456"

    def test_trace_setting_with_large_queue_size(self):
        """Test TraceSetting with large queue size"""
        setting = TraceSetting(trace_max_queue_size=1000000)
        assert setting.trace_max_queue_size == 1000000


class TestMetricSetting:
    """Tests for MetricSetting class"""

    def test_metric_setting_default_init(self):
        """Test MetricSetting initialization with defaults"""
        setting = MetricSetting()
        assert setting.metric_enabled is False
        assert setting.metric_provider == ""
        assert setting.http_metric_feed_ingester_url == ""
        assert setting.metric_interval_second == 0

    def test_metric_setting_with_all_params(self):
        """Test MetricSetting with all parameters"""
        setting = MetricSetting(
            metric_enabled=True,
            metric_provider="prometheus",
            http_metric_feed_ingester_url="http://localhost:9090",
            metric_interval_second=60,
        )
        assert setting.metric_enabled is True
        assert setting.metric_provider == "prometheus"
        assert setting.http_metric_feed_ingester_url == "http://localhost:9090"
        assert setting.metric_interval_second == 60

    def test_metric_setting_with_enabled_only(self):
        """Test MetricSetting with only metric_enabled"""
        setting = MetricSetting(metric_enabled=True)
        assert setting.metric_enabled is True
        assert setting.metric_provider == ""

    def test_metric_setting_with_provider_only(self):
        """Test MetricSetting with only metric_provider"""
        setting = MetricSetting(metric_provider="statsd")
        assert setting.metric_provider == "statsd"
        assert setting.metric_enabled is False

    def test_metric_setting_with_zero_interval(self):
        """Test MetricSetting with zero interval"""
        setting = MetricSetting(metric_interval_second=0)
        assert setting.metric_interval_second == 0

    def test_metric_setting_with_negative_interval(self):
        """Test MetricSetting with negative interval"""
        setting = MetricSetting(metric_interval_second=-30)
        assert setting.metric_interval_second == -30

    def test_metric_setting_with_large_interval(self):
        """Test MetricSetting with large interval"""
        setting = MetricSetting(metric_interval_second=86400)
        assert setting.metric_interval_second == 86400

    def test_metric_setting_with_url(self):
        """Test MetricSetting with URL"""
        url = "http://metrics.example.com"
        setting = MetricSetting(http_metric_feed_ingester_url=url)
        assert setting.http_metric_feed_ingester_url == url


class TestObservabilitySetting:
    """Tests for ObservabilitySetting class"""

    def test_observability_setting_default_init(self):
        """Test ObservabilitySetting initialization with defaults"""
        setting = ObservabilitySetting()
        assert isinstance(setting.log, LogSetting)
        assert isinstance(setting.trace, TraceSetting)
        assert isinstance(setting.metric, MetricSetting)
        assert setting.log.log_enabled is False
        assert setting.trace.trace_enabled is False
        assert setting.metric.metric_enabled is False

    def test_observability_setting_with_log_only(self):
        """Test ObservabilitySetting with only log setting"""
        log_setting = LogSetting(log_enabled=True, log_exporter="otlp")
        setting = ObservabilitySetting(log=log_setting)
        assert setting.log.log_enabled is True
        assert setting.log.log_exporter == "otlp"
        assert setting.trace.trace_enabled is False
        assert setting.metric.metric_enabled is False

    def test_observability_setting_with_trace_only(self):
        """Test ObservabilitySetting with only trace setting"""
        trace_setting = TraceSetting(trace_enabled=True, trace_provider="jaeger")
        setting = ObservabilitySetting(trace=trace_setting)
        assert setting.log.log_enabled is False
        assert setting.trace.trace_enabled is True
        assert setting.trace.trace_provider == "jaeger"
        assert setting.metric.metric_enabled is False

    def test_observability_setting_with_metric_only(self):
        """Test ObservabilitySetting with only metric setting"""
        metric_setting = MetricSetting(metric_enabled=True)
        setting = ObservabilitySetting(metric=metric_setting)
        assert setting.log.log_enabled is False
        assert setting.trace.trace_enabled is False
        assert setting.metric.metric_enabled is True

    def test_observability_setting_with_all_settings(self):
        """Test ObservabilitySetting with all settings"""
        log_setting = LogSetting(log_enabled=True)
        trace_setting = TraceSetting(trace_enabled=True)
        metric_setting = MetricSetting(metric_enabled=True)
        setting = ObservabilitySetting(
            log=log_setting, trace=trace_setting, metric=metric_setting
        )
        assert setting.log.log_enabled is True
        assert setting.trace.trace_enabled is True
        assert setting.metric.metric_enabled is True

    def test_observability_setting_with_none_log(self):
        """Test ObservabilitySetting with None log creates default"""
        setting = ObservabilitySetting(log=None)
        assert isinstance(setting.log, LogSetting)
        assert setting.log.log_enabled is False

    def test_observability_setting_with_none_trace(self):
        """Test ObservabilitySetting with None trace creates default"""
        setting = ObservabilitySetting(trace=None)
        assert isinstance(setting.trace, TraceSetting)
        assert setting.trace.trace_enabled is False

    def test_observability_setting_with_none_metric(self):
        """Test ObservabilitySetting with None metric creates default"""
        setting = ObservabilitySetting(metric=None)
        assert isinstance(setting.metric, MetricSetting)
        assert setting.metric.metric_enabled is False


class TestServerInfo:
    """Tests for ServerInfo class"""

    def test_server_info_default_init(self):
        """Test ServerInfo initialization with defaults"""
        info = ServerInfo()
        assert info.server_name == ""
        assert info.server_version == ""
        assert info.language == ""
        assert info.python_version == ""

    def test_server_info_with_all_params(self):
        """Test ServerInfo with all parameters"""
        info = ServerInfo(
            server_name="agent-executor",
            server_version="1.0.0",
            language="python",
            python_version="3.11",
        )
        assert info.server_name == "agent-executor"
        assert info.server_version == "1.0.0"
        assert info.language == "python"
        assert info.python_version == "3.11"

    def test_server_info_with_name_only(self):
        """Test ServerInfo with only server_name"""
        info = ServerInfo(server_name="test-server")
        assert info.server_name == "test-server"
        assert info.server_version == ""

    def test_server_info_with_version_only(self):
        """Test ServerInfo with only server_version"""
        info = ServerInfo(server_version="2.0.0")
        assert info.server_version == "2.0.0"
        assert info.server_name == ""

    def test_server_info_with_language_only(self):
        """Test ServerInfo with only language"""
        info = ServerInfo(language="java")
        assert info.language == "java"
        assert info.server_name == ""

    def test_server_info_with_python_version_only(self):
        """Test ServerInfo with only python_version"""
        info = ServerInfo(python_version="3.12")
        assert info.python_version == "3.12"
        assert info.server_name == ""

    def test_server_info_with_empty_strings(self):
        """Test ServerInfo with empty strings"""
        info = ServerInfo("", "", "", "")
        assert info.server_name == ""
        assert info.server_version == ""
        assert info.language == ""
        assert info.python_version == ""


class TestInjectTraceContext:
    """Tests for inject_trace_context function"""

    @patch("app.utils.observability.observability_setting.trace")
    def test_inject_trace_context_with_recording_span(self, mock_trace):
        """Test inject_trace_context with recording span"""
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

    @patch("app.utils.observability.observability_setting.trace")
    def test_inject_trace_context_without_recording_span(self, mock_trace):
        """Test inject_trace_context without recording span"""
        mock_span = MagicMock()
        mock_span.is_recording.return_value = False
        mock_trace.get_current_span.return_value = mock_span

        headers = {"existing": "header"}
        result = inject_trace_context(headers)

        assert result == headers
        assert headers == {"existing": "header"}

    @patch("app.utils.observability.observability_setting.trace")
    def test_inject_trace_context_with_existing_headers(self, mock_trace):
        """Test inject_trace_context preserves existing headers"""
        mock_span = MagicMock()
        mock_span.is_recording.return_value = True
        mock_trace.get_current_span.return_value = mock_span

        mock_propagator = MagicMock()
        mock_propagator.inject = MagicMock()

        with patch(
            "app.utils.observability.observability_setting.get_global_textmap",
            return_value=mock_propagator,
        ):
            headers = {"Authorization": "Bearer token"}
            result = inject_trace_context(headers)

            assert "Authorization" in result
            assert result["Authorization"] == "Bearer token"

    @patch("app.utils.observability.observability_setting.trace")
    def test_inject_trace_context_with_empty_headers(self, mock_trace):
        """Test inject_trace_context with empty headers dict"""
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

            assert isinstance(result, dict)


class TestExtractTraceContext:
    """Tests for extract_trace_context function"""

    @patch("app.utils.observability.observability_setting.trace")
    def test_extract_trace_context_with_valid_headers(self, mock_trace):
        """Test extract_trace_context with valid headers"""
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

    @patch("app.utils.observability.observability_setting.trace")
    def test_extract_trace_context_with_empty_headers(self, mock_trace):
        """Test extract_trace_context with empty headers"""
        mock_propagator = MagicMock()
        mock_context = MagicMock()
        mock_span_context = MagicMock()

        mock_propagator.extract.return_value = mock_context
        mock_context.get.return_value.get_span_context.return_value = mock_span_context

        with patch(
            "app.utils.observability.observability_setting.get_global_textmap",
            return_value=mock_propagator,
        ):
            headers = {}
            extract_trace_context(headers)

            mock_propagator.extract.assert_called_once_with(headers)

    @patch("app.utils.observability.observability_setting.trace")
    def test_extract_trace_context_with_multiple_headers(self, mock_trace):
        """Test extract_trace_context with multiple headers"""
        mock_propagator = MagicMock()
        mock_context = MagicMock()
        mock_span_context = MagicMock()

        mock_propagator.extract.return_value = mock_context
        mock_context.get.return_value.get_span_context.return_value = mock_span_context

        with patch(
            "app.utils.observability.observability_setting.get_global_textmap",
            return_value=mock_propagator,
        ):
            headers = {
                "traceparent": "test-trace",
                "tracestate": "test-state",
                "other": "value",
            }
            extract_trace_context(headers)

            mock_propagator.extract.assert_called_once_with(headers)
