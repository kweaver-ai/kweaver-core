"""
错误定义基类
提供统一的错误结构和调试支持
"""

import sys
import traceback
from typing import Optional


class APIError:
    """
    错误基类

    错误是一个字典结构，包含以下字段：
    - ErrorCode: 错误代码
    - Description: 错误描述
    - Solution: 解决方案
    - Trace: 调试追踪信息（仅在 debug 模式下）

    使用方式：
        error = Error("AgentExecutor.BadRequest.ParamError",
                     "Parameter error!",
                     "Please check your parameter again.")
        error_dict = error.to_dict()
    """

    def __init__(
        self,
        error_code: str,
        description: str,
        solution: str,
        include_trace: Optional[bool] = None,
    ):
        """
        初始化错误对象

        Args:
            error_code: 错误代码，格式为 "Service.Category.ErrorName"
            description: 错误描述，支持国际化
            solution: 解决方案，支持国际化
            include_trace: 是否包含调试追踪信息，默认为 None 时自动从 Config 获取
        """
        self.error_code = error_code
        self.description = description
        self.solution = solution
        self.trace: Optional[str] = None

        # 如果未指定 include_trace，则从 Config 获取
        if include_trace is None:
            from app.common.config import Config

            include_trace = Config.is_debug_mode()

        # 如果需要包含追踪信息，则获取当前调用栈
        if include_trace:
            self._capture_trace()

    def _capture_trace(self) -> None:
        """捕获调试追踪信息"""
        # 获取当前异常信息（如果存在）
        exc_info = sys.exc_info()
        if exc_info[0] is not None:
            # 有活动异常，格式化完整堆栈
            self.trace = "".join(traceback.format_exception(*exc_info))
        else:
            # 没有活动异常，获取当前调用栈（排除当前函数和 __init__）
            stack = traceback.extract_stack()[:-2]
            self.trace = "".join(traceback.format_list(stack))

    def to_dict(self) -> dict:
        """
        转换为字典格式

        Returns:
            dict: 包含错误信息的字典
        """
        error_dict = {
            "ErrorCode": self.error_code,
            "Description": self.description,
            "Solution": self.solution,
        }

        # 只有在 trace 存在时才添加
        if self.trace is not None:
            error_dict["Trace"] = self.trace

        return error_dict

    def __repr__(self) -> str:
        return f"Error(error_code={self.error_code!r})"

    def __str__(self) -> str:
        return self.error_code

    @staticmethod
    def from_dict(error_dict: dict, include_trace: Optional[bool] = None) -> "APIError":
        """
        从字典创建错误对象

        Args:
            error_dict: 错误字典，包含 ErrorCode, Description, Solution
            include_trace: 是否包含追踪信息，默认为 None 时自动从 Config 获取

        Returns:
            APIError: 错误对象
        """
        error = APIError(
            error_code=error_dict.get(
                "ErrorCode", "AgentExecutor.InternalServerError.UnknownError"
            ),
            description=error_dict.get("Description", "Unknown error"),
            solution=error_dict.get("Solution", "Please check the service."),
            include_trace=include_trace,
        )

        # 如果字典中已经包含 Trace，则使用它
        if "Trace" in error_dict:
            error.trace = error_dict["Trace"]

        return error
