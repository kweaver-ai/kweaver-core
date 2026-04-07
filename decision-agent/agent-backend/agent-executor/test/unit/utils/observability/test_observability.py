"""单元测试 - utils/observability/observability 模块"""

from unittest.mock import patch


class TestInitObservability:
    """测试 init_observability 函数"""

    @patch("app.utils.observability.observability.init_log_provider")
    @patch("app.utils.observability.observability.init_trace_provider")
    def test_init_with_log_enabled(self, mock_init_trace, mock_init_log):
        """测试启用日志时初始化"""
        from app.utils.observability.observability_setting import (
            ObservabilitySetting,
            LogSetting,
            TraceSetting,
            ServerInfo,
        )
        from app.utils.observability.observability import init_observability

        server_info = ServerInfo(server_name="test", server_version="1.0")
        setting = ObservabilitySetting(
            log=LogSetting(log_enabled=True), trace=TraceSetting(trace_enabled=False)
        )

        init_observability(server_info, setting)

        mock_init_log.assert_called_once_with(server_info, setting.log)
        mock_init_trace.assert_not_called()

    @patch("app.utils.observability.observability.init_log_provider")
    @patch("app.utils.observability.observability.init_trace_provider")
    def test_init_with_trace_enabled(self, mock_init_trace, mock_init_log):
        """测试启用追踪时初始化"""
        from app.utils.observability.observability_setting import (
            ObservabilitySetting,
            LogSetting,
            TraceSetting,
            ServerInfo,
        )
        from app.utils.observability.observability import init_observability

        server_info = ServerInfo(server_name="test", server_version="1.0")
        setting = ObservabilitySetting(
            log=LogSetting(log_enabled=False), trace=TraceSetting(trace_enabled=True)
        )

        init_observability(server_info, setting)

        mock_init_log.assert_not_called()
        mock_init_trace.assert_called_once_with(server_info, setting.trace)

    @patch("app.utils.observability.observability.init_log_provider")
    @patch("app.utils.observability.observability.init_trace_provider")
    def test_init_with_both_enabled(self, mock_init_trace, mock_init_log):
        """测试同时启用日志和追踪"""
        from app.utils.observability.observability_setting import (
            ObservabilitySetting,
            LogSetting,
            TraceSetting,
            ServerInfo,
        )
        from app.utils.observability.observability import init_observability

        server_info = ServerInfo(server_name="test", server_version="1.0")
        setting = ObservabilitySetting(
            log=LogSetting(log_enabled=True), trace=TraceSetting(trace_enabled=True)
        )

        init_observability(server_info, setting)

        mock_init_log.assert_called_once_with(server_info, setting.log)
        mock_init_trace.assert_called_once_with(server_info, setting.trace)

    @patch("app.utils.observability.observability.init_log_provider")
    @patch("app.utils.observability.observability.init_trace_provider")
    def test_init_with_none_enabled(self, mock_init_trace, mock_init_log):
        """测试都禁用时不初始化"""
        from app.utils.observability.observability_setting import (
            ObservabilitySetting,
            LogSetting,
            TraceSetting,
            ServerInfo,
        )
        from app.utils.observability.observability import init_observability

        server_info = ServerInfo(server_name="test", server_version="1.0")
        setting = ObservabilitySetting(
            log=LogSetting(log_enabled=False), trace=TraceSetting(trace_enabled=False)
        )

        init_observability(server_info, setting)

        mock_init_log.assert_not_called()
        mock_init_trace.assert_not_called()


class TestShutdownObservability:
    """测试 shutdown_observability 函数"""

    @patch("app.utils.observability.observability.shutdown_log_provider")
    def test_shutdown_calls_shutdown_log_provider(self, mock_shutdown_log):
        """测试关闭时调用shutdown_log_provider"""
        from app.utils.observability.observability import shutdown_observability

        shutdown_observability()

        mock_shutdown_log.assert_called_once()
