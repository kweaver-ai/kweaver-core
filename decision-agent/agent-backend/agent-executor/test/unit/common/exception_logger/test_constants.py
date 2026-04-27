# -*- coding: utf-8 -*-
"""
Unit tests for app/common/exception_logger/constants.py module
"""

import pytest


class TestExceptionLoggerConstants:
    """Tests for exception_logger constants"""

    @pytest.mark.asyncio
    async def test_project_root_constant(self):
        """Test PROJECT_ROOT constant is defined correctly"""
        # Import after potential path setup
        from app.common.exception_logger.constants import PROJECT_ROOT

        assert PROJECT_ROOT is not None
        assert isinstance(PROJECT_ROOT, str)
        # Should point to project root (parent of app directory)
        assert "agent-executor" in PROJECT_ROOT or "agent-backend" in PROJECT_ROOT

    @pytest.mark.asyncio
    async def test_log_dir_constant(self):
        """Test EXCEPTION_LOG_DIR constant"""
        from app.common.exception_logger.constants import EXCEPTION_LOG_DIR

        assert EXCEPTION_LOG_DIR == "log/exceptions"

    @pytest.mark.asyncio
    async def test_log_file_constants(self):
        """Test log file name constants"""
        from app.common.exception_logger.constants import (
            EXCEPTION_LOG_SIMPLE,
            EXCEPTION_LOG_DETAILED,
        )

        assert EXCEPTION_LOG_SIMPLE == "exception_simple.log"
        assert EXCEPTION_LOG_DETAILED == "exception_detailed.log"

    @pytest.mark.asyncio
    async def test_ansi_reset_codes(self):
        """Test ANSI reset codes"""
        from app.common.exception_logger.constants import RESET, BOLD, DIM, UNDERLINE

        assert RESET == "\033[0m"
        assert BOLD == "\033[1m"
        assert DIM == "\033[2m"
        assert UNDERLINE == "\033[4m"

    @pytest.mark.asyncio
    async def test_colors_constant(self):
        """Test COLORS dictionary"""
        from app.common.exception_logger.constants import COLORS

        assert isinstance(COLORS, dict)
        assert "timestamp" in COLORS
        assert "error" in COLORS
        assert "critical" in COLORS
        assert "warning" in COLORS
        assert "caller" in COLORS
        assert "key" in COLORS
        assert "value" in COLORS
        assert "error_value" in COLORS
        assert "border" in COLORS
        assert "exception_type" in COLORS
        assert "exception_msg" in COLORS
        assert "traceback" in COLORS
        assert "project_code" in COLORS
        assert "separator" in COLORS

    @pytest.mark.asyncio
    async def test_level_emoji_constant(self):
        """Test LEVEL_EMOJI dictionary"""
        from app.common.exception_logger.constants import LEVEL_EMOJI

        assert isinstance(LEVEL_EMOJI, dict)
        assert "ERROR" in LEVEL_EMOJI
        assert "CRITICAL" in LEVEL_EMOJI
        assert LEVEL_EMOJI["ERROR"] == "❌"
        assert LEVEL_EMOJI["CRITICAL"] == "🔥"

    @pytest.mark.asyncio
    async def test_border_constants(self):
        """Test border character constants"""
        from app.common.exception_logger.constants import (
            BORDER_DOUBLE,
            BORDER_SINGLE,
            BORDER_DOT,
        )

        assert BORDER_DOUBLE == "═"
        assert BORDER_SINGLE == "─"
        assert BORDER_DOT == "┄"

    @pytest.mark.asyncio
    async def test_border_width_constant(self):
        """Test BORDER_WIDTH constant"""
        from app.common.exception_logger.constants import BORDER_WIDTH

        assert BORDER_WIDTH == 100
        assert isinstance(BORDER_WIDTH, int)

    @pytest.mark.asyncio
    async def test_colors_are_ansi_codes(self):
        """Test that all color values are ANSI escape codes"""
        from app.common.exception_logger.constants import COLORS

        for color_name, color_value in COLORS.items():
            assert color_value.startswith("\033[")
            assert color_value.endswith("m")
