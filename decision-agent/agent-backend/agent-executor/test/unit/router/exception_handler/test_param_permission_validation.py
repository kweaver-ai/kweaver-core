# -*- coding:utf-8 -*-
"""单元测试 - 参数异常处理器"""

import pytest
from unittest.mock import patch, Mock
from fastapi import Request
from app.common.errors import ParamException, AgentPermissionException
from fastapi.exceptions import RequestValidationError


@pytest.mark.asyncio
class TestHandleParamException:
    """测试 handle_param_exception 函数"""

    async def test_handles_param_exception(self):
        """测试处理参数异常"""
        from app.router.exception_handler.param_handler import handle_param_exception

        request = Mock(spec=Request)
        exc = ParamException("Test param error")

        with patch("app.router.exception_handler.param_handler.struct_logger"):
            response = handle_param_exception(request, exc)

        assert response.status_code == 400

    async def test_returns_json_response(self):
        """测试返回JSON响应"""
        from app.router.exception_handler.param_handler import handle_param_exception

        request = Mock(spec=Request)
        exc = ParamException("Test error")

        with patch("app.router.exception_handler.param_handler.struct_logger"):
            response = handle_param_exception(request, exc)

        # Response should be JSON
        assert response.media_type == "application/json"
        assert response.status_code == 400

    async def test_logs_error(self):
        """测试错误日志记录"""
        from app.router.exception_handler.param_handler import handle_param_exception

        request = Mock(spec=Request)
        request.url.path = "/api/test"
        request.method = "POST"
        exc = ParamException("Test error")

        with patch(
            "app.router.exception_handler.param_handler.struct_logger"
        ) as mock_logger:
            handle_param_exception(request, exc)

            # Check that error was logged
            mock_logger.error.assert_called_once()


@pytest.mark.asyncio
class TestHandlePermissionException:
    """测试 handle_permission_exception 函数"""

    async def test_handles_permission_exception(self):
        """测试处理权限异常"""
        from app.router.exception_handler.permission_handler import (
            handle_permission_exception,
        )

        request = Mock(spec=Request)
        exc = AgentPermissionException(agent_id="agent123", user_id="user456")

        with patch("app.router.exception_handler.permission_handler.struct_logger"):
            response = handle_permission_exception(request, exc)

        assert response.status_code == 403

    async def test_returns_forbidden(self):
        """测试返回403禁止访问"""
        from app.router.exception_handler.permission_handler import (
            handle_permission_exception,
        )

        request = Mock(spec=Request)
        exc = AgentPermissionException(agent_id="agent123", user_id="user456")

        with patch("app.router.exception_handler.permission_handler.struct_logger"):
            response = handle_permission_exception(request, exc)

        assert response.status_code == 403

    async def test_logs_permission_error(self):
        """测试权限错误日志记录"""
        from app.router.exception_handler.permission_handler import (
            handle_permission_exception,
        )

        request = Mock(spec=Request)
        request.url.path = "/api/protected"
        request.method = "GET"
        exc = AgentPermissionException(agent_id="agent123", user_id="user456")

        with patch(
            "app.router.exception_handler.permission_handler.struct_logger"
        ) as mock_logger:
            handle_permission_exception(request, exc)

            # Check that error was logged
            mock_logger.error.assert_called_once()


@pytest.mark.asyncio
class TestHandleValidationException:
    """测试 handle_param_error 函数 (validation_handler.py)"""

    async def test_handles_validation_error_missing_field(self):
        """测试处理缺失字段错误"""
        from app.router.exception_handler.validation_handler import handle_param_error

        request = Mock(spec=Request)
        request.url = Mock(path="/api/test")
        request.method = "POST"

        # Create a mock validation error with proper structure
        exc = Mock(spec=RequestValidationError)
        exc.errors.return_value = [
            {"loc": ("body", "field1"), "type": "missing", "msg": "Field1 is required"},
        ]

        with patch("app.router.exception_handler.validation_handler.struct_logger"):
            response = handle_param_error(request, exc)

        assert response.status_code == 400

    async def test_handles_string_too_short(self):
        """测试处理字符串过短错误"""
        from app.router.exception_handler.validation_handler import handle_param_error

        request = Mock(spec=Request)
        request.url = Mock(path="/api/register")
        request.method = "POST"

        exc = Mock(spec=RequestValidationError)
        exc.errors.return_value = [
            {
                "loc": ("body", "username"),
                "type": "string_too_short",
                "ctx": {"min_length": 3},
            }
        ]

        with patch("app.router.exception_handler.validation_handler.struct_logger"):
            response = handle_param_error(request, exc)

        assert response.status_code == 400

    async def test_with_multiple_errors(self):
        """测试多个验证错误"""
        from app.router.exception_handler.validation_handler import handle_param_error

        request = Mock(spec=Request)
        request.url = Mock(path="/api/register")
        request.method = "POST"

        exc = Mock(spec=RequestValidationError)
        exc.errors.return_value = [
            {
                "loc": ("body", "email"),
                "type": "value_error.email",
                "msg": "Invalid email",
            },
            {
                "loc": ("body", "age"),
                "type": "type_error.integer",
                "msg": "Not an integer",
            },
        ]

        with patch("app.router.exception_handler.validation_handler.struct_logger"):
            response = handle_param_error(request, exc)

        assert response.status_code == 400

    async def test_logs_validation_error(self):
        """测试错误日志记录"""
        from app.router.exception_handler.validation_handler import handle_param_error

        request = Mock(spec=Request)
        request.url = Mock(path="/api/register")
        request.method = "POST"

        exc = Mock(spec=RequestValidationError)
        exc.errors.return_value = [
            {
                "loc": ("body", "username"),
                "type": "string_too_short",
                "ctx": {"min_length": 3},
            }
        ]

        with patch(
            "app.router.exception_handler.validation_handler.struct_logger"
        ) as mock_logger:
            handle_param_error(request, exc)

            # Check that error was logged
            mock_logger.error.assert_called_once()
