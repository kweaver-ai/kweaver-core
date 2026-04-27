"""
Dolphin SDK 模型错误定义
"""

from gettext import gettext as _l
from app.common.errors.api_error_class import APIError


def DolphinSDKModelError() -> APIError:
    """
    Dolphin SDK 模型错误

    Returns:
        APIError: Error 对象
    """
    return APIError(
        error_code="AgentExecutor.DolphinSDKException.ModelExecption",
        description=_l("DolphinSDKException!"),
        solution=_l("Please check the service."),
    )
