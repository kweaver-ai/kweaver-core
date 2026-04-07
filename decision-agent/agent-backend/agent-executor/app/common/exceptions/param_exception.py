"""
参数异常
"""

from app.common.exceptions.base_exception import BaseException
from app.common.errors.custom_errors_pkg import ParamError


class ParamException(BaseException):
    """
    参数异常

    用于表示参数验证失败的情况
    """

    def __init__(self, error_details: str = "", error_link: str = ""):
        """
        初始化参数异常

        Args:
            error_details: 详细错误信息，描述哪个参数有问题
            error_link: 错误文档链接
        """
        # ParamError() 会自动从 Config 获取 debug 模式
        super().__init__(
            error=ParamError(),
            error_details=error_details,
            error_link=error_link,
        )
