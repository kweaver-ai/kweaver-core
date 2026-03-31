"""
会话运行异常
"""

from app.common.exceptions.base_exception import BaseException
from app.common.errors.custom_errors_pkg import ConversationRunningError


class ConversationRunningException(BaseException):
    """
    会话运行异常

    用于表示会话正在运行中，不能执行新的操作
    """

    def __init__(self, error_details: str = "", error_link: str = ""):
        """
        初始化会话运行异常

        Args:
            error_details: 详细错误信息
            error_link: 错误文档链接
        """
        # ConversationRunningError() 会自动从 Config 获取 debug 模式
        super().__init__(
            error=ConversationRunningError(),
            error_details=error_details,
            error_link=error_link,
        )
