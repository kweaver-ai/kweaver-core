"""
异常处理器包
统一导入和注册所有异常处理器
"""

from fastapi.exceptions import RequestValidationError

from app.common.errors import ParamException, AgentPermissionException, CodeException

from .validation_handler import handle_param_error
from .param_handler import handle_param_exception
from .permission_handler import handle_permission_exception
from .code_handler import handle_code_exception

# from .unknown_handler import handle_unknown_exception
from .enhanced_unknown_handler import handle_enhanced_unknown_exception


def register_exception_handlers(app):
    """
    注册所有异常处理器到 FastAPI 应用

    Args:
        app: FastAPI 应用实例
    """
    # 请求参数验证异常
    app.add_exception_handler(RequestValidationError, handle_param_error)

    # 参数异常
    app.add_exception_handler(ParamException, handle_param_exception)

    # 权限异常
    app.add_exception_handler(AgentPermissionException, handle_permission_exception)

    # 代码异常
    app.add_exception_handler(CodeException, handle_code_exception)

    # 未知异常（使用增强版处理器）
    # app.add_exception_handler(Exception, handle_unknown_exception)  # 旧版本，已注释
    app.add_exception_handler(Exception, handle_enhanced_unknown_exception)


# 导出所有处理器函数，方便单独使用
__all__ = [
    "handle_param_error",
    "handle_param_exception",
    "handle_permission_exception",
    "handle_code_exception",
    # "handle_unknown_exception",
    "handle_enhanced_unknown_exception",
    "register_exception_handlers",
]
