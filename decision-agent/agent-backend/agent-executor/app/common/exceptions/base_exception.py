"""
基础异常类
所有自定义异常都应继承此类
"""

import json
import gettext
from typing import Callable, Optional

from app.common.errors.api_error_class import APIError


class BaseException(Exception):
    """
    基础异常类

    所有自定义异常的基类，提供统一的错误格式化功能

    Attributes:
        error: Error 对象，包含错误代码、描述、解决方案等信息
        error_details: 详细的错误信息
        error_link: 相关错误文档链接
        description: 自定义描述（优先级高于 error 中的描述）
    """

    error: APIError
    error_details: Optional[str] = None
    error_link: Optional[str] = None
    description: Optional[str] = None

    def __init__(
        self,
        error: APIError,
        error_details: str = "",
        error_link: str = "",
        description: str = "",
    ):
        """
        初始化异常

        Args:
            error: Error 对象
            error_details: 详细错误信息
            error_link: 错误文档链接
            description: 自定义描述
        """
        self.error = error

        super().__init__(str(self.error))
        self.error_details = error_details
        self.error_link = error_link
        self.description = description

    def FormatHttpError(self, tr: Callable[..., str] = gettext.gettext) -> dict:
        """
        格式化为 HTTP 错误响应格式

        Args:
            tr: 翻译函数，用于国际化

        Returns:
            dict: 包含以下字段的字典：
                - description: 错误描述
                - error_code: 错误代码
                - error_details: 详细错误信息
                - error_link: 错误文档链接
                - solution: 解决方案
                - trace: 调试追踪信息（仅在 debug 模式下）
        """
        error_code = self.error.error_code

        # 确定描述优先级：description > error_details > error.description > error_code
        if self.description:
            description = tr(self.description)
        elif self.error.description:
            description = tr(self.error.description)
        elif self.error_details:
            description = tr(self.error_details)
        else:
            description = error_code

        # 确定详细信息
        if self.error_details:
            error_details = tr(self.error_details)
        else:
            error_details = tr(self.error.description)

        # 构建基本响应
        response = {
            "description": description,
            "error_code": error_code,
            "error_details": error_details,
            "error_link": self.error_link,
            "solution": tr(self.error.solution),
        }

        # 如果 Error 对象包含 trace 信息，则添加到响应中（仅在 debug 模式下）
        if self.error.trace:
            response["trace"] = self.error.trace

        return response

    def FormatLogError(self, tr: Callable[..., str] = gettext.gettext) -> dict:
        """
        格式化为日志错误响应格式

        Args:
            tr: 翻译函数，用于国际化

        Returns:
            dict: 包含以下字段的字典：
                - description: 错误描述
                - error_code: 错误代码
                - error_details: 详细错误信息
                - error_link: 错误文档链接
                - solution: 解决方案
        """
        response = self.FormatHttpError(tr)
        # 去除trace字段
        response.pop("trace", None)
        return response

    def __repr__(self) -> str:
        return json.dumps(self.FormatLogError(), ensure_ascii=False)

    def __str__(self) -> str:
        return json.dumps(self.FormatLogError(), ensure_ascii=False)
