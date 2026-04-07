# -*- coding:utf-8 -*-
"""单元测试 - 代码异常处理器"""

import pytest
from unittest.mock import patch, Mock
from fastapi import Request
from app.common.errors import CodeException
from app.common.errors.api_error_class import APIError


@pytest.mark.asyncio
class TestHandleCodeException:
    """测试 handle_code_exception 函数"""

    async def test_handles_code_exception(self):
        """测试处理代码异常"""
        from app.router.exception_handler.code_handler import handle_code_exception

        request = Mock(spec=Request)
        error = APIError(
            error_code="AgentExecutor.InternalServerError.CodeError",
            description="Test code error",
            solution="Please check your code.",
        )
        exc = CodeException(error)

        with patch("app.router.exception_handler.code_handler.struct_logger"):
            response = handle_code_exception(request, exc)

        assert response.status_code == 500

    async def test_logs_error_correctly(self):
        """测试正确记录错误日志"""
        from app.router.exception_handler.code_handler import handle_code_exception

        request = Mock(spec=Request)
        request.url.path = "/api/test"
        request.method = "POST"
        error = APIError(
            error_code="AgentExecutor.InternalServerError.CodeError",
            description="Test error",
            solution="Please check your code.",
        )
        exc = CodeException(error)

        with patch(
            "app.router.exception_handler.code_handler.struct_logger"
        ) as mock_logger:
            handle_code_exception(request, exc)

            # Check that error was logged
            mock_logger.error.assert_called_once()

    async def test_returns_json_response(self):
        """测试返回JSON响应"""
        from app.router.exception_handler.code_handler import handle_code_exception

        request = Mock(spec=Request)
        error = APIError(
            error_code="AgentExecutor.InternalServerError.CodeError",
            description="Test error",
            solution="Please check your code.",
        )
        exc = CodeException(error)

        with patch("app.router.exception_handler.code_handler.struct_logger"):
            response = handle_code_exception(request, exc)

        # Response should be JSON
        assert response.media_type == "application/json"
