# -*- coding: utf-8 -*-
"""单元测试 - validation_handler 模块"""

import pytest
from unittest.mock import MagicMock, patch


class TestHandleParamError:
    """测试 handle_param_error 函数"""

    @pytest.fixture
    def mock_request(self):
        """创建 mock Request"""
        request = MagicMock()
        request.url.path = "/api/test"
        request.method = "POST"
        return request

    def test_handle_missing_param(self, mock_request):
        """测试缺少参数错误"""

        exc = MagicMock()
        exc.errors.return_value = [
            {"type": "missing", "loc": ("body", "name"), "msg": "field required"}
        ]

        with patch(
            "app.router.exception_handler.validation_handler.GetRequestLangFunc"
        ) as mock_lang:
            mock_lang.return_value = lambda x: x

            from app.router.exception_handler.validation_handler import (
                handle_param_error,
            )

            result = handle_param_error(mock_request, exc)

            assert result.status_code == 400

    def test_handle_string_type_error(self, mock_request):
        """测试字符串类型错误"""
        exc = MagicMock()
        exc.errors.return_value = [
            {"type": "string_type", "loc": ("body", "age"), "msg": "not a string"}
        ]

        with patch(
            "app.router.exception_handler.validation_handler.GetRequestLangFunc"
        ) as mock_lang:
            mock_lang.return_value = lambda x: x

            from app.router.exception_handler.validation_handler import (
                handle_param_error,
            )

            result = handle_param_error(mock_request, exc)

            assert result.status_code == 400

    def test_handle_int_type_error(self, mock_request):
        """测试整数类型错误"""
        exc = MagicMock()
        exc.errors.return_value = [
            {"type": "int_type", "loc": ("body", "name"), "msg": "not an int"}
        ]

        with patch(
            "app.router.exception_handler.validation_handler.GetRequestLangFunc"
        ) as mock_lang:
            mock_lang.return_value = lambda x: x

            from app.router.exception_handler.validation_handler import (
                handle_param_error,
            )

            result = handle_param_error(mock_request, exc)

            assert result.status_code == 400

    def test_handle_float_type_error(self, mock_request):
        """测试浮点类型错误"""
        exc = MagicMock()
        exc.errors.return_value = [
            {"type": "float_type", "loc": ("body", "count"), "msg": "not a float"}
        ]

        with patch(
            "app.router.exception_handler.validation_handler.GetRequestLangFunc"
        ) as mock_lang:
            mock_lang.return_value = lambda x: x

            from app.router.exception_handler.validation_handler import (
                handle_param_error,
            )

            result = handle_param_error(mock_request, exc)

            assert result.status_code == 400

    def test_handle_list_type_error(self, mock_request):
        """测试列表类型错误"""
        exc = MagicMock()
        exc.errors.return_value = [
            {"type": "list_type", "loc": ("body", "items"), "msg": "not a list"}
        ]

        with patch(
            "app.router.exception_handler.validation_handler.GetRequestLangFunc"
        ) as mock_lang:
            mock_lang.return_value = lambda x: x

            from app.router.exception_handler.validation_handler import (
                handle_param_error,
            )

            result = handle_param_error(mock_request, exc)

            assert result.status_code == 400

    def test_handle_dict_type_error(self, mock_request):
        """测试字典类型错误"""
        exc = MagicMock()
        exc.errors.return_value = [
            {"type": "dict_type", "loc": ("body", "data"), "msg": "not a dict"}
        ]

        with patch(
            "app.router.exception_handler.validation_handler.GetRequestLangFunc"
        ) as mock_lang:
            mock_lang.return_value = lambda x: x

            from app.router.exception_handler.validation_handler import (
                handle_param_error,
            )

            result = handle_param_error(mock_request, exc)

            assert result.status_code == 400

    def test_handle_bool_type_error(self, mock_request):
        """测试布尔类型错误"""
        exc = MagicMock()
        exc.errors.return_value = [
            {"type": "bool_type", "loc": ("body", "enabled"), "msg": "not a bool"}
        ]

        with patch(
            "app.router.exception_handler.validation_handler.GetRequestLangFunc"
        ) as mock_lang:
            mock_lang.return_value = lambda x: x

            from app.router.exception_handler.validation_handler import (
                handle_param_error,
            )

            result = handle_param_error(mock_request, exc)

            assert result.status_code == 400

    def test_handle_string_too_short_error(self, mock_request):
        """测试字符串太短错误"""
        exc = MagicMock()
        exc.errors.return_value = [
            {
                "type": "string_too_short",
                "loc": ("body", "password"),
                "ctx": {"min_length": 8},
            }
        ]

        with patch(
            "app.router.exception_handler.validation_handler.GetRequestLangFunc"
        ) as mock_lang:
            mock_lang.return_value = lambda x: x

            from app.router.exception_handler.validation_handler import (
                handle_param_error,
            )

            result = handle_param_error(mock_request, exc)

            assert result.status_code == 400

    def test_handle_string_too_long_error(self, mock_request):
        """测试字符串太长错误"""
        exc = MagicMock()
        exc.errors.return_value = [
            {
                "type": "string_too_long",
                "loc": ("body", "name"),
                "ctx": {"max_length": 100},
            }
        ]

        with patch(
            "app.router.exception_handler.validation_handler.GetRequestLangFunc"
        ) as mock_lang:
            mock_lang.return_value = lambda x: x

            from app.router.exception_handler.validation_handler import (
                handle_param_error,
            )

            result = handle_param_error(mock_request, exc)

            assert result.status_code == 400

    def test_handle_greater_than_equal_error(self, mock_request):
        """测试大于等于错误"""
        exc = MagicMock()
        exc.errors.return_value = [
            {
                "type": "greater_than_equal",
                "loc": ("body", "age"),
                "ctx": {"ge": 0},
            }
        ]

        with patch(
            "app.router.exception_handler.validation_handler.GetRequestLangFunc"
        ) as mock_lang:
            mock_lang.return_value = lambda x: x

            from app.router.exception_handler.validation_handler import (
                handle_param_error,
            )

            result = handle_param_error(mock_request, exc)

            assert result.status_code == 400

    def test_handle_less_than_equal_error(self, mock_request):
        """测试小于等于错误"""
        exc = MagicMock()
        exc.errors.return_value = [
            {
                "type": "less_than_equal",
                "loc": ("body", "score"),
                "ctx": {"le": 100},
            }
        ]

        with patch(
            "app.router.exception_handler.validation_handler.GetRequestLangFunc"
        ) as mock_lang:
            mock_lang.return_value = lambda x: x

            from app.router.exception_handler.validation_handler import (
                handle_param_error,
            )

            result = handle_param_error(mock_request, exc)

            assert result.status_code == 400

    def test_handle_unknown_error_type(self, mock_request):
        """测试未知错误类型"""
        exc = MagicMock()
        exc.errors.return_value = [
            {
                "type": "unknown_type",
                "loc": ("body", "field"),
                "msg": "unknown error",
                "ctx": {"key": "value"},
                "input": "test_input",
            }
        ]

        with patch(
            "app.router.exception_handler.validation_handler.GetRequestLangFunc"
        ) as mock_lang:
            mock_lang.return_value = lambda x: x

            from app.router.exception_handler.validation_handler import (
                handle_param_error,
            )

            result = handle_param_error(mock_request, exc)

            assert result.status_code == 400

    def test_handle_multiple_errors(self, mock_request):
        """测试多个错误"""
        exc = MagicMock()
        exc.errors.return_value = [
            {"type": "missing", "loc": ("body", "name"), "msg": "field required"},
            {"type": "int_type", "loc": ("body", "age"), "msg": "not an int"},
        ]

        with patch(
            "app.router.exception_handler.validation_handler.GetRequestLangFunc"
        ) as mock_lang:
            mock_lang.return_value = lambda x: x

            from app.router.exception_handler.validation_handler import (
                handle_param_error,
            )

            result = handle_param_error(mock_request, exc)

            assert result.status_code == 400
