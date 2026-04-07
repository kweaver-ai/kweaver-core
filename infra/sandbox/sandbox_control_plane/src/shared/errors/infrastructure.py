"""
基础设施错误

定义基础设施层的错误类型。
"""
from typing import Optional


class InfrastructureError(Exception):
    """基础设施错误基类"""

    def __init__(
        self,
        message: str,
        original_error: Optional[Exception] = None
    ):
        self.message = message
        self.original_error = original_error
        super().__init__(self.message)


class DatabaseError(InfrastructureError):
    """数据库错误"""
    pass


class ConnectionError(InfrastructureError):
    """连接错误"""
    pass


class StorageError(InfrastructureError):
    """存储错误"""
    pass


class HTTPClientError(InfrastructureError):
    """HTTP 客户端错误"""
    pass


class ContainerError(InfrastructureError):
    """容器错误"""
    pass


class KubernetesError(InfrastructureError):
    """Kubernetes 错误"""
    pass


class MessagingError(InfrastructureError):
    """消息队列错误"""
    pass
