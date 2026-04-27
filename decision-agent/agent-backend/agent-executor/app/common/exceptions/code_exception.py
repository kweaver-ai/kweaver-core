"""
代码相关异常
通用的代码执行异常
"""

from app.common.exceptions.base_exception import BaseException
from app.common.errors.api_error_class import APIError


class CodeException(BaseException):
    """
    代码异常

    用于表示代码执行过程中的各类错误
    """

    def __init__(
        self,
        error: APIError,
        error_details: str = "",
        error_link: str = "",
        description: str = "",
    ):
        """
        初始化代码异常

        Args:
            error: Error 对象
            error_details: 详细错误信息
            error_link: 错误文档链接
            description: 自定义描述
        """
        super().__init__(error, error_details, error_link, description)
