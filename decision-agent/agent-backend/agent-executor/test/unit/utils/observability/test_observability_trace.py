"""单元测试 - utils/observability/observability_trace 模块"""

from unittest.mock import MagicMock, patch
import sys


class TestInitTraceProvider:
    """测试 init_trace_provider 函数"""

    @patch("app.utils.observability.observability_trace.TELEMETRY_SDK_AVAILABLE", False)
    def test_returns_early_when_sdk_unavailable(self):
        """测试SDK不可用时直接返回"""
        from app.utils.observability.observability_trace import init_trace_provider
        from app.utils.observability.observability_setting import (
            ServerInfo,
            TraceSetting,
        )

        server_info = ServerInfo(server_name="test", server_version="1.0")
        setting = TraceSetting(trace_enabled=True)

        # Should not raise any error
        init_trace_provider(server_info, setting)

    @patch("app.utils.observability.observability_trace.TELEMETRY_SDK_AVAILABLE", True)
    @patch("app.utils.observability.observability_trace.set_service_info")
    @patch("app.utils.observability.observability_trace.set_tracer_provider")
    @patch("app.common.config.Config")
    def test_returns_when_trace_disabled_in_config(
        self, m_config, m_set_tracer, m_set_service
    ):
        """测试配置中禁用追踪时直接返回"""
        m_config.is_o11y_trace_enabled.return_value = False
        from app.utils.observability.observability_trace import init_trace_provider
        from app.utils.observability.observability_setting import (
            ServerInfo,
            TraceSetting,
        )

        server_info = ServerInfo(server_name="test", server_version="1.0")
        setting = TraceSetting(trace_enabled=True)

        init_trace_provider(server_info, setting)

        m_set_service.assert_called_once()
        m_set_tracer.assert_not_called()

    @patch("app.utils.observability.observability_trace.TELEMETRY_SDK_AVAILABLE", True)
    @patch("app.utils.observability.observability_trace.set_service_info")
    @patch("app.utils.observability.observability_trace.set_tracer_provider")
    @patch("app.utils.observability.observability_trace.TracerProvider")
    @patch("app.utils.observability.observability_trace.BatchSpanProcessor")
    @patch("app.utils.observability.observability_trace.ConsoleSpanExporter")
    @patch("app.utils.observability.observability_trace.trace_resource")
    @patch("app.common.config.Config")
    def test_console_exporter_initialization(
        self,
        m_config,
        m_trace_resource,
        m_console_exporter,
        m_batch_processor,
        m_tracer_provider,
        m_set_tracer,
        m_set_service,
    ):
        """测试使用console导出器初始化"""
        # Mock exporter modules
        sys.modules["exporter.ar_trace.trace_exporter"] = MagicMock(
            ARTraceExporter=MagicMock(return_value=MagicMock())
        )
        sys.modules["exporter.public.client"] = MagicMock(
            HTTPClient=MagicMock(return_value=MagicMock())
        )
        sys.modules["exporter.public.public"] = MagicMock(
            WithAnyRobotURL=MagicMock(return_value="http://test.url")
        )

        try:
            m_config.is_o11y_trace_enabled.return_value = True
            m_trace_resource.return_value = MagicMock()
            from app.utils.observability.observability_trace import init_trace_provider
            from app.utils.observability.observability_setting import (
                ServerInfo,
                TraceSetting,
            )

            mock_exporter = MagicMock()
            m_console_exporter.return_value = mock_exporter
            mock_processor = MagicMock()
            m_batch_processor.return_value = mock_processor
            mock_provider = MagicMock()
            m_tracer_provider.return_value = mock_provider

            server_info = ServerInfo(server_name="test_service", server_version="1.0")
            setting = TraceSetting(
                trace_enabled=True, trace_provider="console", trace_max_queue_size=2048
            )

            init_trace_provider(server_info, setting)

            m_set_service.assert_called_once()
            m_console_exporter.assert_called_once()
            m_batch_processor.assert_called_once()
            m_tracer_provider.assert_called_once()
            m_set_tracer.assert_called_once_with(mock_provider)
        finally:
            # Clean up
            if "exporter.ar_trace.trace_exporter" in sys.modules:
                del sys.modules["exporter.ar_trace.trace_exporter"]
            if "exporter.public.client" in sys.modules:
                del sys.modules["exporter.public.client"]
            if "exporter.public.public" in sys.modules:
                del sys.modules["exporter.public.public"]

    @patch("app.utils.observability.observability_trace.TELEMETRY_SDK_AVAILABLE", True)
    @patch("app.utils.observability.observability_trace.set_service_info")
    @patch("app.utils.observability.observability_trace.trace_resource")
    @patch("app.common.config.Config")
    def test_http_exporter_initialization(
        self, m_config, m_trace_resource, m_set_service
    ):
        """测试使用HTTP导出器初始化"""
        m_config.is_o11y_trace_enabled.return_value = True
        m_trace_resource.return_value = MagicMock()

        import sys

        sys.modules["exporter.ar_trace.trace_exporter"] = MagicMock(
            ARTraceExporter=MagicMock(return_value=MagicMock())
        )
        sys.modules["exporter.public.client"] = MagicMock(
            HTTPClient=MagicMock(return_value=MagicMock())
        )
        sys.modules["exporter.public.public"] = MagicMock(
            WithAnyRobotURL=MagicMock(return_value="http://test.url")
        )

        try:
            from app.utils.observability.observability_trace import init_trace_provider
            from app.utils.observability.observability_setting import (
                ServerInfo,
                TraceSetting,
            )

            server_info = ServerInfo(server_name="test_service", server_version="1.0")
            setting = TraceSetting(
                trace_enabled=True,
                trace_provider="http",
                http_trace_feed_ingester_url="http://test.url",
            )

            init_trace_provider(server_info, setting)

            m_set_service.assert_called_once()
        finally:
            del sys.modules["exporter.ar_trace.trace_exporter"]
            del sys.modules["exporter.public.client"]
            del sys.modules["exporter.public.public"]
