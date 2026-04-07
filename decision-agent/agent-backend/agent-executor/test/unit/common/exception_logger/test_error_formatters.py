# -*- coding: utf-8 -*-
"""
Unit tests for app/common/exception_logger/formatter_pkg module
"""

from datetime import datetime

from app.common.exception_logger.formatter_pkg.error_formatters import (
    format_error_console,
    format_error_file_simple,
    format_error_file_detailed,
    format_multiple_errors_separator,
)


class TestFormatErrorConsole:
    """Tests for format_error_console function"""

    def test_format_error_console_basic(self):
        """测试基本控制台错误格式化"""
        exc = ValueError("Test error message")

        result = format_error_console(exc)

        assert isinstance(result, str)
        assert "ValueError" in result
        assert "Test error message" in result

    def test_format_error_console_with_timestamp(self):
        """测试带时间戳的控制台错误格式化"""
        exc = RuntimeError("Test error")
        timestamp = datetime(2024, 1, 1, 12, 0, 0)

        result = format_error_console(exc, timestamp=timestamp)

        assert isinstance(result, str)
        assert "2024-01-01" in result

    def test_format_error_console_with_request_info(self):
        """测试带请求信息的控制台错误格式化"""
        exc = TypeError("Type error")
        request_info = {"method": "POST", "path": "/api/test", "status": 500}

        result = format_error_console(exc, request_info=request_info)

        assert isinstance(result, str)
        # Request info should be included
        assert "POST" in result or "/api/test" in result

    def test_format_error_console_long_message_truncated(self):
        """测试长消息被截断"""
        long_message = "x" * 200
        exc = ValueError(long_message)

        result = format_error_console(exc)

        assert isinstance(result, str)
        # Message should be truncated
        assert "..." in result

    def test_format_error_console_with_all_params(self):
        """测试带所有参数的控制台错误格式化"""
        exc = KeyError("missing_key")
        timestamp = datetime(2024, 6, 15, 14, 30, 0)
        request_info = {"user_id": "12345", "action": "test_action"}

        result = format_error_console(
            exc, timestamp=timestamp, request_info=request_info
        )

        assert isinstance(result, str)
        assert "KeyError" in result
        assert "2024-06-15" in result


class TestFormatErrorFileSimple:
    """Tests for format_error_file_simple function"""

    def test_format_error_file_simple_basic(self):
        """测试基本文件错误格式化（简单）"""
        exc = ValueError("File error")

        result = format_error_file_simple(exc)

        assert isinstance(result, str)
        assert "ValueError" in result
        assert "File error" in result
        # Should not have color codes
        assert "\033[" not in result  # No ANSI color codes

    def test_format_error_file_simple_with_timestamp(self):
        """测试带时间戳的简单文件错误格式化"""
        exc = RuntimeError("Runtime error")
        timestamp = datetime(2024, 3, 15, 10, 30, 0)

        result = format_error_file_simple(exc, timestamp=timestamp)

        assert isinstance(result, str)
        assert "2024-03-15" in result

    def test_format_error_file_simple_with_request_info(self):
        """测试带请求信息的简单文件错误格式化"""
        exc = TypeError("Type error")
        request_info = {"endpoint": "/api/v1/resource"}

        result = format_error_file_simple(exc, request_info=request_info)

        assert isinstance(result, str)
        # Should include traceback section
        assert "Traceback:" in result


class TestFormatErrorFileDetailed:
    """Tests for format_error_file_detailed function"""

    def test_format_error_file_detailed_basic(self):
        """测试基本文件错误格式化（详细）"""
        exc = ValueError("Detailed error")

        result = format_error_file_detailed(exc)

        assert isinstance(result, str)
        assert "ValueError" in result
        assert "Detailed error" in result
        # Should include full traceback
        assert "Full Traceback:" in result

    def test_format_error_file_detailed_with_timestamp(self):
        """测试带时间戳的详细文件错误格式化"""
        exc = RuntimeError("Runtime error")
        timestamp = datetime(2024, 12, 25, 16, 45, 0)

        result = format_error_file_detailed(exc, timestamp=timestamp)

        assert isinstance(result, str)
        assert "2024-12-25" in result

    def test_format_error_file_detailed_with_request_info(self):
        """测试带请求信息的详细文件错误格式化"""
        exc = AttributeError("Attribute not found")
        request_info = {"service": "auth", "operation": "login"}

        result = format_error_file_detailed(exc, request_info=request_info)

        assert isinstance(result, str)
        assert "AttributeError" in result


class TestFormatMultipleErrorsSeparator:
    """Tests for format_multiple_errors_separator function"""

    def test_format_separator_first_error(self):
        """测试第一个错误的分隔符"""
        result = format_multiple_errors_separator(1, 3, colorize=False)

        assert "Exception 1/3" in result

    def test_format_separator_middle_error(self):
        """测试中间错误的分隔符"""
        result = format_multiple_errors_separator(2, 5, colorize=False)

        assert "Exception 2/5" in result

    def test_format_separator_last_error(self):
        """测试最后一个错误的分隔符"""
        result = format_multiple_errors_separator(5, 5, colorize=False)

        assert "Exception 5/5" in result

    def test_format_separator_with_colorize(self):
        """测试带颜色的分隔符"""
        result = format_multiple_errors_separator(1, 2, colorize=True)

        assert isinstance(result, str)
        assert "Exception 1/2" in result
        # Colorized version should have ANSI codes
        assert "\033[" in result or "Exception" in result

    def test_format_separator_without_colorize(self):
        """测试不带颜色的分隔符"""
        result = format_multiple_errors_separator(1, 1, colorize=False)

        assert isinstance(result, str)
        assert "Exception 1/1" in result
        # Non-colorized version should not have ANSI codes
        assert "\033[" not in result

    def test_format_separator_large_index(self):
        """测试大索引的分隔符"""
        result = format_multiple_errors_separator(99, 100, colorize=False)

        assert "Exception 99/100" in result
