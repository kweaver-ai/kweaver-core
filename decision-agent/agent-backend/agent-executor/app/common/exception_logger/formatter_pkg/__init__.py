# -*- coding:utf-8 -*-
"""
错误日志格式化器包

提供友好的错误日志展示格式，支持表格、emoji、颜色等
"""

from .table_drawer import TableDrawer
from .request_info import format_request_info
from .error_header import format_error_header, format_error_footer
from .error_formatters import (
    format_error_console,
    format_error_file_simple,
    format_error_file_detailed,
    format_multiple_errors_separator,
)

__all__ = [
    # 表格绘制
    "TableDrawer",
    # 请求信息格式化
    "format_request_info",
    # 错误头部/尾部
    "format_error_header",
    "format_error_footer",
    # 错误日志格式化
    "format_error_console",
    "format_error_file_simple",
    "format_error_file_detailed",
    "format_multiple_errors_separator",
]
