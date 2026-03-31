"""
文件错误定义
"""

from gettext import gettext as _l
from app.common.errors.api_error_class import APIError


def AgentExecutor_File_ParseError() -> APIError:
    """
    文件解析错误

    Returns:
        APIError: Error 对象
    """
    return APIError(
        error_code="AgentExecutor.InternalError.ParseFileError",
        description=_l("Failed to parse file."),
        solution=_l("Supports parsing text files only."),
    )
