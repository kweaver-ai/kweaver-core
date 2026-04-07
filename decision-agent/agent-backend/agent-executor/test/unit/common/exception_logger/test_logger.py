# -*- coding: utf-8 -*-
"""
Unit tests for app/common/exception_logger/logger.py module
"""

import tempfile
from unittest.mock import MagicMock, patch
from datetime import datetime

import pytest


class TestExceptionLogger:
    """Tests for ExceptionLogger class"""

    @pytest.fixture
    def mock_config(self):
        """Mock Config with is_write_exception_log_to_file attribute"""
        mock_config = MagicMock()
        mock_config.app.is_write_exception_log_to_file = False
        return mock_config

    @pytest.fixture
    def temp_log_dir(self):
        """Create a temporary directory for logs"""
        with tempfile.TemporaryDirectory() as tmpdir:
            yield tmpdir

    @pytest.mark.asyncio
    async def test_singleton_pattern(self, mock_config, temp_log_dir):
        """Test ExceptionLogger implements singleton pattern"""
        with patch("app.common.exception_logger.logger.Config", mock_config):
            with patch(
                "app.common.exception_logger.constants.EXCEPTION_LOG_DIR", temp_log_dir
            ):
                from app.common.exception_logger.logger import ExceptionLogger

                logger1 = ExceptionLogger()
                logger2 = ExceptionLogger()

                assert logger1 is logger2

    @pytest.mark.asyncio
    async def test_initialization(self, mock_config, temp_log_dir):
        """Test ExceptionLogger initialization"""
        with patch("app.common.exception_logger.logger.Config", mock_config):
            with patch(
                "app.common.exception_logger.constants.EXCEPTION_LOG_DIR", temp_log_dir
            ):
                from app.common.exception_logger.logger import ExceptionLogger

                logger = ExceptionLogger()

                assert logger._initialized is True
                assert hasattr(logger, "_logger")
                assert hasattr(logger, "_simple_logger")
                assert hasattr(logger, "_detailed_logger")

    @pytest.mark.asyncio
    async def test_log_error_basic(self, mock_config, temp_log_dir):
        """Test log_error with basic exception"""
        with patch("app.common.exception_logger.logger.Config", mock_config):
            with patch(
                "app.common.exception_logger.constants.EXCEPTION_LOG_DIR", temp_log_dir
            ):
                from app.common.exception_logger.logger import ExceptionLogger

                logger = ExceptionLogger()

                # Mock the formatters
                with patch(
                    "app.common.exception_logger.logger.format_error_console"
                ) as mock_console:
                    mock_console.return_value = "Console error message"

                    exc = ValueError("Test error")
                    logger.log_error(exc)

                    mock_console.assert_called_once()

    @pytest.mark.asyncio
    async def test_log_error_with_request_info(self, mock_config, temp_log_dir):
        """Test log_error with request information"""
        with patch("app.common.exception_logger.logger.Config", mock_config):
            with patch(
                "app.common.exception_logger.constants.EXCEPTION_LOG_DIR", temp_log_dir
            ):
                from app.common.exception_logger.logger import ExceptionLogger

                logger = ExceptionLogger()

                with patch(
                    "app.common.exception_logger.logger.format_error_console"
                ) as mock_console:
                    mock_console.return_value = "Console error message"

                    exc = ValueError("Test error")
                    request_info = {"method": "GET", "path": "/api/test"}
                    logger.log_error(exc, request_info=request_info)

                    mock_console.assert_called_once()
                    # Check that request_info was passed to formatter
                    call_args = mock_console.call_args
                    assert call_args[0][2] == request_info

    @pytest.mark.asyncio
    async def test_log_error_with_timestamp(self, mock_config, temp_log_dir):
        """Test log_error with custom timestamp"""
        with patch("app.common.exception_logger.logger.Config", mock_config):
            with patch(
                "app.common.exception_logger.constants.EXCEPTION_LOG_DIR", temp_log_dir
            ):
                from app.common.exception_logger.logger import ExceptionLogger

                logger = ExceptionLogger()

                with patch(
                    "app.common.exception_logger.logger.format_error_console"
                ) as mock_console:
                    mock_console.return_value = "Console error message"

                    exc = ValueError("Test error")
                    timestamp = datetime(2024, 1, 1, 12, 0, 0)
                    logger.log_error(exc, timestamp=timestamp)

                    mock_console.assert_called_once()
                    # The timestamp is passed as a positional argument
                    # Verify the mock was called with the correct timestamp somewhere in args
                    call_args_str = str(mock_console.call_args)
                    assert "2024" in call_args_str

    @pytest.mark.asyncio
    async def test_log_error_writes_to_file_when_enabled(self, temp_log_dir):
        """Test log_error writes to file when config enabled"""
        mock_config = MagicMock()
        mock_config.app.is_write_exception_log_to_file = True

        with patch("app.common.exception_logger.logger.Config", mock_config):
            with patch(
                "app.common.exception_logger.constants.EXCEPTION_LOG_DIR", temp_log_dir
            ):
                from app.common.exception_logger.logger import ExceptionLogger

                logger = ExceptionLogger()

                with patch(
                    "app.common.exception_logger.logger.format_error_console"
                ) as mock_console:
                    with patch(
                        "app.common.exception_logger.logger.format_error_file_simple"
                    ) as mock_simple:
                        with patch(
                            "app.common.exception_logger.logger.format_error_file_detailed"
                        ) as mock_detailed:
                            mock_console.return_value = "Console"
                            mock_simple.return_value = "Simple"
                            mock_detailed.return_value = "Detailed"

                            exc = ValueError("Test error")
                            logger.log_error(exc)

                            mock_simple.assert_called_once()
                            mock_detailed.assert_called_once()

    @pytest.mark.asyncio
    async def test_log_exception_group(self, mock_config, temp_log_dir):
        """Test log_exception_group with multiple exceptions"""
        with patch("app.common.exception_logger.logger.Config", mock_config):
            with patch(
                "app.common.exception_logger.constants.EXCEPTION_LOG_DIR", temp_log_dir
            ):
                from app.common.exception_logger.logger import ExceptionLogger

                logger = ExceptionLogger()

                with patch(
                    "app.common.exception_logger.logger.format_error_console"
                ) as mock_console:
                    with patch(
                        "app.common.exception_logger.logger.format_multiple_errors_separator"
                    ) as mock_sep:
                        mock_console.return_value = "Error message"
                        mock_sep.return_value = "Separator"

                        exceptions = [ValueError("Error 1"), TypeError("Error 2")]
                        logger.log_exception_group(exceptions)

                        # Should be called twice (once for each exception)
                        assert mock_console.call_count == 2

    @pytest.mark.asyncio
    async def test_log_exception_with_exception_group(self, mock_config, temp_log_dir):
        """Test log_exception handles ExceptionGroup"""
        with patch("app.common.exception_logger.logger.Config", mock_config):
            with patch(
                "app.common.exception_logger.constants.EXCEPTION_LOG_DIR", temp_log_dir
            ):
                from app.common.exception_logger.logger import ExceptionLogger

                logger = ExceptionLogger()

                with patch.object(logger, "log_exception_group") as mock_group:
                    with patch.object(logger, "log_error") as mock_error:
                        # Create a mock ExceptionGroup
                        mock_exception_group = MagicMock()
                        mock_exception_group.__class__.__name__ = "ExceptionGroup"
                        mock_exception_group.exceptions = [
                            ValueError("Error 1"),
                            TypeError("Error 2"),
                        ]

                        logger.log_exception(mock_exception_group)

                        mock_group.assert_called_once()

    @pytest.mark.asyncio
    async def test_log_exception_with_regular_exception(
        self, mock_config, temp_log_dir
    ):
        """Test log_exception handles regular exceptions"""
        with patch("app.common.exception_logger.logger.Config", mock_config):
            with patch(
                "app.common.exception_logger.constants.EXCEPTION_LOG_DIR", temp_log_dir
            ):
                from app.common.exception_logger.logger import ExceptionLogger

                logger = ExceptionLogger()

                with patch.object(logger, "log_error") as mock_error:
                    exc = ValueError("Regular error")
                    logger.log_exception(exc)

                    mock_error.assert_called_once_with(exc, None, None)

    @pytest.mark.asyncio
    async def test_log_exception_defaults_to_now(self, mock_config, temp_log_dir):
        """Test log_error uses current time when no timestamp provided"""
        with patch("app.common.exception_logger.logger.Config", mock_config):
            with patch(
                "app.common.exception_logger.constants.EXCEPTION_LOG_DIR", temp_log_dir
            ):
                from app.common.exception_logger.logger import ExceptionLogger

                logger = ExceptionLogger()

                with patch(
                    "app.common.exception_logger.logger.format_error_console"
                ) as mock_console:
                    mock_console.return_value = "Error message"

                    with patch(
                        "app.common.exception_logger.logger.datetime"
                    ) as mock_datetime:
                        mock_now = datetime(2024, 12, 25, 10, 30, 0)
                        mock_datetime.now.return_value = mock_now

                        exc = ValueError("Test error")
                        logger.log_error(exc)

                        # Check that datetime.now() was called
                        mock_datetime.now.assert_called()


class TestGlobalExceptionLogger:
    """Tests for the global exception_logger instance"""

    @pytest.mark.asyncio
    async def test_global_exception_logger_exists(self):
        """Test that global exception_logger instance exists"""
        with patch("app.common.exception_logger.logger.Config") as mock_config:
            mock_config.return_value.app.is_write_exception_log_to_file = False

            with patch(
                "app.common.exception_logger.constants.EXCEPTION_LOG_DIR",
                "/tmp/test_logs",
            ):
                # Force reimport to get a fresh instance
                import importlib
                from app.common.exception_logger import logger

                importlib.reload(logger)

                from app.common.exception_logger.logger import exception_logger

                assert exception_logger is not None
                assert isinstance(exception_logger, logger.ExceptionLogger)
