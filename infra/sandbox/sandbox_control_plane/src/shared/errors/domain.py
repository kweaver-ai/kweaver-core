"""
领域错误

定义领域层的错误类型。
"""
from typing import Any, Optional


class DomainError(Exception):
    """领域错误基类"""

    def __init__(self, message: str, details: Optional[dict[str, Any]] = None):
        self.message = message
        self.details = details or {}
        super().__init__(self.message)


class NotFoundError(DomainError):
    """未找到错误"""
    pass


class ValidationError(DomainError):
    """验证错误"""
    pass


class InvalidStatusError(DomainError):
    """无效状态错误"""
    pass


class ResourceLimitError(DomainError):
    """资源限制错误"""
    pass


class SessionExpiredError(DomainError):
    """会话过期错误"""
    pass


class ExecutionTimeoutError(DomainError):
    """执行超时错误"""
    pass


class ExecutionCrashedError(DomainError):
    """执行崩溃错误"""
    pass


class TemplateNotFoundError(DomainError):
    """模板未找到错误"""
    pass


class NodeUnavailableError(DomainError):
    """节点不可用错误"""
    pass


class ConflictError(DomainError):
    """资源冲突错误"""
    pass
