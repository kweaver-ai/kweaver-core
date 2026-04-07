"""
参数错误定义
"""

from gettext import gettext as _l
from app.common.errors.api_error_class import APIError


def ParamError() -> APIError:
    """
    参数错误

    Returns:
        APIError: Error 对象
    """
    return APIError(
        error_code="AgentExecutor.BadRequest.ParamError",
        description=_l("Parameter error!"),
        solution=_l("Please check your parameter again."),
    )
