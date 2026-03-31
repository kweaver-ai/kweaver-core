"""
未知异常处理器
"""

import os
import sys
import traceback
from fastapi import Request
from fastapi.responses import JSONResponse


from app.common.struct_logger import struct_logger
from app.utils.common import GetRequestLangFunc, GetUnknowError


def handle_unknown_exception(request: Request, exc: Exception):
    """
    处理未知异常

    Args:
        request: FastAPI 请求对象
        exc: 未知异常（可能是 ExceptionGroup）

    Returns:
        JSONResponse: 错误响应
    """

    # 处理 ExceptionGroup（Python 3.11+）
    actual_exc = exc
    if hasattr(exc, "__class__") and exc.__class__.__name__ == "ExceptionGroup":
        # ExceptionGroup 包含多个异常，取第一个实际异常
        if hasattr(exc, "exceptions") and exc.exceptions:
            actual_exc = exc.exceptions[0]
            struct_logger.console_logger.warning(
                f"ExceptionGroup contains {len(exc.exceptions)} exceptions, using first one",
                exception_count=len(exc.exceptions),
            )

    message = "handle_unknown_exception: {}".format(repr(actual_exc))

    # 记录异常日志
    struct_logger.console_logger.error(message, exc_info=actual_exc)

    tr = GetRequestLangFunc(request)
    traceback_list = traceback.extract_tb(sys.exc_info()[2])
    # 异常抛出位置
    if traceback_list:
        file_name, line_number, func_name, line_text = traceback_list[-1]
        file_name = os.path.basename(file_name).split(".")[0]
    else:
        file_name = "common"
        func_name = "common"

    return JSONResponse(
        status_code=500,
        content=GetUnknowError(file_name, func_name, repr(actual_exc), tr),
    )
