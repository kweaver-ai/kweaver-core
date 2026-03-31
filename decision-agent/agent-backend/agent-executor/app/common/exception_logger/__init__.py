# -*- coding:utf-8 -*-
"""
异常日志模块

专门用于记录异常日志，提供友好的格式化输出：
1. 控制台日志：仅项目代码的 Traceback，带颜色高亮
2. 简单文件日志：仅项目代码的 Traceback
3. 详细文件日志：完整的 Traceback（包含第三方库）

使用示例:
    from app.common.exception_logger import exception_logger

    try:
        some_operation()
    except Exception as e:
        exception_logger.log_exception(
            e,
            request_info={
                "method": "POST",
                "path": "/api/test",
                "client_ip": "127.0.0.1",
                "account_id": "user_123",
            }
        )
"""

from .logger import ExceptionLogger, exception_logger
from .constants import EXCEPTION_LOG_DIR, EXCEPTION_LOG_SIMPLE, EXCEPTION_LOG_DETAILED
from .traceback_pkg import (
    is_project_file,
    filter_traceback_frames,
    format_exception_simple,
    format_exception_detailed,
)
from .formatter_pkg import (
    format_error_console,
    format_error_file_simple,
    format_error_file_detailed,
)

__all__ = [
    # 主要接口
    "exception_logger",
    "ExceptionLogger",
    # 常量
    "EXCEPTION_LOG_DIR",
    "EXCEPTION_LOG_SIMPLE",
    "EXCEPTION_LOG_DETAILED",
    # Traceback 过滤
    "is_project_file",
    "filter_traceback_frames",
    "format_exception_simple",
    "format_exception_detailed",
    # 格式化
    "format_error_console",
    "format_error_file_simple",
    "format_error_file_detailed",
]
