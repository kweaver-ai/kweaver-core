"""
权限异常处理器
"""

import sys
import traceback
from fastapi import Request
from fastapi.responses import JSONResponse

from app.common.errors import AgentPermissionException


import app.common.stand_log as log_oper
from app.common.struct_logger import struct_logger
from app.utils.common import GetRequestLangFunc


def handle_permission_exception(request: Request, exc: AgentPermissionException):
    """
    处理权限异常

    Args:
        request: FastAPI 请求对象
        exc: 权限异常

    Returns:
        JSONResponse: 错误响应
    """
    error_log = log_oper.get_error_log(
        repr(exc), sys._getframe(), traceback.format_exc()
    )
    struct_logger.error(error_log)

    tr = GetRequestLangFunc(request)
    return JSONResponse(status_code=403, content=exc.FormatHttpError(tr))
