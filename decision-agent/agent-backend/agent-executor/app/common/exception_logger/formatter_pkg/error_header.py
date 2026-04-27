# -*- coding:utf-8 -*-
"""
错误头部和尾部格式化
"""

from datetime import datetime
from typing import Optional, Dict, Any

from ..constants import (
    COLORS,
    RESET,
    BOLD,
    LEVEL_EMOJI,
    BORDER_DOUBLE,
    BORDER_DOT,
    BORDER_WIDTH,
)
from .request_info import format_request_info


def format_error_header(
    exc: BaseException,
    timestamp: Optional[datetime] = None,
    request_info: Optional[Dict[str, Any]] = None,
    colorize: bool = False,
) -> str:
    """
    格式化错误头部信息

    Args:
        exc: 异常对象
        timestamp: 时间戳
        request_info: 请求信息
        colorize: 是否添加颜色

    Returns:
        str: 格式化后的头部
    """
    if timestamp is None:
        timestamp = datetime.now()

    timestamp_str = timestamp.strftime("%Y-%m-%d %H:%M:%S")
    exc_type = type(exc).__name__
    exc_msg = str(exc)

    # 截断过长的消息
    if len(exc_msg) > 150:
        exc_msg = exc_msg[:150] + "..."

    lines = []

    if colorize:
        # 顶部边界
        lines.append(f"{COLORS['border']}{BORDER_DOUBLE * BORDER_WIDTH}{RESET}")

        # 错误标题
        emoji = LEVEL_EMOJI.get("ERROR", "❌")
        lines.append(
            f"{emoji} {BOLD}{COLORS['error']}[ERROR]{RESET} "
            f"{BOLD}{COLORS['exception_type']}{exc_type}{RESET} "
            f"{COLORS['timestamp']}({timestamp_str}){RESET}"
        )

        # 错误消息
        lines.append(f"  💬 {COLORS['exception_msg']}{exc_msg}{RESET}")

        # 分隔线
        lines.append(f"{COLORS['border']}{BORDER_DOT * BORDER_WIDTH}{RESET}")

        # 请求信息（详细）
        if request_info:
            lines.append(format_request_info(request_info, colorize=True))
            lines.append(f"{COLORS['border']}{BORDER_DOT * BORDER_WIDTH}{RESET}")
    else:
        # 无颜色版本
        lines.append(BORDER_DOUBLE * BORDER_WIDTH)
        lines.append(f"❌ [ERROR] {exc_type} ({timestamp_str})")
        lines.append(f"  💬 {exc_msg}")
        lines.append(BORDER_DOT * BORDER_WIDTH)

        if request_info:
            lines.append(format_request_info(request_info, colorize=False))
            lines.append(BORDER_DOT * BORDER_WIDTH)

    return "\n".join(lines)


def format_error_footer(colorize: bool = False) -> str:
    """
    格式化错误尾部

    Args:
        colorize: 是否添加颜色

    Returns:
        str: 格式化后的尾部
    """
    if colorize:
        return f"{COLORS['border']}{BORDER_DOUBLE * BORDER_WIDTH}{RESET}\n"
    else:
        return f"{BORDER_DOUBLE * BORDER_WIDTH}\n"
