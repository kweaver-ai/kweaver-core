"""测试 - 异常处理器整合"""

import pytest
from unittest.mock import MagicMock, patch, Mock
from fastapi import Request


@pytest.mark.asyncio
class TestCodeExceptionHandler:
    """测试代码异常处理器"""

    async def test_handles_code_exception(self):
        """测试处理代码异常"""
        from app.router.exception_handler.code_handler import handle_code_exception
        from app.common.errors import CodeException, APIError

        request = MagicMock(spec=Request)
        request.headers = {}

        error = APIError(
            error_code="TEST_ERROR", description="Test error", solution="Test solution"
        )
        exc = CodeException(error)

        with patch("app.router.exception_handler.code_handler.traceback.print_exc"):
            with patch(
                "app.router.exception_handler.code_handler.log_oper.get_error_log"
            ):
                with patch(
                    "app.router.exception_handler.code_handler.struct_logger.error"
                ):
                    response = handle_code_exception(request, exc)

        assert response.status_code == 500


@pytest.mark.asyncio
class TestParamExceptionHandler:
    """测试参数异常处理器"""

    async def test_handles_param_exception(self):
        """测试处理参数异常"""
        from app.router.exception_handler.param_handler import handle_param_exception
        from app.common.errors import ParamException

        request = MagicMock(spec=Request)
        request.headers = {}

        exc = ParamException("Missing parameter")

        with patch("app.router.exception_handler.param_handler.log_oper.get_error_log"):
            with patch(
                "app.router.exception_handler.param_handler.struct_logger.error"
            ):
                response = handle_param_exception(request, exc)

        assert response.status_code == 400


@pytest.mark.asyncio
class TestPermissionExceptionHandler:
    """测试权限异常处理器"""

    async def test_handles_permission_exception(self):
        """测试处理权限异常"""
        from app.router.exception_handler.permission_handler import (
            handle_permission_exception,
        )
        from app.common.errors import AgentPermissionException

        request = MagicMock(spec=Request)
        request.headers = {}

        exc = AgentPermissionException("agent123", "user456")

        with patch(
            "app.router.exception_handler.permission_handler.log_oper.get_error_log"
        ):
            with patch(
                "app.router.exception_handler.permission_handler.struct_logger.error"
            ):
                response = handle_permission_exception(request, exc)

        assert response.status_code == 403


@pytest.mark.asyncio
class TestValidationHandler:
    """测试验证异常处理器"""

    async def test_handles_validation_error(self):
        """测试处理验证错误"""
        from app.router.exception_handler.validation_handler import handle_param_error
        from fastapi.exceptions import RequestValidationError

        request = MagicMock(spec=Request)
        request.url = MagicMock(path="/api/test")
        request.method = "POST"
        request.headers = MagicMock(items=Mock(return_value=[]))

        # Create mock error
        error_dict = {
            "loc": ("body", "field1"),
            "type": "missing",
            "msg": "Field required",
        }
        exc = RequestValidationError([error_dict])

        with patch(
            "app.router.exception_handler.validation_handler.struct_logger.error"
        ):
            response = handle_param_error(request, exc)

        assert response.status_code == 400

    async def test_handles_string_type_error(self):
        """测试处理字符串类型错误"""
        from app.router.exception_handler.validation_handler import handle_param_error
        from fastapi.exceptions import RequestValidationError

        request = MagicMock(spec=Request)
        request.url = MagicMock(path="/api/test")
        request.method = "POST"
        request.headers = MagicMock(items=Mock(return_value=[]))

        error_dict = {
            "loc": ("body", "name"),
            "type": "string_type",
            "msg": "Field must be a string",
        }
        exc = RequestValidationError([error_dict])

        with patch(
            "app.router.exception_handler.validation_handler.struct_logger.error"
        ):
            response = handle_param_error(request, exc)

        assert response.status_code == 400

    async def test_handles_int_type_error(self):
        """测试处理整数类型错误"""
        from app.router.exception_handler.validation_handler import handle_param_error
        from fastapi.exceptions import RequestValidationError

        request = MagicMock(spec=Request)
        request.url = MagicMock(path="/api/test")
        request.method = "POST"
        request.headers = MagicMock(items=Mock(return_value=[]))

        error_dict = {
            "loc": ("body", "age"),
            "type": "int_type",
            "msg": "Field must be an integer",
        }
        exc = RequestValidationError([error_dict])

        with patch(
            "app.router.exception_handler.validation_handler.struct_logger.error"
        ):
            response = handle_param_error(request, exc)

        assert response.status_code == 400

    async def test_handles_list_type_error(self):
        """测试处理列表类型错误"""
        from app.router.exception_handler.validation_handler import handle_param_error
        from fastapi.exceptions import RequestValidationError

        request = MagicMock(spec=Request)
        request.url = MagicMock(path="/api/test")
        request.method = "POST"
        request.headers = MagicMock(items=Mock(return_value=[]))

        error_dict = {
            "loc": ("body", "items"),
            "type": "list_type",
            "msg": "Field must be a list",
        }
        exc = RequestValidationError([error_dict])

        with patch(
            "app.router.exception_handler.validation_handler.struct_logger.error"
        ):
            response = handle_param_error(request, exc)

        assert response.status_code == 400

    async def test_handles_dict_type_error(self):
        """测试处理字典类型错误"""
        from app.router.exception_handler.validation_handler import handle_param_error
        from fastapi.exceptions import RequestValidationError

        request = MagicMock(spec=Request)
        request.url = MagicMock(path="/api/test")
        request.method = "POST"
        request.headers = MagicMock(items=Mock(return_value=[]))

        error_dict = {
            "loc": ("body", "metadata"),
            "type": "dict_type",
            "msg": "Field must be a dict",
        }
        exc = RequestValidationError([error_dict])

        with patch(
            "app.router.exception_handler.validation_handler.struct_logger.error"
        ):
            response = handle_param_error(request, exc)

        assert response.status_code == 400

    async def test_handles_bool_type_error(self):
        """测试处理布尔类型错误"""
        from app.router.exception_handler.validation_handler import handle_param_error
        from fastapi.exceptions import RequestValidationError

        request = MagicMock(spec=Request)
        request.url = MagicMock(path="/api/test")
        request.method = "POST"
        request.headers = MagicMock(items=Mock(return_value=[]))

        error_dict = {
            "loc": ("body", "is_active"),
            "type": "bool_type",
            "msg": "Field must be a boolean",
        }
        exc = RequestValidationError([error_dict])

        with patch(
            "app.router.exception_handler.validation_handler.struct_logger.error"
        ):
            response = handle_param_error(request, exc)

        assert response.status_code == 400

    async def test_handles_string_too_short_error(self):
        """测试处理字符串过短错误"""
        from app.router.exception_handler.validation_handler import handle_param_error
        from fastapi.exceptions import RequestValidationError

        request = MagicMock(spec=Request)
        request.url = MagicMock(path="/api/test")
        request.method = "POST"
        request.headers = MagicMock(items=Mock(return_value=[]))

        error_dict = {
            "loc": ("body", "name"),
            "type": "string_too_short",
            "msg": "String is too short",
            "ctx": {"min_length": 3},
        }
        exc = RequestValidationError([error_dict])

        with patch(
            "app.router.exception_handler.validation_handler.struct_logger.error"
        ):
            response = handle_param_error(request, exc)

        assert response.status_code == 400

    async def test_handles_string_too_long_error(self):
        """测试处理字符串过长错误"""
        from app.router.exception_handler.validation_handler import handle_param_error
        from fastapi.exceptions import RequestValidationError

        request = MagicMock(spec=Request)
        request.url = MagicMock(path="/api/test")
        request.method = "POST"
        request.headers = MagicMock(items=Mock(return_value=[]))

        error_dict = {
            "loc": ("body", "name"),
            "type": "string_too_long",
            "msg": "String is too long",
            "ctx": {"max_length": 50},
        }
        exc = RequestValidationError([error_dict])

        with patch(
            "app.router.exception_handler.validation_handler.struct_logger.error"
        ):
            response = handle_param_error(request, exc)

        assert response.status_code == 400

    async def test_handles_greater_than_equal_error(self):
        """测试处理大于等于错误"""
        from app.router.exception_handler.validation_handler import handle_param_error
        from fastapi.exceptions import RequestValidationError

        request = MagicMock(spec=Request)
        request.url = MagicMock(path="/api/test")
        request.method = "POST"
        request.headers = MagicMock(items=Mock(return_value=[]))

        error_dict = {
            "loc": ("body", "age"),
            "type": "greater_than_equal",
            "msg": "Value is too small",
            "ctx": {"ge": 0},
        }
        exc = RequestValidationError([error_dict])

        with patch(
            "app.router.exception_handler.validation_handler.struct_logger.error"
        ):
            response = handle_param_error(request, exc)

        assert response.status_code == 400

    async def test_handles_less_than_equal_error(self):
        """测试处理小于等于错误"""
        from app.router.exception_handler.validation_handler import handle_param_error
        from fastapi.exceptions import RequestValidationError

        request = MagicMock(spec=Request)
        request.url = MagicMock(path="/api/test")
        request.method = "POST"
        request.headers = MagicMock(items=Mock(return_value=[]))

        error_dict = {
            "loc": ("body", "age"),
            "type": "less_than_equal",
            "msg": "Value is too large",
            "ctx": {"le": 100},
        }
        exc = RequestValidationError([error_dict])

        with patch(
            "app.router.exception_handler.validation_handler.struct_logger.error"
        ):
            response = handle_param_error(request, exc)

        assert response.status_code == 400

    async def test_handles_json_invalid_error(self):
        """测试处理JSON格式错误"""
        from app.router.exception_handler.validation_handler import handle_param_error
        from fastapi.exceptions import RequestValidationError

        request = MagicMock(spec=Request)
        request.url = MagicMock(path="/api/test")
        request.method = "POST"
        request.headers = MagicMock(items=Mock(return_value=[]))

        error_dict = {"loc": ("body",), "type": "json_invalid", "msg": "Invalid JSON"}
        exc = RequestValidationError([error_dict])

        with patch(
            "app.router.exception_handler.validation_handler.struct_logger.error"
        ):
            response = handle_param_error(request, exc)

        assert response.status_code == 400

    async def test_handles_unknown_error_type(self):
        """测试处理未知错误类型"""
        from app.router.exception_handler.validation_handler import handle_param_error
        from fastapi.exceptions import RequestValidationError

        request = MagicMock(spec=Request)
        request.url = MagicMock(path="/api/test")
        request.method = "POST"
        request.headers = MagicMock(items=Mock(return_value=[]))

        error_dict = {
            "loc": ("body", "field"),
            "type": "unknown_type",
            "msg": "Unknown error",
            "input": "test_value",
        }
        exc = RequestValidationError([error_dict])

        with patch(
            "app.router.exception_handler.validation_handler.struct_logger.error"
        ):
            with patch(
                "app.router.exception_handler.validation_handler.inspect.stack",
                return_value=[],
            ):
                response = handle_param_error(request, exc)

        assert response.status_code == 400

    async def test_handles_multiple_errors(self):
        """测试处理多个验证错误"""
        from app.router.exception_handler.validation_handler import handle_param_error
        from fastapi.exceptions import RequestValidationError

        request = MagicMock(spec=Request)
        request.url = MagicMock(path="/api/test")
        request.method = "POST"
        request.headers = MagicMock(items=Mock(return_value=[]))

        error_dict1 = {
            "loc": ("body", "name"),
            "type": "missing",
            "msg": "Field required",
        }
        error_dict2 = {
            "loc": ("body", "age"),
            "type": "int_type",
            "msg": "Must be an integer",
        }
        exc = RequestValidationError([error_dict1, error_dict2])

        with patch(
            "app.router.exception_handler.validation_handler.struct_logger.error"
        ):
            response = handle_param_error(request, exc)

        assert response.status_code == 400

    async def test_handles_list_unique_items_error(self):
        """测试处理列表重复项错误"""
        from app.router.exception_handler.validation_handler import handle_param_error
        from fastapi.exceptions import RequestValidationError

        request = MagicMock(spec=Request)
        request.url = MagicMock(path="/api/test")
        request.method = "POST"
        request.headers = MagicMock(items=Mock(return_value=[]))

        error_dict = {
            "loc": ("body", "tags"),
            "type": "value_error.list.unique_items",
            "msg": "List has duplicate items",
        }
        exc = RequestValidationError([error_dict])

        with patch(
            "app.router.exception_handler.validation_handler.struct_logger.error"
        ):
            response = handle_param_error(request, exc)

        assert response.status_code == 400
