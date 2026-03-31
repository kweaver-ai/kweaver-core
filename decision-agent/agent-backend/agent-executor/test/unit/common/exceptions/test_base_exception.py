# -*- coding: utf-8 -*-
"""
Unit tests for app/common/exceptions/base_exception.py
"""

from unittest.mock import patch


class TestBaseExceptionInit:
    """测试 BaseException 初始化"""

    def test_init_with_error_only(self):
        """测试仅用 error 初始化"""
        from app.common.exceptions.base_exception import BaseException
        from app.common.errors.api_error_class import APIError

        with patch("app.common.config.Config.is_debug_mode", return_value=False):
            error = APIError(
                error_code="TEST_ERROR",
                description="Test description",
                solution="Test solution",
            )

        exc = BaseException(error=error)

        assert exc.error == error
        assert exc.error_details == ""

    def test_init_with_error_and_details(self):
        """测试用 error 和 details 初始化"""
        from app.common.exceptions.base_exception import BaseException
        from app.common.errors.api_error_class import APIError

        with patch("app.common.config.Config.is_debug_mode", return_value=False):
            error = APIError(
                error_code="TEST_ERROR",
                description="Test description",
                solution="Test solution",
            )

        exc = BaseException(error=error, error_details="Additional details")

        assert exc.error == error
        assert exc.error_details == "Additional details"

    def test_str_representation(self):
        """测试字符串表示"""
        from app.common.exceptions.base_exception import BaseException
        from app.common.errors.api_error_class import APIError

        with patch("app.common.config.Config.is_debug_mode", return_value=False):
            error = APIError(
                error_code="TEST_ERROR",
                description="Test description",
                solution="Test solution",
            )

        exc = BaseException(error=error, error_details="Test details")

        str_repr = str(exc)
        assert "TEST_ERROR" in str_repr

    def test_repr_representation(self):
        """测试 repr 表示"""
        from app.common.exceptions.base_exception import BaseException
        from app.common.errors.api_error_class import APIError

        with patch("app.common.config.Config.is_debug_mode", return_value=False):
            error = APIError(
                error_code="TEST_ERROR",
                description="Test description",
                solution="Test solution",
            )

        exc = BaseException(error=error)

        repr_str = repr(exc)
        assert "TEST_ERROR" in repr_str

    def test_format_http_error(self):
        """测试 FormatHttpError 方法"""
        from app.common.exceptions.base_exception import BaseException
        from app.common.errors.api_error_class import APIError

        with patch("app.common.config.Config.is_debug_mode", return_value=False):
            error = APIError(
                error_code="TEST_ERROR",
                description="Test description",
                solution="Test solution",
            )

        exc = BaseException(error=error, error_details="Additional details")

        result = exc.FormatHttpError()

        assert result["error_code"] == "TEST_ERROR"
        # description 优先级: self.description > error.description > error_details
        assert result["description"] == "Test description"
        assert result["solution"] == "Test solution"

    def test_format_log_error(self):
        """测试 FormatLogError 方法"""
        from app.common.exceptions.base_exception import BaseException
        from app.common.errors.api_error_class import APIError

        with patch("app.common.config.Config.is_debug_mode", return_value=False):
            error = APIError(
                error_code="TEST_ERROR",
                description="Test description",
                solution="Test solution",
            )

        exc = BaseException(error=error, error_details="Additional details")

        result = exc.FormatLogError()

        assert result["error_code"] == "TEST_ERROR"
        # description 优先级: self.description > error.description > error_details
        assert result["description"] == "Test description"
        assert "trace" not in result

    def test_init_with_all_params(self):
        """测试使用所有参数初始化"""
        from app.common.exceptions.base_exception import BaseException
        from app.common.errors.api_error_class import APIError

        with patch("app.common.config.Config.is_debug_mode", return_value=False):
            error = APIError(
                error_code="TEST_ERROR",
                description="Test description",
                solution="Test solution",
            )

        exc = BaseException(
            error=error,
            error_details="Additional details",
            error_link="https://example.com/error",
            description="Custom description",
        )

        assert exc.error == error
        assert exc.error_details == "Additional details"
        assert exc.error_link == "https://example.com/error"
        assert exc.description == "Custom description"
