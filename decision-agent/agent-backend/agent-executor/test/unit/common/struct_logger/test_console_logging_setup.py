# -*- coding: utf-8 -*-
"""
Unit tests for app/common/struct_logger/console_logging_setup module
"""

from unittest.mock import patch
import logging

import pytest


class TestCreateStdoutHandler:
    """Tests for _create_stdout_handler function"""

    @pytest.mark.asyncio
    async def test_create_stdout_handler(self):
        """Test stdout handler creation"""
        from app.common.struct_logger.console_logging_setup import (
            _create_stdout_handler,
        )

        handler = _create_stdout_handler()

        assert isinstance(handler, logging.StreamHandler)
        # Verify it's a stdout handler by checking the stream
        assert handler.stream is not None

    @pytest.mark.asyncio
    async def test_stdout_handler_has_filter(self):
        """Test that stdout handler has filter"""
        from app.common.struct_logger.console_logging_setup import (
            _create_stdout_handler,
        )

        handler = _create_stdout_handler()

        assert len(handler.filters) == 1
        filter_obj = handler.filters[0]

        # Filter should allow INFO and below
        debug_record = logging.LogRecord(
            name="test",
            level=logging.DEBUG,
            pathname="test.py",
            lineno=1,
            msg="test",
            args=(),
            exc_info=None,
        )
        assert filter_obj.filter(debug_record) is True  # DEBUG (10) <= INFO (20)

        warning_record = logging.LogRecord(
            name="test",
            level=logging.WARNING,
            pathname="test.py",
            lineno=1,
            msg="test",
            args=(),
            exc_info=None,
        )
        assert filter_obj.filter(warning_record) is False  # WARNING (30) > INFO (20)


class TestCreateStderrHandler:
    """Tests for _create_stderr_handler function"""

    @pytest.mark.asyncio
    async def test_create_stderr_handler(self):
        """Test stderr handler creation"""
        from app.common.struct_logger.console_logging_setup import (
            _create_stderr_handler,
        )

        handler = _create_stderr_handler()

        assert isinstance(handler, logging.StreamHandler)
        # Verify it's a stderr handler by checking the stream
        assert handler.stream is not None

    @pytest.mark.asyncio
    async def test_stderr_handler_has_filter(self):
        """Test that stderr handler has filter"""
        from app.common.struct_logger.console_logging_setup import (
            _create_stderr_handler,
        )

        handler = _create_stderr_handler()

        assert len(handler.filters) == 1
        filter_obj = handler.filters[0]

        # Filter should allow WARNING and above
        info_record = logging.LogRecord(
            name="test",
            level=logging.INFO,
            pathname="test.py",
            lineno=1,
            msg="test",
            args=(),
            exc_info=None,
        )
        assert filter_obj.filter(info_record) is False  # INFO (20) < WARNING (30)

        warning_record = logging.LogRecord(
            name="test",
            level=logging.WARNING,
            pathname="test.py",
            lineno=1,
            msg="test",
            args=(),
            exc_info=None,
        )
        assert filter_obj.filter(warning_record) is True  # WARNING (30) >= WARNING (30)


class TestSetupConsoleLogging:
    """Tests for setup_console_logging function"""

    def setup_method(self):
        """Clear handlers before each test"""
        stdlib_logger = logging.getLogger("agent-executor-console")
        stdlib_logger.handlers.clear()

    def teardown_method(self):
        """Clear handlers after each test to prevent pollution"""
        stdlib_logger = logging.getLogger("agent-executor-console")
        stdlib_logger.handlers.clear()

    @pytest.mark.asyncio
    async def test_setup_console_logging_returns_logger(self):
        """Test that setup_console_logging returns a logger"""
        with patch(
            "app.common.struct_logger.console_logging_setup.Config"
        ) as mock_config:
            mock_config.app.get_stdlib_log_level.return_value = logging.INFO

            from app.common.struct_logger.console_logging_setup import (
                setup_console_logging,
            )

            logger = setup_console_logging()

            assert logger is not None
            # Should be a structlog BoundLogger
            assert hasattr(logger, "info")
            assert hasattr(logger, "error")
            assert hasattr(logger, "debug")

    @pytest.mark.asyncio
    async def test_setup_console_logging_creates_logger(self):
        """Test that setup_console_logging creates correct logger"""
        # Clear any existing handlers first
        import logging

        stdlib_logger = logging.getLogger("agent-executor-console")
        stdlib_logger.handlers.clear()

        with patch(
            "app.common.struct_logger.console_logging_setup.Config"
        ) as mock_config:
            mock_config.app.get_stdlib_log_level.return_value = logging.INFO

            from app.common.struct_logger.console_logging_setup import (
                setup_console_logging,
            )

            logger = setup_console_logging()

            # Check that the logger was created
            stdlib_logger = logging.getLogger("agent-executor-console")
            assert stdlib_logger is not None
            # Should have at least stdout and stderr handlers
            assert len(stdlib_logger.handlers) >= 2

    @pytest.mark.asyncio
    async def test_setup_console_logging_log_level(self):
        """Test that setup_console_logging respects log level"""
        # Clear any existing handlers first
        import logging

        stdlib_logger = logging.getLogger("agent-executor-console")
        stdlib_logger.handlers.clear()

        with patch(
            "app.common.struct_logger.console_logging_setup.Config"
        ) as mock_config:
            mock_config.app.get_stdlib_log_level.return_value = logging.DEBUG

            from app.common.struct_logger.console_logging_setup import (
                setup_console_logging,
            )

            logger = setup_console_logging()

            stdlib_logger = logging.getLogger("agent-executor-console")
            assert stdlib_logger.level == logging.DEBUG

    @pytest.mark.asyncio
    async def test_setup_console_logging_no_propagate(self):
        """Test that console logger doesn't propagate"""
        # Clear any existing handlers first
        import logging

        stdlib_logger = logging.getLogger("agent-executor-console")
        stdlib_logger.handlers.clear()

        with patch(
            "app.common.struct_logger.console_logging_setup.Config"
        ) as mock_config:
            mock_config.app.get_stdlib_log_level.return_value = logging.INFO

            from app.common.struct_logger.console_logging_setup import (
                setup_console_logging,
            )

            logger = setup_console_logging()

            stdlib_logger = logging.getLogger("agent-executor-console")
            assert stdlib_logger.propagate is False


class TestModuleImports:
    """Tests for module imports"""

    @pytest.mark.asyncio
    async def test_module_imports(self):
        """Test that console_logging_setup module can be imported"""
        from app.common.struct_logger import console_logging_setup

        assert console_logging_setup is not None
        assert hasattr(console_logging_setup, "setup_console_logging")
        assert hasattr(console_logging_setup, "_create_stdout_handler")
        assert hasattr(console_logging_setup, "_create_stderr_handler")
