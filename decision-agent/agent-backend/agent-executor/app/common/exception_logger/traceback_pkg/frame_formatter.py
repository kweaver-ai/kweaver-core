# -*- coding:utf-8 -*-
"""
Traceback Frame 格式化
"""

import traceback
from typing import List

from ..constants import COLORS, RESET, BOLD, DIM


def format_traceback_frames(
    frames: List[traceback.FrameSummary], colorize: bool = False, show_tree: bool = True
) -> List[str]:
    """
    格式化 traceback frames 为字符串列表，带树状结构展示

    Args:
        frames: FrameSummary 列表
        colorize: 是否添加颜色
        show_tree: 是否显示树状结构

    Returns:
        List[str]: 格式化后的字符串列表
    """
    lines = []
    total_frames = len(frames)

    for idx, frame in enumerate(frames):
        is_last = idx == total_frames - 1

        # 树状结构符号
        if show_tree:
            tree_prefix = "└── " if is_last else "├── "
            tree_continue = "    " if is_last else "│   "
        else:
            tree_prefix = "  "
            tree_continue = "    "

        # 使用完整路径（支持 IDE 点击跳转）
        full_path = frame.filename

        if colorize:
            # 文件路径行（完整路径，可点击跳转）
            file_line = (
                f"{COLORS['border']}{tree_prefix}{RESET}"
                f"{COLORS['caller']}📄 {full_path}:{frame.lineno}{RESET}"
                f"{DIM} in {RESET}"
                f"{BOLD}{COLORS['key']}{frame.name}(){RESET}"
            )
            # 代码行
            if frame.line:
                code_line = (
                    f"{COLORS['border']}{tree_continue}{RESET}"
                    f"{COLORS['project_code']}→ {frame.line.strip()}{RESET}"
                )
            else:
                code_line = None
        else:
            # 无颜色版本也使用完整路径
            file_line = f"{tree_prefix}📄 {full_path}:{frame.lineno} in {frame.name}()"
            code_line = f"{tree_continue}→ {frame.line.strip()}" if frame.line else None

        lines.append(file_line)
        if code_line:
            lines.append(code_line)

    return lines
