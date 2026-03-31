# -*- coding: utf-8 -*-
"""
Unit tests for app/common/struct_logger/file_logging_setup module
"""

from unittest.mock import MagicMock, patch
import logging

import pytest


class TestSetupFileLogging:
    """Tests for setup_file_logging function"""

    def setup_method(self):
        """Clear handlers before each test"""
        stdlib_logger = logging.getLogger("agent-executor-file")
        stdlib_logger.handlers.clear()

    def teardown_method(self):
        """Clear handlers after each test to prevent pollution"""
        stdlib_logger = logging.getLogger("agent-executor-file")
        stdlib_logger.handlers.clear()

    @pytest.mark.asyncio
    async def test_setup_file_logging_returns_logger(self):
        """Test that setup_file_logging returns a logger"""
        with patch("app.common.struct_logger.file_logging_setup.Config") as mock_config:
            mock_config.app.get_stdlib_log_level.return_value = logging.INFO

            from app.common.struct_logger.file_logging_setup import setup_file_logging

            logger = setup_file_logging()

            assert logger is not None
            # Should be a structlog BoundLogger
            assert hasattr(logger, "info")
            assert hasattr(logger, "error")
            assert hasattr(logger, "debug")

    @pytest.mark.asyncio
    async def test_setup_file_logging_creates_handler(self):
        """Test that setup_file_logging creates TimedRotatingFileHandler"""
        # Patch at the module level before importing
        with patch("app.common.struct_logger.file_logging_setup.Config") as mock_config:
            mock_config.app.get_stdlib_log_level.return_value = logging.INFO

            with patch(
                "app.common.struct_logger.file_logging_setup.TimedRotatingFileHandler"
            ) as mock_handler:
                mock_handler_instance = MagicMock()
                mock_handler_instance.setLevel = MagicMock()
                mock_handler_instance.setFormatter = MagicMock()
                mock_handler.return_value = mock_handler_instance

                from app.common.struct_logger.file_logging_setup import (
                    setup_file_logging,
                )

                logger = setup_file_logging()

                # Verify TimedRotatingFileHandler was called
                assert mock_handler.called
                call_args = mock_handler.call_args
                assert call_args[0][0] == "log/agent-executor.log"
                assert call_args[1]["when"] == "midnight"
                assert call_args[1]["interval"] == 1
                assert call_args[1]["backupCount"] == 30

    @pytest.mark.asyncio
    async def test_setup_file_logging_log_level(self):
        """Test that setup_file_logging respects log level from Config"""
        with patch("app.common.struct_logger.file_logging_setup.Config") as mock_config:
            mock_config.app.get_stdlib_log_level.return_value = logging.DEBUG

            from app.common.struct_logger.file_logging_setup import setup_file_logging

            logger = setup_file_logging()

            # Get the logger to verify level
            stdlib_logger = logging.getLogger("agent-executor-file")
            assert stdlib_logger.level == logging.DEBUG

    @pytest.mark.asyncio
    async def test_setup_file_logging_no_propagate(self):
        """Test that file logger doesn't propagate"""
        with patch("app.common.struct_logger.file_logging_setup.Config") as mock_config:
            mock_config.app.get_stdlib_log_level.return_value = logging.INFO

            from app.common.struct_logger.file_logging_setup import setup_file_logging

            logger = setup_file_logging()

            # Get the logger to verify propagate setting
            stdlib_logger = logging.getLogger("agent-executor-file")
            assert stdlib_logger.propagate is False

    @pytest.mark.asyncio
    async def test_setup_file_logging_logger_name(self):
        """Test that setup_file_logging creates logger with correct name"""
        with patch("app.common.struct_logger.file_logging_setup.Config") as mock_config:
            mock_config.app.get_stdlib_log_level.return_value = logging.INFO

            from app.common.struct_logger.file_logging_setup import setup_file_logging

            logger = setup_file_logging()

            # Get the logger to verify name
            stdlib_logger = logging.getLogger("agent-executor-file")
            assert stdlib_logger.name == "agent-executor-file"

    @pytest.mark.asyncio
    async def test_setup_file_logging_handler_level(self):
        """Test that file handler is set to NOTSET level"""
        with patch("app.common.struct_logger.file_logging_setup.Config") as mock_config:
            mock_config.app.get_stdlib_log_level.return_value = logging.INFO

            with patch(
                "app.common.struct_logger.file_logging_setup.TimedRotatingFileHandler"
            ) as mock_handler:
                mock_handler_instance = MagicMock()
                mock_handler_instance.setLevel = MagicMock()
                mock_handler_instance.setFormatter = MagicMock()
                mock_handler.return_value = mock_handler_instance

                from app.common.struct_logger.file_logging_setup import (
                    setup_file_logging,
                )

                logger = setup_file_logging()

                # Handler should be set to NOTSET (logger controls filtering)
                mock_handler_instance.setLevel.assert_called_with(logging.NOTSET)


class TestModuleImports:
    """Tests for module imports"""

    @pytest.mark.asyncio
    async def test_module_imports(self):
        """Test that file_logging_setup module can be imported"""
        from app.common.struct_logger import file_logging_setup

        assert file_logging_setup is not None
        assert hasattr(file_logging_setup, "setup_file_logging")
