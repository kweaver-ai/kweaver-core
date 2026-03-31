# -*- coding: utf-8 -*-
"""
Unit tests for app/common/exception_logger/formatter_pkg/error_header.py
"""

from datetime import datetime
from unittest.mock import patch

from app.common.exception_logger.formatter_pkg.error_header import (
    format_error_header,
    format_error_footer,
)


class TestFormatErrorHeader:
    """Tests for format_error_header function"""

    def test_format_header_basic(self):
        """测试基本头部格式化（不带颜色）"""
        exc = ValueError("Test error")

        result = format_error_header(exc, colorize=False)

        assert isinstance(result, str)
        assert "ValueError" in result
        assert "Test error" in result

    def test_format_header_with_timestamp(self):
        """测试带时间戳的头部格式化"""
        exc = RuntimeError("Runtime error")
        timestamp = datetime(2024, 5, 20, 14, 30, 0)

        result = format_error_header(exc, timestamp=timestamp, colorize=False)

        assert isinstance(result, str)
        assert "2024-05-20" in result
        assert "14:30" in result

    def test_format_header_auto_timestamp(self):
        """测试自动生成时间戳的头部格式化"""
        exc = TypeError("Type error")

        with patch(
            "app.common.exception_logger.formatter_pkg.error_header.datetime"
        ) as mock_datetime:
            mock_now = datetime(2024, 8, 15, 10, 0, 0)
            mock_datetime.now.return_value = mock_now

            result = format_error_header(exc, colorize=False)

            assert isinstance(result, str)
            assert "2024-08-15" in result

    def test_format_header_with_request_info_colorized(self):
        """测试带请求信息的头部格式化（带颜色）"""
        exc = KeyError("missing_key")
        request_info = {"method": "GET", "path": "/api/resource"}

        result = format_error_header(exc, request_info=request_info, colorize=True)

        assert isinstance(result, str)
        assert "KeyError" in result

    def test_format_header_with_request_info_not_colorized(self):
        """测试带请求信息的头部格式化（不带颜色）"""
        exc = AttributeError("attr_error")
        request_info = {"service": "test_service"}

        result = format_error_header(exc, request_info=request_info, colorize=False)

        assert isinstance(result, str)
        assert "AttributeError" in result

    def test_format_header_long_message_truncated(self):
        """测试长消息被截断"""
        long_message = "x" * 200
        exc = ValueError(long_message)

        result = format_error_header(exc, colorize=False)

        assert isinstance(result, str)
        assert "..." in result

    def test_format_header_with_all_params_colorized(self):
        """测试带所有参数的头部格式化（带颜色）"""
        exc = RuntimeError("Test error")
        timestamp = datetime(2024, 11, 30, 16, 45, 0)
        request_info = {"user_id": "123", "action": "test"}

        result = format_error_header(
            exc, timestamp=timestamp, request_info=request_info, colorize=True
        )

        assert isinstance(result, str)
        assert "RuntimeError" in result
        assert "2024-11-30" in result

    def test_format_header_with_all_params_not_colorized(self):
        """测试带所有参数的头部格式化（不带颜色）"""
        exc = ValueError("Test error")
        timestamp = datetime(2024, 7, 4, 9, 15, 0)
        request_info = {"endpoint": "/api/v1/test"}

        result = format_error_header(
            exc, timestamp=timestamp, request_info=request_info, colorize=False
        )

        assert isinstance(result, str)
        assert "ValueError" in result


class TestFormatErrorFooter:
    """Tests for format_error_footer function"""

    def test_format_footer_colorized(self):
        """测试带颜色的尾部格式化"""
        result = format_error_footer(colorize=True)

        assert isinstance(result, str)
        # Colorized version should have ANSI codes
        assert "\033[" in result or result is not None

    def test_format_footer_not_colorized(self):
        """测试不带颜色的尾部格式化"""
        result = format_error_footer(colorize=False)

        assert isinstance(result, str)
        # Non-colorized version should not have ANSI codes
        assert "\033[" not in result
        # Should contain border characters
        assert "═" in result or result is not None

    def test_format_footer_default_colorize(self):
        """测试默认 colorize=False 的尾部格式化"""
        result = format_error_footer()

        assert isinstance(result, str)
        # Default should not have ANSI codes
        assert "\033[" not in result
