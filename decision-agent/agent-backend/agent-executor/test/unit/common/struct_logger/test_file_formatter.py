# -*- coding: utf-8 -*-
"""
Unit tests for app/common/struct_logger/file_formatter module
"""

from unittest.mock import MagicMock
import json

import pytest


class TestFormatFileLog:
    """Tests for format_file_log function"""

    @pytest.mark.asyncio
    async def test_format_file_log_basic(self):
        """Test basic file log formatting"""
        from app.common.struct_logger.file_formatter import format_file_log

        logger = MagicMock()
        event_dict = {
            "timestamp": "2024-01-01 12:00:00",
            "level": "info",
            "event": "Test message",
        }

        result = format_file_log(logger, "info", event_dict)

        assert isinstance(result, str)
        assert "2024-01-01 12:00:00" in result
        assert "INFO" in result
        assert "{" in result  # JSON content

    @pytest.mark.asyncio
    async def test_format_file_log_with_context(self):
        """Test file log with context fields"""
        from app.common.struct_logger.file_formatter import format_file_log

        logger = MagicMock()
        event_dict = {
            "timestamp": "2024-01-01 12:00:00",
            "level": "info",
            "event": "Test message",
            "user_id": "123",
            "action": "login",
        }

        result = format_file_log(logger, "info", event_dict)

        assert isinstance(result, str)
        assert "user_id" in result
        assert "123" in result
        assert "action" in result
        assert "login" in result

    @pytest.mark.asyncio
    async def test_format_file_log_sorts_keys(self):
        """Test that file log JSON has sorted keys"""
        from app.common.struct_logger.file_formatter import format_file_log

        logger = MagicMock()
        event_dict = {
            "timestamp": "2024-01-01 12:00:00",
            "level": "info",
            "z_key": "last",
            "a_key": "first",
            "m_key": "middle",
        }

        result = format_file_log(logger, "info", event_dict)

        # Extract JSON part (after the last " - ")
        json_part = result.split(" - ", 2)[-1]
        parsed = json.loads(json_part)

        keys = list(parsed.keys())
        # Keys should be sorted
        assert keys == sorted(keys)

    @pytest.mark.asyncio
    async def test_format_file_log_error_level(self):
        """Test file log with ERROR level"""
        from app.common.struct_logger.file_formatter import format_file_log

        logger = MagicMock()
        event_dict = {
            "timestamp": "2024-01-01 12:00:00",
            "level": "error",
            "event": "Error occurred",
        }

        result = format_file_log(logger, "error", event_dict)

        assert isinstance(result, str)
        assert "ERROR" in result
        assert "Error occurred" in result

    @pytest.mark.asyncio
    async def test_format_file_log_with_nested_dict(self):
        """Test file log with nested dictionary"""
        from app.common.struct_logger.file_formatter import format_file_log

        logger = MagicMock()
        event_dict = {
            "timestamp": "2024-01-01 12:00:00",
            "level": "info",
            "event": "Test",
            "metadata": {"key1": "value1", "key2": "value2"},
        }

        result = format_file_log(logger, "info", event_dict)

        assert isinstance(result, str)
        assert "metadata" in result
        assert "key1" in result

    @pytest.mark.asyncio
    async def test_format_file_log_with_list(self):
        """Test file log with list value"""
        from app.common.struct_logger.file_formatter import format_file_log

        logger = MagicMock()
        event_dict = {
            "timestamp": "2024-01-01 12:00:00",
            "level": "info",
            "event": "Test",
            "items": ["item1", "item2", "item3"],
        }

        result = format_file_log(logger, "info", event_dict)

        assert isinstance(result, str)
        assert "items" in result
        assert "item1" in result

    @pytest.mark.asyncio
    async def test_format_file_log_unserializable_object(self):
        """Test file log with unserializable object (uses safe serialization)"""
        from app.common.struct_logger.file_formatter import format_file_log

        class CustomObject:
            def __str__(self):
                return "CustomObject()"

        logger = MagicMock()
        event_dict = {
            "timestamp": "2024-01-01 12:00:00",
            "level": "info",
            "event": "Test",
            "custom": CustomObject(),
        }

        result = format_file_log(logger, "info", event_dict)

        assert isinstance(result, str)
        # Should successfully serialize using safe_json_serialize
        assert "custom" in result

    @pytest.mark.asyncio
    async def test_format_file_log_format_structure(self):
        """Test file log format structure"""
        from app.common.struct_logger.file_formatter import format_file_log

        logger = MagicMock()
        event_dict = {
            "timestamp": "2024-01-01 12:00:00",
            "level": "info",
            "event": "Test message",
        }

        result = format_file_log(logger, "info", event_dict)

        # Format should be: "timestamp - level - {json}"
        parts = result.split(" - ")
        assert len(parts) == 3
        assert parts[0] == "2024-01-01 12:00:00"
        assert parts[1] == "INFO"
        # Third part should be valid JSON
        json.loads(parts[2])

    @pytest.mark.asyncio
    async def test_format_file_log_empty_event_dict(self):
        """Test file log with minimal event dict"""
        from app.common.struct_logger.file_formatter import format_file_log

        logger = MagicMock()
        event_dict = {"timestamp": "2024-01-01 12:00:00", "level": "info"}

        result = format_file_log(logger, "info", event_dict)

        assert isinstance(result, str)
        assert "2024-01-01 12:00:00" in result
        assert "INFO" in result

    @pytest.mark.asyncio
    async def test_format_file_log_unicode(self):
        """Test file log with unicode content"""
        from app.common.struct_logger.file_formatter import format_file_log

        logger = MagicMock()
        event_dict = {
            "timestamp": "2024-01-01 12:00:00",
            "level": "info",
            "event": "测试消息",
            "user": "用户123",
        }

        result = format_file_log(logger, "info", event_dict)

        assert isinstance(result, str)
        assert "测试消息" in result
        assert "用户123" in result


class TestModuleImports:
    """Tests for module imports"""

    @pytest.mark.asyncio
    async def test_module_imports(self):
        """Test that file_formatter module can be imported"""
        from app.common.struct_logger import file_formatter

        assert file_formatter is not None
        assert hasattr(file_formatter, "format_file_log")
