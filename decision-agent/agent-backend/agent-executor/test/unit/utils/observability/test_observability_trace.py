"""Unit tests for app/utils/observability/observability_trace.py."""

from unittest.mock import MagicMock, patch


class TestInitTraceProvider:
    @patch("app.utils.observability.observability_trace.set_tracer_provider")
    @patch("app.utils.observability.observability_trace.TracerProvider")
    @patch("app.utils.observability.observability_trace.BatchSpanProcessor")
    @patch("app.utils.observability.observability_trace.Resource")
    def test_init_trace_provider_uses_otlp_endpoint_without_provider_field(
        self,
        m_resource,
        m_batch_processor,
        m_tracer_provider,
        m_set_tracer,
    ):
        from app.utils.observability.observability_trace import init_trace_provider
        from app.utils.observability.observability_setting import ServerInfo, TraceSetting

        mock_exporter = MagicMock()
        mock_processor = MagicMock()
        mock_provider = MagicMock()
        m_batch_processor.return_value = mock_processor
        m_tracer_provider.return_value = mock_provider
        m_resource.create.return_value = MagicMock()

        server_info = ServerInfo(server_name="test-service", server_version="1.0.0")
        setting = TraceSetting(
            trace_enabled=True,
            otlp_endpoint="otelcol-contrib:4318",
            environment="staging",
            sampling_rate=0.3,
            trace_max_queue_size=1024,
        )

        with patch(
            "opentelemetry.exporter.otlp.proto.http.trace_exporter.OTLPSpanExporter",
            return_value=mock_exporter,
        ) as mock_otlp_exporter:
            init_trace_provider(server_info, setting)

        mock_otlp_exporter.assert_called_once_with(
            endpoint="http://otelcol-contrib:4318/v1/traces"
        )
        m_batch_processor.assert_called_once_with(
            span_exporter=mock_exporter,
            schedule_delay_millis=2000,
            max_queue_size=1024,
        )
        m_tracer_provider.assert_called_once()
        m_set_tracer.assert_called_once_with(mock_provider)
