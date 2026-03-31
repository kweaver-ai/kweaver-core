# -*- coding:utf-8 -*-
"""
错误日志格式化函数
"""

from datetime import datetime
from typing import Optional, Dict, Any

from ..constants import COLORS, RESET, BOLD, BORDER_SINGLE, BORDER_WIDTH
from ..traceback_pkg import format_exception_simple, format_exception_detailed
from .error_header import format_error_header, format_error_footer


def format_error_console(
    exc: BaseException,
    timestamp: Optional[datetime] = None,
    request_info: Optional[Dict[str, Any]] = None,
) -> str:
    """
    格式化控制台错误日志（仅项目代码）

    Args:
        exc: 异常对象
        timestamp: 时间戳
        request_info: 请求信息

    Returns:
        str: 格式化后的错误日志
    """
    lines = []

    # 头部
    lines.append(format_error_header(exc, timestamp, request_info, colorize=True))

    # Traceback（仅项目代码，带颜色）
    lines.append(f"  {BOLD}{COLORS['exception_type']}Traceback:{RESET}")
    traceback_str = format_exception_simple(exc, colorize=True)
    for line in traceback_str.split("\n"):
        lines.append(f"  {line}")

    # 尾部
    lines.append(format_error_footer(colorize=True))

    return "\n".join(lines)


def format_error_file_simple(
    exc: BaseException,
    timestamp: Optional[datetime] = None,
    request_info: Optional[Dict[str, Any]] = None,
) -> str:
    """
    格式化简单文件错误日志（仅项目代码，无颜色）

    Args:
        exc: 异常对象
        timestamp: 时间戳
        request_info: 请求信息

    Returns:
        str: 格式化后的错误日志
    """
    lines = []

    # 头部
    lines.append(format_error_header(exc, timestamp, request_info, colorize=False))

    # Traceback（仅项目代码）
    lines.append("  Traceback:")
    traceback_str = format_exception_simple(exc, colorize=False)
    for line in traceback_str.split("\n"):
        lines.append(f"  {line}")

    # 尾部
    lines.append(format_error_footer(colorize=False))

    return "\n".join(lines)


def format_error_file_detailed(
    exc: BaseException,
    timestamp: Optional[datetime] = None,
    request_info: Optional[Dict[str, Any]] = None,
) -> str:
    """
    格式化详细文件错误日志（包含所有代码，无颜色）

    Args:
        exc: 异常对象
        timestamp: 时间戳
        request_info: 请求信息

    Returns:
        str: 格式化后的错误日志
    """
    lines = []

    # 头部
    lines.append(format_error_header(exc, timestamp, request_info, colorize=False))

    # Traceback（完整）
    lines.append("  Full Traceback:")
    traceback_str = format_exception_detailed(exc, colorize=False)
    for line in traceback_str.split("\n"):
        lines.append(f"  {line}")

    # 尾部
    lines.append(format_error_footer(colorize=False))

    return "\n".join(lines)


def format_multiple_errors_separator(
    index: int, total: int, colorize: bool = False
) -> str:
    """
    格式化多个错误之间的分隔符

    Args:
        index: 当前错误索引（从1开始）
        total: 错误总数
        colorize: 是否添加颜色

    Returns:
        str: 分隔符字符串
    """
    separator_text = f" Exception {index}/{total} "
    padding = (BORDER_WIDTH - len(separator_text)) // 2

    if colorize:
        return (
            f"\n{COLORS['separator']}"
            f"{BORDER_SINGLE * padding}{separator_text}{BORDER_SINGLE * padding}"
            f"{RESET}\n"
        )
    else:
        return f"\n{BORDER_SINGLE * padding}{separator_text}{BORDER_SINGLE * padding}\n"
