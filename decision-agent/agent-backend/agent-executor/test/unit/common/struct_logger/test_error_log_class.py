# -*- coding: utf-8 -*-
"""
Unit tests for app/common/struct_logger/error_log_class module
"""

from unittest.mock import MagicMock

import pytest


class TestErrorLog:
    """Tests for ErrorLog class"""

    @pytest.mark.asyncio
    async def test_error_log_init_with_message(self):
        """Test ErrorLog initialization with message only"""
        from app.common.struct_logger.error_log_class import ErrorLog

        error_log = ErrorLog("Test error message")

        assert error_log.message == "Test error message"
        assert error_log.caller_frame is None
        assert error_log.caller_traceback is None
        assert "message" in error_log.to_dict()

    @pytest.mark.asyncio
    async def test_error_log_init_with_caller_frame(self):
        """Test ErrorLog initialization with caller frame"""
        from app.common.struct_logger.error_log_class import ErrorLog

        mock_frame = MagicMock()
        mock_frame.f_code.co_filename = "test.py"
        mock_frame.f_lineno = 42

        error_log = ErrorLog("Test error", caller_frame=mock_frame)

        assert error_log.message == "Test error"
        assert error_log.caller_frame == mock_frame

    @pytest.mark.asyncio
    async def test_error_log_init_with_traceback(self):
        """Test ErrorLog initialization with traceback"""
        from app.common.struct_logger.error_log_class import ErrorLog

        error_log = ErrorLog("Test error", caller_traceback="Traceback here")

        assert error_log.message == "Test error"
        assert error_log.caller_traceback == "Traceback here"

    @pytest.mark.asyncio
    async def test_error_log_to_dict(self):
        """Test ErrorLog to_dict method"""
        from app.common.struct_logger.error_log_class import ErrorLog

        error_log = ErrorLog("Test error")

        result = error_log.to_dict()

        assert isinstance(result, dict)
        assert "message" in result
        assert "caller" in result
        assert "stack" in result
        assert "time" in result
        assert result["message"] == "Test error"

    @pytest.mark.asyncio
    async def test_error_log_to_json(self):
        """Test ErrorLog to_json method"""
        from app.common.struct_logger.error_log_class import ErrorLog

        error_log = ErrorLog("Test error 测试")

        result = error_log.to_json()

        assert isinstance(result, str)
        assert "Test error" in result
        assert "测试" in result

    @pytest.mark.asyncio
    async def test_error_log_get_message(self):
        """Test ErrorLog get_message method"""
        from app.common.struct_logger.error_log_class import ErrorLog

        error_log = ErrorLog("Test message")

        assert error_log.get_message() == "Test message"

    @pytest.mark.asyncio
    async def test_error_log_get_caller(self):
        """Test ErrorLog get_caller method"""
        from app.common.struct_logger.error_log_class import ErrorLog

        mock_frame = MagicMock()
        mock_frame.f_code.co_filename = "test.py"
        mock_frame.f_lineno = 42

        error_log = ErrorLog("Test error", caller_frame=mock_frame)

        caller = error_log.get_caller()
        assert "test.py" in caller
        assert "42" in caller

    @pytest.mark.asyncio
    async def test_error_log_get_stack(self):
        """Test ErrorLog get_stack method"""
        from app.common.struct_logger.error_log_class import ErrorLog

        error_log = ErrorLog("Test error", caller_traceback="Stack trace")

        assert error_log.get_stack() == "Stack trace"

    @pytest.mark.asyncio
    async def test_error_log_get_time(self):
        """Test ErrorLog get_time method"""
        from app.common.struct_logger.error_log_class import ErrorLog

        error_log = ErrorLog("Test error")

        time_str = error_log.get_time()
        assert time_str is not None
        assert len(time_str) > 0

    @pytest.mark.asyncio
    async def test_error_log_str(self):
        """Test ErrorLog __str__ method"""
        from app.common.struct_logger.error_log_class import ErrorLog

        error_log = ErrorLog("Test error")

        str_result = str(error_log)

        assert "ErrorLog" in str_result
        assert "Test error" in str_result

    @pytest.mark.asyncio
    async def test_error_log_repr(self):
        """Test ErrorLog __repr__ method"""
        from app.common.struct_logger.error_log_class import ErrorLog

        error_log = ErrorLog("Test error")

        repr_result = repr(error_log)

        assert "ErrorLog" in repr_result
        assert "Test error" in repr_result


class TestCreateErrorLog:
    """Tests for create_error_log function"""

    @pytest.mark.asyncio
    async def test_create_error_log_basic(self):
        """Test create_error_log with basic parameters"""
        from app.common.struct_logger.error_log_class import create_error_log

        error_log = create_error_log("Test error")

        assert error_log.message == "Test error"
        assert isinstance(error_log, object)

    @pytest.mark.asyncio
    async def test_create_error_log_with_traceback(self):
        """Test create_error_log with traceback enabled"""
        from app.common.struct_logger.error_log_class import create_error_log

        error_log = create_error_log("Test error", include_traceback=True)

        assert error_log.message == "Test error"
        assert error_log.caller_traceback is not None


class TestGetErrorLogDict:
    """Tests for get_error_log_dict function"""

    @pytest.mark.asyncio
    async def test_get_error_log_dict_basic(self):
        """Test get_error_log_dict with basic parameters"""
        from app.common.struct_logger.error_log_class import get_error_log_dict

        mock_frame = MagicMock()
        mock_frame.f_code.co_filename = "test.py"
        mock_frame.f_lineno = 42

        result = get_error_log_dict("Test error", mock_frame)

        assert isinstance(result, dict)
        assert result["message"] == "Test error"


class TestGetErrorLogJson:
    """Tests for get_error_log_json function"""

    @pytest.mark.asyncio
    async def test_get_error_log_json_basic(self):
        """Test get_error_log_json with basic parameters"""
        from app.common.struct_logger.error_log_class import get_error_log_json

        mock_frame = MagicMock()
        mock_frame.f_code.co_filename = "test.py"
        mock_frame.f_lineno = 42

        result = get_error_log_json("Test error", mock_frame)

        assert isinstance(result, str)
        assert "Test error" in result


class TestModuleImports:
    """Tests for module imports"""

    @pytest.mark.asyncio
    async def test_module_imports(self):
        """Test that error_log_class module can be imported"""
        from app.common.struct_logger import error_log_class

        assert error_log_class is not None
        assert hasattr(error_log_class, "ErrorLog")
        assert hasattr(error_log_class, "create_error_log")
        assert hasattr(error_log_class, "get_error_log_dict")
        assert hasattr(error_log_class, "get_error_log_json")
