from typing import Any, Dict, Optional


class MemoryException(Exception):
    """记忆服务基础异常类"""

    def __init__(
        self,
        message: str,
        status_code: int = 500,
        error_code: str = "AgentMemory.Internal.Error",
        details: Optional[Dict[str, Any]] = None,
    ):
        self.message = message
        self.status_code = status_code
        self.error_code = error_code
        self.details = details or {}
        super().__init__(self.message)


class MemoryNotFoundError(MemoryException):
    """记忆不存在异常"""

    def __init__(self, memory_id: str):
        super().__init__(
            message=f"Memory not found: {memory_id}",
            status_code=404,
            error_code="AgentMemory.NotFound.Memory",
            details={"memory_id": memory_id},
        )


class MemoryValidationError(MemoryException):
    """记忆验证异常"""

    def __init__(self, message: str, details: Optional[Dict[str, Any]] = None):
        super().__init__(
            message=message,
            status_code=400,
            error_code="AgentMemory.Validation.Error",
            details=details,
        )


class MemoryOperationError(MemoryException):
    """记忆操作异常"""

    def __init__(self, message: str, details: Optional[Dict[str, Any]] = None):
        super().__init__(
            message=message,
            status_code=500,
            error_code="AgentMemory.Operation.Failed",
            details=details,
        )


class MemoryServiceError(MemoryException):
    """记忆服务异常"""

    def __init__(self, message: str, details: Optional[Dict[str, Any]] = None):
        super().__init__(
            message=message,
            status_code=503,
            error_code="AgentMemory.Service.Unavailable",
            details=details,
        )
