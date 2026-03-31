"""
会话运行中错误定义
"""

from gettext import gettext as _l
from app.common.errors.api_error_class import APIError


def ConversationRunningError() -> APIError:
    """
    会话运行中错误

    Returns:
        APIError: Error 对象
    """
    return APIError(
        error_code="AgentExecutor.ConflictError.ConversationRunning",
        description=_l("Conversation is running!"),
        solution=_l("Please wait for the conversation to finish."),
    )
