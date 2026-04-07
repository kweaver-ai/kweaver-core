"""
外部服务错误定义
"""

from gettext import gettext as _l
from app.common.errors.api_error_class import APIError


def ExternalServiceError() -> APIError:
    """
    外部服务错误

    Returns:
        APIError: Error 对象
    """
    return APIError(
        error_code="AgentExecutor.InternalError.ExternalServiceError",
        description=_l("External service error!"),
        solution=_l("Please check the service."),
    )
