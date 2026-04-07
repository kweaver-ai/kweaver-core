# -*- coding:utf-8 -*-
"""
Traceback 过滤功能
"""

import traceback
from typing import List, Optional
from types import TracebackType

from ..constants import PROJECT_ROOT


def is_project_file(filepath: str) -> bool:
    """
    判断文件是否属于项目代码

    Args:
        filepath: 文件路径

    Returns:
        bool: 是否为项目代码
    """
    if not filepath:
        return False

    # 排除虚拟环境和 site-packages
    excluded_patterns = [
        ".venv",
        "site-packages",
        "dist-packages",
        "/lib/python",
        "/.pyenv/",
    ]

    for pattern in excluded_patterns:
        if pattern in filepath:
            return False

    # 检查是否在项目根目录下
    return filepath.startswith(PROJECT_ROOT)


def filter_traceback_frames(
    tb: Optional[TracebackType], include_all: bool = False
) -> List[traceback.FrameSummary]:
    """
    过滤 traceback frames，可选择只保留项目代码

    Args:
        tb: Traceback 对象
        include_all: 是否包含所有代码（包括第三方库）

    Returns:
        List[traceback.FrameSummary]: 过滤后的 frame 列表
    """
    if tb is None:
        return []

    frames = traceback.extract_tb(tb)

    if include_all:
        return list(frames)

    # 只保留项目代码
    return [frame for frame in frames if is_project_file(frame.filename)]
