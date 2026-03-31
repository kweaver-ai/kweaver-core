# -*- coding: utf-8 -*-
"""
Unit tests for app/common/struct_logger/logger module
"""

from unittest.mock import MagicMock, patch

import pytest


class TestStructLoggerSingleton:
    """Tests for StructLogger singleton pattern"""

    @pytest.mark.asyncio
    async def test_singleton_pattern(self):
        """Test that StructLogger is a singleton"""
        # Reset the singleton
        from app.common.struct_logger import logger

        logger.StructLogger._instance = None
        logger.StructLogger._initialized = False

        from app.common.struct_logger.logger import StructLogger

        instance1 = StructLogger()
        instance2 = StructLogger()

        assert instance1 is instance2

    @pytest.mark.asyncio
    async def test_singleton_persists(self):
        """Test that singleton persists but don't reload"""
        # Reset the singleton
        from app.common.struct_logger import logger

        logger.StructLogger._instance = None
        logger.StructLogger._initialized = False

        from app.common.struct_logger.logger import StructLogger

        instance1 = StructLogger()
        # Don't reload - just check the same instance
        from app.common.struct_logger.logger import StructLogger as StructLogger2

        instance2 = StructLogger2()

        assert instance1 is instance2

    @pytest.mark.asyncio
    async def test_init_only_once(self):
        """Test that __init__ only runs once"""
        # Reset the singleton
        from app.common.struct_logger import logger

        logger.StructLogger._instance = None
        logger.StructLogger._initialized = False

        from app.common.struct_logger.logger import StructLogger

        with patch.object(StructLogger, "_setup_logging") as mock_setup:
            instance1 = StructLogger()
            instance2 = StructLogger()

            # _setup_logging should only be called once
            mock_setup.assert_called_once()


class TestStructLoggerLoggingMethods:
    """Tests for StructLogger logging methods"""

    @pytest.mark.asyncio
    async def test_debug_method(self):
        """Test debug logging method"""
        # Reset the singleton
        from app.common.struct_logger import logger

        logger.StructLogger._instance = None
        logger.StructLogger._initialized = False

        from app.common.struct_logger.logger import StructLogger

        with patch("app.common.struct_logger.logger.setup_file_logging") as mock_file:
            with patch(
                "app.common.struct_logger.logger.setup_console_logging"
            ) as mock_console:
                mock_file_logger = MagicMock()
                mock_console_logger = MagicMock()
                mock_file.return_value = mock_file_logger
                mock_console.return_value = mock_console_logger

                struct_logger = StructLogger()
                struct_logger.debug("Test debug", key="value")

                mock_file_logger.debug.assert_called_once_with(
                    "Test debug", key="value"
                )
                mock_console_logger.debug.assert_called_once_with(
                    "Test debug", key="value"
                )

    @pytest.mark.asyncio
    async def test_info_method(self):
        """Test info logging method"""
        # Reset the singleton
        from app.common.struct_logger import logger

        logger.StructLogger._instance = None
        logger.StructLogger._initialized = False

        from app.common.struct_logger.logger import StructLogger

        with patch("app.common.struct_logger.logger.setup_file_logging") as mock_file:
            with patch(
                "app.common.struct_logger.logger.setup_console_logging"
            ) as mock_console:
                mock_file_logger = MagicMock()
                mock_console_logger = MagicMock()
                mock_file.return_value = mock_file_logger
                mock_console.return_value = mock_console_logger

                struct_logger = StructLogger()
                struct_logger.info("Test info", user_id="123")

                mock_file_logger.info.assert_called_once_with(
                    "Test info", user_id="123"
                )
                mock_console_logger.info.assert_called_once_with(
                    "Test info", user_id="123"
                )

    @pytest.mark.asyncio
    async def test_warning_method(self):
        """Test warning logging method"""
        # Reset the singleton
        from app.common.struct_logger import logger

        logger.StructLogger._instance = None
        logger.StructLogger._initialized = False

        from app.common.struct_logger.logger import StructLogger

        with patch("app.common.struct_logger.logger.setup_file_logging") as mock_file:
            with patch(
                "app.common.struct_logger.logger.setup_console_logging"
            ) as mock_console:
                mock_file_logger = MagicMock()
                mock_console_logger = MagicMock()
                mock_file.return_value = mock_file_logger
                mock_console.return_value = mock_console_logger

                struct_logger = StructLogger()
                struct_logger.warning("Test warning")

                mock_file_logger.warning.assert_called_once_with("Test warning")
                mock_console_logger.warning.assert_called_once_with("Test warning")

    @pytest.mark.asyncio
    async def test_warn_alias(self):
        """Test warn method is alias for warning"""
        # Reset the singleton
        from app.common.struct_logger import logger

        logger.StructLogger._instance = None
        logger.StructLogger._initialized = False

        from app.common.struct_logger.logger import StructLogger

        with patch("app.common.struct_logger.logger.setup_file_logging") as mock_file:
            with patch(
                "app.common.struct_logger.logger.setup_console_logging"
            ) as mock_console:
                mock_file_logger = MagicMock()
                mock_console_logger = MagicMock()
                mock_file.return_value = mock_file_logger
                mock_console.return_value = mock_console_logger

                struct_logger = StructLogger()
                struct_logger.warn("Test warn")

                mock_file_logger.warning.assert_called_once_with("Test warn")
                mock_console_logger.warning.assert_called_once_with("Test warn")

    @pytest.mark.asyncio
    async def test_error_method(self):
        """Test error logging method"""
        # Reset the singleton
        from app.common.struct_logger import logger

        logger.StructLogger._instance = None
        logger.StructLogger._initialized = False

        from app.common.struct_logger.logger import StructLogger

        with patch("app.common.struct_logger.logger.setup_file_logging") as mock_file:
            with patch(
                "app.common.struct_logger.logger.setup_console_logging"
            ) as mock_console:
                mock_file_logger = MagicMock()
                mock_console_logger = MagicMock()
                mock_file.return_value = mock_file_logger
                mock_console.return_value = mock_console_logger

                struct_logger = StructLogger()
                struct_logger.error("Test error", code=500)

                mock_file_logger.error.assert_called_once_with("Test error", code=500)
                mock_console_logger.error.assert_called_once_with(
                    "Test error", code=500
                )

    @pytest.mark.asyncio
    async def test_fatal_method(self):
        """Test fatal logging method"""
        # Reset the singleton
        from app.common.struct_logger import logger

        logger.StructLogger._instance = None
        logger.StructLogger._initialized = False

        from app.common.struct_logger.logger import StructLogger

        with patch("app.common.struct_logger.logger.setup_file_logging") as mock_file:
            with patch(
                "app.common.struct_logger.logger.setup_console_logging"
            ) as mock_console:
                mock_file_logger = MagicMock()
                mock_console_logger = MagicMock()
                mock_file.return_value = mock_file_logger
                mock_console.return_value = mock_console_logger

                struct_logger = StructLogger()
                struct_logger.fatal("Test fatal")

                # fatal uses critical method
                mock_file_logger.critical.assert_called_once_with("Test fatal")
                mock_console_logger.critical.assert_called_once_with("Test fatal")


class TestStructLoggerBind:
    """Tests for StructLogger bind method"""

    @pytest.mark.asyncio
    async def test_bind_method(self):
        """Test bind method returns file_logger with context"""
        # Reset the singleton
        from app.common.struct_logger import logger

        logger.StructLogger._instance = None
        logger.StructLogger._initialized = False

        from app.common.struct_logger.logger import StructLogger

        with patch("app.common.struct_logger.logger.setup_file_logging") as mock_file:
            with patch(
                "app.common.struct_logger.logger.setup_console_logging"
            ) as mock_console:
                mock_file_logger = MagicMock()
                mock_bound_logger = MagicMock()
                mock_file_logger.bind.return_value = mock_bound_logger
                mock_console_logger = MagicMock()
                mock_file.return_value = mock_file_logger
                mock_console.return_value = mock_console_logger

                struct_logger = StructLogger()
                result = struct_logger.bind(request_id="123", user_id="456")

                # Should return file_logger.bind result
                mock_file_logger.bind.assert_called_once_with(
                    request_id="123", user_id="456"
                )
                assert result is mock_bound_logger


class TestStructLoggerProperties:
    """Tests for StructLogger properties"""

    @pytest.mark.asyncio
    async def test_file_logger_property(self):
        """Test file_logger property"""
        # Reset the singleton
        from app.common.struct_logger import logger

        logger.StructLogger._instance = None
        logger.StructLogger._initialized = False

        from app.common.struct_logger.logger import StructLogger

        with patch("app.common.struct_logger.logger.setup_file_logging") as mock_file:
            with patch(
                "app.common.struct_logger.logger.setup_console_logging"
            ) as mock_console:
                mock_file_logger = MagicMock()
                mock_console_logger = MagicMock()
                mock_file.return_value = mock_file_logger
                mock_console.return_value = mock_console_logger

                struct_logger = StructLogger()
                result = struct_logger.file_logger

                assert result is mock_file_logger

    @pytest.mark.asyncio
    async def test_console_logger_property(self):
        """Test console_logger property"""
        # Reset the singleton
        from app.common.struct_logger import logger

        logger.StructLogger._instance = None
        logger.StructLogger._initialized = False

        from app.common.struct_logger.logger import StructLogger

        with patch("app.common.struct_logger.logger.setup_file_logging") as mock_file:
            with patch(
                "app.common.struct_logger.logger.setup_console_logging"
            ) as mock_console:
                mock_file_logger = MagicMock()
                mock_console_logger = MagicMock()
                mock_file.return_value = mock_file_logger
                mock_console.return_value = mock_console_logger

                struct_logger = StructLogger()
                result = struct_logger.console_logger

                assert result is mock_console_logger


class TestModuleImports:
    """Tests for module imports"""

    @pytest.mark.asyncio
    async def test_module_imports(self):
        """Test that logger module can be imported"""
        from app.common.struct_logger import logger

        assert logger is not None
        assert hasattr(logger, "StructLogger")
