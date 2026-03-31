"""
错误和异常模块

提供统一的错误定义和异常类体系

- 错误定义（errors）：定义各类错误的错误码、描述和解决方案
- 异常类（exceptions）：用于抛出和捕获的异常对象

向后兼容：
    为了保持向后兼容性，这里同时导出了旧的异常类和新的异常类
    新代码应该使用 app.common.exceptions 中的异常类

注意：
    为了避免循环导入，异常类使用延迟导入（通过 __getattr__）
"""

# 导入错误定义
from app.common.errors.api_error_class import APIError
from app.common.errors.custom_errors_pkg import (
    ParamError,
    AgentPermissionError,
    DolphinSDKModelError,
    DolphinSDKSkillError,
    DolphinSDKBaseError,
    ConversationRunningError,
)
from app.common.errors.external_errors import ExternalServiceError
from app.common.errors.file_errors import AgentExecutor_File_ParseError

# 异常类使用延迟导入，避免循环导入
# 这些类可以从 app.common.exceptions 直接导入
_exceptions_module = None


def _get_exceptions_module():
    """延迟导入 exceptions 模块，避免循环导入"""
    global _exceptions_module
    if _exceptions_module is None:
        from app.common import exceptions

        _exceptions_module = exceptions
    return _exceptions_module


# 使用 __getattr__ 实现延迟导入
def __getattr__(name: str):
    """延迟导入异常类，避免循环导入"""
    if name in (
        "BaseException",
        "CodeException",
        "ParamException",
        "AgentPermissionException",
        "DolphinSDKException",
        "ConversationRunningException",
    ):
        return getattr(_get_exceptions_module(), name)
    raise AttributeError(f"module '{__name__}' has no attribute '{name}'")


__all__ = [
    # 错误定义
    "APIError",
    "ParamError",
    "AgentPermissionError",
    "DolphinSDKModelError",
    "DolphinSDKSkillError",
    "DolphinSDKBaseError",
    "ConversationRunningError",
    "ExternalServiceError",
    "AgentExecutor_File_ParseError",
    # 异常类（通过 __getattr__ 延迟导入）
    "BaseException",
    "CodeException",
    "ParamException",
    "AgentPermissionException",
    "DolphinSDKException",
    "ConversationRunningException",
]
