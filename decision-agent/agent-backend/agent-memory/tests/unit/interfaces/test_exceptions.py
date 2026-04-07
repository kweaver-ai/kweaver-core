import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "../../.."))

import pytest
from src.interfaces.api.exceptions import (
    MemoryException,
    MemoryNotFoundError,
    MemoryValidationError,
    MemoryOperationError,
    MemoryServiceError,
)


class TestMemoryException:
    def test_memory_exception_creation(self):
        """Test creating a memory exception"""
        exc = MemoryException(
            message="Test error",
            status_code=500,
            error_code="Test.Error",
            details={"key": "value"},
        )
        assert exc.message == "Test error"
        assert exc.status_code == 500
        assert exc.error_code == "Test.Error"
        assert exc.details == {"key": "value"}

    def test_memory_exception_defaults(self):
        """Test memory exception with default values"""
        exc = MemoryException(message="Test error")
        assert exc.message == "Test error"
        assert exc.status_code == 500
        assert exc.error_code == "AgentMemory.Internal.Error"
        assert exc.details == {}

    def test_memory_exception_as_exception(self):
        """Test that memory exception can be raised and caught"""
        with pytest.raises(MemoryException) as exc:
            raise MemoryException(message="Test error")
        assert str(exc.value) == "Test error"

    def test_memory_exception_with_custom_status_code(self):
        """Test memory exception with custom status code"""
        exc = MemoryException(message="Test error", status_code=404)
        assert exc.status_code == 404

    def test_memory_exception_with_none_details(self):
        """Test memory exception with None details"""
        exc = MemoryException(message="Test error", details=None)
        assert exc.details == {}


class TestMemoryNotFoundError:
    def test_memory_not_found_error_creation(self):
        """Test creating a memory not found error"""
        exc = MemoryNotFoundError(memory_id="mem123")
        assert exc.status_code == 404
        assert exc.error_code == "AgentMemory.NotFound.Memory"
        assert exc.details == {"memory_id": "mem123"}
        assert "mem123" in exc.message

    def test_memory_not_found_error_message(self):
        """Test memory not found error message format"""
        exc = MemoryNotFoundError(memory_id="mem456")
        assert "Memory not found: mem456" in str(exc)

    def test_memory_not_found_error_is_memory_exception(self):
        """Test that MemoryNotFoundError is a MemoryException"""
        exc = MemoryNotFoundError(memory_id="mem123")
        assert isinstance(exc, MemoryException)


class TestMemoryValidationError:
    def test_memory_validation_error_creation(self):
        """Test creating a memory validation error"""
        exc = MemoryValidationError(
            message="Invalid input", details={"field": "user_id"}
        )
        assert exc.status_code == 400
        assert exc.error_code == "AgentMemory.Validation.Error"
        assert exc.details == {"field": "user_id"}

    def test_memory_validation_error_defaults(self):
        """Test memory validation error with default values"""
        exc = MemoryValidationError(message="Invalid input")
        assert exc.status_code == 400
        assert exc.error_code == "AgentMemory.Validation.Error"
        assert exc.details == {}

    def test_memory_validation_error_message(self):
        """Test memory validation error message"""
        message = "User ID cannot be empty"
        exc = MemoryValidationError(message=message)
        assert exc.message == message

    def test_memory_validation_error_is_memory_exception(self):
        """Test that MemoryValidationError is a MemoryException"""
        exc = MemoryValidationError(message="Invalid input")
        assert isinstance(exc, MemoryException)


class TestMemoryOperationError:
    def test_memory_operation_error_creation(self):
        """Test creating a memory operation error"""
        exc = MemoryOperationError(
            message="Operation failed", details={"operation": "update"}
        )
        assert exc.status_code == 500
        assert exc.error_code == "AgentMemory.Operation.Failed"
        assert exc.details == {"operation": "update"}

    def test_memory_operation_error_defaults(self):
        """Test memory operation error with default values"""
        exc = MemoryOperationError(message="Operation failed")
        assert exc.status_code == 500
        assert exc.error_code == "AgentMemory.Operation.Failed"
        assert exc.details == {}

    def test_memory_operation_error_with_memory_id_in_details(self):
        """Test memory operation error with memory_id in details"""
        exc = MemoryOperationError(
            message="Delete failed", details={"memory_id": "mem123"}
        )
        assert exc.details["memory_id"] == "mem123"

    def test_memory_operation_error_is_memory_exception(self):
        """Test that MemoryOperationError is a MemoryException"""
        exc = MemoryOperationError(message="Operation failed")
        assert isinstance(exc, MemoryException)


class TestMemoryServiceError:
    def test_memory_service_error_creation(self):
        """Test creating a memory service error"""
        exc = MemoryServiceError(
            message="Service unavailable", details={"service": "database"}
        )
        assert exc.status_code == 503
        assert exc.error_code == "AgentMemory.Service.Unavailable"
        assert exc.details == {"service": "database"}

    def test_memory_service_error_defaults(self):
        """Test memory service error with default values"""
        exc = MemoryServiceError(message="Service unavailable")
        assert exc.status_code == 503
        assert exc.error_code == "AgentMemory.Service.Unavailable"
        assert exc.details == {}

    def test_memory_service_error_status_code(self):
        """Test memory service error status code"""
        exc = MemoryServiceError(message="Service unavailable")
        assert exc.status_code == 503

    def test_memory_service_error_is_memory_exception(self):
        """Test that MemoryServiceError is a MemoryException"""
        exc = MemoryServiceError(message="Service unavailable")
        assert isinstance(exc, MemoryException)


class TestExceptionChaining:
    def test_exception_chaining(self):
        """Test that exceptions can be caught by parent class"""
        with pytest.raises(MemoryException) as exc_info:
            raise MemoryNotFoundError(memory_id="mem123")

        assert isinstance(exc_info.value, MemoryException)
        assert isinstance(exc_info.value, MemoryNotFoundError)

    def test_catching_specific_exception_type(self):
        """Test catching specific exception type"""
        with pytest.raises(MemoryNotFoundError):
            raise MemoryNotFoundError(memory_id="mem123")

    def test_not_catching_different_exception_type(self):
        """Test that different exception types are not caught"""
        try:
            raise MemoryNotFoundError(memory_id="mem123")
        except MemoryValidationError:
            assert False, (
                "Should not catch MemoryNotFoundError as MemoryValidationError"
            )
        except MemoryNotFoundError:
            pass


class TestExceptionDetails:
    def test_exception_details_dict(self):
        """Test exception details as dictionary"""
        exc = MemoryOperationError(
            message="Error", details={"key1": "value1", "key2": "value2"}
        )
        assert len(exc.details) == 2
        assert exc.details["key1"] == "value1"

    def test_exception_details_updates(self):
        """Test that exception details can be updated"""
        exc = MemoryValidationError(message="Error", details={"field": "email"})
        exc.details["error_type"] = "invalid_format"
        assert exc.details["error_type"] == "invalid_format"
