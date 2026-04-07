"""
执行器错误单元测试

测试执行器相关错误类。
"""
import pytest

from src.infrastructure.executors.errors import (
    ExecutorError,
    ExecutorConnectionError,
    ExecutorTimeoutError,
    ExecutorUnavailableError,
    ExecutorResponseError,
    ExecutorValidationError,
)


class TestExecutorError:
    """执行器错误基类测试"""

    def test_is_exception(self):
        """测试继承自 Exception"""
        error = ExecutorError("Test error")
        assert isinstance(error, Exception)

    def test_message(self):
        """测试错误消息"""
        error = ExecutorError("Test error message")
        assert str(error) == "Test error message"


class TestExecutorConnectionError:
    """执行器连接错误测试"""

    def test_init(self):
        """测试初始化"""
        error = ExecutorConnectionError(
            executor_url="http://localhost:8080",
            reason="Connection refused"
        )

        assert error.executor_url == "http://localhost:8080"
        assert error.reason == "Connection refused"

    def test_message_format(self):
        """测试消息格式"""
        error = ExecutorConnectionError(
            executor_url="http://localhost:8080",
            reason="Connection refused"
        )

        assert "http://localhost:8080" in str(error)
        assert "Connection refused" in str(error)

    def test_inherits_from_executor_error(self):
        """测试继承自 ExecutorError"""
        error = ExecutorConnectionError("http://localhost:8080", "test")
        assert isinstance(error, ExecutorError)

    def test_empty_reason(self):
        """测试空原因"""
        error = ExecutorConnectionError("http://localhost:8080")

        assert error.executor_url == "http://localhost:8080"
        assert error.reason == ""
        assert "http://localhost:8080" in str(error)


class TestExecutorTimeoutError:
    """执行器超时错误测试"""

    def test_init(self):
        """测试初始化"""
        error = ExecutorTimeoutError(
            executor_url="http://localhost:8080",
            timeout=30.0
        )

        assert error.executor_url == "http://localhost:8080"
        assert error.timeout == 30.0

    def test_message_format(self):
        """测试消息格式"""
        error = ExecutorTimeoutError(
            executor_url="http://localhost:8080",
            timeout=30.0
        )

        assert "http://localhost:8080" in str(error)
        assert "30.0" in str(error)
        assert "timed out" in str(error).lower()

    def test_inherits_from_executor_error(self):
        """测试继承自 ExecutorError"""
        error = ExecutorTimeoutError("http://localhost:8080", 30.0)
        assert isinstance(error, ExecutorError)


class TestExecutorUnavailableError:
    """执行器不可用错误测试"""

    def test_init(self):
        """测试初始化"""
        error = ExecutorUnavailableError(
            executor_url="http://localhost:8080",
            status="unhealthy"
        )

        assert error.executor_url == "http://localhost:8080"
        assert error.status == "unhealthy"

    def test_message_format(self):
        """测试消息格式"""
        error = ExecutorUnavailableError(
            executor_url="http://localhost:8080",
            status="unhealthy"
        )

        assert "http://localhost:8080" in str(error)
        assert "unavailable" in str(error).lower()
        assert "unhealthy" in str(error)

    def test_empty_status(self):
        """测试空状态"""
        error = ExecutorUnavailableError("http://localhost:8080")

        assert error.executor_url == "http://localhost:8080"
        assert error.status == ""

    def test_inherits_from_executor_error(self):
        """测试继承自 ExecutorError"""
        error = ExecutorUnavailableError("http://localhost:8080", "test")
        assert isinstance(error, ExecutorError)


class TestExecutorResponseError:
    """执行器响应错误测试"""

    def test_init(self):
        """测试初始化"""
        error = ExecutorResponseError(
            executor_url="http://localhost:8080",
            status_code=500,
            message="Internal Server Error"
        )

        assert error.executor_url == "http://localhost:8080"
        assert error.status_code == 500
        assert error.message == "Internal Server Error"

    def test_message_format(self):
        """测试消息格式"""
        error = ExecutorResponseError(
            executor_url="http://localhost:8080",
            status_code=500,
            message="Internal Server Error"
        )

        assert "http://localhost:8080" in str(error)
        assert "500" in str(error)
        assert "Internal Server Error" in str(error)

    def test_empty_message(self):
        """测试空消息"""
        error = ExecutorResponseError(
            executor_url="http://localhost:8080",
            status_code=404
        )

        assert error.executor_url == "http://localhost:8080"
        assert error.status_code == 404
        assert error.message == ""

    def test_inherits_from_executor_error(self):
        """测试继承自 ExecutorError"""
        error = ExecutorResponseError("http://localhost:8080", 500, "test")
        assert isinstance(error, ExecutorError)


class TestExecutorValidationError:
    """执行器验证错误测试"""

    def test_init(self):
        """测试初始化"""
        error = ExecutorValidationError(
            executor_url="http://localhost:8080",
            validation_errors=["Missing field: code", "Invalid timeout"]
        )

        assert error.executor_url == "http://localhost:8080"
        assert error.validation_errors == ["Missing field: code", "Invalid timeout"]

    def test_message_format(self):
        """测试消息格式"""
        error = ExecutorValidationError(
            executor_url="http://localhost:8080",
            validation_errors=["Missing field: code"]
        )

        assert "http://localhost:8080" in str(error)
        assert "rejected request" in str(error).lower()
        assert "Missing field: code" in str(error)

    def test_empty_validation_errors(self):
        """测试空验证错误列表"""
        error = ExecutorValidationError(
            executor_url="http://localhost:8080",
            validation_errors=[]
        )

        assert error.executor_url == "http://localhost:8080"
        assert error.validation_errors == []

    def test_inherits_from_executor_error(self):
        """测试继承自 ExecutorError"""
        error = ExecutorValidationError("http://localhost:8080", [])
        assert isinstance(error, ExecutorError)
