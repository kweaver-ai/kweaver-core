# -*- coding: utf-8 -*-
"""
Unit tests for app/common/struct_logger/constants module
"""

import pytest


class TestConstants:
    """Tests for constant definitions"""

    @pytest.mark.asyncio
    async def test_log_dir_constant(self):
        """Test LOG_DIR constant"""
        from app.common.struct_logger.constants import LOG_DIR

        assert LOG_DIR == "log"
        assert isinstance(LOG_DIR, str)

    @pytest.mark.asyncio
    async def test_system_log_constant(self):
        """Test SYSTEM_LOG constant"""
        from app.common.struct_logger.constants import SYSTEM_LOG

        assert SYSTEM_LOG == "SystemLog"
        assert isinstance(SYSTEM_LOG, str)

    @pytest.mark.asyncio
    async def test_business_log_constant(self):
        """Test BUSINESS_LOG constant"""
        from app.common.struct_logger.constants import BUSINESS_LOG

        assert BUSINESS_LOG == "BusinessLog"
        assert isinstance(BUSINESS_LOG, str)

    @pytest.mark.asyncio
    async def test_ansi_reset_code(self):
        """Test ANSI reset code"""
        from app.common.struct_logger.constants import RESET

        assert RESET == "\033[0m"

    @pytest.mark.asyncio
    async def test_ansi_bold_code(self):
        """Test ANSI bold code"""
        from app.common.struct_logger.constants import BOLD

        assert BOLD == "\033[1m"

    @pytest.mark.asyncio
    async def test_ansi_dim_code(self):
        """Test ANSI dim code"""
        from app.common.struct_logger.constants import DIM

        assert DIM == "\033[2m"

    @pytest.mark.asyncio
    async def test_colors_dictionary(self):
        """Test COLORS dictionary"""
        from app.common.struct_logger.constants import COLORS

        assert isinstance(COLORS, dict)
        assert "timestamp" in COLORS
        assert "debug" in COLORS
        assert "info" in COLORS
        assert "warning" in COLORS
        assert "error" in COLORS
        assert "critical" in COLORS
        assert "caller" in COLORS
        assert "key" in COLORS
        assert "value" in COLORS
        assert "error_value" in COLORS
        assert "border" in COLORS
        assert "exception_type" in COLORS
        assert "exception_msg" in COLORS
        assert "traceback" in COLORS

        # Verify they are ANSI codes (start with \033)
        for color_name, color_code in COLORS.items():
            assert color_code.startswith("\033")

    @pytest.mark.asyncio
    async def test_level_emoji_dictionary(self):
        """Test LEVEL_EMOJI dictionary"""
        from app.common.struct_logger.constants import LEVEL_EMOJI

        assert isinstance(LEVEL_EMOJI, dict)
        assert "DEBUG" in LEVEL_EMOJI
        assert "INFO" in LEVEL_EMOJI
        assert "WARNING" in LEVEL_EMOJI
        assert "ERROR" in LEVEL_EMOJI
        assert "CRITICAL" in LEVEL_EMOJI

        # Verify emoji values
        assert LEVEL_EMOJI["DEBUG"] == "🔍"
        assert LEVEL_EMOJI["INFO"] == "ℹ️"
        assert LEVEL_EMOJI["WARNING"] == "⚠️"
        assert LEVEL_EMOJI["ERROR"] == "❌"
        assert LEVEL_EMOJI["CRITICAL"] == "🔥"


class TestModuleImports:
    """Tests for module imports"""

    @pytest.mark.asyncio
    async def test_module_imports(self):
        """Test that constants module can be imported"""
        from app.common.struct_logger import constants

        assert constants is not None
        assert hasattr(constants, "LOG_DIR")
        assert hasattr(constants, "SYSTEM_LOG")
        assert hasattr(constants, "BUSINESS_LOG")
        assert hasattr(constants, "COLORS")
        assert hasattr(constants, "LEVEL_EMOJI")
        assert hasattr(constants, "RESET")
        assert hasattr(constants, "BOLD")
        assert hasattr(constants, "DIM")
