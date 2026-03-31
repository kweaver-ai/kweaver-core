"""
函数相关错误定义
"""

from gettext import gettext as _l
from app.common.errors.api_error_class import APIError


def AgentExecutor_Function_CodeError() -> APIError:
    """
    函数代码错误

    Returns:
        APIError: Error 对象
    """
    return APIError(
        error_code="AgentExecutor.InternalError.ParseCodeError",
        description=_l("There's an error in the code. "),
        solution=_l("Please check the code."),
    )


def AgentExecutor_Function_InputError() -> APIError:
    """
    函数输入参数错误

    Returns:
        APIError: Error 对象
    """
    return APIError(
        error_code="AgentExecutor.InternalError.RunCodeError",
        description=_l("Incorrect input parameters. "),
        solution=_l("Please check the input arguments."),
    )


def AgentExecutor_Function_RunError() -> APIError:
    """
    函数运行错误

    Returns:
        APIError: Error 对象
    """
    return APIError(
        error_code="AgentExecutor.InternalError.RunCodeError",
        description=_l("Code Execution Failure"),
        solution=_l("Please check the code."),
    )


def AgentExecutor_Function_OutputError() -> APIError:
    """
    函数输出错误

    Returns:
        APIError: Error 对象
    """
    return APIError(
        error_code="AgentExecutor.InternalError.RunCodeError",
        description=_l("Return of the function is not JSON serializable"),
        solution=_l("Please check the code."),
    )
