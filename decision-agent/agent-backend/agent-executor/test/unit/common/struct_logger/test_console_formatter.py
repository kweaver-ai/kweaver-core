# -*- coding: utf-8 -*-
"""
Unit tests for app/common/struct_logger/console_formatter module
"""

from unittest.mock import MagicMock

import pytest


class TestParseEventContent:
    """Tests for _parse_event_content function"""

    @pytest.mark.asyncio
    async def test_parse_simple_string_event(self):
        """Test parsing a simple string event"""
        from app.common.struct_logger.console_formatter import _parse_event_content

        event, extracted = _parse_event_content("Simple message")

        assert event == "Simple message"
        assert extracted == {}

    @pytest.mark.asyncio
    async def test_parse_dict_event_with_description(self):
        """Test parsing dict event with description field"""
        from app.common.struct_logger.console_formatter import _parse_event_content

        raw_event = {"description": "Test description", "key1": "value1"}
        event, extracted = _parse_event_content(raw_event)

        assert event == "Test description"
        assert extracted == {"key1": "value1"}

    @pytest.mark.asyncio
    async def test_parse_dict_event_with_message(self):
        """Test parsing dict event with message field"""
        from app.common.struct_logger.console_formatter import _parse_event_content

        raw_event = {"message": "Test message", "key1": "value1"}
        event, extracted = _parse_event_content(raw_event)

        assert event == "Test message"
        assert extracted == {"key1": "value1"}

    @pytest.mark.asyncio
    async def test_parse_dict_event_without_special_fields(self):
        """Test parsing dict event without description/message"""
        from app.common.struct_logger.console_formatter import _parse_event_content

        raw_event = {"key1": "value1", "key2": "value2"}
        event, extracted = _parse_event_content(raw_event)

        assert event == str(raw_event)
        assert "key1" in extracted
        assert "key2" in extracted

    @pytest.mark.asyncio
    async def test_long_event_truncation(self):
        """Test that long events are truncated"""
        from app.common.struct_logger.console_formatter import _parse_event_content

        long_event = "x" * 200
        event, extracted = _parse_event_content(long_event)

        assert len(event) == 153  # 150 + "..."
        assert event.endswith("...")


class TestFormatLogHeader:
    """Tests for _format_log_header function"""

    @pytest.mark.asyncio
    async def test_format_log_header_info(self):
        """Test formatting log header for INFO level"""
        from app.common.struct_logger.console_formatter import _format_log_header

        result = _format_log_header("INFO", "Test event", "2024-01-01 12:00:00")

        assert "INFO" in result
        assert "Test event" in result
        assert "2024-01-01 12:00:00" in result
        assert "ℹ️" in result  # INFO emoji

    @pytest.mark.asyncio
    async def test_format_log_header_error(self):
        """Test formatting log header for ERROR level"""
        from app.common.struct_logger.console_formatter import _format_log_header

        result = _format_log_header("ERROR", "Error event", "2024-01-01 12:00:00")

        assert "ERROR" in result
        assert "Error event" in result
        assert "❌" in result  # ERROR emoji

    @pytest.mark.asyncio
    async def test_format_log_header_unknown_level(self):
        """Test formatting log header with unknown level"""
        from app.common.struct_logger.console_formatter import _format_log_header

        result = _format_log_header("UNKNOWN", "Test event", "2024-01-01 12:00:00")

        assert "UNKNOWN" in result
        assert "Test event" in result
        assert "📝" in result  # Default emoji


class TestFormatLocationAndType:
    """Tests for _format_location_and_type function"""

    @pytest.mark.asyncio
    async def test_format_with_caller_and_type(self):
        """Test formatting with both caller and type"""
        from app.common.struct_logger.console_formatter import _format_location_and_type

        lines = _format_location_and_type("file.py:123", "BusinessLog")

        assert len(lines) == 2
        assert any("file.py:123" in line for line in lines)
        assert any("BusinessLog" in line for line in lines)

    @pytest.mark.asyncio
    async def test_format_with_caller_only(self):
        """Test formatting with only caller"""
        from app.common.struct_logger.console_formatter import _format_location_and_type

        lines = _format_location_and_type("file.py:123", None)

        assert len(lines) == 1
        assert "file.py:123" in lines[0]

    @pytest.mark.asyncio
    async def test_format_with_type_only(self):
        """Test formatting with only type"""
        from app.common.struct_logger.console_formatter import _format_location_and_type

        lines = _format_location_and_type(None, "SystemLog")

        assert len(lines) == 1
        assert "SystemLog" in lines[0]

    @pytest.mark.asyncio
    async def test_format_with_none_values(self):
        """Test formatting with both values as None"""
        from app.common.struct_logger.console_formatter import _format_location_and_type

        lines = _format_location_and_type(None, None)

        assert len(lines) == 0


class TestFormatContextFields:
    """Tests for _format_context_fields function"""

    @pytest.mark.asyncio
    async def test_format_empty_event_dict(self):
        """Test formatting empty event dict"""
        from app.common.struct_logger.console_formatter import _format_context_fields

        lines = _format_context_fields({})

        assert len(lines) == 0

    @pytest.mark.asyncio
    async def test_format_with_simple_key_value(self):
        """Test formatting simple key-value pairs"""
        from app.common.struct_logger.console_formatter import _format_context_fields

        event_dict = {"key1": "value1", "key2": "value2"}
        lines = _format_context_fields(event_dict)

        # Should have header and at least the values
        assert len(lines) > 0
        assert any("key1" in line for line in lines)
        assert any("key2" in line for line in lines)

    @pytest.mark.asyncio
    async def test_format_skips_reserved_keys(self):
        """Test that reserved keys are skipped"""
        from app.common.struct_logger.console_formatter import _format_context_fields

        event_dict = {
            "logger": "test",
            "level": "INFO",
            "timestamp": "2024-01-01",
            "stack": "test_stack",
        }
        lines = _format_context_fields(event_dict)

        # Reserved keys should not appear in output
        assert not any("logger:" in line for line in lines)
        assert not any("level:" in line for line in lines)

    @pytest.mark.asyncio
    async def test_format_with_dict_value(self):
        """Test formatting with dict value"""
        from app.common.struct_logger.console_formatter import _format_context_fields

        event_dict = {"config": {"key": "value"}}
        lines = _format_context_fields(event_dict)

        assert len(lines) > 0

    @pytest.mark.asyncio
    async def test_format_with_list_value(self):
        """Test formatting with list value"""
        from app.common.struct_logger.console_formatter import _format_context_fields

        event_dict = {"items": ["item1", "item2"]}
        lines = _format_context_fields(event_dict)

        assert len(lines) > 0


class TestFormatConsoleLogMain:
    """Tests for format_console_log main function"""

    @pytest.mark.asyncio
    async def test_format_console_log_basic(self):
        """Test basic console log formatting"""
        from app.common.struct_logger.console_formatter import format_console_log

        logger = MagicMock()
        event_dict = {
            "timestamp": "2024-01-01 12:00:00",
            "level": "info",
            "event": "Test message",
        }

        result = format_console_log(logger, "info", event_dict)

        assert isinstance(result, str)
        assert "Test message" in result
        assert "INFO" in result

    @pytest.mark.asyncio
    async def test_format_console_log_with_context(self):
        """Test console log with context fields"""
        from app.common.struct_logger.console_formatter import format_console_log

        logger = MagicMock()
        event_dict = {
            "timestamp": "2024-01-01 12:00:00",
            "level": "info",
            "event": "Test message",
            "user_id": "123",
            "action": "login",
        }

        result = format_console_log(logger, "info", event_dict)

        assert isinstance(result, str)
        assert "Test message" in result
        assert "user_id" in result

    @pytest.mark.asyncio
    async def test_format_console_log_error_level(self):
        """Test console log with ERROR level"""
        from app.common.struct_logger.console_formatter import format_console_log

        logger = MagicMock()
        event_dict = {
            "timestamp": "2024-01-01 12:00:00",
            "level": "error",
            "event": "Error occurred",
        }

        result = format_console_log(logger, "error", event_dict)

        assert isinstance(result, str)
        assert "ERROR" in result
        assert "Error occurred" in result
        # ERROR level should use different border character
        assert "═" in result


class TestModuleImports:
    """Tests for module imports"""

    @pytest.mark.asyncio
    async def test_module_imports(self):
        """Test that console_formatter module can be imported"""
        from app.common.struct_logger import console_formatter

        assert console_formatter is not None
        assert hasattr(console_formatter, "format_console_log")
        assert hasattr(console_formatter, "_parse_event_content")


class TestParseEventContentExtended:
    """Extended tests for _parse_event_content function - covering JSON parsing"""

    @pytest.mark.asyncio
    async def test_parse_json_string_event(self):
        """Test parsing JSON string event"""
        from app.common.struct_logger.console_formatter import _parse_event_content

        json_event = '{"description": "JSON event", "extra": "data"}'
        event, extracted = _parse_event_content(json_event)

        assert "JSON event" in event
        assert "extra" in extracted

    @pytest.mark.asyncio
    async def test_parse_long_json_string_event(self):
        """Test parsing long string event that triggers JSON parsing"""
        from app.common.struct_logger.console_formatter import _parse_event_content

        long_event = "x" * 250
        event, extracted = _parse_event_content(long_event)

        # Long strings without valid JSON should be returned as-is (but truncated)
        assert event.endswith("...")

    @pytest.mark.asyncio
    async def test_parse_string_with_embedded_json(self):
        """Test parsing string with embedded JSON"""
        from app.common.struct_logger.console_formatter import _parse_event_content

        # The function looks for JSON pattern within the string using regex
        # It only parses JSON if the string starts with { or is > 200 chars
        mixed_event = '{"description": "Test message", "key": "value", "other": "data"}'
        event, extracted = _parse_event_content(mixed_event)

        assert "Test message" in event
        assert "key" in extracted
        assert "other" in extracted

    @pytest.mark.asyncio
    async def test_parse_malformed_json_string(self):
        """Test parsing malformed JSON string - should handle gracefully"""
        from app.common.struct_logger.console_formatter import _parse_event_content

        malformed_json = '{"description": "test", invalid json}'
        event, extracted = _parse_event_content(malformed_json)

        # Should fall back to using the original string
        assert event == malformed_json


class TestFormatCallStack:
    """Tests for _format_call_stack function"""

    @pytest.mark.asyncio
    async def test_format_call_stack_string_format(self):
        """Test formatting call stack in string format"""
        from app.common.struct_logger.console_formatter import _format_call_stack

        stack_str = """Traceback (most recent call last):
  File "test.py", line 10, in <module>
    raise ValueError("test")
ValueError: test"""

        lines = _format_call_stack(stack_str)

        assert len(lines) > 0
        assert any("call_stack" in line for line in lines)
        assert any("test.py" in line for line in lines)

    @pytest.mark.asyncio
    async def test_format_call_stack_list_format(self):
        """Test formatting call stack in list format"""
        from app.common.struct_logger.console_formatter import _format_call_stack

        stack_list = [
            {"file": "/path/to/file.py", "line": 42, "function": "test_func"},
            {"file": "/path/to/other.py", "line": 10, "function": "another_func"},
        ]

        lines = _format_call_stack(stack_list)

        assert len(lines) > 0
        assert any("file.py" in line for line in lines)
        assert any("other.py" in line for line in lines)
        assert any("line 42" in line for line in lines)

    @pytest.mark.asyncio
    async def test_format_call_stack_list_with_partial_data(self):
        """Test formatting call stack list with partial data"""
        from app.common.struct_logger.console_formatter import _format_call_stack

        stack_list = [
            {"file": "/path/to/file.py"},  # Missing line and function
            {"line": 10, "function": "test"},  # Missing file
        ]

        lines = _format_call_stack(stack_list)

        # Should handle partial data gracefully
        assert len(lines) > 0

    @pytest.mark.asyncio
    async def test_format_call_stack_unknown_format(self):
        """Test formatting call stack with unknown format"""
        from app.common.struct_logger.console_formatter import _format_call_stack

        lines = _format_call_stack(12345)

        assert len(lines) > 0
        assert any("12345" in line for line in lines)


class TestFormatErrorDetails:
    """Tests for _format_error_details function"""

    @pytest.mark.asyncio
    async def test_format_error_details_dict(self):
        """Test formatting error details as dict"""
        from app.common.struct_logger.console_formatter import _format_error_details

        error_details = {"error_code": "ERR001", "message": "Test error"}

        lines = _format_error_details(error_details)

        assert len(lines) > 0
        assert any("Error Details" in line for line in lines)
        assert any("error_code" in line for line in lines)

    @pytest.mark.asyncio
    async def test_format_error_details_string(self):
        """Test formatting error details as string"""
        from app.common.struct_logger.console_formatter import _format_error_details

        lines = _format_error_details("Simple error message")

        assert len(lines) > 0
        assert any("Error Details" in line for line in lines)
        assert any("Simple error message" in line for line in lines)

    @pytest.mark.asyncio
    async def test_format_error_details_none(self):
        """Test formatting None error details"""
        from app.common.struct_logger.console_formatter import _format_error_details

        lines = _format_error_details(None)

        assert len(lines) == 0

    @pytest.mark.asyncio
    async def test_format_error_details_empty(self):
        """Test formatting empty error details"""
        from app.common.struct_logger.console_formatter import _format_error_details

        lines = _format_error_details({})

        assert len(lines) == 0


class TestFormatTracebackLine:
    """Tests for _format_traceback_line function"""

    @pytest.mark.asyncio
    async def test_format_traceback_file_line(self):
        """Test formatting traceback file line"""
        from app.common.struct_logger.console_formatter import _format_traceback_line

        line = '  File "/path/to/file.py", line 42, in test_function'
        result = _format_traceback_line(line)

        assert "file.py" in result

    @pytest.mark.asyncio
    async def test_format_traceback_exception_line(self):
        """Test formatting traceback exception line"""
        from app.common.struct_logger.console_formatter import _format_traceback_line

        line = "ValueError: test error"
        result = _format_traceback_line(line)

        assert "test error" in result

    @pytest.mark.asyncio
    async def test_format_traceback_code_line(self):
        """Test formatting traceback code line"""
        from app.common.struct_logger.console_formatter import _format_traceback_line

        line = "    result = some_function()"
        result = _format_traceback_line(line)

        assert "some_function" in result


class TestFormatExceptionTraceback:
    """Tests for _format_exception_traceback function"""

    @pytest.mark.asyncio
    async def test_format_exception_traceback_with_exc_info_string(self):
        """Test formatting exception traceback with exc_info string"""
        from app.common.struct_logger.console_formatter import (
            _format_exception_traceback,
        )

        exc_info = (
            "Traceback (most recent call last):\n  File test.py\nValueError: test"
        )
        lines = _format_exception_traceback(exc_info, None, None)

        assert len(lines) > 0
        assert any("Exception Traceback" in line for line in lines)

    @pytest.mark.asyncio
    async def test_format_exception_traceback_with_exc_str(self):
        """Test formatting exception traceback with exc_str"""
        from app.common.struct_logger.console_formatter import (
            _format_exception_traceback,
        )

        exc_str = "Error: test error"
        lines = _format_exception_traceback(None, exc_str, None)

        assert len(lines) > 0
        assert any("Exception Traceback" in line for line in lines)

    @pytest.mark.asyncio
    async def test_format_exception_traceback_with_stack_str(self):
        """Test formatting exception traceback with stack_str"""
        from app.common.struct_logger.console_formatter import (
            _format_exception_traceback,
        )

        stack_str = "Stack line 1\nStack line 2"
        lines = _format_exception_traceback(None, None, stack_str)

        assert len(lines) > 0
        assert any("Exception Traceback" in line for line in lines)

    @pytest.mark.asyncio
    async def test_format_exception_traceback_with_exception_object(self):
        """Test formatting exception traceback with exception object"""
        from app.common.struct_logger.console_formatter import (
            _format_exception_traceback,
        )

        exc = ValueError("Test error")
        lines = _format_exception_traceback(exc, None, None)

        assert len(lines) > 0
        assert any("Exception Traceback" in line for line in lines)

    @pytest.mark.asyncio
    async def test_format_exception_traceback_no_info(self):
        """Test formatting exception traceback with no info"""
        from app.common.struct_logger.console_formatter import (
            _format_exception_traceback,
        )

        lines = _format_exception_traceback(None, None, None)

        assert len(lines) == 0


class TestFormatContextFieldsExtended:
    """Extended tests for _format_context_fields function"""

    @pytest.mark.asyncio
    async def test_format_with_validation_errors(self):
        """Test formatting validation_errors field"""
        from app.common.struct_logger.console_formatter import _format_context_fields

        event_dict = {
            "validation_errors": [
                {"field": "email", "message": "Invalid email"},
                {"field": "age", "message": "Must be positive"},
            ]
        }
        lines = _format_context_fields(event_dict)

        assert len(lines) > 0
        assert any("validation_errors" in line for line in lines)

    @pytest.mark.asyncio
    async def test_format_with_error_related_keys(self):
        """Test that error-related keys get error color"""
        from app.common.struct_logger.console_formatter import _format_context_fields

        event_dict = {
            "error": "Some error",
            "error_code": "E001",
            "error_details": {"info": "details"},
            "normal_field": "normal_value",
        }
        lines = _format_context_fields(event_dict)

        assert len(lines) > 0
        assert any("error" in line for line in lines)
        assert any("normal_field" in line for line in lines)

    @pytest.mark.asyncio
    async def test_format_with_unserializable_dict(self):
        """Test formatting with dict that has unserializable values"""
        from app.common.struct_logger.console_formatter import _format_context_fields

        # Create an object that can't be directly serialized
        class CustomObject:
            def __str__(self):
                return "CustomObject"

        event_dict = {"object_field": CustomObject()}
        lines = _format_context_fields(event_dict)

        # Should handle gracefully using str()
        assert len(lines) > 0

    @pytest.mark.asyncio
    async def test_format_with_multiline_dict_value(self):
        """Test formatting dict value with multiple lines"""
        from app.common.struct_logger.console_formatter import _format_context_fields

        event_dict = {
            "config": {"key1": "value1", "key2": "value2", "nested": {"a": 1}}
        }
        lines = _format_context_fields(event_dict)

        assert len(lines) > 0
        # Multi-line JSON should be properly indented
        assert any("config:" in line for line in lines)


class TestFormatConsoleLogExtended:
    """Extended tests for format_console_log main function"""

    @pytest.mark.asyncio
    async def test_format_with_call_stack_in_event_dict(self):
        """Test console log with call_stack in event dict"""
        from app.common.struct_logger.console_formatter import format_console_log

        logger = MagicMock()
        event_dict = {
            "timestamp": "2024-01-01 12:00:00",
            "level": "error",
            "event": "Error with stack",
            "stack": "Stack trace line 1\nStack trace line 2",
        }

        result = format_console_log(logger, "error", event_dict)

        assert isinstance(result, str)
        # Stack should appear in output
        assert "Exception Traceback" in result or "stack" in result.lower()

    @pytest.mark.asyncio
    async def test_format_with_exception_object(self):
        """Test console log with exception object"""
        from app.common.struct_logger.console_formatter import format_console_log

        logger = MagicMock()
        exc = ValueError("Test exception")
        event_dict = {
            "timestamp": "2024-01-01 12:00:00",
            "level": "error",
            "event": "Exception occurred",
            "exception": exc,
        }

        result = format_console_log(logger, "error", event_dict)

        assert isinstance(result, str)
        assert "Exception Traceback" in result

    @pytest.mark.asyncio
    async def test_format_with_error_details(self):
        """Test console log with error_details"""
        from app.common.struct_logger.console_formatter import format_console_log

        logger = MagicMock()
        event_dict = {
            "timestamp": "2024-01-01 12:00:00",
            "level": "error",
            "event": "Error with details",
            "error_details": {"code": "E001", "message": "Detailed error"},
        }

        result = format_console_log(logger, "error", event_dict)

        assert isinstance(result, str)
        assert "Error Details" in result
        assert "E001" in result

    @pytest.mark.asyncio
    async def test_format_with_caller_and_log_type(self):
        """Test console log with caller and log_type"""
        from app.common.struct_logger.console_formatter import format_console_log

        logger = MagicMock()
        event_dict = {
            "timestamp": "2024-01-01 12:00:00",
            "level": "info",
            "event": "Test message",
            "caller": "module.py:42",
            "log_type": "BusinessLog",
        }

        result = format_console_log(logger, "info", event_dict)

        assert isinstance(result, str)
        assert "module.py:42" in result
        assert "BusinessLog" in result
        # Separator should be present when location lines exist
        assert "┄" in result

    @pytest.mark.asyncio
    async def test_format_populates_extracted_info(self):
        """Test that extracted info from event is populated"""
        from app.common.struct_logger.console_formatter import format_console_log

        logger = MagicMock()
        event_dict = {
            "timestamp": "2024-01-01 12:00:00",
            "level": "info",
            "event": {"description": "Main message", "extra": "data"},
        }

        result = format_console_log(logger, "info", event_dict)

        assert isinstance(result, str)
        assert "Main message" in result
        assert "extra" in result

    @pytest.mark.asyncio
    async def test_format_with_all_fields(self):
        """Test console log with all possible fields"""
        from app.common.struct_logger.console_formatter import format_console_log

        logger = MagicMock()
        exc = ValueError("Test")
        event_dict = {
            "timestamp": "2024-01-01 12:00:00",
            "level": "error",
            "event": "Full test",
            "caller": "test.py:1",
            "log_type": "TestType",
            "user_id": "123",
            "action": "test_action",
            "error_details": {"info": "details"},
            "exception": exc,
            "extra_field": "extra_value",
        }

        result = format_console_log(logger, "error", event_dict)

        assert isinstance(result, str)
        assert "Full test" in result
        assert "test.py:1" in result
        assert "TestType" in result
        assert "user_id" in result
        assert "extra_field" in result

    @pytest.mark.asyncio
    async def test_format_critical_level_uses_double_border(self):
        """Test CRITICAL level uses double border character"""
        from app.common.struct_logger.console_formatter import format_console_log

        logger = MagicMock()
        event_dict = {
            "timestamp": "2024-01-01 12:00:00",
            "level": "critical",
            "event": "Critical error",
        }

        result = format_console_log(logger, "critical", event_dict)

        assert isinstance(result, str)
        # CRITICAL level should use double border
        assert "═" in result
