"""
文件日志格式化器
"""

import json
from structlog.types import EventDict, WrappedLogger

from .utils import safe_json_serialize


def format_file_log(
    logger: WrappedLogger, method_name: str, event_dict: EventDict
) -> str:
    """
    格式化文件日志为: "时间 - 级别 - {JSON内容}"
    例如: "2025-10-30 15:42:57 - ERROR - {\"event\":\"错误消息\",\"caller\":\"app.py:123\"}"

    Args:
        logger: 日志记录器
        method_name: 日志方法名
        event_dict: 事件字典

    Returns:
        格式化后的日志字符串
    """
    # 提取时间戳和日志级别
    timestamp = event_dict.pop("timestamp", "")
    level = event_dict.pop("level", "").upper()

    # 安全地序列化 event_dict
    try:
        # 先尝试直接序列化
        json_content = json.dumps(event_dict, ensure_ascii=False, sort_keys=True)
    except (TypeError, ValueError) as e:
        # 如果失败,使用安全序列化
        try:
            safe_event_dict = safe_json_serialize(event_dict)
            json_content = json.dumps(
                safe_event_dict, ensure_ascii=False, sort_keys=True
            )
        except Exception as fallback_error:
            # 最后的兜底方案
            json_content = json.dumps(
                {
                    "error": "Failed to serialize log data",
                    "original_error": str(e),
                    "fallback_error": str(fallback_error),
                    "event_keys": list(event_dict.keys())
                    if isinstance(event_dict, dict)
                    else "not_a_dict",
                },
                ensure_ascii=False,
            )

    # 构造最终格式并直接返回字符串
    return f"{timestamp} - {level} - {json_content}"
