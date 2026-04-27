"""单元测试 - logic/agent_core_logic_v2/exception 模块"""

import pytest
from unittest.mock import patch


@pytest.mark.asyncio
class TestExceptionHandler:
    """测试 ExceptionHandler 类"""

    @patch("app.logic.agent_core_logic_v2.exception.get_format_error_info")
    @patch("app.logic.agent_core_logic_v2.exception.struct_logger")
    @patch("app.logic.agent_core_logic_v2.exception.log_oper")
    @patch("app.logic.agent_core_logic_v2.exception.exception_logger")
    async def test_handle_exception_basic(
        self, m_exception_logger, m_log_oper, m_struct_logger, m_format_error
    ):
        """测试基本异常处理"""
        m_format_error.return_value = {
            "error_code": "TEST_ERROR",
            "message": "Test error",
        }

        from app.logic.agent_core_logic_v2.exception import ExceptionHandler

        exc = Exception("Test exception")
        res = {}
        headers = {"Authorization": "Bearer token123"}

        await ExceptionHandler.handle_exception(exc, res, headers)

        assert res["status"] == "Error"
        assert "error" in res
        m_exception_logger.log_exception.assert_called_once()
        m_format_error.assert_called_once_with(headers, exc)

    @patch("app.logic.agent_core_logic_v2.exception.get_format_error_info")
    @patch("app.logic.agent_core_logic_v2.exception.struct_logger")
    @patch("app.logic.agent_core_logic_v2.exception.log_oper")
    @patch("app.logic.agent_core_logic_v2.exception.exception_logger")
    async def test_handle_exception_with_dict_res(
        self, m_exception_logger, m_log_oper, m_struct_logger, m_format_error
    ):
        """测试字典结果的异常处理"""
        m_format_error.return_value = {
            "error_code": "TEST_ERROR",
            "message": "Test error",
        }
        m_log_oper.get_error_log.return_value = "Error log"

        from app.logic.agent_core_logic_v2.exception import ExceptionHandler

        exc = ValueError("Invalid value")
        res = {}
        headers = {}

        await ExceptionHandler.handle_exception(exc, res, headers)

        assert res["status"] == "Error"
        assert "error" in res
        assert res["error"]["error_code"] == "TEST_ERROR"

    @patch("app.logic.agent_core_logic_v2.exception.get_format_error_info")
    @patch("app.logic.agent_core_logic_v2.exception.struct_logger")
    @patch("app.logic.agent_core_logic_v2.exception.log_oper")
    @patch("app.logic.agent_core_logic_v2.exception.exception_logger")
    async def test_handle_exception_with_custom_error(
        self, m_exception_logger, m_log_oper, m_struct_logger, m_format_error
    ):
        """测试自定义错误异常处理"""
        m_format_error.return_value = {
            "error_code": "CUSTOM_ERROR",
            "message": "Custom error message",
            "details": "Error details",
        }
        m_log_oper.get_error_log.return_value = "Custom error log"

        from app.logic.agent_core_logic_v2.exception import ExceptionHandler

        exc = RuntimeError("Runtime error occurred")
        res = {}
        headers = {"X-Request-ID": "12345"}

        await ExceptionHandler.handle_exception(exc, res, headers)

        assert res["status"] == "Error"
        assert res["error"]["error_code"] == "CUSTOM_ERROR"
        m_exception_logger.log_exception.assert_called_once()

    @patch("app.logic.agent_core_logic_v2.exception.get_format_error_info")
    @patch("app.logic.agent_core_logic_v2.exception.struct_logger")
    @patch("app.logic.agent_core_logic_v2.exception.log_oper")
    @patch("app.logic.agent_core_logic_v2.exception.exception_logger")
    async def test_handle_exception_logs_to_console(
        self, m_exception_logger, m_log_oper, m_struct_logger, m_format_error
    ):
        """测试异常被记录到控制台"""
        m_format_error.return_value = {"error_code": "TEST_ERROR"}
        m_log_oper.get_error_log.return_value = "Formatted error log"

        from app.logic.agent_core_logic_v2.exception import ExceptionHandler

        exc = Exception("Console log test")
        res = {}
        headers = {}

        await ExceptionHandler.handle_exception(exc, res, headers)

        # Verify console logger was called
        m_struct_logger.console_logger.error.assert_called_once()
        error_log_arg = m_struct_logger.console_logger.error.call_args[0][0]
        assert error_log_arg == "Formatted error log"
