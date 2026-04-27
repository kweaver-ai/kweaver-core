"""
基础设施错误单元测试

测试基础设施层的错误类型。
"""
import pytest

from src.shared.errors.infrastructure import (
    InfrastructureError,
    DatabaseError,
    ConnectionError,
    StorageError,
    HTTPClientError,
    ContainerError,
    KubernetesError,
    MessagingError,
)


class TestInfrastructureError:
    """基础设施错误基类测试"""

    def test_is_exception(self):
        """测试继承自 Exception"""
        error = InfrastructureError("Test error")
        assert isinstance(error, Exception)

    def test_message(self):
        """测试错误消息"""
        error = InfrastructureError("Test error message")
        assert str(error) == "Test error message"
        assert error.message == "Test error message"

    def test_with_original_error(self):
        """测试带原始错误"""
        original = ValueError("Original error")
        error = InfrastructureError("Wrapper error", original_error=original)

        assert error.message == "Wrapper error"
        assert error.original_error is original

    def test_without_original_error(self):
        """测试不带原始错误"""
        error = InfrastructureError("Test error")

        assert error.message == "Test error"
        assert error.original_error is None


class TestDatabaseError:
    """数据库错误测试"""

    def test_inherits_from_infrastructure_error(self):
        """测试继承自 InfrastructureError"""
        error = DatabaseError("Database connection failed")
        assert isinstance(error, InfrastructureError)

    def test_message(self):
        """测试错误消息"""
        error = DatabaseError("Connection timeout")
        assert str(error) == "Connection timeout"

    def test_with_original_error(self):
        """测试带原始错误"""
        original = ConnectionRefusedError("Connection refused")
        error = DatabaseError("Database error", original_error=original)

        assert error.original_error is original


class TestConnectionError:
    """连接错误测试"""

    def test_inherits_from_infrastructure_error(self):
        """测试继承自 InfrastructureError"""
        error = ConnectionError("Connection failed")
        assert isinstance(error, InfrastructureError)

    def test_message(self):
        """测试错误消息"""
        error = ConnectionError("Failed to connect to server")
        assert str(error) == "Failed to connect to server"

    def test_with_original_error(self):
        """测试带原始错误"""
        original = TimeoutError("Connection timeout")
        error = ConnectionError("Connection error", original_error=original)

        assert error.original_error is original


class TestStorageError:
    """存储错误测试"""

    def test_inherits_from_infrastructure_error(self):
        """测试继承自 InfrastructureError"""
        error = StorageError("Storage error")
        assert isinstance(error, InfrastructureError)

    def test_message(self):
        """测试错误消息"""
        error = StorageError("Failed to upload file")
        assert str(error) == "Failed to upload file"

    def test_with_original_error(self):
        """测试带原始错误"""
        original = IOError("Disk full")
        error = StorageError("Storage error", original_error=original)

        assert error.original_error is original


class TestHTTPClientError:
    """HTTP 客户端错误测试"""

    def test_inherits_from_infrastructure_error(self):
        """测试继承自 InfrastructureError"""
        error = HTTPClientError("HTTP error")
        assert isinstance(error, InfrastructureError)

    def test_message(self):
        """测试错误消息"""
        error = HTTPClientError("Request timeout")
        assert str(error) == "Request timeout"

    def test_with_original_error(self):
        """测试带原始错误"""
        original = Exception("Network error")
        error = HTTPClientError("HTTP client error", original_error=original)

        assert error.original_error is original


class TestContainerError:
    """容器错误测试"""

    def test_inherits_from_infrastructure_error(self):
        """测试继承自 InfrastructureError"""
        error = ContainerError("Container error")
        assert isinstance(error, InfrastructureError)

    def test_message(self):
        """测试错误消息"""
        error = ContainerError("Failed to start container")
        assert str(error) == "Failed to start container"

    def test_with_original_error(self):
        """测试带原始错误"""
        original = RuntimeError("Docker daemon not running")
        error = ContainerError("Container error", original_error=original)

        assert error.original_error is original


class TestKubernetesError:
    """Kubernetes 错误测试"""

    def test_inherits_from_infrastructure_error(self):
        """测试继承自 InfrastructureError"""
        error = KubernetesError("Kubernetes error")
        assert isinstance(error, InfrastructureError)

    def test_message(self):
        """测试错误消息"""
        error = KubernetesError("Pod failed to start")
        assert str(error) == "Pod failed to start"

    def test_with_original_error(self):
        """测试带原始错误"""
        original = Exception("API server unavailable")
        error = KubernetesError("Kubernetes error", original_error=original)

        assert error.original_error is original


class TestMessagingError:
    """消息队列错误测试"""

    def test_inherits_from_infrastructure_error(self):
        """测试继承自 InfrastructureError"""
        error = MessagingError("Messaging error")
        assert isinstance(error, InfrastructureError)

    def test_message(self):
        """测试错误消息"""
        error = MessagingError("Failed to publish message")
        assert str(error) == "Failed to publish message"

    def test_with_original_error(self):
        """测试带原始错误"""
        original = Exception("Queue not found")
        error = MessagingError("Messaging error", original_error=original)

        assert error.original_error is original


class TestErrorHierarchy:
    """错误层次结构测试"""

    def test_all_errors_inherit_from_infrastructure_error(self):
        """测试所有错误都继承自 InfrastructureError"""
        errors = [
            DatabaseError("test"),
            ConnectionError("test"),
            StorageError("test"),
            HTTPClientError("test"),
            ContainerError("test"),
            KubernetesError("test"),
            MessagingError("test"),
        ]

        for error in errors:
            assert isinstance(error, InfrastructureError)
            assert isinstance(error, Exception)

    def test_errors_can_be_caught_by_base_class(self):
        """测试可以通过基类捕获所有错误"""
        errors_to_raise = [
            DatabaseError("db error"),
            ContainerError("container error"),
            StorageError("storage error"),
        ]

        for error in errors_to_raise:
            try:
                raise error
            except InfrastructureError as e:
                assert e is error
            except Exception:
                pytest.fail("Error should be caught by InfrastructureError")
