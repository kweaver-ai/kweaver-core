"""
Dolphin SDK 基础错误定义
"""

from gettext import gettext as _l
from app.common.errors.api_error_class import APIError


def DolphinSDKBaseError() -> APIError:
    """
    Dolphin SDK 基础错误

    Returns:
        APIError: Error 对象
    """
    return APIError(
        error_code="AgentExecutor.DolphinSDKException.BaseExecption",
        description=_l("DolphinSDKException!"),
        solution=_l("Please check the service."),
    )
