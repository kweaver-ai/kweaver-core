"""
Dolphin SDK 技能错误定义
"""

from gettext import gettext as _l
from app.common.errors.api_error_class import APIError


def DolphinSDKSkillError() -> APIError:
    """
    Dolphin SDK 技能错误

    Returns:
        APIError: Error 对象
    """
    return APIError(
        error_code="AgentExecutor.DolphinSDKException.SkillExecption",
        description=_l("DolphinSDKException!"),
        solution=_l("Please check the service."),
    )
