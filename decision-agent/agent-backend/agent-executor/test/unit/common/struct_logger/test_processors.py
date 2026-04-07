# -*- coding: utf-8 -*-
"""
Unit tests for app/common/struct_logger/processors module
"""

from unittest.mock import MagicMock

import pytest


class TestAddCallerInfo:
    """Tests for add_caller_info processor"""

    @pytest.mark.asyncio
    async def test_add_caller_info_basic(self):
        """Test that add_caller_info adds caller information"""
        from app.common.struct_logger.processors import add_caller_info

        logger = MagicMock()
        method_name = "info"
        event_dict = {}

        result = add_caller_info(logger, method_name, event_dict)

        assert "caller" in result
        assert isinstance(result["caller"], str)
        assert ":" in result["caller"]  # Should contain filename:lineno format

    @pytest.mark.asyncio
    async def test_add_caller_info_preserves_existing_fields(self):
        """Test that add_caller_info preserves existing event fields"""
        from app.common.struct_logger.processors import add_caller_info

        logger = MagicMock()
        method_name = "error"
        event_dict = {"user_id": "123", "action": "login"}

        result = add_caller_info(logger, method_name, event_dict)

        assert result["user_id"] == "123"
        assert result["action"] == "login"
        assert "caller" in result

    @pytest.mark.asyncio
    async def test_add_caller_info_overwrites_existing(self):
        """Test that add_caller_info overwrites existing caller field"""
        from app.common.struct_logger.processors import add_caller_info

        logger = MagicMock()
        method_name = "debug"
        event_dict = {"caller": "old_value"}

        result = add_caller_info(logger, method_name, event_dict)

        assert result["caller"] != "old_value"
        assert ":" in result["caller"]

    @pytest.mark.asyncio
    async def test_add_caller_info_returns_event_dict(self):
        """Test that add_caller_info returns the modified event_dict"""
        from app.common.struct_logger.processors import add_caller_info

        logger = MagicMock()
        method_name = "info"
        event_dict = {"key": "value"}

        result = add_caller_info(logger, method_name, event_dict)

        # Should return the same dict object (modified in place)
        assert result is event_dict

    @pytest.mark.asyncio
    async def test_add_caller_info_different_log_levels(self):
        """Test add_caller_info with different log levels"""
        from app.common.struct_logger.processors import add_caller_info

        logger = MagicMock()

        for level in ["debug", "info", "warning", "error", "critical"]:
            event_dict = {}
            result = add_caller_info(logger, level, event_dict)

            assert "caller" in result
            assert isinstance(result["caller"], str)


class TestModuleImports:
    """Tests for module imports"""

    @pytest.mark.asyncio
    async def test_module_imports(self):
        """Test that processors module can be imported"""
        from app.common.struct_logger import processors

        assert processors is not None
        assert hasattr(processors, "add_caller_info")
