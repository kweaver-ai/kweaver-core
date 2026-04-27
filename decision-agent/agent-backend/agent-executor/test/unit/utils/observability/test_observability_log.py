"""Unit tests for app/utils/observability/observability_log.py."""

from unittest.mock import MagicMock, patch

import pytest


class TestNullLogger:
    def test_null_logger_methods_are_noop(self):
        from app.utils.observability.observability_log import NullLogger

        logger = NullLogger()
        logger.info("info")
        logger.error("error")
        logger.warn("warn")
        logger.debug("debug")
        logger.trace("trace")
        logger.shutdown()


class TestStandardOtelLogger:
    def test_info_emits_log_record(self):
        from app.utils.observability.observability_log import StandardOtelLogger

        otel_logger = MagicMock()
        provider = MagicMock()
        logger = StandardOtelLogger(otel_logger, provider, level="info")

        logger.info("hello", attributes={"foo": "bar"})

        otel_logger.emit.assert_called_once()
        kwargs = otel_logger.emit.call_args.kwargs
        assert kwargs["body"] == "hello"
        assert kwargs["severity_text"] == "INFO"
        assert kwargs["attributes"]["foo"] == "bar"

    def test_debug_respects_log_level(self):
        from app.utils.observability.observability_log import StandardOtelLogger

        otel_logger = MagicMock()
        provider = MagicMock()
        logger = StandardOtelLogger(otel_logger, provider, level="warn")

        logger.debug("skip me")

        otel_logger.emit.assert_not_called()


class TestInitLogProvider:
    @patch("app.utils.observability.observability_log._build_standard_logger")
    def test_init_log_provider_uses_standard_logger(self, mock_build_standard):
        from app.utils.observability.observability_log import init_log_provider
        from app.utils.observability.observability_setting import LogSetting, ServerInfo

        server_info = ServerInfo(server_name="test", server_version="1.0")
        setting = LogSetting(log_enabled=True, log_level="info")
        mock_build_standard.return_value = MagicMock()

        init_log_provider(server_info, setting)

        mock_build_standard.assert_called_once_with(server_info, setting)

    def test_init_log_provider_keeps_null_logger_when_disabled(self):
        from app.utils.observability.observability_log import (
            NullLogger,
            get_logger,
            init_log_provider,
        )
        from app.utils.observability.observability_setting import LogSetting, ServerInfo

        server_info = ServerInfo(server_name="test", server_version="1.0")
        init_log_provider(server_info, LogSetting(log_enabled=False))

        assert isinstance(get_logger(), NullLogger)


class TestShutdownLogProvider:
    @patch("app.utils.observability.observability_log.logger")
    def test_shutdown_calls_logger_shutdown(self, mock_logger):
        from app.utils.observability.observability_log import shutdown_log_provider

        shutdown_log_provider()

        mock_logger.shutdown.assert_called_once()
