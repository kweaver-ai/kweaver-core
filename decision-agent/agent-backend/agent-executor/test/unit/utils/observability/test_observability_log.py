"""单元测试 - utils/observability/observability_log 模块"""

import pytest
from unittest.mock import MagicMock, patch


class TestNullLogger:
    """测试 NullLogger 类"""

    def test_null_logger_info(self):
        """测试info方法是空操作"""
        from app.utils.observability.observability_log import NullLogger

        logger = NullLogger()
        logger.info("test message")  # Should not raise

    def test_null_logger_error(self):
        """测试error方法是空操作"""
        from app.utils.observability.observability_log import NullLogger

        logger = NullLogger()
        logger.error("test error")  # Should not raise

    def test_null_logger_warn(self):
        """测试warn方法是空操作"""
        from app.utils.observability.observability_log import NullLogger

        logger = NullLogger()
        logger.warn("test warning")  # Should not raise

    def test_null_logger_debug(self):
        """测试debug方法是空操作"""
        from app.utils.observability.observability_log import NullLogger

        logger = NullLogger()
        logger.debug("test debug")  # Should not raise

    def test_null_logger_fatal(self):
        """测试fatal方法是空操作"""
        from app.utils.observability.observability_log import NullLogger

        logger = NullLogger()
        logger.fatal("test fatal")  # Should not raise, not exit

    def test_null_logger_trace(self):
        """测试trace方法是空操作"""
        from app.utils.observability.observability_log import NullLogger

        logger = NullLogger()
        logger.trace("test trace")  # Should not raise

    def test_null_logger_set_level(self):
        """测试set_level方法是空操作"""
        from app.utils.observability.observability_log import NullLogger

        logger = NullLogger()
        logger.set_level("info")  # Should not raise

    def test_null_logger_get_level(self):
        """测试get_level返回0"""
        from app.utils.observability.observability_log import NullLogger

        logger = NullLogger()
        assert logger.get_level() == 0

    def test_null_logger_set_exporters(self):
        """测试set_exporters方法是空操作"""
        from app.utils.observability.observability_log import NullLogger

        logger = NullLogger()
        logger.set_exporters(MagicMock())  # Should not raise

    def test_null_logger_shutdown(self):
        """测试shutdown方法是空操作"""
        from app.utils.observability.observability_log import NullLogger

        logger = NullLogger()
        logger.shutdown()  # Should not raise


class TestGetCallerInfo:
    """测试 get_caller_info 函数"""

    def test_returns_caller_information(self):
        """测试返回调用者信息"""
        from app.utils.observability.observability_log import get_caller_info

        result = get_caller_info()

        assert isinstance(result, str)
        assert ":" in result  # Should contain file:line:function format


class TestLoggingFunctions:
    """测试日志记录函数"""

    @patch("app.utils.observability.observability_log.logger")
    def test_info_function(self, m_logger):
        """测试info函数"""
        from app.utils.observability.observability_log import info

        info("test message")

        m_logger.info.assert_called_once()

    @patch("app.utils.observability.observability_log.logger")
    def test_error_function(self, m_logger):
        """测试error函数"""
        from app.utils.observability.observability_log import error

        error("test error")

        m_logger.error.assert_called_once()

    @patch("app.utils.observability.observability_log.logger")
    def test_warn_function(self, m_logger):
        """测试warn函数"""
        from app.utils.observability.observability_log import warn

        warn("test warning")

        m_logger.warn.assert_called_once()

    @patch("app.utils.observability.observability_log.logger")
    def test_debug_function(self, m_logger):
        """测试debug函数"""
        from app.utils.observability.observability_log import debug

        debug("test debug")

        m_logger.debug.assert_called_once()

    @patch("app.utils.observability.observability_log.logger")
    def test_trace_function(self, m_logger):
        """测试trace函数"""
        from app.utils.observability.observability_log import info

        info("test trace")

        m_logger.info.assert_called_once()

    @patch("app.utils.observability.observability_log.logger", None)
    def test_info_when_logger_is_none(self):
        """测试logger为None时不调用"""
        from app.utils.observability.observability_log import info

        # Should not raise
        info("test message")

    @patch("app.utils.observability.observability_log.logger", None)
    def test_error_when_logger_is_none(self):
        """测试logger为None时不调用"""
        from app.utils.observability.observability_log import error

        # Should not raise
        error("test error")

    @patch("app.utils.observability.observability_log.logger", None)
    def test_warn_when_logger_is_none(self):
        """测试logger为None时不调用"""
        from app.utils.observability.observability_log import warn

        # Should not raise
        warn("test warning")

    @patch("app.utils.observability.observability_log.logger", None)
    def test_debug_when_logger_is_none(self):
        """测试logger为None时不调用"""
        from app.utils.observability.observability_log import debug

        # Should not raise
        debug("test debug")

    @patch("app.utils.observability.observability_log.logger", None)
    def test_fatal_when_logger_is_none_exits(self):
        """测试logger为None时fatal会退出"""
        from app.utils.observability.observability_log import fatal
        import pytest

        with pytest.raises(SystemExit):
            fatal("test fatal")

    @patch("app.utils.observability.observability_log.logger")
    def test_fatal_exits_after_logging(self, m_logger):
        """测试fatal在日志后退出"""
        from app.utils.observability.observability_log import fatal

        with pytest.raises(SystemExit):
            fatal("test fatal")

        m_logger.fatal.assert_called_once()


class TestInitLogProvider:
    """测试 init_log_provider 函数"""

    @patch("app.utils.observability.observability_log.TELEMETRY_SDK_AVAILABLE", False)
    def test_returns_early_when_sdk_unavailable(self):
        """测试SDK不可用时直接返回"""
        from app.utils.observability.observability_log import init_log_provider
        from app.utils.observability.observability_setting import ServerInfo, LogSetting

        server_info = ServerInfo(server_name="test", server_version="1.0")
        setting = LogSetting(log_enabled=True)

        # Should not raise any error
        init_log_provider(server_info, setting)

    @patch("app.utils.observability.observability_log.TELEMETRY_SDK_AVAILABLE", True)
    @patch("app.utils.observability.observability_log.set_service_info")
    @patch("app.utils.observability.observability_log.SamplerLogger")
    @patch("app.utils.observability.observability_log.log_resource")
    @patch("app.common.config.Config")
    def test_console_exporter_initialization(
        self, m_config, m_log_resource, m_sampler_logger, m_set_service
    ):
        """测试使用console导出器初始化"""
        m_config.is_o11y_log_enabled.return_value = True
        m_log_resource.return_value = MagicMock()

        mock_logger = MagicMock()
        m_sampler_logger.return_value = mock_logger

        from app.utils.observability.observability_log import init_log_provider
        from app.utils.observability.observability_setting import ServerInfo, LogSetting

        server_info = ServerInfo(server_name="test_service", server_version="1.0")
        setting = LogSetting(log_enabled=True, log_exporter="console")

        init_log_provider(server_info, setting)

        m_set_service.assert_called_once()
        m_sampler_logger.assert_called()
        mock_logger.set_exporters.assert_called_once_with()

    @patch("app.utils.observability.observability_log.TELEMETRY_SDK_AVAILABLE", True)
    @patch("app.utils.observability.observability_log.set_service_info")
    @patch("app.common.config.Config")
    def test_returns_when_log_disabled_in_config(self, m_config, m_set_service):
        """测试配置中禁用日志时直接返回"""
        m_config.is_o11y_log_enabled.return_value = False
        from app.utils.observability.observability_log import init_log_provider
        from app.utils.observability.observability_setting import ServerInfo, LogSetting

        server_info = ServerInfo(server_name="test", server_version="1.0")
        setting = LogSetting(log_enabled=True)

        init_log_provider(server_info, setting)

        m_set_service.assert_called_once()

    @patch("app.utils.observability.observability_log.TELEMETRY_SDK_AVAILABLE", True)
    @patch("app.utils.observability.observability_log.set_service_info")
    @patch("app.utils.observability.observability_log.SamplerLogger")
    @patch("app.utils.observability.observability_log.log_resource")
    @patch("app.common.config.Config")
    def test_http_exporter_initialization(
        self, m_config, m_log_resource, m_sampler_logger, m_set_service
    ):
        """测试使用HTTP导出器初始化"""
        m_config.is_o11y_log_enabled.return_value = True
        m_log_resource.return_value = MagicMock()

        import sys

        sys.modules["exporter.ar_log.log_exporter"] = MagicMock(
            ARLogExporter=MagicMock(return_value=MagicMock())
        )
        sys.modules["exporter.public.client"] = MagicMock(
            HTTPClient=MagicMock(return_value=MagicMock())
        )
        sys.modules["exporter.public.public"] = MagicMock(
            WithAnyRobotURL=MagicMock(return_value="http://test.url")
        )

        try:
            mock_logger = MagicMock()
            m_sampler_logger.return_value = mock_logger

            from app.utils.observability.observability_log import init_log_provider
            from app.utils.observability.observability_setting import (
                ServerInfo,
                LogSetting,
            )

            server_info = ServerInfo(server_name="test_service", server_version="1.0")
            setting = LogSetting(
                log_enabled=True,
                log_exporter="http",
                http_log_feed_ingester_url="http://test.url",
            )

            init_log_provider(server_info, setting)

            m_set_service.assert_called_once()
            mock_logger.set_exporters.assert_called_once()
        finally:
            del sys.modules["exporter.ar_log.log_exporter"]
            del sys.modules["exporter.public.client"]
            del sys.modules["exporter.public.public"]


class TestGetLogger:
    """测试 get_logger 函数"""

    def test_returns_logger_instance(self):
        """测试返回logger实例"""
        from app.utils.observability.observability_log import get_logger

        logger = get_logger()

        assert logger is not None
        # Should have expected methods
        assert hasattr(logger, "info")
        assert hasattr(logger, "error")
        assert hasattr(logger, "warn")
        assert hasattr(logger, "debug")

    @patch("app.utils.observability.observability_log.logger", None)
    def test_creates_null_logger_when_none(self):
        """测试logger为None时创建NullLogger"""
        from app.utils.observability.observability_log import get_logger

        logger = get_logger()

        assert logger is not None
        # Should be NullLogger which doesn't crash
        logger.info("test")
        logger.error("test")


class TestShutdownLogProvider:
    """测试 shutdown_log_provider 函数"""

    @patch("app.utils.observability.observability_log.logger", None)
    def test_returns_when_logger_is_none(self):
        """测试logger为None时直接返回"""
        from app.utils.observability.observability_log import shutdown_log_provider

        shutdown_log_provider()  # Should not raise

    @patch("app.utils.observability.observability_log.logger")
    def test_calls_shutdown_on_logger(self, m_logger):
        """测试调用logger的shutdown方法"""
        from app.utils.observability.observability_log import shutdown_log_provider

        shutdown_log_provider()

        m_logger.shutdown.assert_called_once()
