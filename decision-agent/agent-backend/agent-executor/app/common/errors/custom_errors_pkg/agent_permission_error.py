"""
Agent 权限错误定义
"""

from gettext import gettext as _l
from app.common.errors.api_error_class import APIError


def AgentPermissionError() -> APIError:
    """
    Agent 权限错误

    Returns:
        APIError: Error 对象
    """
    return APIError(
        error_code="AgentExecutor.Forbidden.PermissionError",
        description=_l("You do not have permission to execute this agent!"),
        solution=_l("Please check your permission again."),
    )
