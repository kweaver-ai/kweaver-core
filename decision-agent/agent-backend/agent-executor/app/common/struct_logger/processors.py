"""
结构化日志处理器
"""

import sys
from structlog.types import EventDict, WrappedLogger


def add_caller_info(
    logger: WrappedLogger, method_name: str, event_dict: EventDict
) -> EventDict:
    """
    添加调用者信息到日志

    Args:
        logger: 日志记录器
        method_name: 日志方法名
        event_dict: 事件字典

    Returns:
        更新后的事件字典
    """
    # 获取调用栈
    frame = sys._getframe()
    # 向上查找，跳过日志框架的调用
    depth = 1
    while frame.f_back and depth < 10:
        frame = frame.f_back
        code = frame.f_code
        filename = code.co_filename

        # 跳过 structlog 和当前模块的栈帧
        if "structlog" in filename or "struct_logger" in filename:
            depth += 1
            continue

        # 找到实际调用者
        event_dict["caller"] = f"{filename}:{frame.f_lineno}"
        break

    return event_dict
