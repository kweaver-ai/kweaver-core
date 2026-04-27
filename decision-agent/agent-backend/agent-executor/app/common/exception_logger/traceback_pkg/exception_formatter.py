# -*- coding:utf-8 -*-
"""
异常格式化（简单和详细）
"""

import traceback

from ..constants import COLORS, RESET, BOLD
from .filter import filter_traceback_frames
from .frame_formatter import format_traceback_frames
from .exception_chain import extract_exception_chain


def format_exception_simple(exc: BaseException, colorize: bool = False) -> str:
    """
    格式化异常（仅项目代码）

    Args:
        exc: 异常对象
        colorize: 是否添加颜色

    Returns:
        str: 格式化后的异常字符串
    """
    lines = []

    # 处理 ExceptionGroup
    if hasattr(exc, "exceptions") and exc.__class__.__name__ == "ExceptionGroup":
        if colorize:
            lines.append(
                f"{COLORS['warning']}📦 ExceptionGroup contains {len(exc.exceptions)} exception(s):{RESET}"
            )
        else:
            lines.append(
                f"📦 ExceptionGroup contains {len(exc.exceptions)} exception(s):"
            )

        for i, sub_exc in enumerate(exc.exceptions, 1):
            if colorize:
                lines.append(
                    f"\n{COLORS['separator']}─── 🔴 Exception {i}/{len(exc.exceptions)} ───{RESET}"
                )
            else:
                lines.append(f"\n─── 🔴 Exception {i}/{len(exc.exceptions)} ───")
            lines.append(format_exception_simple(sub_exc, colorize))
        return "\n".join(lines)

    # 提取异常链
    chain = extract_exception_chain(exc)

    for i, (current_exc, tb) in enumerate(reversed(chain)):
        if i > 0:
            if colorize:
                lines.append(
                    f"\n{COLORS['separator']}⚠️  During handling of the above exception, another exception occurred:{RESET}\n"
                )
            else:
                lines.append(
                    "\n⚠️  During handling of the above exception, another exception occurred:\n"
                )

        # Traceback 头
        if colorize:
            lines.append(f"{BOLD}📚 Call Stack (project code only):{RESET}")
        else:
            lines.append("📚 Call Stack (project code only):")

        # 过滤后的 frames
        frames = filter_traceback_frames(tb, include_all=False)

        if frames:
            lines.extend(format_traceback_frames(frames, colorize))
        else:
            if colorize:
                lines.append(
                    f"  {COLORS['timestamp']}🚧 (No project code in traceback){RESET}"
                )
            else:
                lines.append("  🚧 (No project code in traceback)")

        # 异常类型和消息
        exc_type = type(current_exc).__name__
        exc_msg = str(current_exc)

        if colorize:
            lines.append(
                f"\n{BOLD}💥 {COLORS['exception_type']}{exc_type}{RESET}: {COLORS['exception_msg']}{exc_msg}{RESET}"
            )
        else:
            lines.append(f"\n💥 {exc_type}: {exc_msg}")

    return "\n".join(lines)


def format_exception_detailed(exc: BaseException, colorize: bool = False) -> str:
    """
    格式化异常（包含所有代码）

    Args:
        exc: 异常对象
        colorize: 是否添加颜色

    Returns:
        str: 格式化后的异常字符串
    """
    lines = []

    # 处理 ExceptionGroup
    if hasattr(exc, "exceptions") and exc.__class__.__name__ == "ExceptionGroup":
        lines.append("ExceptionGroup contains multiple exceptions:")
        for i, sub_exc in enumerate(exc.exceptions, 1):
            lines.append(f"\n--- Exception {i} ---")
            lines.append(format_exception_detailed(sub_exc, colorize))
        return "\n".join(lines)

    # 使用标准库格式化完整异常
    tb_lines = traceback.format_exception(type(exc), exc, exc.__traceback__)
    full_traceback = "".join(tb_lines)

    if colorize:
        # 对完整 traceback 添加颜色
        result_lines = []
        for line in full_traceback.split("\n"):
            if line.strip().startswith("File"):
                result_lines.append(f"{COLORS['caller']}{line}{RESET}")
            elif line.strip().startswith("Traceback"):
                result_lines.append(f"{BOLD}{line}{RESET}")
            elif ":" in line and not line.startswith(" "):
                # 异常类型行
                result_lines.append(f"{COLORS['exception_type']}{line}{RESET}")
            else:
                result_lines.append(f"{COLORS['traceback']}{line}{RESET}")
        return "\n".join(result_lines)

    return full_traceback
